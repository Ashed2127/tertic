package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Ashed2127/tertic/internal/agent"
	"github.com/Ashed2127/tertic/internal/harness"
	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/task"
	"github.com/Ashed2127/tertic/internal/tools"
)

// Ensure Engine explicitly satisfies the standard harness.Harness interface
var _ harness.Harness = (*Engine)(nil)

// ProgressCallback receives updates when an agent makes progress
type ProgressCallback func(agentRole, message string)

// Engine is the central orchestration harness running the multi-agent system
type Engine struct {
	agentRegistry *agent.Registry
	modelRegistry *model.Registry
	toolValidator *tools.ToolValidator
	taskQueue     *task.Queue
	sharedContext map[string]interface{}
	contextMu     sync.RWMutex
	onProgress    ProgressCallback

	// Resource control for Android low-memory constraints
	sequentialOnly bool
	sem            chan struct{}
}

// Config configures the multi-agent engine
type Config struct {
	MaxConcurrentAgents int
	SequentialOnly      bool
	OnProgress          ProgressCallback
}

func NewEngine(agentReg *agent.Registry, modelReg *model.Registry, cfg Config) *Engine {
	if cfg.MaxConcurrentAgents <= 0 {
		cfg.MaxConcurrentAgents = 1 // Safe default for Termux on Android
	}

	return &Engine{
		agentRegistry:  agentReg,
		modelRegistry:  modelReg,
		toolValidator:  tools.NewToolValidator(),
		taskQueue:      task.NewQueue(),
		sharedContext:  make(map[string]interface{}),
		onProgress:     cfg.OnProgress,
		sequentialOnly: cfg.SequentialOnly,
		sem:            make(chan struct{}, cfg.MaxConcurrentAgents),
	}
}

// Harness implementation
func (e *Engine) GetModelProvider(agentName string) (model.Provider, error) {
	return e.modelRegistry.Get(agentName)
}

func (e *Engine) GetSharedHistory() []model.Message {
	e.contextMu.RLock()
	defer e.contextMu.RUnlock()
	if raw, ok := e.sharedContext["_conversation_history"]; ok {
		if history, ok := raw.([]model.Message); ok {
			copied := make([]model.Message, len(history))
			copy(copied, history)
			return copied
		}
	}
	return nil
}

func (e *Engine) AppendSharedHistory(msgs ...model.Message) {
	e.contextMu.Lock()
	defer e.contextMu.Unlock()
	var history []model.Message
	if raw, ok := e.sharedContext["_conversation_history"]; ok {
		if h, ok := raw.([]model.Message); ok {
			history = h
		}
	}
	history = append(history, msgs...)
	// Keep history bounded to avoid OOM on mobile
	if len(history) > 40 {
		history = history[len(history)-40:]
	}
	e.sharedContext["_conversation_history"] = history
}

