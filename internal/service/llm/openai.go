package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/SalvucciFacundo/novel-tui/internal/domain"
)

// OpenAICompatibleProvider implements domain.LLMProvider for OpenAI and compatible endpoints (OpenRouter, vLLM, etc.).
type OpenAICompatibleProvider struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewOpenAICompatibleProvider creates a new OpenAICompatibleProvider.
func NewOpenAICompatibleProvider(baseURL string, apiKey string, client *http.Client) *OpenAICompatibleProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if client == nil {
		client = &http.Client{}
	}
	return &OpenAICompatibleProvider{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  client,
	}
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type openAIChoice struct {
	Index        int         `json:"index"`
	Delta        openAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type openAIChatChunk struct {
	ID      string         `json:"id"`
	Choices []openAIChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// StreamChat sends a chat completion request with SSE streaming.
func (p *OpenAICompatibleProvider) StreamChat(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error) {
	endpoint := strings.TrimRight(p.BaseURL, "/")
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = endpoint + "/chat/completions"
	}

	messages := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openAIMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	bodyData := openAIChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Stream:      true,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}

	payload, err := json.Marshal(bodyData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.APIKey))
	}

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai connection failed: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		respBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai returned error status %d: %s", resp.StatusCode, string(respBytes))
	}

	outChan := make(chan domain.StreamChunk, 16)

	go func() {
		defer resp.Body.Close()
		defer close(outChan)

		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			if !strings.HasPrefix(line, "data:") {
				continue
			}

			dataPayload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if dataPayload == "[DONE]" {
				select {
				case outChan <- domain.StreamChunk{Done: true}:
				case <-ctx.Done():
				}
				return
			}

			var chunk openAIChatChunk
			if err := json.Unmarshal([]byte(dataPayload), &chunk); err != nil {
				select {
				case outChan <- domain.StreamChunk{Err: fmt.Errorf("failed to decode sse json: %w", err), Done: true}:
				case <-ctx.Done():
				}
				return
			}

			if chunk.Error != nil && chunk.Error.Message != "" {
				select {
				case outChan <- domain.StreamChunk{Err: fmt.Errorf("openai error: %s", chunk.Error.Message), Done: true}:
				case <-ctx.Done():
				}
				return
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				isDone := chunk.Choices[0].FinishReason != nil && *chunk.Choices[0].FinishReason != ""

				if content != "" || isDone {
					select {
					case outChan <- domain.StreamChunk{Content: content, Done: isDone}:
					case <-ctx.Done():
						return
					}
				}
			}
		}

		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			select {
			case outChan <- domain.StreamChunk{Err: fmt.Errorf("stream read error: %w", err), Done: true}:
			case <-ctx.Done():
			}
		}
	}()

	return outChan, nil
}
