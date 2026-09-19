package components

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/rivo/uniseg"

	"github.com/SalvucciFacundo/novel-tui/internal/service/spell"
	"github.com/SalvucciFacundo/novel-tui/internal/ui/theme"
)

// This file implements the coordinate mapping between rendered textarea view
// rows and buffer rune offsets, plus the underline/selection overlay.
//
// Mapping strategy: every rendered content row carries its buffer line number
// in the gutter (bubbles textarea prints " %*v "), so parsing the rendered
// view yields authoritative (bufferLine, colStart) pairs without replicating
// the soft-wrap algorithm. Continuation rows inherit the previous line and
// accumulate rune counts. End-of-buffer filler rows are harmless: clicks
// there clamp to the last line end.

var (
	ansiSequence = regexp.MustCompile("\x1b\\[[0-9;]*m")
	// gutterNumber matches the " %*v " line-number gutter. Width is
	// validated in code: normal gutters are exactly digits+2 cells, wider
	// ones only for line numbers exceeding the digit width. This rejects
	// continuation rows whose content happens to start with digits.
	gutterNumber = regexp.MustCompile(`^ ( *)(\d+) `)
)

// mappedRow describes one rendered view row of the textarea.
type mappedRow struct {
	bufferLine int // 0-indexed buffer line, -1 for filler/unknown
	colStart   int // buffer rune offset where this row's content starts (within line)
	content    []rune
	prefix     int // plain-rune count of prompt+gutter before content
	gutter     int // gutter cells (prompt excluded), varies with line numbers
}

// promptCells returns how many leading cells the textarea prompt occupies.
func (m EditorModel) promptCells() int {
	return uniseg.StringWidth(m.textarea.Prompt)
}

// gutterCells returns the gutter width for a line number with the given
// digit count. Format is " %*v " (leading space, padded number, space).
func (m EditorModel) gutterCells() int {
	digits := len(strconv.Itoa(m.textarea.MaxHeight))
	if digits < 1 {
		digits = 1
	}
	return digits + 2
}

// stripANSI removes SGR escape sequences, keeping visible runes.
func stripANSI(s string) string {
	return ansiSequence.ReplaceAllString(s, "")
}

// mapViewRows parses the textarea view into per-row buffer mappings.
// promptCells and gutter are skipped; trailing padding is trimmed.
//
// Column accounting is sequential, not arithmetic: every rendered row holds
// a contiguous slice of its buffer line (the soft-wrap never reorders), so
// a continuation row's colStart is located by searching its content in the
// buffer line right after the previous row's end. This stays exact no
// matter how the wrapper shuffles separator spaces across row edges.
func (m EditorModel) mapViewRows(view string) []mappedRow {
	promptW := m.promptCells()
	gutterW := m.gutterCells()
	bufLines := m.bufferLines()
	var rows []mappedRow
	curLine := -1
	curCol := 0
	rowGutter := 0
	for _, rawLine := range strings.Split(view, "\n") {
		rowGutter = 0
		numbered := false
		plain := stripANSI(rawLine)
		runes := []rune(plain)
		if len(runes) <= promptW {
			continue
		}
		rest := string(runes[promptW:])
		content := rest
		prefix := promptW
		if match := gutterNumber.FindStringSubmatch(rest); match != nil {
			leadSpaces := 1 + len(match[1])
			num, err := strconv.Atoi(match[2])
			if err != nil || num < 1 {
				continue
			}
			glen := leadSpaces + 1 + len(match[2])
			// Accept only exact gutter widths: digits+2 normally, or wider
			// solely for over-wide line numbers. Anything else is content
			// starting with digits, i.e. a continuation row.
			isGutter := glen == gutterW ||
				(len(match[2]) > gutterW-2 && glen == len(match[2])+2)
			if isGutter {
				curLine = num - 1
				curCol = 0
				numbered = true
				prefix += glen
			rowGutter = glen
				if len([]rune(rest)) > glen {
					content = string([]rune(rest)[glen:])
				} else {
					content = ""
				}
			} else if curLine < 0 {
				continue
			} else {
				prefix += gutterW
				if len([]rune(rest)) > gutterW {
					content = string([]rune(rest)[gutterW:])
				} else {
					content = ""
				}
			}
		} else if strings.HasPrefix(rest, "~") {
			// End-of-buffer filler row: not addressable content.
			rows = append(rows, mappedRow{bufferLine: -1})
			continue
		} else {
			// Continuation of the previous buffer line (blank gutter).
			if curLine < 0 {
				continue
			}
			prefix += gutterW
			rowGutter = gutterW
			if len([]rune(rest)) > gutterW {
				content = string([]rune(rest)[gutterW:])
			} else {
				content = ""
			}
		}
		contentRunes := []rune(strings.TrimRight(content, " "))
		if numbered {
			curCol = 0
		} else if len(contentRunes) > 0 && curLine >= 0 && curLine < len(bufLines) {
			curCol = locateContinuationCol(bufLines[curLine], contentRunes, curCol)
		}
		rows = append(rows, mappedRow{bufferLine: curLine, colStart: curCol, content: contentRunes, prefix: prefix, gutter: rowGutter})
		curCol += len(contentRunes)
	}
	return rows
}

