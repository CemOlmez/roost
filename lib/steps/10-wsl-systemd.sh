STEP_ROLES="box client"; STEP_WSL_ONLY=1
step_title() { echo "WSL runs systemd (needed for background services)"; }
step_check() { has_systemd; }
step_apply() {
  need_sudo
  if grep -q '^\[boot\]' /etc/wsl.conf 2>/dev/null; then
    sudo sed -i '/^\[boot\]/a systemd=true' /etc/wsl.conf
  else
    printf '[boot]\nsystemd=true\n' | sudo tee -a /etc/wsl.conf >/dev/null
  fi
  warn "Restart WSL to activate: in PowerShell run  wsl --shutdown  then reopen Ubuntu and re-run: roost init"
  exit 3
}
