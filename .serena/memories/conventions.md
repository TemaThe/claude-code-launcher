# Conventions

- Keep the architecture split into small internal packages with explicit responsibilities; avoid folding config/render/launch logic into the CLI entrypoint.
- Session launch behavior should be derivable from explicit generated files and argv, not implicit global Claude state.
- Tests are package-local and behavior-focused; fake binaries and temp directories are preferred over mocks.
- JSON artifacts are pretty-printed.
- Built-in profile behavior (`min`, `serena`) is part of the contract and should stay covered by tests.