package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// runCleanZones reads a YAML file, validates DNS records and nameservers by pinging them,
// comments out unreachable entries, and generates PTR records for A records.
func runCleanZones(filePath string, pingTimeout time.Duration, numWorkers int, dryRun bool) error {
	if filePath == "" {
		return fmt.Errorf("--file is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return err
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
		pingTimeout,
		numWorkers,
	)

	// apply results single-threaded
	for _, r := range results {
		if !r.ok {
			commentOut(r.job.Node, "unreachable")
		}
	}

	createMissingPTRs(root)

	if dryRun {
		fmt.Println("# dry-run enabled, no output written")
		return nil
	}

	out, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

// kv creates a YAML mapping value node with the given key and string value.
func kv(k, v string) *ast.MappingValueNode {
	return &ast.MappingValueNode{
		Key:   &ast.StringNode{Value: k},
		Value: &ast.StringNode{Value: v},
	}
}
