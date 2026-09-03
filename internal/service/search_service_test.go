package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service"
)

type mockChapterRepo struct {
	chapters []domain.Chapter
	contents map[string]string
	saveErr  error
	listErr  error
}

func newMockChapterRepo() *mockChapterRepo {
	return &mockChapterRepo{
		contents: make(map[string]string),
	}
}

func (m *mockChapterRepo) ListAll() ([]domain.Chapter, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var res []domain.Chapter
	for _, ch := range m.chapters {
		c := ch
		if content, ok := m.contents[ch.ID]; ok {
			c.Content = content
		}
		res = append(res, c)
	}
	return res, nil
}

func (m *mockChapterRepo) LoadContent(id string) (string, error) {
	if content, ok := m.contents[id]; ok {
		return content, nil
	}
	return "", errors.New("chapter not found")
}

func (m *mockChapterRepo) SaveContent(id string, content string) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.contents[id] = content
	for i := range m.chapters {
		if m.chapters[i].ID == id {
			m.chapters[i].Content = content
			break
		}
	}
	return nil
}

func (m *mockChapterRepo) Create(title string) (domain.Chapter, error) {
	ch := domain.Chapter{
		ID:       title,
		Title:    title,
		FilePath: title + ".txt",
	}
	m.chapters = append(m.chapters, ch)
	return ch, nil
}

func TestSearchService_Search_EmptyQuery(t *testing.T) {
	repo := newMockChapterRepo()
	svc := service.NewSearchService(repo)

	matches, err := svc.Search("", false)
	if err != nil {
		t.Fatalf("unexpected error on empty query: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty query, got %d", len(matches))
	}
}

func TestSearchService_Search_CaseSensitiveAndInsensitive(t *testing.T) {
	repo := newMockChapterRepo()
	repo.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1", FilePath: "capitulos/01_cap1.txt"},
		{ID: "02_cap2", Title: "Capítulo 2", FilePath: "capitulos/02_cap2.txt"},
	}
	repo.contents["01_cap1"] = "El Dragón despertó en la cueva.\nEl dragón rugió ferozmente."
	repo.contents["02_cap2"] = "No había ningún Dragón aquí.\nTodo estaba en calma."

	svc := service.NewSearchService(repo)

	// 1. Case-sensitive search for "Dragón"
	matches, err := svc.Search("Dragón", true)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 case-sensitive matches for 'Dragón', got %d", len(matches))
	}

	if matches[0].ChapterID != "01_cap1" || matches[0].LineNumber != 1 || matches[0].Column != 4 {
		t.Errorf("unexpected match 0: %+v", matches[0])
	}
	if matches[1].ChapterID != "02_cap2" || matches[1].LineNumber != 1 || matches[1].Column != 17 {
		t.Errorf("unexpected match 1: %+v", matches[1])
	}

	// 2. Case-insensitive search for "dragón"
	matchesAll, err := svc.Search("dragón", false)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(matchesAll) != 3 {
		t.Fatalf("expected 3 case-insensitive matches for 'dragón', got %d", len(matchesAll))
	}
}

func TestSearchService_Search_MultipleMatchesOnSameLine(t *testing.T) {
	repo := newMockChapterRepo()
	repo.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1", FilePath: "capitulos/01_cap1.txt"},
	}
	repo.contents["01_cap1"] = "hola mundo, hola a todos, hola de nuevo"

	svc := service.NewSearchService(repo)
	matches, err := svc.Search("hola", false)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches on same line, got %d", len(matches))
	}

	expectedCols := []int{1, 13, 27}
	for i, expCol := range expectedCols {
		if matches[i].Column != expCol {
			t.Errorf("match %d: expected column %d, got %d", i, expCol, matches[i].Column)
		}
		if matches[i].LineNumber != 1 {
			t.Errorf("match %d: expected line 1, got %d", i, matches[i].LineNumber)
		}
	}
}

func TestSearchService_Search_SpecialRegexCharacters(t *testing.T) {
	repo := newMockChapterRepo()
	repo.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1", FilePath: "capitulos/01_cap1.txt"},
	}
	repo.contents["01_cap1"] = "Versión [1.0] (beta)* lista.\nOtra línea normal."

	svc := service.NewSearchService(repo)
	matches, err := svc.Search("[1.0] (beta)*", false)
	if err != nil {
		t.Fatalf("Search with special characters failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Column != 9 {
		t.Errorf("expected column 9, got %d", matches[0].Column)
	}
}

func TestSearchService_ReplaceAll_EmptyQuery(t *testing.T) {
	repo := newMockChapterRepo()
	svc := service.NewSearchService(repo)

	res, err := svc.ReplaceAll("", "replacement", false)
	if err != nil {
		t.Fatalf("unexpected error on empty replace query: %v", err)
	}
	if res.TotalReplaced != 0 || res.ChaptersAffected != 0 {
		t.Errorf("expected 0 replacements, got %+v", res)
	}
}

