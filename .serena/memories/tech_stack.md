# Tech Stack

- Language: Go.
- Module: `claude-code-launcher`.
- Stdlib-heavy implementation with YAML parsing via `gopkg.in/yaml.v3`.
- TUI dependency surface includes Bubble Tea and related Charm packages (`bubbletea`, `bubbles`, `lipgloss`), though the current scaffold uses Bubble Tea directly.
- Testing uses the standard `go test` runner across `./...`.