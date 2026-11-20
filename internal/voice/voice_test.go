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
func (m *mockEngine) StartRecording(ctx context.Context, onStatus func(string)) (*ActiveRecording, error) {
	if onStatus != nil {
		onStatus("Mock recording started...")
	}
	return &ActiveRecording{
		AudioPath: "/tmp/mock.wav",
		stopped:   false,
	}, nil
}
func (m *mockEngine) Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if onStatus != nil {
		onStatus("Mock transcribing...")
	}
	return m.text, nil
}
func (m *mockEngine) RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error) {
	rec, err := m.StartRecording(ctx, onStatus)
	if err != nil {
		return "", err
	}
	_ = rec.Stop()
	return m.Transcribe(ctx, rec.AudioPath, onStatus)
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

func TestMockEngineManualStopRecording(t *testing.T) {
	mock := &mockEngine{available: true, text: "Build a calculator in HTML"}
	var statusMsg string
	rec, err := mock.StartRecording(context.Background(), func(s string) {
		statusMsg = s
	})
	if err != nil {
		t.Fatalf("unexpected error starting recording: %v", err)
	}
	_ = statusMsg

	// Verify we can stop recording on demand
	err = rec.Stop()
	if err != nil {
		t.Fatalf("unexpected error stopping recording: %v", err)
	}

	res, err := mock.Transcribe(context.Background(), rec.AudioPath, func(s string) {
		statusMsg = s
	})
	if err != nil {
		t.Fatalf("unexpected error transcribing: %v", err)
	}
	if res != "Build a calculator in HTML" {
		t.Fatalf("unexpected transcribed text: %q", res)
	}
}

func TestActiveRecordingDoubleStop(t *testing.T) {
	rec := &ActiveRecording{AudioPath: "/tmp/test.wav"}
	if err := rec.Stop(); err != nil {
		t.Fatalf("first stop failed: %v", err)
	}
	if err := rec.Stop(); err != nil {
		t.Fatalf("second stop should be idempotent: %v", err)
	}
}

func TestVoiceStatusReport(t *testing.T) {
	mgr := NewManager()
	report := mgr.StatusReport()
	if !strings.Contains(report, "Input Mode: TEXT") {
		t.Fatalf("report missing input mode: %s", report)
	}
}
