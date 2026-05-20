package resolve

import (
	"testing"

	"claude-code-launcher/internal/config"
)

func TestResolveBuiltInMinProfile(t *testing.T) {
	cfg, err := config.Parse([]byte(config.DefaultConfigYAML()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	spec, err := Build(cfg, Selection{ProfileName: "min"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if spec.Profile.Name != "min" {
		t.Fatalf("profile name = %q", spec.Profile.Name)
	}
	if spec.Profile.HookProfile != "none" {
		t.Fatalf("hook profile = %q", spec.Profile.HookProfile)
	}
	if spec.Profile.MCPProfile != "none" {
		t.Fatalf("mcp profile = %q", spec.Profile.MCPProfile)
	}
	if !spec.Profile.Flags.StrictMCPConfig || !spec.Profile.Flags.ExcludeDynamicSystemPromptSections {
		t.Fatalf("expected built-in min flags to be enabled")
	}
	if got := spec.Profile.Plugins["serena@claude-plugins-official"]; got != config.TriStateFalse {
		t.Fatalf("serena plugin = %v", got)
	}
}

func TestResolveBuiltInSerenaProfile(t *testing.T) {
	cfg, err := config.Parse([]byte(config.DefaultConfigYAML()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	spec, err := Build(cfg, Selection{ProfileName: "serena"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if spec.Profile.MCPProfile != "serena" {
		t.Fatalf("mcp profile = %q", spec.Profile.MCPProfile)
	}
	if got := spec.Profile.Plugins["gopls-lsp@claude-plugins-official"]; got != config.TriStateTrue {
		t.Fatalf("gopls plugin = %v", got)
	}
}

func TestResolveOmitsInheritedPluginsFromSettings(t *testing.T) {
	cfg, err := config.Parse([]byte(config.DefaultConfigYAML()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	cfg.Plugins["sample"] = config.PluginSpec{Default: config.TriStateInherit}
	cfg.Profiles["custom"] = config.Profile{
		Plugins: map[string]*config.TriState{
			"sample": config.TriStateInherit.Ptr(),
		},
	}

	spec, err := Build(cfg, Selection{ProfileName: "custom"})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if _, ok := spec.Settings.Plugins["sample"]; ok {
		t.Fatalf("expected inherited plugin to be omitted from settings")
	}
}

func TestResolveUnknownProfile(t *testing.T) {
	cfg, err := config.Parse([]byte(config.DefaultConfigYAML()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if _, err := Build(cfg, Selection{ProfileName: "missing"}); err == nil {
		t.Fatalf("expected unknown profile error")
	}
}

func TestBuildArgvIncludesManagedFlagsAndPassthrough(t *testing.T) {
	cfg, err := config.Parse([]byte(config.DefaultConfigYAML()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	spec, err := Build(cfg, Selection{
		ProfileName: "serena",
		ExtraArgs:   []string{"--model", "opus"},
		PassThrough: []string{"--", "hello"},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	argv := FinalArgv(spec, "/tmp/settings.json", "/tmp/mcp.json")
	want := []string{
		cfg.Defaults.ClaudeBinary,
		"--settings", "/tmp/settings.json",
		"--mcp-config", "/tmp/mcp.json",
		"--strict-mcp-config",
		"--exclude-dynamic-system-prompt-sections",
		"--model", "opus",
		"--", "hello",
	}
	if len(argv) != len(want) {
		t.Fatalf("argv len = %d want %d: %#v", len(argv), len(want), argv)
	}
	for i := range want {
		if argv[i] != want[i] {
			t.Fatalf("argv[%d] = %q want %q", i, argv[i], want[i])
		}
	}
}
