package gyroaim

import (
	"log/slog"
	"math"
	"sync"
	"time"
)

// diag collects counters about the gyro aim pipeline and logs a summary every 2 s
// while GTA V is connected, so problems can be located from the SISR log file:
// no sensor samples (SDL/controller), not aiming (mod/shared memory), or no output change.
type diag struct {
	mu            sync.Mutex
	rawSensor     int // every SDL sensor event, before any filtering
	gyroSamples   int
	steamUpdates  int
	aimingSamples int
	sends         int
	lastYaw       float64
	lastPitch     float64
	lastRX        int16
	lastRY        int16
	biasPitch     float64
	biasYaw       float64
	haveBias      bool
}

var (
	stats     diag
	startOnce sync.Once
)

func startDiag() {
	startOnce.Do(func() {
		go func() {
			wasConnected := false
			for range time.Tick(2 * time.Second) {
				connected, aiming := Status()
				if connected != wasConnected {
					slog.Info("GyroAim: GTA V connection changed", "connected", connected)
					wasConnected = connected
				}
				stats.mu.Lock()
				if connected {
					slog.Info("GyroAim: 2s summary",
						"aimingNow", aiming,
						"rawSensorEvents", stats.rawSensor,
						"gyroSamples", stats.gyroSamples,
						"steamUpdates", stats.steamUpdates,
						"samplesWhileAiming", stats.aimingSamples,
						"stateSends", stats.sends,
						"lastYawDps", int(stats.lastYaw),
						"lastPitchDps", int(stats.lastPitch),
						"lastRX", stats.lastRX,
						"lastRY", stats.lastRY,
						"biasPitchDps", math.Round(stats.biasPitch*100)/100,
						"biasYawDps", math.Round(stats.biasYaw*100)/100,
						"calibrated", stats.haveBias,
					)
				}
				stats.rawSensor, stats.gyroSamples, stats.steamUpdates, stats.aimingSamples, stats.sends = 0, 0, 0, 0, 0
				stats.mu.Unlock()
			}
		}()
	})
}

func diagRecord(fromSteam, aiming, sent bool, st *State, rx, ry int16) {
	startDiag()
	stats.mu.Lock()
	defer stats.mu.Unlock()
	if fromSteam {
		stats.steamUpdates++
	} else {
		stats.gyroSamples++
		stats.lastYaw, stats.lastPitch = st.yaw, st.pitch
	}
	if aiming {
		stats.aimingSamples++
	}
	if sent {
		stats.sends++
	}
	stats.lastRX, stats.lastRY = rx, ry
	stats.biasPitch, stats.biasYaw, stats.haveBias = st.Bias()
}

// DiagRawSensor counts an SDL sensor event as soon as SISR receives it.
func DiagRawSensor() {
	startDiag()
	stats.mu.Lock()
	stats.rawSensor++
	stats.mu.Unlock()
}

// LastBias returns the most recent gyro bias estimate (deg/s) seen by the pipeline.
func LastBias() (pitch, yaw float64, ok bool) {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	return stats.biasPitch, stats.biasYaw, stats.haveBias
}
