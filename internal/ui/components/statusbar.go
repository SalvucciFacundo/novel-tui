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

	// 1. Left Section: Styled command badges
	badgeStyle := m.styles.StatusCommandBadge
	badgeHome := badgeStyle.Render("[Ctrl+H: Inicio]")
	badgeAI := badgeStyle.Render("[Ctrl+A: IA]")
	badgeSave := badgeStyle.Render("[Ctrl+S: Guardar]")
	badgeNew := badgeStyle.Render("[Ctrl+N: Nuevo Cap]")
	badgeTab := badgeStyle.Render("[Tab: Panel]")

	leftSection := lipgloss.JoinHorizontal(lipgloss.Center,
		badgeHome, " ",
		badgeAI, " ",
		badgeSave, " ",
		badgeNew, " ",
		badgeTab,
	)

	// 2. Right Section: Metrics + Save badge
	readingMins := m.Metrics.ReadingTime
	if readingMins < 1 && m.Metrics.WordCount > 0 {
		readingMins = 1
	}
	metricsStr := fmt.Sprintf("%d palabras | %d caracteres | ~%d min",
		m.Metrics.WordCount,
		m.Metrics.CharCount,
		readingMins,
	)
	renderedMetrics := m.styles.StatusMetrics.Render(metricsStr)

	var savePill string
	if m.Metrics.IsDirty {
		savePill = m.styles.StatusDirty.Render("[Modificado*]")
	} else {
		savePill = m.styles.StatusSaved.Render("[Guardado]")
	}

	rightSection := lipgloss.JoinHorizontal(lipgloss.Center, renderedMetrics, " ", savePill, " ")

	// 3. Layout calculation
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
