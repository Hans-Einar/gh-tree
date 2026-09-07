package runtime

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/Hans-Einar/gh-tree/internal/application/api"
	"github.com/Hans-Einar/gh-tree/internal/runtime/broker"
	"github.com/Hans-Einar/gh-tree/internal/runtime/brokerassets"
	"golang.org/x/sys/windows"
)

// These ordinary selected test functions are parent/user fixtures only. The
// broker under test is always the actual selected committed broker/cmd image.
// No TestMain, private dispatcher or native production source is substituted.
func bindingArgs(marker string) []string {
	for i, arg := range os.Args {
		if arg == marker {
			return os.Args[i+1:]
		}
	}
	return nil
}

func bindingMachine(t *testing.T) (uint16, uint16) {
	t.Helper()
	var process, native uint16
	if err := windows.IsWow64Process2(windows.CurrentProcess(), &process, &native); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	f, err := pe.Open(exe)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	compiled := map[string]uint16{"386": pe.IMAGE_FILE_MACHINE_I386, "amd64": pe.IMAGE_FILE_MACHINE_AMD64, "arm64": pe.IMAGE_FILE_MACHINE_ARM64}[goruntime.GOARCH]
	if f.Machine != compiled || (process != 0 && process != compiled) {
		t.Fatalf("compiled=%s PE=%04x actual=%04x native=%04x", goruntime.GOARCH, f.Machine, process, native)
	}
	return native, compiled
}

func TestWindowsCommittedHelperBinding(t *testing.T) {
	native, actual := bindingMachine(t)
	if native != actual {
		t.Skip("matrix driver runs on native amd64/ARM64; emulated parents execute the selected binding fixture")
	}
	arches := []string{"386"}
	helperArch := "amd64"
	switch native {
	case pe.IMAGE_FILE_MACHINE_AMD64:
	case pe.IMAGE_FILE_MACHINE_ARM64:
		arches = []string{"amd64", "386"}
		helperArch = "arm64"
		for _, arch := range []string{"amd64", "arm64"} {
			if _, err := brokerassets.Load(arch); err == nil {
				t.Fatal("native ARM64 unexpectedly embeds a helper")
			}
		}
	case pe.IMAGE_FILE_MACHINE_I386:
		t.Skip("native 32-bit Windows uses its own executable and has no embedded-helper route")
	default:
		t.Fatal("unsupported native machine")
	}
	manifestBytes, err := os.ReadFile("brokerassets/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Targets []struct {
			Arch, SHA256 string
			Machine      uint16
		}
	}
	if err = json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	expected := ""
	for _, target := range manifest.Targets {
		if target.Arch == helperArch && target.Machine == native {
			expected = target.SHA256
		}
	}
	if expected == "" {
		t.Fatal("manifest lacks native target")
	}
	user, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("native driver=%s/%04x Go=%s manifestSHA256=%x helperSHA256=%s", goruntime.GOARCH, native, goruntime.Version(), sha256.Sum256(manifestBytes), expected)
	for _, arch := range arches {
		t.Run(arch, func(t *testing.T) {
			parent := filepath.Join(t.TempDir(), "binding-parent-"+arch+".exe")
			ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, filepath.Join(goruntime.GOROOT(), "bin", "go.exe"), "test", "-c", "-o", parent, ".")
			for _, v := range os.Environ() {
				k, _, _ := strings.Cut(v, "=")
				if !strings.EqualFold(k, "GOARCH") && !strings.EqualFold(k, "CGO_ENABLED") {
					cmd.Env = append(cmd.Env, v)
				}
			}
			cmd.Env = append(cmd.Env, "GOARCH="+arch, "CGO_ENABLED=0")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("compile owned %s parent: %v\n%s", arch, err, out)
			}
			cmd = exec.CommandContext(ctx, parent, "-test.run=^TestWindowsCommittedHelperParent$", "-test.v", "-test.timeout=90s", "--", "--committed-helper-parent", user, helperArch, expected, arch)
			out, err := cmd.CombinedOutput()
			t.Logf("executed %s parent, native helper %s:\n%s", arch, helperArch, out)
			if err != nil {
				t.Fatalf("actual parent execution: %v", err)
			}
			if !bytes.Contains(out, []byte("COMMITTED_BINDING_COMPLETE")) {
				t.Fatal("parent did not execute binding assertions")
			}
		})
	}
}

