// Package spell provides a pure-Go (CGO-free) spellchecker for the editor.
//
// Dictionaries for Spanish and English are embedded as gzipped word lists
// (see data/SOURCES.md) and loaded into a sorted slice, so Check is a binary
// search. Suggestions are computed on demand with a banded Levenshtein
// distance capped at 2 over same-length word buckets.
package spell

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"
)

//go:embed data/en.txt.gz
var enData []byte

//go:embed data/es.txt.gz
var esData []byte

// MaxSuggestionDistance caps the edit distance for candidate suggestions.
const MaxSuggestionDistance = 2

// Token is a checkable word found in a text with rune offsets.
type Token struct {
	Word       string
	StartRune  int
	EndRune    int // exclusive
	Line       int // 0-indexed buffer line
	LineOffset int // rune offset of the token start within its line
}

// Checker holds the combined dictionaries and user custom words.
type Checker struct {
	mu     sync.RWMutex
	words  []string // sorted, deduped; Check is a binary search
	custom map[string]struct{}

	customPath string
}

// Load builds a Checker with the embedded Spanish and English dictionaries.
// Custom words are read from customPath when it exists (created on first
// AddCustom). Pass "" to disable custom-word persistence.
func Load(customPath string) (*Checker, error) {
	seen := make(map[string]struct{})
	var words []string
	for _, blob := range [][]byte{enData, esData} {
		list, err := readGzLines(blob)
		if err != nil {
			return nil, err
		}
		for _, w := range list {
			w = Normalize(w)
			if w == "" {
				continue
			}
			if _, ok := seen[w]; ok {
				continue
			}
			seen[w] = struct{}{}
			words = append(words, w)
		}
	}
	sort.Strings(words)

	c := &Checker{
		words:      words,
		custom:     make(map[string]struct{}),
		customPath: customPath,
	}
	if customPath != "" {
		_ = c.loadCustom()
	}
	return c, nil
}

// WordCount reports the number of dictionary entries (both languages).
func (c *Checker) WordCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.words)
}

// editAlphabet holds candidate runes for suggestion generation (both
// languages, including accented vowels).
var editAlphabet = []rune("abcdefghijklmnopqrstuvwxyzáéíóúüñç")

// Normalize lowercases a word for dictionary lookup.
func Normalize(word string) string {
	return strings.ToLower(strings.TrimSpace(word))
}

// DefaultCustomPath returns ~/.config/novel-tui/custom_words.txt for
// user-approved words.
func DefaultCustomPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "novel-tui", "custom_words.txt")
}

// Check reports whether word is correctly spelled (either language, custom
// words, digits, or apostrophe/hyphen compositions of valid parts).
func (c *Checker) Check(word string) bool {
	w := Normalize(word)
	if w == "" {
		return true
	}
	// Tokens with digits (chapter numbers, years, model names) always pass.
	for _, r := range w {
		if unicode.IsDigit(r) {
			return true
		}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.lookupLocked(w) {
		return true
	}
	// Contractions: check the stem before the apostrophe (don't->don,
	// can't->can, it's->it).
	if i := strings.IndexAny(w, "'’"); i > 0 {
		if c.lookupLocked(w[:i]) {
			return true
		}
	}
	// Hyphenated compounds pass when every part is valid.
	if strings.Contains(w, "-") {
		parts := strings.Split(w, "-")
		allValid := len(parts) > 1
		for _, p := range parts {
			if p == "" || !c.lookupLocked(p) {
				allValid = false
				break
			}
		}
		if allValid {
			return true
		}
	}
	return false
}

func (c *Checker) lookupLocked(w string) bool {
	if _, ok := c.custom[w]; ok {
		return true
	}
	i := sort.SearchStrings(c.words, w)
	return i < len(c.words) && c.words[i] == w
}
// Suggestions returns up to max ordered correction candidates for word.
//
// Generate-and-test in three cheap rounds instead of O(dictionary) distance
// scans (which collapse under -race):
//  1. every distance-1 edit of the query (full alphabet),
//  2. distance-1 edits of each accent variant (covers accent+letter combos,
//     the dominant Spanish error class),
//  3. deletes and transposes of round-1 strings (cheap second-order edits).
//
// Ranking: accent-folded distance first (so "corason" prefers "corazón"
// over accentless lookalikes), then exact distance, shared prefix,
// alphabetically.
func (c *Checker) Suggestions(word string, max int) []string {
	w := Normalize(word)
	if w == "" || max <= 0 {
		return nil
	}
	// Suggest from the apostrophe stem as well (don't->don...).
	if i := strings.IndexAny(w, "'’"); i > 0 {
		w = w[:i]
	}
	qr := []rune(w)
	if len(qr) < 2 {
		return nil
	}

	type candidate struct {
		foldDist int
		dist     int
		prefix   int
		word     string
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	probed := make(map[string]struct{})
	var hits []string
	probe := func(s string) {
		if _, dup := probed[s]; dup || s == w {
			return
		}
		probed[s] = struct{}{}
		i := sort.SearchStrings(c.words, s)
		if i < len(c.words) && c.words[i] == s {
			hits = append(hits, s)
		}
	}

	round1 := edits1(qr, editAlphabet)
	for s := range round1 {
		probe(s)
	}
	for _, variant := range accentVariants(qr) {
		for s := range edits1([]rune(variant), editAlphabet) {
			probe(s)
		}
	}
	for s := range round1 {
		for _, d := range deletesTransposes([]rune(s)) {
			probe(d)
		}
	}

	foldedQuery := foldDiacritics(w)
	out := make([]candidate, 0, len(hits))
	for _, dictWord := range hits {
		fd, _ := osaBounded(foldedQuery, foldDiacritics(dictWord), MaxSuggestionDistance+2)
		d, _ := osaBounded(w, dictWord, MaxSuggestionDistance+2)
		out = append(out, candidate{foldDist: fd, dist: d, word: dictWord, prefix: commonPrefixRunes(w, dictWord)})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].foldDist != out[j].foldDist {
			return out[i].foldDist < out[j].foldDist
		}
		if out[i].dist != out[j].dist {
			return out[i].dist < out[j].dist
		}
		if out[i].prefix != out[j].prefix {
			return out[i].prefix > out[j].prefix
		}
		return out[i].word < out[j].word
	})
	if len(out) > max {
		out = out[:max]
	}
	res := make([]string, len(out))
	for i, cand := range out {
		res[i] = cand.word
	}
	return res
}

