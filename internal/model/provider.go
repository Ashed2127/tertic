package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Ashed2127/tertic/internal/gemini"
	"github.com/Ashed2127/tertic/internal/ollama"
	"github.com/Ashed2127/tertic/internal/openrouter"
	"github.com/Ashed2127/tertic/internal/tools"
)

// Message is a unified chat message across model providers
type Message struct {
	Role       string            `json:"role"` // system, user, assistant, tool
	Content    string            `json:"content"`
	ToolCallID string            `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCallRequest `json:"tool_calls,omitempty"`
}

// ToolCallRequest represents a requested tool invocation
type ToolCallRequest struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ModelOptions represents per-invocation or per-agent configuration
type ModelOptions struct {
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}

// Response is the unified model output
type Response struct {
	Content      string            `json:"content"`
	ToolCalls    []ToolCallRequest `json:"tool_calls,omitempty"`
	FinishReason string            `json:"finish_reason,omitempty"`
}

// Provider is the modular interface implemented by all model backends
type Provider interface {
	Name() string
	ModelID() string
	Generate(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition, opts ModelOptions) (*Response, error)
	SupportsNativeTools() bool
}

// OpenRouterProvider wraps OpenRouter API
type OpenRouterProvider struct {
	client *openrouter.Client
	model  string
}

func NewOpenRouterProvider(apiKey, model string) *OpenRouterProvider {
	if model == "" {
		model = openrouter.DefaultModel
	}
	return &OpenRouterProvider{
		client: openrouter.NewClient(apiKey, model),
		model:  model,
	}
}

func (p *OpenRouterProvider) Name() string               { return "openrouter" }
func (p *OpenRouterProvider) ModelID() string            { return p.model }
func (p *OpenRouterProvider) SupportsNativeTools() bool  { return true }

func (p *OpenRouterProvider) Generate(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition, opts ModelOptions) (*Response, error) {
	var orMsgs []openrouter.Message
	for _, m := range messages {
		var tcs []openrouter.ToolCall
		for _, tc := range m.ToolCalls {
			argsBytes, _ := tools.MarshalJSON(tc.Arguments)
			tcs = append(tcs, openrouter.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openrouter.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argsBytes),
				},
			})
		}
		orMsgs = append(orMsgs, openrouter.Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
			ToolCalls:  tcs,
		})
	}

	req := openrouter.ChatCompletionRequest{
		Model:    p.model,
		Messages: orMsgs,
	}
	if len(toolDefs) > 0 {
		req.Tools = openrouter.ConvertToolDefinitions(toolDefs)
	}

	resp, err := p.client.Chat(req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from openrouter")
	}

	choice := resp.Choices[0]
	res := &Response{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
	}

	for _, tc := range choice.Message.ToolCalls {
		var args map[string]interface{}
		_ = tools.UnmarshalJSON([]byte(tc.Function.Arguments), &args)
		res.ToolCalls = append(res.ToolCalls, ToolCallRequest{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	return res, nil
}

// OllamaProvider wraps local Ollama models
type OllamaProvider struct {
	client *ollama.Client
	model  string
}

func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	if baseURL == "" {
		baseURL = ollama.DefaultHost
	}
	return &OllamaProvider{
		client: ollama.NewClient(baseURL, model),
		model:  model,
	}
}

func (p *OllamaProvider) Name() string               { return "ollama" }
func (p *OllamaProvider) ModelID() string            { return p.model }
func (p *OllamaProvider) SupportsNativeTools() bool  { return true }

func (p *OllamaProvider) Generate(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition, opts ModelOptions) (*Response, error) {
	var orMsgs []openrouter.Message
	for _, m := range messages {
		var tcs []openrouter.ToolCall
		for _, tc := range m.ToolCalls {
			argsBytes, _ := tools.MarshalJSON(tc.Arguments)
			tcs = append(tcs, openrouter.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openrouter.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argsBytes),
				},
			})
		}
		orMsgs = append(orMsgs, openrouter.Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
			ToolCalls:  tcs,
		})
	}

	req := openrouter.ChatCompletionRequest{
		Model:    p.model,
		Messages: orMsgs,
	}
	if len(toolDefs) > 0 {
		req.Tools = openrouter.ConvertToolDefinitions(toolDefs)
	}

	resp, err := p.client.Chat(req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned from ollama")
	}

	choice := resp.Choices[0]
	res := &Response{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
	}

	for _, tc := range choice.Message.ToolCalls {
		var args map[string]interface{}
		_ = tools.UnmarshalJSON([]byte(tc.Function.Arguments), &args)
		res.ToolCalls = append(res.ToolCalls, ToolCallRequest{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: args,
		})
	}

	return res, nil
}

// GeminiProvider wraps Google Gemini API
type GeminiProvider struct {
	client *gemini.Client
	model  string
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	if model == "" {
		model = gemini.DefaultModel
	}
	return &GeminiProvider{
		client: gemini.NewClient(apiKey, model),
		model:  model,
	}
}

func (p *GeminiProvider) Name() string               { return "gemini" }
func (p *GeminiProvider) ModelID() string            { return p.model }
func (p *GeminiProvider) SupportsNativeTools() bool  { return true }

func (p *GeminiProvider) Generate(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition, opts ModelOptions) (*Response, error) {
	var gemContents []gemini.Content
	var systemText string

	for _, m := range messages {
		if m.Role == "system" {
			systemText += m.Content + "\n"
			continue
		}

		var parts []gemini.Part
		if m.Content != "" {
			parts = append(parts, gemini.Part{Text: m.Content})
		}
		for _, tc := range m.ToolCalls {
			parts = append(parts, gemini.Part{
				FunctionCall: &gemini.FunctionCall{
					ID:   tc.ID,
					Name: tc.Name,
					Args: tc.Arguments,
				},
			})
		}

		role := m.Role
		if role == "tool" {
			role = "user"
			parts = append(parts, gemini.Part{
				FunctionResponse: &gemini.FunctionResponse{
					Name:     m.ToolCallID,
					Response: map[string]interface{}{"result": m.Content},
				},
			})
		}

		gemContents = append(gemContents, gemini.Content{
			Role:  role,
			Parts: parts,
		})
	}

	req := gemini.GenerateContentRequest{
		Contents: gemContents,
	}
	if strings.TrimSpace(systemText) != "" {
		req.SystemInstruction = &gemini.SystemInstruction{
			Parts: []gemini.Part{{Text: strings.TrimSpace(systemText)}},
		}
	}
	if len(toolDefs) > 0 {
		req.Tools = []gemini.ToolDeclarationWrapper{
			{FunctionDeclarations: toolDefs},
		}
	}

	resp, err := p.client.Generate(req)
	if err != nil {
		return nil, err
	}
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from gemini")
	}

	cand := resp.Candidates[0]
	res := &Response{}

	for _, part := range cand.Content.Parts {
		if part.Text != "" {
			res.Content += part.Text
		}
		if part.FunctionCall != nil {
			res.ToolCalls = append(res.ToolCalls, ToolCallRequest{
				ID:        part.FunctionCall.ID,
				Name:      part.FunctionCall.Name,
				Arguments: part.FunctionCall.Args,
			})
		}
	}

	return res, nil
}

// Registry maintains active model providers
type Registry struct {
	providers map[string]Provider
	defaultP  Provider
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

func (r *Registry) Register(name string, p Provider) {
	r.providers[name] = p
	if r.defaultP == nil {
		r.defaultP = p
	}
}

func (r *Registry) SetDefault(p Provider) {
	r.defaultP = p
}

func (r *Registry) Get(name string) (Provider, error) {
	if p, ok := r.providers[name]; ok {
		return p, nil
	}
	if r.defaultP != nil {
		return r.defaultP, nil
	}
	return nil, fmt.Errorf("no model provider configured for %s", name)
}

func (r *Registry) List() []string {
	var list []string
	for k, p := range r.providers {
		list = append(list, fmt.Sprintf("%s (%s - %s)", k, p.Name(), p.ModelID()))
	}
	return list
}
