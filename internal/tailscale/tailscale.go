// Package tailscale reads the local Tailscale node's state through the
// tailscale CLI's JSON output.
package tailscale

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cemolmez/roost/internal/host"
)

// Node is a device on the tailnet.
type Node struct {
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
	OS       string `json:"OS"`
	Online   bool   `json:"Online"`
}

// IsPhone reports whether the node is a phone or tablet.
func (n Node) IsPhone() bool { return n.OS == "android" || n.OS == "iOS" }

// Status is the subset of `tailscale status --json` roost uses.
type Status struct {
	BackendState string          `json:"BackendState"`
	Self         Node            `json:"Self"`
	Peer         map[string]Node `json:"Peer"`
}

// Running reports whether the node is logged in and connected.
func (s *Status) Running() bool { return s.BackendState == "Running" }

// Peers returns the other devices sorted by name.
func (s *Status) Peers() []Node {
	out := make([]Node, 0, len(s.Peer))
	for _, p := range s.Peer {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].HostName < out[j].HostName })
	return out
}

// ParseStatus decodes `tailscale status --json` output.
func ParseStatus(b []byte) (*Status, error) {
	var s Status
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parse tailscale status: %w", err)
	}
	return &s, nil
}

// GetStatus runs `tailscale status --json`.
func GetStatus(ctx context.Context, h host.Host) (*Status, error) {
	// tailscale exits non-zero when logged out but still prints JSON.
	out, err := h.Run(ctx, "tailscale", "status", "--json")
	if out == "" && err != nil {
		return nil, err
	}
	return ParseStatus([]byte(out))
}

// Prefs is the subset of `tailscale debug prefs` roost uses.
type Prefs struct {
	RunSSH       bool   `json:"RunSSH"`
	Hostname     string `json:"Hostname"`
	OperatorUser string `json:"OperatorUser"`
}

// GetPrefs runs `tailscale debug prefs`.
func GetPrefs(ctx context.Context, h host.Host) (*Prefs, error) {
	out, err := h.Run(ctx, "tailscale", "debug", "prefs")
	if err != nil {
		return nil, err
	}
	var p Prefs
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		return nil, fmt.Errorf("parse tailscale prefs: %w", err)
	}
	return &p, nil
}
