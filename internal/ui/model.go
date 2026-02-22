package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"xtldr/internal/model"
)

type Copier interface {
	Copy(text string) error
}

type Loader func() ([]model.Candidate, error)

type candidatesLoadedMsg struct {
	candidates []model.Candidate
	err        error
}

type spinnerTickMsg struct{}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var loadingStages = []string{
	"🧭 Understanding your request",
	"🛡️ Designing safe command options",
	"🧪 Preparing executable candidates",
}

type Model struct {
	candidates      []model.Candidate
	selected        int
	copier          Copier
	loader          Loader
	showExplanation bool
	loading         bool
	loadingStarted  time.Time
	loadErr         string
	statusText      string
	spinIndex       int
	chosen          string
}

func NewModel(candidates []model.Candidate, copier Copier, showExplanation bool) Model {
	return Model{
		candidates:      candidates,
		copier:          copier,
		showExplanation: showExplanation,
	}
}

func NewLoadingModel(loader Loader, copier Copier, showExplanation bool) Model {
	return Model{
		copier:          copier,
		loader:          loader,
		showExplanation: showExplanation,
		loading:         true,
		loadingStarted:  time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.loader == nil {
		return nil
	}
	return tea.Batch(m.loadCandidatesCmd(), spinnerTickCmd())
}

func spinnerTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m Model) loadCandidatesCmd() tea.Cmd {
	loader := m.loader
	return func() tea.Msg {
		candidates, err := loader()
		return candidatesLoadedMsg{
			candidates: candidates,
			err:        err,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerTickMsg:
		if m.loading {
			m.spinIndex = (m.spinIndex + 1) % len(spinnerFrames)
			return m, spinnerTickCmd()
		}
	case candidatesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.loadErr = msg.err.Error()
			m.statusText = "❌ Failed to generate candidates."
			return m, nil
		}
		m.candidates = msg.candidates
		m.statusText = fmt.Sprintf("✨ Generated %d candidates.", len(msg.candidates))
	case tea.KeyMsg:
		if m.loading {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}
		if m.loadErr != "" {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.move(-1)
		case "down", "j":
			m.move(1)
		case "enter":
			m.confirmSelected()
			return m, tea.Quit
		case "c":
			m.copySelected()
		}
	}
	return m, nil
}

func (m *Model) move(step int) {
	if len(m.candidates) == 0 {
		return
	}
	m.selected = (m.selected + step + len(m.candidates)) % len(m.candidates)
}

func (m *Model) copySelected() {
	if len(m.candidates) == 0 {
		m.statusText = "⚠️ No command available to copy."
		return
	}
	if m.copier == nil {
		m.statusText = "⚠️ Clipboard is not configured."
		return
	}
	command := m.candidates[m.selected].Command
	if err := m.copier.Copy(command); err != nil {
		m.statusText = fmt.Sprintf("❌ Copy failed: %v", err)
		return
	}
	m.statusText = "📋 Copied selected command."
}

func (m *Model) confirmSelected() {
	if len(m.candidates) == 0 {
		return
	}
	m.chosen = m.candidates[m.selected].Command
}

func (m Model) SelectedCommand() string {
	return m.chosen
}

func (m Model) View() string {
	if m.loading {
		return m.renderLoading()
	}
	if m.loadErr != "" {
		return m.renderError()
	}
	if len(m.candidates) == 0 {
		return "No command candidates.\n"
	}

	commandPanel := m.renderCommandPanel()
	footer := m.renderFooter()
	if !m.showExplanation {
		return lipgloss.JoinVertical(lipgloss.Left, commandPanel, "", footer)
	}
	explanationPanel := m.renderExplanationPanel()
	return lipgloss.JoinVertical(lipgloss.Left, commandPanel, "", explanationPanel, "", footer)
}

