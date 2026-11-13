//go:build live

package steps

import (
	"os/exec"
	"strings"
	"testing"
)

// go test -tags live -run TestLiveEncodedCommand ./internal/steps  (WSL only; read-only)
func TestLiveEncodedCommand(t *testing.T) {
	script, err := windowsHostScript("Ubuntu", `DOM\o'brien`)
	if err != nil {
		t.Fatal(err)
	}
	// Bind the same prelude, but run a harmless body instead of the setup script.
	prelude := script[:strings.Index(script, "# roost-windows.ps1")]
	probe := prelude + `Write-Output "distro=$Distro user=$TaskUser"`
	out, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", encodePowerShell(probe)).Output()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	if got := strings.TrimSpace(strings.ReplaceAll(string(out), "\r", "")); got != `distro=Ubuntu user=DOM\o'brien` {
		t.Errorf("got %q", got)
	}
	t.Logf("encoded script length: %d chars", len(encodePowerShell(script)))
}
