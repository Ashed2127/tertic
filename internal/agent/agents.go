package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/task"
	"github.com/Ashed2127/tertic/internal/tools"
)

// Agent defines the interface for specialized autonomous agents
type Agent interface {
	Name() string
	Role() string
	Description() string
	Instructions() string
	AllowedTools() []string
	Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error)
}

// Result is the outcome of an agent's execution
type Result struct {
	Summary   string            `json:"summary"`
	Data      interface{}       `json:"data,omitempty"`
	NextTasks []*task.Task      `json:"next_tasks,omitempty"`
	Passed    bool              `json:"passed"`
	Feedback  string            `json:"feedback,omitempty"`
}

// Harness exposes execution services and resources to agents
type Harness interface {
	GetModelProvider(agentName string) (model.Provider, error)
	ExecuteTool(agentName, toolName string, args map[string]interface{}) (string, error)
	GetSharedContext() map[string]interface{}
	SetSharedContext(key string, val interface{})
	GetToolDefinitions(toolNames []string) []tools.ToolDefinition
	NotifyProgress(agentName, status string)
}

// BaseAgent contains common attributes
type BaseAgent struct {
	name        string
	role        string
	description string
	allowed     []string
}

func (b *BaseAgent) Name() string            { return b.name }
func (b *BaseAgent) Role() string            { return b.role }
func (b *BaseAgent) Description() string     { return b.description }
func (b *BaseAgent) AllowedTools() []string  { return b.allowed }

// -------------------------------------------------------------
// 1. Coordinator Agent
// -------------------------------------------------------------
type CoordinatorAgent struct {
	BaseAgent
}

func NewCoordinatorAgent() *CoordinatorAgent {
	return &CoordinatorAgent{
		BaseAgent: BaseAgent{
			name:        "coordinator",
			role:        "Coordinator",
			description: "Evaluates objectives, routes work, integrates results and verifies final delivery",
			allowed:     []string{"read_file", "list_dir"},
		},
	}
}

func (a *CoordinatorAgent) Instructions() string {
	return `You are the Coordinator Agent in the Tertic multi-agent software engineering system on Android.
Your job is to understand the overall user objective, determine task complexity, coordinate specialized agents, and evaluate final outcomes.
When analyzing a user request:
- If it's a simple greeting or question, answer directly.
- If it requires coding, breakdown the goals and ensure the developer and reviewer execute the plan thoroughly.`
}

func (a *CoordinatorAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Evaluating objective & integrating work...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	systemMsg := model.Message{Role: "system", Content: a.Instructions()}
	userMsg := model.Message{Role: "user", Content: input}

	resp, err := p.Generate(ctx, []model.Message{systemMsg, userMsg}, nil, model.ModelOptions{Temperature: 0.2})
	if err != nil {
		return nil, err
	}

	return &Result{
		Summary: resp.Content,
		Passed:  true,
	}, nil
}

// -------------------------------------------------------------
// 2. Planner Agent
// -------------------------------------------------------------
type PlannerAgent struct {
	BaseAgent
}

func NewPlannerAgent() *PlannerAgent {
	return &PlannerAgent{
		BaseAgent: BaseAgent{
			name:        "planner",
			role:        "Planner",
			description: "Breaks objectives into atomic structured tasks with dependencies and acceptance criteria",
			allowed:     []string{"list_dir", "read_file"},
		},
	}
}

func (a *PlannerAgent) Instructions() string {
	return `You are the Planner Agent in Tertic.
Break complex coding objectives into a minimal, sequential list of atomic tasks for Developer and Reviewer agents.
You MUST output ONLY a valid JSON array of tasks in this schema:
[
  {
    "task_id": "task_001",
    "assigned_agent": "developer",
    "objective": "Brief description of work",
    "dependencies": [],
    "acceptance_criteria": ["Criteria 1", "Criteria 2"]
  }
]
Do not wrap in markdown or backticks if possible, return raw parseable JSON.`
}

