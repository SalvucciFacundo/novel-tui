package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// FileChapterRepository implements domain.ChapterRepository using the local filesystem.
type FileChapterRepository struct {
	baseDir string
}

// NewFileChapterRepository creates a new FileChapterRepository.
// It ensures the target capitulos (or legacy chapters) directory exists.
func NewFileChapterRepository(baseDir string) (*FileChapterRepository, error) {
	baseDir = ExpandHome(baseDir)
	repo := &FileChapterRepository{baseDir: baseDir}
	dir := repo.chaptersDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chapters directory: %w", err)
	}
	return repo, nil
}

func (r *FileChapterRepository) chaptersDir() string {
	capitulos := filepath.Join(r.baseDir, "capitulos")
	if _, err := os.Stat(capitulos); err == nil {
		return capitulos
	}
	legacyChapters := filepath.Join(r.baseDir, "chapters")
	if _, err := os.Stat(legacyChapters); err == nil {
		return legacyChapters
	}
	return capitulos
}

// ListAll scans and parses all .txt and .md files in the chapters/capitulos directory.
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
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".txt" && ext != ".md" {
			continue
		}

		id := strings.TrimSuffix(filename, filepath.Ext(filename))
		fullPath := filepath.Join(dir, filename)

		contentBytes, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		content := string(contentBytes)
		title := extractTitle(id, content)
		wordCount := countWords(content)

		relDir := filepath.Base(dir)
		relPath := filepath.Join(relDir, filename)
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

// resolveFile resolves the actual filename for a given chapter ID (.txt, .md, or exact).
func (r *FileChapterRepository) resolveFile(id string) string {
	dir := r.chaptersDir()
	// Try txt first
	txtPath := filepath.Join(dir, id+".txt")
	if _, err := os.Stat(txtPath); err == nil {
		return txtPath
	}
	// Try md
	mdPath := filepath.Join(dir, id+".md")
	if _, err := os.Stat(mdPath); err == nil {
		return mdPath
	}
	// Try exact id
	exactPath := filepath.Join(dir, id)
	if _, err := os.Stat(exactPath); err == nil {
		return exactPath
	}

	// Default fallback: if dir base is "chapters", use .md for backward compat, else .txt
	if filepath.Base(dir) == "chapters" {
		return mdPath
	}
	return txtPath
}

// LoadContent reads the full content for a given chapter ID.
func (r *FileChapterRepository) LoadContent(id string) (string, error) {
	filePath := r.resolveFile(id)
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read chapter %s: %w", id, err)
	}
	return string(contentBytes), nil
}

// SaveContent writes content atomically using a temporary file.
func (r *FileChapterRepository) SaveContent(id string, content string) error {
	dir := r.chaptersDir()
	targetPath := r.resolveFile(id)
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

	dir := r.chaptersDir()
	isLegacy := filepath.Base(dir) == "chapters"
	ext := ".txt"
	if isLegacy {
		ext = ".md"
	}

	slug := slugify(title)

	// Check if this is sequential (starts with number) or needs sequence
	entries, _ := os.ReadDir(dir)
	maxSeq := 0
	numRegex := regexp.MustCompile(`^(\d+)`)
	for _, entry := range entries {
		matches := numRegex.FindStringSubmatch(entry.Name())
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil && num > maxSeq {
				maxSeq = num
			}
		}
	}

	var filename string
	var id string
	if isLegacy {
		// Legacy chapter naming: chapter-1-the-gathering.md
		baseSlug := slug
		counter := 1
		id = baseSlug
		for {
			candidate := filepath.Join(dir, id+ext)
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				break
			}
			id = fmt.Sprintf("%s-%d", baseSlug, counter)
			counter++
		}
		filename = id + ext
	} else {
		// New convention: 01_capitulo_1.txt
		nextSeq := maxSeq + 1
		id = fmt.Sprintf("%02d_%s", nextSeq, strings.ReplaceAll(slug, "-", "_"))
		filename = id + ext
	}

	initialContent := fmt.Sprintf("# %s\n\n", title)
	targetPath := filepath.Join(dir, filename)
	if err := os.WriteFile(targetPath, []byte(initialContent), 0644); err != nil {
		return domain.Chapter{}, err
	}

	relDir := filepath.Base(dir)
	return domain.Chapter{
		ID:        id,
		Title:     title,
		FilePath:  filepath.Join(relDir, filename),
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
