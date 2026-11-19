package orchestrator

import (
	"context"
	"testing"

	"github.com/Ashed2127/tertic/internal/agent"
	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/tools"
)

// MockProvider provides synthetic responses without making network requests
type MockProvider struct {
	response string
}

func (m *MockProvider) Name() string              { return "mock" }
func (m *MockProvider) ModelID() string           { return "mock-model" }
func (m *MockProvider) SupportsNativeTools() bool { return false }
func (m *MockProvider) Generate(ctx context.Context, msgs []model.Message, tdefs []tools.ToolDefinition, opts model.ModelOptions) (*model.Response, error) {
	return &model.Response{
		Content: m.response,
	}, nil
}

func TestEngineWorkflow(t *testing.T) {
	agentReg := agent.NewRegistry()
	modelReg := model.NewRegistry()
	mock := &MockProvider{
		response: `[{"task_id":"task_001","assigned_agent":"developer","objective":"Create test component","acceptance_criteria":["component exists"]}]`,
	}
	modelReg.SetDefault(mock)

	var recordedProgress []string
	cfg := Config{
		MaxConcurrentAgents: 1,
		SequentialOnly:      true,
		OnProgress: func(role, msg string) {
			recordedProgress = append(recordedProgress, role+": "+msg)
		},
	}

	engine := NewEngine(agentReg, modelReg, cfg)
	result, err := engine.ExecuteWorkflow(context.Background(), "Build simple component")
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}

	if result == "" {
		t.Fatalf("expected non-empty result")
	}

	if len(recordedProgress) == 0 {
		t.Fatalf("expected progress notifications from agents")
	}
}

func TestToolPermissionEnforcement(t *testing.T) {
	agentReg := agent.NewRegistry()
	modelReg := model.NewRegistry()
	engine := NewEngine(agentReg, modelReg, Config{MaxConcurrentAgents: 1})

	// Researcher is NOT allowed to execute_bash or write_file
	_, err := engine.ExecuteTool("researcher", "execute_bash", map[string]interface{}{"command": "ls"})
	if err == nil {
		t.Fatalf("expected permission denied for researcher using execute_bash")
	}

	// Developer IS allowed to write_file, but empty path must fail validation
	_, err = engine.ExecuteTool("developer", "write_file", map[string]interface{}{"path": "", "content": "test"})
	if err == nil {
		t.Fatalf("expected validation error for empty path")
	}
}
