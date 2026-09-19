package components

import (
	"sort"
	"strings"
	"unicode"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
	"github.com/SalvucciFacundo/novel-tui/internal/service/spell"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// EditorKeyMap defines keybindings for the editor.
type EditorKeyMap struct {
	Save key.Binding
}

// DefaultEditorKeyMap returns standard editor keybindings.
func DefaultEditorKeyMap() EditorKeyMap {
	return EditorKeyMap{
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save chapter"),
		),
	}
}

// EditorModel wraps textarea.Model and provides rich novel writing features:
// spellcheck underlines, keyboard/mouse text selection, clipboard, and a
// right-click context menu with suggestions.
type EditorModel struct {
	textarea textarea.Model

	ActiveChapter domain.Chapter
	Metrics       domain.EditorMetrics
	IsDirty       bool
	Focused       bool

	Width  int
	Height int

	// Spellcheck state (checker arrives async via SpellReadyMsg).
	spellChecker *spell.Checker
	spellVal     string
	spellBad     map[string]struct{}

	// Text selection as buffer rune offsets (anchor + cursor side).
	selActive bool
	selAnchor int
	selFocus  int
	dragging  bool

	// Clipboard fallback when the system clipboard is unavailable.
	clipFallback string

	styles theme.Styles
	keys   EditorKeyMap
}

// SetSpellChecker attaches the background-loaded spellchecker.
func (m *EditorModel) SetSpellChecker(c *spell.Checker) {
	m.spellChecker = c
	m.spellVal = ""
	m.spellBad = nil
}

// SelectedText returns the current selection ("" when inactive or empty).
func (m EditorModel) SelectedText() string {
	a, b, ok := m.selectionRange()
	if !ok {
		return ""
	}
	return string([]rune(m.textarea.Value())[a:b])
}

// selectionRange normalizes the selection to [start, end).
func (m EditorModel) selectionRange() (int, int, bool) {
	if !m.selActive {
		return 0, 0, false
	}
	a, b := m.selAnchor, m.selFocus
	if a > b {
		a, b = b, a
	}
	total := len([]rune(m.textarea.Value()))
	if a < 0 {
		a = 0
	}
	if b > total {
		b = total
	}
	if a >= b {
		return 0, 0, false
	}
	return a, b, true
}

// clearSelection drops any active selection.
func (m *EditorModel) clearSelection() {
	m.selActive = false
	m.dragging = false
}

// NewEditorModel constructs an EditorModel.
func NewEditorModel(styles theme.Styles) EditorModel {
	ta := textarea.New()
	ta.Placeholder = "Begin writing your story here..."
	ta.ShowLineNumbers = true
	ta.CharLimit = 0 // unlimited

	// Custom styles for textarea
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(styles.AppContainer.GetBackground())
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(theme.CurrentTheme.Muted)
	ta.BlurredStyle.LineNumber = lipgloss.NewStyle().Foreground(theme.CurrentTheme.BorderBlurred)

	return EditorModel{
		textarea: ta,
		styles:   styles,
		keys:     DefaultEditorKeyMap(),
	}
}

