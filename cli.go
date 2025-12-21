package main

import (
	_ "embed"
	"fmt"
	"os"
	"time"
)

//go:embed completions/bash.sh
var bashCompletion string

//go:embed completions/zsh.sh
var zshCompletion string

// ───────── kong CLI struct for flag parsing ─────────

type CLI struct {
	CleanZones CleanZonesCmd `cmd:"" help:"Clean and validate DNS zones"`
	Completion CompletionCmd `cmd:"" help:"Generate shell completion script"`
}

type CleanZonesCmd struct {
	File    string        `help:"YAML file to process" required:""`
	Timeout time.Duration `help:"Ping timeout" default:"2s"`
	Workers int           `help:"Number of parallel ping workers" default:"8"`
	DryRun  bool          `help:"Do not modify output"`
}

type CompletionCmd struct {
	Shell string `arg:"" enum:"bash,zsh" help:"Shell type"`
}

func (c *CompletionCmd) Print() {
	fmt.Fprintln(os.Stderr, `
# To enable:
#   bash: source <(dnsctl completion bash)
#   zsh : source <(dnsctl completion zsh)
`)

	switch c.Shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	}
}
