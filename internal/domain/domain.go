package domain

// Chapter represents a single writing unit.
type Chapter struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	FilePath  string `json:"file_path"`
	WordCount int    `json:"word_count"`
	Content   string `json:"content,omitempty"`
}

// Character tracks lore and context for the secondary sidebar tab.
type Character struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
}

// EditorMetrics represents real-time typing metrics.
type EditorMetrics struct {
	WordCount   int  `json:"word_count"`
	CharCount   int  `json:"char_count"`
	ReadingTime int  `json:"reading_time"` // in minutes
	IsDirty     bool `json:"is_dirty"`
}

// ChapterRepository defines data persistence contracts for chapters.
type ChapterRepository interface {
	ListAll() ([]Chapter, error)
	LoadContent(id string) (string, error)
	SaveContent(id string, content string) error
	Create(title string) (Chapter, error)
}

// CharacterRepository defines data persistence contracts for characters/lore.
type CharacterRepository interface {
	ListAll() ([]Character, error)
	SaveAll(chars []Character) error
}
