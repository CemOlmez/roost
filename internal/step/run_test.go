package step

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host/hosttest"
	"github.com/cemolmez/roost/internal/ui"
)

func newEnv(role config.Role, wsl bool, answers string) (*Env, *bytes.Buffer) {
	var out bytes.Buffer
	return &Env{
		Host:   hosttest.New(),
		Config: config.Config{Role: role, Name: "devbox"},
		UI:     ui.New(&out, strings.NewReader(answers)),
		IsWSL:  wsl,
	}, &out
}

func title(s string) func(*Env) string { return func(*Env) string { return s } }

// flipStep fails until fixed.
func flipStep(id string, fixErr error) (Step, *bool) {
	fixed := false
	return Step{
		ID: id, Title: title(id), Roles: []config.Role{config.RoleBox},
		Check: func(context.Context, *Env) Result {
			if fixed {
				return Pass()
			}
			return Failf(id + " is broken")
		},
		Fix: func(context.Context, *Env) error {
			if fixErr != nil {
				return fixErr
			}
			fixed = true
			return nil
		},
	}, &fixed
}

func TestRunCheckModeNeverFixes(t *testing.T) {
	t.Parallel()
	env, out := newEnv(config.RoleBox, false, "")
	s, fixed := flipStep("a", nil)
	rep := Run(context.Background(), env, []Step{s}, ModeCheck)
	if *fixed {
		t.Fatal("ModeCheck ran Fix")
	}
	if rep.Count(StatusBroken) != 1 || rep.Healthy() {
		t.Errorf("report = %+v", rep)
	}
	if !strings.Contains(out.String(), "a is broken") {
		t.Errorf("output missing detail:\n%s", out)
	}
}

func TestRunFixModeFixesOnYes(t *testing.T) {
	t.Parallel()
	env, _ := newEnv(config.RoleBox, false, "y\n")
	s, fixed := flipStep("a", nil)
	rep := Run(context.Background(), env, []Step{s}, ModeFix)
	if !*fixed || !rep.Healthy() {
		t.Errorf("fixed=%v report=%+v", *fixed, rep)
	}
}

func TestRunFixModeRespectsNo(t *testing.T) {
	t.Parallel()
	env, _ := newEnv(config.RoleBox, false, "n\n")
	s, fixed := flipStep("a", nil)
	Run(context.Background(), env, []Step{s}, ModeFix)
	if *fixed {
		t.Error("fixed after answering no")
	}
}

func TestRunFixFailureIsReported(t *testing.T) {
	t.Parallel()
	env, out := newEnv(config.RoleBox, false, "y\n")
	s, _ := flipStep("a", fmt.Errorf("apt exploded"))
	rep := Run(context.Background(), env, []Step{s}, ModeFix)
	if rep.Outcomes[0].Status != StatusBroken || !strings.Contains(out.String(), "apt exploded") {
		t.Errorf("report=%+v out=%s", rep, out)
	}
}

func TestRunStopsOnRestartRequired(t *testing.T) {
	t.Parallel()
	env, _ := newEnv(config.RoleBox, false, "y\ny\n")
	first, _ := flipStep("first", fmt.Errorf("wsl --shutdown needed: %w", ErrRestartRequired))
	second, secondFixed := flipStep("second", nil)
	rep := Run(context.Background(), env, []Step{first, second}, ModeFix)
	if !rep.Stopped || len(rep.Outcomes) != 1 || *secondFixed {
		t.Errorf("expected stop after first step; report=%+v", rep)
	}
}

func TestRunSkipsByRoleAndWSL(t *testing.T) {
	t.Parallel()
	env, _ := newEnv(config.RoleClient, false, "")
	boxOnly, _ := flipStep("box-only", nil)
	wslOnly := Step{ID: "wsl", Title: title("wsl"), Roles: []config.Role{config.RoleClient}, WSLOnly: true,
		Check: func(context.Context, *Env) Result { return Pass() }}
	rep := Run(context.Background(), env, []Step{boxOnly, wslOnly}, ModeCheck)
	if len(rep.Outcomes) != 0 {
		t.Errorf("expected all skipped, got %+v", rep.Outcomes)
	}
}

func TestRunManualStep(t *testing.T) {
	t.Parallel()
	env, out := newEnv(config.RoleBox, false, "")
	s := Step{ID: "bios", Title: title("BIOS"), Roles: []config.Role{config.RoleBox},
		Check:  func(context.Context, *Env) Result { return Failf("") },
		Manual: func(context.Context, *Env) []string { return []string{"press Del"} }}
	rep := Run(context.Background(), env, []Step{s}, ModeFix)
	if rep.Count(StatusManual) != 1 || rep.Healthy() || !strings.Contains(out.String(), "press Del") {
		t.Errorf("report=%+v out=%s", rep, out)
	}
	if rep.Stopped {
		t.Error("manual step stopped the run")
	}
}
