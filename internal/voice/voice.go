package voice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// InputMode represents the active input interaction paradigm
type InputMode string

const (
	ModeText   InputMode = "text"
	ModeVoice  InputMode = "voice"
	ModeHybrid InputMode = "hybrid"
)

// Engine defines speech capture & transcription abstraction
type Engine interface {
	Name() string
	IsAvailable() (bool, string)
	RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error)
}

// TermuxAPIEngine uses termux-speech-to-text or Termux dialog
type TermuxAPIEngine struct{}

func (t *TermuxAPIEngine) Name() string {
	return "Termux:API Speech Recognition"
}

func (t *TermuxAPIEngine) IsAvailable() (bool, string) {
	if _, err := exec.LookPath("termux-speech-to-text"); err == nil {
		return true, "termux-speech-to-text binary found"
	}
	if _, err := exec.LookPath("termux-dialog"); err == nil {
		return true, "termux-dialog binary found (fallback text/voice entry)"
	}
	return false, "Termux:API package not installed (run: pkg install termux-api && install Termux:API apk)"
}

func (t *TermuxAPIEngine) RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error) {
	if _, err := exec.LookPath("termux-speech-to-text"); err == nil {
		if onStatus != nil {
			onStatus("Listening for speech (Android SpeechRecognizer)...")
		}
		cmd := exec.CommandContext(ctx, "termux-speech-to-text")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("termux-speech-to-text error: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	if _, err := exec.LookPath("termux-dialog"); err == nil {
		if onStatus != nil {
			onStatus("Opening Android voice/text dialog...")
		}
		cmd := exec.CommandContext(ctx, "termux-dialog", "speech", "-t", "Speak your instruction to Tertic")
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("termux-dialog speech error: %w", err)
		}
		// termux-dialog output is json like {"code": 0, "text": "..."}
		s := strings.TrimSpace(string(out))
		if strings.Contains(s, `"text":`) {
			idx := strings.Index(s, `"text":`)
			sub := s[idx+7:]
			sub = strings.Trim(sub, ` "}	\r\n`)
			return sub, nil
		}
		return s, nil
	}

	return "", fmt.Errorf("termux speech tools not available")
}

// WhisperCLIEngine uses whisper.cpp or whisper CLI if available locally on Termux/Linux
type WhisperCLIEngine struct{}

func (w *WhisperCLIEngine) Name() string {
	return "Local Whisper CLI"
}

func (w *WhisperCLIEngine) IsAvailable() (bool, string) {
	if _, err := exec.LookPath("whisper"); err == nil {
		return true, "whisper command found"
	}
	if _, err := exec.LookPath("whisper-cpp"); err == nil {
		return true, "whisper-cpp command found"
	}
	return false, "whisper CLI not found"
}

func (w *WhisperCLIEngine) RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error) {
	// Check recording tool: sox, arecord, ffmpeg, or termux-microphone-record
	tmpAudio := fmt.Sprintf("/tmp/tertic_voice_%d.wav", time.Now().UnixNano())
	defer os.Remove(tmpAudio)

	recordCmd := findRecordCommand(tmpAudio)
	if recordCmd == nil {
		return "", fmt.Errorf("no audio recording utility found (arecord, ffmpeg, sox, or termux-microphone-record)")
	}

	if onStatus != nil {
		onStatus("Recording audio (Speak now)...")
	}

	// Record for up to 10 seconds or until context cancelled
	recCtx, recCancel := context.WithTimeout(ctx, 10*time.Second)
	defer recCancel()

	cmd := exec.CommandContext(recCtx, recordCmd[0], recordCmd[1:]...)
	_ = cmd.Run()

	if _, err := os.Stat(tmpAudio); os.IsNotExist(err) {
		return "", fmt.Errorf("audio file not recorded")
	}

	if onStatus != nil {
		onStatus("Transcribing with local Whisper...")
	}

	whisperBin := "whisper"
	if _, err := exec.LookPath("whisper-cpp"); err == nil {
		whisperBin = "whisper-cpp"
	}

	whisperCmd := exec.CommandContext(ctx, whisperBin, tmpAudio)
	out, err := whisperCmd.Output()
	if err != nil {
		return "", fmt.Errorf("whisper transcription failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func findRecordCommand(outPath string) []string {
	if _, err := exec.LookPath("arecord"); err == nil {
		return []string{"arecord", "-d", "8", "-f", "cd", "-t", "wav", outPath}
	}
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return []string{"ffmpeg", "-y", "-f", "pulse", "-i", "default", "-t", "8", outPath}
	}
	if _, err := exec.LookPath("sox"); err == nil {
		return []string{"rec", "-r", "16000", "-c", "1", outPath, "trim", "0", "8"}
	}
	return nil
}

// Manager coordinates speech detection and active input mode
type Manager struct {
	mu           sync.RWMutex
	currentMode  InputMode
	engines      []Engine
	activeEngine Engine
}

// NewManager initializes the voice manager with available recognition backends
func NewManager() *Manager {
	m := &Manager{
		currentMode: ModeText,
		engines: []Engine{
			&TermuxAPIEngine{},
			&WhisperCLIEngine{},
		},
	}

	// Detect best available engine
	for _, eng := range m.engines {
		if avail, _ := eng.IsAvailable(); avail {
			m.activeEngine = eng
			break
		}
	}

	return m
}

// GetMode returns the current active input mode
func (m *Manager) GetMode() InputMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMode
}

// SetMode updates the active input mode
func (m *Manager) SetMode(mode InputMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentMode = mode
}

// ActiveEngine returns the current speech engine or nil if none
func (m *Manager) ActiveEngine() Engine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeEngine
}

// StatusReport returns human-readable status of microphone & speech recognizers
func (m *Manager) StatusReport() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Input Mode: %s\n", strings.ToUpper(string(m.currentMode))))
	sb.WriteString("Speech Recognition Engines:\n")
	for _, eng := range m.engines {
		avail, details := eng.IsAvailable()
		status := "NOT AVAILABLE"
		if avail {
			status = "READY"
		}
		sb.WriteString(fmt.Sprintf(" • %-28s [%s] (%s)\n", eng.Name(), status, details))
	}
	if m.activeEngine != nil {
		sb.WriteString(fmt.Sprintf("\nActive Engine: %s\n", m.activeEngine.Name()))
	} else {
		sb.WriteString("\nActive Engine: None (Keyboard text input active). To enable voice on Termux:\n")
		sb.WriteString("  1. pkg install termux-api\n")
		sb.WriteString("  2. Install Termux:API app from F-Droid & grant Microphone permission.\n")
	}
	return sb.String()
}
