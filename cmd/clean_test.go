package cmd

import (
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
