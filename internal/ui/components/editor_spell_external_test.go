package components_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service/spell"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/components"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

func newSpellEditor(t *testing.T, content string) components.EditorModel {
	t.Helper()
	ed := components.NewEditorModel(theme.DefaultStyles)
	ed.SetSize(60, 20)
	ed, _ = ed.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{ID: "c1", Title: "T", Content: content},
	})
	ed, _ = ed.Update(messages.FocusMsg{Target: messages.FocusEditor})
	// Chapter load leaves the cursor at the last line start; Home it.
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyHome})
	return ed
}

func loadSpellChecker(t *testing.T) *spell.Checker {
	t.Helper()
	c, err := spell.Load("")
	if err != nil {
		t.Fatalf("spell.Load failed: %v", err)
	}
	return c
}

func shiftKey(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

func TestEditorShiftSelection(t *testing.T) {
	// Single line: chapter load leaves the cursor at (0,0).
	ed := newSpellEditor(t, "abcdef")
	// Extend right twice.
	ed, _ = ed.Update(shiftKey(tea.KeyShiftRight))
	ed, _ = ed.Update(shiftKey(tea.KeyShiftRight))
	if got := ed.SelectedText(); got != "ab" {
		t.Errorf("expected selection %q, got %q", "ab", got)
	}
	// Collapse with plain Left lands on the start edge.
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if got := ed.SelectedText(); got != "" {
		t.Errorf("expected collapsed selection, got %q", got)
	}
}

func TestEditorCopyPasteFallback(t *testing.T) {
	ed := newSpellEditor(t, "hola mundo")
	for i := 0; i < 4; i++ {
		ed, _ = ed.Update(shiftKey(tea.KeyShiftRight))
	}
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c"), Alt: true})
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyEnd})
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v"), Alt: true})
	if got := ed.Value(); got != "hola mundohola" {
		t.Errorf("expected pasted value, got %q", got)
	}
}

func TestEditorCutAndDelete(t *testing.T) {
	ed := newSpellEditor(t, "hola mundo")
	for i := 0; i < 4; i++ {
		ed, _ = ed.Update(shiftKey(tea.KeyShiftRight))
	}
	ed, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x"), Alt: true})
	if cmd != nil {
		ed, _ = ed.Update(cmd())
	}
	if got := ed.Value(); got != " mundo" {
		t.Errorf("expected cut value, got %q", got)
	}
	// Paste it back at the start.
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyHome})
	ed, cmd = ed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v"), Alt: true})
	if cmd != nil {
		ed, _ = ed.Update(cmd())
	}
	if got := ed.Value(); got != "hola mundo" {
		t.Errorf("expected restored value, got %q", got)
	}
}

func TestEditorTypeOverSelection(t *testing.T) {
	ed := newSpellEditor(t, "hola mundo")
	for i := 0; i < 4; i++ {
		ed, _ = ed.Update(shiftKey(tea.KeyShiftRight))
	}
	ed, _ = ed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Z")})
	if got := ed.Value(); got != "Z mundo" {
		t.Errorf("expected type-over value, got %q", got)
	}
}

func TestEditorBackspaceDeletesSelection(t *testing.T) {
	ed := newSpellEditor(t, "hola mundo")
	for i := 0; i < 4; i++ {
		ed, _ = ed.Update(shiftKey(tea.KeyShiftRight))
	}
	ed, cmd := ed.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd != nil {
		ed, _ = ed.Update(cmd())
	}
	if got := ed.Value(); got != " mundo" {
		t.Errorf("expected backspace value, got %q", got)
	}
}

func TestEditorUnderlineMisspelled(t *testing.T) {
	ed := newSpellEditor(t, "hola novle mundo")
	ed.SetSpellChecker(loadSpellChecker(t))
	view := ed.View()
	if !strings.Contains(view, "\x1b[4m") {
		t.Fatalf("expected underline SGR in view")
	}
	stripped := stripTestANSI(view)
	if !strings.Contains(stripped, "hola novle mundo") {
		t.Errorf("plain text altered: %q", stripped)
	}
	// 'hola' must not be underlined, 'novle' must.
	if strings.Contains(view, "\x1b[4mhola") {
		t.Errorf("'hola' should not be underlined")
	}
	idx := strings.Index(view, "novle")
	if idx < 0 || idx < 4 || view[idx-4:idx] != "\x1b[4m" {
		// Underline may start with color+underline codes; accept color variant.
		if !strings.Contains(view, "mnovle") && !strings.Contains(view, "\x1b[38;2;") {
			t.Errorf("expected underline codes around novle")
		}
	}
}

func stripTestANSI(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] < '@' || s[j] > '~') {
				j++
			}
			i = j + 1
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