func bindingSpec(t *testing.T, user, mode string) broker.StartSpec {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "binding-cwd.txt"), []byte("exact acquired cwd"), 0600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	var info windows.ByHandleFileInformation
	err = windows.GetFileInformationByHandle(windows.Handle(file.Fd()), &info)
	if err != nil {
		file.Close()
		t.Fatal(err)
	}
	birth := uint64(info.CreationTime.HighDateTime)<<32 | uint64(info.CreationTime.LowDateTime)
	id, err := broker.ObserveDirectory(file, fmt.Sprintf("birth-filetime:%d", birth))
	closeErr := file.Close()
	if err != nil || closeErr != nil {
		t.Fatal(errors.Join(err, closeErr))
	}
	return broker.StartSpec{ParentID: uint64(os.Getpid()), OperationID: 1, RootLocator: root, RootIdentity: id, ProjectIdentity: id, Executable: user, Arguments: []string{"-test.run=^TestWindowsCommittedHelperUser$", "--", "--committed-helper-user", mode}, Environment: os.Environ(), Rows: 24, Columns: 80, Terminal: mode == "terminal"}
}

type bindingOutput struct {
	mu   sync.Mutex
	ring outputRing
}

func (o *bindingOutput) append(stream api.OutputStream, b []byte) {
	o.mu.Lock()
	defer o.mu.Unlock()
	seq, _ := api.NewSessionSequence(1)
	if err := o.ring.append(stream, seq, b); err != nil {
		panic(err)
	}
}
func (o *bindingOutput) text() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.textLocked()
}
func (o *bindingOutput) wait(t *testing.T, marker string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !strings.Contains(o.text(), marker) {
		if time.Now().After(deadline) {
			t.Fatalf("missing %q output=%q", marker, o.text())
		}
		time.Sleep(time.Millisecond)
	}
}

func bindingExtract(t *testing.T, arch, expected string) *broker.WindowsImage {
	t.Helper()
	asset, err := brokerassets.Load(arch)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(asset.Bytes)
	if asset.SHA256 != expected || hex.EncodeToString(hash[:]) != expected || asset.Protocol != broker.ProtocolVersion {
		t.Fatal("selected embedded image differs from exact manifest")
	}
	image, err := broker.ExtractWindowsImage(asset.Bytes, asset.Machine, hash, asset.Protocol)
	if image != nil {
		t.Cleanup(func() {
			if e := image.Cleanup(); e != nil {
				t.Error(e)
			}
		})
	}
	if err != nil {
		t.Fatal(err)
	}
	disk, err := os.ReadFile(image.Path())
	if err != nil || sha256.Sum256(disk) != hash {
		t.Fatalf("retained extracted bytes mismatch: %v", err)
	}
	return image
}

// Acquire a read-only retained observation of this test's uniquely extracted
// child. The numerical census is never used to kill or grant cleanup authority.
func bindingHelperHandle(t *testing.T, path string, native uint16) windows.Handle {
	t.Helper()
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ParentProcessID != uint32(os.Getpid()) {
			continue
		}
		h, e := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, entry.ProcessID)
		if e != nil {
			continue
		}
		buf := make([]uint16, 32768)
		n := uint32(len(buf))
		e = windows.QueryFullProcessImageName(h, 0, &buf[0], &n)
		if e != nil || !strings.EqualFold(windows.UTF16ToString(buf[:n]), path) {
			windows.CloseHandle(h)
			continue
		}
		var process, actual uint16
		e = windows.IsWow64Process2(h, &process, &actual)
		if e != nil || process != 0 || actual != native {
			windows.CloseHandle(h)
			t.Fatalf("executing helper machine=%04x native=%04x expected=%04x err=%v", process, actual, native, e)
		}
		t.Cleanup(func() {
			if e := windows.CloseHandle(h); e != nil {
				t.Error(e)
			}
		})
		t.Logf("retained actual committed helper native=%04x emulation=%04x", actual, process)
		return h
	}
	t.Fatal("actual committed helper child not found")
	return 0
}

