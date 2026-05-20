# ccx

`ccx` launches isolated Claude Code sessions from explicit per-session `settings.json` and `mcp.json` files. It does not mutate global Claude config, project `.claude` state, or ad hoc `.mcp.json` files.

The v1 implementation is Unix-first and uses Bubble Tea for the interactive TUI layer over a testable Go core.

## Safety Guarantee

- Reads existing Claude config only for diagnostics.
- Writes generated session artifacts under `/tmp/ccx/<timestamp>-<pid>/`.
- Always launches Claude with explicit `--settings` and `--mcp-config` paths.
- Never merges implicit hooks, plugins, or MCP servers from global/project Claude config.

## Install

```bash
go build ./cmd/ccx
```

## Usage

```bash
ccx
ccx min
ccx serena
ccx --profile serena
ccx --dry-run serena
ccx --print min
ccx --json serena
ccx --no-exec min
ccx serena --extra-arg --model --extra-arg opus -- --resume
```

## Config

Default config path: `~/.config/ccx/config.yaml`

On first run, `ccx` writes a starter config with built-in `min` and `serena` profiles plus MCP stubs:

```yaml
defaults:
  claude_binary: claude
  temp_root: /tmp/ccx
plugins:
  serena@claude-plugins-official:
    default: inherit
  gopls-lsp@claude-plugins-official:
    default: inherit
profiles:
  min:
    hook_profile: none
    mcp_profile: none
  serena:
    hook_profile: none
    mcp_profile: serena
```

## Dry Run

`ccx --dry-run <profile>` prints:

- The exact Claude command.
- Pretty `settings.json`.
- Pretty `mcp.json`.
- Pretty `manifest.json`.

That output is sufficient to reconstruct the launch outside `ccx`.

## Bubble Tea

The interactive interface is built on [Bubble Tea](https://github.com/charmbracelet/bubbletea), with supporting Charm libraries for components and styling.

## Alias Migration

If you previously used shell aliases or wrapper scripts to toggle Claude settings, move those presets into named `profiles` in `config.yaml` and launch them with `ccx <profile>`.

## Troubleshooting

- `claude binary not found`: set `defaults.claude_binary` or fix `PATH`.
- `mcp command not found`: adjust the selected MCP profile so each configured server command exists locally.
- `disabled MCP server` warning: `ccx` detected a server disabled in `~/.claude.json`; the generated session still stays isolated, but the warning is useful when comparing behavior.
