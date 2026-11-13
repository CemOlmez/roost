STEP_ROLES="box"; STEP_WSL_ONLY=1
step_title() { echo "Windows: WSL starts at logon, PC never sleeps on AC, hibernate/Fast Startup off, wake-on-LAN"; }
_ps1_check() {
  # Reads state without admin rights. Uses hex values only, so it works on any Windows language.
  win_ps "\$t=[bool](Get-ScheduledTask -TaskName RoostKeepAlive -ErrorAction SilentlyContinue)
\$h=(Get-ItemProperty HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Power).HibernateEnabled
\$hex=@(powercfg /q SCHEME_CURRENT SUB_SLEEP STANDBYIDLE | ForEach-Object { [regex]::Match(\$_,'0x[0-9a-fA-F]{8}\s*\$').Value.Trim() } | Where-Object { \$_ })
\$s=\$hex[-2]
\"TASK=\$t HIB=\$h STANDBY_AC=\$s\""
}
step_check() {
  local r; r="$(_ps1_check)"
  [[ "$r" == *"TASK=True"* && "$r" == *"HIB=0"* && "$r" == *"STANDBY_AC=0x00000000"* ]] \
    || { STEP_DETAIL="${r:-could not query Windows}"; return 1; }
}
step_apply() {
  local tmp
  tmp="$(win_temp)" || die "could not find Windows TEMP"
  cp "$ROOST_HOME/windows/roost-windows.ps1" "$tmp/roost-windows.ps1"
  say "  Windows will ask for admin permission (UAC) — click Yes."
  win_ps "Start-Process powershell -Verb RunAs -Wait -ArgumentList '-NoProfile -ExecutionPolicy Bypass -File \"$(wslpath -w "$tmp/roost-windows.ps1")\" -Distro $(wsl_distro)'"
  [[ -f "$tmp/roost-windows.log" ]] && sed 's/\r$//; s/^/      /' "$tmp/roost-windows.log"
}
