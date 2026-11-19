package task

import (
	"testing"
)

func TestTaskQueueDependencies(t *testing.T) {
	q := NewQueue()

	task0 := &Task{
		TaskID:        "task_000",
		AssignedAgent: "researcher",
		Objective:     "Inspect codebase",
	}
	task1 := &Task{
		TaskID:        "task_001",
		AssignedAgent: "developer",
		Objective:     "Write auth logic",
		Dependencies:  []string{"task_000"},
	}

	_ = q.Add(task0)
	_ = q.Add(task1)

	// First ready task must be task_000 because task_001 depends on it
	next := q.NextReady()
	if next == nil || next.TaskID != "task_000" {
		t.Fatalf("expected task_000 to be ready, got %v", next)
	}

	// While task_000 is in_progress, no new ready task
	second := q.NextReady()
	if second != nil {
		t.Fatalf("expected no ready tasks while dependencies pending, got %v", second)
	}

	// Mark task_000 as completed
	q.UpdateStatus("task_000", StatusCompleted, "Found existing code", "")

	// Now task_001 should become ready
	third := q.NextReady()
	if third == nil || third.TaskID != "task_001" {
		t.Fatalf("expected task_001 to become ready after dependency completed, got %v", third)
	}

	q.UpdateStatus("task_001", StatusCompleted, "Auth implemented", "")
	if !q.AllCompleted() {
		t.Fatalf("expected all tasks to be completed")
	}
}