// edits1 returns every string at edit distance exactly 1 from q (deletes,
// adjacent transposes, replaces and inserts over alpha), plus q itself.
func edits1(q []rune, alpha []rune) map[string]struct{} {
	out := make(map[string]struct{}, len(q)*len(alpha))
	s := func(r []rune) string { return string(r) }
	out[s(q)] = struct{}{}
	for i := range q {
		// Delete q[i].
		out[s(append(append([]rune{}, q[:i]...), q[i+1:]...))] = struct{}{}
		// Transpose q[i], q[i+1].
		if i+1 < len(q) {
			t := append([]rune{}, q...)
			t[i], t[i+1] = t[i+1], t[i]
			out[s(t)] = struct{}{}
		}
		// Replace q[i].
		for _, r := range alpha {
			if r == q[i] {
				continue
			}
			t := append([]rune{}, q...)
			t[i] = r
			out[s(t)] = struct{}{}
		}
		// Insert before q[i].
		for _, r := range alpha {
			t := append([]rune{}, q[:i]...)
			t = append(t, r)
			t = append(t, q[i:]...)
			out[s(t)] = struct{}{}
		}
	}
	// Insert at the end.
	for _, r := range alpha {
		out[s(append(append([]rune{}, q...), r))] = struct{}{}
	}
	return out
}

// deletesTransposes returns deletes and adjacent transposes of q (the cheap
// subset of distance-1 edits, used for second-order expansion).
func deletesTransposes(q []rune) []string {
	var out []string
	for i := range q {
		out = append(out, string(append(append([]rune{}, q[:i]...), q[i+1:]...)))
		if i+1 < len(q) {
			t := append([]rune{}, q...)
			t[i], t[i+1] = t[i+1], t[i]
			out = append(out, string(t))
		}
	}
	return out
}

// accentAlts maps vowels to their accented counterparts and back (ñ is
// intentionally excluded: n ≠ ñ in Spanish).
var accentAlts = map[rune][]rune{
	'a': {'á'}, 'á': {'a'},
	'e': {'é'}, 'é': {'e'},
	'i': {'í'}, 'í': {'i'},
	'o': {'ó'}, 'ó': {'o'},
	'u': {'ú', 'ü'}, 'ú': {'u'}, 'ü': {'u'},
}

// accentVariants returns the query plus accent-toggled variants: the full
// power set when few vowels are present, single-position toggles otherwise
// (capped to keep generation cheap).
func accentVariants(q []rune) []string {
	var positions []int
	for i, r := range q {
		if _, ok := accentAlts[r]; ok {
			positions = append(positions, i)
		}
	}
	variants := map[string]struct{}{string(q): {}}
	add := func(base []rune, pos int, alt rune) {
		t := append([]rune{}, base...)
		t[pos] = alt
		variants[string(t)] = struct{}{}
	}
	if len(positions) > 5 {
		for _, p := range positions {
			for _, alt := range accentAlts[q[p]] {
				add(q, p, alt)
			}
		}
	} else {
		current := [][]rune{append([]rune{}, q...)}
		for _, p := range positions {
			var next [][]rune
			for _, base := range current {
				next = append(next, base)
				for _, alt := range accentAlts[base[p]] {
					t := append([]rune{}, base...)
					t[p] = alt
					next = append(next, t)
				}
			}
			current = next
		}
		for _, v := range current {
			variants[string(v)] = struct{}{}
		}
	}
	out := make([]string, 0, len(variants))
	for v := range variants {
		out = append(out, v)
	}
	return out
}

