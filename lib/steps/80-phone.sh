STEP_ROLES="box client"; STEP_MANUAL=1
step_title() { echo "A phone is on your tailnet (Tailscale app) for SSH/Termius access"; }
step_check() {
  local r; r="$(ts_py 'sum(1 for p in (d.get("Peer") or {}).values() if p.get("OS") in ("android","iOS"))')"
  [[ "${r:-0}" -gt 0 ]]
}
step_manual() {
  hint "On your phone install: Claude app (for Remote Control), Tailscale (sign in, same account), Termius (SSH)."
  hint "In Termius add host: $NAME  user: $USER  — Tailscale SSH handles the login, no password or key needed."
}
