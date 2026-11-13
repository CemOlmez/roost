package tailscale

import (
	"context"
	"os"
	"testing"

	"github.com/cemolmez/roost/internal/host/hosttest"
)

func TestParseStatus(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile("testdata/status.json")
	if err != nil {
		t.Fatal(err)
	}
	s, err := ParseStatus(b)
	if err != nil {
		t.Fatalf("ParseStatus: %v", err)
	}
	if !s.Running() {
		t.Errorf("Running() = false, want true")
	}
	if s.Self.HostName != "devbox" {
		t.Errorf("Self.HostName = %q", s.Self.HostName)
	}
	peers := s.Peers()
	if len(peers) != 2 || peers[0].HostName != "laptop" || peers[1].HostName != "phone" {
		t.Fatalf("Peers() = %+v", peers)
	}
	if !peers[1].IsPhone() || peers[0].IsPhone() {
		t.Errorf("IsPhone wrong: %+v", peers)
	}
}

func TestGetStatusLoggedOutStillParses(t *testing.T) {
	t.Parallel()
	h := hosttest.New().On("tailscale status --json", `{"BackendState":"NeedsLogin","Self":{},"Peer":null}`, os.ErrInvalid)
	s, err := GetStatus(context.Background(), h)
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if s.Running() {
		t.Errorf("Running() = true for NeedsLogin")
	}
}

func TestGetPrefs(t *testing.T) {
	t.Parallel()
	h := hosttest.New().On("tailscale debug prefs", `{"RunSSH":true,"Hostname":"devbox","OperatorUser":"dev"}`, nil)
	p, err := GetPrefs(context.Background(), h)
	if err != nil {
		t.Fatalf("GetPrefs: %v", err)
	}
	if !p.RunSSH || p.Hostname != "devbox" || p.OperatorUser != "dev" {
		t.Errorf("GetPrefs() = %+v", p)
	}
}
