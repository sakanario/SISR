package gyroaim

import (
	"math"
	"sync"
	"time"
)

// Gyro bias calibration.
//
// A gyro never reads exactly 0 at rest; a leftover of a few deg/s ("bias") turns into a
// constant camera drift while aiming. The bias is estimated two ways:
//   - automatically, whenever the controller is held still and the player is not aiming;
//   - manually, by averaging all samples during a short window requested from the UI
//     (controller resting on a table).
const (
	shortAlpha     = 0.05 // EMA used to detect stillness (~50 ms at ~375 samples/s)
	stillThreshold = 2.5  // deg/s; samples closer than this to the short mean count as "still"
	stillSamples   = 300  // ~0.8 s of stillness before auto calibration trusts the mean
	biasAlpha      = 0.01 // how fast the bias follows once still
	maxBias        = 15.0 // deg/s; larger "biases" are real motion, never learned
)

type calibration struct {
	biasPitch, biasYaw float64
	meanPitch, meanYaw float64
	still              int
	haveBias           bool

	manualFor  time.Time // which manual request is being collected
	sumP, sumY float64
	n          int
}

var manual struct {
	mu    sync.Mutex
	until time.Time
}

// RequestCalibration averages the gyro of every connected controller for d and uses the
// result as its bias. The controller should rest still on a surface meanwhile.
func RequestCalibration(d time.Duration) {
	manual.mu.Lock()
	manual.until = time.Now().Add(d)
	manual.mu.Unlock()
}

// Calibrating reports whether a manual calibration window is running.
func Calibrating() bool {
	manual.mu.Lock()
	defer manual.mu.Unlock()
	return time.Now().Before(manual.until)
}

// observe feeds one raw gyro sample (deg/s) into the calibration.
func (c *calibration) observe(rawPitch, rawYaw float64, aiming bool, now time.Time) {
	manual.mu.Lock()
	until := manual.until
	manual.mu.Unlock()

	// Manual window: collect, then commit once the window has passed.
	if !until.IsZero() {
		if now.Before(until) {
			if !c.manualFor.Equal(until) {
				c.manualFor, c.sumP, c.sumY, c.n = until, 0, 0, 0
			}
			c.sumP += rawPitch
			c.sumY += rawYaw
			c.n++
			return
		}
		if c.manualFor.Equal(until) && c.n > 0 {
			c.biasPitch, c.biasYaw = c.sumP/float64(c.n), c.sumY/float64(c.n)
			c.haveBias = true
			c.n = 0
		}
	}

	// Automatic: track a short mean and learn it as bias after a still period.
	c.meanPitch += shortAlpha * (rawPitch - c.meanPitch)
	c.meanYaw += shortAlpha * (rawYaw - c.meanYaw)
	if math.Abs(rawPitch-c.meanPitch) < stillThreshold && math.Abs(rawYaw-c.meanYaw) < stillThreshold {
		c.still++
	} else {
		c.still = 0
	}
	if aiming || c.still < stillSamples ||
		math.Abs(c.meanPitch) > maxBias || math.Abs(c.meanYaw) > maxBias {
		return
	}
	if !c.haveBias {
		c.biasPitch, c.biasYaw, c.haveBias = c.meanPitch, c.meanYaw, true
		return
	}
	c.biasPitch += biasAlpha * (c.meanPitch - c.biasPitch)
	c.biasYaw += biasAlpha * (c.meanYaw - c.biasYaw)
}
