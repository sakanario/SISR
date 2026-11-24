// Package gyroaim exposes the GTA V gyro aim settings to the UI.
package gyroaim

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Alia5/SISR/cmd"
	ga "github.com/Alia5/SISR/gyroaim"
	"github.com/danielgtaylor/huma/v2"
)

type Settings struct {
	Enabled     bool    `json:"enabled"`
	Sensitivity float64 `json:"sensitivity" minimum:"0.1" maximum:"20" doc:"Multiplier; full stick at 360/sensitivity deg/s"`
	PowerCurve  float64 `json:"powerCurve" minimum:"0.1" maximum:"4" doc:"0.1 aggressive, 1 linear, 4 relaxed"`
	MinOutput   float64 `json:"minOutput" minimum:"0" maximum:"0.9" doc:"Smallest stick deflection past the deadband"`
	Deadband    float64 `json:"deadband" minimum:"0" maximum:"20" doc:"deg/s ignored as jitter"`
	InvertX     bool    `json:"invertX"`
	InvertY     bool    `json:"invertY"`
}

type SettingsResponse struct {
	Body Settings
}

type SettingsRequest struct {
	Body Settings
}

type StatusResponse struct {
	Body struct {
		GameConnected bool    `json:"gameConnected" doc:"GTA V with GyroAim.asi is running and updating"`
		Aiming        bool    `json:"aiming" doc:"Player is free-aiming right now"`
		Calibrating   bool    `json:"calibrating" doc:"A manual gyro calibration is collecting samples"`
		Calibrated    bool    `json:"calibrated" doc:"A gyro bias estimate exists"`
		BiasPitch     float64 `json:"biasPitch" doc:"Estimated gyro bias, deg/s"`
		BiasYaw       float64 `json:"biasYaw" doc:"Estimated gyro bias, deg/s"`
	}
}

const calibrationWindow = 2 * time.Second

func Register(a huma.API, c *cmd.SISRContext) {
	huma.Register(a, huma.Operation{
		Method: http.MethodGet,
		Path:   "/api/v1/gyroaim",
	}, func(ctx context.Context, _ *struct{}) (*SettingsResponse, error) {
		c.Config.Lock()
		defer c.Config.Unlock()
		return &SettingsResponse{Body: Settings(ga.SavedFrom(c.Config.ControllerEmulation))}, nil
	})

	huma.Register(a, huma.Operation{
		Method: http.MethodPut,
		Path:   "/api/v1/gyroaim",
	}, func(ctx context.Context, req *SettingsRequest) (*SettingsResponse, error) {
		c.Config.Lock()
		defer c.Config.Unlock()
		ga.Saved(req.Body).ApplyTo(c.Config.ControllerEmulation)
		if err := ga.Save(c.Config.ControllerEmulation); err != nil {
			slog.Error("Failed to save gyro aim settings", "error", err)
			return nil, huma.Error500InternalServerError("settings applied but could not be saved", err)
		}
		return &SettingsResponse{Body: Settings(ga.SavedFrom(c.Config.ControllerEmulation))}, nil
	})

	huma.Register(a, huma.Operation{
		Method: http.MethodGet,
		Path:   "/api/v1/gyroaim/status",
	}, func(ctx context.Context, _ *struct{}) (*StatusResponse, error) {
		resp := &StatusResponse{}
		resp.Body.GameConnected, resp.Body.Aiming = ga.Status()
		resp.Body.Calibrating = ga.Calibrating()
		resp.Body.BiasPitch, resp.Body.BiasYaw, resp.Body.Calibrated = ga.LastBias()
		return resp, nil
	})

	huma.Register(a, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/api/v1/gyroaim/calibrate",
		Description: "Average the gyro for 2 s and use it as bias; keep the controller still meanwhile",
	}, func(ctx context.Context, _ *struct{}) (*struct{}, error) {
		ga.RequestCalibration(calibrationWindow)
		return nil, nil
	})
}