// Init initializes the editor.
func (m EditorModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the editor component.
func (m EditorModel) Update(msg tea.Msg) (EditorModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case messages.FocusMsg:
		m.Focused = (msg.Target == messages.FocusEditor)
		if m.Focused {
			cmds = append(cmds, m.textarea.Focus())
		} else {
			m.textarea.Blur()
		}

	case messages.SpellReadyMsg:
		m.SetSpellChecker(msg.Checker)
		return m, nil

	case messages.ApplySpellSuggestionMsg:
		return m, m.applySuggestion(msg.Start, msg.End, msg.Replacement)

	case messages.AddCustomWordMsg:
		if m.spellChecker != nil && msg.Word != "" {
			_ = m.spellChecker.AddCustom(msg.Word)
			m.spellVal = "" // force underline recompute
		}
		return m, nil

	case messages.EditorCopyMsg:
		m.copyOp()
		return m, nil

	case messages.EditorCutMsg:
		return m, m.cutOp()

	case messages.EditorPasteMsg:
		return m, m.pasteOp()

	case messages.ChapterSelectedMsg:
		m.ActiveChapter = msg.Chapter
		m.textarea.SetValue(msg.Chapter.Content)
		m.textarea.CursorStart()
		m.clearSelection()
		m.IsDirty = false
		m.Metrics = service.CalculateMetrics(msg.Chapter.Content, false)

		return m, func() tea.Msg {
			return messages.TextChangedMsg{
				ChapterID: m.ActiveChapter.ID,
				Content:   msg.Chapter.Content,
				Metrics:   m.Metrics,
			}
		}

	case messages.SaveCompletedMsg:
		if msg.Success && msg.ChapterID == m.ActiveChapter.ID {
			m.IsDirty = false
			m.Metrics.IsDirty = false
		}

	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseWheelUp:
			for i := 0; i < 3; i++ {
				m.textarea.CursorUp()
			}
			return m, nil

		case tea.MouseWheelDown:
			for i := 0; i < 3; i++ {
				m.textarea.CursorDown()
			}
			return m, nil
		}

		// Left press/drag/release: text selection. Right press: context menu.
		// Both the legacy Type field and the new Button+Action pair are
		// honored so synthetic test messages keep working.
		isLeftPress := msg.Type == tea.MouseLeft ||
			(msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress)
		isMotion := msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionMotion
		isRelease := msg.Type == tea.MouseRelease ||
			(msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease)
		isRightPress := msg.Type == tea.MouseRight ||
			(msg.Button == tea.MouseButtonRight && msg.Action == tea.MouseActionPress)

		switch {
		case isRightPress:
			m.focusEditor(&cmds)
			if off, ok := m.offsetAtCell(msg.X, msg.Y); ok {
				return m, m.openContextMenu(off)
			}
			return m, tea.Batch(cmds...)

		case isLeftPress:
			m.focusEditor(&cmds)
			if off, ok := m.offsetAtCell(msg.X, msg.Y); ok {
				m.setCursorOffset(off)
				m.selAnchor = off
				m.selFocus = off
				m.selActive = false
				m.dragging = true
			}
			return m, tea.Batch(cmds...)

		case isMotion:
			if m.dragging {
				if off, ok := m.offsetAtCell(msg.X, msg.Y); ok {
					m.setCursorOffset(off)
					m.selFocus = off
					m.selActive = off != m.selAnchor
				}
				return m, nil
			}

		case isRelease:
			m.dragging = false
			return m, nil
		}

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Save) {
			content := m.textarea.Value()
			chapterID := m.ActiveChapter.ID
			if chapterID != "" {
				return m, func() tea.Msg {
					return messages.SaveRequestedMsg{
						ChapterID: chapterID,
						Content:   content,
					}
				}
			}
		}

		if !m.Focused {
			return m, nil
		}

		// F2 opens the context menu at the cursor.
		if msg.Type == tea.KeyF2 {
			return m, m.openContextMenu(m.cursorOffset())
		}

		// Alt+C/X/V clipboard ops (work even with an active selection).
		if msg.Alt && len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'c', 'C':
				m.copyOp()
				return m, nil
			case 'x', 'X':
				return m, m.cutOp()
			case 'v', 'V':
				return m, m.pasteOp()
			}
		}

		// Shift+arrows extend the selection (handled here, not forwarded:
		// the textarea has no selection concept).
		switch msg.Type {
		case tea.KeyShiftLeft:
			m.extendSelectionHoriz(-1)
			return m, nil
		case tea.KeyShiftRight:
			m.extendSelectionHoriz(1)
			return m, nil
		case tea.KeyShiftUp:
			m.extendSelectionVert(-1)
			return m, nil
		case tea.KeyShiftDown:
			m.extendSelectionVert(1)
			return m, nil
		}

		// Selection-aware editing.
		if m.selActive {
			switch msg.Type {
			case tea.KeyBackspace, tea.KeyDelete:
				return m, m.deleteSelection()
			case tea.KeyEnter:
				m.deleteSelectionSilent()
			case tea.KeyRunes:
				if len(msg.Runes) == 1 && !msg.Alt && unicode.IsPrint(msg.Runes[0]) {
					m.deleteSelectionSilent()
				}
			case tea.KeyLeft, tea.KeyUp:
				m.collapseSelectionToStart()
				return m, nil
			case tea.KeyRight, tea.KeyDown:
				m.collapseSelectionToEnd()
				return m, nil
			default:
				m.clearSelection()
			}
		}

		prevValue := m.textarea.Value()
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)

		curValue := m.textarea.Value()
		if curValue != prevValue {
			m.IsDirty = true
			m.Metrics = service.CalculateMetrics(curValue, true)
			cmds = append(cmds, func() tea.Msg {
				return messages.TextChangedMsg{
					ChapterID: m.ActiveChapter.ID,
					Content:   curValue,
					Metrics:   m.Metrics,
				}
			})
		}
		return m, tea.Batch(cmds...)
	}

	if m.Focused {
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)
	}

	return m, tea.Batch(cmds...)
}

