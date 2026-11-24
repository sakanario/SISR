<img src="docs/SISR.svg" align="right" width="128"/>
<br />

[![Build Status](https://github.com/alia5/SISR/actions/workflows/snapshots.yml/badge.svg)](https://github.com/alia5/SISR/actions/workflows/snapshots.yml)
[![License: GPL-3.0](https://img.shields.io/github/license/alia5/SISR)](https://github.com/alia5/SISR/blob/main/LICENSE.txt)
[![Release](https://img.shields.io/github/v/release/alia5/SISR?include_prereleases&sort=semver)](https://github.com/alia5/SISR/releases)
[![Issues](https://img.shields.io/github/issues/alia5/SISR)](https://github.com/alia5/SISR/issues)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/alia5/SISR/pulls)
[![Downloads](https://img.shields.io/github/downloads/alia5/SISR/total?logo=github)](https://github.com/alia5/SISR/releases)
[![Discord](https://img.shields.io/discord/368823110817808384?logo=discord&logoColor=white&label=Discord&color=%23535fe5
)](https://discord.gg/hs34MtcHJY)

> [!NOTE]
> **This is a fork of [Alia5/SISR](https://github.com/Alia5/SISR) that adds GTA V gyro aiming.**
> The rest of this README is the upstream one. For SISR itself, releases and support, use the upstream repo.

## GTA V gyro aim (this fork)

When emulating an **Xbox 360** controller, this fork adds your real controller's gyro (e.g. Switch Pro Controller) to the right stick, **only while GTA V says you are aiming**. That works on foot and when aiming from a vehicle (LB drive-by), which Steam's gyro activators can't tell apart.

The aim state comes from **GyroAim.asi**, a ScriptHookV plugin: [sakanario/gta5-gyro-aim-mod](https://github.com/sakanario/gta5-gyro-aim-mod). It publishes the aim state to the shared memory `Local\GyroAimState`, which this fork reads.

```
GTA V ──(ScriptHookV)──> GyroAim.asi ──(shared memory)──> SISR fork ──> virtual Xbox 360 pad ──> GTA V
```

**What's added**

- **Gyro Aim card** in the SISR UI: sensitivity (full stick at `360 / sensitivity` °/s), joystick power curve, minimum stick output, deadband, invert X/Y and gyro calibration, plus live "game connected / aiming" status. Changes apply live and are saved to `%APPDATA%\SISR\gyroaim.json`.
- **Gyro bias calibration:** automatic while the controller is still and you are not aiming, or manual (2 s) from the card. This stops the aim slowly drifting.
- **Fix for an upstream bug:** the VIIPER state queue dropped the newest state when full, which could leave the game stuck on an old input (e.g. running forever).
- API: `GET/PUT /api/v1/gyroaim`, `GET /api/v1/gyroaim/status`, `POST /api/v1/gyroaim/calibrate`, and `SISR_GYRO_AIM_*` environment variables for the defaults.

**Setup**

1. Install ScriptHookV and `GyroAim.asi` in GTA V Enhanced, following the [mod README](https://github.com/sakanario/gta5-gyro-aim-mod#install). Story Mode only.
2. Steam → Settings → Controller: *Enable Steam Input for Switch Pro controllers* **ON**, *Enable Steam Input for Xbox controllers* **OFF**. With the Xbox option on, Steam wraps SISR's virtual pad and input can drop after a few minutes.
3. In SISR's controller layout in Steam, set **Gyro Behavior = None**, because this fork applies the gyro itself.
4. Start SISR with Xbox 360 emulation (the default). `run-gyroaim.ps1` starts `dist\SISR.exe` with a debug log in `dist\logs`.
5. Start GTA V, open the **Gyro Aim** card in SISR and tune it. If the aim drifts while you hold still, press *Calibrate* with the controller on a table.

**Building** is the same as upstream (Go with cgo + SDL3, Svelte UI). Build only the command, because `go build ./...` fails on an unused package:

```powershell
go build -o dist/SISR.exe ./cmd/sisr
```

---

# SISR ✂️

**S**team **I**nput **S**ystem **R**edirector

SISR (pronounced "scissor") redirects Steam Input configurations to the system level (localhost or network).  

SISR takes controllers it receives from Steam (via Steam Input)
and forwards them as emulated but more compatible controllers (indistinguishable from real hardware) to the OS.

This approach makes SISR unique for controller rebinding tools as it uses the established Steam Input system
and gives access to it's full feature set, supports every controller Steam supports,
and provides access to thousands of community configurations (web-browsable on [SteamInputDB](https://steaminputdb.com/)).

It can be used to circumvent issues with games and applications that
do not support Steam Input or otherwise pose challenges, like (but not limited to):

- Games with aggressive anti-cheat systems
- Emulators
- Windows Store games/apps
- Games with broken Steam Input support

All while still providing you the full feature set of SteamInput!

SISR can also be used to "tunnel"/forward Steam Input configurations over the network to other machines, including Keyboard/Mouse.  
This makes it possible to use devices like a Steam Deck as a dedicated controller without the need to stream the entire game.

The emulated controllers (and Keyboard/Mouse) are indistinguishable from real hardware and show up at system level.  
SISR achieves this by utilizing [VIIPER](https://github.com/Alia5/VIIPER) (requires **USBIP**).  
Unlike its predecessor [GlosSI](https://github.com/Alia5/GlosSI), it does not use the unmaintained [ViGEm](https://github.com/ViGEm/ViGEmBus) driver.

> [!WARNING]
**Highly experimental work in progress.** Everything is subject to change and may or may not work.  
Expect bugs, crashes, and missing features.

> [!IMPORTANT]
You **do** get the full functionality of SteamInput, _without having to launch your games from Steam_  
For this Steam must still be running in the background, though.   
**Please read the [introduction post](https://alia5.github.io/SISR/main/getting-started/introduction/) before you get started.**    


## ✨🛣️ Features / Roadmap

- Full SteamInput featureset while emulating compatible controllers (indistinguishable from real hardware)
  - Xbox360 _or_
  - DualShock 4
      - Gyro Passthrough
- **Non Steam Mode*
- Multi-platform support (Windows, Linux)
   Multiple operation modes
    - Standalone background service (To be improved)
    - Steam overlay window mode
- Networked operation across computers
    - Use devices like a SteamDeck as dedicated controller without streaming the whole game/display
- 🚧 Xbox One controller emulation
- 🚧 Bundling multiple devices into a single controller
- 🚧 Automatic HidHide integration

## Documentation / FAQ / Help

Read the [documentation](https://alia5.github.io/SISR/)!

## 📝 Contributing

PRs welcome!
See [GitHub Issues](https://github.com/Alia5/SISR/issues) for open tasks.

## 📄 License

```license
SISR - Steam Input System Redirector

Copyright (C) 2025-2026 Peter Repukat

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```




