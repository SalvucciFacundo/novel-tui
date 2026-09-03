package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// SearchService handles novel-wide global search and replacement across chapters.
type SearchService struct {
	chapterRepo domain.ChapterRepository
}

// NewSearchService creates a new SearchService.
func NewSearchService(chapterRepo domain.ChapterRepository) *SearchService {
	return &SearchService{
		chapterRepo: chapterRepo,
	}
}

// Search scans all chapters in the repository for occurrences of query.
func (s *SearchService) Search(query string, caseSensitive bool) ([]domain.SearchMatch, error) {
	if s == nil || s.chapterRepo == nil || query == "" {
		return []domain.SearchMatch{}, nil
	}

	chapters, err := s.chapterRepo.ListAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list chapters: %w", err)
	}

	queryRunes := []rune(query)
	qLen := len(queryRunes)
	if qLen == 0 {
		return []domain.SearchMatch{}, nil
	}

	var matches []domain.SearchMatch

	for _, ch := range chapters {
		content := ch.Content
		if content == "" {
			loaded, err := s.chapterRepo.LoadContent(ch.ID)
			if err == nil {
				content = loaded
			}
		}

		lines := strings.Split(content, "\n")
		for lineIdx, rawLine := range lines {
			line := strings.TrimRight(rawLine, "\r")
			lineRunes := []rune(line)
			lLen := len(lineRunes)

			for i := 0; i <= lLen-qLen; i++ {
				matched := true
				for k := 0; k < qLen; k++ {
					rLine := lineRunes[i+k]
					rQuery := queryRunes[k]
					if caseSensitive {
						if rLine != rQuery {
							matched = false
							break
						}
					} else {
						if unicode.ToLower(rLine) != unicode.ToLower(rQuery) {
							matched = false
							break
						}
					}
				}

				if matched {
					matchText := string(lineRunes[i : i+qLen])
					matches = append(matches, domain.SearchMatch{
						ChapterID:    ch.ID,
						ChapterTitle: ch.Title,
						FilePath:     ch.FilePath,
						LineNumber:   lineIdx + 1,
						Column:       i + 1,
						LineText:     line,
						MatchText:    matchText,
					})
					i += qLen - 1
				}
			}
		}
	}

	return matches, nil
}

// ReplaceAll replaces all occurrences of query with replacement across all chapters.
func (s *SearchService) ReplaceAll(query string, replacement string, caseSensitive bool) (domain.SearchReplaceResult, error) {
	result := domain.SearchReplaceResult{}
	if s == nil || s.chapterRepo == nil || query == "" {
		return result, nil
	}

	chapters, err := s.chapterRepo.ListAll()
	if err != nil {
		return result, fmt.Errorf("failed to list chapters: %w", err)
	}

	var pattern *regexp.Regexp
	if !caseSensitive {
		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(query))
		if err != nil {
			return result, fmt.Errorf("invalid search pattern: %w", err)
		}
		pattern = re
	}

	for _, ch := range chapters {
		content := ch.Content
		if content == "" {
			loaded, err := s.chapterRepo.LoadContent(ch.ID)
			if err == nil {
				content = loaded
			}
		}

		var count int
		var newContent string

		if caseSensitive {
			count = strings.Count(content, query)
			if count > 0 {
				newContent = strings.ReplaceAll(content, query, replacement)
			}
		} else {
			allIndices := pattern.FindAllStringIndex(content, -1)
			count = len(allIndices)
			if count > 0 {
				newContent = pattern.ReplaceAllLiteralString(content, replacement)
			}
		}

		if count > 0 {
			if err := s.chapterRepo.SaveContent(ch.ID, newContent); err != nil {
				return result, fmt.Errorf("failed to save chapter %s: %w", ch.ID, err)
			}
			result.TotalMatches += count
			result.TotalReplaced += count
			result.ChaptersAffected++
		}
	}

	return result, nil
}
