package cmd

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

// ping sends a single ICMP echo request to the given IP and returns true if successful.
func ping(ip string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", ip)
	return cmd.Run() == nil
}

type pingJob struct {
	IP   string
	Node ast.Node
}

type pingResult struct {
	job pingJob
	ok  bool
}

// runPingWorkers concurrently pings multiple hosts and returns the results.
// It spawns the specified number of worker goroutines to process jobs in parallel.
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

// collectNameserverJobs extracts all unique nameserver IP addresses from the YAML root node.
func collectNameserverJobs(root *ast.MappingNode) []pingJob {
	var jobs []pingJob

	seenIPs := make(map[string]struct{})

	for _, mv := range root.Values {
		key := mv.Key.(*ast.StringNode).Value
		if !strings.HasPrefix(key, "nameservers") {
			continue
		}

		seq := mv.Value.(*ast.SequenceNode)
		for _, item := range seq.Values {
			m := item.(*ast.MappingNode)

			ip := stringValue(m, "ip_address")
			if ip == "" {
				continue
			}

			if _, alreadySeen := seenIPs[ip]; alreadySeen {
				continue
			}

			seenIPs[ip] = struct{}{}

			jobs = append(jobs, pingJob{
				IP:   ip,
				Node: item,
			})
		}
	}

	return jobs
}

// collectDNSRecordJobs extracts all unique A and AAAA record IPs from DNS records and sub-zones.
func collectDNSRecordJobs(root *ast.MappingNode) []pingJob {
	var jobs []pingJob

	seenIPs := make(map[string]struct{})

	for _, section := range []string{"dns_records", "sub_zone_records"} {
		n := mappingValue(root, section)
		if n == nil {
			continue
		}

		seq := n.(*ast.SequenceNode)
		for _, item := range seq.Values {
			m := item.(*ast.MappingNode)

			recordType := stringValue(m, "type")
			if recordType != "A" && recordType != "AAAA" {
				continue
			}

			ip := stringValue(m, "record_value")
			if ip == "" {
				continue
			}

			if _, alreadySeen := seenIPs[ip]; alreadySeen {
				continue
			}

			seenIPs[ip] = struct{}{}

			jobs = append(jobs, pingJob{
				IP:   ip,
				Node: item,
			})
		}
	}

	return jobs
}