// SetSize sets allocated dimensions for the editor component.
func (m *EditorModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h

	// Internal textarea size within border
	innerW := w - 4
	innerH := h - 2
	if innerW < 10 {
		innerW = 10
	}
	if innerH < 5 {
		innerH = 5
	}
	m.textarea.SetWidth(innerW)
	m.textarea.SetHeight(innerH)
}

// Value returns the current text in the buffer.
func (m EditorModel) Value() string {
	return m.textarea.Value()
}

// Line returns the current 0-indexed line of the cursor.
func (m EditorModel) Line() int {
	return m.textarea.Line()
}

// GotoLine moves the cursor to the specified 1-indexed line.
func (m *EditorModel) GotoLine(line int) {
	// Move to the very top line
	for m.textarea.Line() > 0 {
		m.textarea.CursorUp()
	}
	m.textarea.SetCursor(0)

	targetLineIdx := line - 1
	if targetLineIdx < 0 {
		targetLineIdx = 0
	}
	lineCount := m.textarea.LineCount()
	if targetLineIdx >= lineCount {
		targetLineIdx = lineCount - 1
	}

	for m.textarea.Line() < targetLineIdx {
		m.textarea.CursorDown()
	}
}

// SetCursorPosition moves the cursor to the specified 1-indexed line and column.
func (m *EditorModel) SetCursorPosition(line, col int) {
	m.GotoLine(line)
	if col > 1 {
		m.textarea.SetCursor(col - 1)
	} else {
		m.textarea.SetCursor(0)
	}
}

// focusEditor focuses the editor, emitting the global focus message once.
func (m *EditorModel) focusEditor(cmds *[]tea.Cmd) {
	if !m.Focused {
		m.Focused = true
		*cmds = append(*cmds, m.textarea.Focus(), func() tea.Msg {
			return messages.FocusMsg{Target: messages.FocusEditor}
		})
	}
}

// extendSelectionHoriz extends the selection one rune left (-1) or right (+1).
func (m *EditorModel) extendSelectionHoriz(dir int) {
	if !m.selActive {
		m.selAnchor = m.cursorOffset()
	}
	off := m.cursorOffset() + dir
	total := len([]rune(m.textarea.Value()))
	if off < 0 {
		off = 0
	}
	if off > total {
		off = total
	}
	m.setCursorOffset(off)
	m.selFocus = off
	m.selActive = off != m.selAnchor
}

// extendSelectionVert extends the selection one visual row up (-1)/down (+1).
func (m *EditorModel) extendSelectionVert(dir int) {
	if !m.selActive {
		m.selAnchor = m.cursorOffset()
	}
	if dir < 0 {
		m.textarea.CursorUp()
	} else {
		m.textarea.CursorDown()
	}
	off := m.cursorOffset()
	m.selFocus = off
	m.selActive = off != m.selAnchor
}

// collapseSelectionToStart/End collapses the selection to one edge.
func (m *EditorModel) collapseSelectionToStart() {
	if a, _, ok := m.selectionRange(); ok {
		m.setCursorOffset(a)
	}
	m.clearSelection()
}

// collapseSelectionToEnd collapses the selection to one edge.
func (m *EditorModel) collapseSelectionToEnd() {
	if _, b, ok := m.selectionRange(); ok {
		m.setCursorOffset(b)
	}
	m.clearSelection()
}

