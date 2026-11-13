STEP_ROLES="box client"; STEP_MANUAL=1
step_title() { echo "Claude Code logged in (Pro/Max/Team subscription — API keys can't use Remote Control)"; }
step_check() { [[ -s "$HOME/.claude/.credentials.json" ]]; }
step_manual() { hint "Run:  claude   then type  /login  and follow the link. Re-run roost init afterwards."; }