// locateContinuationCol finds where a wrapped continuation row starts inside
// its buffer line, searching at or after prevEnd (the previous row's end).
// Rows starting with spaces continue exactly at prevEnd; otherwise the first
// content runes are located forward (they always follow, since rows hold
// contiguous slices in order).
func locateContinuationCol(line string, content []rune, prevEnd int) int {
	lead := 0
	for lead < len(content) && content[lead] == ' ' {
		lead++
	}
	if lead == len(content) {
		return prevEnd
	}
	lineRunes := []rune(line)
	if prevEnd < 0 {
		prevEnd = 0
	}
	if prevEnd > len(lineRunes) {
		prevEnd = len(lineRunes)
	}
	if lead > 0 {
		return prevEnd
	}
	needle := content
	if len(needle) > 12 {
		needle = needle[:12]
	}
	idx := indexRunes(lineRunes[prevEnd:], needle)
	if idx < 0 {
		return prevEnd
	}
	return prevEnd + idx
}

// indexRunes reports the first index of needle in hay (rune-wise), or -1.
func indexRunes(hay, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
outer:
	for i := 0; i+len(needle) <= len(hay); i++ {
		for j := range needle {
			if hay[i+j] != needle[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}

// bufferLines splits the current value into lines (same split the textarea uses).
func (m EditorModel) bufferLines() []string {
	return strings.Split(m.textarea.Value(), "\n")
}

// lineStartOffset returns the buffer rune offset where line starts.
func (m EditorModel) lineStartOffset(line int) int {
	off := 0
	lines := m.bufferLines()
	for i := 0; i < line && i < len(lines); i++ {
		off += len([]rune(lines[i])) + 1 // +1 for '\n'
	}
	return off
}

// cursorOffset returns the cursor position as a buffer rune offset.
func (m EditorModel) cursorOffset() int {
	row := m.textarea.Line()
	lines := m.bufferLines()
	if row < 0 {
		row = 0
	}
	if row >= len(lines) {
		row = len(lines) - 1
	}
	off := m.lineStartOffset(row)
	col := m.textarea.LineInfo().CharOffset
	lineLen := len([]rune(lines[row]))
	if col < 0 {
		col = 0
	}
	if col > lineLen {
		col = lineLen
	}
	return off + col
}

// offsetToLineCol converts a buffer rune offset to (line, col-in-runes).
func (m EditorModel) offsetToLineCol(off int) (line, col int) {
	if off < 0 {
		off = 0
	}
	lines := m.bufferLines()
	for i, text := range lines {
		l := len([]rune(text))
		if off <= l {
			return i, off
		}
		off -= l + 1
	}
	if len(lines) == 0 {
		return 0, 0
	}
	last := len(lines) - 1
	return last, len([]rune(lines[last]))
}

// setCursorOffset moves the cursor to a buffer rune offset.
func (m *EditorModel) setCursorOffset(off int) {
	line, col := m.offsetToLineCol(off)
	m.GotoLine(line + 1) // GotoLine is 1-indexed
	m.textarea.SetCursor(col)
}

// offsetAtCell maps editor-local (ex, ey) view cells to a buffer rune offset.
// ex/ey include the panel border; content starts at (1,1). Out-of-range
// clicks clamp to the nearest content.
func (m EditorModel) offsetAtCell(ex, ey int) (int, bool) {
	rows := m.mapViewRows(m.textarea.View())
	y := ey - 1 // panel top border
	if y < 0 {
		y = 0
	}
	if y >= len(rows) {
		y = len(rows) - 1
	}
	if len(rows) == 0 {
		return 0, false
	}
	row := rows[y]
	if row.bufferLine < 0 {
		// Filler below content: end of buffer.
		return len([]rune(m.textarea.Value())), true
	}
	lines := m.bufferLines()
	if row.bufferLine >= len(lines) {
		return len([]rune(m.textarea.Value())), true
	}
	lineRunes := []rune(lines[row.bufferLine])
	// ex counts cells from the panel edge: subtract border, prompt, gutter.
	targetCell := ex - 1 - m.promptCells() - row.gutter
	if targetCell < 0 {
		targetCell = 0
	}
	cells := 0
	col := 0
	for col < len(row.content) && row.colStart+col < len(lineRunes) && cells < targetCell {
		cells += uniseg.StringWidth(string(lineRunes[row.colStart+col]))
		col++
	}
	lineOff := row.colStart + col
	if lineOff > len([]rune(lines[row.bufferLine])) {
		lineOff = len([]rune(lines[row.bufferLine]))
	}
	return m.lineStartOffset(row.bufferLine) + lineOff, true
}

// wordAtOffset returns the token containing buffer rune offset off.
func (m EditorModel) wordAtOffset(off int) (spell.Token, bool) {
	line, col := m.offsetToLineCol(off)
	lines := m.bufferLines()
	if line >= len(lines) {
		return spell.Token{}, false
	}
	for _, t := range spell.Tokenize(lines[line]) {
		if col >= t.LineOffset && col < t.LineOffset+len([]rune(t.Word)) {
			lineStart := m.lineStartOffset(line)
			t.StartRune = lineStart + t.LineOffset
			t.EndRune = t.StartRune + len([]rune(t.Word))
			t.Line = line
			return t, true
		}
	}
	return spell.Token{}, false
}

// --- overlay rendering ---

// spellStyleCodes returns open/close SGR for red underline from the theme.
// The close sequence only resets underline+foreground (never a full reset),
// so ambient row styles (cursor line background, cursor) survive.
func spellStyleCodes() (open, close string) {
	r, g, b := hexToRGB(string(theme.CurrentTheme.Error))
	return fmt.Sprintf("\x1b[4m\x1b[38;2;%d;%d;%dm", r, g, b), "\x1b[24;39m"
}

// hexToRGB parses "#rrggbb" into components (falls back to plain red 31).
func hexToRGB(h string) (r, g, b int) {
	var ri, gi, bi uint8
	if len(h) == 7 && h[0] == '#' {
		_, err := fmt.Sscanf(h[1:], "%02x%02x%02x", &ri, &gi, &bi)
		if err == nil {
			return int(ri), int(gi), int(bi)
		}
	}
	return 243, 139, 168 // CatppuccinMocha Error
}

// selectionCodes returns open/close SGR for reverse-video selection.
func selectionCodes() (open, close string) {
	return "\x1b[7m", "\x1b[27m"
}

// overlaySpan is a styled rune span within one rendered row (plain indexes).
type overlaySpan struct {
	start, end  int // plain-rune offsets in the row's visible text
	open, close string
}

// decorateRow injects overlay spans into a rendered row, walking plain-rune
// indexes so existing ANSI sequences are never split.
func decorateRow(raw string, spans []overlaySpan) string {
	if len(spans) == 0 {
		return raw
	}
	var out strings.Builder
	plainIdx := 0
	spanIdx := 0
	i := 0
	openSpans := map[int]overlaySpan{}
	// syncSpans closes spans ending at plainIdx and opens spans starting
	// there. It runs at every position — rune or ANSI — so boundaries
	// inside styled regions (e.g. around the cursor) stay exact.
	syncSpans := func() {
		for id, sp := range openSpans {
			if sp.end == plainIdx {
				out.WriteString(sp.close)
				delete(openSpans, id)
			}
		}
		for spanIdx < len(spans) && spans[spanIdx].start == plainIdx {
			out.WriteString(spans[spanIdx].open)
			openSpans[spanIdx] = spans[spanIdx]
			spanIdx++
		}
	}
	for i < len(raw) {
		syncSpans()
		if loc := ansiSequence.FindStringIndex(raw[i:]); loc != nil && loc[0] == 0 {
			out.WriteString(raw[i : i+loc[1]])
			i += loc[1]
			continue
		}
		_, size := utf8.DecodeRuneInString(raw[i:])
		if size == 0 {
			break
		}
		out.WriteString(raw[i : i+size])
		i += size
		plainIdx++
	}
	syncSpans()
	for _, sp := range openSpans {
		out.WriteString(sp.close)
	}
	return out.String()
}
