package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// SpellMenuAction identifies what a menu row does when activated.
type SpellMenuAction int

const (
	SpellMenuSuggest SpellMenuAction = iota
	SpellMenuAddWord
	SpellMenuCopy
	SpellMenuCut
	SpellMenuPaste
)

// SpellMenuItem is one selectable row of the context menu.
type SpellMenuItem struct {
	Label  string
	Action SpellMenuAction
	// Replacement is set for suggestion rows.
	Replacement string
}

// MaxSpellSuggestions caps suggestion rows in the menu.
const MaxSpellSuggestions = 8

// SpellMenuModel is a centered context menu opened by right-click (or F2)
// on a misspelled word. It offers suggestions plus clipboard actions.
type SpellMenuModel struct {
	Active      bool
	Word        string
	WordStart   int
	WordEnd     int
	Searching   bool
	Suggestions []string
	Items       []SpellMenuItem
	Cursor      int

	width  int
	height int
	styles theme.Styles
}

// NewSpellMenuModel creates an inactive context menu.
func NewSpellMenuModel(styles theme.Styles) SpellMenuModel {
	return SpellMenuModel{styles: styles}
}

// SetSize updates the screen dimensions used to center the menu.
func (m *SpellMenuModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Open shows the menu for word at buffer range [start, end).
func (m *SpellMenuModel) Open(word string, start, end int) {
	m.Active = true
	m.Word = word
	m.WordStart = start
	m.WordEnd = end
	m.Searching = true
	m.Suggestions = nil
	m.Cursor = 0
	m.rebuildItems()
}

// SetSuggestions fills suggestion rows (async arrival); stale results for a
// different word/range are ignored by the caller via word+range match.
func (m *SpellMenuModel) SetSuggestions(word string, start, end int, suggestions []string) {
	if !m.Active || m.Word != word || m.WordStart != start || m.WordEnd != end {
		return
	}
	m.Searching = false
	if len(suggestions) > MaxSpellSuggestions {
		suggestions = suggestions[:MaxSpellSuggestions]
	}
	m.Suggestions = suggestions
	if m.Cursor >= len(m.Items) {
		m.Cursor = 0
	}
	m.rebuildItems()
}

// Close hides the menu.
func (m *SpellMenuModel) Close() {
	m.Active = false
	m.Items = nil
	m.Suggestions = nil
	m.Cursor = 0
}

func (m *SpellMenuModel) rebuildItems() {
	var items []SpellMenuItem
	for _, s := range m.Suggestions {
		items = append(items, SpellMenuItem{
			Label:       fmt.Sprintf("  %s", s),
			Action:      SpellMenuSuggest,
			Replacement: s,
		})
	}
	items = append(items,
		SpellMenuItem{Label: fmt.Sprintf("＋ Añadir %q al diccionario", m.Word), Action: SpellMenuAddWord},
		SpellMenuItem{Label: "⧉ Copiar", Action: SpellMenuCopy},
		SpellMenuItem{Label: "✂ Cortar", Action: SpellMenuCut},
		SpellMenuItem{Label: "⎘ Pegar", Action: SpellMenuPaste},
	)
	m.Items = items
	if m.Cursor >= len(items) && len(items) > 0 {
		m.Cursor = len(items) - 1
	}
}

// activateCmd builds the message for the current row.
func (m SpellMenuModel) activateCmd() tea.Cmd {
	if len(m.Items) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Items) {
		return nil
	}
	item := m.Items[m.Cursor]
	switch item.Action {
	case SpellMenuSuggest:
		start, end, repl := m.WordStart, m.WordEnd, item.Replacement
		return func() tea.Msg {
			return messages.ApplySpellSuggestionMsg{Start: start, End: end, Replacement: repl}
		}
	case SpellMenuAddWord:
		word := m.Word
		return func() tea.Msg {
			return messages.AddCustomWordMsg{Word: word}
		}
	case SpellMenuCopy:
		return func() tea.Msg { return messages.EditorCopyMsg{} }
	case SpellMenuCut:
		return func() tea.Msg { return messages.EditorCutMsg{} }
	case SpellMenuPaste:
		return func() tea.Msg { return messages.EditorPasteMsg{} }
	}
	return nil
}

// Update handles navigation and activation for the menu.
func (m SpellMenuModel) Update(msg tea.Msg) (SpellMenuModel, tea.Cmd) {
	if !m.Active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Close()
			return m, nil
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
			return m, nil
		case "down", "j":
			if m.Cursor < len(m.Items)-1 {
				m.Cursor++
			}
			return m, nil
		case "enter":
			cmd := m.activateCmd()
			m.Close()
			return m, cmd
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			if idx, ok := m.rowAt(msg.X, msg.Y); ok {
				m.Cursor = idx
				cmd := m.activateCmd()
				m.Close()
				return m, cmd
			}
		}
	}
	return m, nil
}

// boxView renders the menu box without screen positioning.
func (m SpellMenuModel) boxView() string {
	const boxWidth = 44
	title := fmt.Sprintf("«%s»", m.Word)
	if m.Word == "" {
		title = "Portapapeles"
	}
	title = truncateCells(title, boxWidth-4)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.CurrentTheme.Highlight).
		Width(boxWidth - 4)
	var rows []string
	rows = append(rows, titleStyle.Render(title))
	if m.Searching {
		rows = append(rows, m.styles.ListSubtitle.Render("  buscando sugerencias…"))
	} else if len(m.Suggestions) == 0 {
		rows = append(rows, m.styles.ListSubtitle.Render("  sin sugerencias"))
	}
	for i, item := range m.Items {
		label := truncateCells(item.Label, boxWidth-6)
		if i == m.Cursor {
			rows = append(rows, m.styles.ListItemActive.Width(boxWidth-4).Render(label))
		} else {
			rows = append(rows, m.styles.ListItem.Width(boxWidth-4).Render(label))
		}
	}
	rows = append(rows, m.styles.ListSubtitle.Render("enter elegir · esc cerrar"))
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.CurrentTheme.BorderFocused).
		Background(theme.CurrentTheme.CardBg).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)
}

// truncateCells shortens s to fit cells, rune-safe with an ellipsis.
func truncateCells(s string, maxCells int) string {
	if lipgloss.Width(s) <= maxCells {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)+"...") > maxCells {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}

// rowAt maps absolute screen cells to an item index. The box layout is fully
// deterministic (fixed width, single-line rows), so this stays exact.
func (m SpellMenuModel) rowAt(x, y int) (int, bool) {
	box := m.boxView()
	boxW := lipgloss.Width(box)
	boxH := lipgloss.Height(box)
	ox := (m.width - boxW) / 2
	oy := (m.height - boxH) / 2
	relY := y - oy
	// Row 0: top border; row 1: padding-top; row 2: title; rows 3..: items
	// (each item row may be preceded by searching/empty hint rows).
	headerRows := 3
	if m.Searching || len(m.Suggestions) == 0 {
		headerRows = 4
	}
	idx := relY - headerRows
	if idx < 0 || idx >= len(m.Items) {
		return 0, false
	}
	if x < ox || x >= ox+boxW {
		return 0, false
	}
	return idx, true
}

// View renders the centered overlay (empty when inactive).
func (m SpellMenuModel) View() string {
	if !m.Active {
		return ""
	}
	if m.width <= 0 || m.height <= 0 {
		return m.boxView()
	}
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		m.boxView(),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(theme.CurrentTheme.Background),
	)
}
