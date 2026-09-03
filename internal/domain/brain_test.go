package domain_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

func TestBrainFactJSONSerialization(t *testing.T) {
	fact := domain.BrainFact{
		ID:        "fact-1",
		Topic:     "Personajes",
		Concept:   "Kuno",
		Fact:      "Perdió el brazo en la batalla del norte",
		Type:      domain.FactTypeCharacter,
		ChapterID: "01",
		Tags:      []string{"herida", "batalla"},
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(fact)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var parsed domain.BrainFact
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if parsed.ID != fact.ID || parsed.Concept != fact.Concept || parsed.Type != fact.Type {
		t.Errorf("expected fact match, got %+v", parsed)
	}
}

func TestTimelineEventJSONSerialization(t *testing.T) {
	now := time.Now()
	event := domain.TimelineEvent{
		ID:                 "tl-1",
		ChronologicalOrder: 1,
		Period:             "Era Antigua",
		Title:              "Forja de la Espada del Alba",
		Description:        "Los primeros artesanos forjan la espada con fuego sagrado",
		Characters:         []string{"Aurelio", "Kuno"},
		ChapterID:          "cap-0",
		CreatedAt:          now,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	var parsed domain.TimelineEvent
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if parsed.ID != event.ID || parsed.ChronologicalOrder != 1 || parsed.Period != "Era Antigua" {
		t.Errorf("expected event match, got %+v", parsed)
	}
	if len(parsed.Characters) != 2 || parsed.Characters[0] != "Aurelio" {
		t.Errorf("expected characters match, got %+v", parsed.Characters)
	}

	// Triangulation: minimal event without optional fields
	minEvent := domain.TimelineEvent{
		Title:       "Evento sin detalles",
		Description: "Breve descripción",
	}
	minData, err := json.Marshal(minEvent)
	if err != nil {
		t.Fatalf("unexpected marshal error on minimal event: %v", err)
	}
	var minParsed domain.TimelineEvent
	if err := json.Unmarshal(minData, &minParsed); err != nil {
		t.Fatalf("unexpected unmarshal error on minimal event: %v", err)
	}
	if minParsed.Title != "Evento sin detalles" || len(minParsed.Characters) != 0 {
		t.Errorf("expected minimal event match, got %+v", minParsed)
	}
}

