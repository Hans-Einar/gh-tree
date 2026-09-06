package broker

import (
	"context"
	"fmt"
	"golang.org/x/sys/unix"
	"runtime"
	"syscall"
	"unsafe"
)

func census(ctx context.Context) ([]processFact, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	rows, err := boundedDarwinTable(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) > maxCensusRecords {
		return nil, ErrCensus
	}
	result := make([]processFact, 0, len(rows))
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		pid := int(row.Proc.P_pid)
		if pid == 0 {
			continue
		}
		if pid < 0 {
			return nil, ErrCensus
		}
		sid, err := unix.Getsid(pid)
		if err == unix.ESRCH {
			continue
		}
		if err != nil {
			return nil, err
		}
		verified, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
		if err == unix.ESRCH {
			continue
		}
		if err != nil {
			return nil, err
		}
		if verified.Proc.P_pid != row.Proc.P_pid || verified.Proc.P_starttime != row.Proc.P_starttime || verified.Eproc.Pgid != row.Eproc.Pgid {
			return nil, ErrCensus
		}
		// Darwin proc.h: SIDL1/SRUN2/SSLEEP3/SSTOP4/SZOMB5. Eproc.Sess
		// is a kernel pointer-shaped field and is never interpreted as SID.
		state := row.Proc.P_stat
		if state < 1 || state > 5 {
			return nil, ErrCensus
		}
		result = append(result, processFact{pid: pid, parent: int(row.Eproc.Ppid), group: int(row.Eproc.Pgid), session: sid, identity: fmt.Sprintf("%d:%d", row.Proc.P_starttime.Sec, row.Proc.P_starttime.Usec), live: state != 5, stopped: state == 4})
	}
	return result, nil
}

// x/sys SysctlKinfoProcSlice allocates the reported table before a caller can
// check its size and retries ENOMEM without a limit. The same native sysctl MIB
// with the pinned x/sys KinfoProc layout permits a bounded acquisition here.
func boundedDarwinTable(ctx context.Context) ([]unix.KinfoProc, error) {
	// Apple bsd/sys/sysctl.h: KERN_PROC=14 and KERN_PROC_ALL=0.
	// https://github.com/apple-oss-distributions/xnu/blob/main/bsd/sys/sysctl.h
	mib := [3]int32{unix.CTL_KERN, 14, 0}
	for attempt := 0; attempt < 3; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var length uintptr
		_, _, errno := syscall.Syscall6(unix.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), uintptr(len(mib)), 0, uintptr(unsafe.Pointer(&length)), 0, 0)
		runtime.KeepAlive(mib)
		if errno != 0 {
			return nil, errno
		}
		if length == 0 || length%unix.SizeofKinfoProc != 0 || length/unix.SizeofKinfoProc > maxCensusRecords {
			return nil, ErrCensus
		}
		rows := make([]unix.KinfoProc, length/unix.SizeofKinfoProc)
		capacity := length
		_, _, errno = syscall.Syscall6(unix.SYS___SYSCTL, uintptr(unsafe.Pointer(&mib[0])), uintptr(len(mib)), uintptr(unsafe.Pointer(&rows[0])), uintptr(unsafe.Pointer(&length)), 0, 0)
		runtime.KeepAlive(mib)
		runtime.KeepAlive(rows)
		if errno == syscall.ENOMEM {
			continue
		}
		if errno != 0 {
			return nil, errno
		}
		if length > capacity || length%unix.SizeofKinfoProc != 0 {
			return nil, ErrCensus
		}
		return rows[:length/unix.SizeofKinfoProc], nil
	}
	return nil, ErrCensus
}
