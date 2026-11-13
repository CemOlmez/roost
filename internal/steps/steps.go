// Package steps is roost's setup checklist: every item `roost init` sets up
// and `roost doctor` checks, in the order they run. To add a capability, add
// a step here (see docs/ARCHITECTURE.md, "Adding a step").
package steps

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host"
	"github.com/cemolmez/roost/internal/step"
)

//go:embed assets/*
var assetFS embed.FS

// Asset returns an embedded file from assets/.
func Asset(name string) []byte {
	b, err := assetFS.ReadFile("assets/" + name)
	if err != nil {
		panic(fmt.Sprintf("steps: missing embedded asset %q", name)) // programming error, caught by tests
	}
	return b
}

var (
	both    = []config.Role{config.RoleBox, config.RoleClient}
	boxOnly = []config.Role{config.RoleBox}
)

// All returns the checklist in run order.
func All() []step.Step {
	return []step.Step{
		WSLSystemd(),
		Packages(),
		TailscaleInstalled(),
		TailscaleConnected(),
		ClaudeInstalled(),
		ClaudeLoggedIn(),
		Linger(),
		RemoteControlService(),
		WSLConfig(),
		WindowsHost(),
		WindowsAutologon(),
		BIOS(),
		Phone(),
	}
}

// ByID finds a step by its ID.
func ByID(id string) (step.Step, bool) {
	for _, s := range All() {
		if s.ID == id {
			return s, true
		}
	}
	return step.Step{}, false
}

func title(s string) func(*step.Env) string { return func(*step.Env) string { return s } }

// sudo runs a command as root, attached to the terminal so sudo can ask for a password.
func sudo(ctx context.Context, env *step.Env, args ...string) error {
	return env.Host.RunTTY(ctx, "sudo", args...)
}

// sudoWriteFile writes a root-owned file by staging it in roost's config dir
// and moving it into place with `sudo install`.
func sudoWriteFile(ctx context.Context, env *step.Env, path, content string) error {
	tmp := filepath.Join(config.Dir(env.Host), ".staging-"+filepath.Base(path))
	if err := env.Host.WriteFile(tmp, []byte(content), 0o600); err != nil {
		return err
	}
	return sudo(ctx, env, "install", "-m", "0644", tmp, path)
}

// shell runs a shell pipeline attached to the terminal (used for vendor installers).
func shell(ctx context.Context, env *step.Env, script string) error {
	return env.Host.RunTTY(ctx, "sh", "-c", script)
}

// readText returns a file's content, or "" if it does not exist.
func readText(h host.Host, path string) string {
	b, err := h.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

// confirmedPath is the marker file for a manual step the user has confirmed
// with `roost confirm <id>` (used where the state can't be detected).
func confirmedPath(h host.Host, id string) string {
	return filepath.Join(config.Dir(h), "confirmed", id)
}

// Confirm records that the user completed a manual step.
func Confirm(h host.Host, id string) error {
	return h.WriteFile(confirmedPath(h, id), []byte("confirmed\n"), 0o600)
}
