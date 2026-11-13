// Package ui prints roost's checklist output and asks the user questions.
// Colors are used only when writing to a terminal.
package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// UI writes human output and reads answers.
type UI struct {
	out   io.Writer
	in    *bufio.Reader
	color bool
	// Yes answers every question with its default (for --yes and scripts).
	Yes bool
}

// New returns a UI. Colors are enabled when out is a terminal and NO_COLOR is unset.
func New(out io.Writer, in io.Reader) *UI {
	return &UI{out: out, in: bufio.NewReader(in), color: isTerminal(out) && os.Getenv("NO_COLOR") == ""}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}

const (
	reset  = "\x1b[0m"
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	red    = "\x1b[31m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
)

func (u *UI) paint(code, s string) string {
	if !u.color {
		return s
	}
	return code + s + reset
}

// Bold returns s in bold.
func (u *UI) Bold(s string) string { return u.paint(bold, s) }

// Dim returns s dimmed.
func (u *UI) Dim(s string) string { return u.paint(dim, s) }

// Println prints a line.
func (u *UI) Println(a ...any) { _, _ = fmt.Fprintln(u.out, a...) }

// Printf prints formatted text.
func (u *UI) Printf(format string, a ...any) { _, _ = fmt.Fprintf(u.out, format, a...) }

// Heading prints a section heading.
func (u *UI) Heading(s string) { u.Printf("\n%s\n", u.Bold(s)) }

// OK prints a passed check.
func (u *UI) OK(s string) { u.Printf("  %s %s\n", u.paint(green, "✓"), s) }

// Fail prints a failed check.
func (u *UI) Fail(s string) { u.Printf("  %s %s\n", u.paint(red, "✗"), s) }

// Warn prints something that needs the user's attention.
func (u *UI) Warn(s string) { u.Printf("  %s %s\n", u.paint(yellow, "!"), s) }

// Skip prints a check that does not apply here.
func (u *UI) Skip(s string) { u.Printf("  %s %s\n", u.Dim("-"), u.Dim(s)) }

// Hint prints an indented detail line under a check.
func (u *UI) Hint(s string) { u.Printf("      %s\n", u.Dim(s)) }

// Confirm asks a yes/no question. def is the answer for Enter or --yes.
func (u *UI) Confirm(question string, def bool) bool {
	if u.Yes {
		return def
	}
	prompt := "[y/N]"
	if def {
		prompt = "[Y/n]"
	}
	ans := u.read(fmt.Sprintf("  %s %s ", question, prompt))
	if ans == "" {
		return def
	}
	return strings.HasPrefix(strings.ToLower(ans), "y")
}

// Ask asks for a value, returning def for an empty answer or --yes.
func (u *UI) Ask(question, def string) string {
	if u.Yes {
		return def
	}
	if ans := u.read(fmt.Sprintf("  %s [%s] ", question, def)); ans != "" {
		return ans
	}
	return def
}

func (u *UI) read(prompt string) string {
	u.Printf("%s", prompt)
	line, err := u.in.ReadString('\n')
	if err != nil && line == "" {
		u.Println()
	}
	return strings.TrimSpace(line)
}
