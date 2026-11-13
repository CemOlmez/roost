package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host"
	"github.com/cemolmez/roost/internal/tailscale"
	"github.com/cemolmez/roost/internal/ui"
)

// tmuxSession is the persistent terminal `roost attach` opens.
const tmuxSession = "main"

func newStatusCmd(app *App) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "This machine, your devices, and Remote Control",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cfg, _, err := loadConfig(app)
			if err != nil {
				return err
			}
			st := gatherStatus(ctx, app, cfg)
			if asJSON {
				enc := json.NewEncoder(app.Out)
				enc.SetIndent("", "  ")
				return enc.Encode(st)
			}
			printStatus(ui.New(app.Out, app.In), st)
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print as JSON")
	return cmd
}

type statusView struct {
	Role          config.Role      `json:"role"`
	Name          string           `json:"name"`
	WSLDistro     string           `json:"wsl_distro,omitempty"`
	RemoteControl string           `json:"remote_control,omitempty"`
	TmuxSessions  string           `json:"tmux_sessions"`
	Tailnet       string           `json:"tailnet"`
	Devices       []tailscale.Node `json:"devices"`
}

func gatherStatus(ctx context.Context, app *App, cfg config.Config) statusView {
	v := statusView{Role: cfg.Role, Name: cfg.Name}
	if host.IsWSL(app.Host) {
		v.WSLDistro = host.WSLDistro(app.Host)
	}
	if cfg.Role == config.RoleBox {
		v.RemoteControl, _ = app.Host.Run(ctx, "systemctl", "--user", "is-active", "roost-rc")
	}
	v.TmuxSessions, _ = app.Host.Run(ctx, "tmux", "list-sessions", "-F", "#S")
	st, err := tailscale.GetStatus(ctx, app.Host)
	switch {
	case err != nil:
		v.Tailnet = "unavailable"
	case !st.Running():
		v.Tailnet = st.BackendState
	default:
		v.Tailnet = "connected"
		v.Devices = append([]tailscale.Node{st.Self}, st.Peers()...)
	}
	return v
}

func printStatus(u *ui.UI, v statusView) {
	u.Heading("This machine")
	role := string(v.Role)
	if role == "" {
		role = "not set up (run: roost init)"
	}
	u.Printf("  role: %s   name: %s\n", role, v.Name)
	if v.WSLDistro != "" {
		u.Printf("  WSL distro: %s\n", v.WSLDistro)
	}
	if v.RemoteControl != "" {
		u.Printf("  remote control: %s\n", v.RemoteControl)
	}
	sessions := v.TmuxSessions
	if sessions == "" {
		sessions = "none"
	}
	u.Printf("  terminals (tmux): %s\n", sessions)

	u.Heading("Your devices (tailnet: " + v.Tailnet + ")")
	for _, d := range v.Devices {
		dot := "○"
		if d.Online {
			dot = "●"
		}
		u.Printf("  %s %-24s %s\n", dot, d.HostName, d.OS)
	}
}

func newAttachCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Open your persistent terminal (on a client: on the box, over SSH)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, ok, err := loadConfig(app)
			if err != nil {
				return err
			}
			if ok && cfg.Role == config.RoleClient {
				box := cfg.Box
				if box == "" {
					box = config.DefaultBoxName
				}
				return app.Host.RunTTY(cmd.Context(), "ssh", "-t", app.Host.Username()+"@"+box,
					"tmux new-session -A -s "+tmuxSession)
			}
			return app.Host.RunTTY(cmd.Context(), "tmux", "new-session", "-A", "-s", tmuxSession)
		},
	}
}

func newRCCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rc",
		Short: "Manage the always-on Claude Remote Control service (box only)",
	}
	systemctl := func(action string) *cobra.Command {
		return &cobra.Command{
			Use:   action,
			Short: action + " the Remote Control service",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				if _, err := app.Host.Run(cmd.Context(), "systemctl", "--user", action, "roost-rc"); err != nil {
					return err
				}
				ui.New(app.Out, app.In).OK("roost-rc: " + action)
				return nil
			},
		}
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "attach",
			Short: "Show the Remote Control screen (QR code, sessions). Detach with Ctrl-b d",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				err := app.Host.RunTTY(cmd.Context(), "tmux", "attach-session", "-t", "roost-rc")
				if err != nil {
					return errors.New("remote control isn't running — check: roost doctor")
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "logs",
			Short: "Show recent service logs",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				return app.Host.RunTTY(cmd.Context(), "journalctl", "--user", "-u", "roost-rc", "-n", "50", "--no-pager")
			},
		},
		systemctl("start"),
		systemctl("stop"),
		systemctl("restart"),
	)
	return cmd
}

func newVersionCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		Run: func(*cobra.Command, []string) {
			_, _ = fmt.Fprintf(app.Out, "roost %s (%s)\n", app.Version, app.Commit)
		},
	}
}
