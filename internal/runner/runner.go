package runner

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"claude-code-launcher/internal/resolve"
)

type Result struct {
	SessionDir       string
	SettingsPath     string
	MCPPath          string
	ManifestPath     string
	LaunchScriptPath string
}

func Prepare(root string, spec resolve.LaunchSpec) (Result, error) {
	sessionDir := filepath.Join(root, fmt.Sprintf("%d-%d", spec.Timestamp.Unix(), spec.PID))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return Result{}, err
	}
	result := Result{
		SessionDir:       sessionDir,
		SettingsPath:     filepath.Join(sessionDir, "settings.json"),
		MCPPath:          filepath.Join(sessionDir, "mcp.json"),
		ManifestPath:     filepath.Join(sessionDir, "manifest.json"),
		LaunchScriptPath: filepath.Join(sessionDir, "launch.sh"),
	}
	for path, body := range map[string][]byte{
		result.SettingsPath: spec.SettingsJSON,
		result.MCPPath:      spec.MCPJSON,
		result.ManifestPath: spec.ManifestJSON,
	} {
		if err := os.WriteFile(path, body, 0o644); err != nil {
			return Result{}, err
		}
	}
	argv := rewriteManagedPaths(spec.Argv, result.SettingsPath, result.MCPPath)
	script := buildLaunchScript(argv, spec.EnvOverrides)
	if err := os.WriteFile(result.LaunchScriptPath, []byte(script), 0o755); err != nil {
		return Result{}, err
	}
	return result, nil
}

func Run(root string, spec resolve.LaunchSpec, stdout, stderr io.Writer, stdin any) (Result, error) {
	result, err := Prepare(root, spec)
	if err != nil {
		return Result{}, err
	}
	argv := rewriteManagedPaths(spec.Argv, result.SettingsPath, result.MCPPath)
	commandLine := shellJoin(argv)
	switch {
	case spec.DryRun:
		_, _ = fmt.Fprintf(stdout, "COMMAND:\n%s\n\nSETTINGS.JSON:\n%s\n\nMCP.JSON:\n%s\n\nMANIFEST.JSON:\n%s\n", commandLine, spec.SettingsJSON, spec.MCPJSON, spec.ManifestJSON)
		return result, nil
	case spec.PrintOnly || spec.NoExec:
		_, _ = fmt.Fprintln(stdout, commandLine)
		return result, nil
	case spec.JSONOnly:
		_, _ = stdout.Write(spec.ManifestJSON)
		if len(spec.ManifestJSON) == 0 || spec.ManifestJSON[len(spec.ManifestJSON)-1] != '\n' {
			_, _ = fmt.Fprintln(stdout)
		}
		return result, nil
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = spec.CWD
	cmd.Env = append(os.Environ(), envList(spec.EnvOverrides)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if in, ok := stdin.(*os.File); ok {
		cmd.Stdin = in
	}
	return result, cmd.Run()
}

func ExitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if err == nil {
		return 0
	}
	return 1
}

func rewriteManagedPaths(argv []string, settingsPath, mcpPath string) []string {
	out := append([]string(nil), argv...)
	for i := 0; i < len(out)-1; i++ {
		if out[i] == "--settings" {
			out[i+1] = settingsPath
		}
		if out[i] == "--mcp-config" {
			out[i+1] = mcpPath
		}
	}
	return out
}

func buildLaunchScript(argv []string, env map[string]string) string {
	var b strings.Builder
	b.WriteString("#!/bin/sh\nset -eu\n")
	for key, value := range env {
		b.WriteString("export ")
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(strconv.Quote(value))
		b.WriteString("\n")
	}
	b.WriteString("exec ")
	b.WriteString(shellJoin(argv))
	b.WriteString("\n")
	return b.String()
}

func shellJoin(argv []string) string {
	quoted := make([]string, len(argv))
	for i, arg := range argv {
		quoted[i] = strconv.Quote(arg)
	}
	return strings.Join(quoted, " ")
}

func envList(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for key, value := range env {
		out = append(out, key+"="+value)
	}
	return out
}
