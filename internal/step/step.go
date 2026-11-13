// Package step is the engine behind `roost init` and `roost doctor`.
//
// A Step is one checkable item ("Tailscale SSH is on"). Every step can be
// checked; most can also be fixed automatically. Steps that need a person
// (BIOS, logging in) provide instructions instead. doctor runs Check only;
// init runs Check and offers Fix for anything broken.
package step

import (
	"context"
	"errors"
	"slices"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host"
	"github.com/cemolmez/roost/internal/ui"
)

// Env is what a step works with.
type Env struct {
	Host   host.Host
	Config config.Config
	UI     *ui.UI
	// IsWSL is computed once per run.
	IsWSL bool
}

// Result is the outcome of a Check.
type Result struct {
	OK bool
	// Detail explains a failure ("Tailscale SSH is off").
	Detail string
}

// Pass is a successful Result.
func Pass() Result { return Result{OK: true} }

// Failf is a failed Result with an explanation.
func Failf(detail string) Result { return Result{Detail: detail} }

// Step is one item of the setup checklist.
type Step struct {
	// ID is a stable kebab-case identifier ("tailscale-ssh").
	ID string
	// Title describes the desired state, phrased as done ("Tailscale SSH is on").
	Title func(*Env) string
	// Roles this step applies to.
	Roles []config.Role
	// WSLOnly skips the step outside WSL.
	WSLOnly bool
	// Check inspects the machine. It must not change anything.
	Check func(context.Context, *Env) Result
	// Fix makes Check pass. nil means the step needs the user (see Manual).
	Fix func(context.Context, *Env) error
	// Manual returns instructions for steps the user has to do themselves.
	Manual func(context.Context, *Env) []string
}

// ErrRestartRequired is returned by a Fix that only takes effect after the
// environment restarts (e.g. enabling systemd in WSL). The run stops there.
var ErrRestartRequired = errors.New("restart required")

// Applies reports whether s is relevant on this machine.
func (s Step) Applies(env *Env) bool {
	if s.WSLOnly && !env.IsWSL {
		return false
	}
	return slices.Contains(s.Roles, env.Config.Role)
}

// Status is the final state of a step after a run.
type Status int

const (
	// StatusOK means the check passes.
	StatusOK Status = iota
	// StatusBroken means the check fails and roost can fix it.
	StatusBroken
	// StatusManual means the user has to act.
	StatusManual
	// StatusSkipped means the step does not apply here.
	StatusSkipped
)

func (s Status) String() string {
	return [...]string{"ok", "broken", "manual", "skipped"}[s]
}
