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

var ctx = context.Background()

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
		if s.Confirmable && s.Fix != nil {
			t.Errorf("step %q is Confirmable but also auto-fixable", s.ID)
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
	on := hosttest.New().File("/proc/1/comm", "systemd\n")
	if r := WSLSystemd().Check(ctx, env(on, config.RoleBox)); !r.OK {
		t.Errorf("systemd running: %+v", r)
	}

	pending := hosttest.New().File("/proc/1/comm", "init\n").File(wslConf, "[boot]\nsystemd=true\n")
	if r := WSLSystemd().Check(ctx, env(pending, config.RoleBox)); r.OK || !strings.Contains(r.Detail, "restarted") {
		t.Errorf("pending restart: %+v", r)
	}

	const staging = "/home/dev/.config/roost/.staging-tmp"
	off := hosttest.New().File("/proc/1/comm", "init\n").File(wslConf, "[user]\ndefault=dev\n").
		On("sudo install -T -m 0644 -o root -g root "+staging+" /etc/wsl.conf", "", nil)
	err := WSLSystemd().Fix(ctx, env(off, config.RoleBox))
	if !errors.Is(err, step.ErrRestartRequired) {
		t.Fatalf("Fix error = %v, want ErrRestartRequired", err)
	}
	if got := string(off.Written[staging]); got != "[user]\ndefault=dev\n\n[boot]\nsystemd=true\n" {
		t.Errorf("staged wsl.conf = %q", got)
	}
	if off.Exists(staging) {
		t.Error("staging file left behind")
	}
}

func TestPackages(t *testing.T) {
	t.Parallel()
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

	noApt := hosttest.New()
	if err := Packages().Fix(ctx, env(noApt, config.RoleClient)); err == nil || !strings.Contains(err.Error(), "no apt-get") {
		t.Errorf("Fix without apt-get = %v", err)
	}
}

func TestVendorInstallers(t *testing.T) {
	t.Parallel()
	h := hosttest.New().
		On("sh -c curl -fsSL https://tailscale.com/install.sh | sh", "", nil).
		On("sh -c curl -fsSL https://claude.ai/install.sh | bash", "", nil)
	e := env(h, config.RoleBox)
	for _, s := range []step.Step{TailscaleInstalled(), ClaudeInstalled()} {
		if r := s.Check(ctx, e); r.OK {
			t.Errorf("%s passed with nothing installed", s.ID)
		}
		if err := s.Fix(ctx, e); err != nil {
			t.Errorf("%s Fix: %v", s.ID, err)
		}
	}
	// claude installed by its native installer, not yet on PATH (fresh shell).
	h.File("/home/dev/.local/bin/claude", "")
	if r := ClaudeInstalled().Check(ctx, e); !r.OK {
		t.Errorf("claude in ~/.local/bin not found: %+v", r)
	}
}

func TestClaudeLoggedIn(t *testing.T) {
	t.Parallel()
	h := hosttest.New()
	e := env(h, config.RoleBox)
	if r := ClaudeLoggedIn().Check(ctx, e); r.OK {
		t.Error("passed without credentials")
	}
	h.Env["CLAUDE_CONFIG_DIR"] = "/data/claude"
	h.File("/data/claude/.credentials.json", "{}")
	if r := ClaudeLoggedIn().Check(ctx, e); !r.OK {
		t.Errorf("CLAUDE_CONFIG_DIR not honoured: %+v", r)
	}
}

func TestLinger(t *testing.T) {
	t.Parallel()
	h := hosttest.New().
		On("loginctl show-user dev -p Linger --value", "no", nil).
		On("sudo loginctl enable-linger dev", "", nil)
	e := env(h, config.RoleBox)
	if r := Linger().Check(ctx, e); r.OK {
		t.Error("passed with linger off")
	}
	if err := Linger().Fix(ctx, e); err != nil {
		t.Fatal(err)
	}
	h.On("loginctl show-user dev -p Linger --value", "yes", nil)
	if r := Linger().Check(ctx, e); !r.OK {
		t.Errorf("Check = %+v", r)
	}
}

const runningStatus = `{"BackendState":"Running","Self":{"HostName":"devbox","DNSName":"devbox.tail1.ts.net.","OS":"linux"},"Peer":{"a":{"HostName":"phone","OS":"iOS"}}}`

