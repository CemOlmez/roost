package steps

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cemolmez/roost/internal/config"
	"github.com/cemolmez/roost/internal/host/hosttest"
	"github.com/cemolmez/roost/internal/step"
	"github.com/cemolmez/roost/internal/ui"
)

func env(h *hosttest.Fake, role config.Role) *step.Env {
	return &step.Env{
		Host:   h,
		Config: config.Config{Role: role, Name: "devbox"},
		UI:     ui.New(&bytes.Buffer{}, strings.NewReader("")),
		IsWSL:  true,
	}
}

func TestCatalogIsWellFormed(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, s := range All() {
		if s.ID == "" || s.Title == nil || s.Check == nil || len(s.Roles) == 0 {
			t.Errorf("step %q is missing ID, Title, Check or Roles", s.ID)
		}
		if s.Fix == nil && s.Manual == nil {
			t.Errorf("step %q has neither Fix nor Manual — the user would get no help", s.ID)
		}
		if seen[s.ID] {
			t.Errorf("duplicate step ID %q", s.ID)
		}
		seen[s.ID] = true
		if _, ok := ByID(s.ID); !ok {
			t.Errorf("ByID(%q) not found", s.ID)
		}
	}
	for _, a := range []string{"roost-rc.service.tmpl", "roost-windows.ps1"} {
		if len(Asset(a)) == 0 {
			t.Errorf("asset %s is empty", a)
		}
	}
}

func TestWSLSystemd(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	on := hosttest.New().File("/proc/1/comm", "systemd\n")
	if r := WSLSystemd().Check(ctx, env(on, config.RoleBox)); !r.OK {
		t.Errorf("systemd running: %+v", r)
	}

	pending := hosttest.New().File("/proc/1/comm", "init\n").File(wslConf, "[boot]\nsystemd=true\n")
	if r := WSLSystemd().Check(ctx, env(pending, config.RoleBox)); r.OK || !strings.Contains(r.Detail, "restarted") {
		t.Errorf("pending restart: %+v", r)
	}

	off := hosttest.New().File("/proc/1/comm", "init\n").File(wslConf, "[user]\ndefault=dev\n").
		On("sudo install -m 0644 /home/dev/.config/roost/.staging-wsl.conf /etc/wsl.conf", "", nil)
	err := WSLSystemd().Fix(ctx, env(off, config.RoleBox))
	if !errors.Is(err, step.ErrRestartRequired) {
		t.Fatalf("Fix error = %v, want ErrRestartRequired", err)
	}
	staged := string(off.Files["/home/dev/.config/roost/.staging-wsl.conf"])
	if staged != "[user]\ndefault=dev\n\n[boot]\nsystemd=true\n" {
		t.Errorf("staged wsl.conf = %q", staged)
	}
}

func TestPackages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := hosttest.New().Bin("tmux", "apt-get").
		On("sudo apt-get update -qq", "", nil).
		On("sudo apt-get install -y -qq git curl", "", nil)
	e := env(h, config.RoleClient)
	if r := Packages().Check(ctx, e); r.OK || r.Detail != "missing: git, curl" {
		t.Errorf("Check = %+v", r)
	}
	if err := Packages().Fix(ctx, e); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if !h.Ran("sudo apt-get install -y -qq git curl") {
		t.Errorf("calls = %v", h.Calls)
	}
}

const runningStatus = `{"BackendState":"Running","Self":{"HostName":"devbox","OS":"linux"},"Peer":{"a":{"HostName":"phone","OS":"iOS"}}}`

func TestTailscaleConnected(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name, status, prefs, detail string
	}{
		{"all good", runningStatus, `{"RunSSH":true,"OperatorUser":"dev"}`, ""},
		{"logged out", `{"BackendState":"NeedsLogin","Self":{}}`, `{}`, "not logged in"},
		{"wrong name", strings.Replace(runningStatus, `"devbox"`, `"laptop"`, 1), `{"RunSSH":true,"OperatorUser":"dev"}`, `want "devbox"`},
		{"ssh off", runningStatus, `{"RunSSH":false,"OperatorUser":"dev"}`, "SSH is off"},
		{"no operator", runningStatus, `{"RunSSH":true}`, "operator"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h := hosttest.New().On("tailscale status --json", tt.status, nil).On("tailscale debug prefs", tt.prefs, nil)
			r := TailscaleConnected().Check(ctx, env(h, config.RoleBox))
			if tt.detail == "" && !r.OK {
				t.Errorf("want OK, got %+v", r)
			}
			if tt.detail != "" && (r.OK || !strings.Contains(r.Detail, tt.detail)) {
				t.Errorf("want failure containing %q, got %+v", tt.detail, r)
			}
		})
	}
}