func (e *Engine) ExecuteTool(agentName, toolName string, args map[string]interface{}) (string, error) {
	// 1. Permission check
	ag, ok := e.agentRegistry.Get(agentName)
	if !ok {
		return "", fmt.Errorf("unknown agent: %s", agentName)
	}

	allowed := false
	for _, t := range ag.AllowedTools() {
		if t == toolName {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("permission denied: agent %s is not permitted to use tool %s", agentName, toolName)
	}

	// 2. Argument validation & security checks
	if err := e.toolValidator.Validate(toolName, args); err != nil {
		return "", fmt.Errorf("validation error: %w", err)
	}

	// 3. Execution
	e.NotifyProgress("Action", toolActionLabel(toolName, args))
	return tools.ExecuteTool(toolName, args)
}

// toolActionLabel produces a Devin-style status line for the spinner
func toolActionLabel(toolName string, args map[string]interface{}) string {
	path, _ := args["path"].(string)
	cmd, _ := args["command"].(string)
	if i := strings.IndexByte(cmd, '\n'); i != -1 {
		cmd = cmd[:i]
	}
	if len(cmd) > 80 {
		cmd = cmd[:77] + "..."
	}
	switch toolName {
	case "execute_bash":
		if cmd != "" {
			return "Running: " + cmd
		}
		return "Running command..."
	case "write_file", "edit_file":
		if path != "" {
			return "Editing: " + path
		}
		return "Editing file..."
	case "read_file":
		if path != "" {
			return "Reading: " + path
		}
		return "Reading file..."
	case "list_dir":
		return "Listing directory..."
	default:
		return "Working..."
	}
}

func (e *Engine) GetSharedContext() map[string]interface{} {
	e.contextMu.RLock()
	defer e.contextMu.RUnlock()
	copied := make(map[string]interface{})
	for k, v := range e.sharedContext {
		copied[k] = v
	}
	return copied
}

func (e *Engine) SetSharedContext(key string, val interface{}) {
	e.contextMu.Lock()
	defer e.contextMu.Unlock()
	e.sharedContext[key] = val
}

func (e *Engine) GetToolDefinitions(toolNames []string) []tools.ToolDefinition {
	var defs []tools.ToolDefinition
	allowedMap := make(map[string]bool)
	for _, t := range toolNames {
		allowedMap[t] = true
	}

	for _, d := range tools.GetToolDefinitions() {
		if allowedMap[d.Name] {
			defs = append(defs, d)
		}
	}
	return defs
}

func (e *Engine) NotifyProgress(agentRole, message string) {
	if e.onProgress != nil {
		e.onProgress(agentRole, message)
	}
}

// ExecuteWorkflow executes the complete 5-stage multi-agent loop
func (e *Engine) ExecuteWorkflow(ctx context.Context, userObjective string) (string, error) {
	// Resource lock: ensures sequential model execution on Android by default
	e.sem <- struct{}{}
	defer func() { <-e.sem }()

	// 1. Coordinator: evaluate objective
	coord, _ := e.agentRegistry.Get("coordinator")
	coordRes, err := coord.Execute(ctx, e, userObjective, nil)
	if err != nil {
		return "", fmt.Errorf("coordinator error: %w", err)
	}

	// 2. Planner: generate atomic tasks with dependencies
	planner, _ := e.agentRegistry.Get("planner")
	planRes, err := planner.Execute(ctx, e, userObjective, nil)
	if err != nil {
		return "", fmt.Errorf("planner error: %w", err)
	}

	for _, t := range planRes.NextTasks {
		_ = e.taskQueue.Add(t)
	}

	// 3. Researcher: gather context for the project
	researcher, _ := e.agentRegistry.Get("researcher")
	researchRes, _ := researcher.Execute(ctx, e, userObjective, nil)
	if researchRes != nil {
		e.SetSharedContext("project_context", researchRes.Summary)
	}

	// 4. Autonomous Task Loop (Developer -> Reviewer -> Feedback)
	maxLoopSteps := 20
	steps := 0

	for e.taskQueue.HasPendingOrRunning() && steps < maxLoopSteps {
		steps++
		t := e.taskQueue.NextReady()
		if t == nil {
			// No ready task right now, brief pause for dependencies
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Delegate to assigned agent
		targetAgentName := t.AssignedAgent
		if targetAgentName == "" {
			targetAgentName = "developer"
		}

		targetAgent, ok := e.agentRegistry.Get(targetAgentName)
		if !ok {
			targetAgent, _ = e.agentRegistry.Get("developer")
		}

		devRes, devErr := targetAgent.Execute(ctx, e, t.Objective, t)
		if devErr != nil {
			t.RetryCount++
			if t.RetryCount <= t.MaxRetries {
				e.taskQueue.UpdateStatus(t.TaskID, task.StatusPending, "", devErr.Error())
				continue
			}
			e.taskQueue.UpdateStatus(t.TaskID, task.StatusFailed, "", devErr.Error())
			continue
		}

		// 5. Reviewer validates against acceptance criteria
		reviewer, _ := e.agentRegistry.Get("reviewer")
		reviewRes, revErr := reviewer.Execute(ctx, e, fmt.Sprintf("Task: %s\nDeveloper Output: %s", t.Objective, devRes.Summary), t)

		if revErr == nil && !reviewRes.Passed && t.RetryCount < t.MaxRetries {
			// Failed acceptance criteria: return to developer with reviewer feedback
			t.RetryCount++
			e.NotifyProgress("Reviewer", fmt.Sprintf("Revising (%s)", reviewRes.Feedback))
			t.Objective = fmt.Sprintf("%s\nReviewer feedback: %s", t.Objective, reviewRes.Feedback)
			e.taskQueue.UpdateStatus(t.TaskID, task.StatusPending, "", reviewRes.Feedback)
			continue
		}

		e.taskQueue.UpdateStatus(t.TaskID, task.StatusCompleted, devRes.Summary, "")
	}

	// 6. Coordinator final verification
	e.NotifyProgress("Coordinator", "Finalizing...")
	var finalReport string
	if e.taskQueue.AllCompleted() {
		finalReport = fmt.Sprintf("Successfully completed all planned tasks for: %s", userObjective)
	} else {
		finalReport = fmt.Sprintf("Execution finished with partial status:\n%s", e.taskQueue.ToJSON())
	}

	_ = coordRes
	return finalReport, nil
}
