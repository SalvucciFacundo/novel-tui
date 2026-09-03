package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// Focus targets in SearchModalModel
const (
	searchFocusInput = iota
	searchFocusReplace
	searchFocusResults
)

// SearchModalKeyMap defines keybindings for the search modal.
type SearchModalKeyMap struct {
	Close          key.Binding
	Submit         key.Binding
	NextFocus      key.Binding
	PrevFocus      key.Binding
	ToggleReplace  key.Binding
	ToggleCase     key.Binding
	NextResult     key.Binding
	PrevResult     key.Binding
}

// DefaultSearchModalKeyMap returns standard search modal keybindings.
func DefaultSearchModalKeyMap() SearchModalKeyMap {
	return SearchModalKeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cerrar"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirmar / saltar"),
		),
		NextFocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "siguiente campo"),
		),
		PrevFocus: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "campo anterior"),
		),
		ToggleReplace: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "modo reemplazar"),
		),
		ToggleCase: key.NewBinding(
			key.WithKeys("ctrl+t", "alt+c"),
			key.WithHelp("ctrl+t", "sensible a mayúsculas"),
		),
		NextResult: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "siguiente coincidencia"),
		),
		PrevResult: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "anterior coincidencia"),
		),
	}
}

// SearchModalModel is a floating dialog for global novel search and replace.
type SearchModalModel struct {
	Active        bool
	IsReplaceMode bool
	CaseSensitive bool

	SearchInput  textinput.Model
	ReplaceInput textinput.Model

	Matches       []domain.SearchMatch
	SelectedMatch int
	FocusIndex    int // 0: SearchInput, 1: ReplaceInput, 2: Results

	SearchFunc func(query string, caseSensitive bool) ([]domain.SearchMatch, error)

	width  int
	height int
	styles theme.Styles
	keys   SearchModalKeyMap
}

// NewSearchModalModel constructs a new SearchModalModel.
func NewSearchModalModel(styles theme.Styles) SearchModalModel {
	si := textinput.New()
	si.Placeholder = "Buscar en toda la novela..."
	si.Prompt = "🔍 "
	si.CharLimit = 120
	si.Width = 50
	si.Focus()

	ri := textinput.New()
	ri.Placeholder = "Reemplazar con..."
	ri.Prompt = "🔄 "
	ri.CharLimit = 120
	ri.Width = 50

	return SearchModalModel{
		Active:        false,
		IsReplaceMode: false,
		CaseSensitive: false,
		SearchInput:   si,
		ReplaceInput:  ri,
		Matches:       []domain.SearchMatch{},
		SelectedMatch: 0,
		FocusIndex:    searchFocusInput,
		styles:        styles,
		keys:          DefaultSearchModalKeyMap(),
	}
}

// Show opens the search modal and focuses the search input.
func (m *SearchModalModel) Show() {
	m.Active = true
	m.FocusIndex = searchFocusInput
	m.SearchInput.Focus()
	m.ReplaceInput.Blur()
	m.PerformSearch()
}

// PerformSearch executes SearchFunc if configured and updates Matches.
func (m *SearchModalModel) PerformSearch() {
	if m.SearchFunc != nil {
		matches, err := m.SearchFunc(m.SearchInput.Value(), m.CaseSensitive)
		if err == nil {
			m.SetMatches(matches)
		}
	}
}

// Hide closes and resets the search modal.
func (m *SearchModalModel) Hide() {
	m.Active = false
	m.SearchInput.Blur()
	m.ReplaceInput.Blur()
}

// SetMatches updates the search results and adjusts the selection cursor.
func (m *SearchModalModel) SetMatches(matches []domain.SearchMatch) {
	m.Matches = matches
	if len(m.Matches) == 0 {
		m.SelectedMatch = 0
	} else if m.SelectedMatch >= len(m.Matches) {
		m.SelectedMatch = len(m.Matches) - 1
	}
}

