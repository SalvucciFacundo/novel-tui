package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// FileChapterRepository implements domain.ChapterRepository using the local filesystem.
type FileChapterRepository struct {
	baseDir string
}

// NewFileChapterRepository creates a new FileChapterRepository.
// It ensures the target chapters directory exists.
func NewFileChapterRepository(baseDir string) (*FileChapterRepository, error) {
	chaptersDir := filepath.Join(baseDir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chapters directory: %w", err)
	}
	return &FileChapterRepository{baseDir: baseDir}, nil
}

func (r *FileChapterRepository) chaptersDir() string {
	return filepath.Join(r.baseDir, "chapters")
}

// ListAll scans and parses all .md files in the chapters/ directory.
func (r *FileChapterRepository) ListAll() ([]domain.Chapter, error) {
	dir := r.chaptersDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read chapters directory: %w", err)
	}

	var chapters []domain.Chapter
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}

		filename := entry.Name()
		id := strings.TrimSuffix(filename, filepath.Ext(filename))
		fullPath := filepath.Join(dir, filename)

		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		content := string(contentBytes)
		title := extractTitle(id, content)
		wordCount := countWords(content)

		relPath := filepath.Join("chapters", filename)
		chapters = append(chapters, domain.Chapter{
			ID:        id,
			Title:     title,
			FilePath:  relPath,
			WordCount: wordCount,
			Content:   content,
		})
	}

	// Sort chapters by ID / filename for consistent ordering
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].ID < chapters[j].ID
	})

	return chapters, nil
}

// LoadContent reads the full content for a given chapter ID.
func (r *FileChapterRepository) LoadContent(id string) (string, error) {
	filePath := filepath.Join(r.chaptersDir(), id+".md")
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read chapter %s: %w", id, err)
	}
	return string(contentBytes), nil
}

// SaveContent writes content atomically using a temporary file.
func (r *FileChapterRepository) SaveContent(id string, content string) error {
	dir := r.chaptersDir()
	targetPath := filepath.Join(dir, id+".md")
	tmpPath := filepath.Join(dir, fmt.Sprintf(".%s.tmp", id))

	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file to target %s: %w", targetPath, err)
	}

	return nil
}

// Create generates a new chapter file with the given title and returns domain.Chapter.
func (r *FileChapterRepository) Create(title string) (domain.Chapter, error) {
	if strings.TrimSpace(title) == "" {
		title = "Untitled Chapter"
	}

	slug := slugify(title)
	dir := r.chaptersDir()

	// Ensure unique slug
	baseSlug := slug
	counter := 1
	for {
		candidate := filepath.Join(dir, slug+".md")
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	initialContent := fmt.Sprintf("# %s\n\n", title)
	if err := r.SaveContent(slug, initialContent); err != nil {
		return domain.Chapter{}, err
	}

	return domain.Chapter{
		ID:        slug,
		Title:     title,
		FilePath:  filepath.Join("chapters", slug+".md"),
		WordCount: countWords(initialContent),
		Content:   initialContent,
	}, nil
}

// Helpers
var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "chapter"
	}
	return s
}

func extractTitle(id, content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title := strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			if title != "" {
				return title
			}
		}
	}
	// Fallback: humanize ID
	formatted := strings.ReplaceAll(id, "-", " ")
	formatted = strings.ReplaceAll(formatted, "_", " ")
	return strings.Title(formatted)
}

func countWords(s string) int {
	fields := strings.Fields(s)
	return len(fields)
}
