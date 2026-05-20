package main

import (
	"os"

	"claude-code-launcher/internal/app"
	"claude-code-launcher/internal/runner"
)

func main() {
	err := app.Run(app.Options{
		Args:   os.Args[1:],
		Env:    os.Environ(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		os.Exit(runner.ExitCode(err))
	}
}