// SetSize updates modal viewport dimensions.
func (m *SearchModalModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Init initializes the search modal.
func (m SearchModalModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles events for the search modal.
func (m SearchModalModel) Update(msg tea.Msg) (SearchModalModel, tea.Cmd) {
	switch msg.(type) {
	case messages.OpenGlobalSearchMsg:
		m.Show()
		return m, textinput.Blink

	case messages.CloseGlobalSearchMsg:
		m.Hide()
		return m, nil
	}

	if !m.Active {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Close):
			m.Hide()
			return m, func() tea.Msg { return messages.CloseGlobalSearchMsg{} }

		case key.Matches(msg, m.keys.ToggleReplace):
			if !m.IsReplaceMode {
				m.IsReplaceMode = true
				m.FocusIndex = searchFocusReplace
				m.SearchInput.Blur()
				cmds = append(cmds, m.ReplaceInput.Focus())
			} else if m.FocusIndex == searchFocusReplace && strings.TrimSpace(m.SearchInput.Value()) != "" {
				// If already in replace mode and focused on replace input, execute replace
				query := m.SearchInput.Value()
				replacement := m.ReplaceInput.Value()
				caseSensitive := m.CaseSensitive
				m.Hide()
				return m, func() tea.Msg {
					return messages.GlobalReplaceMsg{
						Query:         query,
						Replacement:   replacement,
						CaseSensitive: caseSensitive,
					}
				}
			} else {
				m.IsReplaceMode = !m.IsReplaceMode
				if !m.IsReplaceMode && m.FocusIndex == searchFocusReplace {
					m.FocusIndex = searchFocusInput
					m.ReplaceInput.Blur()
					cmds = append(cmds, m.SearchInput.Focus())
				}
			}
			return m, tea.Batch(cmds...)

		case key.Matches(msg, m.keys.ToggleCase):
			m.CaseSensitive = !m.CaseSensitive
			m.PerformSearch()
			return m, nil

		case key.Matches(msg, m.keys.NextFocus):
			if m.IsReplaceMode {
				m.FocusIndex = (m.FocusIndex + 1) % 3
			} else {
				if m.FocusIndex == searchFocusInput {
					m.FocusIndex = searchFocusResults
				} else {
					m.FocusIndex = searchFocusInput
				}
			}
			m.updateFocus()
			return m, nil

		case key.Matches(msg, m.keys.PrevFocus):
			if m.IsReplaceMode {
				m.FocusIndex = (m.FocusIndex + 2) % 3
			} else {
				if m.FocusIndex == searchFocusInput {
					m.FocusIndex = searchFocusResults
				} else {
					m.FocusIndex = searchFocusInput
				}
			}
			m.updateFocus()
			return m, nil

		case key.Matches(msg, m.keys.NextResult):
			if len(m.Matches) > 0 {
				if m.SelectedMatch < len(m.Matches)-1 {
					m.SelectedMatch++
				}
			}
			if m.FocusIndex != searchFocusResults && (msg.Type == tea.KeyDown) {
				// Allow arrow down to move result selection without switching focus entirely
				return m, nil
			}

		case key.Matches(msg, m.keys.PrevResult):
			if len(m.Matches) > 0 {
				if m.SelectedMatch > 0 {
					m.SelectedMatch--
				}
			}
			if m.FocusIndex != searchFocusResults && (msg.Type == tea.KeyUp) {
				return m, nil
			}

		case key.Matches(msg, m.keys.Submit):
			if m.FocusIndex == searchFocusReplace {
				query := m.SearchInput.Value()
				replacement := m.ReplaceInput.Value()
				caseSensitive := m.CaseSensitive
				if strings.TrimSpace(query) != "" {
					m.Hide()
					return m, func() tea.Msg {
						return messages.GlobalReplaceMsg{
							Query:         query,
							Replacement:   replacement,
							CaseSensitive: caseSensitive,
						}
					}
				}
			} else {
				if len(m.Matches) > 0 && m.SelectedMatch >= 0 && m.SelectedMatch < len(m.Matches) {
					targetMatch := m.Matches[m.SelectedMatch]
					m.Hide()
					return m, func() tea.Msg {
						return messages.JumpToMatchMsg{Match: targetMatch}
					}
				}
			}
		}
	}

	// Update text inputs based on focus
	if m.FocusIndex == searchFocusInput {
		prevVal := m.SearchInput.Value()
		var tiCmd tea.Cmd
		m.SearchInput, tiCmd = m.SearchInput.Update(msg)
		cmds = append(cmds, tiCmd)
		if m.SearchInput.Value() != prevVal {
			m.PerformSearch()
		}
	} else if m.FocusIndex == searchFocusReplace {
		var tiCmd tea.Cmd
		m.ReplaceInput, tiCmd = m.ReplaceInput.Update(msg)
		cmds = append(cmds, tiCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *SearchModalModel) updateFocus() {
	switch m.FocusIndex {
	case searchFocusInput:
		m.SearchInput.Focus()
		m.ReplaceInput.Blur()
	case searchFocusReplace:
		m.SearchInput.Blur()
		m.ReplaceInput.Focus()
	case searchFocusResults:
		m.SearchInput.Blur()
		m.ReplaceInput.Blur()
	}
}

// View renders the centered floating search & replace modal.
func (m SearchModalModel) View() string {
	if !m.Active {
		return ""
	}

	modalWidth := 74
	if m.width > 0 && modalWidth > m.width-4 {
		modalWidth = m.width - 4
	}
	if modalWidth < 40 {
		modalWidth = 40
	}

	innerWidth := modalWidth - 6
	if innerWidth < 30 {
		innerWidth = 30
	}

	cardBg := theme.CurrentTheme.CardBg

	// 1. Header / Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Highlight).
		Background(cardBg).
		Width(innerWidth)

	modeText := "Búsqueda Global"
	if m.IsReplaceMode {
		modeText = "Búsqueda y Reemplazo Global"
	}
	title := titleStyle.Render("🔍 " + modeText)

	// 2. Search Input Box
	searchBoxBorder := theme.CurrentTheme.BorderBlurred
	if m.FocusIndex == searchFocusInput {
		searchBoxBorder = theme.CurrentTheme.BorderFocused
	}
	searchBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(searchBoxBorder).
		Background(cardBg).
		Padding(0, 1).
		Width(innerWidth).
		Render(m.SearchInput.View())

	var replaceBox string
	if m.IsReplaceMode {
		replaceBoxBorder := theme.CurrentTheme.BorderBlurred
		if m.FocusIndex == searchFocusReplace {
			replaceBoxBorder = theme.CurrentTheme.BorderFocused
		}
		replaceBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(replaceBoxBorder).
			Background(cardBg).
			Padding(0, 1).
			Width(innerWidth).
			MarginTop(1).
			Render(m.ReplaceInput.View())
	}

	// 3. Status Badge & Options
	var badgeText string
	query := strings.TrimSpace(m.SearchInput.Value())
	if query == "" {
		badgeText = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted).Render("Escribe un término para buscar")
	} else if len(m.Matches) == 0 {
		badgeText = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Warning).Render("⚠ Sin coincidencias")
	} else {
		chapMap := make(map[string]bool)
		for _, match := range m.Matches {
			chapMap[match.ChapterID] = true
		}
		numMatches := len(m.Matches)
		numChaps := len(chapMap)

		matchWord := "coincidencias"
		if numMatches == 1 {
			matchWord = "coincidencia"
		}
		chapWord := "capítulos"
		if numChaps == 1 {
			chapWord = "capítulo"
		}
		badgeText = lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.CurrentTheme.Success).
			Render(fmt.Sprintf("✓ %d %s en %d %s", numMatches, matchWord, numChaps, chapWord))
	}

	caseOpt := "[Aa: No sensible]"
	if m.CaseSensitive {
		caseOpt = "[Aa: Sensible]"
	}
	caseBadge := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted).Render(caseOpt)

	statusLine := lipgloss.JoinHorizontal(lipgloss.Center,
		badgeText,
		"  ",
		caseBadge,
	)

	// 4. Results List
	var resultsList string
	if len(m.Matches) > 0 {
		maxVisible := 6
		startIdx := 0
		if m.SelectedMatch >= maxVisible {
			startIdx = m.SelectedMatch - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(m.Matches) {
			endIdx = len(m.Matches)
		}

		var renderedItems []string
		for i := startIdx; i < endIdx; i++ {
			match := m.Matches[i]
			prefix := "  "
			itemStyle := lipgloss.NewStyle().Foreground(theme.CurrentTheme.Foreground)
			if i == m.SelectedMatch {
				prefix = "❯ "
				itemStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(theme.CurrentTheme.Highlight).
					Background(theme.CurrentTheme.BorderBlurred)
			}

			chapTag := lipgloss.NewStyle().
				Foreground(theme.CurrentTheme.Secondary).
				Render(fmt.Sprintf("[%s L%d]", match.ChapterTitle, match.LineNumber))

			linePreview := strings.TrimSpace(match.LineText)
			if len(linePreview) > 40 {
				linePreview = linePreview[:37] + "..."
			}

			row := fmt.Sprintf("%s%s %s", prefix, chapTag, linePreview)
			renderedItems = append(renderedItems, itemStyle.Width(innerWidth).Render(row))
		}

		resultsBoxBorder := theme.CurrentTheme.BorderBlurred
		if m.FocusIndex == searchFocusResults {
			resultsBoxBorder = theme.CurrentTheme.BorderFocused
		}
		resultsList = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(resultsBoxBorder).
			Background(cardBg).
			Padding(0, 1).
			Width(innerWidth).
			MarginTop(1).
			Render(strings.Join(renderedItems, "\n"))
	}

	// 5. Help Footer
	helpStyle := lipgloss.NewStyle().
		Foreground(theme.CurrentTheme.Muted).
		Background(cardBg).
		Width(innerWidth).
		MarginTop(1)

	var helpText string
	if m.IsReplaceMode {
		helpText = helpStyle.Render("[Tab] Foco • [Enter] Reemplazar todo • [Ctrl+R] Ocultar reemplazo • [Esc] Cerrar")
	} else {
		helpText = helpStyle.Render("[Tab] Foco • [Enter] Saltar a línea • [Ctrl+R] Modo reemplazo • [Esc] Cerrar")
	}

	// Assemble modal components
	var elements []string
	elements = append(elements, title, searchBox)
	if replaceBox != "" {
		elements = append(elements, replaceBox)
	}
	elements = append(elements, statusLine)
	if resultsList != "" {
		elements = append(elements, resultsList)
	}
	elements = append(elements, helpText)

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)
	content = lipgloss.NewStyle().Background(cardBg).Render(content)

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Background(cardBg).
		Padding(1, 2).
		Width(modalWidth)

	dialog := cardStyle.Render(content)

	if m.width == 0 || m.height == 0 {
		return dialog
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}
