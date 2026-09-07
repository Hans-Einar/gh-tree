package broker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type nativeOutput struct {
	stream byte
	file   *os.File
}

// userProcess has one OS-thread owner through creation/debug detachment. Each
// allocation is assigned here before the next fallible operation.
type userProcess struct {
	cwd                   *AcquiredDirectory
	job                   windows.Handle
	debug                 debugOwner
	hpc                   windows.Handle
	input                 *os.File
	outputs               []nativeOutput
	childHandles          []windows.Handle
	stdin, stdout, stderr windows.Handle
	attributes            *windows.ProcThreadAttributeListContainer
	terminalClosed        chan struct{}
	exit                  uint32
	rootWaited            bool
	hook                  func(string)
	closeTerminal         func(windows.Handle)
}

func (p *userProcess) pipe() (windows.Handle, windows.Handle, error) {
	var r, w windows.Handle
	err := windows.CreatePipe(&r, &w, nil, 64<<10)
	if r != 0 {
		p.childHandles = append(p.childHandles, r)
	}
	if w != 0 {
		p.childHandles = append(p.childHandles, w)
	}
	return r, w, err
}

func (p *userProcess) take(h windows.Handle, name string) *os.File {
	for i, owned := range p.childHandles {
		if owned == h {
			p.childHandles[i] = 0
			break
		}
	}
	return os.NewFile(uintptr(h), name)
}

func (p *userProcess) prepare(spec StartSpec) error {
	var err error
	p.cwd, err = acquireCwd(spec, p.hook)
	if err != nil {
		return err
	}
	p.job, err = newJob()
	if err != nil {
		return err
	}
	if p.hook != nil {
		p.hook("inner-job-created")
	}
	r, w, err := p.pipe()
	if err != nil {
		return err
	}
	p.stdin, p.input = r, p.take(w, "user-input")
	r, w, err = p.pipe()
	if err != nil {
		return err
	}
	p.stdout = w
	stream := byte(1)
	if spec.Terminal {
		stream = 3
	}
	p.outputs = append(p.outputs, nativeOutput{stream, p.take(r, "user-output")})
	if spec.Terminal {
		err = windows.CreatePseudoConsole(windows.Coord{X: int16(spec.Columns), Y: int16(spec.Rows)}, p.stdin, p.stdout, 0, &p.hpc)
		if err != nil {
			return err
		}
		if p.hook != nil {
			p.hook("conpty-created")
		}
	} else {
		r, w, err = p.pipe()
		if err != nil {
			return err
		}
		p.stderr = w
		p.outputs = append(p.outputs, nativeOutput{2, p.take(r, "user-stderr")})
	}
	return nil
}

func environmentBlock(entries []string) ([]uint16, error) {
	var block []uint16
	for _, entry := range entries {
		part, err := windows.UTF16FromString(entry)
		if err != nil {
			return nil, err
		}
		block = append(block, part...)
	}
	block = append(block, 0)
	if len(block) == 1 {
		block = append(block, 0)
	}
	return block, nil
}

func environmentValue(entries []string, key string) string {
	for _, entry := range entries {
		name, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(name, key) {
			return value
		}
	}
	return ""
}

