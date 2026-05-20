package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"claude-code-launcher/internal/config"
	"claude-code-launcher/internal/render"
	"claude-code-launcher/internal/resolve"
)

type triState int

const (
	triStateTrue triState = iota
	triStateFalse
	triStateInherit
)

type focusArea int

const (
	focusProfiles focusArea = iota
	focusPlugins
	focusPreview
)

type Model struct {
	config          *config.Config
	profiles        []string
	selectedIndex   int
	selectedProfile string
	initialProfile  string
	pluginOrder     []string
	pluginState     map[string]triState
	focus           focusArea
	width           int
	height          int
	preview         string
	commandPreview  string
	resultProfile   string
	cancelled       bool
}

func NewModel(cfg *config.Config) Model {
	if cfg == nil {
		parsed, _ := config.Parse([]byte(config.DefaultConfigYAML()))
		cfg = parsed
	}
	profiles := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	selected := "min"
	index := 0
	if len(profiles) > 0 {
		for i, name := range profiles {
			if name == selected {
				index = i
				break
			}
		}
		selected = profiles[index]
	}
	pluginOrder := make([]string, 0, len(cfg.Plugins))
	pluginState := make(map[string]triState, len(cfg.Plugins))
	for name := range cfg.Plugins {
		pluginOrder = append(pluginOrder, name)
		pluginState[name] = triStateInherit
	}
	sort.Strings(pluginOrder)
	m := Model{
		config:          cfg,
		profiles:        profiles,
		selectedIndex:   index,
		selectedProfile: selected,
		initialProfile:  selected,
		pluginOrder:     pluginOrder,
		pluginState:     pluginState,
		focus:           focusProfiles,
	}
	m.refreshPreview()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
	case tea.KeyMsg:
		switch typed.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "tab":
			m.focus = (m.focus + 1) % 3
		case "j", "down":
			if m.focus == focusProfiles {
				m.moveProfile(1)
			}
		case "k", "up":
			if m.focus == focusProfiles {
				m.moveProfile(-1)
			}
		case " ":
			if m.focus == focusPlugins {
				m.cyclePlugin()
			}
		case "r":
			m.reset()
		case "enter":
			m.resultProfile = m.selectedProfile
			return m, tea.Quit
		}
		m.refreshPreview()
	}
	return m, nil
}

func (m Model) View() string {
	sections := []string{
		m.renderProfiles(),
		m.renderPlugins(),
		m.renderStub("MCP Profile", m.config.Profiles[m.selectedProfile].MCPProfile),
		m.renderStub("Hook Profile", m.config.Profiles[m.selectedProfile].HookProfile),
		m.renderStub("Prompt/Context Flags", m.renderFlags()),
		m.renderStub("Claude Flags", m.commandPreview),
		m.renderStub("Env Vars", m.renderEnv()),
		m.renderStub("Command Preview", m.commandPreview),
		m.renderStub("JSON Preview", m.preview),
	}
	footer := "\nKeys: j/k move, tab focus, space toggle plugin, enter launch, r reset, q quit"
	return strings.Join(sections, "\n\n") + footer
}

func (m *Model) cyclePlugin() {
	if len(m.pluginOrder) == 0 {
		return
	}
	name := m.pluginOrder[0]
	switch m.pluginState[name] {
	case triStateTrue:
		m.pluginState[name] = triStateFalse
	case triStateFalse:
		m.pluginState[name] = triStateInherit
	default:
		m.pluginState[name] = triStateTrue
	}
}

func (m *Model) reset() {
	m.selectedIndex = 0
	for i, name := range m.profiles {
		if name == m.initialProfile {
			m.selectedIndex = i
			break
		}
	}
	m.selectedProfile = m.initialProfile
	for name := range m.pluginState {
		m.pluginState[name] = triStateInherit
	}
	m.refreshPreview()
}

func (m *Model) moveProfile(delta int) {
	if len(m.profiles) == 0 {
		return
	}
	m.selectedIndex = (m.selectedIndex + delta + len(m.profiles)) % len(m.profiles)
	m.selectedProfile = m.profiles[m.selectedIndex]
}

func (m *Model) refreshPreview() {
	spec, err := resolve.Build(m.config, resolve.Selection{ProfileName: m.selectedProfile})
	if err != nil {
		m.preview = err.Error()
		m.commandPreview = err.Error()
		return
	}
	settingsJSON, _ := render.Settings(spec.Settings)
	mcpJSON, _ := render.MCP(spec.MCP)
	m.preview = string(settingsJSON) + "\n\n" + string(mcpJSON)
	m.commandPreview = strings.Join(resolve.FinalArgv(spec, "/tmp/settings.json", "/tmp/mcp.json"), " ")
}

func (m Model) renderProfiles() string {
	var b strings.Builder
	b.WriteString("Profile Selection")
	for i, profile := range m.profiles {
		marker := " "
		if i == m.selectedIndex {
			marker = ">"
		}
		b.WriteString("\n" + marker + " " + profile)
	}
	return b.String()
}

func (m Model) renderPlugins() string {
	var b strings.Builder
	b.WriteString("Plugin Toggles")
	for _, name := range m.pluginOrder {
		b.WriteString("\n- " + name + ": " + m.pluginState[name].String())
	}
	return b.String()
}

func (m Model) renderFlags() string {
	profile := m.config.Profiles[m.selectedProfile]
	return fmt.Sprintf("strict_mcp_config=%t, exclude_dynamic_system_prompt_sections=%t", profile.Flags.StrictMCPConfig, profile.Flags.ExcludeDynamicSystemPromptSections)
}

func (m Model) renderEnv() string {
	profile := m.config.Profiles[m.selectedProfile]
	if len(profile.Env) == 0 {
		return "(none)"
	}
	parts := make([]string, 0, len(profile.Env))
	for key, value := range profile.Env {
		parts = append(parts, key+"="+value)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func (m Model) renderStub(title, body string) string {
	if body == "" {
		body = "(none)"
	}
	return title + "\n" + body
}

func (t triState) String() string {
	switch t {
	case triStateTrue:
		return "true"
	case triStateFalse:
		return "false"
	default:
		return "inherit"
	}
}
