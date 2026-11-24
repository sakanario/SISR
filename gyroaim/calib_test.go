package gyroaim

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func degToRadSample(pitch, yaw float64) [3]float32 {
	return [3]float32{float32(pitch * math.Pi / 180), float32(yaw * math.Pi / 180), 0}
}

// feed sends n noisy samples around (pitch, yaw) deg/s, 375 Hz apart.
func feed(s *State, n int, pitch, yaw float64, aiming bool, start time.Time) time.Time {
	r := rand.New(rand.NewSource(1))
	t := start
	for i := 0; i < n; i++ {
		s.SetGyro(degToRadSample(pitch+r.NormFloat64()*0.3, yaw+r.NormFloat64()*0.3))
		s.UpdateCalibration(aiming, t)
		t = t.Add(time.Second / 375)
	}
	return t
}

func TestAutoCalibrationRemovesRestingBias(t *testing.T) {
	var s State
	feed(&s, 1500, -6, 1.5, false, time.Now()) // resting controller with a downward bias
	p, y, ok := s.Bias()
	if !ok || math.Abs(p+6) > 0.3 || math.Abs(y-1.5) > 0.3 {
		t.Fatalf("bias = %.2f/%.2f ok=%v, want ~-6/1.5", p, y, ok)
	}
	// With the bias removed, a resting controller must not move the stick.
	cfg := testCfg
	cfg.Deadband = 1
	s.SetGyro(degToRadSample(-6, 1.5))
	gx, gy := gyroToStick(cfg, s.yaw, s.pitch)
	if gx != 0 || gy != 0 {
		t.Fatalf("resting output = %d,%d, want 0,0", gx, gy)
	}
}

func TestAutoCalibrationSkippedWhileAiming(t *testing.T) {
	var s State
	feed(&s, 1500, -6, 0, true, time.Now())
	if _, _, ok := s.Bias(); ok {
		t.Fatal("bias learned while aiming")
	}
}

func TestAutoCalibrationIgnoresFastTurn(t *testing.T) {
	var s State
	feed(&s, 1500, 40, 0, false, time.Now()) // steady 40 deg/s turn is motion, not bias
	if _, _, ok := s.Bias(); ok {
		t.Fatal("learned a fast turn as bias")
	}
}

func TestManualCalibration(t *testing.T) {
	var s State
	start := time.Now()
	manual.mu.Lock()
	manual.until = start.Add(time.Second)
	manual.mu.Unlock()
	defer func() { manual.mu.Lock(); manual.until = time.Time{}; manual.mu.Unlock() }()

	end := feed(&s, 375, 3, -2, true, start) // collected even while aiming
	feed(&s, 1, 3, -2, true, end.Add(10*time.Millisecond))
	p, y, ok := s.Bias()
	if !ok || math.Abs(p-3) > 0.1 || math.Abs(y+2) > 0.1 {
		t.Fatalf("bias = %.2f/%.2f ok=%v, want ~3/-2", p, y, ok)
	}
}
