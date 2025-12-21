package main

import (
	"fmt"
	"os"

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

	data, err := os.ReadFile(cli.File)
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

	results := runPingWorkers(allJobs, cli.Timeout, cli.Workers)

	// Apply results SINGLE-THREADED
	for _, r := range results {
		if !r.ok {
			commentOut(r.job.Node, r.job.Why)
		}
	}

	createMissingPTRs(root)

	if cli.DryRun {
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
