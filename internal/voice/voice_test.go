package voice

import (
	"context"
	"strings"
	"testing"
)

type mockEngine struct {
	available bool
	text      string
}

func (m *mockEngine) Name() string { return "Mock Engine" }
func (m *mockEngine) IsAvailable() (bool, string) {
	if m.available {
		return true, "mock ready"
	}
	return false, "mock not ready"
}
func (m *mockEngine) RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error) {
	if onStatus != nil {
		onStatus("Mock recording...")
	}
	return m.text, nil
}

func TestVoiceManagerModes(t *testing.T) {
	mgr := NewManager()
	if mgr.GetMode() != ModeText {
		t.Fatalf("expected initial mode text, got %v", mgr.GetMode())
	}

	mgr.SetMode(ModeVoice)
	if mgr.GetMode() != ModeVoice {
		t.Fatalf("expected voice mode, got %v", mgr.GetMode())
	}

	mgr.SetMode(ModeHybrid)
	if mgr.GetMode() != ModeHybrid {
		t.Fatalf("expected hybrid mode, got %v", mgr.GetMode())
	}
}

func TestMockEngineExecution(t *testing.T) {
	mock := &mockEngine{available: true, text: "Build a calculator in HTML"}
	var statusMsg string
	res, err := mock.RecordAndTranscribe(context.Background(), func(s string) {
		statusMsg = s
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "Build a calculator in HTML" {
		t.Fatalf("unexpected transcribed text: %q", res)
	}
	if statusMsg != "Mock recording..." {
		t.Fatalf("unexpected status msg: %q", statusMsg)
	}
}

func TestVoiceStatusReport(t *testing.T) {
	mgr := NewManager()
	report := mgr.StatusReport()
	if !strings.Contains(report, "Input Mode: TEXT") {
		t.Fatalf("report missing input mode: %s", report)
	}
}
