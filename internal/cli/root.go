// Package cli wires roost's commands (cobra) to the step engine.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host"
	"github.com/cemolmez/roost/internal/step"
	"github.com/cemolmez/roost/internal/ui"
)

// App holds roost's dependencies. Tests build one with a fake Host.
type App struct {
	Host    host.Host
	In      io.Reader
	Out     io.Writer
	Err     io.Writer
	Version string
	Commit  string
}

// errUnhealthy makes the process exit 1 without printing an extra error
// (the checklist output already says what's wrong).
var errUnhealthy = errors.New("setup incomplete")

// Execute runs roost and returns the process exit code.
func Execute(ctx context.Context, app *App, args []string) int {
	cmd := NewRootCmd(app)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(ctx)
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errUnhealthy):
		return 1
	default:
		_, _ = fmt.Fprintf(app.Err, "roost: %v\n", err)
		return 1
	}
}

// NewRootCmd builds the command tree.
func NewRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:   "roost",
		Short: "Your dev box, reachable from anywhere",
		Long: `roost turns a machine you own (a home PC with WSL, a VPS, a Raspberry Pi) into an
always-on Linux dev box you can reach from your phone (Claude Remote Control,
SSH) and your laptop — and that comes back by itself after reboots and power cuts.

Roles:
  box     the machine that runs your work 24/7
  client  a laptop you connect from`,
		Version:       app.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetIn(app.In)
	root.SetOut(app.Out)
	root.SetErr(app.Err)
	root.SetVersionTemplate("roost {{.Version}}\n")

	root.AddCommand(
		newInitCmd(app),
		newDoctorCmd(app),
		newStatusCmd(app),
		newAttachCmd(app),
		newRCCmd(app),
		newConfirmCmd(app),
		newVersionCmd(app),
	)
	return root
}

// newEnv builds the step environment for a run.
func newEnv(app *App, cfg config.Config, u *ui.UI) *step.Env {
	return &step.Env{Host: app.Host, Config: cfg, UI: u, IsWSL: host.IsWSL(app.Host)}
}

// loadConfig loads the config, treating "not set up" as an empty config.
func loadConfig(app *App) (config.Config, bool, error) {
	cfg, err := config.Load(app.Host)
	if errors.Is(err, config.ErrNotFound) {
		return cfg, false, nil
	}
	return cfg, err == nil, err
}

// guessRole suggests a role: laptops are clients, everything else is a box.
func guessRole(ctx context.Context, app *App) config.Role {
	if host.IsWSL(app.Host) && host.IsLaptop(ctx, app.Host) {
		return config.RoleClient
	}
	return config.RoleBox
}
