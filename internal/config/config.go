// Package config loads and saves roost's per-machine settings
// (~/.config/roost/config.toml).
package config

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/cemolmez/roost/internal/host"
)

// Role is what a machine is for.
type Role string

const (
	// RoleBox runs your work 24/7 (home PC, VPS, Pi). Phones and laptops connect to it.
	RoleBox Role = "box"
	// RoleClient is a laptop you connect from.
	RoleClient Role = "client"
)

// Valid reports whether r is a known role.
func (r Role) Valid() bool { return r == RoleBox || r == RoleClient }

// Config is the per-machine configuration.
type Config struct {
	// Role of this machine.
	Role Role `toml:"role"`
	// Name of this machine on the tailnet (e.g. "devbox").
	Name string `toml:"name"`
	// Box is the tailnet name of the box a client connects to.
	Box string `toml:"box,omitempty"`
	// RCDir is where Claude Remote Control starts new sessions (box only).
	RCDir string `toml:"rc_dir,omitempty"`
}

// DefaultBoxName is the suggested tailnet name for a box.
const DefaultBoxName = "devbox"

var nameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Validate checks the config for a complete, usable setup.
func (c Config) Validate() error {
	if !c.Role.Valid() {
		return fmt.Errorf("role must be %q or %q, got %q", RoleBox, RoleClient, c.Role)
	}
	if !nameRe.MatchString(c.Name) {
		return fmt.Errorf("name %q must be lowercase letters, digits and dashes (a valid hostname)", c.Name)
	}
	if c.Role == RoleClient && c.Box != "" && !nameRe.MatchString(c.Box) {
		return fmt.Errorf("box %q must be a valid hostname", c.Box)
	}
	if c.RCDir != "" {
		if err := SafePath(c.RCDir); err != nil {
			return fmt.Errorf("rc_dir: %w", err)
		}
	}
	return nil
}

// SafePath checks that p is an absolute, clean path without characters that
// would need escaping in a systemd unit or a shell command line.
func SafePath(p string) error {
	if !filepath.IsAbs(p) || filepath.Clean(p) != p {
		return fmt.Errorf("%q must be an absolute path like /home/you/projects", p)
	}
	if strings.ContainsAny(p, " \t\n\r\"'\\%$`;&|<>*?") {
		return fmt.Errorf("%q contains spaces or special characters roost can't safely use", p)
	}
	return nil
}

// Dir returns roost's config directory.
func Dir(h host.Host) string {
	if x := h.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "roost")
	}
	return filepath.Join(h.HomeDir(), ".config", "roost")
}

// Path returns the config file path.
func Path(h host.Host) string { return filepath.Join(Dir(h), "config.toml") }

// ErrNotFound means roost has not been set up on this machine yet.
var ErrNotFound = errors.New("roost is not set up on this machine yet (run: roost init)")

// Load reads the config file. It returns ErrNotFound if there is none.
func Load(h host.Host) (Config, error) {
	var c Config
	b, err := h.ReadFile(Path(h))
	if host.IsNotExist(err) {
		return c, ErrNotFound
	}
	if err != nil {
		return c, fmt.Errorf("read config: %w", err)
	}
	if _, err := toml.Decode(string(b), &c); err != nil {
		return c, fmt.Errorf("parse %s: %w", Path(h), err)
	}
	if err := c.Validate(); err != nil {
		return c, fmt.Errorf("%s: %w (fix the file or run: roost init)", Path(h), err)
	}
	return c, nil
}

// Save writes the config file.
func Save(h host.Host, c Config) error {
	if err := c.Validate(); err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("# roost config — edit freely, then run: roost init\n")
	if err := toml.NewEncoder(&buf).Encode(c); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := h.WriteFile(Path(h), buf.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
