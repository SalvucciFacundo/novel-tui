package domain_test

import (
	"strings"
	"testing"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

func TestContentRatings(t *testing.T) {
	ratings := domain.AllContentRatings()
	if len(ratings) != 4 {
		t.Fatalf("expected 4 content ratings, got %d", len(ratings))
	}

	expected := map[domain.ContentRating]string{
		domain.RatingAllAges:   "Todos los públicos",
		domain.RatingTeen:      "Juvenil 13+",
		domain.RatingMature18:  "Maduro +18",
		domain.RatingExplicit21: "Explícito +21 / R-18",
	}

	for rating, label := range expected {
		name := domain.ContentRatingLabel(rating)
		if name != label {
			t.Errorf("expected label %q for rating %q, got %q", label, rating, name)
		}
	}
}

func TestDefaultGenreCatalog(t *testing.T) {
	catalog := domain.DefaultGenreCatalog()
	if len(catalog) < 18 {
		t.Fatalf("expected at least 18 genre definitions in catalog, got %d", len(catalog))
	}

	requiredIDs := []string{
		// Standard
		"high_fantasy",
		"dark_fantasy",
		"cyberpunk",
		"space_opera",
		"mystery_detective",
		"psychological_thriller",
		"cosmic_horror",
		"slice_of_life",
		"urban_fantasy",
		"contemporary_romance",
		// Japanese Light Novel & Webnovel Tropes
		"isekai_litrpg",
		"xianxia_cultivation",
		"otome_villainess",
		"yandere_obsession",
		// Explicit & Japanese R-18 Tropes
		"smut_explicit",
		"isekai_harem_r18",
		"monster_girls_r18",
		"femdom_r18",
		"dark_romance_taboo",
		"magical_corruption_r18",
		"netori_ntr_drama",
		"bdsm_power_dynamics",
	}

	catalogMap := make(map[string]domain.GenreDefinition)
	for _, g := range catalog {
		if g.ID == "" {
			t.Errorf("genre ID must not be empty: %+v", g)
		}
		if g.Name == "" {
			t.Errorf("genre Name must not be empty for ID %s", g.ID)
		}
		if g.Category == "" {
			t.Errorf("genre Category must not be empty for ID %s", g.ID)
		}
		if g.Description == "" {
			t.Errorf("genre Description must not be empty for ID %s", g.ID)
		}
		if g.EditorDirective == "" {
			t.Errorf("genre EditorDirective must not be empty for ID %s", g.ID)
		}
		catalogMap[g.ID] = g
	}

	for _, id := range requiredIDs {
		if _, ok := catalogMap[id]; !ok {
			t.Errorf("missing required genre ID: %s", id)
		}
	}
}

func TestGetGenreByID(t *testing.T) {
	genre, ok := domain.GetGenreByID("yandere_obsession")
	if !ok {
		t.Fatalf("expected to find yandere_obsession genre")
	}
	if genre.ID != "yandere_obsession" {
		t.Errorf("expected ID yandere_obsession, got %s", genre.ID)
	}

	_, notFound := domain.GetGenreByID("non_existent_genre_xyz")
	if notFound {
		t.Errorf("expected false for non-existent genre")
	}
}

func TestComposeEditorPrompt_Layers(t *testing.T) {
	// 1. Base persona layer
	promptBase := domain.ComposeEditorPrompt(domain.RatingAllAges, nil, "")
	if !strings.Contains(strings.ToLower(promptBase), "15+") || !strings.Contains(strings.ToLower(promptBase), "editor") {
		t.Errorf("expected senior editor persona in base prompt, got:\n%s", promptBase)
	}
	if !strings.Contains(strings.ToLower(promptBase), "todos los públicos") {
		t.Errorf("expected AllAges rating directive in prompt, got:\n%s", promptBase)
	}

	// 2. Explicit +21 R-18 Rating Layer
	promptR18 := domain.ComposeEditorPrompt(domain.RatingExplicit21, []string{"smut_explicit", "isekai_harem_r18"}, "Mantén el ritmo ágil.")
	if !strings.Contains(promptR18, "R-18") && !strings.Contains(promptR18, "+21") {
		t.Errorf("expected explicit rating directive, got:\n%s", promptR18)
	}
	if !strings.Contains(strings.ToLower(promptR18), "sin censura") && !strings.Contains(strings.ToLower(promptR18), "cero censura") {
		t.Errorf("expected zero censorship directive in +21 R-18 prompt, got:\n%s", promptR18)
	}

	// Check matched genre directives
	genreSmut, _ := domain.GetGenreByID("smut_explicit")
	if !strings.Contains(promptR18, genreSmut.Name) {
		t.Errorf("expected genre name %q in composed prompt", genreSmut.Name)
	}

	genreHarem, _ := domain.GetGenreByID("isekai_harem_r18")
	if !strings.Contains(promptR18, genreHarem.Name) {
		t.Errorf("expected genre name %q in composed prompt", genreHarem.Name)
	}

	// 3. Mature +18 and Teen ratings
	promptMature := domain.ComposeEditorPrompt(domain.RatingMature18, []string{"dark_fantasy"}, "")
	if !strings.Contains(promptMature, "MADURO +18") {
		t.Errorf("expected mature 18 rating directive, got:\n%s", promptMature)
	}

	promptTeen := domain.ComposeEditorPrompt(domain.RatingTeen, []string{"isekai_litrpg"}, "")
	if !strings.Contains(promptTeen, "JUVENIL 13+") {
		t.Errorf("expected teen rating directive, got:\n%s", promptTeen)
	}

	// 4. Unknown genre IDs should be ignored gracefully
	promptUnknown := domain.ComposeEditorPrompt(domain.RatingTeen, []string{"unknown_xyz_123"}, "")
	if strings.Contains(promptUnknown, "unknown_xyz_123") {
		t.Errorf("unknown genre ID should not leak as valid genre block")
	}
}

func TestNovelSettingsDefaults(t *testing.T) {
	settings := domain.DefaultNovelSettings()
	if settings.Rating != domain.RatingTeen {
		t.Errorf("expected default rating %q, got %q", domain.RatingTeen, settings.Rating)
	}
	if len(settings.Genres) != 0 {
		t.Errorf("expected empty default genres, got %v", settings.Genres)
	}
}
