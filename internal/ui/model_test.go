package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"xtldr/internal/model"
)

type mockCopier struct {
	copied string
	err    error
}

func (m *mockCopier) Copy(text string) error {
	m.copied = text
	return m.err
}

func sampleCandidates() []model.Candidate {
	return []model.Candidate{
		{Command: "ls -la", Title: "List files", Description: "List all files."},
		{Command: "pwd", Title: "Show path", Description: "Show cwd."},
	}
}

func TestUpDownNavigation(t *testing.T) {
	m := NewModel(sampleCandidates(), &mockCopier{}, false)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(Model)
	if updated.selected != 1 {
		t.Fatalf("expected selected index 1, got %d", updated.selected)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = next.(Model)
	if updated.selected != 0 {
		t.Fatalf("expected selected index 0, got %d", updated.selected)
	}
}

func TestCopySelectedCommand(t *testing.T) {
	copier := &mockCopier{}
	m := NewModel(sampleCandidates(), copier, false)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(Model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	updated = next.(Model)

	if copier.copied != "pwd" {
		t.Fatalf("expected copied command pwd, got %q", copier.copied)
	}
	if updated.statusText == "" {
		t.Fatalf("expected status text after copy")
	}
}

func TestEnterSelectsCommand(t *testing.T) {
	m := NewModel(sampleCandidates(), &mockCopier{}, false)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(Model)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(Model)

	if updated.SelectedCommand() != "pwd" {
		t.Fatalf("expected selected command pwd, got %q", updated.SelectedCommand())
	}
}

func TestDefaultViewHidesExplanationPanel(t *testing.T) {
	m := NewModel(sampleCandidates(), &mockCopier{}, false)
	view := m.View()
	if strings.Contains(view, "Command Explanation") {
		t.Fatalf("did not expect explanation panel by default")
	}
}

func TestEFlagViewShowsSeparatedExplanationPanel(t *testing.T) {
	m := NewModel(sampleCandidates(), &mockCopier{}, true)
	view := m.View()
	if !strings.Contains(view, "Command Explanation") {
		t.Fatalf("expected explanation panel when enabled")
	}
	if !strings.Contains(view, "Command Candidates") {
		t.Fatalf("expected command panel when explanation is enabled")
	}
}

func TestLoadingStateView(t *testing.T) {
	m := NewLoadingModel(func() ([]model.Candidate, error) {
		return sampleCandidates(), nil
	}, &mockCopier{}, false)

	view := m.View()
	if !strings.Contains(view, "Smart command generation in progress") {
		t.Fatalf("expected processing state in loading view")
	}
	if !strings.Contains(view, "live generation") {
		t.Fatalf("expected modern loading meta line")
	}

	next, _ := m.Update(candidatesLoadedMsg{candidates: sampleCandidates()})
	updated := next.(Model)
	if updated.loading {
		t.Fatalf("expected loading to be false after candidates loaded")
	}
	if len(updated.candidates) != 2 {
		t.Fatalf("expected candidates loaded into model")
	}
}
