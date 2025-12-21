package main

import "time"

// ───────── kong CLI struct for flag parsing ─────────

type CLI struct {
	File    string        `arg:"" required:"" help:"YAML file to process"`
	Timeout time.Duration `help:"Ping timeout" default:"2s"`
	Workers int           `help:"Number of parallel ping workers" default:"8"`
	DryRun  bool          `help:"Do not modify output"`
}