func TestTailscaleConnected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, status, prefs, detail string
	}{
		{"all good", runningStatus, `{"RunSSH":true}`, ""},
		{"logged out", `{"BackendState":"NeedsLogin","Self":{}}`, `{}`, "not connected"},
		{"wrong name", strings.ReplaceAll(runningStatus, "devbox", "laptop"), `{"RunSSH":true}`, `want "devbox"`},
		{"name taken", strings.Replace(runningStatus, "devbox.tail1", "devbox-1.tail1", 1), `{"RunSSH":true}`, "taken by another device"},
		{"ssh off", runningStatus, `{"RunSSH":false}`, "SSH is off"},
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

func TestTailscaleFix(t *testing.T) {
	t.Parallel()
	running := hosttest.New().On("tailscale status --json", runningStatus, nil).
		On("sudo tailscale set --ssh --hostname=devbox", "", nil)
	if err := TailscaleConnected().Fix(ctx, env(running, config.RoleBox)); err != nil {
		t.Fatalf("Fix: %v (calls %v)", err, running.Calls)
	}
	if running.Ran("sudo tailscale up") {
		t.Error("ran `tailscale up` on a node that was already up")
	}

	stopped := hosttest.New().On("tailscale status --json", `{"BackendState":"Stopped","Self":{}}`, nil).
		On("sudo tailscale set --ssh --hostname=devbox", "", nil).
		On("sudo tailscale up", "", nil)
	if err := TailscaleConnected().Fix(ctx, env(stopped, config.RoleBox)); err != nil || !stopped.Ran("sudo tailscale up") {
		t.Errorf("Fix: %v (calls %v)", err, stopped.Calls)
	}
}

func TestPhone(t *testing.T) {
	t.Parallel()
	h := hosttest.New().On("tailscale status --json", runningStatus, nil)
	if r := Phone().Check(ctx, env(h, config.RoleBox)); !r.OK {
		t.Errorf("Check = %+v", r)
	}
	none := hosttest.New().On("tailscale status --json", `{"BackendState":"Running","Self":{"HostName":"devbox"}}`, nil)
	if r := Phone().Check(ctx, env(none, config.RoleBox)); r.OK {
		t.Error("passed without a phone")
	}
}

func rcHost() *hosttest.Fake {
	return hosttest.New().Bin("claude", "tmux").
		On("systemctl --user daemon-reload", "", nil).
		On("systemctl --user enable roost-rc", "", nil).
		On("systemctl --user restart roost-rc", "", nil).
		On("systemctl --user is-active --quiet roost-rc", "", nil).
		On("tmux -L roost-rc has-session -t roost-rc", "", nil)
}

func TestRemoteControlService(t *testing.T) {
	t.Parallel()
	h := rcHost()
	e := env(h, config.RoleBox)

	if r := RemoteControlService().Check(ctx, e); r.OK || r.Detail != "service not installed" {
		t.Errorf("before fix: %+v", r)
	}
	if err := RemoteControlService().Fix(ctx, e); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	unit := string(h.Files["/home/dev/.config/systemd/user/roost-rc.service"])
	for _, want := range []string{
		`Environment="PATH=/usr/bin:/home/dev/.local/bin:/usr/local/bin:/bin"`,
		"ExecStart=/usr/bin/tmux -L roost-rc new-session -d -s roost-rc -c /home/dev /usr/bin/claude remote-control --name devbox",
		"ExecStop=/usr/bin/tmux -L roost-rc kill-session -t roost-rc",
		"WorkingDirectory=/home/dev",
		"Restart=always",
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("unit missing %q:\n%s", want, unit)
		}
	}
	if strings.Contains(unit, "CLAUDE_CONFIG_DIR") {
		t.Error("CLAUDE_CONFIG_DIR set without being configured")
	}
	if r := RemoteControlService().Check(ctx, e); !r.OK {
		t.Errorf("after fix: %+v", r)
	}

	// Changing the config makes the installed unit stale.
	e.Config.RCDir = "/home/dev/projects"
	if r := RemoteControlService().Check(ctx, e); r.OK || !strings.Contains(r.Detail, "out of date") {
		t.Errorf("stale unit: %+v", r)
	}
	if err := RemoteControlService().Fix(ctx, e); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Fix with missing rc_dir = %v", err)
	}
}

func TestRemoteControlServiceDetectsDeadClaude(t *testing.T) {
	t.Parallel()
	h := rcHost()
	e := env(h, config.RoleBox)
	if err := RemoteControlService().Fix(ctx, e); err != nil {
		t.Fatal(err)
	}
	h.On("tmux -L roost-rc has-session -t roost-rc", "", errors.New("no session"))
	if r := RemoteControlService().Check(ctx, e); r.OK || !strings.Contains(r.Detail, "Claude isn't running") {
		t.Errorf("Check = %+v", r)
	}
}

