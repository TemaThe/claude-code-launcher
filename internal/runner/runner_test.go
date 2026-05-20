package runner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"claude-code-launcher/internal/resolve"
)

func TestWriteArtifactsAndCommand(t *testing.T) {
	root := t.TempDir()
	spec := resolve.LaunchSpec{
		Profile:      resolve.ResolvedProfile{Name: "min"},
		CWD:          root,
		ClaudeBinary: "/usr/bin/claude",
		SettingsJSON: []byte("{\n  \"a\": 1\n}\n"),
		MCPJSON:      []byte("{\n  \"mcpServers\": {}\n}\n"),
		ManifestJSON: []byte("{\n  \"profile\": \"min\"\n}\n"),
		Argv:         []string{"/usr/bin/claude", "--settings", "SETTINGS", "--mcp-config", "MCP"},
		Timestamp:    time.Unix(100, 0).UTC(),
		PID:          42,
	}

	result, err := Prepare(root, spec)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	for _, path := range []string{result.SettingsPath, result.MCPPath, result.ManifestPath, result.LaunchScriptPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file %s: %v", path, err)
		}
	}
	content, err := os.ReadFile(result.LaunchScriptPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if !strings.Contains(string(content), "--settings") {
		t.Fatalf("launch script missing argv")
	}
}

func TestLaunchNoExecWritesArtifacts(t *testing.T) {
	root := t.TempDir()
	out := &strings.Builder{}
	spec := resolve.LaunchSpec{
		Profile:      resolve.ResolvedProfile{Name: "min"},
		CWD:          root,
		ClaudeBinary: "/usr/bin/claude",
		SettingsJSON: []byte("{}\n"),
		MCPJSON:      []byte("{\"mcpServers\":{}}\n"),
		ManifestJSON: []byte("{}\n"),
		Argv:         []string{"/usr/bin/claude"},
		Timestamp:    time.Unix(100, 0).UTC(),
		PID:          42,
		NoExec:       true,
	}
	if _, err := Run(root, spec, out, out, out); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(out.String(), "/usr/bin/claude") {
		t.Fatalf("expected command output")
	}
}

func TestRunRealLaunchPropagatesExitCode(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "claude")
	script := "#!/bin/sh\nexit 7\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	spec := resolve.LaunchSpec{
		Profile:      resolve.ResolvedProfile{Name: "min"},
		CWD:          root,
		ClaudeBinary: bin,
		SettingsJSON: []byte("{}\n"),
		MCPJSON:      []byte("{\"mcpServers\":{}}\n"),
		ManifestJSON: []byte("{}\n"),
		Argv:         []string{bin},
		Timestamp:    time.Unix(100, 0).UTC(),
		PID:          42,
	}
	_, err := Run(root, spec, os.Stdout, os.Stderr, os.Stdin)
	if err == nil {
		t.Fatalf("expected exit error")
	}
	if code := ExitCode(err); code != 7 {
		t.Fatalf("exit code = %d", code)
	}
}
