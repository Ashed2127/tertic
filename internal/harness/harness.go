package harness

import (
	"context"

	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/task"
	"github.com/Ashed2127/tertic/internal/tools"
)

// Agent defines the standard interface for any autonomous agent in the system
type Agent interface {
	Name() string
	Role() string
	Description() string
	Instructions() string
	AllowedTools() []string
	Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error)
}

// Result is the standardized outcome produced by an agent's turn
type Result struct {
	Summary   string       `json:"summary"`
	Data      interface{}  `json:"data,omitempty"`
	NextTasks []*task.Task `json:"next_tasks,omitempty"`
	Passed    bool         `json:"passed"`
	Feedback  string       `json:"feedback,omitempty"`
}

// Harness defines the standard harness interface providing execution runtime,
// tool permissions, memory storage, model connectivity, and progress notifications.
type Harness interface {
	// Model routing
	GetModelProvider(agentName string) (model.Provider, error)

	// Tool execution & security checks
	ExecuteTool(agentName, toolName string, args map[string]interface{}) (string, error)
	GetToolDefinitions(toolNames []string) []tools.ToolDefinition

	// Context and shared workspace memory
	GetSharedContext() map[string]interface{}
	SetSharedContext(key string, val interface{})

	// Cross-turn conversation memory
	GetSharedHistory() []model.Message
	AppendSharedHistory(msgs ...model.Message)

	// Lifecycle & progress events
	NotifyProgress(agentRole, status string)
}
