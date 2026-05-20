package manifest

import (
	"testing"
	"time"

	"claude-code-launcher/internal/resolve"
)

func TestBuildManifestRedactsEnv(t *testing.T) {
	m := Build(resolve.LaunchSpec{
		Profile:      resolve.ResolvedProfile{Name: "min"},
		CWD:          "/work",
		ClaudeBinary: "/bin/claude",
		SettingsPath: "/tmp/settings.json",
		MCPPath:      "/tmp/mcp.json",
		Argv:         []string{"/bin/claude"},
		EnvOverrides: map[string]string{"API_TOKEN": "secret", "HOME": "/work"},
		DryRun:       true,
		Timestamp:    time.Unix(100, 0).UTC(),
		PID:          99,
		CCXVersion:   "dev",
	}, []string{"token"})
	if m.EnvOverrides["API_TOKEN"] != "[REDACTED]" {
		t.Fatalf("expected env redaction")
	}
	if m.Profile != "min" || m.PID != 99 {
		t.Fatalf("unexpected manifest data: %#v", m)
	}
}
