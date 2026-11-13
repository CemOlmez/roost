STEP_ROLES="box"
step_title() { echo "Your services keep running when no terminal is open (systemd linger)"; }
step_check() { [[ "$(loginctl show-user "$USER" -p Linger --value 2>/dev/null)" == yes ]]; }
step_apply() { need_sudo; sudo loginctl enable-linger "$USER"; }