func (a *PlannerAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Creating structured execution plan...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf("Create a structured execution plan for this goal:\n%s", input)
	messages := []model.Message{
		{Role: "system", Content: a.Instructions()},
		{Role: "user", Content: prompt},
	}

	resp, err := p.Generate(ctx, messages, nil, model.ModelOptions{Temperature: 0.1})
	if err != nil {
		return nil, err
	}

	// Clean JSON
	cleanJSON := resp.Content
	startIdx := strings.Index(cleanJSON, "[")
	endIdx := strings.LastIndex(cleanJSON, "]")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		cleanJSON = cleanJSON[startIdx : endIdx+1]
	}

	var parsedTasks []*task.Task
	if err := json.Unmarshal([]byte(cleanJSON), &parsedTasks); err != nil {
		// Fallback to single primary task
		parsedTasks = []*task.Task{
			{
				TaskID:             "task_001",
				AssignedAgent:      "developer",
				Objective:          input,
				AcceptanceCriteria: []string{"Task executes cleanly and delivers working code"},
				Status:             task.StatusPending,
			},
		}
	}

	return &Result{
		Summary:   fmt.Sprintf("Created plan with %d task(s)", len(parsedTasks)),
		NextTasks: parsedTasks,
		Passed:    true,
	}, nil
}

// -------------------------------------------------------------
// 3. Researcher Agent
// -------------------------------------------------------------
type ResearcherAgent struct {
	BaseAgent
}

func NewResearcherAgent() *ResearcherAgent {
	return &ResearcherAgent{
		BaseAgent: BaseAgent{
			name:        "researcher",
			role:        "Researcher",
			description: "Inspects codebase structure, reads project files, and gathers architectural context",
			allowed:     []string{"read_file", "list_dir"},
		},
	}
}

func (a *ResearcherAgent) Instructions() string {
	return `You are the Researcher Agent in Tertic.
Your task is to inspect directories and read relevant project files using available read_file and list_dir tools.
Provide concise architectural insights, library details, and patterns to the Developer agent.`
}

func (a *ResearcherAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Inspecting project context...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	toolDefs := h.GetToolDefinitions(a.AllowedTools())
	messages := []model.Message{
		{Role: "system", Content: a.Instructions()},
		{Role: "user", Content: fmt.Sprintf("Gather technical context for: %s", input)},
	}

	resp, err := p.Generate(ctx, messages, toolDefs, model.ModelOptions{Temperature: 0.2})
	if err != nil {
		return nil, err
	}

	// If tools requested, execute them
	for _, tc := range resp.ToolCalls {
		out, toolErr := h.ExecuteTool(a.Name(), tc.Name, tc.Arguments)
		if toolErr != nil {
			out = fmt.Sprintf("Error: %v", toolErr)
		}
		messages = append(messages, model.Message{
			Role:       "tool",
			ToolCallID: tc.Name,
			Content:    out,
		})
	}

	return &Result{
		Summary: resp.Content,
		Passed:  true,
	}, nil
}

// -------------------------------------------------------------
// 4. Developer Agent
// -------------------------------------------------------------
type DeveloperAgent struct {
	BaseAgent
}

func NewDeveloperAgent() *DeveloperAgent {
	return &DeveloperAgent{
		BaseAgent: BaseAgent{
			name:        "developer",
			role:        "Developer",
			description: "Implements software solutions, writes code, edits files, and executes commands",
			allowed:     []string{"execute_bash", "write_file", "read_file", "edit_file", "list_dir"},
		},
	}
}

func (a *DeveloperAgent) Instructions() string {
	return `You are the Developer Agent in Tertic.
You write, refactor, and test working code on Android/Termux.
When assigned a task, invoke tools (write_file, edit_file, execute_bash) directly to deliver the solution.
Verify that created code compiles and runs.`
}