// osaBounded computes the Optimal String Alignment distance between a and b
// (runes): Levenshtein plus adjacent transposition (novle->novel costs 1).
// It reports ok=false early when the distance exceeds max.
func osaBounded(a, b string, max int) (dist int, ok bool) {
	ar, br := []rune(a), []rune(b)
	if abs(len(ar)-len(br)) > max {
		return 0, false
	}
	prev2 := make([]int, len(br)+1)
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		cur[0] = i
		rowMin := cur[0]
		for j := 1; j <= len(br); j++ {
			cost := 0
			if ar[i-1] != br[j-1] {
				cost = 1
			}
			best := min3(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && ar[i-1] == br[j-2] && ar[i-2] == br[j-1] {
				if prev2[j-2]+1 < best {
					best = prev2[j-2] + 1
				}
			}
			cur[j] = best
			if best < rowMin {
				rowMin = best
			}
		}
		if rowMin > max {
			return 0, false
		}
		prev2, prev = prev, cur
	}
	if prev[len(br)] > max {
		return 0, false
	}
	return prev[len(br)], true
}

// diacriticFold maps accented Latin vowels (and ç) to their base letter so
// ranking tolerates missing/wrong accents. NOTE: ñ is intentionally NOT
// folded (n ≠ ñ in Spanish) and ü is kept (güe/güi would collide with gue).
var diacriticFold = map[rune]rune{
	'á': 'a', 'à': 'a', 'ä': 'a', 'â': 'a',
	'é': 'e', 'è': 'e', 'ë': 'e', 'ê': 'e',
	'í': 'i', 'ì': 'i', 'ï': 'i', 'î': 'i',
	'ó': 'o', 'ò': 'o', 'ö': 'o', 'ô': 'o',
	'ú': 'u', 'ù': 'u', 'û': 'u',
	'ý': 'y', 'ÿ': 'y',
	'ç': 'c',
}

// foldDiacritics strips accents per diacriticFold (input must be lowercased).
func foldDiacritics(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	changed := false
	for _, r := range s {
		if base, ok := diacriticFold[r]; ok {
			b.WriteRune(base)
			changed = true
		} else {
			b.WriteRune(r)
		}
	}
	if !changed {
		return s
	}
	return b.String()
}

// commonPrefixRunes counts leading equal runes of a and b.
func commonPrefixRunes(a, b string) int {
	ar, br := []rune(a), []rune(b)
	n := 0
	for n < len(ar) && n < len(br) && ar[n] == br[n] {
		n++
	}
	return n
}

// Tokenize splits text into checkable word tokens with rune offsets.
// Letters form words; interior apostrophes and hyphens join (don't,
// ciencia-ficción). Line is the 0-indexed buffer line of each token.
func Tokenize(text string) []Token {
	var tokens []Token
	line := 0
	lineStart := 0 // rune offset where the current line starts
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		r := runes[i]
		if r == '\n' {
			line++
			lineStart = i + 1
			i++
			continue
		}
		if !unicode.IsLetter(r) {
			i++
			continue
		}
		start := i
		for i < len(runes) && (unicode.IsLetter(runes[i]) || isInteriorJoiner(runes, i)) {
			i++
		}
		word := strings.Trim(string(runes[start:i]), "'’-\u2010\u2011")
		if word != "" {
			tokens = append(tokens, Token{
				Word:       word,
				StartRune:  start,
				EndRune:    start + len([]rune(word)),
				Line:       line,
				LineOffset: start - lineStart,
			})
		}
	}
	return tokens
}

// isInteriorJoiner reports whether runes[i] is an apostrophe/hyphen joining
// two letters (so "don't" and "bien-estar" stay single tokens).
func isInteriorJoiner(runes []rune, i int) bool {
	r := runes[i]
	if r != '\'' && r != '’' && r != '-' && r != '\u2010' && r != '\u2011' {
		return false
	}
	return i > 0 && i+1 < len(runes) && unicode.IsLetter(runes[i-1]) && unicode.IsLetter(runes[i+1])
}

// Misspelled returns the tokens in text that fail Check.
func (c *Checker) Misspelled(text string) []Token {
	var bad []Token
	for _, t := range Tokenize(text) {
		if !c.Check(t.Word) {
			bad = append(bad, t)
		}
	}
	return bad
}

// AddCustom records word as user-approved (in memory and on disk).
func (c *Checker) AddCustom(word string) error {
	w := Normalize(word)
	if w == "" {
		return nil
	}
	c.mu.Lock()
	c.custom[w] = struct{}{}
	path := c.customPath
	c.mu.Unlock()
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.WriteString(w + "\n")
	f.Close()
	return err
}

func (c *Checker) loadCustom() error {
	data, err := os.ReadFile(c.customPath)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, line := range strings.Split(string(data), "\n") {
		if w := Normalize(line); w != "" {
			c.custom[w] = struct{}{}
		}
	}
	return nil
}

func readGzLines(blob []byte) ([]string, error) {
	zr, err := gzip.NewReader(bytes.NewReader(blob))
	if err != nil {
		return nil, err
	}
	var lines []string
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	zr.Close()
	return lines, sc.Err()
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
