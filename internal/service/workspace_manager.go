package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

var (
	ErrEmptyTitle       = errors.New("novel title cannot be empty")
	ErrNovelExists      = errors.New("a novel with this name already exists")
	ErrInvalidPath      = errors.New("invalid novel path or path traversal detected")
	ErrNovelNotFound    = errors.New("novel directory does not exist")
	ErrChaptersNotFound = errors.New("capitulos directory not found")
)

// WorkspaceManager defines operations for managing multi-project workspaces on disk.
type WorkspaceManager interface {
	Initialize(rootDir string) error
	ListRecentNovels(rootDir string) ([]domain.NovelMetadata, error)
	CreateNovel(rootDir, title string) (*domain.NovelMetadata, error)
	CreateChapter(novelDir, chapterTitle string) (string, error)
}

// DefaultWorkspaceManager implements WorkspaceManager.
type DefaultWorkspaceManager struct{}

// NewWorkspaceManager returns a new instance of DefaultWorkspaceManager.
func NewWorkspaceManager() *DefaultWorkspaceManager {
	return &DefaultWorkspaceManager{}
}

// Initialize ensures the workspace root directory exists.
func (m *DefaultWorkspaceManager) Initialize(rootDir string) error {
	resolved := repository.ExpandHome(rootDir)
	if resolved == "" {
		return errors.New("root directory cannot be empty")
	}
	return os.MkdirAll(resolved, 0755)
}

// Slugify converts a title into a safe directory or filename slug.
func Slugify(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}

	// Normalize characters
	var b strings.Builder
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case unicode.IsSpace(r) || r == '-' || r == '_':
			b.WriteRune('_')
		}
	}

	slug := b.String()
	// Replace consecutive underscores
	dupUnderscore := regexp.MustCompile(`_+`)
	slug = dupUnderscore.ReplaceAllString(slug, "_")
	slug = strings.Trim(slug, "_")
	return slug
}

// ListRecentNovels scans rootDir for novel directories and returns their metadata.
func (m *DefaultWorkspaceManager) ListRecentNovels(rootDir string) ([]domain.NovelMetadata, error) {
	resolved := repository.ExpandHome(rootDir)
	if err := m.Initialize(resolved); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(resolved)
	if err != nil {
		return nil, fmt.Errorf("failed to read root directory %s: %w", resolved, err)
	}

	var novels []domain.NovelMetadata

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		novelDir := filepath.Join(resolved, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		capitulosDir := filepath.Join(novelDir, "capitulos")
		if _, err := os.Stat(capitulosDir); os.IsNotExist(err) {
			// Check legacy chapters dir
			capitulosDir = filepath.Join(novelDir, "chapters")
		}

		chapterCount := 0
		lastMod := info.ModTime()

		if chEntries, err := os.ReadDir(capitulosDir); err == nil {
			for _, chEntry := range chEntries {
				if chEntry.IsDir() || strings.HasPrefix(chEntry.Name(), ".") {
					continue
				}
				name := strings.ToLower(chEntry.Name())
				if strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".md") {
					chapterCount++
					if chInfo, err := chEntry.Info(); err == nil {
						if chInfo.ModTime().After(lastMod) {
							lastMod = chInfo.ModTime()
						}
					}
				}
			}
		}

		title := entry.Name()
		novels = append(novels, domain.NovelMetadata{
			Title:        title,
			AbsolutePath: novelDir,
			ChapterCount: chapterCount,
			LastModified: lastMod,
		})
	}

	// Sort by LastModified descending
	sort.Slice(novels, func(i, j int) bool {
		return novels[i].LastModified.After(novels[j].LastModified)
	})

	return novels, nil
}

