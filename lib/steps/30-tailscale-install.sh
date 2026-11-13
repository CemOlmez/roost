STEP_ROLES="box client"
step_title() { echo "Tailscale installed (private network between your devices)"; }
step_check() { has tailscale; }
step_apply() { need_sudo; curl -fsSL https://tailscale.com/install.sh | sh; }