func bindingUserHandle(t *testing.T, imagePath, output string, native uint16) windows.Handle {
	t.Helper()
	_, tail, ok := strings.Cut(output, "pid=")
	if !ok || len(strings.Fields(tail)) == 0 {
		t.Fatal("missing actual user identity")
	}
	pid, err := strconv.ParseUint(strings.Fields(tail)[0], 10, 32)
	if err != nil {
		t.Fatal(err)
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := windows.CloseHandle(h); err != nil {
			t.Error(err)
		}
	})
	buf := make([]uint16, 32768)
	n := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &n); err != nil || !strings.EqualFold(windows.UTF16ToString(buf[:n]), imagePath) {
		t.Fatalf("actual user image mismatch: %v", err)
	}
	var process, actual uint16
	if err := windows.IsWow64Process2(h, &process, &actual); err != nil || process != 0 || actual != native {
		t.Fatalf("actual user machine=%04x native=%04x expected=%04x err=%v", process, actual, native, err)
	}
	return h
}

func bindingJoin(t *testing.T, c *broker.WindowsClient, imagePath string) broker.WindowsFact {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	f, e := c.Wait(ctx)
	if !f.CleanupComplete || len(f.Residuals) != 0 {
		t.Fatalf("native owned cleanup incomplete: %+v %v", f, e)
	}
	if _, err := os.Stat(imagePath); !os.IsNotExist(err) {
		t.Fatalf("owned extraction survived joined cleanup: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(imagePath)); !os.IsNotExist(err) {
		t.Fatalf("owned extraction directory survived: %v", err)
	}
	return f
}

func TestWindowsCommittedHelperParent(t *testing.T) {
	args := bindingArgs("--committed-helper-parent")
	if len(args) == 0 {
		return
	}
	if len(args) != 4 {
		t.Fatal("invalid parent fixture args")
	}
	user, arch, expected, parentArch := args[0], args[1], args[2], args[3]
	native, actual := bindingMachine(t)
	if goruntime.GOARCH != parentArch || native == actual {
		t.Fatal("parent must actually execute requested emulated architecture")
	}
	route, embedded, err := broker.MachineRoute()
	if err != nil || route != native || !embedded {
		t.Fatalf("actual native routing %04x %t %v", route, embedded, err)
	}
	if arch == "amd64" && native != pe.IMAGE_FILE_MACHINE_AMD64 || arch == "arm64" && native != pe.IMAGE_FILE_MACHINE_ARM64 {
		t.Fatal("wrong actual host")
	}
	t.Logf("executing parent=%s image=%04x native=%04x manifest imageSHA256=%s", goruntime.GOARCH, actual, native, expected)
	for _, mode := range []string{"terminal", "blocked", "flood", "missing", "stale"} {
		t.Run(mode, func(t *testing.T) {
			s := bindingSpec(t, user, mode)
			if mode == "missing" {
				s.Executable = filepath.Join(s.RootLocator, "nonexistent-binding-command.exe")
			}
			if mode == "stale" {
				other := bindingSpec(t, user, "unused")
				s.ProjectIdentity = other.ProjectIdentity
			}
			image := bindingExtract(t, arch, expected)
			path := image.Path()
			var output bindingOutput
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			c, start, startErr := broker.StartWindows(ctx, broker.WindowsConfig{SessionID: 72, Spec: s, Image: path, Extraction: image, Output: output.append, GracePeriod: 20 * time.Millisecond, ForcePeriod: 3 * time.Second})
			if c != nil {
				t.Cleanup(func() { c.Stop(); bindingJoin(t, c, path) })
			}
			if mode == "missing" || mode == "stale" {
				cause, stage, sentinel := broker.WindowsNotFoundFailure, api.ProcessContainment, error(os.ErrNotExist)
				if mode == "stale" {
					cause, stage, sentinel = broker.WindowsCwdFailure, api.CwdAcquisition, broker.ErrCwd
				}
				var failure *broker.WindowsFailure
				if c == nil || start.Established || !errors.As(startErr, &failure) || failure.Cause != cause || failure.Stage != stage || failure.Cleanup || !errors.Is(startErr, sentinel) {
					t.Fatalf("committed native typed failure erased: start=%+v err=%v typed=%+v", start, startErr, failure)
				}
				final := bindingJoin(t, c, path)
				if !errors.Is(final.Err, sentinel) || strings.Contains(output.text(), "BINDING_USER_READY") {
					t.Fatalf("failure facts/user execution: %+v %q", final, output.text())
				}
				return
			}
			if startErr != nil || c == nil || !start.Established || start.Cwd != s.RootLocator || start.Architecture.NativeMachine != native || start.Architecture.ImageMachine != native || start.Architecture.ProcessMachine != 0 || start.Architecture.InitialBreakpoint != 0x80000003 {
				t.Fatalf("committed startup/native cwd barrier: %+v %v", start, startErr)
			}

			if mode == "flood" {
				final := bindingJoin(t, c, path)
				if final.Err != nil || !final.RootExited || final.ExitCode != 0 {
					t.Fatalf("flood final %+v", final)
				}
				output.mu.Lock()
				defer output.mu.Unlock()
				if output.ring.end < 1<<20 || output.ring.end-output.ring.start != outputCapacity || !strings.Contains(output.textLocked(), "BINDING_FLOOD_DONE") {
					t.Fatalf("committed output drain/ring end=%d start=%d", output.ring.end, output.ring.start)
				}
				return
			}
			output.wait(t, "BINDING_USER_READY")
			helper := bindingHelperHandle(t, path, native)
			userHandle := bindingUserHandle(t, user, output.text(), native)

			canceled, stop := context.WithCancel(context.Background())
			stop()
			d, e := c.Write(canceled, []byte("CANCELED_BINDING\r"))
			if !errors.Is(e, context.Canceled) || d.Dispatched || d.Receipt != nil {
				t.Fatalf("canceled native admission %+v %v", d, e)
			}
			if mode == "terminal" {
				d, e = c.Resize(ctx, 31, 101)
				if e != nil || !d.Completed || !d.Dispatched || d.Receipt == nil {
					t.Fatalf("resize %+v %v", d, e)
				}
				d, e = c.Write(ctx, []byte("COMMITTED_INPUT\r"))
				if e != nil || !d.Completed || d.Delivered != 16 || d.Receipt == nil {
					t.Fatalf("input %+v %v", d, e)
				}
				final := bindingJoin(t, c, path)
				if final.Err != nil || final.ExitCode != 0 || !final.RootExited || !final.Quiescent || !strings.Contains(output.text(), "INPUT COMMITTED_INPUT SIZE 101 31") || strings.Contains(output.text(), "CANCELED_BINDING") {
					t.Fatalf("terminal final %+v output=%q", final, output.text())
				}
			} else {
				data := bytes.Repeat([]byte{'A'}, 65484)
				d, e = c.Write(ctx, data)
				if e != nil || !d.Completed || d.Delivered != uint32(len(data)) {
					t.Fatalf("actual first pipe fill %+v %v", d, e)
				}
				short, stop := context.WithTimeout(context.Background(), 100*time.Millisecond)
				pending, e := c.Write(short, data)
				stop()
				if !errors.Is(e, context.DeadlineExceeded) || pending.Completed || !pending.Dispatched || pending.Receipt == nil {
					t.Fatalf("pending receipt lost: %+v %v", pending, e)
				}
				refused, e := c.Write(ctx, []byte{'B'})
				if !errors.Is(e, broker.ErrWindowsControlsBusy) || refused.Dispatched || refused.Receipt != nil {
					t.Fatalf("native input bound: %+v %v", refused, e)
				}
				c.Stop()
				c.Stop()
				final := bindingJoin(t, c, path)
				if final.Err != nil || !final.Quiescent || !final.RootExited {
					t.Fatalf("Stop before receipt observation %+v", final)
				}
				for i := 0; i < 3; i++ {
					known, e := pending.Receipt.Wait(canceled)
					if !errors.Is(e, io.ErrShortWrite) || !known.Completed || !known.Dispatched || known.Accepted != uint32(len(data)) || known.Delivered >= known.Accepted || known.Receipt != pending.Receipt {
						t.Fatalf("committed eventual receipt %+v %v", known, e)
					}
				}
			}

			state, e := windows.WaitForSingleObject(userHandle, 0)
			if e != nil || state != windows.WAIT_OBJECT_0 {
				t.Fatalf("retained actual user survives cleanup: %d %v", state, e)
			}
			state, e = windows.WaitForSingleObject(helper, 0)
			if e != nil || state != windows.WAIT_OBJECT_0 {
				t.Fatalf("retained actual helper survives cleanup: %d %v", state, e)
			}
		})
	}
	t.Log("COMMITTED_BINDING_COMPLETE")
}

func (o *bindingOutput) textLocked() string {
	b := make([]byte, o.ring.end-o.ring.start)
	for i := range b {
		b[i] = o.ring.bytes[(o.ring.start+uint64(i))%outputCapacity]
	}
	return string(b)
}

func TestWindowsCommittedHelperUser(t *testing.T) {
	args := bindingArgs("--committed-helper-user")
	if len(args) == 0 {
		return
	}
	if len(args) != 1 {
		t.Fatal("invalid user fixture")
	}
	b, err := os.ReadFile("binding-cwd.txt")
	if err != nil || string(b) != "exact acquired cwd" {
		t.Fatalf("actual cwd: %q %v", b, err)
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".gh-tree-start-") {
			t.Fatal("startup anchor reached user initialization")
		}
	}
	debug, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("IsDebuggerPresent").Call()
	if debug != 0 {
		t.Fatal("debugger retained into user code")
	}
	native, actual := bindingMachine(t)
	if native != actual {
		t.Fatal("user fixture must execute natively")
	}
	fmt.Printf("BINDING_USER_READY native=%04x image=%04x pid=%s\n", native, actual, strconv.Itoa(os.Getpid()))
	switch args[0] {
	case "terminal":
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		var info windows.ConsoleScreenBufferInfo
		if err = windows.GetConsoleScreenBufferInfo(windows.Handle(os.Stdout.Fd()), &info); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("INPUT %s SIZE %d %d\n", strings.TrimSpace(line), info.Window.Right-info.Window.Left+1, info.Window.Bottom-info.Window.Top+1)
	case "blocked":
		time.Sleep(45 * time.Second)
		t.Fatal("owned blocked fixture expired before Stop")
	case "flood":
		block := bytes.Repeat([]byte{0, 0xff, 0x1b, 0x80, 'A', 'B', 'C', 'D'}, 32768)
		for i := 0; i < 4; i++ {
			if _, err := os.Stdout.Write(block); err != nil {
				t.Fatal(err)
			}
		}
		fmt.Println("BINDING_FLOOD_DONE")
	default:
		t.Fatal("unexpected user mode")
	}
}
