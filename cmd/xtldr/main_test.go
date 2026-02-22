package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelpCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run([]string{"help"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help usage output")
	}
	if !strings.Contains(out.String(), "Hide Command Explanation panel") {
		t.Fatalf("expected -e flag help to hide explanation panel")
	}
}

func TestRunVersionCommand(t *testing.T) {
	origVersion, origCommit, origBuildDate := Version, Commit, BuildDate
	t.Cleanup(func() {
		Version, Commit, BuildDate = origVersion, origCommit, origBuildDate
	})
	Version, Commit, BuildDate = "v1.2.3", "abc123", "2026-02-22T00:00:00Z"

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	got := out.String()
	if !strings.Contains(got, "v1.2.3") || !strings.Contains(got, "abc123") {
		t.Fatalf("unexpected version output: %s", got)
	}
}

func TestRunVersionFlag(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"--version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out.String(), "xtldr version") {
		t.Fatalf("expected version output, got %q", out.String())
	}
}