func (m Model) renderLoading() string {
	badgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Bold(true)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	spinnerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	stageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("222"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("159"))
	tipStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	frame := spinnerFrames[m.spinIndex%len(spinnerFrames)]
	stage := loadingStages[(m.spinIndex/10)%len(loadingStages)]
	bar := animatedLoadingBar(m.spinIndex, 10)
	elapsed := time.Since(m.loadingStarted)
	if m.loadingStarted.IsZero() {
		elapsed = 0
	}

	lines := []string{
		badgeStyle.Render(" xtldr "),
		"",
		titleStyle.Render("🚀 Smart command generation in progress"),
		subtitleStyle.Render("🤖 Please hold on while Copilot prepares polished suggestions."),
		"",
		spinnerStyle.Render(frame+"  ⚙️ Working: ") + stageStyle.Render(stage),
		progressStyle.Render("📊 "+bar) + metaStyle.Render("  📡 live generation"),
		metaStyle.Render(fmt.Sprintf("⏱ Elapsed: %s", elapsed.Truncate(time.Second))),
		"",
		tipStyle.Render("💡 Tip: press q to cancel"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Background(lipgloss.Color("236")).
		Padding(1, 3).
		Width(90)
	return box.Render(strings.Join(lines, "\n"))
}

func animatedLoadingBar(step, width int) string {
	if width <= 0 {
		return ""
	}
	pos := step % width
	parts := make([]string, width)
	for i := 0; i < width; i++ {
		if i <= pos {
			parts[i] = "▰"
		} else {
			parts[i] = "▱"
		}
	}
	return strings.Join(parts, "")
}

func (m Model) renderError() string {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	lines := []string{
		titleStyle.Render("❌ Failed to generate command candidates"),
		"",
		m.loadErr,
		"",
		muted.Render("Press q to quit 👋."),
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("160")).
		Padding(1, 2).
		Width(90)
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderCommandPanel() string {
	badgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Bold(true)
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Bold(true)
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	selectedCmdStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	normalCmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	indexStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	selectedMarkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("219")).Bold(true)
	selectedDescStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("110"))

	lines := []string{
		badgeStyle.Render(" xtldr results "),
		"",
		titleStyle.Render("🧾 Command Candidates"),
		subtitleStyle.Render("🎯 Review options and pick the one you want to run."),
		"",
	}
	for i, candidate := range m.candidates {
		if i == m.selected {
			marker := selectedMarkStyle.Render("▶ ")
			lines = append(lines, marker+selectedCmdStyle.Render(candidate.Command))
			if candidate.Description != "" {
				lines = append(lines, selectedDescStyle.Render("   💬 "+candidate.Description))
			}
			lines = append(lines, "")
			continue
		}
		prefix := indexStyle.Render(fmt.Sprintf("%d) ", i+1))
		lines = append(lines, " "+prefix+normalCmdStyle.Render(candidate.Command))
	}
	lines = append(lines, hintStyle.Render("💡 Tip: use ↑/↓ or j/k to navigate, Enter to confirm."))

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("69")).
		Background(lipgloss.Color("236")).
		Padding(1, 3).
		Width(110)
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderExplanationPanel() string {
	selected := m.candidates[m.selected]

	badgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("97")).
		Padding(0, 1).
		Bold(true)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("183"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	commandStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	argStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("153"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("60"))

	lines := []string{
		badgeStyle.Render(" xtldr explain "),
		"",
		titleStyle.Render("🧠 Command Explanation"),
		subtitleStyle.Render("🔍 Detailed breakdown for the currently selected command."),
		"",
		labelStyle.Render("🎯 Title      ") + valueStyle.Render(fallback(selected.Title, selected.Command)),
		labelStyle.Render("⚡ Executable ") + commandStyle.Render(selected.Command),
		labelStyle.Render("📝 Summary    ") + valueStyle.Render(selected.Description),
		dividerStyle.Render(strings.Repeat("─", 72)),
		"",
		labelStyle.Render("🧩 Parameters:"),
	}

	if len(selected.Args) == 0 {
		lines = append(lines, "  "+mutedStyle.Render("ℹ️ No explicit parameters were detected."))
	} else {
		for _, arg := range selected.Args {
			lines = append(lines, "  "+argStyle.Render("- "+fallback(arg.Name, "(positional)")))
			if arg.Example != "" {
				lines = append(lines, "    "+mutedStyle.Render("example: "+arg.Example))
			}
			if arg.Meaning != "" {
				lines = append(lines, "    "+valueStyle.Render("meaning: "+arg.Meaning))
			}
		}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("97")).
		Background(lipgloss.Color("235")).
		Padding(1, 3).
		Width(110)
	return box.Render(strings.Join(lines, "\n"))
}

func (m Model) renderFooter() string {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("147"))
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("84")).Bold(true)
	help := helpStyle.Render("⌨️ Keys: ↑/↓ or j/k navigate, Enter prints + copies, c copy, q quit | -e hides Explanation")
	if m.statusText == "" {
		return help
	}
	return help + " | " + statusStyle.Render(m.statusText)
}

func fallback(primary, alternative string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return alternative
}
