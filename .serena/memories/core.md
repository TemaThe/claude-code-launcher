# Core

- Purpose: `ccx` is a Unix-first Go CLI/TUI that launches isolated Claude Code sessions from generated per-session config artifacts.
- Entry point: `cmd/ccx/main.go` calls `internal/app.Run`.
- Main flow: config load/create in `internal/config`, profile resolution in `internal/resolve`, JSON rendering in `internal/render`, manifest in `internal/manifest`, safety checks in `internal/preflight`, artifact writing and execution in `internal/runner`.
- Interactive UI code lives under `internal/tui`; current implementation is a minimal Bubble Tea model scaffold.
- Generated session artifacts are written under `/tmp/ccx/<timestamp>-<pid>/`.
- Root docs are in `README.md`.
- Read `mem:tech_stack` for dependencies and `mem:task_completion` for required verification.