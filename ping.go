package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/goccy/go-yaml/ast"
)

//
// ─── PING ───────────────────────────────────────────────────────────────────────
//

func ping(ip string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", ip)
	return cmd.Run() == nil
}

type pingJob struct {
	IP   string
	Node ast.Node
	Why  string
}

type pingResult struct {
	job pingJob
	ok  bool
}

func runPingWorkers(
	ctx context.Context,
	jobs []pingJob,
	timeout time.Duration,
	workers int,
) []pingResult {

	jobCh := make(chan pingJob)
	resCh := make(chan pingResult)

	var completed int64
	total := int64(len(jobs))

	// workers
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobCh:
					if !ok {
						return
					}

					okPing := ping(job.IP, timeout)

					// send result
					resCh <- pingResult{job: job, ok: okPing}

					// progress update (stderr only)
					n := atomic.AddInt64(&completed, 1)
					fmt.Fprintf(os.Stderr, "\rPinging: %d/%d", n, total)
				}
			}
		}()
	}

	// feed jobs
	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			select {
			case <-ctx.Done():
				return
			case jobCh <- j:
			}
		}
	}()

	results := make([]pingResult, 0, len(jobs))

	for i := 0; i < len(jobs); i++ {
		select {
		case <-ctx.Done():
			fmt.Fprintln(os.Stderr) // newline before exit
			return results

		case res := <-resCh:
			results = append(results, res)
		}
	}

	// finish progress line cleanly
	fmt.Fprintln(os.Stderr)

	return results
}

// ───────── Collect Nameserver and DNSRecord pingJobs ─────────
func collectNameserverJobs(root *ast.MappingNode) []pingJob {
	var jobs []pingJob

	for _, mv := range root.Values {
		key := mv.Key.(*ast.StringNode).Value
		if !strings.HasPrefix(key, "nameservers") {
			continue
		}

		seq := mv.Value.(*ast.SequenceNode)
		for _, item := range seq.Values {
			m := item.(*ast.MappingNode)
			ip := stringValue(m, "ip_address")
			if ip != "" {
				jobs = append(jobs, pingJob{
					IP:   ip,
					Node: item,
					Why:  "nameserver unreachable",
				})
			}
		}
	}

	return jobs
}

func collectDNSRecordJobs(root *ast.MappingNode) []pingJob {
	var jobs []pingJob

	for _, section := range []string{"dns_records", "sub_zone_records"} {
		n := mappingValue(root, section)
		if n == nil {
			continue
		}

		seq := n.(*ast.SequenceNode)
		for _, item := range seq.Values {
			m := item.(*ast.MappingNode)
			typ := stringValue(m, "type")

			if typ != "A" && typ != "AAAA" {
				continue
			}

			ip := stringValue(m, "record_value")
			if ip == "" {
				continue
			}

			jobs = append(jobs, pingJob{
				IP:   ip,
				Node: item,
				Why:  "record unreachable",
			})
		}
	}

	return jobs
}
