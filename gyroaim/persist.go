package gyroaim

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Alia5/SISR/config"
)

// Saved is the on-disk form of the gyro aim settings changed from the UI.
type Saved struct {
	Enabled     bool    `json:"enabled"`
	Sensitivity float64 `json:"sensitivity"`
	PowerCurve  float64 `json:"powerCurve"`
	MinOutput   float64 `json:"minOutput"`
	Deadband    float64 `json:"deadband"`
	InvertX     bool    `json:"invertX"`
	InvertY     bool    `json:"invertY"`
}

// SavedPath is where UI changes are stored: %APPDATA%\SISR\gyroaim.json on Windows.
func SavedPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "SISR", "gyroaim.json"), nil
}

// SavedFrom copies the gyro fields out of the controller emulation config.
func SavedFrom(c *config.ControllerEmulation) Saved {
	return Saved{
		Enabled:     c.GyroAim,
		Sensitivity: c.GyroAimSensitivity,
		PowerCurve:  c.GyroAimPowerCurve,
		MinOutput:   c.GyroAimMinOutput,
		Deadband:    c.GyroAimDeadband,
		InvertX:     c.GyroAimInvertX,
		InvertY:     c.GyroAimInvertY,
	}
}

// ApplyTo writes the saved values into the controller emulation config.
func (s Saved) ApplyTo(c *config.ControllerEmulation) {
	c.GyroAim = s.Enabled
	c.GyroAimSensitivity = s.Sensitivity
	c.GyroAimPowerCurve = s.PowerCurve
	c.GyroAimMinOutput = s.MinOutput
	c.GyroAimDeadband = s.Deadband
	c.GyroAimInvertX = s.InvertX
	c.GyroAimInvertY = s.InvertY
}

// LoadInto applies previously saved UI settings on top of flags/defaults.
// A missing file is not an error.
func LoadInto(c *config.ControllerEmulation) error {
	path, err := SavedPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	s := SavedFrom(c) // fields missing from the file keep their current value
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	s.ApplyTo(c)
	return nil
}

// Save stores the current gyro settings so they survive a restart.
func Save(c *config.ControllerEmulation) error {
	path, err := SavedPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(SavedFrom(c), "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Status reports whether GTA V (GyroAim.asi) is publishing aim state and whether the
// player is aiming right now.
func Status() (gameConnected, aiming bool) {
	flags, ok := shm.read()
	return ok, ok && flags&flagFreeAiming != 0
}
