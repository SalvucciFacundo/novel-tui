package llm_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
	"github.com/SalvucciFacundo/novel-tui/internal/service/llm"
)

func TestOllamaProvider_StreamChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}

		chunks := []string{
			`{"model":"qwen2.5:7b","message":{"role":"assistant","content":"Hola"},"done":false}`,
			`{"model":"qwen2.5:7b","message":{"role":"assistant","content":" mundo"},"done":false}`,
			`{"model":"qwen2.5:7b","message":{"role":"assistant","content":"!"},"done":true}`,
		}

		for _, chunk := range chunks {
			fmt.Fprintln(w, chunk)
			flusher.Flush()
		}
	}))
	defer server.Close()

	provider := llm.NewOllamaProvider(server.URL, server.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, err := provider.StreamChat(ctx, domain.ChatRequest{
		Model: "qwen2.5:7b",
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Di hola"},
		},
		Temperature: 0.7,
	})
	if err != nil {
		t.Fatalf("unexpected error starting stream: %v", err)
	}

	var collected string
	var doneReceived bool
	for chunk := range stream {
		if chunk.Err != nil {
			t.Fatalf("unexpected chunk error: %v", chunk.Err)
		}
		collected += chunk.Content
		if chunk.Done {
			doneReceived = true
		}
	}

	if collected != "Hola mundo!" {
		t.Errorf("expected collected 'Hola mundo!', got '%s'", collected)
	}
	if !doneReceived {
		t.Errorf("expected done flag received")
	}
}

func TestOllamaProvider_StreamChat_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := llm.NewOllamaProvider(server.URL, server.Client())
	ctx := context.Background()

	_, err := provider.StreamChat(ctx, domain.ChatRequest{
		Model: "qwen2.5:7b",
	})
	if err == nil {
		t.Fatalf("expected error for 500 status code, got nil")
	}
}

