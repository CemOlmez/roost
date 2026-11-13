package steps

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cemolmez/roost/internal/host"
	"github.com/cemolmez/roost/internal/ini"
	"github.com/cemolmez/roost/internal/step"
)

const wslConf = "/etc/wsl.conf"

// WSLSystemd: systemd runs inside WSL, so background services can start on their own.
func WSLSystemd() step.Step {
	return step.Step{
		ID:      "wsl-systemd",
		Title:   title("WSL runs systemd (needed for background services)"),
		Roles:   both,
		WSLOnly: true,
		Check: func(_ context.Context, env *step.Env) step.Result {
			if host.HasSystemd(env.Host) {
				return step.Pass()
			}
			if v, _ := ini.Get(readText(env.Host, wslConf), "boot", "systemd"); v == "true" {
				return step.Failf("enabled in " + wslConf + " but WSL hasn't restarted yet")
			}
			return step.Failf("not enabled in " + wslConf)
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			content := ini.Set(readText(env.Host, wslConf), "boot", "systemd", "true")
			if err := sudoWriteFile(ctx, env, wslConf, content); err != nil {
				return err
			}
			return fmt.Errorf("systemd enabled — restart WSL: in PowerShell run  wsl --shutdown, reopen Ubuntu, then run  roost init  again: %w",
				step.ErrRestartRequired)
		},
	}
}

var basePackages = []string{"tmux", "git", "curl"}

// Packages: the command-line tools roost relies on.
func Packages() step.Step {
	return step.Step{
		ID:    "packages",
		Title: title("Base packages installed: " + strings.Join(basePackages, ", ")),
		Roles: both,
		Check: func(_ context.Context, env *step.Env) step.Result {
			if m := missing(env.Host, basePackages); len(m) > 0 {
				return step.Failf("missing: " + strings.Join(m, ", "))
			}
			return step.Pass()
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			if env.Host.LookPath("apt-get") == "" {
				return errors.New("this distro has no apt-get — install these with your package manager: " +
					strings.Join(missing(env.Host, basePackages), " "))
			}
			if err := sudo(ctx, env, "apt-get", "update", "-qq"); err != nil {
				return err
			}
			args := append([]string{"apt-get", "install", "-y", "-qq"}, missing(env.Host, basePackages)...)
			return sudo(ctx, env, args...)
		},
	}
}

func missing(h host.Host, bins []string) []string {
	var out []string
	for _, b := range bins {
		if h.LookPath(b) == "" {
			out = append(out, b)
		}
	}
	return out
}

// Linger: the user's systemd services keep running with no terminal open.
func Linger() step.Step {
	return step.Step{
		ID:    "linger",
		Title: title("Your services run even when no terminal is open (systemd linger)"),
		Roles: boxOnly,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			out, err := env.Host.Run(ctx, "loginctl", "show-user", env.Host.Username(), "-p", "Linger", "--value")
			if err == nil && out == "yes" {
				return step.Pass()
			}
			return step.Failf("linger is off")
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			return sudo(ctx, env, "loginctl", "enable-linger", env.Host.Username())
		},
	}
}
