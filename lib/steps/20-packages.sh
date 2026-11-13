STEP_ROLES="box client"
PKGS=(tmux git curl python3)
step_title() { echo "Base packages: ${PKGS[*]}"; }
step_check() { local p; for p in "${PKGS[@]}"; do has "$p" || { STEP_DETAIL="missing: $p"; return 1; }; done; }
step_apply() { need_sudo; sudo apt-get update -qq && sudo apt-get install -y -qq "${PKGS[@]}"; }