func (a *DeveloperAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	taskDesc := input
	if currentTask != nil {
		taskDesc = fmt.Sprintf("Task [%s]: %s\nAcceptance Criteria: %v", currentTask.TaskID, currentTask.Objective, currentTask.AcceptanceCriteria)
	}

	h.NotifyProgress(a.Role(), "Implementing task...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	toolDefs := h.GetToolDefinitions(a.AllowedTools())
	messages := []model.Message{
		{Role: "system", Content: a.Instructions()},
		{Role: "user", Content: taskDesc},
	}

	maxSteps := 10
	var finalSummary string
	for step := 0; step < maxSteps; step++ {
		resp, err := p.Generate(ctx, messages, toolDefs, model.ModelOptions{Temperature: 0.1})
		if err != nil {
			return nil, err
		}

		if resp.Content != "" {
			finalSummary = resp.Content
			messages = append(messages, model.Message{Role: "assistant", Content: resp.Content})
		}

		if len(resp.ToolCalls) == 0 {
			break
		}

		for _, tc := range resp.ToolCalls {
			toolOut, tErr := h.ExecuteTool(a.Name(), tc.Name, tc.Arguments)
			if tErr != nil {
				toolOut = fmt.Sprintf("Tool error: %v", tErr)
			}
			messages = append(messages, model.Message{
				Role:       "tool",
				ToolCallID: tc.Name,
				Content:    toolOut,
			})
		}
	}

	return &Result{
		Summary: finalSummary,
		Passed:  true,
	}, nil
}

// -------------------------------------------------------------
// 5. Reviewer Agent
// -------------------------------------------------------------
type ReviewerAgent struct {
	BaseAgent
}

func NewReviewerAgent() *ReviewerAgent {
	return &ReviewerAgent{
		BaseAgent: BaseAgent{
			name:        "reviewer",
			role:        "Reviewer",
			description: "Audits implementations against acceptance criteria, tests results, and verifies quality",
			allowed:     []string{"execute_bash", "read_file", "list_dir"},
		},
	}
}

func (a *ReviewerAgent) Instructions() string {
	return `You are the Reviewer Agent in Tertic.
Audit the work against acceptance criteria. Read generated files and optionally run bash verification tests.
Decide if the task passed or failed.
If issues exist, provide actionable feedback for the Developer agent.`
}

func (a *ReviewerAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Reviewing changes & verifying criteria...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	criteria := "General correctness"
	if currentTask != nil && len(currentTask.AcceptanceCriteria) > 0 {
		criteria = strings.Join(currentTask.AcceptanceCriteria, ", ")
	}

	prompt := fmt.Sprintf("Audit the implementation for: %s\nAcceptance Criteria: %s\nReport whether it passes.", input, criteria)
	toolDefs := h.GetToolDefinitions(a.AllowedTools())

	messages := []model.Message{
		{Role: "system", Content: a.Instructions()},
		{Role: "user", Content: prompt},
	}

	resp, err := p.Generate(ctx, messages, toolDefs, model.ModelOptions{Temperature: 0.1})
	if err != nil {
		return nil, err
	}

	passed := true
	lower := strings.ToLower(resp.Content)
	if strings.Contains(lower, "failed") || strings.Contains(lower, "error") || strings.Contains(lower, "missing") {
		passed = false
	}

	return &Result{
		Summary:  resp.Content,
		Passed:   passed,
		Feedback: resp.Content,
	}, nil
}

// Registry manages all registered agents
type Registry struct {
	agents map[string]Agent
}

func NewRegistry() *Registry {
	r := &Registry{agents: make(map[string]Agent)}
	r.Register(NewCoordinatorAgent())
	r.Register(NewPlannerAgent())
	r.Register(NewResearcherAgent())
	r.Register(NewDeveloperAgent())
	r.Register(NewReviewerAgent())
	return r
}

func (r *Registry) Register(a Agent) {
	r.agents[a.Name()] = a
}

func (r *Registry) Get(name string) (Agent, bool) {
	a, ok := r.agents[name]
	return a, ok
}

func (r *Registry) List() []Agent {
	var list []Agent
	for _, a := range []string{"coordinator", "planner", "researcher", "developer", "reviewer"} {
		if ag, ok := r.agents[a]; ok {
			list = append(list, ag)
		}
	}
	return list
}
