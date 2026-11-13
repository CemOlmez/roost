# shellcheck shell=bash
# Shared helpers for roost. Sourced by bin/roost.

ROOST_CONFIG_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/roost"
ROOST_CONFIG="$ROOST_CONFIG_DIR/config"

# ---------- output ----------
if [[ -t 1 ]]; then
  C_RESET=$'\e[0m'; C_DIM=$'\e[2m'; C_BOLD=$'\e[1m'
  C_GREEN=$'\e[32m'; C_RED=$'\e[31m'; C_YELLOW=$'\e[33m'; C_BLUE=$'\e[34m'
else
  C_RESET=''; C_DIM=''; C_BOLD=''; C_GREEN=''; C_RED=''; C_YELLOW=''; C_BLUE=''
fi

say()   { printf '%s\n' "$*"; }
head1() { printf '\n%s%s%s\n' "$C_BOLD" "$*" "$C_RESET"; }
ok()    { printf '  %s✓%s %s\n' "$C_GREEN" "$C_RESET" "$*"; }
bad()   { printf '  %s✗%s %s\n' "$C_RED" "$C_RESET" "$*"; }
warn()  { printf '  %s!%s %s\n' "$C_YELLOW" "$C_RESET" "$*"; }
skip()  { printf '  %s-%s %s%s%s\n' "$C_DIM" "$C_RESET" "$C_DIM" "$*" "$C_RESET"; }
hint()  { printf '      %s%s%s\n' "$C_DIM" "$*" "$C_RESET"; }
die()   { printf '%serror:%s %s\n' "$C_RED" "$C_RESET" "$*" >&2; exit 1; }

# ask "question" default(y|n) -> returns 0 for yes
ask() {
  local q="$1" def="${2:-y}" ans prompt
  [[ "${ROOST_YES:-0}" == 1 ]] && return 0
  [[ "$def" == y ]] && prompt="[Y/n]" || prompt="[y/N]"
  read -r -p "  $q $prompt " ans </dev/tty || ans=""
  ans="${ans:-$def}"
  [[ "$ans" =~ ^[Yy] ]]
}

# ask_value "question" default -> echoes answer
ask_value() {
  local q="$1" def="$2" ans
  if [[ "${ROOST_YES:-0}" == 1 ]]; then echo "$def"; return; fi
  read -r -p "  $q [$def] " ans </dev/tty || ans=""
  echo "${ans:-$def}"
}

need_sudo() {
  sudo -n true 2>/dev/null && return 0
  say "  ${C_DIM}(sudo needs your Linux password)${C_RESET}"
  sudo -v
}

# ---------- environment detection ----------
is_wsl()      { grep -qi microsoft /proc/sys/kernel/osrelease 2>/dev/null; }
has_systemd() { [[ "$(ps -p 1 -o comm= 2>/dev/null)" == systemd ]]; }
has()         { command -v "$1" >/dev/null 2>&1; }

win_ps() {
  # Run a PowerShell command on the Windows host (WSL only); strips CRs.
  powershell.exe -NoProfile -NonInteractive -Command "$1" 2>/dev/null | tr -d '\r'
}

win_user_home() {
  # /mnt/c/Users/<name> of the Windows user running WSL
  local p
  p="$(win_ps '$env:USERPROFILE')"
  [[ -n "$p" ]] && wslpath -u "$p"
}

win_temp() {
  local p
  p="$(win_ps '$env:TEMP')"
  [[ -n "$p" ]] && wslpath -u "$p"
}

wsl_distro() { echo "${WSL_DISTRO_NAME:-Ubuntu}"; }

# ---------- config ----------
# Keys: ROLE (box|client), NAME (tailnet hostname), BOX (client: which box to attach to), RC_DIR (where Remote Control spawns sessions)
load_config() {
  ROLE=""; NAME=""; BOX=""; RC_DIR=""
  # shellcheck disable=SC1090
  [[ -f "$ROOST_CONFIG" ]] && source "$ROOST_CONFIG"
  : "${RC_DIR:=$HOME}"
}

save_config() {
  mkdir -p "$ROOST_CONFIG_DIR"
  cat >"$ROOST_CONFIG" <<EOF
# roost config — edit freely, then run: roost init
ROLE=$ROLE
NAME=$NAME
BOX=$BOX
RC_DIR=$RC_DIR
EOF
}

# ---------- tailscale helpers (no jq needed) ----------
ts_py() {
  # ts_py '<python expr using d>'  — evaluates against `tailscale status --json`
  tailscale status --json 2>/dev/null | python3 -c "import json,sys
try: d=json.load(sys.stdin)
except Exception: sys.exit(1)
print($1)" 2>/dev/null
}
ts_pref() {
  tailscale debug prefs 2>/dev/null | python3 -c "import json,sys
try: d=json.load(sys.stdin)
except Exception: sys.exit(1)
print(d.get('$1'))" 2>/dev/null
}
