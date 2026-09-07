package broker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var kernel = windows.NewLazySystemDLL("kernel32.dll")

func nativeCall(name string, args ...uintptr) error {
	r, _, e := kernel.NewProc(name).Call(args...)
	if r == 0 {
		return fmt.Errorf("%s: %w", name, e)
	}
	return nil
}

func closeHandle(h *windows.Handle) error {
	if *h == 0 || *h == windows.InvalidHandle {
		return nil
	}
	err := windows.CloseHandle(*h)
	if err == nil {
		*h = 0
	}
	return err
}

func newJob() (windows.Handle, error) {
	h, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, err
	}
	var limit windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	limit.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(h, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limit)), uint32(unsafe.Sizeof(limit)))
	if err != nil {
		return h, err // caller owns even a partially configured Job
	}
	return h, nil
}

func activeProcesses(job windows.Handle) (uint32, error) {
	var accounting struct {
		User, Kernel, PeriodUser, PeriodKernel int64
		Faults, Total, Active, Terminated      uint32
	}
	err := windows.QueryInformationJobObject(job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil)
	return accounting.Active, err
}

func waitJob(ctx context.Context, job windows.Handle) error {
	for {
		n, err := activeProcesses(job)
		if err != nil || n == 0 {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func waitProcess(ctx context.Context, process windows.Handle) (uint32, error) {
	for {
		result, err := windows.WaitForSingleObject(process, 20)
		if err != nil {
			return 0, err
		}
		if result == windows.WAIT_OBJECT_0 {
			var exit uint32
			err = windows.GetExitCodeProcess(process, &exit)
			return exit, err
		}
		if result != uint32(windows.WAIT_TIMEOUT) {
			return 0, errors.New("unexpected process wait result")
		}
		if err = ctx.Err(); err != nil {
			return 0, err
		}
	}
}

type fileIdentity struct {
	Volume uint64
	ID     [16]byte
}

func identity(h windows.Handle) (fileIdentity, error) {
	// uint64 storage supplies native FileIdInfo alignment on 386 as well.
	var aligned [3]uint64
	b := unsafe.Slice((*byte)(unsafe.Pointer(&aligned[0])), 24)
	err := windows.GetFileInformationByHandleEx(h, windows.FileIdInfo, &b[0], 24)
	v := fileIdentity{Volume: binary.LittleEndian.Uint64(b)}
	copy(v.ID[:], b[8:])
	if err == nil && v.ID == [16]byte{} {
		err = errors.New("unavailable file identity")
	}
	return v, err
}

func openRelative(parent windows.Handle, name string, access, share, disposition, options uint32) (windows.Handle, error) {
	u, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, err
	}
	oa := windows.OBJECT_ATTRIBUTES{Length: uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})), RootDirectory: parent, ObjectName: u, Attributes: windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE}
	var iosb windows.IO_STATUS_BLOCK
	var h windows.Handle
	err = windows.NtCreateFile(&h, access|windows.SYNCHRONIZE, &oa, &iosb, nil, windows.FILE_ATTRIBUTE_NORMAL, share, disposition, options|windows.FILE_OPEN_REPARSE_POINT|windows.FILE_SYNCHRONOUS_IO_NONALERT, 0, 0)
	return h, err
}
