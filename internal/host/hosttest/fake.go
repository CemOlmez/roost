// Package hosttest provides an in-memory host.Host for tests.
package hosttest

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cemolmez/roost/internal/host"
)

// Fake is a scriptable host.Host. Commands are matched by their full command
// line ("name arg1 arg2"); unmatched commands fail.
type Fake struct {
	mu    sync.Mutex
	Files map[string][]byte
	// Dirs are directories that exist (the home directory always does).
	Dirs map[string]bool
	// Written keeps the last content written to every path, even after Remove.
	Written  map[string][]byte
	Bins     map[string]bool
	Env      map[string]string
	Home     string
	User     string
	Commands map[string]Result
	// Calls records every command line run, in order.
	Calls []string
}

// Result is a scripted command outcome.
type Result struct {
	Out string
	Err error
}

var _ host.Host = (*Fake)(nil)

// New returns an empty fake for user "dev" with home /home/dev.
func New() *Fake {
	return &Fake{
		Files:    map[string][]byte{},
		Dirs:     map[string]bool{"/home/dev": true},
		Written:  map[string][]byte{},
		Bins:     map[string]bool{},
		Env:      map[string]string{},
		Home:     "/home/dev",
		User:     "dev",
		Commands: map[string]Result{},
	}
}

// On scripts the result of a command line.
func (f *Fake) On(cmdline, out string, err error) *Fake {
	f.Commands[cmdline] = Result{Out: out, Err: err}
	return f
}

// File sets a file's content.
func (f *Fake) File(path, content string) *Fake {
	f.Files[path] = []byte(content)
	return f
}

// Dir marks a directory as existing.
func (f *Fake) Dir(path string) *Fake {
	f.Dirs[path] = true
	return f
}

// Bin marks an executable as installed.
func (f *Fake) Bin(names ...string) *Fake {
	for _, n := range names {
		f.Bins[n] = true
	}
	return f
}

// Ran reports whether a command line was run.
func (f *Fake) Ran(cmdline string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, c := range f.Calls {
		if c == cmdline {
			return true
		}
	}
	return false
}

func (f *Fake) run(name string, args ...string) (string, error) {
	line := strings.TrimSpace(name + " " + strings.Join(args, " "))
	f.mu.Lock()
	f.Calls = append(f.Calls, line)
	r, ok := f.Commands[line]
	f.mu.Unlock()
	if !ok {
		return "", fmt.Errorf("hosttest: unscripted command %q", line)
	}
	return r.Out, r.Err
}

// Run implements host.Host.
func (f *Fake) Run(_ context.Context, name string, args ...string) (string, error) {
	return f.run(name, args...)
}

// RunTTY implements host.Host.
func (f *Fake) RunTTY(_ context.Context, name string, args ...string) error {
	_, err := f.run(name, args...)
	return err
}

// LookPath implements host.Host.
func (f *Fake) LookPath(name string) string {
	if f.Bins[name] {
		return "/usr/bin/" + name
	}
	return ""
}

// ReadFile implements host.Host.
func (f *Fake) ReadFile(path string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.Files[path]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
	}
	return b, nil
}

// WriteFile implements host.Host.
func (f *Fake) WriteFile(path string, data []byte, _ fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Files[filepath.Clean(path)] = append([]byte(nil), data...)
	f.Written[filepath.Clean(path)] = append([]byte(nil), data...)
	return nil
}

// Exists implements host.Host.
func (f *Fake) Exists(path string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.Files[path]
	return ok || f.Dirs[path]
}

// CreateTemp implements host.Host. The name is dir/pattern with "*" → "tmp".
func (f *Fake) CreateTemp(dir, pattern string, data []byte) (string, error) {
	p := filepath.Join(dir, strings.Replace(pattern, "*", "tmp", 1))
	return p, f.WriteFile(p, data, 0o600)
}

// Remove implements host.Host.
func (f *Fake) Remove(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.Files, path)
	return nil
}

// MkdirAll implements host.Host.
func (f *Fake) MkdirAll(string, fs.FileMode) error { return nil }

// Getenv implements host.Host.
func (f *Fake) Getenv(key string) string { return f.Env[key] }

// HomeDir implements host.Host.
func (f *Fake) HomeDir() string { return f.Home }

// Username implements host.Host.
func (f *Fake) Username() string { return f.User }
