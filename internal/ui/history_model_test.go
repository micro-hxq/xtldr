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
		{Request: "find large files", CreatedAt: time.Now()},
		{Request: "git status", CreatedAt: time.Now()},
	}
	m := NewHistoryModel(sessions, "")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := next.(HistoryModel)
	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = next.(HistoryModel)

	if updated.SelectedRequest() != "git status" {
		t.Fatalf("expected selected request git status, got %q", updated.SelectedRequest())
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
