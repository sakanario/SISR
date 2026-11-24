//go:build !windows

package gyroaim

const (
	flagFreeAiming = 1 << 0
	flagAimCam     = 1 << 1
	flagInVehicle  = 1 << 2
)

type reader struct{}

var shm reader

// GTA V only runs on Windows; elsewhere aim state is never available.
func (r *reader) read() (uint32, bool) { return 0, false }
func (r *reader) close()               {}
