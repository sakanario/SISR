package config

type UpdateNotify string

const (
	UpdateNotifyNone       UpdateNotify = "none"
	UpdateNotifyStable     UpdateNotify = "stable"
	UpdateNotifyPrerelease UpdateNotify = "prerelease"
	UpdateNotifyAlways     UpdateNotify = "always"
)

type Log struct {
	Level string `aliases:"l" default:"info" help:"Log level: trace, debug, info, warn, error" env:"SISR_LOG_LEVEL"`
	File  string `aliases:"lf" help:"Log file path" env:"SISR_LOG_FILE"`
}

type Global struct {
	ConfigPath   string `help:"Path to configuration file (json|yaml|toml)" name:"config" env:"SISR_CONFIG"`
	Log          `embed:"" prefix:"log."`
	PlatformOpts `embed:""`
	Marker       bool `help:"dummy, used for Steam Input layout" hidden:""`
}

type API struct {
	ListenAddress   string `default:"localhost:0" help:"API listen address; by default will pick random free port and listen on localhost only" env:"SISR_API_LISTEN_ADDRESS"`
	CORSOrigins     string `help:"CORS allowed origins" default:"https://steamloopback.host,http://steamloopback.host" env:"SISR_API_CORS_ORIGINS"`
	FrontendAddress string `help:"Frontend address (optional)" hidden:""`
}

type Steam struct {
	InstallDir         string `help:"Steam installation directory (optional, will attempt to auto-detect if not set)" env:"SISR_STEAM_INSTALL_DIR"`
	UserID             uint32 `help:"Active Steam user ID (optional, will attempt to auto-detect if not set)" env:"SISR_STEAM_USER_ID"`
	CEFRemoteDebugPort uint16 `help:"CEF remote debugging port (optional, will attempt to auto-detect if not set)" env:"SISR_STEAM_CEF_REMOTE_DEBUG_PORT"`
}

type RunMisc struct {
	NoSteam       bool `default:"false" help:"Run in no-Steam mode" env:"SISR_NO_STEAM"`
	InitialLaunch bool `default:"false" hidden:""`
}

type AutoUpdate struct {
	UpdateNotify UpdateNotify `default:"stable" enum:"none,stable,prerelease,always" help:"Update notification level" env:"SISR_UPDATE_NOTIFY"`
}

type Window struct {
	MaxFPS     uint32 `default:"60" help:"Maximim FPS for SteamOverlay/UI (Does not affect inputs)" env:"SISR_MAX_FPS"`
	Fullscreen bool   `default:"true" aliases:"f" help:"Fullscreen overlay"  env:"SISR_FULLSCREEN"`
	Show       bool   `default:"false" aliases:"w" help:"Shows window on startup; when used w/ fullscreen=true: will enable Steam Overlay; when used with fullscreen=false: will show UI; when false SISR will only show up in system tray"   env:"SISR_SHOW_WINDOW"`
}

type ControllerEmulation struct {
	DefaultControllerType string `default:"xbox360" aliases:"ct" enum:"xbox360,dualshock4,dualsense,dualsenseedge,ns2pro" help:"Default controller type for emulation" env:"SISR_DEFAULT_CONTROLLER_TYPE"`

	GyroPassthrough       bool `default:"true" aliases:"gp" help:"Enable gyro passthrough for supported controllers" env:"SISR_GYRO_PASSTHROUGH"`
	TouchpadPassthrough   bool `default:"true" aliases:"tp" help:"Enable touchpad passthrough for supported controllers" env:"SISR_TOUCHPAD_PASSTHROUGH"`
	BackButtonPassthrough bool `default:"false" aliases:"bbp" help:"Enable back button passthrough for supported controllers (e.g. DualSense Edge)" env:"SISR_BACK_BUTTON_PASSTHROUGH"`

	AllowSteamDesktopLayout bool `default:"false" help:"Allow/Use Steam's desktop configuration for emulated controllers" env:"SISR_ALLOW_STEAM_DESKTOP_LAYOUT"`
	ImmediateSensorUpdates  bool `default:"true" help:"Immediately send sensor updates to VIIPER instead of waiting for the next input report; may reduce latency at the cost of increased CPU usage" env:"SISR_IMMEDIATE_SENSOR_UPDATES"`

	// GTA V gyro aim (needs GyroAim.asi in the game folder). Xbox 360 emulation only.
	GyroAim            bool    `default:"true" help:"Add gyro to the right stick while GTA V (GyroAim.asi) reports the player is aiming" env:"SISR_GYRO_AIM"`
	GyroAimSensitivity float64 `default:"3" help:"Gyro aim: sensitivity multiplier (full stick at 360/sensitivity deg/s)" env:"SISR_GYRO_AIM_SENSITIVITY"`
	GyroAimPowerCurve  float64 `default:"1" help:"Gyro aim: joystick power curve; 0.1 aggressive, 1 linear, 4 relaxed" env:"SISR_GYRO_AIM_POWER_CURVE"`
	GyroAimMinOutput   float64 `default:"0.15" help:"Gyro aim: smallest stick deflection (0-1) for movement past the deadband, to get past the game's stick deadzone" env:"SISR_GYRO_AIM_MIN_OUTPUT"`
	GyroAimDeadband    float64 `default:"1" help:"Gyro aim: angular speed in deg/s below which gyro is ignored (jitter/drift)" env:"SISR_GYRO_AIM_DEADBAND"`
	GyroAimInvertX     bool    `default:"false" help:"Gyro aim: invert horizontal gyro" env:"SISR_GYRO_AIM_INVERT_X"`
	GyroAimInvertY     bool    `default:"false" help:"Gyro aim: invert vertical gyro" env:"SISR_GYRO_AIM_INVERT_Y"`
}

type Viiper struct {
	Address  string `default:"localhost:3242" aliases:"va" help:"VIIPER server address" env:"SISR_VIIPER_ADDRESS"`
	Password string `aliases:"vp" help:"VIIPER server password" env:"SISR_VIIPER_PASSWORD"`
}

type KbMEmuation struct {
	KeyboardMouseEmulation bool `default:"false" aliases:"kbm" help:"Forward Keyboard/Mouse when running over a network" env:"SISR_KBM_EMULATION"`
}
