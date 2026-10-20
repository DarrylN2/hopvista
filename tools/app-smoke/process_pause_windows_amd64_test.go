//go:build windows && amd64

package appsmoke

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

var processDLL = syscall.NewLazyDLL("ntdll.dll")
var suspendProcess = processDLL.NewProc("NtSuspendProcess")
var resumeProcess = processDLL.NewProc("NtResumeProcess")
var kernelDLL = syscall.NewLazyDLL("kernel32.dll")
var createJobObject = kernelDLL.NewProc("CreateJobObjectW")
var setJobInformation = kernelDLL.NewProc("SetInformationJobObject")
var assignProcessToJob = kernelDLL.NewProc("AssignProcessToJobObject")

// A Go test timeout terminates the test process without running t.Cleanup.
// Owning the child in a kill-on-close Windows job makes the kernel remove a
// paused child when that happens. The layout is JOBOBJECT_EXTENDED_LIMIT_INFORMATION
// on Windows x64; LimitFlags is at byte 16 and the structure is 144 bytes.
func ownDisposablePackagedProcess(t *testing.T, p *packagedProcess) {
	t.Helper()
	job, _, err := createJobObject.Call(0, 0)
	if job == 0 {
		t.Fatalf("create disposable child job: %v", err)
	}
	// The job is registered after the existing packaged-process cleanup, so
	// stop the child first while its handle is still valid. Abrupt test death
	// instead closes the job in the kernel and kills the child automatically.
	t.Cleanup(func() {
		p.stop(t)
		_ = syscall.CloseHandle(syscall.Handle(job))
	})
	var limits [144]byte
	binary.LittleEndian.PutUint32(limits[16:20], 0x2000) // JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	set, _, err := setJobInformation.Call(job, 9, uintptr(unsafe.Pointer(&limits[0])), uintptr(len(limits)))
	if set == 0 {
		t.Fatalf("set disposable child job limits: %v", err)
	}
	const assignAccess = 0x0101 // PROCESS_SET_QUOTA | PROCESS_TERMINATE
	process, err := syscall.OpenProcess(assignAccess, false, uint32(p.cmd.Process.Pid))
	if err != nil {
		t.Fatalf("open disposable child for job: %v", err)
	}
	defer syscall.CloseHandle(process)
	assigned, _, err := assignProcessToJob.Call(job, uintptr(process))
	if assigned == 0 {
		t.Fatalf("assign disposable child to kill-on-close job: %v", err)
	}
}

// Pause only the disposable child executable launched by this test. NTSTATUS
// errors are reported rather than silently turning a failed pause into timing
// evidence. Always resume before process cleanup, even on a test failure.
func pausePackagedUntil(t *testing.T, p *packagedProcess, resumeAt time.Time) {
	t.Helper()
	const processSuspendResume = 0x0800
	handle, err := syscall.OpenProcess(processSuspendResume, false, uint32(p.cmd.Process.Pid))
	if err != nil {
		t.Fatalf("open disposable packaged process for pause: %v", err)
	}
	defer syscall.CloseHandle(handle)
	call := func(proc *syscall.LazyProc) error {
		status, _, _ := proc.Call(uintptr(handle))
		if status != 0 {
			return fmt.Errorf("%s returned NTSTATUS %#x", proc.Name, status)
		}
		return nil
	}
	if err := call(suspendProcess); err != nil {
		t.Fatal(err)
	}
	pausedAt := time.Now().UTC()
	defer func() {
		if err := call(resumeProcess); err != nil {
			t.Errorf("resume disposable packaged process: %v", err)
			return
		}
		t.Logf("harness-only packaged process pause: suspended=%s resumed=%s planned=%s; native stall was not induced by the network", pausedAt.Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano), resumeAt.Format(time.RFC3339Nano))
	}()
	if wait := time.Until(resumeAt); wait > 0 {
		time.Sleep(wait)
	}
}
