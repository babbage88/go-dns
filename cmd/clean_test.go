package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunCleanZones_RequiresFileFlag(t *testing.T) {
	err := runCleanZones("", time.Second, 1, true)
	if err == nil {
		t.Fatalf("runCleanZones returned nil, want error")
	}
	if err.Error() != "--file is required" {
		t.Fatalf("runCleanZones error = %q, want %q", err.Error(), "--file is required")
	}
}

func TestRunCleanZones_ReadFailure(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	err := runCleanZones(missing, time.Second, 1, true)
	if err == nil {
		t.Fatalf("runCleanZones returned nil, want read error")
	}
}

func TestRunCleanZones_ParseFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")

	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("failed to write invalid YAML fixture: %v", err)
	}

	err := runCleanZones(path, time.Second, 1, true)
	if err == nil {
		t.Fatalf("runCleanZones returned nil, want parse error")
	}
}

func TestRunCleanZones_DryRunNoJobs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "zones.yaml")

	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("failed to write YAML fixture: %v", err)
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	runErr := runCleanZones(path, time.Second, 1, true)

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	if runErr != nil {
		t.Fatalf("runCleanZones returned error: %v", runErr)
	}

	if string(out) != "# dry-run enabled, no output written\n" {
		t.Fatalf("runCleanZones stdout = %q, want dry-run message", string(out))
	}
}
