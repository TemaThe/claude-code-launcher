package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-code-launcher/internal/config"
	"claude-code-launcher/internal/manifest"
	"claude-code-launcher/internal/preflight"
	"claude-code-launcher/internal/render"
	"claude-code-launcher/internal/resolve"
	"claude-code-launcher/internal/runner"
	"claude-code-launcher/internal/tui"
)

type multiFlag []string

func (m *multiFlag) String() string { return fmt.Sprint([]string(*m)) }
func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type Options struct {
	Args   []string
	Env    []string
	Stdout io.Writer
	Stderr io.Writer
}

func Run(opts Options) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	fs := flag.NewFlagSet("ccx", flag.ContinueOnError)
	fs.SetOutput(opts.Stderr)
	var (
		profile    = fs.String("profile", "", "")
		configPath = fs.String("config", "", "")
		dryRun     = fs.Bool("dry-run", false, "")
		printOnly  = fs.Bool("print", false, "")
		jsonOnly   = fs.Bool("json", false, "")
		noExec     = fs.Bool("no-exec", false, "")
		extra      multiFlag
	)
	fs.Var(&extra, "extra-arg", "")
	if err := fs.Parse(opts.Args); err != nil {
		return err
	}
	rest := fs.Args()
	cfg, _, err := loadConfigWithEnv(*configPath, opts.Env)
	if err != nil {
		return err
	}
	interactive := !*dryRun && !*printOnly && !*jsonOnly && !*noExec && *profile == "" && len(rest) == 0 && len(extra) == 0
	if interactive {
		selected, err := tui.Run(cfg)
		if err != nil {
			if err == tui.ErrCancelled {
				return nil
			}
			return err
		}
		*profile = selected
	}
	if *profile == "" && len(rest) > 0 {
		*profile = rest[0]
		rest = rest[1:]
	}
	spec, err := resolve.Build(cfg, resolve.Selection{
		ProfileName: *profile,
		ExtraArgs:   append([]string(nil), extra...),
		PassThrough: rest,
	})
	if err != nil {
		return err
	}
	home := homeFromEnv(opts.Env)
	cwd, _ := os.Getwd()
	spec.CWD = cwd
	spec.DryRun = *dryRun
	spec.NoExec = *noExec
	spec.PrintOnly = *printOnly
	spec.JSONOnly = *jsonOnly
	spec.Timestamp = time.Now().UTC()
	spec.PID = os.Getpid()
	settingsJSON, err := render.Settings(spec.Settings)
	if err != nil {
		return err
	}
	mcpJSON, err := render.MCP(spec.MCP)
	if err != nil {
		return err
	}
	spec.SettingsJSON = append(settingsJSON, '\n')
	spec.MCPJSON = append(mcpJSON, '\n')
	spec.SettingsPath = filepath.Join(cfg.Defaults.TempRoot, "settings.json")
	spec.MCPPath = filepath.Join(cfg.Defaults.TempRoot, "mcp.json")
	spec.Argv = resolve.FinalArgv(spec, spec.SettingsPath, spec.MCPPath)
	manifestJSON, err := manifest.JSON(manifest.Build(spec, cfg.Defaults.RedactPatterns))
	if err != nil {
		return err
	}
	spec.ManifestJSON = append(manifestJSON, '\n')
	warnings, err := preflight.Validate(cfg.Defaults.TempRoot, home, spec)
	if err != nil {
		return err
	}
	spec.Warnings = warnings
	_, err = runner.Run(cfg.Defaults.TempRoot, spec, opts.Stdout, opts.Stderr, os.Stdin)
	return err
}

func loadConfigWithEnv(path string, env []string) (*config.Config, bool, error) {
	if path == "" {
		path = config.DefaultPath(homeFromEnv(env))
	}
	return config.Load(path)
}

func homeFromEnv(env []string) string {
	for _, entry := range env {
		if strings.HasPrefix(entry, "HOME=") {
			return strings.TrimPrefix(entry, "HOME=")
		}
	}
	home, _ := os.UserHomeDir()
	return home
}
