package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"xtldr/internal/history"
)

func TestHistoryModelSelectsRequest(t *testing.T) {
	sessions := []history.Session{
		{Request: "find large files", Command: "find . -size +1G", CreatedAt: time.Now()},
		{Request: "git status", Command: "git status", CreatedAt: time.Now()},
	}
	m := NewHistoryModel(sessions, "")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(HistoryModel)

	if updated.SelectedRequest() != "git status" {
		t.Fatalf("expected selected request git status, got %q", updated.SelectedRequest())
	}
	if updated.SelectedSession().Command != "git status" {
		t.Fatalf("expected selected command git status, got %q", updated.SelectedSession().Command)
	}
}

func TestHistoryViewContainsHeaderAndQuery(t *testing.T) {
	m := NewHistoryModel([]history.Session{{Request: "git status"}}, "git")
	view := m.View()
	if !strings.Contains(view, "Session History") {
		t.Fatalf("expected history title in view")
	}
	if !strings.Contains(view, "Search: git") {
		t.Fatalf("expected query in history view")
	}
}

func TestHistoryInteractiveSearchFiltersSessions(t *testing.T) {
	m := NewHistoryModel([]history.Session{
		{Request: "git status"},
		{Request: "find large files"},
	}, "")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	updated = next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	updated = next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	updated = next.(HistoryModel)

	view := updated.View()
	if !strings.Contains(view, "Search: git") {
		t.Fatalf("expected active search query in view")
	}
	if !strings.Contains(view, "git status") {
		t.Fatalf("expected git session in filtered view")
	}
	if strings.Contains(view, "find large files") {
		t.Fatalf("did not expect non-matching session in filtered view")
	}
}

func TestHistorySearchModeEnterDoesNotSelect(t *testing.T) {
	m := NewHistoryModel([]history.Session{
		{Request: "git status"},
		{Request: "go test ./..."},
	}, "")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	updated := next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	updated = next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	updated = next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	updated = next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(HistoryModel)

	if updated.SelectedRequest() != "" {
		t.Fatalf("expected no request selected when pressing enter in search mode")
	}
}
