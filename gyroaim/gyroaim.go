// Package gyroaim turns the real controller's gyro into right-stick movement on the
// emulated Xbox 360 pad, but only while GTA V reports that the player is aiming.
// The aim state comes from GyroAim.asi via shared memory (see shm_windows.go).
package gyroaim

import (
	"math"
	"time"
)

// FullScaleAt1x is the angular speed (deg/s) that fully deflects the stick at sensitivity 1x.
// Sensitivity 3x therefore reaches full deflection at 120 deg/s.
const FullScaleAt1x = 360.0

// Settings mirrors the --gyro-aim-* options and the UI sliders.
type Settings struct {
	Enabled bool
	// Multiplier like Steam's "Gyro Sensitivity": full stick at FullScaleAt1x/Sensitivity deg/s.
	Sensitivity float64
	// Like Steam's "Joystick Power Curve": 0.1 extremely aggressive, 1 linear, 4 extremely relaxed.
	PowerCurve float64
	// Stick deflection (0..1) used for the slowest movement that passes the deadband;
	// lets small motions get past the game's own stick deadzone.
	MinOutput float64
	// Angular speed (deg/s) below which gyro input is ignored (hand jitter, drift).
	Deadband float64
	InvertX  bool
	InvertY  bool
}

// State is the per-device gyro state. Callers hold the device lock.
type State struct {
	rawPitch, rawYaw float64 // latest angular speed in deg/s (SDL axes: x = pitch, y = yaw)
	pitch, yaw       float64 // same, with the calibrated bias removed
	baseRX           int16   // right stick as delivered by Steam Input, XInput convention
	baseRY           int16
	active           bool // gyro contributed to the last output
	cal              calibration
}

// SetGyro stores the latest gyroscope sample from SDL (rad/s).
func (s *State) SetGyro(data [3]float32) {
	const radToDeg = 180.0 / math.Pi
	s.rawPitch = float64(data[0]) * radToDeg
	s.rawYaw = float64(data[1]) * radToDeg
	s.pitch, s.yaw = s.rawPitch-s.cal.biasPitch, s.rawYaw-s.cal.biasYaw
}

// UpdateCalibration feeds the latest sample into bias calibration and re-applies the bias.
func (s *State) UpdateCalibration(aiming bool, now time.Time) {
	s.cal.observe(s.rawPitch, s.rawYaw, aiming, now)
	s.pitch, s.yaw = s.rawPitch-s.cal.biasPitch, s.rawYaw-s.cal.biasYaw
}

// Bias returns the current gyro bias estimate in deg/s.
func (s *State) Bias() (pitch, yaw float64, ok bool) {
	return s.cal.biasPitch, s.cal.biasYaw, s.cal.haveBias
}

// SetBaseStick stores the right stick coming from Steam Input before gyro is added.
func (s *State) SetBaseStick(rx, ry int16) {
	s.baseRX, s.baseRY = rx, ry
}

// WasActive reports whether gyro changed the last output (so the stick must be reset
// once aiming stops).
func (s *State) WasActive() bool { return s.active }

// Aiming reports whether GTA V currently says the player is free-aiming
// (on foot with ZL or drive-by with L).
func Aiming() bool {
	flags, ok := shm.read()
	return ok && flags&flagFreeAiming != 0
}

// RightStick returns the right stick to send: Steam's stick plus gyro while aiming.
func (s *State) RightStick(cfg Settings, aiming bool) (rx, ry int16) {
	if !cfg.Enabled || !aiming {
		s.active = false
		return s.baseRX, s.baseRY
	}
	gx, gy := gyroToStick(cfg, s.yaw, s.pitch)
	s.active = true
	return addAxis(s.baseRX, gx), addAxis(s.baseRY, gy)
}

// gyroToStick maps angular speed to stick deflection in XInput units.
// Turning left (positive yaw) pushes the stick left; tilting up (positive pitch) pushes it up.
func gyroToStick(cfg Settings, yaw, pitch float64) (int16, int16) {
	x, y := -yaw, pitch
	if cfg.InvertX {
		x = -x
	}
	if cfg.InvertY {
		y = -y
	}
	speed := math.Hypot(x, y)
	if speed <= cfg.Deadband || cfg.Sensitivity <= 0 {
		return 0, 0
	}
	// Scale on the vector length so diagonal motion keeps its direction.
	fullScale := FullScaleAt1x / cfg.Sensitivity
	t := math.Min(1, (speed-cfg.Deadband)/math.Max(1e-6, fullScale-cfg.Deadband))
	if cfg.PowerCurve > 0 && cfg.PowerCurve != 1 {
		t = math.Pow(t, cfg.PowerCurve)
	}
	minOut := math.Max(0, math.Min(1, cfg.MinOutput))
	deflection := minOut + (1-minOut)*t
	scale := deflection * math.MaxInt16 / speed
	return int16(math.Round(x * scale)), int16(math.Round(y * scale))
}

func addAxis(a, b int16) int16 {
	return int16(max(math.MinInt16, min(math.MaxInt16, int32(a)+int32(b))))
}

// Close releases the shared memory mapping.
func Close() { shm.close() }
