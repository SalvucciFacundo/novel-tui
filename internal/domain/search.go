package domain

// SearchMatch represents an occurrence found in a chapter during global search.
type SearchMatch struct {
	ChapterID    string `json:"chapter_id"`
	ChapterTitle string `json:"chapter_title"`
	FilePath     string `json:"file_path"`
	LineNumber   int    `json:"line_number"` // 1-indexed
	Column       int    `json:"column"`      // 1-indexed (character column)
	LineText     string `json:"line_text"`
	MatchText    string `json:"match_text"`
}

// SearchReplaceResult summarizes the outcome of a global find-and-replace operation.
type SearchReplaceResult struct {
	TotalMatches     int `json:"total_matches"`
	TotalReplaced    int `json:"total_replaced"`
	ChaptersAffected int `json:"chapters_affected"`
}