func TestOllamaProvider_StreamChat_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		flusher, _ := w.(http.Flusher)
		fmt.Fprintln(w, `{"model":"m","message":{"role":"assistant","content":"part 1"},"done":false}`)
		flusher.Flush()
		time.Sleep(500 * time.Millisecond)
		fmt.Fprintln(w, `{"model":"m","message":{"role":"assistant","content":"part 2"},"done":true}`)
		flusher.Flush()
	}))
	defer server.Close()

	provider := llm.NewOllamaProvider(server.URL, server.Client())
	ctx, cancel := context.WithCancel(context.Background())

	stream, err := provider.StreamChat(ctx, domain.ChatRequest{Model: "m"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Read first chunk
	chunk := <-stream
	if chunk.Content != "part 1" {
		t.Errorf("expected 'part 1', got '%s'", chunk.Content)
	}

	// Cancel context
	cancel()

	// Drain remaining stream; should close quickly
	for range stream {
	}
}

func TestOpenAICompatibleProvider_StreamChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Buenos"}}]}`,
			`data: {"choices":[{"delta":{"content":" días"}}]}`,
			`data: {"choices":[{"delta":{"content":"!"},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			fmt.Fprintf(w, "%s\n\n", chunk)
			flusher.Flush()
		}
	}))
	defer server.Close()

	provider := llm.NewOpenAICompatibleProvider(server.URL, "test-key", server.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, err := provider.StreamChat(ctx, domain.ChatRequest{
		Model: "gpt-4o",
		Messages: []domain.ChatMessage{
			{Role: "user", Content: "Saluda"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var collected string
	var isDone bool
	for chunk := range stream {
		if chunk.Err != nil {
			t.Fatalf("unexpected error in chunk: %v", chunk.Err)
		}
		collected += chunk.Content
		if chunk.Done {
			isDone = true
		}
	}

	if collected != "Buenos días!" {
		t.Errorf("expected 'Buenos días!', got '%s'", collected)
	}
	if !isDone {
		t.Errorf("expected done to be true")
	}
}

func TestOpenAICompatibleProvider_StreamChat_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_api_key"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := llm.NewOpenAICompatibleProvider(server.URL, "bad-key", server.Client())
	_, err := provider.StreamChat(context.Background(), domain.ChatRequest{Model: "gpt-4o"})
	if err == nil {
		t.Fatalf("expected error for 401 response, got nil")
	}
}

func TestContextBuilder(t *testing.T) {
	tempDir := t.TempDir()

	// Create personajes.json
	chars := []domain.Character{
		{
			Name:        "Elena",
			Role:        "Protagonista",
			Description: "Maga de fuego",
			Notes:       "Odia los secretos",
		},
		{
			Name:        "Mateo",
			Role:        "Antagonista",
			Description: "Caballero traidor",
		},
	}
	charData, _ := json.Marshal(chars)
	_ = os.WriteFile(filepath.Join(tempDir, "personajes.json"), charData, 0644)

	// Create notas.txt
	notes := "El giro argumental ocurre cuando Elena descubre la espada antigua."
	_ = os.WriteFile(filepath.Join(tempDir, "notas.txt"), []byte(notes), 0644)

	builder := llm.NewContextBuilder()

	// Test full compilation with High effort
	ctxPrompt := builder.BuildContext(llm.ContextParams{
		NovelDir:           tempDir,
		Genre:              "Fantasía Oscura",
		GenrePrompt:        "Eres un asistente para fantasía oscura.",
		ActiveChapterTitle: "02_la_emboscada.txt",
		ActiveChapterText:  "La noche era fría y la niebla cubría el sendero.",
		EffortLevel:        domain.EffortHigh,
	})

	if !strings.Contains(ctxPrompt, "Eres un asistente para fantasía oscura.") {
		t.Errorf("expected genre prompt in context")
	}
	if !strings.Contains(ctxPrompt, "Actúa como editor literario senior") {
		t.Errorf("expected high effort instruction in context")
	}
	if !strings.Contains(ctxPrompt, "Elena (Rol: Protagonista) - Maga de fuego [Notas: Odia los secretos]") {
		t.Errorf("expected Elena character lore in context")
	}
	if !strings.Contains(ctxPrompt, "Mateo (Rol: Antagonista) - Caballero traidor") {
		t.Errorf("expected Mateo character lore in context")
	}
	if !strings.Contains(ctxPrompt, notes) {
		t.Errorf("expected notes in context")
	}
	if !strings.Contains(ctxPrompt, "--- INICIO CAPÍTULO ACTUAL (02_la_emboscada.txt) ---") {
		t.Errorf("expected active chapter header in context")
	}
	if !strings.Contains(ctxPrompt, "La noche era fría y la niebla cubría el sendero.") {
		t.Errorf("expected chapter text in context")
	}

	// Test EffortLow
	lowPrompt := builder.BuildContext(llm.ContextParams{
		EffortLevel: domain.EffortLow,
	})
	if !strings.Contains(lowPrompt, "Sé conciso, directo y breve") {
		t.Errorf("expected low effort instruction")
	}

	// Test EffortMedium
	medPrompt := builder.BuildContext(llm.ContextParams{
		EffortLevel: domain.EffortMedium,
	})
	if !strings.Contains(medPrompt, "Actúa como co-escritor creativo") {
		t.Errorf("expected medium effort instruction")
	}
	// Test with BrainFacts
	brainFacts := []domain.BrainFact{
		{
			Topic:   "Personajes",
			Concept: "Kuno",
			Fact:    "Perdió el brazo izquierdo",
		},
		{
			Topic:   "Lore",
			Concept: "Espada del Alba",
			Fact:    "Brilla en presencia de sombras",
		},
	}
	brainPrompt := builder.BuildContext(llm.ContextParams{
		BrainFacts: brainFacts,
	})
	if !strings.Contains(brainPrompt, "--- MEMORIA DE CONTINUIDAD Y HECHOS (BRAIN) ---") {
		t.Errorf("expected brain header in context")
	}
	if !strings.Contains(brainPrompt, "• [Personajes] Kuno: Perdió el brazo izquierdo") {
		t.Errorf("expected Kuno brain fact in context")
	}
	if !strings.Contains(brainPrompt, "• [Lore] Espada del Alba: Brilla en presencia de sombras") {
		t.Errorf("expected Espada brain fact in context")
	}
}

func TestContextBuilder_Truncation(t *testing.T) {
	builder := &llm.ContextBuilder{
		MaxChapterChars: 50,
	}

	longText := strings.Repeat("Palabra ", 20) // ~160 chars
	prompt := builder.BuildContext(llm.ContextParams{
		ActiveChapterText: longText,
	})

	if !strings.Contains(prompt, llm.TruncationNotice) {
		t.Errorf("expected truncation notice for long chapter text")
	}
}

func TestFactory(t *testing.T) {
	p1, err := llm.NewProvider(domain.LLMConfig{
		Provider: "ollama",
		BaseURL:  "http://localhost:11434",
	})
	if err != nil || p1 == nil {
		t.Fatalf("failed to create ollama provider: %v", err)
	}

	p2, err := llm.NewProvider(domain.LLMConfig{
		Provider: "openai",
		BaseURL:  "https://api.openai.com/v1",
	})
	if err != nil || p2 == nil {
		t.Fatalf("failed to create openai provider: %v", err)
	}
}
