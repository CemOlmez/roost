STEP_ROLES="box client"
step_title() { echo "Claude Code installed"; }
step_check() { has claude || [[ -x "$HOME/.local/bin/claude" ]]; }
step_apply() { curl -fsSL https://claude.ai/install.sh | bash; }