// CreateNovel scaffolds a new novel workspace directory structure.
func (m *DefaultWorkspaceManager) CreateNovel(rootDir, title string) (*domain.NovelMetadata, error) {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return nil, ErrEmptyTitle
	}
	if strings.Contains(trimmedTitle, "..") || strings.Contains(trimmedTitle, "/") || strings.Contains(trimmedTitle, "\\") {
		return nil, ErrInvalidPath
	}

	resolvedRoot := repository.ExpandHome(rootDir)
	if err := m.Initialize(resolvedRoot); err != nil {
		return nil, err
	}

	slug := Slugify(trimmedTitle)
	if slug == "" {
		slug = "novela"
	}

	novelDir := filepath.Join(resolvedRoot, slug)

	// Path traversal protection
	rel, err := filepath.Rel(resolvedRoot, novelDir)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return nil, ErrInvalidPath
	}

	// Collision check
	if _, err := os.Stat(novelDir); !os.IsNotExist(err) {
		return nil, ErrNovelExists
	}

	// Create directories
	capitulosDir := filepath.Join(novelDir, "capitulos")
	if err := os.MkdirAll(capitulosDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create capitulos directory: %w", err)
	}

	// Starter chapter: 01_capitulo_1.txt
	initialChapterPath := filepath.Join(capitulosDir, "01_capitulo_1.txt")
	initialChapterContent := fmt.Sprintf("# %s - Capítulo 1\n\nComienza a escribir aquí...\n", trimmedTitle)
	if err := os.WriteFile(initialChapterPath, []byte(initialChapterContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create initial chapter: %w", err)
	}

	// Starter personajes.json
	personajesPath := filepath.Join(novelDir, "personajes.json")
	if err := os.WriteFile(personajesPath, []byte("[]\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to create personajes.json: %w", err)
	}

	// Starter notas.txt
	notasPath := filepath.Join(novelDir, "notas.txt")
	notasContent := fmt.Sprintf("Notas y libreta de ideas para: %s\n\n", trimmedTitle)
	if err := os.WriteFile(notasPath, []byte(notasContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create notas.txt: %w", err)
	}

	return &domain.NovelMetadata{
		Title:        trimmedTitle,
		AbsolutePath: novelDir,
		ChapterCount: 1,
		LastModified: time.Now(),
	}, nil
}

// CreateChapter creates a new chapter with sequential numbering in the novel's capitulos directory.
func (m *DefaultWorkspaceManager) CreateChapter(novelDir, chapterTitle string) (string, error) {
	resolvedNovelDir := repository.ExpandHome(novelDir)
	if _, err := os.Stat(resolvedNovelDir); os.IsNotExist(err) {
		return "", ErrNovelNotFound
	}

	capitulosDir := filepath.Join(resolvedNovelDir, "capitulos")
	if _, err := os.Stat(capitulosDir); os.IsNotExist(err) {
		// Fallback to chapters if present
		legacyDir := filepath.Join(resolvedNovelDir, "chapters")
		if _, lErr := os.Stat(legacyDir); lErr == nil {
			capitulosDir = legacyDir
		} else {
			if err := os.MkdirAll(capitulosDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create capitulos dir: %w", err)
			}
		}
	}

	trimmedTitle := strings.TrimSpace(chapterTitle)
	if trimmedTitle == "" {
		trimmedTitle = "Nuevo Capítulo"
	}

	entries, err := os.ReadDir(capitulosDir)
	if err != nil {
		return "", fmt.Errorf("failed to read capitulos directory: %w", err)
	}

	maxSeq := 0
	numRegex := regexp.MustCompile(`^(\d+)`)

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		matches := numRegex.FindStringSubmatch(entry.Name())
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				if num > maxSeq {
					maxSeq = num
				}
			}
		}
	}

	nextSeq := maxSeq + 1
	slug := Slugify(trimmedTitle)
	if slug == "" {
		slug = "capitulo"
	}

	filename := fmt.Sprintf("%02d_%s.txt", nextSeq, slug)
	chapterPath := filepath.Join(capitulosDir, filename)

	initialContent := fmt.Sprintf("# %s\n\n", trimmedTitle)
	if err := os.WriteFile(chapterPath, []byte(initialContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create chapter file %s: %w", chapterPath, err)
	}

	return chapterPath, nil
}
