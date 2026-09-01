package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// StatusBarModel renders the bottom metric and status bar.
type StatusBarModel struct {
	ChapterTitle string
	Metrics      domain.EditorMetrics

	Width  int
	styles theme.Styles
}

// NewStatusBarModel creates a new StatusBarModel.
func NewStatusBarModel(styles theme.Styles) StatusBarModel {
	return StatusBarModel{
		ChapterTitle: "No Chapter Selected",
		Metrics: domain.EditorMetrics{
			WordCount:   0,
			CharCount:   0,
			ReadingTime: 0,
			IsDirty:     false,
		},
		styles: styles,
	}
}

// Init initializes status bar.
func (m StatusBarModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming status messages.
func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.ChapterSelectedMsg:
		m.ChapterTitle = msg.Chapter.Title
		m.Metrics = service.CalculateMetrics(msg.Chapter.Content, false)

	case messages.TextChangedMsg:
		m.Metrics = msg.Metrics

	case messages.SaveCompletedMsg:
		if msg.Success {
			m.Metrics.IsDirty = false
		}
	}
	return m, nil
}

// SetWidth updates width of the status bar.
func (m *StatusBarModel) SetWidth(w int) {
	m.Width = w
}

// View renders the status bar.
func (m StatusBarModel) View() string {
	if m.Width <= 0 {
		return ""
	}

	// Title badge
	title := m.ChapterTitle
	if len(title) > 28 {
		title = title[:25] + "..."
	}
	titlePill := m.styles.StatusTitle.Render(" " + title + " ")

	// Save status badge
	var savePill string
	if m.Metrics.IsDirty {
		savePill = m.styles.StatusDirty.Render("[Modified*]")
	} else {
		savePill = m.styles.StatusSaved.Render("[Saved]")
	}

	// Hotkey hints
	hints := m.styles.StatusHint.Render("Tab: Switch | Ctrl+S: Save | n: New")

	leftSection := lipgloss.JoinHorizontal(lipgloss.Center, titlePill, " ", savePill, "  ", hints)

	// Metrics string
	metricsStr := fmt.Sprintf("%d words | %d chars | %s",
		m.Metrics.WordCount,
		m.Metrics.CharCount,
		service.FormatReadingTime(m.Metrics.ReadingTime),
	)
	rightSection := m.styles.StatusMetrics.Render(metricsStr + " ")

	// Sizing and spacing
	leftWidth := lipgloss.Width(leftSection)
	rightWidth := lipgloss.Width(rightSection)

	spaceWidth := m.Width - leftWidth - rightWidth
	if spaceWidth < 1 {
		spaceWidth = 1
	}

	spacer := lipgloss.NewStyle().
		Background(m.styles.StatusBarContainer.GetBackground()).
		Width(spaceWidth).
		Render("")

	barContent := lipgloss.JoinHorizontal(lipgloss.Center, leftSection, spacer, rightSection)

	return m.styles.StatusBarContainer.
		Width(m.Width).
		Render(barContent)
}
