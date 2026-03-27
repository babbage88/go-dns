package cmd

import (
	"context"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/goccy/go-yaml/ast"
)

func requireLocalPingSuccess(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("ping"); err != nil {
		t.Skip("ping binary not available in test environment")
	}

	if !ping("127.0.0.1", 2*time.Second) {
		t.Skip("ICMP ping to localhost is blocked in this test environment")
	}
}

// TestPing_Localhost tests ping to localhost when ICMP is available.
func TestPing_Localhost(t *testing.T) {
	requireLocalPingSuccess(t)

	result := ping("127.0.0.1", 2*time.Second)
	if !result {
		t.Fatalf("ping(127.0.0.1) = false, want true")
	}
}

// TestPing_InvalidIP tests ping to an unreachable IP address.
func TestPing_InvalidIP(t *testing.T) {
	result := ping("192.0.2.1", 1*time.Second)
	if result {
		t.Fatalf("ping(192.0.2.1) = true, want false")
	}
}

// TestPing_Timeout tests ping with a very short timeout.
func TestPing_Timeout(t *testing.T) {
	_ = ping("127.0.0.1", 1*time.Millisecond)
}

// TestRunPingWorkers_EmptyJobs tests ping workers with empty job list.
func TestRunPingWorkers_EmptyJobs(t *testing.T) {
	ctx := context.Background()
	jobs := []pingJob{}

	results := runPingWorkers(ctx, jobs, 1*time.Second, 4)

	if len(results) != 0 {
		t.Fatalf("runPingWorkers with empty jobs returned %d results, want 0", len(results))
	}
}

// TestRunPingWorkers_SingleHost tests ping workers with a single host.
func TestRunPingWorkers_SingleHost(t *testing.T) {
	requireLocalPingSuccess(t)

	ctx := context.Background()
	jobs := []pingJob{{
		IP:   "127.0.0.1",
		Node: &ast.StringNode{Value: "localhost"},
	}}

	results := runPingWorkers(ctx, jobs, 2*time.Second, 1)

	if len(results) != 1 {
		t.Fatalf("runPingWorkers returned %d results, want 1", len(results))
	}

	if !results[0].ok {
		t.Fatalf("runPingWorkers(127.0.0.1).ok = false, want true")
	}

	if results[0].job.IP != "127.0.0.1" {
		t.Fatalf("runPingWorkers result IP = %s, want 127.0.0.1", results[0].job.IP)
	}
}

// TestRunPingWorkers_MultipleHosts tests ping workers with multiple hosts.
func TestRunPingWorkers_MultipleHosts(t *testing.T) {
	requireLocalPingSuccess(t)

	ctx := context.Background()
	jobs := []pingJob{
		{IP: "127.0.0.1", Node: &ast.StringNode{Value: "localhost1"}},
		{IP: "192.0.2.1", Node: &ast.StringNode{Value: "unreachable"}},
		{IP: "127.0.0.1", Node: &ast.StringNode{Value: "localhost2"}},
	}

	results := runPingWorkers(ctx, jobs, 1*time.Second, 2)

	if len(results) != 3 {
		t.Fatalf("runPingWorkers returned %d results, want 3", len(results))
	}

	successCount := 0
	for _, r := range results {
		if r.ok {
			successCount++
		}
	}

	if successCount < 2 {
		t.Fatalf("runPingWorkers succeeded %d times, want at least 2", successCount)
	}
}

// TestRunPingWorkers_ContextCancellation tests ping workers with context cancellation.
func TestRunPingWorkers_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	jobs := []pingJob{
		{IP: "127.0.0.1", Node: &ast.StringNode{Value: "host1"}},
		{IP: "127.0.0.1", Node: &ast.StringNode{Value: "host2"}},
		{IP: "127.0.0.1", Node: &ast.StringNode{Value: "host3"}},
	}

	cancel()

	results := runPingWorkers(ctx, jobs, 2*time.Second, 2)
	if results == nil {
		t.Fatalf("runPingWorkers returned nil, want []pingResult")
	}
}

// TestRunPingWorkers_WithMultipleWorkers tests ping workers with different worker counts.
func TestRunPingWorkers_WithMultipleWorkers(t *testing.T) {
	requireLocalPingSuccess(t)

	workerCounts := []int{1, 2, 4, 8}

	for _, workers := range workerCounts {
		workers := workers
		t.Run("Workers="+strconv.Itoa(workers), func(t *testing.T) {
			ctx := context.Background()
			jobs := []pingJob{
				{IP: "127.0.0.1", Node: &ast.StringNode{Value: "host1"}},
				{IP: "127.0.0.1", Node: &ast.StringNode{Value: "host2"}},
				{IP: "127.0.0.1", Node: &ast.StringNode{Value: "host3"}},
			}

			results := runPingWorkers(ctx, jobs, 2*time.Second, workers)
			if len(results) != len(jobs) {
				t.Fatalf("runPingWorkers(%d workers) returned %d results, want %d", workers, len(results), len(jobs))
			}
		})
	}
}
