package host

import (
	"context"
	"fmt"
	"strings"
)

// IsWSL reports whether roost is running inside Windows Subsystem for Linux.
func IsWSL(h Host) bool {
	b, err := h.ReadFile("/proc/sys/kernel/osrelease")
	return err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft")
}

// HasSystemd reports whether PID 1 is systemd.
func HasSystemd(h Host) bool {
	b, err := h.ReadFile("/proc/1/comm")
	return err == nil && strings.TrimSpace(string(b)) == "systemd"
}

// WSLDistro returns the name of the current WSL distribution.
func WSLDistro(h Host) string {
	if d := h.Getenv("WSL_DISTRO_NAME"); d != "" {
		return d
	}
	return "Ubuntu"
}

// PowerShell runs a script with Windows PowerShell on the Windows host (WSL only).
func PowerShell(ctx context.Context, h Host, script string) (string, error) {
	return h.Run(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
}

// WindowsPath converts a Windows path (C:\Users\me) to its WSL path (/mnt/c/Users/me).
func WindowsPath(ctx context.Context, h Host, winPath string) (string, error) {
	return h.Run(ctx, "wslpath", "-u", winPath)
}

// WSLToWindows converts a WSL path to a Windows path.
func WSLToWindows(ctx context.Context, h Host, path string) (string, error) {
	return h.Run(ctx, "wslpath", "-w", path)
}

// WindowsEnvDir resolves a Windows environment variable holding a directory
// (USERPROFILE, TEMP) to its WSL path.
func WindowsEnvDir(ctx context.Context, h Host, name string) (string, error) {
	win, err := PowerShell(ctx, h, "$env:"+name)
	if err != nil {
		return "", err
	}
	if win == "" {
		return "", fmt.Errorf("windows environment variable %s is empty", name)
	}
	return WindowsPath(ctx, h, win)
}

// IsLaptop reports whether the Windows host is a mobile PC (WSL only).
func IsLaptop(ctx context.Context, h Host) bool {
	out, err := PowerShell(ctx, h, "(Get-CimInstance Win32_ComputerSystem).PCSystemType")
	return err == nil && strings.TrimSpace(out) == "2"
}
