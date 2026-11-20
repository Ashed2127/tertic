package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Ashed2127/tertic/internal/tools"
)

const (
	DefaultModel = "google/gemma-4-31b-it:free"
	APIBaseURL   = "https://openrouter.ai/api/v1/chat/completions"
)

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
	req.Model = c.Model

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
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse json response: %w, raw: %s", err, string(bodyBytes))
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s (code %d)", chatResp.Error.Message, chatResp.Error.Code)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenRouter API")
	}

	return &chatResp, nil
}
