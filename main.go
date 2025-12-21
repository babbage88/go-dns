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

	if cli.MaybePrintShellCompletion() {
		return
	}

	// read input file
	data, err := os.ReadFile(cli.File)
	if err != nil {
		panic(err)
	}

	// parse YAML with comments preserved
	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	root := file.Docs[0].Body.(*ast.MappingNode)

	// collect ping jobs
	nsJobs := collectNameserverJobs(root)
	dnsJobs := collectDNSRecordJobs(root)

	allJobs := append(nsJobs, dnsJobs...)

	// setup SIGINT / SIGTERM cancellation
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// run parallel pings
	results := runPingWorkers(
		ctx,
		allJobs,
		cli.Timeout,
		cli.Workers,
	)

	// apply results SINGLE-THREADED
	for _, r := range results {
		if r.ok {
			continue
		}

		commentOut(r.job.Node, "unreachable")
	}

	// generate missing PTRs (unchanged)
	createMissingPTRs(root)

	if cli.DryRun {
		fmt.Println("# dry-run enabled, no output written")
		return
	}

	// write output
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
