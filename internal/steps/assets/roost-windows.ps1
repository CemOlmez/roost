# roost-windows.ps1 — Windows-side setup for a roost "box" (a PC that hosts your WSL dev box).
#
# roost runs this once, elevated, passing it to PowerShell in memory
# (-EncodedCommand), so there is no script file another process could swap
# while the UAC prompt is open. roost prepends two variables:
#   $Distro    WSL distribution to keep running (validated: [A-Za-z0-9._-]+)
#   $TaskUser  the Windows account that ran roost (DOMAIN\user), not the admin
#              account UAC may have elevated to
# Safe to run more than once.

$ErrorActionPreference = "Continue"
function Ok($msg)   { Write-Host "OK    $msg" -ForegroundColor Green }
function Warn($msg) { Write-Host "WARN  $msg" -ForegroundColor Yellow }

# 1. Keep WSL alive: at the user's logon, start the distro in a hidden window
#    and hold it open. systemd inside WSL then starts Tailscale and roost-rc.
$taskName  = "RoostKeepAlive"
$action    = New-ScheduledTaskAction -Execute "powershell.exe" `
               -Argument "-NoProfile -WindowStyle Hidden -Command wsl.exe -d $Distro --exec sleep infinity"
$trigger   = New-ScheduledTaskTrigger -AtLogOn -User $TaskUser
$principal = New-ScheduledTaskPrincipal -UserId $TaskUser -LogonType Interactive -RunLevel Limited
$settings  = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries `
               -ExecutionTimeLimit ([TimeSpan]::Zero) -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Principal $principal `
  -Settings $settings -Description "roost: keep WSL distro '$Distro' running for $TaskUser" -Force | Out-Null
Start-ScheduledTask -TaskName $taskName
Ok "scheduled task $taskName (starts WSL at $TaskUser's logon, running now)"

# 2. Never sleep on AC power. The screen may still turn off.
powercfg /change standby-timeout-ac 0
powercfg /change hibernate-timeout-ac 0
Ok "sleep on AC power: never"

# 3. Hibernate off — also turns off Fast Startup, which breaks 'power on after AC loss' and wake-on-LAN.
powercfg /hibernate off
Ok "hibernate + Fast Startup: off"

# 4. Wired network cards: wake on magic packet (wake-on-LAN). Best effort.
$wired = Get-NetAdapter -Physical -ErrorAction SilentlyContinue | Where-Object { $_.MediaType -eq "802.3" }
if (-not $wired) { Warn "no wired (Ethernet) adapter found — wake-on-LAN needs a cable" }
foreach ($a in $wired) {
  try {
    Set-NetAdapterPowerManagement -Name $a.Name -WakeOnMagicPacket Enabled -ErrorAction Stop
    Ok "wake on magic packet: $($a.Name)  (MAC $($a.MacAddress))"
  } catch {
    Warn "could not enable wake on $($a.Name): $($_.Exception.Message)"
  }
}

Write-Host ""
Read-Host "Done. Press Enter to close this window and return to roost"