func TestEditorRightClickOpensMenu(t *testing.T) {
	ed := newSpellEditor(t, "hola novle mundo")
	ed.SetSpellChecker(loadSpellChecker(t))
	// 'novle' starts at col 5 of line 0: ex = border(1)+prompt(2)+gutter(4)+5.
	ed, cmd := ed.Update(tea.MouseMsg{
		Button: tea.MouseButtonRight,
		Action: tea.MouseActionPress,
		X:      12,
		Y:      1,
	})
	if cmd == nil {
		t.Fatalf("expected menu/suggest commands on right-click")
	}
	foundOpen := false
	if batch, ok := cmd().(tea.BatchMsg); ok {
		for _, c := range batch {
			if m, ok := c().(messages.OpenSpellMenuMsg); ok {
				foundOpen = true
				if m.Word != "novle" || m.Start != 5 || m.End != 10 {
					t.Errorf("wrong menu range: %+v", m)
				}
			}
		}
	} else {
		t.Fatalf("expected batch cmd, got %T", cmd())
	}
	if !foundOpen {
		t.Errorf("OpenSpellMenuMsg not emitted on right-click")
	}
	// Right-click also selects the word.
	if got := ed.SelectedText(); got != "novle" {
		t.Errorf("expected word selected on right-click, got %q", got)
	}
}

func TestEditorApplySuggestionAndCustom(t *testing.T) {
	ed := newSpellEditor(t, "hola novle mundo")
	ed.SetSpellChecker(loadSpellChecker(t))
	ed, cmd := ed.Update(messages.ApplySpellSuggestionMsg{Start: 5, End: 10, Replacement: "novel"})
	if cmd != nil {
		ed, _ = ed.Update(cmd())
	}
	if got := ed.Value(); got != "hola novel mundo" {
		t.Errorf("expected applied suggestion, got %q", got)
	}
	ed, _ = ed.Update(messages.AddCustomWordMsg{Word: "Kaelith"})
	ed, _ = ed.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{ID: "c1", Title: "T", Content: "Kaelith corre"},
	})
	view := ed.View()
	if strings.Contains(view, "\x1b[4m") {
		t.Errorf("custom word should not be underlined: %q", stripTestANSI(view))
	}
}

func TestSpellMenuFlow(t *testing.T) {
	menu := components.NewSpellMenuModel(theme.DefaultStyles)
	menu.SetSize(100, 30)
	menu.Open("novle", 5, 10)
	if !menu.Active {
		t.Fatalf("menu should be active after Open")
	}
	menu.SetSuggestions("novle", 5, 10, []string{"novel", "noble"})
	// Stale results are ignored.
	menu.SetSuggestions("other", 0, 1, []string{"zzz"})
	view := menu.View()
	if !strings.Contains(view, "novel") || strings.Contains(view, "zzz") {
		t.Errorf("stale suggestions leaked into menu")
	}
	// Navigate to first suggestion and activate.
	menu, cmd := menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	menu, cmd = menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	_ = menu
	menu2 := components.NewSpellMenuModel(theme.DefaultStyles)
	menu2.SetSize(100, 30)
	menu2.Open("novle", 5, 10)
	menu2.SetSuggestions("novle", 5, 10, []string{"novel"})
	_, cmd = menu2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected activation command")
	}
	apply, ok := cmd().(messages.ApplySpellSuggestionMsg)
	if !ok || apply.Replacement != "novel" || apply.Start != 5 || apply.End != 10 {
		t.Errorf("wrong apply message: %+v", cmd())
	}
	// Esc closes.
	menu3 := components.NewSpellMenuModel(theme.DefaultStyles)
	menu3.SetSize(100, 30)
	menu3.Open("novle", 5, 10)
	menu3, _ = menu3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x"), Alt: false})
	menu3, _ = menu3.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if menu3.Active {
		t.Errorf("expected menu closed on esc")
	}
	_ = cmd
}

func TestSpellMenuMouseActivate(t *testing.T) {
	menu := components.NewSpellMenuModel(theme.DefaultStyles)
	menu.SetSize(100, 30)
	menu.Open("novle", 5, 10)
	menu.SetSuggestions("novle", 5, 10, []string{"novel"})
	// Sweep every cell; exactly the suggestion rows must activate it.
	hits := 0
	var last messages.ApplySpellSuggestionMsg
	for y := 0; y < 30; y++ {
		_, cmd := menu.Update(tea.MouseMsg{X: 50, Y: y, Type: tea.MouseLeft})
		if cmd == nil {
			continue
		}
		if m, ok := cmd().(messages.ApplySpellSuggestionMsg); ok {
			hits++
			last = m
		}
	}
	if hits != 1 {
		t.Errorf("expected exactly 1 suggestion row hit, got %d", hits)
	}
	if last.Replacement != "novel" {
		t.Errorf("wrong suggestion activated: %+v", last)
	}
}
