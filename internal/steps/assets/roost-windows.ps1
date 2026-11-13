# roost-windows.ps1 — Windows-side setup for a roost "box" (a PC that hosts your WSL dev box).
# Run elevated. roost init copies this to %TEMP% and launches it with a UAC prompt.
# Safe to run more than once. Writes a log next to itself.

param(
  [string]$Distro = "Ubuntu"
)

$ErrorActionPreference = "Continue"
$log = Join-Path $PSScriptRoot "roost-windows.log"
"roost-windows.ps1  $(Get-Date -Format s)  distro=$Distro" | Out-File $log -Encoding utf8

function Log($msg) { $msg | Tee-Object -FilePath $log -Append | Out-Host }

# 1. Keep WSL alive: a hidden task at logon that starts the distro and holds it open.
#    systemd inside WSL then starts Tailscale and roost-rc on its own.
$taskName = "RoostKeepAlive"
$action   = New-ScheduledTaskAction -Execute "wsl.exe" -Argument "-d $Distro --exec /bin/sh -c `"exec sleep infinity`""
$trigger  = New-ScheduledTaskTrigger -AtLogOn -User $env:USERNAME
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries `
              -ExecutionTimeLimit ([TimeSpan]::Zero) -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -Hidden
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Settings $settings `
  -Description "roost: keep WSL distro '$Distro' running" -Force | Out-Null
Start-ScheduledTask -TaskName $taskName
Log "OK  scheduled task $taskName (starts WSL at logon, runs now)"

# 2. Never sleep on AC power. The screen may still turn off.
powercfg /change standby-timeout-ac 0
powercfg /change hibernate-timeout-ac 0
Log "OK  sleep on AC power: never"

# 3. Hibernate off — this also turns off Fast Startup, which breaks 'power on after AC loss' and wake-on-LAN.
powercfg /hibernate off
Log "OK  hibernate + Fast Startup: off"

# 4. Wired network cards: allow wake on magic packet (wake-on-LAN).
$wired = Get-NetAdapter -Physical -ErrorAction SilentlyContinue | Where-Object { $_.MediaType -eq "802.3" }
if (-not $wired) { Log "WARN no wired (Ethernet) adapter found — wake-on-LAN needs a cable" }
foreach ($a in $wired) {
  try {
    Set-NetAdapterPowerManagement -Name $a.Name -WakeOnMagicPacket Enabled -ErrorAction Stop
    Log "OK  wake on magic packet: $($a.Name)  MAC $($a.MacAddress)"
  } catch {
    Log "WARN could not enable wake on $($a.Name): $($_.Exception.Message)"
  }
}

Log "Done. You can close this window."
Start-Sleep -Seconds 3
