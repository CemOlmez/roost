STEP_ROLES="box client"
step_title() { echo "Tailscale connected as '$NAME' with Tailscale SSH on"; }
step_check() {
  local state host ssh
  state="$(ts_py 'd["BackendState"]')"
  [[ "$state" == Running ]] || { STEP_DETAIL="not logged in (state: ${state:-unknown})"; return 1; }
  host="$(ts_py 'd["Self"]["HostName"]')"; ssh="$(ts_pref RunSSH)"
  [[ "$ssh" == True ]] || { STEP_DETAIL="Tailscale SSH is off"; return 1; }
  [[ "${host,,}" == "${NAME,,}" ]] || { STEP_DETAIL="hostname is '$host', want '$NAME'"; return 1; }
}
step_apply() {
  need_sudo
  if [[ "$(ts_py 'd["BackendState"]')" == Running ]]; then
    sudo tailscale set --ssh --hostname="$NAME"
  else
    say "  A login link will appear — open it and sign in with the same account as your phone."
    sudo tailscale up --ssh --hostname="$NAME"
  fi
}