// spliceRunes replaces buffer range [a, b) with insert.
func spliceRunes(value string, a, b int, insert string) string {
	runes := []rune(value)
	if a < 0 {
		a = 0
	}
	if b > len(runes) {
		b = len(runes)
	}
	if a > b {
		a, b = b, a
	}
	out := make([]rune, 0, len(runes)-(b-a)+len([]rune(insert)))
	out = append(out, runes[:a]...)
	out = append(out, []rune(insert)...)
	out = append(out, runes[b:]...)
	return string(out)
}

// commitValue replaces the buffer, repositions the cursor, and emits change
// tracking (dirty flag, metrics, TextChangedMsg).
func (m *EditorModel) commitValue(newVal string, cursorOff int) tea.Cmd {
	m.textarea.SetValue(newVal)
	m.setCursorOffset(cursorOff)
	m.IsDirty = true
	m.Metrics = service.CalculateMetrics(newVal, true)
	chapterID := m.ActiveChapter.ID
	return func() tea.Msg {
		return messages.TextChangedMsg{
			ChapterID: chapterID,
			Content:   newVal,
			Metrics:   m.Metrics,
		}
	}
}

// deleteSelection removes the selection and returns the change command.
func (m *EditorModel) deleteSelection() tea.Cmd {
	a, b, ok := m.selectionRange()
	if !ok {
		return nil
	}
	m.clearSelection()
	return m.commitValue(spliceRunes(m.textarea.Value(), a, b, ""), a)
}

// deleteSelectionSilent removes the selection without emitting (used when the
// triggering keystroke will forward and emit itself, e.g. type-over).
func (m *EditorModel) deleteSelectionSilent() {
	a, b, ok := m.selectionRange()
	if !ok {
		return
	}
	m.clearSelection()
	m.textarea.SetValue(spliceRunes(m.textarea.Value(), a, b, ""))
	m.setCursorOffset(a)
	m.IsDirty = true
}

// applySuggestion replaces [start, end) with the chosen correction.
func (m *EditorModel) applySuggestion(start, end int, replacement string) tea.Cmd {
	m.clearSelection()
	return m.commitValue(spliceRunes(m.textarea.Value(), start, end, replacement), start+len([]rune(replacement)))
}

// setClipboard writes text to the system clipboard (internal fallback).
func (m *EditorModel) setClipboard(s string) {
	if err := clipboard.WriteAll(s); err != nil {
		m.clipFallback = s
	}
}

// getClipboard reads the system clipboard (internal fallback).
func (m *EditorModel) getClipboard() string {
	if s, err := clipboard.ReadAll(); err == nil {
		return s
	}
	return m.clipFallback
}

// copyOp copies the selection to the clipboard (no-op without selection;
// the context menu pre-selects its word so menu copy works).
func (m *EditorModel) copyOp() {
	if text := m.SelectedText(); text != "" {
		m.setClipboard(text)
	}
}

// cutOp cuts the selection to the clipboard.
func (m *EditorModel) cutOp() tea.Cmd {
	if text := m.SelectedText(); text != "" {
		m.setClipboard(text)
		return m.deleteSelection()
	}
	return nil
}

// pasteOp inserts the clipboard at the cursor, replacing the selection.
func (m *EditorModel) pasteOp() tea.Cmd {
	text := m.getClipboard()
	if text == "" {
		return nil
	}
	if a, b, ok := m.selectionRange(); ok {
		m.clearSelection()
		return m.commitValue(spliceRunes(m.textarea.Value(), a, b, text), a+len([]rune(text)))
	}
	off := m.cursorOffset()
	return m.commitValue(spliceRunes(m.textarea.Value(), off, off, text), off+len([]rune(text)))
}

