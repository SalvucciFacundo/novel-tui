package components

import (
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/messages"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// newSpellTestEditor builds an editor with known content for mapping tests.
func newSpellTestEditor(content string) EditorModel {
	ed := NewEditorModel(theme.DefaultStyles)
	ed.SetSize(40, 20)
	ed, _ = ed.Update(messages.ChapterSelectedMsg{
		Chapter: domain.Chapter{ID: "c1", Title: "T", Content: content},
	})
	ed, _ = ed.Update(messages.FocusMsg{Target: messages.FocusEditor})
	return ed
}

func TestMapViewRowsBasic(t *testing.T) {
	ed := newSpellTestEditor("hola mundo\nsegunda línea")
	rows := ed.mapViewRows(ed.textarea.View())
	if len(rows) < 2 {
		t.Fatalf("expected >=2 mapped rows, got %d", len(rows))
	}
	if rows[0].bufferLine != 0 || rows[0].colStart != 0 {
		t.Errorf("row 0 should map to line 0 col 0, got %+v", rows[0])
	}
	if string(rows[0].content) != "hola mundo" {
		t.Errorf("row 0 content = %q", string(rows[0].content))
	}
	if rows[1].bufferLine != 1 || rows[1].colStart != 0 {
		t.Errorf("row 1 should map to line 1 col 0, got %+v", rows[1])
	}
}

func TestMapViewRowsWrapped(t *testing.T) {
	long := "esta es una línea deliberadamente larga que debe envolverse en varias filas visibles"
	ed := newSpellTestEditor(long + "\ncorta")
	rows := ed.mapViewRows(ed.textarea.View())
	if len(rows) < 3 {
		t.Fatalf("expected wrapped rows, got %d: %q", len(rows), ed.textarea.View())
	}
	// Sequential-partition invariant: every row's content must equal the
	// exact buffer slice [colStart, colStart+len), in increasing order.
	lineRunes := []rune(long)
	prevEnd := -1
	seen := 0
	for _, r := range rows {
		if r.bufferLine != 0 {
			break
		}
		seen++
		end := r.colStart + len(r.content)
		if r.colStart < prevEnd {
			t.Fatalf("row %+v overlaps previous end %d", r, prevEnd)
		}
		if end > len(lineRunes) || string(lineRunes[r.colStart:end]) != string(r.content) {
			t.Fatalf("row content %+v is not the buffer slice", r)
		}
		prevEnd = end
	}
	if seen < 2 {
		t.Errorf("expected multiple wrapped rows for line 0, got %d", seen)
	}
}

func TestOffsetAndWordMapping(t *testing.T) {
	ed := newSpellTestEditor("hola novle mundo")
	ex := 1 + ed.promptCells() + ed.gutterCells() + 5 // col of 'n' in novle
	off, ok := ed.offsetAtCell(ex, 1)
	if !ok {
		t.Fatalf("offsetAtCell failed")
	}
	tok, ok := ed.wordAtOffset(off)
	if !ok || tok.Word != "novle" {
		t.Errorf("expected word novle at click, got %+v (off=%d)", tok, off)
	}
	if tok.StartRune != 5 || tok.EndRune != 10 {
		t.Errorf("expected range [5,10), got [%d,%d)", tok.StartRune, tok.EndRune)
	}
}

func TestCursorOffsetRoundTrip(t *testing.T) {
	ed := newSpellTestEditor("corazón\nhola")
	// NOTE: chapter load leaves the cursor wherever SetValue put it, so
	// place it explicitly (offsets are rune-based, accents included).
	ed.setCursorOffset(0)
	if off := ed.cursorOffset(); off != 0 {
		t.Fatalf("expected cursor 0, got %d", off)
	}
	ed.setCursorOffset(9) // 'h' of hola (7 runes + \n + 1)
	if off := ed.cursorOffset(); off != 9 {
		t.Errorf("expected cursor 9 after set, got %d", off)
	}
	line, col := ed.offsetToLineCol(9)
	if line != 1 || col != 1 {
		t.Errorf("expected (1,1) for offset 9, got (%d,%d)", line, col)
	}
}

func TestDecorateRowKeepsANSI(t *testing.T) {
	raw := "\x1b[38;5;1m" + "hola novle" + "\x1b[0m"
	spans := []overlaySpan{{start: 5, end: 10, open: "\x1b[4m", close: "\x1b[24m"}}
	got := decorateRow(raw, spans)
	plain := stripANSI(got)
	if plain != "hola novle" {
		t.Errorf("plain text changed: %q", plain)
	}
	if !strings.Contains(got, "\x1b[4mnovle\x1b[24m") {
		t.Errorf("expected wrapped word, got %q", got)
	}
}
