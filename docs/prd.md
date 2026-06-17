# back-cal — Product Requirements Document

## Overview

`back-cal` is a Go CLI tool that fetches a dual calendar image from a remote API and composites it onto the desktop wallpaper, positioned in the top-right corner. It runs once per day, triggered automatically at Windows login and workstation unlock via Task Scheduler.

## Problem Statement

A dual calendar widget is available as an SVG/BMP via a remote HTTP API. There is no native Windows mechanism to display it as a persistent desktop widget. The goal is to embed it into the desktop wallpaper so it is always visible without requiring a running background process or third-party widget framework.

The user works at hot-desks, so monitor resolution varies daily. The solution must detect the primary monitor at runtime and adapt accordingly.

## Goals

- Fetch the calendar image from the API once per day
- Composite it onto a full-screen canvas and set it as the Windows wallpaper
- Position the calendar top-right with configurable padding
- Trigger automatically at login and workstation unlock with no user interaction
- Compile and build on Linux (development machine); run on Windows 11 (work machine)

## Non-Goals

- Setting the wallpaper on Linux or macOS
- Rasterizing SVG locally
- Running as a persistent background service or daemon
- Displaying anything other than the calendar image

---

## CLI Interface

Binary name: `back-cal`

### Subcommands

| Command | Description |
|---|---|
| `back-cal` | Alias for `run` (default subcommand) |
| `back-cal run` | Fetch, composite, and set wallpaper. Skips if already ran today. |
| `back-cal run --force` | Same as `run` but bypasses the datestamp guard |
| `back-cal setup` | Register Task Scheduler tasks (logon + unlock triggers) |
| `back-cal config init` | Write a commented default config file to `%APPDATA%\back-cal\config.yaml` |

---

## Configuration

Config file location: `%APPDATA%\back-cal\config.yaml`