func TestSearchService_ReplaceAll_CaseSensitive(t *testing.T) {
	repo := newMockChapterRepo()
	repo.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1", FilePath: "capitulos/01_cap1.txt"},
		{ID: "02_cap2", Title: "Capítulo 2", FilePath: "capitulos/02_cap2.txt"},
	}
	repo.contents["01_cap1"] = "El Dragón despertó.\nEl dragón rugió."
	repo.contents["02_cap2"] = "El Dragón voló."

	svc := service.NewSearchService(repo)
	res, err := svc.ReplaceAll("Dragón", "Fénix", true)
	if err != nil {
		t.Fatalf("ReplaceAll failed: %v", err)
	}

	if res.TotalMatches != 2 || res.TotalReplaced != 2 || res.ChaptersAffected != 2 {
		t.Errorf("unexpected replace result: %+v", res)
	}

	if repo.contents["01_cap1"] != "El Fénix despertó.\nEl dragón rugió." {
		t.Errorf("unexpected content in cap1: %s", repo.contents["01_cap1"])
	}
	if repo.contents["02_cap2"] != "El Fénix voló." {
		t.Errorf("unexpected content in cap2: %s", repo.contents["02_cap2"])
	}
}

func TestSearchService_ReplaceAll_CaseInsensitive(t *testing.T) {
	repo := newMockChapterRepo()
	repo.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1", FilePath: "capitulos/01_cap1.txt"},
		{ID: "02_cap2", Title: "Capítulo 2", FilePath: "capitulos/02_cap2.txt"},
	}
	repo.contents["01_cap1"] = "El DRAGÓN y el dragón y el Dragón."
	repo.contents["02_cap2"] = "Sin monstruos aquí."

	svc := service.NewSearchService(repo)
	res, err := svc.ReplaceAll("dragón", "Fénix", false)
	if err != nil {
		t.Fatalf("ReplaceAll failed: %v", err)
	}

	if res.TotalMatches != 3 || res.TotalReplaced != 3 || res.ChaptersAffected != 1 {
		t.Errorf("unexpected replace result: %+v", res)
	}

	expectedCap1 := "El Fénix y el Fénix y el Fénix."
	if repo.contents["01_cap1"] != expectedCap1 {
		t.Errorf("unexpected content in cap1: %s", repo.contents["01_cap1"])
	}
}

func TestSearchService_ReplaceAll_SpecialReplacementTokens(t *testing.T) {
	repo := newMockChapterRepo()
	repo.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1", FilePath: "capitulos/01_cap1.txt"},
	}
	repo.contents["01_cap1"] = "Precios: $100 y $200"

	svc := service.NewSearchService(repo)
	res, err := svc.ReplaceAll("$100", "$1 (USD)", false)
	if err != nil {
		t.Fatalf("ReplaceAll failed: %v", err)
	}
	if res.TotalReplaced != 1 {
		t.Fatalf("expected 1 replacement, got %d", res.TotalReplaced)
	}
	if !strings.Contains(repo.contents["01_cap1"], "$1 (USD) y $200") {
		t.Errorf("expected literal replacement without $ expansion, got %s", repo.contents["01_cap1"])
	}
}

func TestSearchService_ErrorsAndEdgeCases(t *testing.T) {
	// 1. Nil receiver / nil repo
	var nilSvc *service.SearchService
	m, err := nilSvc.Search("test", false)
	if err != nil || len(m) != 0 {
		t.Errorf("expected empty result on nil service search")
	}
	res, err := nilSvc.ReplaceAll("test", "rep", false)
	if err != nil || res.TotalReplaced != 0 {
		t.Errorf("expected zero result on nil service replace")
	}

	// 2. List error
	repo := newMockChapterRepo()
	repo.listErr = errors.New("io error")
	svc := service.NewSearchService(repo)
	_, err = svc.Search("test", false)
	if err == nil {
		t.Errorf("expected error when ListAll fails")
	}
	_, err = svc.ReplaceAll("test", "rep", false)
	if err == nil {
		t.Errorf("expected error when ListAll fails during replace")
	}

	// 3. Save error during ReplaceAll
	repo2 := newMockChapterRepo()
	repo2.chapters = []domain.Chapter{
		{ID: "01_cap1", Title: "Capítulo 1"},
	}
	repo2.contents["01_cap1"] = "un dragón aquí"
	repo2.saveErr = errors.New("disk full")
	svc2 := service.NewSearchService(repo2)
	_, err = svc2.ReplaceAll("dragón", "fénix", false)
	if err == nil {
		t.Errorf("expected error when SaveContent fails")
	}
}