// openContextMenu selects the word at off (when misspelled), opens the menu,
// and kicks async suggestion computation.
func (m *EditorModel) openContextMenu(off int) tea.Cmd {
	word, start, end := "", off, off
	if m.spellChecker != nil {
		if tok, ok := m.wordAtOffset(off); ok && !m.spellChecker.Check(tok.Word) {
			word, start, end = tok.Word, tok.StartRune, tok.EndRune
			m.selAnchor, m.selFocus = start, end
			m.selActive = true
		}
	}
	checker := m.spellChecker
	openCmd := func() tea.Msg {
		return messages.OpenSpellMenuMsg{Word: word, Start: start, End: end}
	}
	if checker == nil || word == "" {
		return openCmd
	}
	suggestCmd := func() tea.Msg {
		return messages.SpellSuggestMsg{
			Word:        word,
			Start:       start,
			End:         end,
			Suggestions: checker.Suggestions(word, MaxSpellSuggestions),
		}
	}
	return tea.Batch(openCmd, suggestCmd)
}

// decorateContent overlays selection (reverse video) and misspelled words
// (red underline) onto the rendered textarea view.
func (m *EditorModel) decorateContent(view string) string {
	rows := m.mapViewRows(view)
	if len(rows) == 0 {
		return view
	}
	lines := m.bufferLines()
	visible := make(map[int]bool)
	for _, r := range rows {
		if r.bufferLine >= 0 && r.bufferLine < len(lines) {
			visible[r.bufferLine] = true
		}
	}
	// Misspelled words across visible lines (cached per buffer value).
	val := m.textarea.Value()
	if m.spellChecker != nil && (m.spellVal != val || m.spellBad == nil) {
		bad := make(map[string]struct{})
		for ln := range visible {
			for _, t := range spell.Tokenize(lines[ln]) {
				if !m.spellChecker.Check(t.Word) {
					bad[spell.Normalize(t.Word)] = struct{}{}
				}
			}
		}
		m.spellBad = bad
		m.spellVal = val
	}
	selA, selB, hasSel := m.selectionRange()
	spellOpen, spellClose := spellStyleCodes()
	selOpen, selClose := selectionCodes()

	rawLines := strings.Split(view, "\n")
	var out []string
	for i, raw := range rawLines {
		if i >= len(rows) {
			out = append(out, raw)
			continue
		}
		row := rows[i]
		if row.bufferLine < 0 {
			out = append(out, raw)
			continue
		}
		prefix := row.prefix
		var spans []overlaySpan
		// Spell spans from row-local tokens (plain indexes include prefix).
		if len(m.spellBad) > 0 {
			for _, t := range spell.Tokenize(string(row.content)) {
				if _, ok := m.spellBad[spell.Normalize(t.Word)]; ok {
					spans = append(spans, overlaySpan{
						start: prefix + t.LineOffset,
						end:   prefix + t.LineOffset + len([]rune(t.Word)),
						open:  spellOpen, close: spellClose,
					})
				}
			}
		}
		// Selection span from buffer offsets.
		if hasSel {
			rowStart := m.lineStartOffset(row.bufferLine) + row.colStart
			rowEnd := rowStart + len(row.content)
			if selB > rowStart && selA < rowEnd {
				a := selA
				if a < rowStart {
					a = rowStart
				}
				b := selB
				if b > rowEnd {
					b = rowEnd
				}
				selStart, selEnd := prefix+(a-rowStart), prefix+(b-rowStart)
				// Suppress spell spans overlapped by the selection.
				kept := spans[:0]
				for _, sp := range spans {
					if sp.end <= selStart || sp.start >= selEnd {
						kept = append(kept, sp)
					}
				}
				spans = kept
				spans = append(spans, overlaySpan{start: selStart, end: selEnd, open: selOpen, close: selClose})
				sortSpans(spans)
			}
		}
		out = append(out, decorateRow(raw, spans))
	}
	return strings.Join(out, "\n")
}

// sortSpans orders spans by start offset (decorateRow requires order).
func sortSpans(spans []overlaySpan) {
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
}

// View renders the editor panel with selection and spellcheck overlays.
func (m EditorModel) View() string {
	panelStyle := m.styles.BlurredPanel
	if m.Focused {
		panelStyle = m.styles.FocusedPanel
	}

	content := m.textarea.View()
	if m.textarea.Value() != "" && (m.selActive || (m.spellChecker != nil)) {
		content = m.decorateContent(content)
	}

	return panelStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}
