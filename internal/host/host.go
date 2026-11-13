// Package host is roost's only door to the machine it runs on: running
// commands, reading and writing files, and environment lookups. Everything
// else in roost talks to a Host, so steps can be tested against a fake.
package host

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Host is everything roost needs from the operating system.
type Host interface {
	// Run executes a command and returns its trimmed stdout.
	// A non-zero exit returns an error that includes stderr.
	Run(ctx context.Context, name string, args ...string) (string, error)
	// RunTTY executes a command attached to the user's terminal
	// (for sudo password prompts, installers and interactive tools).
	RunTTY(ctx context.Context, name string, args ...string) error
	// LookPath reports the full path of an executable, or "" if not found.
	LookPath(name string) string

	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	Exists(path string) bool
	MkdirAll(path string, perm fs.FileMode) error
	// CreateTemp writes data to a new, exclusively created 0600 file in dir
	// (name from pattern, as os.CreateTemp) and returns its path.
	CreateTemp(dir, pattern string, data []byte) (string, error)
	Remove(path string) error

	Getenv(key string) string
	HomeDir() string
	Username() string
}

// Local is the real Host.
type Local struct{}

var _ Host = Local{}

// Run implements Host.
func (Local) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(strings.ReplaceAll(stdout.String(), "\r", ""))
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = out
		}
		if msg != "" {
			return out, fmt.Errorf("%s: %w: %s", name, err, msg)
		}
		return out, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

// RunTTY implements Host.
func (Local) RunTTY(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// LookPath implements Host.
func (Local) LookPath(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

// ReadFile implements Host.
func (Local) ReadFile(path string) ([]byte, error) { return os.ReadFile(path) } //nolint:gosec // paths are roost-chosen config locations

// WriteFile implements Host. It creates parent directories as needed.
func (Local) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

// Exists implements Host.
func (Local) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MkdirAll implements Host.
func (Local) MkdirAll(path string, perm fs.FileMode) error { return os.MkdirAll(path, perm) }

// CreateTemp implements Host.
func (Local) CreateTemp(dir, pattern string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// Remove implements Host.
func (Local) Remove(path string) error { return os.Remove(path) }

// Getenv implements Host.
func (Local) Getenv(key string) string { return os.Getenv(key) }

// HomeDir implements Host.
func (Local) HomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return os.Getenv("HOME")
	}
	return h
}

// Username implements Host.
func (Local) Username() string {
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return os.Getenv("LOGNAME")
}

// IsNotExist reports whether err means a file does not exist.
func IsNotExist(err error) bool { return errors.Is(err, fs.ErrNotExist) }
