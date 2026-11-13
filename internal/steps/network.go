package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cemolmez/roost/internal/step"
	"github.com/cemolmez/roost/internal/tailscale"
)

// TailscaleInstalled: the private network between your devices.
func TailscaleInstalled() step.Step {
	return step.Step{
		ID:    "tailscale-installed",
		Title: title("Tailscale installed (private network between your devices)"),
		Roles: both,
		Check: func(_ context.Context, env *step.Env) step.Result {
			if env.Host.LookPath("tailscale") != "" {
				return step.Pass()
			}
			return step.Failf("tailscale not found")
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			return shell(ctx, env, "curl -fsSL https://tailscale.com/install.sh | sh")
		},
	}
}

// TailscaleConnected: logged in, reachable under the configured name, Tailscale SSH on.
func TailscaleConnected() step.Step {
	return step.Step{
		ID: "tailscale-connected",
		Title: func(env *step.Env) string {
			return fmt.Sprintf("Tailscale connected as %q with Tailscale SSH on", env.Config.Name)
		},
		Roles: both,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			st, err := tailscale.GetStatus(ctx, env.Host)
			if err != nil {
				return step.Failf("can't read tailscale status: " + err.Error())
			}
			if !st.Running() {
				return step.Failf("not connected (state: " + st.BackendState + ")")
			}
			if got := st.Self.MagicDNSName(); !strings.EqualFold(got, env.Config.Name) {
				if strings.EqualFold(st.Self.HostName, env.Config.Name) {
					return step.Failf(fmt.Sprintf("the name %q is taken by another device, so this one is %q — remove the old device at https://login.tailscale.com/admin/machines, then run roost init", env.Config.Name, got))
				}
				return step.Failf(fmt.Sprintf("this machine is called %q on the tailnet, want %q", got, env.Config.Name))
			}
			p, err := tailscale.GetPrefs(ctx, env.Host)
			if err != nil {
				return step.Failf("can't read tailscale prefs: " + err.Error())
			}
			if !p.RunSSH {
				return step.Failf("Tailscale SSH is off")
			}
			return step.Pass()
		},
		Fix: func(ctx context.Context, env *step.Env) error {
			// `set` works in any state and keeps the user's other settings;
			// `up` with flags would refuse if non-default prefs exist.
			if err := sudo(ctx, env, "tailscale", "set", "--ssh", "--hostname="+env.Config.Name); err != nil {
				return err
			}
			if st, err := tailscale.GetStatus(ctx, env.Host); err == nil && st.Running() {
				return nil
			}
			env.UI.Hint("A login link will appear — open it and sign in with the same account as your phone.")
			return sudo(ctx, env, "tailscale", "up")
		},
	}
}

// Phone: a phone is on the tailnet, for Termius/SSH access on the go.
func Phone() step.Step {
	return step.Step{
		ID:    "phone",
		Title: title("A phone is on your tailnet (for Termius / SSH on the go)"),
		Roles: both,
		Check: func(ctx context.Context, env *step.Env) step.Result {
			st, err := tailscale.GetStatus(ctx, env.Host)
			if err != nil {
				return step.Failf("")
			}
			for _, p := range st.Peers() {
				if p.IsPhone() {
					return step.Pass()
				}
			}
			return step.Failf("no phone found")
		},
		Manual: func(_ context.Context, env *step.Env) []string {
			target := env.Config.Name
			if env.Config.Box != "" {
				target = env.Config.Box
			}
			return []string{
				"On your phone install: Claude (Remote Control), Tailscale (same account), Termius (terminal).",
				fmt.Sprintf("In Termius add host %q, user %q — Tailscale SSH logs you in, no password or key.", target, env.Host.Username()),
			}
		},
	}
}
