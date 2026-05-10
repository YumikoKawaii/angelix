package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/YumikoKawaii/angelix/internal/auth"
	"github.com/YumikoKawaii/angelix/internal/cli"
	claudeexec "github.com/YumikoKawaii/angelix/internal/exec"
	"github.com/YumikoKawaii/angelix/internal/otel"
)

type CLI struct {
	Init   InitCmd   `cmd:"" help:"Set up ~/.angelix/config.json interactively"`
	Setup  SetupCmd  `cmd:"" help:"Download the claude binary into ~/.angelix/bin/claude"`
	Status StatusCmd `cmd:"" help:"Check server connectivity and show usage metrics"`
	Run    RunCmd    `cmd:"" default:"withargs" help:"Pass through to claude (default)"`
}

type InitCmd struct{}

func (c *InitCmd) Run() error { return cli.Init() }

type SetupCmd struct {
	Version string `help:"Claude version to download (default: latest)" short:"v"`
}

func (c *SetupCmd) Run() error { return cli.Setup(c.Version) }

type StatusCmd struct{}

func (c *StatusCmd) Run() error { return cli.Status() }

type RunCmd struct {
	Args []string `arg:"" optional:"" passthrough:""`
}

func (c *RunCmd) Run() error {
	cfg, err := auth.LoadConfig()
	if err != nil {
		if !auth.IsNotConfigured(err) {
			return err
		}
		fmt.Println("Welcome to angelix — let's get you set up first.")
		fmt.Println()
		if err := cli.Init(); err != nil {
			return err
		}
		fmt.Println()
		cfg, err = auth.LoadConfig()
		if err != nil {
			return err
		}
	}
	creds, err := auth.FetchCredentials(cfg.ServerURL, cfg.MemberToken)
	if err != nil {
		return err
	}
	if err := auth.WriteClaudeCredentials(creds.AccessToken); err != nil {
		return fmt.Errorf("write claude credentials: %w", err)
	}
	return claudeexec.RunClaude(c.Args, otel.BuildEnv(otel.Config{
		ServerURL:   cfg.ServerURL,
		MemberToken: cfg.MemberToken,
	}))
}

// angelixCommands is the set of first arguments that angelix handles itself.
// Anything else is forwarded to claude via the implicit "run" command.
var angelixCommands = map[string]bool{
	"init": true, "setup": true, "status": true, "run": true,
	"--help": true, "-h": true, "help": true,
}

func main() {
	// If the first argument isn't a known angelix subcommand, prepend "run"
	// so that flags like --resume, --continue, etc. pass through to claude
	// rather than being rejected by kong.
	if len(os.Args) > 1 && !angelixCommands[os.Args[1]] {
		os.Args = append([]string{os.Args[0], "run"}, os.Args[1:]...)
	}

	var k CLI
	ctx := kong.Parse(&k,
		kong.Name("angelix"),
		kong.Description("Claude Code wrapper — credential injection and usage telemetry"),
		kong.UsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
