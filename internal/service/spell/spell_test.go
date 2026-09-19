package spell

import (
	"path/filepath"
	"testing"
	"time"
)

func loadTestChecker(t *testing.T) *Checker {
	t.Helper()
	c, err := Load(filepath.Join(t.TempDir(), "custom_words.txt"))
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if c.WordCount() < 1000000 {
		t.Fatalf("expected 1M+ combined entries, got %d", c.WordCount())
	}
	return c
}

func TestCheckBothLanguages(t *testing.T) {
	c := loadTestChecker(t)
	valid := []string{
		// Spanish (incl. conjugations and accents)
		"hablábamos", "había", "corazón", "rápidamente", "también",
		"novela", "dragón", "murmullo",
		// English
		"novel", "dragon", "whisper", "written", "quickly",
		// Case-insensitive
		"Corazón", "NOVEL",
		// Contractions and compounds
		"don't", "it's", "ciencia-ficción",
		// Non-words that must pass
		"cap12", "2026", "",
	}
	for _, w := range valid {
		if !c.Check(w) {
			t.Errorf("expected %q to pass Check", w)
		}
	}

	invalid := []string{"hablabamos", "corason", "novle", "writen", "draggon", "xqzpt"}
	for _, w := range invalid {
		if c.Check(w) {
			t.Errorf("expected %q to fail Check", w)
		}
	}
}

func TestSuggestions(t *testing.T) {
	c := loadTestChecker(t)
	cases := []struct {
		typo string
		want string
	}{
		{"novle", "novel"},
		{"corason", "corazón"},
		{"draggon", "dragon"},
		{"hablabamos", "hablábamos"},
	}
	for _, tt := range cases {
		start := time.Now()
		got := c.Suggestions(tt.typo, 8)
		if elapsed := time.Since(start); elapsed > 5*time.Second {
			t.Errorf("Suggestions(%q) took too long: %v", tt.typo, elapsed)
		}
		found := false
		for _, s := range got {
			if s == tt.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Suggestions(%q) = %v, want %q included", tt.typo, got, tt.want)
		}
	}

	if got := c.Suggestions("x", 6); len(got) != 0 {
		t.Errorf("expected no suggestions for single rune, got %v", got)
	}
}

func TestCustomWords(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom_words.txt")
	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// NOTE: real names like Aurelio ship in the hunspell Spanish dictionary,
	// so the test uses an invented fantasy name.
	if c.Check("Kaelith") {
		t.Fatalf("expected invented name to fail before adding")
	}
	if err := c.AddCustom("Kaelith"); err != nil {
		t.Fatalf("AddCustom failed: %v", err)
	}
	if !c.Check("kaelith") {
		t.Errorf("expected custom word to pass (case-insensitive)")
	}

	// Persistence across reloads.
	c2, err := Load(path)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	if !c2.Check("Kaelith") {
		t.Errorf("expected custom word to persist across reloads")
	}
}

func TestTokenize(t *testing.T) {
	tokens := Tokenize("Hola, mundo!\ndon't bien-estar 123")
	words := []string{"Hola", "mundo", "don't", "bien-estar"}
	if len(tokens) != len(words) {
		t.Fatalf("expected %d tokens, got %+v", len(words), tokens)
	}
	for i, w := range words {
		if tokens[i].Word != w {
			t.Errorf("token %d: expected %q, got %q", i, w, tokens[i].Word)
		}
	}
	if tokens[2].Line != 1 || tokens[2].LineOffset != 0 {
		t.Errorf("expected don't at line 1 offset 0, got %+v", tokens[2])
	}
	// "123" is not a token (no letters).
	for _, tok := range tokens {
		if tok.Word == "123" {
			t.Errorf("digits should not tokenize: %+v", tok)
		}
	}
}