func resolveWindowsCommand(spec StartSpec, cwd string) (string, string, error) {
	name := spec.Executable
	var bases []string
	if filepath.IsAbs(name) {
		bases = []string{name}
	} else if strings.ContainsAny(name, `/\:`) {
		if filepath.VolumeName(name) != "" {
			return "", "", errors.New("drive-relative executable unsupported")
		}
		bases = []string{filepath.Join(cwd, name)}
	} else {
		bases = append(bases, filepath.Join(cwd, name))
		for _, dir := range filepath.SplitList(environmentValue(spec.Environment, "PATH")) {
			if dir == "" {
				continue
			}
			if !filepath.IsAbs(dir) {
				dir = filepath.Join(cwd, dir)
			}
			bases = append(bases, filepath.Join(dir, name))
		}
	}
	extensions := []string{""}
	if filepath.Ext(name) == "" {
		extensions = []string{".com", ".exe", ".bat", ".cmd"}
		if value := environmentValue(spec.Environment, "PATHEXT"); value != "" {
			extensions = strings.Split(value, ";")
		}
	}
	var executable string
	for _, base := range bases {
		for _, ext := range extensions {
			if ext != "" && (!strings.HasPrefix(ext, ".") || strings.ContainsAny(ext, `/\:`)) {
				continue
			}
			candidate := base + ext
			if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
				executable = candidate
				break
			}
		}
		if executable != "" {
			break
		}
	}
	if executable == "" {
		return "", "", os.ErrNotExist
	}
	args := append([]string{executable}, spec.Arguments...)
	ext := strings.ToLower(filepath.Ext(executable))
	if ext != ".cmd" && ext != ".bat" {
		return executable, windows.ComposeCommandLine(args), nil
	}
	// The sole reviewed shell carrier. Quoting here is deliberately distinct
	// from native argv quoting: all operands are quoted, including empty ones.
	quoted := make([]string, len(args))
	for i, arg := range args {
		if strings.ContainsAny(arg, "\x00\r\n\"%") {
			return "", "", errors.New("unsafe batch carrier operand")
		}
		trailing := len(arg) - len(strings.TrimRight(arg, `\`))
		quoted[i] = `"` + arg + strings.Repeat(`\`, trailing) + `"`
	}
	system, err := windows.GetSystemDirectory()
	if err != nil {
		return "", "", err
	}
	cmd := filepath.Join(system, "cmd.exe")
	return cmd, windows.EscapeArg(cmd) + ` /D /V:OFF /S /C "` + strings.Join(quoted, " ") + `"`, nil
}

func (p *userProcess) start(ctx context.Context, spec StartSpec) error {
	if err := assertStartupLayout(); err != nil {
		return err
	}
	executable, command, err := resolveWindowsCommand(spec, p.cwd.Path())
	if err != nil {
		return err
	}
	app, err := windows.UTF16PtrFromString(executable)
	if err != nil {
		return err
	}
	cmd, err := windows.UTF16PtrFromString(command)
	if err != nil {
		return err
	}
	dir, err := windows.UTF16PtrFromString(p.cwd.Path())
	if err != nil {
		return err
	}
	env, err := environmentBlock(spec.Environment)
	if err != nil {
		return err
	}
	p.attributes, err = windows.NewProcThreadAttributeList(1)
	if err != nil {
		return err
	}
	if spec.Terminal {
		// This attribute takes the opaque HPCON value itself, not a pointer to
		// Go memory. Keep it a uintptr through the native syscall boundary.
		err = nativeCall("UpdateProcThreadAttribute", uintptr(unsafe.Pointer(p.attributes.List())), 0, windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE, uintptr(p.hpc), unsafe.Sizeof(p.hpc), 0, 0)
	} else {
		handles := []windows.Handle{p.stdin, p.stdout, p.stderr}
		for _, h := range handles {
			if err = windows.SetHandleInformation(h, windows.HANDLE_FLAG_INHERIT, windows.HANDLE_FLAG_INHERIT); err != nil {
				return err
			}
		}
		err = p.attributes.Update(windows.PROC_THREAD_ATTRIBUTE_HANDLE_LIST, unsafe.Pointer(&handles[0]), uintptr(len(handles))*unsafe.Sizeof(handles[0]))
		runtime.KeepAlive(handles)
	}
	if err != nil {
		return err
	}
	si := windows.StartupInfoEx{}
	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = windows.STARTF_USESHOWWINDOW | windows.STARTF_USESTDHANDLES
	si.ShowWindow = windows.SW_HIDE
	si.StdInput, si.StdOutput, si.StdErr = p.stdin, p.stdout, p.stderr
	si.ProcThreadAttributeList = p.attributes.List()
	flags := uint32(windows.CREATE_SUSPENDED | windows.CREATE_UNICODE_ENVIRONMENT | windows.EXTENDED_STARTUPINFO_PRESENT | 2)
	if !spec.Terminal {
		flags |= windows.CREATE_NO_WINDOW
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	err = windows.CreateProcess(app, cmd, nil, nil, !spec.Terminal, flags, &env[0], dir, &si.StartupInfo, &p.debug.process)
	runtime.KeepAlive(env)
	runtime.KeepAlive(p.attributes)
	if err != nil {
		return err
	}
	p.debug.attached = true
	if p.hook != nil {
		p.hook("user-created-suspended")
	}
	if err = windows.AssignProcessToJobObject(p.job, p.debug.process.Process); err != nil {
		return err
	}
	if p.hook != nil {
		p.hook("user-assigned-inner-job")
	}
	if err = p.closeChildHandles(); err != nil {
		return err
	}
	if err = p.debug.barrier(ctx, p.cwd, p.hook); err != nil {
		return err
	}
	p.attributes.Delete()
	p.attributes = nil
	return closeHandle(&p.debug.process.Thread)
}

func (p *userProcess) closeChildHandles() error {
	var err error
	for i := range p.childHandles {
		err = errors.Join(err, closeHandle(&p.childHandles[i]))
	}
	if err == nil {
		p.childHandles = nil
	}
	return err
}

func closeFile(file **os.File) error {
	if *file == nil {
		return nil
	}
	err := (*file).Close()
	if err == nil || errors.Is(err, os.ErrClosed) {
		*file = nil
		return nil
	}
	return err
}

// cleanup is called by the debug creating thread on failed startup too. It
// never releases the process/Job capability on an unproved wait or membership.
// Output readers belong to the parent and continue draining through HPCON close.
func (p *userProcess) cleanup(ctx context.Context) error {
	var result error
	if p.job != 0 {
		result = errors.Join(result, windows.TerminateJobObject(p.job, 1))
	}
	if p.debug.process.Process != 0 && !p.rootWaited {
		// Assignment may have failed while the root was still suspended.
		var terminateErr error
		state, waitErr := windows.WaitForSingleObject(p.debug.process.Process, 0)
		if waitErr != nil {
			result = errors.Join(result, waitErr)
		} else if state != windows.WAIT_OBJECT_0 {
			terminateErr = windows.TerminateProcess(p.debug.process.Process, 1)
		}
		if p.debug.attached {
			result = errors.Join(result, p.debug.continuation(), p.debug.detach())
		}
		exit, err := waitProcess(ctx, p.debug.process.Process)
		if err == nil {
			p.exit = exit
			p.rootWaited = true
		} else {
			result = errors.Join(result, terminateErr)
		}
		result = errors.Join(result, err)
	}
	if p.job != 0 {
		result = errors.Join(result, waitJob(ctx, p.job))
	}
	if result != nil {
		return result
	}
	result = errors.Join(result, p.closeChildHandles(), closeFile(&p.input))
	if p.hpc != 0 && p.terminalClosed == nil {
		p.terminalClosed = make(chan struct{})
		hpc, done := p.hpc, p.terminalClosed
		closer := p.closeTerminal
		if closer == nil {
			closer = windows.ClosePseudoConsole
		}
		go func() { closer(hpc); close(done) }()
	}
	if p.terminalClosed != nil {
		select {
		case <-p.terminalClosed:
			p.hpc = 0
		case <-ctx.Done():
			return errors.Join(result, ctx.Err())
		}
	}
	if p.cwd != nil {
		result = errors.Join(result, p.cwd.Close())
	}
	if p.attributes != nil {
		p.attributes.Delete()
		p.attributes = nil
	}
	result = errors.Join(result, p.debug.closeDebugHandles(), closeHandle(&p.debug.process.Thread), closeHandle(&p.debug.process.Process), closeHandle(&p.job))
	return result
}

func boundedCleanup(p *userProcess) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return p.cleanup(ctx)
}
