# Windows Setup Guide

This guide covers building, deploying, and testing `back-cal` on Windows 11.

---

## Prerequisites

| Requirement | Notes |
|---|---|
| Windows 11 | Earlier versions likely work but are untested |
| Local admin | Required once for `back-cal setup` (Task Scheduler) |
| Go 1.21+ | Only needed if building on Windows; skip if cross-compiling from Linux |

---

## 1. Build the Binary

### Option A — Cross-compile from Linux (recommended)

Run this on your Linux dev machine:

```bash
GOOS=windows GOARCH=amd64 go build -o back-cal.exe .
```

Copy `back-cal.exe` to your Windows machine (USB, network share, OneDrive, etc.).

### Option B — Build directly on Windows

Install Go from https://go.dev/dl/, then in PowerShell:

```powershell
git clone https://github.com/philipf/back-cal
cd back-cal
go build -o back-cal.exe .
```

---

## 2. Install the Binary

Place `back-cal.exe` somewhere permanent — it must stay at the same path after you run `back-cal setup`, because that path is hard-coded into the Task Scheduler tasks.

Recommended location:

```powershell
mkdir "$env:LOCALAPPDATA\back-cal"
# copy back-cal.exe there
```

Optionally add that directory to your `PATH` so you can run `back-cal` from any terminal:

```powershell
[System.Environment]::SetEnvironmentVariable(
    "PATH",
    "$env:PATH;$env:LOCALAPPDATA\back-cal",
    "User"
)
```

Restart your terminal after changing `PATH`.

---

## 3. Create and Edit the Config File

```powershell
back-cal config init
```

This writes a commented template to `%APPDATA%\back-cal\config.yaml`. Open it:

```powershell
notepad "$env:APPDATA\back-cal\config.yaml"
```

Fill in at minimum:

```yaml
api:
  slug: "your-radiator-slug"   # e.g. test-daytime_calendar
  token: ""                    # leave blank — set via env var instead
```

### Storing the token securely

Rather than writing the token into the config file, set it as a persistent user environment variable:

```powershell
[System.Environment]::SetEnvironmentVariable("BACK_CAL_TOKEN", "your-token-here", "User")
```

Restart any open terminals for the change to take effect. The Task Scheduler tasks inherit the user environment, so this works for scheduled runs too.

---

## 4. Test a Manual Run

Before registering scheduled tasks, verify the pipeline end-to-end:

```powershell
back-cal run --force
```

Expected output: no output to stdout (everything goes to the log file). Check the result:

1. Your desktop wallpaper should now show the calendar in the top-right corner.
2. Check the log:

```powershell
Get-Content "$env:LOCALAPPDATA\back-cal\back-cal.log" -Tail 20
```

A successful run looks like:

```
2026/06/17 08:31:02 INFO  run started (force=true)
2026/06/17 08:31:02 INFO  primary monitor: 1920x1080
2026/06/17 08:31:03 INFO  fetched 64862 bytes
2026/06/17 08:31:03 INFO  calendar size: 960x540
2026/06/17 08:31:03 INFO  composited wallpaper written to C:\Users\...\back-cal\wallpaper.png
2026/06/17 08:31:03 INFO  wallpaper set
2026/06/17 08:31:03 INFO  done
```

---

## 5. Register Task Scheduler Tasks

Once the manual run works:

```powershell
back-cal setup
```

This registers two tasks under your user account (no UAC prompt):

| Task name | Trigger |
|---|---|
| `back-cal-logon` | User logon |
| `back-cal-unlock` | Workstation unlock (Event ID 4801) |

### Verify in Task Scheduler

Open Task Scheduler (`taskschd.msc`), navigate to **Task Scheduler Library**, and confirm both tasks appear. Check that:

- **General tab**: "Run only when user is logged on" is selected
- **Triggers tab**: triggers match the above
- **Actions tab**: command points to your `back-cal.exe` with argument `run`

### Enable the unlock trigger (Event ID 4801)

The unlock trigger relies on Windows security audit logging. This is **not enabled by default** on most machines. To enable it:

1. Open **Local Security Policy**: `Win + R` → `secpol.msc`
2. Navigate to **Security Settings → Local Policies → Audit Policy**
3. Double-click **Audit logon events**
4. Check both **Success** and **Failure**
5. Click OK

Alternatively, from an elevated PowerShell:

```powershell
auditpol /set /subcategory:"Other Logon/Logoff Events" /success:enable /failure:enable
```

