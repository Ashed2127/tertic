package commands

import (
	"testing"
)

func TestRegistryFilter(t *testing.T) {
	reg := Default()
	all := reg.ListAll()
	if len(all) < 10 {
		t.Fatalf("expected at least 10 registered commands, got %d", len(all))
	}

	// Filter with "/"
	matches := reg.Filter("/", AppState{})
	if len(matches) != len(all) {
		t.Fatalf("expected all commands matching '/', got %d vs %d", len(matches), len(all))
	}

	// Filter with "/mo"
	moMatches := reg.Filter("/mo", AppState{})
	if len(moMatches) < 2 {
		t.Fatalf("expected at least /model and /models for '/mo', got %d", len(moMatches))
	}
	foundModel := false
	foundModels := false
	for _, m := range moMatches {
		if m.Name == "/model" {
			foundModel = true
		}
		if m.Name == "/models" {
			foundModels = true
		}
	}
	if !foundModel || !foundModels {
		t.Fatalf("expected /model and /models in matches, got: %v", moMatches)
	}

	// Filter with "/vo"
	voMatches := reg.Filter("/vo", AppState{})
	if len(voMatches) == 0 || voMatches[0].Name != "/voice" {
		t.Fatalf("expected /voice as top match for '/vo', got: %v", voMatches)
	}

	// Filter with "/ag"
	agMatches := reg.Filter("/ag", AppState{})
	if len(agMatches) == 0 || agMatches[0].Name != "/agents" {
		t.Fatalf("expected /agents for '/ag', got: %v", agMatches)
	}

	// Context-aware filter when recording
	recMatches := reg.Filter("/s", AppState{IsRecording: true})
	if len(recMatches) == 0 || recMatches[0].Name != "/stop" {
		t.Fatalf("expected /stop to be prioritized when recording, got: %v", recMatches)
	}
}

func TestRegistryCategories(t *testing.T) {
	reg := Default()
	seenCategories := make(map[Category]bool)
	for _, cmd := range reg.ListAll() {
		seenCategories[cmd.Category] = true
	}

	expected := []Category{
		CategoryGeneral,
		CategoryModels,
		CategoryAgents,
		CategoryVoice,
		CategoryConfig,
		CategorySession,
		CategoryDeveloper,
	}

	for _, exp := range expected {
		if !seenCategories[exp] {
			t.Errorf("missing category: %s", exp)
		}
	}
}
