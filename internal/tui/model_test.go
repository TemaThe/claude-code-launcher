package tui

import "testing"

func TestNewModelInitialState(t *testing.T) {
	m := NewModel(nil)
	if m.selectedProfile == "" {
		t.Fatalf("expected default selected profile")
	}
}

func TestCyclePluginToggle(t *testing.T) {
	m := NewModel(nil)
	m.pluginOrder = []string{"serena@claude-plugins-official"}
	m.pluginState["serena@claude-plugins-official"] = triStateFalse
	m.focus = focusPlugins
	m.cyclePlugin()
	if got := m.pluginState["serena@claude-plugins-official"]; got != triStateInherit {
		t.Fatalf("plugin state = %v", got)
	}
}

func TestResetRestoresDefaults(t *testing.T) {
	m := NewModel(nil)
	original := m.selectedProfile
	m.selectedProfile = "serena"
	m.reset()
	if m.selectedProfile != original {
		t.Fatalf("expected reset to restore initial profile")
	}
}
