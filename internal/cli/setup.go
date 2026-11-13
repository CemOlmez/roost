package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/step"
	"github.com/cemolmez/roost/internal/steps"
	"github.com/cemolmez/roost/internal/ui"
)

func newInitCmd(app *App) *cobra.Command {
	var (
		flags config.Config
		role  string
		yes   bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Set up this machine (safe to re-run: it only fixes what's missing)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			u := ui.New(app.Out, app.In)
			u.Yes = yes

			cfg, _, err := loadConfig(app)
			if err != nil {
				return err
			}
			if role != "" {
				cfg.Role = config.Role(role)
			}
			if flags.Name != "" {
				cfg.Name = flags.Name
			}
			if flags.Box != "" {
				cfg.Box = flags.Box
			}
			if flags.RCDir != "" {
				cfg.RCDir = flags.RCDir
			}

			u.Heading("roost init")
			if cfg.Role == "" {
				u.Println("  What is this machine?")
				u.Printf("    %s     runs your work 24/7 (home PC, VPS, Pi) — your phone connects to it\n", u.Bold("box"))
				u.Printf("    %s  a laptop you connect from\n", u.Bold("client"))
				cfg.Role = config.Role(strings.ToLower(u.Ask("Role (box/client)?", string(guessRole(ctx, app)))))
			}
			if cfg.Name == "" {
				def := config.DefaultBoxName
				if cfg.Role == config.RoleClient {
					def = "laptop"
				}
				cfg.Name = strings.ToLower(u.Ask("Name for this machine on your private network?", def))
			}
			if cfg.Role == config.RoleClient && cfg.Box == "" {
				cfg.Box = strings.ToLower(u.Ask("Name of your box?", config.DefaultBoxName))
			}
			if err := config.Save(app.Host, cfg); err != nil {
				return err
			}
			u.OK(fmt.Sprintf("saved %s (role=%s, name=%s)", config.Path(app.Host), cfg.Role, cfg.Name))

			u.Heading("Setting up this machine as a " + string(cfg.Role))
			rep := step.Run(ctx, newEnv(app, cfg, u), steps.All(), step.ModeFix)
			summarize(u, rep, step.ModeFix)
			if rep.Healthy() {
				nextSteps(u, cfg)
				return nil
			}
			return errUnhealthy
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "box or client")
	cmd.Flags().StringVar(&flags.Name, "name", "", "this machine's name on the tailnet")
	cmd.Flags().StringVar(&flags.Box, "box", "", "client: the name of the box to connect to")
	cmd.Flags().StringVar(&flags.RCDir, "rc-dir", "", "box: where Remote Control starts new sessions (default: home)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "accept every default and fix without asking")
	return cmd
}

func newDoctorCmd(app *App) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check everything, change nothing (exit 1 if anything needs attention)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			cfg, ok, err := loadConfig(app)
			if err != nil {
				return err
			}
			u := ui.New(app.Out, app.In)
			if asJSON {
				u = ui.New(discard{}, app.In)
			}
			if !ok {
				cfg.Role = guessRole(ctx, app)
				cfg.Name = config.DefaultBoxName
				if cfg.Role == config.RoleClient {
					cfg.Name = "laptop"
				}
				u.Println(u.Dim(fmt.Sprintf("  (not set up yet — checking as a %s named %q; run roost init to choose)", cfg.Role, cfg.Name)))
			}
			u.Heading(fmt.Sprintf("roost doctor — %s %q", cfg.Role, cfg.Name))
			rep := step.Run(ctx, newEnv(app, cfg, u), steps.All(), step.ModeCheck)
			if asJSON {
				enc := json.NewEncoder(app.Out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(rep); err != nil {
					return err
				}
			} else {
				summarize(u, rep, step.ModeCheck)
			}
			if !rep.Healthy() {
				return errUnhealthy
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the result as JSON")
	return cmd
}

func newConfirmCmd(app *App) *cobra.Command {
	return &cobra.Command{
		Use:       "confirm <step>",
		Short:     "Mark a manual step as done when roost can't detect it (e.g. bios)",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bios"},
		RunE: func(_ *cobra.Command, args []string) error {
			s, ok := steps.ByID(args[0])
			if !ok || s.Fix != nil {
				return fmt.Errorf("%q is not a manual step (try: roost confirm bios)", args[0])
			}
			if err := steps.Confirm(app.Host, s.ID); err != nil {
				return err
			}
			ui.New(app.Out, app.In).OK("marked " + s.ID + " as done")
			return nil
		},
	}
}

func summarize(u *ui.UI, rep step.Report, mode step.Mode) {
	u.Println()
	ok, broken, manual := rep.Count(step.StatusOK), rep.Count(step.StatusBroken), rep.Count(step.StatusManual)
	switch {
	case rep.Stopped:
		u.Println(u.Bold("Stopped: a restart is needed (see above). Afterwards run: roost init"))
	case broken == 0 && manual == 0:
		u.Println(u.Bold(fmt.Sprintf("All set. %d checks passed.", ok)))
	default:
		u.Printf("%d ok · %d to fix · %d need you\n", ok, broken, manual)
		if mode == step.ModeCheck && broken > 0 {
			u.Println("Run " + u.Bold("roost init") + " to fix what can be fixed automatically.")
		}
	}
}

func nextSteps(u *ui.UI, cfg config.Config) {
	u.Println()
	switch cfg.Role {
	case config.RoleBox:
		u.Println("Next: on your phone, open the Claude app → Code. " + cfg.Name + " is listed — open it and start typing.")
		u.Println("      To see the Remote Control screen (QR code, sessions): roost rc attach")
	case config.RoleClient:
		u.Println("Next: roost attach   — opens your persistent terminal on " + cfg.Box)
	}
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
