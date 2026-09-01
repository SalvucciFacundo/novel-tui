package service_test

import (
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/service"
)

func TestCalculateMetrics(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		isDirty         bool
		wantWordCount   int
		wantCharCount   int
		wantReadingTime int
	}{
		{
			name:            "empty content",
			content:         "",
			isDirty:         false,
			wantWordCount:   0,
			wantCharCount:   0,
			wantReadingTime: 0,
		},
		{
			name:            "simple sentence",
			content:         "The quick brown fox jumps over the lazy dog.",
			isDirty:         true,
			wantWordCount:   9,
			wantCharCount:   44,
			wantReadingTime: 1,
		},
		{
			name:            "unicode content",
			content:         "こんにちは 世界! Hello World!",
			isDirty:         false,
			wantWordCount:   4,
			wantCharCount:   22,
			wantReadingTime: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := service.CalculateMetrics(tt.content, tt.isDirty)
			if m.WordCount != tt.wantWordCount {
				t.Errorf("WordCount = %d, want %d", m.WordCount, tt.wantWordCount)
			}
			if m.CharCount != tt.wantCharCount {
				t.Errorf("CharCount = %d, want %d", m.CharCount, tt.wantCharCount)
			}
			if m.ReadingTime != tt.wantReadingTime {
				t.Errorf("ReadingTime = %d, want %d", m.ReadingTime, tt.wantReadingTime)
			}
			if m.IsDirty != tt.isDirty {
				t.Errorf("IsDirty = %v, want %v", m.IsDirty, tt.isDirty)
			}
		})
	}
}

func TestFormatReadingTime(t *testing.T) {
	if got := service.FormatReadingTime(0); got != "~1 min read" {
		t.Errorf("FormatReadingTime(0) = %s, want ~1 min read", got)
	}
	if got := service.FormatReadingTime(1); got != "~1 min read" {
		t.Errorf("FormatReadingTime(1) = %s, want ~1 min read", got)
	}
	if got := service.FormatReadingTime(5); got != "~5 min read" {
		t.Errorf("FormatReadingTime(5) = %s, want ~5 min read", got)
	}
}
