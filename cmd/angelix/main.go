package main

import (
	"fmt"
	"os"

	"github.com/YumikoKawaii/angelix/internal/auth"
	"github.com/YumikoKawaii/angelix/internal/cli"
	claudeexec "github.com/YumikoKawaii/angelix/internal/exec"
	"github.com/YumikoKawaii/angelix/internal/otel"
)

const usage = `Usage: angelix [command] [args...]

Commands:
  init     Set up ~/.angelix/config.json interactively
  status   Check server connectivity and show usage metrics
  help     Show this message

Anything else is passed directly to claude.
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := cli.Init(); err != nil {
				fmt.Fprintf(os.Stderr, "angelix init: %v\n", err)
				os.Exit(1)
			}
			return
		case "status":
			if err := cli.Status(); err != nil {
				fmt.Fprintf(os.Stderr, "angelix status: %v\n", err)
				os.Exit(1)
			}
			return
		case "help", "--help", "-h":
			fmt.Print(usage)
			return
		}
	}

	cfg, err := auth.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "angelix: %v\n", err)
		fmt.Fprintf(os.Stderr, "run 'angelix init' to set up your config\n")
		os.Exit(1)
	}

	creds, err := auth.FetchCredentials(cfg.ServerURL, cfg.MemberToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "angelix: credential fetch failed: %v\n", err)
		os.Exit(1)
	}

	extraEnv := append(
		otel.BuildEnv(otel.Config{
			ServerURL:   cfg.ServerURL,
			MemberToken: cfg.MemberToken,
		}),
		"ANTHROPIC_API_KEY="+creds.APIKey,
	)

	if err := claudeexec.RunClaude(os.Args[1:], extraEnv); err != nil {
		fmt.Fprintf(os.Stderr, "angelix: %v\n", err)
		os.Exit(1)
	}
}
