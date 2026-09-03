package repository_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/repository"
)

func TestSQLiteBrainRepository_Lifecycle(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "brain.db")

	repo, err := repository.NewSQLiteBrainRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to init repository: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()

	// 1. Save individual fact
	fact1 := domain.BrainFact{
		ID:        "f1",
		Topic:     "Personajes",
		Concept:   "Kuno",
		Fact:      "Porta la Espada del Alba",
		Type:      domain.FactTypeCharacter,
		ChapterID: "01",
		Tags:      []string{"arma", "protaganista"},
		CreatedAt: time.Now(),
	}

	if err := repo.SaveFact(ctx, fact1); err != nil {
		t.Fatalf("SaveFact failed: %v", err)
	}

	// 2. Save batch facts
	facts := []domain.BrainFact{
		{
			ID:        "f2",
			Topic:     "Geografía",
			Concept:   "Valle de Piedra",
			Fact:      "Región montañosa al norte del reino",
			Type:      domain.FactTypeSetting,
			ChapterID: "01",
			Tags:      []string{"montaña", "norte"},
			CreatedAt: time.Now(),
		},
		{
			ID:        "f3",
			Topic:     "Trama",
			Concept:   "La Profecía",
			Fact:      "Anuncia el regreso de la sombra en el solsticio",
			Type:      domain.FactTypePlot,
			ChapterID: "02",
			Tags:      []string{"lore", "profecía"},
			CreatedAt: time.Now(),
		},
	}

	if err := repo.SaveFacts(ctx, facts); err != nil {
		t.Fatalf("SaveFacts failed: %v", err)
	}

	// 3. Search via FTS5
	searchResults, err := repo.SearchFacts(ctx, "Espada", 10)
	if err != nil {
		t.Fatalf("SearchFacts failed: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].Concept != "Kuno" {
		t.Errorf("expected 1 result for 'Espada', got %d: %+v", len(searchResults), searchResults)
	}

	// Search setting
	settingResults, err := repo.SearchFacts(ctx, "montañosa", 10)
	if err != nil {
		t.Fatalf("SearchFacts failed for setting: %v", err)
	}
	if len(settingResults) != 1 || settingResults[0].Concept != "Valle de Piedra" {
		t.Errorf("expected Valle de Piedra, got %+v", settingResults)
	}

	// 4. List by Topic
	topicResults, err := repo.ListFactsByTopic(ctx, "Personajes")
	if err != nil {
		t.Fatalf("ListFactsByTopic failed: %v", err)
	}
	if len(topicResults) != 1 || topicResults[0].ID != "f1" {
		t.Errorf("expected 1 result for topic Personajes, got %d", len(topicResults))
	}

	// 5. List by Type
	typeResults, err := repo.ListFactsByType(ctx, domain.FactTypePlot)
	if err != nil {
		t.Fatalf("ListFactsByType failed: %v", err)
	}
	if len(typeResults) != 1 || typeResults[0].ID != "f3" {
		t.Errorf("expected 1 result for FactTypePlot, got %d", len(typeResults))
	}

	// 6. List Recent Facts
	recent, err := repo.ListRecentFacts(ctx, 10)
	if err != nil {
		t.Fatalf("ListRecentFacts failed: %v", err)
	}
	if len(recent) != 3 {
		t.Errorf("expected 3 recent facts, got %d", len(recent))
	}

	// 7. Update fact (upsert)
	fact1Updated := fact1
	fact1Updated.Fact = "Porta la Espada del Alba y la Daga de las Sombras"
	if err := repo.SaveFact(ctx, fact1Updated); err != nil {
		t.Fatalf("SaveFact update failed: %v", err)
	}

	updatedSearch, err := repo.SearchFacts(ctx, "Sombras", 10)
	if err != nil {
		t.Fatalf("SearchFacts failed after update: %v", err)
	}
	if len(updatedSearch) == 0 || updatedSearch[0].ID != "f1" {
		t.Errorf("expected updated fact f1, got %+v", updatedSearch)
	}

	// 8. Delete fact
	if err := repo.DeleteFact(ctx, "f3"); err != nil {
		t.Fatalf("DeleteFact failed: %v", err)
	}

	afterDelete, err := repo.SearchFacts(ctx, "Profecía", 10)
	if err != nil {
		t.Fatalf("SearchFacts failed after delete: %v", err)
	}
	if len(afterDelete) != 0 {
		t.Errorf("expected 0 results after delete, got %d", len(afterDelete))
	}

	// 9. Session Summaries
	summary := domain.SessionSummary{
		ID:         "s1",
		Summary:    "Sesión de escritura del capítulo 2 finalizada",
		Highlights: []string{"Kuno llegó al Valle", "Batalla con los lobos"},
		CreatedAt:  time.Now(),
	}

	if err := repo.SaveSessionSummary(ctx, summary); err != nil {
		t.Fatalf("SaveSessionSummary failed: %v", err)
	}

	summaries, err := repo.ListSessionSummaries(ctx, 5)
	if err != nil {
		t.Fatalf("ListSessionSummaries failed: %v", err)
	}
	if len(summaries) != 1 || summaries[0].ID != "s1" || len(summaries[0].Highlights) != 2 {
		t.Errorf("expected 1 summary with 2 highlights, got %+v", summaries)
	}

	// 10. Timeline Events
	event1 := domain.TimelineEvent{
		ID:                 "tl-1",
		ChronologicalOrder: 2,
		Period:             "Capítulo 1",
		Title:              "Llegada al Valle",
		Description:        "Kuno llega al valle de piedra en busca del sabio",
		Characters:         []string{"Kuno"},
		ChapterID:          "01",
		CreatedAt:          time.Now(),
	}

	if err := repo.SaveTimelineEvent(ctx, event1); err != nil {
		t.Fatalf("SaveTimelineEvent failed: %v", err)
	}

	batchEvents := []domain.TimelineEvent{
		{
			ID:                 "tl-0",
			ChronologicalOrder: 1,
			Period:             "Era Antigua",
			Title:              "Forja de la Espada",
			Description:        "Creación del arma legendaria",
			Characters:         []string{"Aurelio"},
			ChapterID:          "00",
			CreatedAt:          time.Now().Add(-1 * time.Hour),
		},
		{
			ID:                 "tl-2",
			ChronologicalOrder: 3,
			Period:             "Capítulo 2",
			Title:              "Enfrentamiento en las Colinas",
			Description:        "Batalla contra las criaturas oscuras",
			Characters:         []string{"Kuno", "Elena"},
			ChapterID:          "02",
			CreatedAt:          time.Now(),
		},
	}

	if err := repo.SaveTimelineEvents(ctx, batchEvents); err != nil {
		t.Fatalf("SaveTimelineEvents failed: %v", err)
	}

	// List ordered by ChronologicalOrder ASC
	events, err := repo.ListTimelineEvents(ctx)
	if err != nil {
		t.Fatalf("ListTimelineEvents failed: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 timeline events, got %d", len(events))
	}
	if events[0].ID != "tl-0" || events[1].ID != "tl-1" || events[2].ID != "tl-2" {
		t.Errorf("expected events ordered by chronological order [tl-0, tl-1, tl-2], got [%s, %s, %s]",
			events[0].ID, events[1].ID, events[2].ID)
	}
	if len(events[2].Characters) != 2 || events[2].Characters[1] != "Elena" {
		t.Errorf("expected characters in event 2, got %+v", events[2].Characters)
	}

	// Update timeline event (upsert)
	event1Updated := event1
	event1Updated.Title = "Llegada al Valle y Encuentro con Elena"
	event1Updated.Characters = []string{"Kuno", "Elena"}
	if err := repo.SaveTimelineEvent(ctx, event1Updated); err != nil {
		t.Fatalf("SaveTimelineEvent update failed: %v", err)
	}

	eventsAfterUpdate, err := repo.ListTimelineEvents(ctx)
	if err != nil {
		t.Fatalf("ListTimelineEvents after update failed: %v", err)
	}
	if len(eventsAfterUpdate) != 3 {
		t.Errorf("expected 3 events after update, got %d", len(eventsAfterUpdate))
	}
	if eventsAfterUpdate[1].Title != "Llegada al Valle y Encuentro con Elena" {
		t.Errorf("expected updated title, got %s", eventsAfterUpdate[1].Title)
	}

	// Delete timeline event
	if err := repo.DeleteTimelineEvent(ctx, "tl-0"); err != nil {
		t.Fatalf("DeleteTimelineEvent failed: %v", err)
	}

	eventsAfterDelete, err := repo.ListTimelineEvents(ctx)
	if err != nil {
		t.Fatalf("ListTimelineEvents after delete failed: %v", err)
	}
	if len(eventsAfterDelete) != 2 {
		t.Fatalf("expected 2 events after delete, got %d", len(eventsAfterDelete))
	}
	// Triangulation: Save empty slice returns nil
	if err := repo.SaveTimelineEvents(ctx, []domain.TimelineEvent{}); err != nil {
		t.Errorf("expected nil on saving empty events slice, got: %v", err)
	}

	// Triangulation: Auto-generated ID and CreatedAt when omitted
	autoEvent := domain.TimelineEvent{
		ChronologicalOrder: 99,
		Title:              "Evento Futuro",
		Description:        "Profecía no cumplida",
	}
	if err := repo.SaveTimelineEvent(ctx, autoEvent); err != nil {
		t.Fatalf("SaveTimelineEvent auto failed: %v", err)
	}
	allAfterAuto, err := repo.ListTimelineEvents(ctx)
	if err != nil {
		t.Fatalf("ListTimelineEvents failed: %v", err)
	}
	foundAuto := false
	for _, ev := range allAfterAuto {
		if ev.Title == "Evento Futuro" {
			foundAuto = true
			if ev.ID == "" {
				t.Errorf("expected generated ID, got empty")
			}
			if ev.CreatedAt.IsZero() {
				t.Errorf("expected non-zero CreatedAt")
			}
		}
	}
	if !foundAuto {
		t.Errorf("expected to find autoEvent in repository")
	}
}

