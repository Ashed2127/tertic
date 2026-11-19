package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Ashed2127/tertic/internal/harness"
	"github.com/Ashed2127/tertic/internal/model"
	"github.com/Ashed2127/tertic/internal/task"
)

// Re-export standard agent interfaces and types from harness
type Agent = harness.Agent
type Result = harness.Result
type Harness = harness.Harness

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
You own the high-level goal, decide task complexity, coordinate specialized agents, and produce the final verified report.

CORE PRINCIPLES:
- DO NOT output code blocks: Always instruct Developer to create files via write_file/edit_file, build/run them via execute_bash, and open them if requested (e.g. termux-open).
- Verification over claims: Never accept a task as done without verification proof.
- Action over narration: Be decisive, direct, and token-efficient.
- Escalate to the user only when blocked by missing credentials or irreversible decisions.

When analyzing a user request:
- If conversational or a question, answer directly and concisely.
- If it requires software work, formulate the objective and ensure the planner, researcher, developer, and reviewer execute the plan thoroughly.`
}

func (a *CoordinatorAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Evaluating objective...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	systemMsg := model.Message{Role: "system", Content: a.Instructions()}
	userMsg := model.Message{Role: "user", Content: input}

	var messages []model.Message
	messages = append(messages, systemMsg)
	// Inject past multi-turn history so the agent never forgets previous steps or user instructions
	if hist := h.GetSharedHistory(); len(hist) > 0 {
		messages = append(messages, hist...)
	}
	messages = append(messages, userMsg)

	resp, err := p.Generate(ctx, messages, nil, model.ModelOptions{Temperature: 0.2})
	if err != nil {
		return nil, err
	}

	// Persist to shared conversation history
	h.AppendSharedHistory(userMsg, model.Message{Role: "assistant", Content: resp.Content})

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
Break complex coding objectives into minimal, sequential, dependency-aware tasks with measurable acceptance criteria.
Keep plans short (prefer <= 8 tasks for normal requests).

You MUST output ONLY a valid JSON array of tasks in this schema:
[
  {
    "task_id": "task_001",
    "assigned_agent": "developer",
    "objective": "Brief description of atomic work",
    "dependencies": [],
    "acceptance_criteria": ["Explicit testable criteria 1", "Explicit testable criteria 2"]
  }
]
Do not wrap in markdown or backticks if possible, return raw parseable JSON.`
}

func (a *PlannerAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Planning...")
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
Gather only the architectural, file-level, and dependency context the Developer actually needs.
Be extremely selective and token-efficient. Use read_file and list_dir tools to inspect key files without dumping full file content.`
}

func (a *ResearcherAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Researching...")
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
Your mission is to implement software solutions on Android/Termux that demonstrably meet acceptance criteria.

CORE BEHAVIOR:
- WRITE DIRECTLY TO DISK: Never print raw code blocks in your text output! Call write_file or edit_file directly to save code into files.
- BUILD & EXECUTE: Immediately invoke execute_bash to build, compile, or run the project (e.g., npm start, go run, python3 main.py).
- OPEN WHEN REQUESTED: If asked to open, display, or preview an app or document, invoke execute_bash with "termux-open <file>", "termux-open-url <url>", or "xdg-open <file>".
- ReAct + Reflect loop: Observe -> Think -> Act -> Reflect.
- Smallest viable change: Prefer precise edit_file over rewriting entire files.
- Verification over claims: Always compile, run tests, or execute checks before finishing.
- Never leave broken intermediate states. Prefer simple, maintainable solutions.`
}

func (a *DeveloperAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	taskDesc := input
	if currentTask != nil {
		taskDesc = fmt.Sprintf("Task [%s]: %s\nAcceptance Criteria: %v", currentTask.TaskID, currentTask.Objective, currentTask.AcceptanceCriteria)
	}

	h.NotifyProgress(a.Role(), "Writing code...")
	p, err := h.GetModelProvider(a.Name())
	if err != nil {
		return nil, err
	}

	toolDefs := h.GetToolDefinitions(a.AllowedTools())
	messages := []model.Message{
		{Role: "system", Content: a.Instructions()},
	}
	// Inject past multi-turn context so developer remembers previous changes & files
	if hist := h.GetSharedHistory(); len(hist) > 0 {
		messages = append(messages, hist...)
	}
	messages = append(messages, model.Message{Role: "user", Content: taskDesc})

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

	if finalSummary != "" {
		h.AppendSharedHistory(model.Message{Role: "assistant", Content: fmt.Sprintf("[%s]: %s", a.Role(), finalSummary)})
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
Audit implementations ruthlessly against acceptance criteria.
Demand concrete proof (compiler/test status, file verification).
If tests fail or criteria are not satisfied, produce concise, actionable feedback for the Developer.
Never give vague criticism.`
}

func (a *ReviewerAgent) Execute(ctx context.Context, h Harness, input string, currentTask *task.Task) (*Result, error) {
	h.NotifyProgress(a.Role(), "Verifying...")
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
