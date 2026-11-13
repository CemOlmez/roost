package config

import (
	"errors"
	"testing"

	"github.com/cemolmez/roost/internal/host/hosttest"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()
	h := hosttest.New()
	want := Config{Role: RoleClient, Name: "laptop", Box: "devbox"}
	if err := Save(h, want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(h)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != want {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
	if Path(h) != "/home/dev/.config/roost/config.toml" {
		t.Errorf("Path() = %q", Path(h))
	}
}

func TestLoadMissing(t *testing.T) {
	t.Parallel()
	if _, err := Load(hosttest.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() error = %v, want ErrNotFound", err)
	}
}

func TestXDGConfigHome(t *testing.T) {
	t.Parallel()
	h := hosttest.New()
	h.Env["XDG_CONFIG_HOME"] = "/tmp/xdg"
	if got := Path(h); got != "/tmp/xdg/roost/config.toml" {
		t.Errorf("Path() = %q", got)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		c    Config
		ok   bool
	}{
		{"box", Config{Role: RoleBox, Name: "devbox"}, true},
		{"client", Config{Role: RoleClient, Name: "laptop-1", Box: "devbox"}, true},
		{"bad role", Config{Role: "server", Name: "devbox"}, false},
		{"uppercase name", Config{Role: RoleBox, Name: "DevBox"}, false},
		{"space in name", Config{Role: RoleBox, Name: "dev box"}, false},
		{"empty name", Config{Role: RoleBox}, false},
		{"bad box", Config{Role: RoleClient, Name: "laptop", Box: "my box"}, false},
		{"rc_dir ok", Config{Role: RoleBox, Name: "devbox", RCDir: "/home/dev/projects"}, true},
		{"rc_dir relative", Config{Role: RoleBox, Name: "devbox", RCDir: "~/code"}, false},
		{"rc_dir space", Config{Role: RoleBox, Name: "devbox", RCDir: "/home/my dev"}, false},
		{"rc_dir newline", Config{Role: RoleBox, Name: "devbox", RCDir: "/home/dev\nExecStartPre=/bin/x"}, false},
		{"rc_dir percent", Config{Role: RoleBox, Name: "devbox", RCDir: "/home/%h"}, false},
	}
	for _, tt := range tests {
		if err := tt.c.Validate(); (err == nil) != tt.ok {
			t.Errorf("%s: Validate() = %v, want ok=%v", tt.name, err, tt.ok)
		}
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	t.Parallel()
	h := hosttest.New().File("/home/dev/.config/roost/config.toml", "role = \"server\"\nname = \"devbox\"\n")
	if _, err := Load(h); err == nil || errors.Is(err, ErrNotFound) {
		t.Errorf("Load() error = %v, want a validation error", err)
	}
}
