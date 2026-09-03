package domain_test

import (
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

func TestSearchMatch_Properties(t *testing.T) {
	match := domain.SearchMatch{
		ChapterID:    "chap-1",
		ChapterTitle: "Capítulo 1: El Inicio",
		FilePath:     "/path/to/chap1.md",
		LineNumber:   10,
		Column:       5,
		LineText:     "Había una vez en el reino lejano.",
		MatchText:    "reino",
	}

	if match.ChapterID != "chap-1" {
		t.Errorf("expected ChapterID 'chap-1', got %s", match.ChapterID)
	}
	if match.LineNumber != 10 {
		t.Errorf("expected LineNumber 10, got %d", match.LineNumber)
	}
	if match.MatchText != "reino" {
		t.Errorf("expected MatchText 'reino', got %s", match.MatchText)
	}
}

func TestSearchReplaceResult_Properties(t *testing.T) {
	result := domain.SearchReplaceResult{
		TotalMatches:     12,
		TotalReplaced:    12,
		ChaptersAffected: 3,
	}

	if result.TotalMatches != 12 {
		t.Errorf("expected TotalMatches 12, got %d", result.TotalMatches)
	}
	if result.TotalReplaced != 12 {
		t.Errorf("expected TotalReplaced 12, got %d", result.TotalReplaced)
	}
	if result.ChaptersAffected != 3 {
		t.Errorf("expected ChaptersAffected 3, got %d", result.ChaptersAffected)
	}
}
