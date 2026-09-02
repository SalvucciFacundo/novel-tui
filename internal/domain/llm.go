package domain

import (
	"context"
	"time"
)

// LLMEffortLevel defines reasoning and creative depth for LLM interactions.
type LLMEffortLevel string

const (
	EffortLow    LLMEffortLevel = "low"
	EffortMedium LLMEffortLevel = "medium"
	EffortHigh   LLMEffortLevel = "high"
)

// StreamChunk represents a streamed token or completion chunk from an LLM provider.
type StreamChunk struct {
	Content string
	Done    bool
	Err     error
}

// ChatMessage represents a single message in a conversation thread.
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// ChatRequest encapsulates all parameters needed to query an LLM stream.
type ChatRequest struct {
	Messages    []ChatMessage
	Model       string
	Temperature float64
	MaxTokens   int
}

// LLMProvider defines the streaming contract for LLM backends (Ollama, OpenAI-compatible, etc.).
type LLMProvider interface {
	StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error)
}
