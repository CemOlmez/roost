package steps

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/step"
)

func claudeConfigDir(env *step.Env) string {
	if d := env.Host.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join(env.Host.HomeDir(), ".claude")
}

func claudeBin(env *step.Env) string {
	if p := env.Host.LookPath("claude"); p != "" {
		return p
	}
	if p := filepath.Join(env.Host.HomeDir(), ".local", "bin", "claude"); env.Host.Exists(p) {
		return p
	}
	return ""
}

// ClaudeInstalled: Claude Code is installed.
func ClaudeInstalled() step.Step {
	return step.Step{
		ID:    "claude-installed",
		Title: title("Claude Code installed"),
		Roles: both,
		Check: func(_ context.Context, env *step.Env) step.Result {
			if claudeBin(env) != "" {
				return step.Pass()
			}
			return step.Failf("claude not found")
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			return shell(ctx, env, "curl -fsSL https://claude.ai/install.sh | bash")
		},
	}
}

// ClaudeLoggedIn: Claude Code has a subscription login (Remote Control can't use API keys).
func ClaudeLoggedIn() step.Step {
	return step.Step{
		ID:    "claude-login",
		Title: title("Claude Code logged in with a Claude subscription (API keys can't use Remote Control)"),
		Roles: both,
		Check: func(_ context.Context, env *step.Env) step.Result {
			if env.Host.Exists(filepath.Join(claudeConfigDir(env), ".credentials.json")) {
				return step.Pass()
			}
			return step.Failf("not logged in")
		},
		Manual: func(context.Context, *step.Env) []string {
			return []string{"Run  claude , type  /login  and follow the link. Then run  roost init  again."}
		},
	}
}

const (
	rcUnitName = "roost-rc.service"
	// rcSocket is the dedicated tmux socket (tmux -L) for the Remote Control session.
	rcSocket = "roost-rc"
)

func rcUnitPath(env *step.Env) string {
	return filepath.Join(env.Host.HomeDir(), ".config", "systemd", "user", rcUnitName)
}

// renderRCUnit renders the systemd unit for the Remote Control service. Every
// path that goes into the unit must pass config.SafePath, because systemd and
// tmux would otherwise split or reinterpret it.
func renderRCUnit(env *step.Env) (string, error) {
	claude := claudeBin(env)
	if claude == "" {
		return "", errors.New("claude is not installed")
	}
	if strings.HasPrefix(claude, "/mnt/") {
		return "", fmt.Errorf("claude resolves to the Windows copy %s — install Claude Code inside WSL", claude)
	}
	tmux := env.Host.LookPath("tmux")
	if tmux == "" {
		return "", errors.New("tmux is not installed")
	}
	home := env.Host.HomeDir()
	data := map[string]string{
		"Path":   strings.Join(uniq(filepath.Dir(claude), filepath.Join(home, ".local", "bin"), "/usr/local/bin", "/usr/bin", "/bin"), ":"),
		"Dir":    rcDir(env),
		"Tmux":   tmux,
		"Claude": claude,
		"Name":   env.Config.Name,
	}
	if d := env.Host.Getenv("CLAUDE_CONFIG_DIR"); d != "" {
		data["ConfigDir"] = d
	}
	for _, k := range []string{"Dir", "Tmux", "Claude", "ConfigDir"} {
		if v := data[k]; v != "" {
			if err := config.SafePath(v); err != nil {
				return "", err
			}
		}
	}
	if err := config.SafePath(home); err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}

	tmpl, err := template.New("unit").Parse(string(Asset("roost-rc.service.tmpl")))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func uniq(items ...string) []string {
	seen := map[string]bool{}
	var out []string
	for _, it := range items {
		if !seen[it] {
			seen[it] = true
			out = append(out, it)
		}
	}
	return out
}

func rcDir(env *step.Env) string {
	if env.Config.RCDir != "" {
		return env.Config.RCDir
	}
	return env.Host.HomeDir()
}

// RemoteControlService: `claude remote-control` runs 24/7 as a user service
// inside its own tmux server.
func RemoteControlService() step.Step {
	return step.Step{
		ID: "remote-control",
		Title: func(env *step.Env) string {
			return fmt.Sprintf("Claude Remote Control always on (service roost-rc, new sessions start in %s)", rcDir(env))
		},
		Roles: boxOnly,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			want, err := renderRCUnit(env)
			if err != nil {
				return step.Failf(err.Error())
			}
			switch got := readText(env.Host, rcUnitPath(env)); {
			case got == "":
				return step.Failf("service not installed")
			case got != want:
				return step.Failf("service file is out of date")
			}
			if _, err := env.Host.Run(ctx, "systemctl", "--user", "is-active", "--quiet", "roost-rc"); err != nil {
				return step.Failf("service installed but not running — see: roost rc logs")
			}
			if _, err := env.Host.Run(ctx, "tmux", "-L", rcSocket, "has-session", "-t", "roost-rc"); err != nil {
				return step.Failf("service is up but Claude isn't running — see: roost rc logs")
			}
			return step.Pass()
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			if !env.Host.Exists(rcDir(env)) {
				return fmt.Errorf("rc_dir %s does not exist", rcDir(env))
			}
			unit, err := renderRCUnit(env)
			if err != nil {
				return err
			}
			if err := env.Host.WriteFile(rcUnitPath(env), []byte(unit), 0o644); err != nil {
				return err
			}
			for _, args := range [][]string{
				{"--user", "daemon-reload"},
				{"--user", "enable", "roost-rc"},
				{"--user", "restart", "roost-rc"},
			} {
				if _, err := env.Host.Run(ctx, "systemctl", args...); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
