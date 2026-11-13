package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestIsNonTerminalFile(t *testing.T) {
	t.Parallel()
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer devnull.Close()
	if !IsNonTerminalFile(devnull) {
		t.Error("/dev/null treated as a terminal")
	}
	if IsNonTerminalFile(strings.NewReader("")) {
		t.Error("a non-file reader must not be reported (tests feed answers this way)")
	}
}

func TestPromptsAndYes(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	u := New(&out, strings.NewReader("n\n\nbox\n"))
	if u.Confirm("ok?", true) {
		t.Error("answered n, got yes")
	}
	if !u.Confirm("ok?", true) {
		t.Error("empty answer should take the default")
	}
	if got := u.Ask("role?", "client"); got != "box" {
		t.Errorf("Ask = %q", got)
	}
	u.Yes = true
	if !u.Confirm("ok?", true) || u.Ask("name?", "devbox") != "devbox" {
		t.Error("--yes should take defaults without reading")
	}
	if strings.Contains(out.String(), "\x1b[") {
		t.Error("colors written to a non-terminal")
	}
}
