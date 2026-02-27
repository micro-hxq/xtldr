package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"xtldr/internal/history"
)

type HistoryModel struct {
	allSessions []history.Session
	sessions    []history.Session
	selected    int
	chosen      history.Session
	query       string
	searching   bool
}

func NewHistoryModel(sessions []history.Session, query string) HistoryModel {
	q := strings.TrimSpace(query)
	m := HistoryModel{
		allSessions: sessions,
		query:       q,
	}
	m.applyFilter()
	return m
}

func (m HistoryModel) Init() tea.Cmd { return nil }

func (m HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.searching = false
				return m, nil
			case "backspace":
				if len(m.query) > 0 {
					queryRunes := []rune(m.query)
					m.query = string(queryRunes[:len(queryRunes)-1])
					m.applyFilter()
				}
				return m, nil
			case "enter":
				m.searching = false
				return m, nil
			}
			if msg.Type == tea.KeyRunes {
				m.query += string(msg.Runes)
				m.applyFilter()
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			m.searching = true
			return m, nil
		case "up", "k":
			if len(m.sessions) == 0 {
				return m, nil
			}
			m.selected = (m.selected - 1 + len(m.sessions)) % len(m.sessions)
		case "down", "j":
			if len(m.sessions) == 0 {
				return m, nil
			}
			m.selected = (m.selected + 1) % len(m.sessions)
		case "enter":
			if len(m.sessions) == 0 {
				return m, tea.Quit
			}
			m.chosen = m.sessions[m.selected]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m HistoryModel) View() string {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	indexStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	selectedMarkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("219")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	badgeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Padding(0, 1).Bold(true)

	lines := []string{
		badgeStyle.Render(" xtldr history "),
		"",
		titleStyle.Render("🕘 Session History"),
		subtitleStyle.Render("🔁 Browse previous requests and press Enter to reuse."),
	}
	if m.query != "" {
		lines = append(lines, "", metaStyle.Render("🔎 Search: "+m.query))
	} else if m.searching {
		lines = append(lines, "", metaStyle.Render("🔎 Search:"))
	}
	lines = append(lines, "")

	if len(m.sessions) == 0 {
		lines = append(lines, metaStyle.Render("No history sessions found."))
	} else {
		for i, session := range m.sessions {
			when := ""
			if !session.CreatedAt.IsZero() {
				when = session.CreatedAt.Local().Format(time.RFC3339)
			}
			entry := session.Request
			if i == m.selected {
				lines = append(lines, selectedMarkStyle.Render("▶ ")+selectedStyle.Render(entry))
			} else {
				lines = append(lines, " "+indexStyle.Render(fmt.Sprintf("%d) ", i+1))+normalStyle.Render(entry))
			}
			if when != "" {
				lines = append(lines, "   "+metaStyle.Render("🕒 "+when))
			}
			if strings.TrimSpace(session.WorkingDir) != "" {
				lines = append(lines, "   "+metaStyle.Render("📁 "+session.WorkingDir))
			}
			lines = append(lines, "")
		}
	}
	if m.searching {
		lines = append(lines, hintStyle.Render("💡 Search mode: type to filter, Backspace delete, Enter/Esc finish search."))
	} else {
		lines = append(lines, hintStyle.Render("💡 Tip: / search, ↑/↓ or j/k navigate, Enter reuse, q quit."))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Background(lipgloss.Color("236")).
		Padding(1, 3).
		Width(110)
	return box.Render(strings.Join(lines, "\n"))
}

func (m HistoryModel) SelectedRequest() string {
	return m.chosen.Request
}

func (m HistoryModel) SelectedSession() history.Session {
	return m.chosen
}

func (m *HistoryModel) applyFilter() {
	m.sessions = history.Search(m.allSessions, m.query)
	if len(m.sessions) == 0 {
		m.selected = 0
		return
	}
	if m.selected >= len(m.sessions) || m.selected < 0 {
		m.selected = 0
	}
}
