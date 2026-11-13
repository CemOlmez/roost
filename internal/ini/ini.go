// Package ini reads and edits the small INI files WSL uses (/etc/wsl.conf,
// %USERPROFILE%\.wslconfig) while preserving everything it doesn't touch:
// comments, ordering, blank lines and unknown keys.
package ini

import (
	"strings"
)

// Get returns the value of key in [section] and whether it was found.
// Section and key names are matched case-insensitively, as WSL does.
func Get(content, section, key string) (string, bool) {
	cur := ""
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if s, ok := sectionName(t); ok {
			cur = s
			continue
		}
		if !strings.EqualFold(cur, section) {
			continue
		}
		if k, v, ok := keyValue(t); ok && strings.EqualFold(k, key) {
			return v, true
		}
	}
	return "", false
}

// Set returns content with key=value in [section]. An existing key is
// replaced in place; otherwise the key is appended to the section, and the
// section is created at the end if missing.
func Set(content, section, key, value string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	entry := key + "=" + value

	cur, secEnd := "", -1
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if s, ok := sectionName(t); ok {
			cur = s
			if strings.EqualFold(s, section) {
				secEnd = i
			}
			continue
		}
		if !strings.EqualFold(cur, section) {
			continue
		}
		if k, _, ok := keyValue(t); ok && strings.EqualFold(k, key) {
			lines[i] = entry
			return join(lines)
		}
		if t != "" {
			secEnd = i
		}
	}

	if secEnd >= 0 {
		lines = append(lines[:secEnd+1], append([]string{entry}, lines[secEnd+1:]...)...)
		return join(lines)
	}
	// Drop trailing blank lines before appending a new section.
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > 0 {
		lines = append(lines, "")
	}
	lines = append(lines, "["+section+"]", entry)
	return join(lines)
}

func join(lines []string) string {
	s := strings.Join(lines, "\n")
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return s
}

func sectionName(t string) (string, bool) {
	if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
		return strings.TrimSpace(t[1 : len(t)-1]), true
	}
	return "", false
}

func keyValue(t string) (key, value string, ok bool) {
	if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, ";") {
		return "", "", false
	}
	k, v, found := strings.Cut(t, "=")
	if !found {
		return "", "", false
	}
	return strings.TrimSpace(k), strings.TrimSpace(v), true
}
