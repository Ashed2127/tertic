package harness

import (
	"context"
	"testing"

	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/task"
	"github.com/Ashed2127/tertic/internal/tools"
)

type mockHarness struct {
	history []model.Message
	context map[string]interface{}
}

func (m *mockHarness) GetModelProvider(agentName string) (model.Provider, error) {
	return nil, nil
}

func (m *mockHarness) ExecuteTool(agentName, toolName string, args map[string]interface{}) (string, error) {
	return "ok", nil
}

func (m *mockHarness) GetToolDefinitions(toolNames []string) []tools.ToolDefinition {
	return nil
}

func (m *mockHarness) GetSharedContext() map[string]interface{} {
	return m.context
}

func (m *mockHarness) SetSharedContext(key string, val interface{}) {
	m.context[key] = val
}

func (m *mockHarness) GetSharedHistory() []model.Message {
	return m.history
}

func (m *mockHarness) AppendSharedHistory(msgs ...model.Message) {
	m.history = append(m.history, msgs...)
}

func (m *mockHarness) NotifyProgress(agentRole, status string) {}

func TestStandardHarnessInterface(t *testing.T) {
	var h Harness = &mockHarness{
		context: make(map[string]interface{}),
	}

	h.SetSharedContext("test_key", "test_val")
	if h.GetSharedContext()["test_key"] != "test_val" {
		t.Fatalf("expected test_val, got %v", h.GetSharedContext()["test_key"])
	}

	h.AppendSharedHistory(model.Message{Role: "user", Content: "hello"})
	if len(h.GetSharedHistory()) != 1 {
		t.Fatalf("expected 1 history message, got %d", len(h.GetSharedHistory()))
	}
}

type mockAgent struct{}

func (m *mockAgent) Name() string            { return "mock" }
func (m *mockAgent) Role() string            { return "Mock" }
func (m *mockAgent) Description() string     { return "Mock agent" }
func (m *mockAgent) Instructions() string    { return "Instructions" }
func (m *mockAgent) AllowedTools() []string  { return []string{"read_file"} }
func (m *mockAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	return &Result{Summary: "done", Passed: true}, nil
}

func TestStandardAgentInterface(t *testing.T) {
	var a Agent = &mockAgent{}
	if a.Name() != "mock" {
		t.Fatalf("expected mock, got %s", a.Name())
	}
	res, err := a.Execute(context.Background(), &mockHarness{context: make(map[string]interface{})}, "test", nil)
	if err != nil || !res.Passed {
		t.Fatalf("unexpected failure: %v", err)
	}
}
