package service

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

const WordsPerMinute = 225

// CalculateMetrics computes word count, char count, and estimated reading time.
func CalculateMetrics(content string, isDirty bool) domain.EditorMetrics {
	words := CountWords(content)
	chars := CountChars(content)
	readingTime := CalculateReadingTime(words)

	return domain.EditorMetrics{
		WordCount:   words,
		CharCount:   chars,
		ReadingTime: readingTime,
		IsDirty:     isDirty,
	}
}

// CountWords counts the number of whitespace-delimited words in the text.
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// CountChars returns the total rune (character) count.
func CountChars(s string) int {
	return utf8.RuneCountInString(s)
}

// CalculateReadingTime calculates estimated reading time in minutes (based on 225 wpm).
func CalculateReadingTime(wordCount int) int {
	if wordCount <= 0 {
		return 0
	}
	mins := int(math.Ceil(float64(wordCount) / float64(WordsPerMinute)))
	if mins < 1 {
		return 1
	}
	return mins
}

// FormatReadingTime formats reading time as "~X min read".
func FormatReadingTime(minutes int) string {
	if minutes <= 1 {
		return "~1 min read"
	}
	return fmt.Sprintf("~%d min read", minutes)
}

// FormatMetrics formats standard metrics string for display.
func FormatMetrics(m domain.EditorMetrics) string {
	return fmt.Sprintf("%d words | %d chars | %s", m.WordCount, m.CharCount, FormatReadingTime(m.ReadingTime))
}
