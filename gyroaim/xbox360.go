package gyroaim

import (
	"encoding"
	"time"

	"github.com/Alia5/SISR/config"
	"github.com/Alia5/VIIPER/device/xbox360"
)

// FromConfig builds Settings from the controller emulation options.
func FromConfig(c *config.ControllerEmulation) Settings {
	return Settings{
		Enabled:     c.GyroAim,
		Sensitivity: c.GyroAimSensitivity,
		PowerCurve:  c.GyroAimPowerCurve,
		MinOutput:   c.GyroAimMinOutput,
		Deadband:    c.GyroAimDeadband,
		InvertX:     c.GyroAimInvertX,
		InvertY:     c.GyroAimInvertY,
	}
}

// ApplyXbox360 writes the right stick (Steam stick + gyro while aiming) into an Xbox 360
// VIIPER state. fromSteam=true means the state's RX/RY were just filled from Steam Input
// and become the new base. Returns true if the state changed and should be sent.
func ApplyXbox360(st *State, cfg Settings, state *encoding.BinaryMarshaler, fromSteam bool) bool {
	s, ok := (*state).(*xbox360.InputState)
	if !ok || s == nil {
		return false
	}
	aiming := cfg.Enabled && Aiming()
	if fromSteam {
		st.SetBaseStick(s.RX, s.RY)
	} else {
		st.UpdateCalibration(aiming, time.Now())
		if !cfg.Enabled {
			diagRecord(false, false, false, st, s.RX, s.RY)
			return false
		}
	}
	if !fromSteam && !aiming && !st.WasActive() {
		diagRecord(false, false, false, st, s.RX, s.RY)
		return false // gyro idle, nothing to update between Steam reports
	}
	rx, ry := st.RightStick(cfg, aiming)
	changed := rx != s.RX || ry != s.RY
	s.RX, s.RY = rx, ry
	send := changed || fromSteam
	diagRecord(fromSteam, aiming, send, st, rx, ry)
	return send
}
