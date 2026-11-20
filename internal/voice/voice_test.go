package voice

import (
	"context"
	"os"
	"path/filepath"
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
	tmp := GenerateAudioPath()
	// Write dummy WAV data (> 44 bytes)
	data := append([]byte("RIFF1234WAVEfmt "), make([]byte, 64)...)
	_ = os.WriteFile(tmp, data, 0644)

	return &ActiveRecording{
		AudioPath: tmp,
		stopped:   false,
	}, nil
}
func (m *mockEngine) Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if err := ValidateAudioFile(audioPath); err != nil {
		return "", err
	}
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

func TestRecordingsDirectory(t *testing.T) {
	dir := GetRecordingsDir()
	if !strings.Contains(dir, ".tertic/recordings") {
		t.Fatalf("unexpected recordings directory: %s", dir)
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("recordings directory was not created: %s", dir)
	}

	p := GenerateAudioPath()
	if filepath.Dir(p) != dir {
		t.Fatalf("generated path %s is not in %s", p, dir)
	}
}

func TestValidateAudioFile(t *testing.T) {
	// Non-existent file
	if err := ValidateAudioFile("/path/does/not/exist.wav"); err == nil {
		t.Fatal("expected error for non-existent file")
	}

	// Empty file
	emptyFile := filepath.Join(GetRecordingsDir(), "empty_test.wav")
	_ = os.WriteFile(emptyFile, []byte{}, 0644)
	defer os.Remove(emptyFile)

	if err := ValidateAudioFile(emptyFile); err == nil {
		t.Fatal("expected error for empty file")
	}

	// Valid WAV file
	validFile := filepath.Join(GetRecordingsDir(), "valid_test.wav")
	validData := append([]byte("RIFFheaderdataWAVEfmt "), make([]byte, 100)...)
	_ = os.WriteFile(validFile, validData, 0644)
	defer os.Remove(validFile)

	if err := ValidateAudioFile(validFile); err != nil {
		t.Fatalf("expected valid file to pass, got: %v", err)
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
	defer os.Remove(rec.AudioPath)
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
	if !strings.Contains(report, ".tertic/recordings") {
		t.Fatalf("report missing recordings directory: %s", report)
	}
}
