package domain

import (
	"context"
	"time"
)

// BrainFactType defines the category of a recorded fact.
type BrainFactType string

const (
	FactTypeCharacter BrainFactType = "character"
	FactTypeLore      BrainFactType = "lore"
	FactTypeSetting   BrainFactType = "setting"
	FactTypePlot      BrainFactType = "plot"
	FactTypeDecision  BrainFactType = "decision"
	FactTypeGeneral   BrainFactType = "general"
)

// BrainFact represents an atomic unit of knowledge or narrative continuity.
type BrainFact struct {
	ID        string        `json:"id"`
	Topic     string        `json:"topic"`     // e.g. "Personajes", "Geografía", "Magia", "Trama"
	Concept   string        `json:"concept"`   // e.g. "Kuno", "Espada del Sol", "Valle de Piedra"
	Fact      string        `json:"fact"`      // e.g. "Kuno perdió el brazo izquierdo en el capítulo 3"
	Type      BrainFactType `json:"type"`      // character, lore, setting, plot, decision
	ChapterID string        `json:"chapter_id,omitempty"`
	Tags      []string      `json:"tags,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// SessionSummary represents a high-level summary of an authoring or planning session.
type SessionSummary struct {
	ID         string    `json:"id"`
	Summary    string    `json:"summary"`
	Highlights []string  `json:"highlights,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// TimelineEvent represents a plot point or historical event in the novel's chronology.
type TimelineEvent struct {
	ID                 string    `json:"id"`
	ChronologicalOrder int       `json:"chronological_order"` // 1, 2, 3...
	Period             string    `json:"period"`              // e.g. "Era Antigua", "Año 452", "Capítulo 1", "Flashback"
	Title              string    `json:"title"`               // e.g. "Forja de la Espada del Alba"
	Description        string    `json:"description"`         // Concise event summary
	Characters         []string  `json:"characters,omitempty"`
	ChapterID          string    `json:"chapter_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// BrainRepository defines data persistence and search contracts for the Brain memory system.
type BrainRepository interface {
	SaveFact(ctx context.Context, fact BrainFact) error
	SaveFacts(ctx context.Context, facts []BrainFact) error
	DeleteFact(ctx context.Context, id string) error
	SearchFacts(ctx context.Context, query string, limit int) ([]BrainFact, error)
	ListFactsByTopic(ctx context.Context, topic string) ([]BrainFact, error)
	ListFactsByType(ctx context.Context, factType BrainFactType) ([]BrainFact, error)
	ListRecentFacts(ctx context.Context, limit int) ([]BrainFact, error)
	SaveSessionSummary(ctx context.Context, summary SessionSummary) error
	ListSessionSummaries(ctx context.Context, limit int) ([]SessionSummary, error)
	SaveTimelineEvent(ctx context.Context, event TimelineEvent) error
	SaveTimelineEvents(ctx context.Context, events []TimelineEvent) error
	ListTimelineEvents(ctx context.Context) ([]TimelineEvent, error)
	DeleteTimelineEvent(ctx context.Context, id string) error
	Close() error
}

// BrainActivityEvent represents a notification emitted when Brain extracts or accesses memory.
type BrainActivityEvent struct {
	Type        string    `json:"type"` // "saved", "retrieved", "summary"
	FactCount   int       `json:"fact_count"`
	Description string    `json:"description"` // e.g. "🧠 Memorizado: 2 hechos nuevos"
	Timestamp   time.Time `json:"timestamp"`
}
