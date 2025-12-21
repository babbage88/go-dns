package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

//
// ─── MAIN ──────────────────────────────────────────────────────────────────────
//

func main() {
	var cli CLI
	kong.Parse(&cli)

	// ─── shell completion subcommand ─────────────────────────────
	if cli.Completion.Shell != "" {
		cli.Completion.Print()
		return
	}

	// ─── clean-zones subcommand ──────────────────────────────────
	run := cli.CleanZones
	if run.File == "" {
		// kong already prints help, just exit cleanly
		return
	}

	data, err := os.ReadFile(run.File)
	if err != nil {
		panic(err)
	}

	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	root := file.Docs[0].Body.(*ast.MappingNode)

	nsJobs := collectNameserverJobs(root)
	dnsJobs := collectDNSRecordJobs(root)
	allJobs := append(nsJobs, dnsJobs...)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	results := runPingWorkers(
		ctx,
		allJobs,
		run.Timeout,
		run.Workers,
	)

	// apply results single-threaded
	for _, r := range results {
		if !r.ok {
			commentOut(r.job.Node, "unreachable")
		}
	}

	createMissingPTRs(root)

	if run.DryRun {
		fmt.Println("# dry-run enabled, no output written")
		return
	}

	out, err := yaml.Marshal(file)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

func kv(k, v string) *ast.MappingValueNode {
	return &ast.MappingValueNode{
		Key:   &ast.StringNode{Value: k},
		Value: &ast.StringNode{Value: v},
	}
}
