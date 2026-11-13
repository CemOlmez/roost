package steps

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/cemolmez/roost/internal/host"
	"github.com/cemolmez/roost/internal/ini"
	"github.com/cemolmez/roost/internal/step"
)

// WSLConfig: the WSL VM never shuts itself down when idle.
func WSLConfig() step.Step {
	path := func(ctx context.Context, env *step.Env) (string, error) {
		dir, err := host.WindowsEnvDir(ctx, env.Host, "USERPROFILE")
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, ".wslconfig"), nil
	}
	return step.Step{
		ID:      "wslconfig",
		Title:   title("Windows: WSL never shuts down when idle (.wslconfig vmIdleTimeout=-1)"),
		Roles:   boxOnly,
		WSLOnly: true,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			p, err := path(ctx, env)
			if err != nil {
				return step.Failf("can't find your Windows profile: " + err.Error())
			}
			if v, _ := ini.Get(readText(env.Host, p), "wsl2", "vmIdleTimeout"); v == "-1" {
				return step.Pass()
			}
			return step.Failf("vmIdleTimeout is not -1 in " + p)
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			p, err := path(ctx, env)
			if err != nil {
				return err
			}
			old := readText(env.Host, p)
			if old != "" {
				if err := env.Host.WriteFile(p+".roost-backup", []byte(old), 0o600); err != nil {
					return err
				}
			}
			content := strings.ReplaceAll(ini.Set(old, "wsl2", "vmIdleTimeout", "-1"), "\n", "\r\n")
			if err := env.Host.WriteFile(p, []byte(content), 0o600); err != nil {
				return err
			}
			env.UI.Hint("Takes effect after the next  wsl --shutdown  (or a reboot).")
			return nil
		},
	}
}

// windowsHostCheck reads Windows state without admin rights. It only parses
// booleans, numbers and hex values, so it works in any Windows display language.
const windowsHostCheck = `$t=Get-ScheduledTask -TaskName RoostKeepAlive -ErrorAction SilentlyContinue
$mine=[bool]($t -and $t.Principal.UserId -and ($t.Principal.UserId -split '\\')[-1] -eq $env:USERNAME)
$h=(Get-ItemProperty HKLM:\SYSTEM\CurrentControlSet\Control\Power).HibernateEnabled
$hex=@(powercfg /q SCHEME_CURRENT SUB_SLEEP STANDBYIDLE | ForEach-Object { [regex]::Match($_,'0x[0-9a-fA-F]{8}\s*$').Value.Trim() } | Where-Object { $_ })
"TASK=$mine HIBERNATE=$h STANDBY_AC=$($hex[-2])"`

var windowsHostRe = regexp.MustCompile(`TASK=(\w+) HIBERNATE=(\d*) STANDBY_AC=(0x[0-9a-fA-F]+|)`)

// parseWindowsHost turns the check script's output into a list of problems.
func parseWindowsHost(out string) []string {
	m := windowsHostRe.FindStringSubmatch(out)
	if m == nil {
		return []string{"could not read Windows settings"}
	}
	var problems []string
	if m[1] != "True" {
		problems = append(problems, "WSL doesn't start at your logon (no RoostKeepAlive task for you)")
	}
	if m[2] != "0" {
		problems = append(problems, "hibernate/Fast Startup is on")
	}
	if !strings.EqualFold(m[3], "0x00000000") {
		problems = append(problems, "PC goes to sleep on AC power")
	}
	return problems
}

var distroRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// psQuote returns s as a single-quoted PowerShell string literal.
func psQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// encodePowerShell encodes a script for powershell.exe -EncodedCommand
// (base64 of UTF-16LE).
func encodePowerShell(script string) string {
	u := utf16.Encode([]rune(script))
	b := make([]byte, 2*len(u))
	for i, r := range u {
		binary.LittleEndian.PutUint16(b[2*i:], r)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// windowsHostScript is the elevated script with its inputs bound.
func windowsHostScript(distro, taskUser string) (string, error) {
	if !distroRe.MatchString(distro) {
		return "", fmt.Errorf("WSL distro name %q has characters roost can't pass safely", distro)
	}
	return "$Distro = " + psQuote(distro) + "\n$TaskUser = " + psQuote(taskUser) + "\n" + string(Asset("roost-windows.ps1")), nil
}

// WindowsHost: WSL starts at logon, no sleep on AC, hibernate/Fast Startup off,
// wake-on-LAN (best effort). The fix runs assets/roost-windows.ps1 once as
// administrator (one UAC prompt), passed in memory.
func WindowsHost() step.Step {
	return step.Step{
		ID:      "windows-host",
		Title:   title("Windows: WSL starts at logon, PC never sleeps on AC power, Fast Startup off"),
		Roles:   boxOnly,
		WSLOnly: true,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			out, err := host.PowerShell(ctx, env.Host, windowsHostCheck)
			if err != nil {
				return step.Failf("could not query Windows: " + err.Error())
			}
			if p := parseWindowsHost(out); len(p) > 0 {
				return step.Failf(strings.Join(p, "; "))
			}
			return step.Pass()
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			user, err := host.PowerShell(ctx, env.Host, `"$env:USERDOMAIN\$env:USERNAME"`)
			if err != nil || user == "" {
				return fmt.Errorf("could not read your Windows user name: %w", err)
			}
			script, err := windowsHostScript(host.WSLDistro(env.Host), user)
			if err != nil {
				return err
			}
			env.UI.Hint("Windows will ask for administrator permission — click Yes, then press Enter in the new window.")
			launch := "Start-Process powershell -Verb RunAs -Wait -ArgumentList " +
				"'-NoProfile','-ExecutionPolicy','Bypass','-EncodedCommand','" + encodePowerShell(script) + "'"
			if _, err := host.PowerShell(ctx, env.Host, launch); err != nil {
				return fmt.Errorf("the administrator prompt was cancelled or failed: %w", err)
			}
			return nil
		},
	}
}

// WindowsAutologon: Windows logs in by itself, so WSL comes back after power cuts and updates.
func WindowsAutologon() step.Step {
	return step.Step{
		ID:      "windows-autologon",
		Title:   title("Windows logs in by itself after a reboot (so the box returns after power cuts)"),
		Roles:   boxOnly,
		WSLOnly: true,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			out, err := host.PowerShell(ctx, env.Host,
				`(Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon').AutoAdminLogon`)
			if err == nil && strings.TrimSpace(out) == "1" {
				return step.Pass()
			}
			return step.Failf("auto-login is off")
		},
		Manual: func(context.Context, *step.Env) []string {
			return []string{
				"Download Autologon from Microsoft: https://learn.microsoft.com/sysinternals/downloads/autologon",
				"Run it, enter your Windows password, click Enable. (roost never sees your password.)",
			}
		},
	}
}

// BIOS: power on after power loss + wake-on-LAN. Can't be detected, so the user confirms it.
func BIOS() step.Step {
	return step.Step{
		ID:          "bios",
		Title:       title("BIOS: PC powers on after a power cut, wake-on-LAN enabled"),
		Roles:       boxOnly,
		WSLOnly:     true,
		Confirmable: true,
		Check: func(_ context.Context, env *step.Env) step.Result {
			if env.Host.Exists(confirmedPath(env.Host, "bios")) {
				return step.Pass()
			}
			return step.Failf("not confirmed yet")
		},
		Manual: func(ctx context.Context, env *step.Env) []string {
			board, err := host.PowerShell(ctx, env.Host,
				`Get-CimInstance Win32_BaseBoard | ForEach-Object { $_.Manufacturer + ' ' + $_.Product }`)
			if err != nil || board == "" {
				board = "your motherboard"
			}
			return []string{
				"Motherboard: " + board + ". Reboot and press Del or F2 to open the BIOS, then set:",
				"  • Restore on AC Power Loss (or AC Back / After Power Failure) → Power On",
				"  • Wake on LAN (or Power On by PCI-E)                         → Enabled",
				"  • ErP / EuP Ready                                            → Disabled",
				fmt.Sprintf("Search \"%s restore on AC power loss\" for the exact menu.", board),
				"When done, run:  roost confirm bios",
			}
		},
	}
}
