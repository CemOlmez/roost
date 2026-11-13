package step

import (
	"context"
	"errors"
	"fmt"
)

// Mode selects what Run does with failing steps.
type Mode int

const (
	// ModeCheck only reports (roost doctor).
	ModeCheck Mode = iota
	// ModeFix offers to fix failing steps (roost init).
	ModeFix
)

// Outcome is one step's result in a Report.
type Outcome struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status Status `json:"-"`
	State  string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Report summarises a run.
type Report struct {
	Outcomes []Outcome `json:"steps"`
	// Stopped is set when a fix needs a restart before continuing.
	Stopped bool `json:"stopped,omitempty"`
}

// Count returns how many outcomes have status s.
func (r Report) Count(s Status) int {
	n := 0
	for _, o := range r.Outcomes {
		if o.Status == s {
			n++
		}
	}
	return n
}

// Healthy reports whether nothing is broken or waiting on the user.
func (r Report) Healthy() bool {
	return !r.Stopped && r.Count(StatusBroken) == 0 && r.Count(StatusManual) == 0
}

// Run checks every step in order, printing as it goes. In ModeFix it asks
// before fixing each broken step and re-checks afterwards.
func Run(ctx context.Context, env *Env, steps []Step, mode Mode) Report {
	var rep Report
	for _, s := range steps {
		if !s.Applies(env) {
			continue
		}
		o, restart := runOne(ctx, env, s, mode)
		o.State = o.Status.String()
		rep.Outcomes = append(rep.Outcomes, o)
		if restart {
			rep.Stopped = true
			break
		}
	}
	return rep
}

// runOne checks (and in ModeFix, fixes) one step. restart reports that a
// fix succeeded but needs a restart before anything else can be checked.
func runOne(ctx context.Context, env *Env, s Step, mode Mode) (o Outcome, restart bool) {
	u := env.UI
	title := s.Title(env)
	o = Outcome{ID: s.ID, Title: title}

	res := s.Check(ctx, env)
	if res.OK {
		u.OK(title)
		o.Status = StatusOK
		return o, false
	}

	if s.Fix == nil {
		u.Warn(title + "  " + u.Bold("[your action]"))
		if s.Manual != nil {
			for _, line := range s.Manual(ctx, env) {
				u.Hint(line)
			}
		}
		o.Status, o.Detail = StatusManual, res.Detail
		return o, false
	}

	u.Fail(title)
	if res.Detail != "" {
		u.Hint(res.Detail)
	}
	o.Status, o.Detail = StatusBroken, res.Detail
	if mode == ModeCheck || !u.Confirm("Fix this now?", true) {
		return o, false
	}

	if err := s.Fix(ctx, env); err != nil {
		if errors.Is(err, ErrRestartRequired) {
			u.Warn(err.Error())
			o.Detail = err.Error()
			return o, true
		}
		u.Fail(fmt.Sprintf("fix failed: %v", err))
		o.Detail = err.Error()
		return o, false
	}

	if res := s.Check(ctx, env); res.OK {
		u.OK("fixed")
		o.Status, o.Detail = StatusOK, ""
		return o, false
	} else if res.Detail != "" {
		o.Detail = res.Detail
	}
	u.Fail("still not right" + suffix(o.Detail))
	return o, false
}

func suffix(detail string) string {
	if detail == "" {
		return ""
	}
	return ": " + detail
}
