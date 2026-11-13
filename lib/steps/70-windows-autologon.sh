STEP_ROLES="box"; STEP_WSL_ONLY=1; STEP_MANUAL=1
step_title() { echo "Windows logs in by itself after a reboot (so WSL comes back after power cuts and updates)"; }
step_check() { [[ "$(win_ps '(Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon").AutoAdminLogon')" == 1 ]]; }
step_manual() {
  hint "Download Autologon from Microsoft Sysinternals: https://learn.microsoft.com/sysinternals/downloads/autologon"
  hint "Run it, enter your Windows password, click Enable. (roost never handles your password.)"
}
