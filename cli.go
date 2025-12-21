package main

import (
	_ "embed"
	"fmt"
	"time"
)

//go:embed completions/bash.sh
var bashCompletion string

//go:embed completions/zsh.sh
var zshCompletion string

// ───────── kong CLI struct for flag parsing ─────────

type ShellCompletionCmd struct {
	Shell string `arg:"" enum:"bash,zsh" help:"Shell type"`
}

type CLI struct {
	File    string        `arg:"" optional:"" help:"YAML file to process"`
	Timeout time.Duration `help:"Ping timeout" default:"2s"`
	Workers int           `help:"Number of parallel ping workers" default:"8"`
	DryRun  bool          `help:"Do not modify output"`

	ShellCompletion ShellCompletionCmd `cmd:"" help:"Generate shell completion script"`
}

func (c *CLI) MaybePrintShellCompletion() bool {
	if c.ShellCompletion.Shell == "" {
		return false
	}

	switch c.ShellCompletion.Shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	}

	return true
}
