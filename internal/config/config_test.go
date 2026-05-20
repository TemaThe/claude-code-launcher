package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDefaultConfigCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, created, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !created {
		t.Fatalf("expected config to be created")
	}
	if cfg == nil {
		t.Fatalf("expected config")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
	if _, ok := cfg.Profiles["serena"]; !ok {
		t.Fatalf("expected built-in serena profile")
	}
}

func TestLoadConfigFromExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
defaults:
  claude_binary: /usr/local/bin/claude
plugins:
  demo:
    default: inherit
profiles:
  test:
    plugins:
      demo: true
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, created, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if created {
		t.Fatalf("did not expect config creation")
	}
	if got := cfg.Defaults.ClaudeBinary; got != "/usr/local/bin/claude" {
		t.Fatalf("ClaudeBinary = %q", got)
	}
	if got := cfg.Profiles["test"].Plugins["demo"]; got == nil || *got != TriStateTrue {
		t.Fatalf("profile plugin override not loaded")
	}
}
