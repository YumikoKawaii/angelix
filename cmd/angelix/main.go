package main

import (
	"fmt"

	"github.com/alecthomas/kong"

	"github.com/YumikoKawaii/angelix/internal/auth"
	"github.com/YumikoKawaii/angelix/internal/cli"
	claudeexec "github.com/YumikoKawaii/angelix/internal/exec"
	"github.com/YumikoKawaii/angelix/internal/otel"
)

type CLI struct {
	Init   InitCmd   `cmd:"" help:"Set up ~/.angelix/config.json interactively"`
	Status StatusCmd `cmd:"" help:"Check server connectivity and show usage metrics"`
	Run    RunCmd    `cmd:"" default:"withargs" help:"Pass through to claude (default)"`
}

type InitCmd struct{}

func (c *InitCmd) Run() error { return cli.Init() }

type StatusCmd struct{}

func (c *StatusCmd) Run() error { return cli.Status() }

type RunCmd struct {
	Args []string `arg:"" optional:"" passthrough:""`
}

func (c *RunCmd) Run() error {
	cfg, err := auth.LoadConfig()
	if err != nil {
		return err
	}
	creds, err := auth.FetchCredentials(cfg.ServerURL, cfg.MemberToken)
	if err != nil {
		return err
	}
	if err := auth.WriteClaudeCredentials(creds.AccessToken, creds.RefreshToken); err != nil {
		return fmt.Errorf("write claude credentials: %w", err)
	}
	return claudeexec.RunClaude(c.Args, otel.BuildEnv(otel.Config{
		ServerURL:   cfg.ServerURL,
		MemberToken: cfg.MemberToken,
	}))
}

func main() {
	var k CLI
	ctx := kong.Parse(&k,
		kong.Name("angelix"),
		kong.Description("Claude Code wrapper — credential injection and usage telemetry"),
		kong.UsageOnError(),
	)
	ctx.FatalIfErrorf(ctx.Run())
}