func TestRCUnitEnvironment(t *testing.T) {
	t.Parallel()
	h := rcHost()
	delete(h.Bins, "claude")
	h.Env["CLAUDE_CONFIG_DIR"] = "/data/claude"
	e := env(h, config.RoleBox)
	if _, err := renderRCUnit(e); err == nil {
		t.Error("rendered a unit without claude installed")
	}
	h.Bin("claude")
	unit, err := renderRCUnit(e)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(unit, `Environment="CLAUDE_CONFIG_DIR=/data/claude"`) {
		t.Errorf("CLAUDE_CONFIG_DIR not passed:\n%s", unit)
	}

	h.Home = "/home/my user"
	if _, err := renderRCUnit(e); err == nil {
		t.Error("accepted a home directory with a space")
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

func TestWindowsHostFix(t *testing.T) {
	t.Parallel()
	script, err := windowsHostScript("Ubuntu", `PC\dev`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(script, "$Distro = 'Ubuntu'\n$TaskUser = 'PC\\dev'\n") {
		t.Errorf("script prelude = %q", script[:60])
	}
	launch := "Start-Process powershell -Verb RunAs -Wait -ArgumentList " +
		"'-NoProfile','-ExecutionPolicy','Bypass','-EncodedCommand','" + encodePowerShell(script) + "'"
	h := hosttest.New().
		On(`powershell.exe -NoProfile -NonInteractive -Command "$env:USERDOMAIN\$env:USERNAME"`, `PC\dev`, nil).
		On("powershell.exe -NoProfile -NonInteractive -Command "+launch, "", nil)
	if err := WindowsHost().Fix(ctx, env(h, config.RoleBox)); err != nil {
		t.Fatalf("Fix: %v (calls %v)", err, h.Calls)
	}
	for _, c := range h.Calls {
		if strings.Contains(c, "roost-windows.ps1") || strings.Contains(c, "TEMP") {
			t.Errorf("script went through a file: %s", c)
		}
	}
}

func TestWindowsHostScriptRejectsUnsafeDistro(t *testing.T) {
	t.Parallel()
	if _, err := windowsHostScript("Ubuntu 24.04 --exec", `PC\dev`); err == nil {
		t.Error("accepted an unsafe distro name")
	}
	s, err := windowsHostScript("Ubuntu", `PC\o'brien`)
	if err != nil || !strings.Contains(s, `$TaskUser = 'PC\o''brien'`) {
		t.Errorf("apostrophe not escaped: %v %q", err, s[:60])
	}
}

func TestEncodePowerShell(t *testing.T) {
	t.Parallel()
	// "hi" in UTF-16LE is 68 00 69 00.
	if got := encodePowerShell("hi"); got != "aABpAA==" {
		t.Errorf("encodePowerShell(hi) = %q", got)
	}
}

func TestWSLConfigFixKeepsUserSettings(t *testing.T) {
	t.Parallel()
	h := hosttest.New().
		On("powershell.exe -NoProfile -NonInteractive -Command $env:USERPROFILE", `C:\Users\dev`, nil).
		On(`wslpath -u C:\Users\dev`, "/mnt/c/Users/dev", nil).
		File("/mnt/c/Users/dev/.wslconfig", "[wsl2]\r\nmemory=16GB\r\n")
	e := env(h, config.RoleBox)
	if err := WSLConfig().Fix(ctx, e); err != nil {
		t.Fatalf("Fix: %v", err)
	}
	if got := string(h.Files["/mnt/c/Users/dev/.wslconfig"]); got != "[wsl2]\r\nmemory=16GB\r\nvmIdleTimeout=-1\r\n" {
		t.Errorf(".wslconfig = %q", got)
	}
	if string(h.Files["/mnt/c/Users/dev/.wslconfig.roost-backup"]) == "" {
		t.Error("no backup written")
	}
	if r := WSLConfig().Check(ctx, e); !r.OK {
		t.Errorf("Check after fix: %+v", r)
	}
}

func TestWindowsAutologon(t *testing.T) {
	t.Parallel()
	const q = `powershell.exe -NoProfile -NonInteractive -Command (Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon').AutoAdminLogon`
	h := hosttest.New().On(q, "0", nil)
	e := env(h, config.RoleBox)
	if r := WindowsAutologon().Check(ctx, e); r.OK {
		t.Error("passed with auto-login off")
	}
	h.On(q, "1", nil)
	if r := WindowsAutologon().Check(ctx, e); !r.OK {
		t.Errorf("Check = %+v", r)
	}
}

func TestBIOSConfirm(t *testing.T) {
	t.Parallel()
	h := hosttest.New()
	e := env(h, config.RoleBox)
	if r := BIOS().Check(ctx, e); r.OK {
		t.Fatal("unconfirmed BIOS passed")
	}
	if err := Confirm(h, "bios"); err != nil {
		t.Fatal(err)
	}
	if r := BIOS().Check(ctx, e); !r.OK {
		t.Errorf("confirmed BIOS failed: %+v", r)
	}
}
