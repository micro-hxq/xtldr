package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"xtldr/internal/model"
)

type fakeGenerator struct {
	candidates []model.Candidate
	err        error
}

type fakeCopier struct {
	err    error
	copied string
}

func (f fakeGenerator) Generate(ctx context.Context, request, workingDir string) ([]model.Candidate, error) {
	_ = ctx
	_ = request
	_ = workingDir
	return f.candidates, f.err
}

func (f *fakeCopier) Copy(text string) error {
	f.copied = text
	return f.err
}

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
	if !strings.Contains(out.String(), "--non-interactive") {
		t.Fatalf("expected non-interactive flag in help output")
	}
	if !strings.Contains(out.String(), "Enable iterative multi-turn refinement") {
		t.Fatalf("expected -i flag help output")
	}
	if !strings.Contains(out.String(), "xtldr history [query]") {
		t.Fatalf("expected history command in help output")
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

func TestRunRoadmapCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := run([]string{"roadmap"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	got := out.String()
	if !strings.Contains(got, "Capability gaps compared with mature CLI tools") {
		t.Fatalf("expected roadmap header, got %q", got)
	}
	if !strings.Contains(got, "manually confirm priorities") {
		t.Fatalf("expected manual confirmation note, got %q", got)
	}
	if !strings.Contains(got, "[x] Session history") {
		t.Fatalf("expected completed session history item, got %q", got)
	}
}

func TestRunHistoryCommandWithNoSessions(t *testing.T) {
	t.Setenv("XTLDR_HISTORY_FILE", t.TempDir()+"/history.jsonl")
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"history"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), "No session history yet") {
		t.Fatalf("expected empty history message, got %q", out.String())
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

func TestRunNonInteractiveFlag(t *testing.T) {
	original := newGenerator
	t.Cleanup(func() { newGenerator = original })
	newGenerator = func() candidateGenerator {
		return fakeGenerator{
			candidates: []model.Candidate{
				{Command: "ls -la"},
				{Command: "echo done"},
			},
		}
	}

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := run([]string{"-n", "list files"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", code, errOut.String())
	}
	if got := out.String(); got != "ls -la\necho done\n" {
		t.Fatalf("unexpected non-interactive output: %q", got)
	}
	if errOut.String() != "" {
		t.Fatalf("expected empty stderr, got %q", errOut.String())
	}
}

func TestPrintCommandsSkipsEmpty(t *testing.T) {
	var out bytes.Buffer
	printCommands(&out, []model.Candidate{
		{Command: "  "},
		{Command: "pwd"},
		{Command: ""},
		{Command: "echo ok"},
	})

	if got := out.String(); got != "pwd\necho ok\n" {
		t.Fatalf("unexpected printed commands: %q", got)
	}
}

func TestAppendRefinement(t *testing.T) {
	got := appendRefinement("list files", "only include hidden files")
	want := "list files\nRefinement: only include hidden files"
	if got != want {
		t.Fatalf("unexpected refinement request: got %q want %q", got, want)
	}
}

func TestOutputSelectedCommand(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	copier := &fakeCopier{}

	if err := outputSelectedCommand(&out, &errOut, copier, "git status"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out.String(); got != "git status\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
	if !strings.Contains(errOut.String(), "Command copied to clipboard") {
		t.Fatalf("expected copy success message, got %q", errOut.String())
	}
	if copier.copied != "git status" {
		t.Fatalf("expected copied command, got %q", copier.copied)
	}
}

func TestOutputSelectedCommandCopyFailure(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	copier := &fakeCopier{err: errors.New("copy failed")}

	err := outputSelectedCommand(&out, &errOut, copier, "pwd")
	if err == nil {
		t.Fatalf("expected error when copy fails")
	}
	if got := out.String(); got != "pwd\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
	if !strings.Contains(errOut.String(), "Failed to copy selected command") {
		t.Fatalf("expected copy failure message, got %q", errOut.String())
	}
}
