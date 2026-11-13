package steps

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"text/template"

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
		Title: title("Claude Code logged in with a Pro/Max/Team plan (API keys can't use Remote Control)"),
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

const rcUnitName = "roost-rc.service"

func rcUnitPath(env *step.Env) string {
	return filepath.Join(env.Host.HomeDir(), ".config", "systemd", "user", rcUnitName)
}

// renderRCUnit renders the systemd unit for the Remote Control service.
func renderRCUnit(env *step.Env) (string, error) {
	tmpl, err := template.New("unit").Parse(string(Asset("roost-rc.service.tmpl")))
	if err != nil {
		return "", err
	}
	tmux := env.Host.LookPath("tmux")
	if tmux == "" {
		tmux = "/usr/bin/tmux"
	}
	home := env.Host.HomeDir()
	data := map[string]string{
		"Path":   filepath.Join(home, ".local", "bin") + ":/usr/local/bin:/usr/bin:/bin",
		"Dir":    rcDir(env),
		"Tmux":   tmux,
		"Claude": claudeBin(env),
		"Name":   env.Config.Name,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func rcDir(env *step.Env) string {
	if env.Config.RCDir != "" {
		return env.Config.RCDir
	}
	return env.Host.HomeDir()
}

// RemoteControlService: `claude remote-control` runs 24/7 as a user service inside tmux.
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
			return step.Pass()
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			if claudeBin(env) == "" {
				return fmt.Errorf("install Claude Code first")
			}
			unit, err := renderRCUnit(env)
			if err != nil {
				return err
			}
			if err := env.Host.WriteFile(rcUnitPath(env), []byte(unit), 0o644); err != nil {
				return err
			}
			if _, err := env.Host.Run(ctx, "systemctl", "--user", "daemon-reload"); err != nil {
				return err
			}
			if _, err := env.Host.Run(ctx, "systemctl", "--user", "enable", "roost-rc"); err != nil {
				return err
			}
			_, err = env.Host.Run(ctx, "systemctl", "--user", "restart", "roost-rc")
			return err
		},
	}
}
