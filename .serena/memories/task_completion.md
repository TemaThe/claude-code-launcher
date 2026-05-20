# Task Completion

- Run `gofmt -w` on touched Go files.
- Run `go test ./...` from the repo root.
- If CLI behavior changed, verify at least one dry-run path such as `go run ./cmd/ccx --dry-run min` or the built binary equivalent.
- Check `README.md` when user-facing flags or guarantees change.