Managed via [viper](https://github.com/spf13/viper). All values have defaults except `api.slug` and `api.token`.

```yaml
api:
  url: https://gotta-go.notnot.uk/v1/frame
  slug: ""        # required: x-radiator-slug header value
  token: ""       # required: x-radiator-token header value (prefer BACK_CAL_TOKEN env var)

canvas:
  background_color: "#000000"  # hex color for the full-screen canvas

padding:
  top: 20    # pixels from top edge
  right: 20  # pixels from right edge

paths:
  wallpaper: ""   # default: %LOCALAPPDATA%\back-cal\wallpaper.bmp
  last_run: ""    # default: %LOCALAPPDATA%\back-cal\last-run.txt
  log: ""         # default: %LOCALAPPDATA%\back-cal\back-cal.log
```

### Environment Variable Override

The API token can be supplied via the environment variable `BACK_CAL_TOKEN`, which takes precedence over the config file value. This avoids storing credentials in plaintext.

---

## Functional Requirements

### FR1 — Fetch calendar image

- HTTP GET to `api.url`
- Request headers:
  - `Accept: image/bmp`
  - `Accept-Encoding: gzip`
  - `x-radiator-slug: <api.slug>`
  - `x-radiator-token: <api.token>`
- Response: BMP image at fixed size (server-determined)

### FR2 — Detect primary monitor

- Enumerate display monitors via `EnumDisplayMonitors` (Windows API)
- Select the monitor flagged as primary (`MONITORINFOF_PRIMARY`)
- Use its pixel dimensions as the canvas size
- If only one monitor exists (e.g. undocked laptop), use that monitor

### FR3 — Composite image

- Create an RGBA canvas at primary monitor resolution
- Fill with `canvas.background_color`
- Decode the fetched BMP
- Place it at `(canvas_width - bmp_width - padding.right, padding.top)`
- Preserve the BMP's original pixel dimensions (no scaling)
- Encode the result as BMP and write to `paths.wallpaper`

### FR4 — Set wallpaper

- Call `SystemParametersInfo(SPI_SETDESKWALLPAPER, 0, path, SPIF_UPDATEINIFILE|SPIF_SENDCHANGE)`
- Wallpaper style should be set to "Tile: No / Position: Center" or "Fill" — the canvas already matches the screen resolution, so no scaling by Windows is needed

### FR5 — Datestamp guard

- On startup, `run` reads `paths.last_run`
- If the file contains today's date (`YYYY-MM-DD`), exit 0 immediately (already ran today)
- On successful wallpaper set, write today's date to `paths.last_run`
- On any error (network, file I/O, API), log the error and exit without updating `paths.last_run` so the next trigger retries

### FR6 — Logging

- Append-only log at `paths.log`
- Log entries: timestamp, level (INFO/ERROR), message
- Logged events: run start, guard check result, fetch result, composite result, wallpaper set result, errors with detail

### FR7 — Task Scheduler setup (`back-cal setup`)

- Creates two Task Scheduler tasks under the current user's context (no elevation required):
  1. **back-cal-logon** — trigger: user logon
  2. **back-cal-unlock** — trigger: event log entry, `Microsoft-Windows-Security-Auditing`, Event ID 4801 (workstation unlock)
- Action for both: run `back-cal run` (resolved to the current executable path via `os.Executable()`)
- If tasks already exist, overwrite them
- Print confirmation to stdout

### FR8 — Config init (`back-cal config init`)

- Write a commented YAML template to `%APPDATA%\back-cal\config.yaml`
- If the file already exists, print a warning and exit without overwriting (unless `--force` is passed)
- Print the path of the written file to stdout

---

## Error Handling

| Scenario | Behaviour |
|---|---|
| Network unreachable | Log error, exit 0, do not update datestamp |
| API returns non-200 | Log status code and body, exit 0, do not update datestamp |
| Primary monitor not found | Log error, exit 1 |
| Config file missing | Log warning, use defaults (will fail if slug/token are empty) |
| Wallpaper file write fails | Log error, exit 0, do not update datestamp |
| Wallpaper set API fails | Log error, exit 0, do not update datestamp |

---

## Project Structure

```
back-cal/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go         # cobra root, viper init
│   ├── run.go          # run subcommand
│   ├── setup.go        # setup subcommand
│   └── config.go       # config init subcommand
├── internal/
│   ├── fetch/
│   │   └── fetch.go    # HTTP fetch logic
│   ├── composite/
│   │   └── composite.go # canvas + BMP placement (pure Go, cross-platform)
│   ├── monitor/
│   │   ├── monitor_windows.go  # EnumDisplayMonitors (build tag: windows)
│   │   └── monitor_other.go    # stub returning error (build tag: !windows)
│   ├── wallpaper/
│   │   ├── wallpaper_windows.go # SystemParametersInfo (build tag: windows)
│   │   └── wallpaper_other.go   # stub (build tag: !windows)
│   └── scheduler/
│       ├── scheduler_windows.go # Task Scheduler COM/XML (build tag: windows)
│       └── scheduler_other.go   # stub (build tag: !windows)
└── docs/
    └── prd.md
```

---

## Platform Notes

| Concern | Detail |
|---|---|
| Build platform | Linux (development), Windows 11 (deployment) |
| Build tags | `//go:build windows` on all OS-specific files; `//go:build !windows` on stubs |
| `go build` on Linux | Succeeds; OS-specific packages compile to no-ops |
| `go test` on Linux | Fetch and composite packages are fully testable |
| Windows version | Windows 11; APIs used are available since Windows XP |
| Elevation | Not required; Task Scheduler tasks run in user context |

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI subcommand framework |
| `github.com/spf13/viper` | Config file + env var management |
| `golang.org/x/sys/windows` | Windows API access (monitor, wallpaper) |
| stdlib `image`, `image/color`, `golang.org/x/image/bmp` | BMP decode/encode and canvas compositing |

---

## Out of Scope / Future Considerations

- Scaling the calendar image to a percentage of screen height (currently fixed size from API)
- Linux wallpaper support
- Multiple calendar widgets / positions
- Tray icon or systray indicator
- Auto-update mechanism