To confirm Event 4801 is now being generated: lock your screen (`Win + L`), unlock it, then check:

```powershell
Get-WinEvent -LogName Security -MaxEvents 10 |
    Where-Object { $_.Id -eq 4801 } |
    Select-Object TimeCreated, Message
```

If no events appear, the audit policy change has not taken effect yet — try a full logoff/logon cycle.

---

## 6. Testing

### Run the full test suite on Windows

From the repo root in PowerShell:

```powershell
go test ./...
```

Expected results:

| Package | Expected |
|---|---|
| `cmd` | PASS — datestamp guard and path resolution tests |
| `internal/composite` | PASS — all compositing tests |
| `internal/fetch` | PASS (unit); integration test skipped (`-short`) |
| `internal/monitor` | `TestPrimaryMonitor_NonWindows` skipped; `TestPrimaryMonitor_Windows` runs and should PASS |
| `internal/scheduler` | `TestRegister_NonWindows` skipped; `TestRegister_Windows` runs — registers test tasks using `cmd.exe`, should PASS |

### Run the live integration test

```powershell
$env:BACK_CAL_SLUG = "test-daytime_calendar"
go test ./internal/fetch/... -run TestBMP_Integration -v
```

Expected: receives a ~65 KB BMP with valid magic bytes.

### Tests that may fail and why

**`TestRegister_Windows` leaves test tasks behind**

The test registers tasks named `back-cal-logon` and `back-cal-unlock` pointing at `cmd.exe`. After running, clean them up:

```powershell
schtasks /delete /tn back-cal-logon /f
schtasks /delete /tn back-cal-unlock /f
```

Then re-run `back-cal setup` to restore the real tasks.

**`TestPrimaryMonitor_Windows` fails with "primary monitor not found"**

This can happen if the binary is run in a non-interactive session (e.g. a CI runner or SSH session without a display). Run it in an interactive PowerShell terminal on the desktop.

**Composite tests fail with file permission errors**

Ensure your temp directory (`%TEMP%`) is writable. These tests use `t.TempDir()` which maps to the OS temp dir.

---

## 7. Troubleshooting

### Wallpaper does not change after unlock

Work through this checklist in order:

1. **Check the log** — `Get-Content "$env:LOCALAPPDATA\back-cal\back-cal.log" -Tail 30`
2. **Check the datestamp guard** — `Get-Content "$env:LOCALAPPDATA\back-cal\last-run.txt"` — if it shows today's date, the guard skipped the run. Use `back-cal run --force` to bypass.
3. **Check Event 4801 is firing** — see [Enable the unlock trigger](#enable-the-unlock-trigger-event-id-4801) above.
4. **Check the task exists** — `schtasks /query /tn back-cal-unlock /fo LIST`
5. **Run manually** — `back-cal run --force` to confirm the pipeline works outside Task Scheduler.
6. **Check task history** — in Task Scheduler, right-click the task → History.

### "api.slug is required" error in log

The config file either doesn't exist or the slug field is empty. Run:

```powershell
back-cal config init
notepad "$env:APPDATA\back-cal\config.yaml"
```

### "api.token is required" error in log

The `BACK_CAL_TOKEN` environment variable is not set and `api.token` is empty in the config. Set the env var:

```powershell
[System.Environment]::SetEnvironmentVariable("BACK_CAL_TOKEN", "your-token", "User")
```

Note: after setting a user environment variable, you must start a **new** PowerShell session for it to be visible. Task Scheduler picks it up on the next logon.

### Wallpaper is set but appears stretched or letterboxed

The composited image should already match your screen resolution exactly. If it looks wrong:

1. Check the log for the monitor dimensions detected: `INFO  primary monitor: WxH`
2. Confirm those match your actual display resolution in **Settings → Display**
3. Check **Personalise → Background** — right-click the desktop and ensure the wallpaper fit is set to **Center** (the `back-cal setup` run sets this via the registry, but some themes override it)

### Task runs but wallpaper path is wrong after moving `back-cal.exe`

Re-run `back-cal setup` from the new location — it re-registers the tasks using `os.Executable()`:

```powershell
back-cal setup
```

### Resetting everything

To start fresh:

```powershell
schtasks /delete /tn back-cal-logon /f
schtasks /delete /tn back-cal-unlock /f
Remove-Item "$env:APPDATA\back-cal" -Recurse -Force
Remove-Item "$env:LOCALAPPDATA\back-cal" -Recurse -Force
```

Then repeat from [Step 3](#3-create-and-edit-the-config-file).
