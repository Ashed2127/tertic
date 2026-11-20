package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Ashed2127/tertic/internal/tools"
)

const (
	DefaultModel = "nvidia/nemotron-3.5-lightning:free"
	APIBaseURL   = "https://openrouter.ai/api/v1/chat/completions"
)

// FallbackFreeModels defines resilient failover models when a free model hits 429 or 404
var FallbackFreeModels = []string{
	"nvidia/nemotron-3.5-lightning:free",
	"google/gemma-4-31b-it:free",
	"google/gemma-4-26b-a4b-it:free",
	"inclusionai/ling-3.2-flash:free",
	"liquid/lfm-3-1.7b:free",
}

type Client struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewClient(apiKey string, model string) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		APIKey: apiKey,
		Model:  model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// OpenAI / OpenRouter Compatible Message Structs
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenRouterTool struct {
	Type     string               `json:"type"`
	Function tools.ToolDefinition `json:"function"`
}

type ChatCompletionRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []OpenRouterTool `json:"tools,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func ConvertToolDefinitions(defs []tools.ToolDefinition) []OpenRouterTool {
	var res []OpenRouterTool
	for _, def := range defs {
		res = append(res, OpenRouterTool{
			Type:     "function",
			Function: def,
		})
	}
	return res
}

func (c *Client) Chat(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	candidateModels := []string{c.Model}
	if strings.HasSuffix(c.Model, ":free") {
		for _, m := range FallbackFreeModels {
			if m != c.Model {
				candidateModels = append(candidateModels, m)
			}
		}
	}

	var lastErr error
	for idx, modelToTry := range candidateModels {
		req.Model = modelToTry

		reqBytes, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequest("POST", APIBaseURL, bytes.NewBuffer(reqBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to build http request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
		httpReq.Header.Set("HTTP-Referer", "https://github.com/Ashed2127/tertic")
		httpReq.Header.Set("X-Title", "Tertic Termux AI Agent")

		resp, err := c.HTTPClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		var chatResp ChatCompletionResponse
		if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
			lastErr = fmt.Errorf("failed to parse json response: %w, raw: %s", err, string(bodyBytes))
			continue
		}

		if chatResp.Error != nil {
			lastErr = fmt.Errorf("OpenRouter API error: %s (code %d)", chatResp.Error.Message, chatResp.Error.Code)
			// Check if rate limited (429) or model deprecated/not found (404/400)
			if (chatResp.Error.Code == 429 || chatResp.Error.Code == 404 || chatResp.Error.Code == 400 || strings.Contains(strings.ToLower(chatResp.Error.Message), "rate-limit")) && idx < len(candidateModels)-1 {
				// Auto-failover to next candidate model
				continue
			}
			return nil, lastErr
		}

		if len(chatResp.Choices) == 0 {
			lastErr = fmt.Errorf("no choices returned from OpenRouter API")
			continue
		}

		// Update active model if fallback succeeded
		c.Model = modelToTry
		return &chatResp, nil
	}

	return nil, lastErr
}
