STEP_ROLES="box"; STEP_WSL_ONLY=1
step_title() { echo "Windows: WSL VM never shuts down when idle (.wslconfig vmIdleTimeout=-1)"; }
_wslconfig() { echo "$(win_user_home)/.wslconfig"; }
step_check() { grep -qE '^\s*vmIdleTimeout\s*=\s*-1' "$(_wslconfig)" 2>/dev/null; }
step_apply() {
  local f; f="$(_wslconfig)"
  [[ -f "$f" ]] && cp "$f" "$f.roost-backup"
  if ! grep -q '^\[wsl2\]' "$f" 2>/dev/null; then printf '[wsl2]\n' >>"$f"; fi
  sed -i '/^\s*vmIdleTimeout\s*=/d' "$f"
  sed -i '/^\[wsl2\]/a vmIdleTimeout=-1' "$f"
  warn "Takes effect after:  wsl --shutdown  (in PowerShell)"
}
