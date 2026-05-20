package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDryRunCreatesDefaultConfigAndPrettyOutput(t *testing.T) {
	home := t.TempDir()
	binDir := t.TempDir()
	claude := filepath.Join(binDir, "claude")
	if err := os.WriteFile(claude, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := Run(Options{
		Args:   []string{"--dry-run", "min"},
		Env:    []string{"HOME=" + home, "PATH=" + binDir},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	configPath := filepath.Join(home, ".config", "ccx", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected default config creation: %v", err)
	}
	out := stdout.String()
	for _, want := range []string{"COMMAND:", "SETTINGS.JSON:", "MCP.JSON:", "MANIFEST.JSON:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("dry-run output missing %q\n%s", want, out)
		}
	}
}

func TestRunJSONModePrintsManifestOnly(t *testing.T) {
	home := t.TempDir()
	binDir := t.TempDir()
	claude := filepath.Join(binDir, "claude")
	if err := os.WriteFile(claude, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	var stdout, stderr bytes.Buffer
	err := Run(Options{
		Args:   []string{"--json", "serena"},
		Env:    []string{"HOME=" + home, "PATH=" + binDir},
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "\"profile\": \"serena\"") {
		t.Fatalf("manifest output missing profile: %s", out)
	}
	if strings.Contains(out, "SETTINGS.JSON:") {
		t.Fatalf("expected manifest-only output")
	}
}
