package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host/hosttest"
)

func run(t *testing.T, h *hosttest.Fake, stdin string, args ...string) (string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	app := &App{Host: h, In: strings.NewReader(stdin), Out: &out, Err: &errOut, Version: "test", Commit: "abc"}
	code := Execute(context.Background(), app, args)
	return out.String() + errOut.String(), code
}

// healthyClient is a non-WSL Linux client with everything in place.
func healthyClient() *hosttest.Fake {
	return hosttest.New().
		Bin("tmux", "git", "curl", "tailscale", "claude").
		File("/proc/sys/kernel/osrelease", "6.1.0-generic").
		File("/home/dev/.claude/.credentials.json", "{}").
		On("tailscale status --json", `{"BackendState":"Running","Self":{"HostName":"laptop","OS":"linux"},"Peer":{"p":{"HostName":"phone","OS":"android"}}}`, nil).
		On("tailscale debug prefs", `{"RunSSH":true,"OperatorUser":"dev"}`, nil)
}

func TestVersion(t *testing.T) {
	t.Parallel()
	out, code := run(t, hosttest.New(), "", "version")
	if code != 0 || out != "roost test (abc)\n" {
		t.Errorf("code=%d out=%q", code, out)
	}
}

func TestDoctorHealthyClient(t *testing.T) {
	t.Parallel()
	h := healthyClient()
	if err := config.Save(h, config.Config{Role: config.RoleClient, Name: "laptop", Box: "devbox"}); err != nil {
		t.Fatal(err)
	}
	out, code := run(t, h, "", "doctor")
	if code != 0 || !strings.Contains(out, "All set.") {
		t.Errorf("code=%d out=\n%s", code, out)
	}
	if strings.Contains(out, "WSL") {
		t.Errorf("WSL-only steps ran on plain Linux:\n%s", out)
	}
}

func TestDoctorJSONAndExitCode(t *testing.T) {
	t.Parallel()
	h := healthyClient()
	delete(h.Bins, "tmux")
	if err := config.Save(h, config.Config{Role: config.RoleClient, Name: "laptop", Box: "devbox"}); err != nil {
		t.Fatal(err)
	}
	out, code := run(t, h, "", "doctor", "--json")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	var rep struct {
		Steps []struct{ ID, Status, Detail string }
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, out)
	}
	found := false
	for _, s := range rep.Steps {
		if s.ID == "packages" && s.Status == "broken" && s.Detail == "missing: tmux" {
			found = true
		}
	}
	if !found {
		t.Errorf("packages step not reported broken: %+v", rep.Steps)
	}
}

func TestInitAsksThenSaves(t *testing.T) {
	t.Parallel()
	h := healthyClient()
	// role, name, box
	out, code := run(t, h, "client\nlaptop\ndevbox\n", "init")
	if code != 0 {
		t.Fatalf("code=%d out=\n%s", code, out)
	}
	cfg, err := config.Load(h)
	if err != nil {
		t.Fatal(err)
	}
	if cfg != (config.Config{Role: config.RoleClient, Name: "laptop", Box: "devbox"}) {
		t.Errorf("saved config = %+v", cfg)
	}
	if !strings.Contains(out, "roost attach") {
		t.Errorf("missing next steps:\n%s", out)
	}
}

func TestInitRejectsBadName(t *testing.T) {
	t.Parallel()
	out, code := run(t, healthyClient(), "", "init", "--role", "box", "--name", "My Box", "--yes")
	if code != 1 || !strings.Contains(out, "valid hostname") {
		t.Errorf("code=%d out=%s", code, out)
	}
}

func TestConfirm(t *testing.T) {
	t.Parallel()
	h := hosttest.New()
	if _, code := run(t, h, "", "confirm", "bios"); code != 0 {
		t.Fatal("confirm bios failed")
	}
	if !h.Exists("/home/dev/.config/roost/confirmed/bios") {
		t.Error("marker not written")
	}
	if out, code := run(t, h, "", "confirm", "packages"); code != 1 || !strings.Contains(out, "not a manual step") {
		t.Errorf("confirm packages: code=%d out=%s", code, out)
	}
}

func TestAttachClientUsesSSH(t *testing.T) {
	t.Parallel()
	h := hosttest.New().On("ssh -t dev@devbox tmux new-session -A -s main", "", nil)
	if err := config.Save(h, config.Config{Role: config.RoleClient, Name: "laptop", Box: "devbox"}); err != nil {
		t.Fatal(err)
	}
	if out, code := run(t, h, "", "attach"); code != 0 {
		t.Errorf("code=%d out=%s calls=%v", code, out, h.Calls)
	}
}

func TestStatusJSON(t *testing.T) {
	t.Parallel()
	h := healthyClient().On("tmux list-sessions -F #S", "main", nil)
	if err := config.Save(h, config.Config{Role: config.RoleClient, Name: "laptop", Box: "devbox"}); err != nil {
		t.Fatal(err)
	}
	out, code := run(t, h, "", "status", "--json")
	if code != 0 || !strings.Contains(out, `"tailnet": "connected"`) || !strings.Contains(out, `"phone"`) {
		t.Errorf("code=%d out=%s", code, out)
	}
}
