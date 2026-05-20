package tui

import (
	"errors"

	tea "github.com/charmbracelet/bubbletea"

	"claude-code-launcher/internal/config"
)

var ErrCancelled = errors.New("tui cancelled")

func Run(cfg *config.Config) (string, error) {
	model := NewModel(cfg)
	program := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		return "", err
	}
	result := finalModel.(Model)
	if result.cancelled {
		return "", ErrCancelled
	}
	return result.resultProfile, nil
}
