package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type TriState string

const (
	TriStateTrue    TriState = "true"
	TriStateFalse   TriState = "false"
	TriStateInherit TriState = "inherit"
)

func (t TriState) Ptr() *TriState {
	v := t
	return &v
}

func (t *TriState) UnmarshalYAML(node *yaml.Node) error {
	switch strings.ToLower(strings.TrimSpace(node.Value)) {
	case "true":
		*t = TriStateTrue
	case "false":
		*t = TriStateFalse
	case "", "inherit":
		*t = TriStateInherit
	default:
		return fmt.Errorf("invalid tri-state %q", node.Value)
	}
	return nil
}

type Defaults struct {
	ClaudeBinary   string   `yaml:"claude_binary"`
	TempRoot       string   `yaml:"temp_root"`
	RedactPatterns []string `yaml:"redact_patterns"`
}

type PluginSpec struct {
	Default TriState `yaml:"default"`
}

type HookProfile struct {
	Hooks map[string]any `yaml:"hooks"`
}

type MCPProfile struct {
	Servers map[string]map[string]any `yaml:"servers"`
}

type Flags struct {
	StrictMCPConfig                    bool `yaml:"strict_mcp_config" json:"strict_mcp_config"`
	ExcludeDynamicSystemPromptSections bool `yaml:"exclude_dynamic_system_prompt_sections" json:"exclude_dynamic_system_prompt_sections"`
}

type SettingsPassthrough map[string]any

type Profile struct {
	HookProfile string               `yaml:"hook_profile"`
	MCPProfile  string               `yaml:"mcp_profile"`
	Plugins     map[string]*TriState `yaml:"plugins"`
	Flags       Flags                `yaml:"flags"`
	Env         map[string]string    `yaml:"env"`
	Settings    SettingsPassthrough  `yaml:"settings"`
}

type Config struct {
	Defaults Defaults               `yaml:"defaults"`
	Plugins  map[string]PluginSpec  `yaml:"plugins"`
	Hooks    map[string]HookProfile `yaml:"hooks"`
	MCP      map[string]MCPProfile  `yaml:"mcp"`
	Profiles map[string]Profile     `yaml:"profiles"`
}

func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Defaults.ClaudeBinary == "" {
		cfg.Defaults.ClaudeBinary = "claude"
	}
	if cfg.Defaults.TempRoot == "" {
		cfg.Defaults.TempRoot = "/tmp/ccx"
	}
	if len(cfg.Defaults.RedactPatterns) == 0 {
		cfg.Defaults.RedactPatterns = []string{"key", "token", "secret", "password"}
	}
	if cfg.Plugins == nil {
		cfg.Plugins = map[string]PluginSpec{}
	}
	if cfg.Hooks == nil {
		cfg.Hooks = map[string]HookProfile{}
	}
	if cfg.MCP == nil {
		cfg.MCP = map[string]MCPProfile{}
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Load(path string) (*Config, bool, error) {
	if path == "" {
		path = DefaultPath(userHome(os.Environ()))
	}
	expanded, err := expandHome(path, userHome(os.Environ()))
	if err != nil {
		return nil, false, err
	}
	if _, err := os.Stat(expanded); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(expanded), 0o755); err != nil {
			return nil, false, err
		}
		if err := os.WriteFile(expanded, []byte(DefaultConfigYAML()), 0o644); err != nil {
			return nil, false, err
		}
		cfg, err := Parse([]byte(DefaultConfigYAML()))
		return cfg, true, err
	} else if err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, false, err
	}
	cfg, err := Parse(data)
	return cfg, false, err
}

func DefaultPath(home string) string {
	return filepath.Join(home, ".config", "ccx", "config.yaml")
}

func expandHome(path, home string) (string, error) {
	if path == "" {
		return "", nil
	}
	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func userHome(env []string) string {
	for _, entry := range env {
		if strings.HasPrefix(entry, "HOME=") {
			return strings.TrimPrefix(entry, "HOME=")
		}
	}
	home, _ := os.UserHomeDir()
	return home
}

func validate(cfg *Config) error {
	for name, profile := range cfg.Profiles {
		if profile.HookProfile != "" {
			if _, ok := cfg.Hooks[profile.HookProfile]; !ok {
				return fmt.Errorf("profile %q references unknown hook profile %q", name, profile.HookProfile)
			}
		}
		if profile.MCPProfile != "" {
			if _, ok := cfg.MCP[profile.MCPProfile]; !ok {
				return fmt.Errorf("profile %q references unknown mcp profile %q", name, profile.MCPProfile)
			}
		}
	}
	return nil
}
