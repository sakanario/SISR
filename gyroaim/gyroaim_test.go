package gyroaim

import (
	"math"
	"testing"
)

var testCfg = Settings{Enabled: true, Sensitivity: 3, PowerCurve: 1, MinOutput: 0.2, Deadband: 1}

func TestPowerCurve(t *testing.T) {
	cfg := testCfg
	cfg.MinOutput = 0
	half := -(1 + (FullScaleAt1x/cfg.Sensitivity-1)/2) // speed giving t = 0.5
	linear, _ := gyroToStick(cfg, half, 0)
	cfg.PowerCurve = 0.4
	aggressive, _ := gyroToStick(cfg, half, 0)
	cfg.PowerCurve = 2
	relaxed, _ := gyroToStick(cfg, half, 0)
	if !(aggressive > linear && linear > relaxed) {
		t.Fatalf("aggressive=%d linear=%d relaxed=%d", aggressive, linear, relaxed)
	}
	if got := float64(linear) / math.MaxInt16; math.Abs(got-0.5) > 0.01 {
		t.Fatalf("linear half = %.3f", got)
	}
}

func TestHigherSensitivityDeflectsMore(t *testing.T) {
	lo, hi := testCfg, testCfg
	hi.Sensitivity = 6
	a, _ := gyroToStick(lo, -30, 0)
	b, _ := gyroToStick(hi, -30, 0)
	if b <= a {
		t.Fatalf("sens 3 -> %d, sens 6 -> %d", a, b)
	}
}

func TestNotAimingPassesSteamStickThrough(t *testing.T) {
	var s State
	s.SetBaseStick(1000, -2000)
	s.SetGyro([3]float32{1, 1, 0})
	rx, ry := s.RightStick(testCfg, false)
	if rx != 1000 || ry != -2000 || s.WasActive() {
		t.Fatalf("got %d,%d active=%v", rx, ry, s.WasActive())
	}
}

func TestDeadbandIgnoresJitter(t *testing.T) {
	x, y := gyroToStick(testCfg, 0.5, 0.5)
	if x != 0 || y != 0 {
		t.Fatalf("jitter produced %d,%d", x, y)
	}
}

func TestTurnLeftPushesStickLeft(t *testing.T) {
	x, y := gyroToStick(testCfg, 60, 0) // positive yaw = turning left
	if x >= 0 || y != 0 {
		t.Fatalf("got %d,%d", x, y)
	}
}

func TestTiltUpPushesStickUp(t *testing.T) {
	_, y := gyroToStick(testCfg, 0, 60)
	if y <= 0 {
		t.Fatalf("got y=%d", y)
	}
}

func TestMinOutputAndSaturation(t *testing.T) {
	x, _ := gyroToStick(testCfg, -1.01, 0) // just past the deadband
	if got := float64(x) / math.MaxInt16; got < 0.19 || got > 0.21 {
		t.Fatalf("min output %.3f", got)
	}
	x, _ = gyroToStick(testCfg, -500, 0) // far past MaxSpeed
	if x != math.MaxInt16 {
		t.Fatalf("saturation %d", x)
	}
}

func TestAimingAddsToSteamStickAndClamps(t *testing.T) {
	var s State
	s.SetBaseStick(30000, 0)
	s.SetGyro([3]float32{0, float32(-200 * math.Pi / 180), 0}) // fast right turn
	rx, _ := s.RightStick(testCfg, true)
	if rx != math.MaxInt16 || !s.WasActive() {
		t.Fatalf("rx=%d active=%v", rx, s.WasActive())
	}
}
