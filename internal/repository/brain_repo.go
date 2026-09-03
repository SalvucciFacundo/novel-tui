package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteBrainRepository implements domain.BrainRepository using SQLite and FTS5.
type SQLiteBrainRepository struct {
	db     *sql.DB
	dbPath string
}

// NewSQLiteBrainRepository creates and migrates a new SQLite brain repository.
func NewSQLiteBrainRepository(dbPath string) (*SQLiteBrainRepository, error) {
	dbPath = ExpandHome(dbPath)
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for brain db: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite brain db: %w", err)
	}

	// Performance and reliability pragmas
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to set pragma %s: %w", p, err)
		}
	}

	repo := &SQLiteBrainRepository{
		db:     db,
		dbPath: dbPath,
	}

	if err := repo.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("brain db migration failed: %w", err)
	}

	return repo, nil
}

func (r *SQLiteBrainRepository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS brain_facts (
		id TEXT PRIMARY KEY,
		topic TEXT NOT NULL,
		concept TEXT NOT NULL,
		fact TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'general',
		chapter_id TEXT NOT NULL DEFAULT '',
		tags TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_brain_facts_topic ON brain_facts(topic);
	CREATE INDEX IF NOT EXISTS idx_brain_facts_concept ON brain_facts(concept);
	CREATE INDEX IF NOT EXISTS idx_brain_facts_type ON brain_facts(type);
	CREATE INDEX IF NOT EXISTS idx_brain_facts_created_at ON brain_facts(created_at);

	CREATE VIRTUAL TABLE IF NOT EXISTS brain_facts_fts USING fts5(
		topic, concept, fact, tags,
		content='brain_facts',
		content_rowid='rowid'
	);

	CREATE TRIGGER IF NOT EXISTS trg_brain_facts_ai AFTER INSERT ON brain_facts BEGIN
		INSERT INTO brain_facts_fts(rowid, topic, concept, fact, tags)
		VALUES (new.rowid, new.topic, new.concept, new.fact, new.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS trg_brain_facts_ad AFTER DELETE ON brain_facts BEGIN
		INSERT INTO brain_facts_fts(brain_facts_fts, rowid, topic, concept, fact, tags)
		VALUES('delete', old.rowid, old.topic, old.concept, old.fact, old.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS trg_brain_facts_au AFTER UPDATE ON brain_facts BEGIN
		INSERT INTO brain_facts_fts(brain_facts_fts, rowid, topic, concept, fact, tags)
		VALUES('delete', old.rowid, old.topic, old.concept, old.fact, old.tags);
		INSERT INTO brain_facts_fts(rowid, topic, concept, fact, tags)
		VALUES (new.rowid, new.topic, new.concept, new.fact, new.tags);
	END;

	CREATE TABLE IF NOT EXISTS session_summaries (
		id TEXT PRIMARY KEY,
		summary TEXT NOT NULL,
		highlights TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_session_summaries_created_at ON session_summaries(created_at);

	CREATE TABLE IF NOT EXISTS brain_timeline_events (
		id TEXT PRIMARY KEY,
		chronological_order INTEGER NOT NULL DEFAULT 0,
		period TEXT NOT NULL DEFAULT '',
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		characters TEXT NOT NULL DEFAULT '[]',
		chapter_id TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_timeline_order ON brain_timeline_events(chronological_order, created_at);
	`

	_, err := r.db.Exec(schema)
	return err
}

// SaveFact inserts or replaces a single fact.
func (r *SQLiteBrainRepository) SaveFact(ctx context.Context, fact domain.BrainFact) error {
	return r.SaveFacts(ctx, []domain.BrainFact{fact})
}

// SaveFacts inserts or replaces multiple facts atomically.
func (r *SQLiteBrainRepository) SaveFacts(ctx context.Context, facts []domain.BrainFact) error {
	if len(facts) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO brain_facts (id, topic, concept, fact, type, chapter_id, tags, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			topic = excluded.topic,
			concept = excluded.concept,
			fact = excluded.fact,
			type = excluded.type,
			chapter_id = excluded.chapter_id,
			tags = excluded.tags,
			created_at = excluded.created_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert fact statement: %w", err)
	}
	defer stmt.Close()

	for _, f := range facts {
		if f.ID == "" {
			f.ID = fmt.Sprintf("fact_%d_%s", time.Now().UnixNano(), sanitizeKey(f.Concept))
		}
		if f.CreatedAt.IsZero() {
			f.CreatedAt = time.Now()
		}
		if f.Type == "" {
			f.Type = domain.FactTypeGeneral
		}

		tagsJSON, err := json.Marshal(f.Tags)
		if err != nil {
			tagsJSON = []byte("[]")
		}

		_, err = stmt.ExecContext(ctx,
			f.ID,
			f.Topic,
			f.Concept,
			f.Fact,
			string(f.Type),
			f.ChapterID,
			string(tagsJSON),
			f.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to execute insert fact: %w", err)
		}
	}

	return tx.Commit()
}

// DeleteFact removes a fact by ID.
func (r *SQLiteBrainRepository) DeleteFact(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM brain_facts WHERE id = ?", id)
	return err
}

// SearchFacts queries FTS5 index for relevant facts, falling back to LIKE if FTS syntax fails.
func (r *SQLiteBrainRepository) SearchFacts(ctx context.Context, query string, limit int) ([]domain.BrainFact, error) {
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return r.ListRecentFacts(ctx, limit)
	}
	if limit <= 0 {
		limit = 10
	}

	// Prepare FTS5 query terms (sanitize special chars for token matching)
	ftsQuery := sanitizeFTS5Query(cleanQuery)

	ftsSQL := `
		SELECT b.id, b.topic, b.concept, b.fact, b.type, b.chapter_id, b.tags, b.created_at
		FROM brain_facts b
		JOIN brain_facts_fts fts ON b.rowid = fts.rowid
		WHERE brain_facts_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, ftsSQL, ftsQuery, limit)
	if err != nil {
		// Fallback to LIKE query if FTS query syntax error occurs
		return r.searchFallbackLike(ctx, cleanQuery, limit)
	}
	defer rows.Close()

	facts, err := scanFacts(rows)
	if err != nil {
		return nil, err
	}

	if len(facts) == 0 {
		// Fallback to substring search if exact token matching gave 0 results
		return r.searchFallbackLike(ctx, cleanQuery, limit)
	}

	return facts, nil
}

func (r *SQLiteBrainRepository) searchFallbackLike(ctx context.Context, query string, limit int) ([]domain.BrainFact, error) {
	likeSQL := `
		SELECT id, topic, concept, fact, type, chapter_id, tags, created_at
		FROM brain_facts
		WHERE concept LIKE ? OR fact LIKE ? OR topic LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, likeSQL, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("fallback search failed: %w", err)
	}
	defer rows.Close()

	return scanFacts(rows)
}

// ListFactsByTopic retrieves all facts for a given topic.
func (r *SQLiteBrainRepository) ListFactsByTopic(ctx context.Context, topic string) ([]domain.BrainFact, error) {
	sqlStr := `
		SELECT id, topic, concept, fact, type, chapter_id, tags, created_at
		FROM brain_facts
		WHERE topic = ?
		ORDER BY concept ASC, created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, sqlStr, topic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFacts(rows)
}

// ListFactsByType retrieves all facts of a specific type (e.g. character, lore).
func (r *SQLiteBrainRepository) ListFactsByType(ctx context.Context, factType domain.BrainFactType) ([]domain.BrainFact, error) {
	sqlStr := `
		SELECT id, topic, concept, fact, type, chapter_id, tags, created_at
		FROM brain_facts
		WHERE type = ?
		ORDER BY topic ASC, concept ASC
	`
	rows, err := r.db.QueryContext(ctx, sqlStr, string(factType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFacts(rows)
}

// ListRecentFacts retrieves the latest recorded facts.
func (r *SQLiteBrainRepository) ListRecentFacts(ctx context.Context, limit int) ([]domain.BrainFact, error) {
	if limit <= 0 {
		limit = 20
	}
	sqlStr := `
		SELECT id, topic, concept, fact, type, chapter_id, tags, created_at
		FROM brain_facts
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, sqlStr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFacts(rows)
}

// SaveSessionSummary records a new session summary.
func (r *SQLiteBrainRepository) SaveSessionSummary(ctx context.Context, summary domain.SessionSummary) error {
	if summary.ID == "" {
		summary.ID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now()
	}

	hlJSON, err := json.Marshal(summary.Highlights)
	if err != nil {
		hlJSON = []byte("[]")
	}

	sqlStr := `
		INSERT INTO session_summaries (id, summary, highlights, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			summary = excluded.summary,
			highlights = excluded.highlights,
			created_at = excluded.created_at
	`
	_, err = r.db.ExecContext(ctx, sqlStr,
		summary.ID,
		summary.Summary,
		string(hlJSON),
		summary.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// ListSessionSummaries returns recent session summaries.
func (r *SQLiteBrainRepository) ListSessionSummaries(ctx context.Context, limit int) ([]domain.SessionSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	sqlStr := `
		SELECT id, summary, highlights, created_at
		FROM session_summaries
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, sqlStr, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.SessionSummary
	for rows.Next() {
		var s domain.SessionSummary
		var hlStr, createdStr string

		if err := rows.Scan(&s.ID, &s.Summary, &hlStr, &createdStr); err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(hlStr), &s.Highlights)
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		summaries = append(summaries, s)
	}

	return summaries, rows.Err()
}

// SaveTimelineEvent inserts or updates a single timeline event.
func (r *SQLiteBrainRepository) SaveTimelineEvent(ctx context.Context, event domain.TimelineEvent) error {
	return r.SaveTimelineEvents(ctx, []domain.TimelineEvent{event})
}

// SaveTimelineEvents inserts or updates multiple timeline events atomically.
func (r *SQLiteBrainRepository) SaveTimelineEvents(ctx context.Context, events []domain.TimelineEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO brain_timeline_events (id, chronological_order, period, title, description, characters, chapter_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			chronological_order = excluded.chronological_order,
			period = excluded.period,
			title = excluded.title,
			description = excluded.description,
			characters = excluded.characters,
			chapter_id = excluded.chapter_id,
			created_at = excluded.created_at
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert timeline event statement: %w", err)
	}
	defer stmt.Close()

	for _, e := range events {
		if e.ID == "" {
			e.ID = fmt.Sprintf("tl_%d_%s", time.Now().UnixNano(), sanitizeKey(e.Title))
		}
		if e.CreatedAt.IsZero() {
			e.CreatedAt = time.Now()
		}

		charsJSON, err := json.Marshal(e.Characters)
		if err != nil {
			charsJSON = []byte("[]")
		}

		_, err = stmt.ExecContext(ctx,
			e.ID,
			e.ChronologicalOrder,
			e.Period,
			e.Title,
			e.Description,
			string(charsJSON),
			e.ChapterID,
			e.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to execute insert timeline event: %w", err)
		}
	}

	return tx.Commit()
}

// ListTimelineEvents returns all timeline events ordered by chronological order ascending and created_at ascending.
func (r *SQLiteBrainRepository) ListTimelineEvents(ctx context.Context) ([]domain.TimelineEvent, error) {
	sqlStr := `
		SELECT id, chronological_order, period, title, description, characters, chapter_id, created_at
		FROM brain_timeline_events
		ORDER BY chronological_order ASC, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, sqlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeline events: %w", err)
	}
	defer rows.Close()

	var events []domain.TimelineEvent
	for rows.Next() {
		var e domain.TimelineEvent
		var charsStr, createdStr string

		if err := rows.Scan(
			&e.ID,
			&e.ChronologicalOrder,
			&e.Period,
			&e.Title,
			&e.Description,
			&charsStr,
			&e.ChapterID,
			&createdStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan timeline event row: %w", err)
		}

		_ = json.Unmarshal([]byte(charsStr), &e.Characters)
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		events = append(events, e)
	}

	return events, rows.Err()
}

// DeleteTimelineEvent removes a timeline event by ID.
func (r *SQLiteBrainRepository) DeleteTimelineEvent(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM brain_timeline_events WHERE id = ?", id)
	return err
}

// Close terminates the underlying database connection.
func (r *SQLiteBrainRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func scanFacts(rows *sql.Rows) ([]domain.BrainFact, error) {
	var facts []domain.BrainFact
	for rows.Next() {
		var f domain.BrainFact
		var typeStr, tagsStr, createdStr string

		if err := rows.Scan(
			&f.ID,
			&f.Topic,
			&f.Concept,
			&f.Fact,
			&typeStr,
			&f.ChapterID,
			&tagsStr,
			&createdStr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan fact row: %w", err)
		}

		f.Type = domain.BrainFactType(typeStr)
		_ = json.Unmarshal([]byte(tagsStr), &f.Tags)
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
		facts = append(facts, f)
	}
	return facts, rows.Err()
}

func sanitizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "_")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "entity"
	}
	return b.String()
}

func sanitizeFTS5Query(q string) string {
	words := strings.Fields(q)
	var terms []string
	for _, w := range words {
		// Strip quotes or special fts5 punctuation
		cleaned := strings.Map(func(r rune) rune {
			if strings.ContainsRune(`"':*^(){}+-~`, r) {
				return -1
			}
			return r
		}, w)
		cleaned = strings.TrimSpace(cleaned)
		if len(cleaned) > 1 {
			terms = append(terms, fmt.Sprintf(`"%s"`, cleaned))
		}
	}
	if len(terms) == 0 {
		return `""`
	}
	return strings.Join(terms, " OR ")
}
