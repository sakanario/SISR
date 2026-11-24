//go:build windows

package gyroaim

import (
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Layout of "Local\GyroAimState", written by GyroAim.asi inside GTA V
// (see gyro-aim-mod/src/gyroaim_ipc.h; both sides must agree).
type sharedState struct {
	Magic     uint32
	Version   uint32
	Flags     uint32
	Frame     uint32
	UpdatedMs uint64
	WriterPid uint32
	Reserved  uint32
}

const (
	mappingName  = `Local\GyroAimState`
	stateMagic   = 0x4F525947 // "GYRO"
	stateVersion = 1
	staleMs      = 250
	retryEvery   = time.Second

	flagFreeAiming = 1 << 0
	flagAimCam     = 1 << 1
	flagInVehicle  = 1 << 2
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procOpenFileMappingW = kernel32.NewProc("OpenFileMappingW")
	procGetTickCount64   = kernel32.NewProc("GetTickCount64")
)

type reader struct {
	mu      sync.Mutex
	handle  windows.Handle
	view    *sharedState
	lastTry time.Time
}

var shm reader

func tickCount64() uint64 {
	r, _, _ := procGetTickCount64.Call()
	return uint64(r)
}

// open maps the shared memory if the game has created it. Retries at most once per second.
func (r *reader) open() bool {
	if r.view != nil {
		return true
	}
	if time.Since(r.lastTry) < retryEvery {
		return false
	}
	r.lastTry = time.Now()

	name, _ := windows.UTF16PtrFromString(mappingName)
	h, _, _ := procOpenFileMappingW.Call(uintptr(windows.FILE_MAP_READ), 0, uintptr(unsafe.Pointer(name)))
	if h == 0 {
		return false
	}
	addr, err := windows.MapViewOfFile(windows.Handle(h), windows.FILE_MAP_READ, 0, 0, unsafe.Sizeof(sharedState{}))
	if err != nil || addr == 0 {
		_ = windows.CloseHandle(windows.Handle(h))
		return false
	}
	r.handle = windows.Handle(h)
	r.view = (*sharedState)(unsafe.Pointer(addr)) //nolint:govet // pointer into a mapped view
	return true
}

// read returns the current game flags, or ok=false if the game is not running,
// not updating (paused/loading/minimised) or the layout does not match.
func (r *reader) read() (flags uint32, ok bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.open() {
		return 0, false
	}
	v := r.view
	if v.Magic != stateMagic || v.Version != stateVersion {
		return 0, false
	}
	updated := v.UpdatedMs
	if updated == 0 || tickCount64()-updated > staleMs {
		return 0, false
	}
	return v.Flags, true
}

func (r *reader) close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.view != nil {
		_ = windows.UnmapViewOfFile(uintptr(unsafe.Pointer(r.view)))
		r.view = nil
	}
	if r.handle != 0 {
		_ = windows.CloseHandle(r.handle)
		r.handle = 0
	}
}