func TestTailscaleFixUsesSetWhenRunning(t *testing.T) {
	t.Parallel()
	h := hosttest.New().On("tailscale status --json", runningStatus, nil).
		On("sudo tailscale set --ssh --hostname=devbox --operator=dev", "", nil)
	if err := TailscaleConnected().Fix(context.Background(), env(h, config.RoleBox)); err != nil {
		t.Fatalf("Fix: %v (calls %v)", err, h.Calls)
	}
}

func TestPhone(t *testing.T) {
	t.Parallel()
	h := hosttest.New().On("tailscale status --json", runningStatus, nil)
	if r := Phone().Check(context.Background(), env(h, config.RoleBox)); !r.OK {
		t.Errorf("Check = %+v", r)
	}
}

func TestRemoteControlService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	h := hosttest.New().Bin("claude", "tmux").
		On("systemctl --user daemon-reload", "", nil).
		On("systemctl --user enable roost-rc", "", nil).
		On("systemctl --user restart roost-rc", "", nil).
		On("systemctl --user is-active --quiet roost-rc", "", nil)
	e := env(h, config.RoleBox)

	if r := RemoteControlService().Check(ctx, e); r.OK || r.Detail != "service not installed" {
		t.Errorf("before fix: %+v", r)
	}
	if err := RemoteControlService().Fix(ctx, e); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	unit := string(h.Files["/home/dev/.config/systemd/user/roost-rc.service"])
	for _, want := range []string{
		`ExecStart=/usr/bin/tmux new-session -d -s roost-rc -c "/home/dev" "/usr/bin/claude remote-control --name devbox"`,
		"WorkingDirectory=/home/dev",
		"Restart=on-failure",
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("unit missing %q:\n%s", want, unit)
		}
	}
	if r := RemoteControlService().Check(ctx, e); !r.OK {
		t.Errorf("after fix: %+v", r)
	}

	// Changing the config makes the installed unit stale.
	e.Config.RCDir = "/home/dev/projects"
	if r := RemoteControlService().Check(ctx, e); r.OK || !strings.Contains(r.Detail, "out of date") {
		t.Errorf("stale unit: %+v", r)
	}
}

func TestParseWindowsHost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		out  string
		want int
	}{
		{"TASK=True HIBERNATE=0 STANDBY_AC=0x00000000", 0},
		{"TASK=False HIBERNATE=1 STANDBY_AC=0x00000384", 3},
		{"TASK=True HIBERNATE=1 STANDBY_AC=0x00000000", 1},
		{"garbage", 1},
	}
	for _, tt := range tests {
		if got := parseWindowsHost(tt.out); len(got) != tt.want {
			t.Errorf("parseWindowsHost(%q) = %v, want %d problems", tt.out, got, tt.want)
		}
	}
}

func TestWSLConfigFixKeepsUserSettings(t *testing.T) {
	t.Parallel()
	h := hosttest.New().
		On("powershell.exe -NoProfile -NonInteractive -Command $env:USERPROFILE", `C:\Users\dev`, nil).
		On(`wslpath -u C:\Users\dev`, "/mnt/c/Users/dev", nil).
		File("/mnt/c/Users/dev/.wslconfig", "[wsl2]\r\nmemory=16GB\r\n")
	e := env(h, config.RoleBox)
	if err := WSLConfig().Fix(context.Background(), e); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	got := string(h.Files["/mnt/c/Users/dev/.wslconfig"])
	if got != "[wsl2]\r\nmemory=16GB\r\nvmIdleTimeout=-1\r\n" {
		t.Errorf(".wslconfig = %q", got)
	}
	if string(h.Files["/mnt/c/Users/dev/.wslconfig.roost-backup"]) == "" {
		t.Error("no backup written")
	}
	if r := WSLConfig().Check(context.Background(), e); !r.OK {
		t.Errorf("Check after fix: %+v", r)
	}
}

func TestBIOSConfirm(t *testing.T) {
	t.Parallel()
	h := hosttest.New()
	e := env(h, config.RoleBox)
	if r := BIOS().Check(context.Background(), e); r.OK {
		t.Fatal("unconfirmed BIOS passed")
	}
	if err := Confirm(h, "bios"); err != nil {
		t.Fatal(err)
	}
	if r := BIOS().Check(context.Background(), e); !r.OK {
		t.Errorf("confirmed BIOS failed: %+v", r)
	}
}
