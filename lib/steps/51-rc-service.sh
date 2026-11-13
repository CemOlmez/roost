STEP_ROLES="box"
UNIT="$HOME/.config/systemd/user/roost-rc.service"
step_title() { echo "Claude Remote Control always running (service roost-rc, sessions start in $RC_DIR)"; }
step_check() {
  [[ -f "$UNIT" ]] || { STEP_DETAIL="service not installed"; return 1; }
  systemctl --user is-active --quiet roost-rc || { STEP_DETAIL="service installed but not running — see: roost rc logs"; return 1; }
}
step_apply() {
  local claude_bin tmux_bin
  claude_bin="$(command -v claude || echo "$HOME/.local/bin/claude")"
  tmux_bin="$(command -v tmux)"
  mkdir -p "$(dirname "$UNIT")"
  cat >"$UNIT" <<UNIT
[Unit]
Description=roost: Claude Remote Control (attach with: roost rc attach)
After=network-online.target

[Service]
Type=forking
Environment=PATH=$HOME/.local/bin:/usr/local/bin:/usr/bin:/bin
WorkingDirectory=$RC_DIR
ExecStart=$tmux_bin new-session -d -s roost-rc -c $RC_DIR "$claude_bin remote-control --name $NAME"
ExecStop=$tmux_bin kill-session -t roost-rc
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
UNIT
  systemctl --user daemon-reload
  systemctl --user enable --now roost-rc
}
