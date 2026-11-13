STEP_ROLES="box"; STEP_WSL_ONLY=1; STEP_MANUAL=1
step_title() { echo "BIOS: power on after power loss + wake on LAN (can't be checked from Windows)"; }
step_check() { [[ -f "$ROOST_CONFIG_DIR/bios-done" ]]; }
step_manual() {
  local board; board="$(win_ps '(Get-CimInstance Win32_BaseBoard | ForEach-Object { $_.Manufacturer + " " + $_.Product })')"
  hint "Motherboard: ${board:-unknown}. Reboot, press Del/F2 to enter BIOS, then set:"
  hint "  • Restore on AC Power Loss  (aka AC Back / After Power Failure) → Power On"
  hint "  • Wake on LAN  (aka Power On by PCI-E)                         → Enabled"
  hint "  • ErP / EuP Ready                                              → Disabled"
  hint "Search: \"${board:-your motherboard} restore on AC power loss\" for the exact menu."
  hint "When done:  touch $ROOST_CONFIG_DIR/bios-done"
}
