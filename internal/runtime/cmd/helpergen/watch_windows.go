package main

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

// One continuous native name-change request covers each input directory.
// The request/event is never rearmed, so insertion followed by removal cannot
// erase invalidation. Existing mutator handles are covered by the kernel too.
// Every completion, including overflow, permanently invalidates the build.
type directoryWatch struct {
	path           string
	handle, event  syscall.Handle
	request        syscall.Overlapped
	output         [1024]uint32 // DWORD-aligned native buffer; overflow also refuses
	completedBytes uint32
	invalid        error
	released       bool
}

var watchKernel = syscall.NewLazyDLL("kernel32.dll")

func openDirectoryWatch(path string) (*directoryWatch, error) {
	return openDirectoryWatchBuffer(path, 4096)
}

func openDirectoryWatchBuffer(path string, bufferBytes uint32) (*directoryWatch, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	h, err := syscall.CreateFile(name, syscall.GENERIC_READ, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OVERLAPPED, 0)
	if err != nil {
		return nil, fmt.Errorf("open input directory watch %q: %w", path, err)
	}
	event, _, err := watchKernel.NewProc("CreateEventW").Call(0, 1, 0, 0)
	if event == 0 {
		syscall.CloseHandle(h)
		return nil, fmt.Errorf("create input watch event %q: %w", path, err)
	}
	w := &directoryWatch{path: path, handle: h, event: syscall.Handle(event)}
	w.request.HEvent = w.event
	var size uint32
	if bufferBytes > uint32(unsafe.Sizeof(w.output)) || bufferBytes < 4 || bufferBytes%4 != 0 {
		syscall.CloseHandle(w.event)
		syscall.CloseHandle(w.handle)
		return nil, fmt.Errorf("invalid native change-buffer size %d", bufferBytes)
	}
	err = syscall.ReadDirectoryChanges(h, (*byte)(unsafe.Pointer(&w.output[0])), bufferBytes, false, syscall.FILE_NOTIFY_CHANGE_FILE_NAME|syscall.FILE_NOTIFY_CHANGE_DIR_NAME, &size, &w.request, 0)
	if err != nil {
		syscall.CloseHandle(w.event)
		syscall.CloseHandle(w.handle)
		return nil, fmt.Errorf("input directory change guard unavailable %q: %w", path, err)
	}
	return w, nil
}

func (w *directoryWatch) check() error {
	if w.invalid != nil {
		return w.invalid
	}
	if w.released {
		return fmt.Errorf("input directory watch already released: %s", w.path)
	}
	state, err := syscall.WaitForSingleObject(w.event, 0)
	if err != nil || state != syscall.WAIT_TIMEOUT {
		child := ""
		if state == syscall.WAIT_OBJECT_0 && w.result(false) == nil && w.output[2] > 0 && w.output[2] <= uint32(unsafe.Sizeof(w.output))-12 && w.output[2]%2 == 0 {
			child = syscall.UTF16ToString(unsafe.Slice((*uint16)(unsafe.Pointer(&w.output[3])), int(w.output[2]/2)))
		}
		w.invalid = fmt.Errorf("input directory invalidated %q (child=%q event=%d): %v", w.path, child, state, err)
	}
	return w.invalid
}

func (w *directoryWatch) result(wait bool) error {
	var done uint32
	waitValue := uintptr(0)
	if wait {
		waitValue = 1
	}
	ok, _, err := watchKernel.NewProc("GetOverlappedResult").Call(uintptr(w.handle), uintptr(unsafe.Pointer(&w.request)), uintptr(unsafe.Pointer(&done)), waitValue)
	if ok == 0 {
		return err
	}
	w.completedBytes = done
	return nil
}

func (w *directoryWatch) close() error {
	if w.released {
		return nil
	}
	before := w.check()
	// Completion with ERROR_OPERATION_ABORTED is accepted only after our own
	// successful cancellation. A break racing cancellation completes normally
	// and permanently invalidates the build. No scan or recapture clears it.
	cancelErr := syscall.CancelIoEx(w.handle, &w.request)
	resultErr := w.result(true) // join before releasing OVERLAPPED/output/event memory
	if before == nil && !(cancelErr == nil && errors.Is(resultErr, syscall.ERROR_OPERATION_ABORTED)) {
		w.invalid = fmt.Errorf("input directory invalidated during release %q: cancel=%v completion=%v", w.path, cancelErr, resultErr)
	}
	eventErr := syscall.CloseHandle(w.event)
	handleErr := syscall.CloseHandle(w.handle)
	w.released = true
	runtime.KeepAlive(w)
	return errors.Join(w.invalid, eventErr, handleErr)
}
