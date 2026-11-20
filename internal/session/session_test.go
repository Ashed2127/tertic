package session

import (
	"os"
	"testing"
	"time"

	"github.com/Ashed2127/tertic/internal/openrouter"
)

func TestSessionSaveLoadList(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tertic_sessions_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", origHome)

	s := &SavedSession{
		ID:        "test-session-123",
		Provider:  "openrouter",
		Model:     "nvidia/nemotron-3.5-lightning:free",
		WorkDir:   "/tmp",
		CreatedAt: time.Now(),
		History: []openrouter.Message{
			{Role: "user", Content: "Hello test session"},
			{Role: "assistant", Content: "Hi there!"},
		},
	}

	if err := SaveSession(s); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := LoadSession("test-session-123")
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if loaded.Title != "Hello test session" {
		t.Errorf("expected title 'Hello test session', got '%s'", loaded.Title)
	}
	if len(loaded.History) != 2 {
		t.Errorf("expected 2 messages, got %d", len(loaded.History))
	}

	list, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 session, got %d", len(list))
	}
	if list[0].ID != "test-session-123" {
		t.Errorf("expected session id test-session-123, got %s", list[0].ID)
	}

	latest, err := LoadLatestSession()
	if err != nil {
		t.Fatalf("LoadLatestSession failed: %v", err)
	}
	if latest.ID != "test-session-123" {
		t.Errorf("expected latest id test-session-123, got %s", latest.ID)
	}
}
