package ini

import "testing"

func TestGet(t *testing.T) {
	t.Parallel()
	content := "# comment\n[boot]\nsystemd = true\n\n[wsl2]\nmemory=22GB\n"
	tests := []struct {
		section, key, want string
		found              bool
	}{
		{"boot", "systemd", "true", true},
		{"BOOT", "SystemD", "true", true},
		{"wsl2", "memory", "22GB", true},
		{"wsl2", "systemd", "", false},
		{"missing", "x", "", false},
	}
	for _, tt := range tests {
		got, found := Get(content, tt.section, tt.key)
		if got != tt.want || found != tt.found {
			t.Errorf("Get(%q,%q) = %q,%v; want %q,%v", tt.section, tt.key, got, found, tt.want, tt.found)
		}
	}
}

func TestSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, section, key, value, want string
	}{
		{
			name: "empty file creates section",
			in:   "", section: "boot", key: "systemd", value: "true",
			want: "[boot]\nsystemd=true\n",
		},
		{
			name: "replaces existing key in place",
			in:   "[boot]\nsystemd=false\ncommand=x\n", section: "boot", key: "systemd", value: "true",
			want: "[boot]\nsystemd=true\ncommand=x\n",
		},
		{
			name: "appends to existing section, keeps others",
			in:   "[wsl2]\nmemory=22GB\nswap=16GB\n\n[experimental]\na=b\n", section: "wsl2", key: "vmIdleTimeout", value: "-1",
			want: "[wsl2]\nmemory=22GB\nswap=16GB\nvmIdleTimeout=-1\n\n[experimental]\na=b\n",
		},
		{
			name: "adds new section after existing content",
			in:   "[user]\ndefault=dev\n\n", section: "boot", key: "systemd", value: "true",
			want: "[user]\ndefault=dev\n\n[boot]\nsystemd=true\n",
		},
		{
			name: "preserves comments and CRLF-free output",
			in:   "# keep me\n[wsl2]\n; also me\nmemory=8GB", section: "wsl2", key: "memory", value: "16GB",
			want: "# keep me\n[wsl2]\n; also me\nmemory=16GB\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Set(tt.in, tt.section, tt.key, tt.value); got != tt.want {
				t.Errorf("Set() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}
