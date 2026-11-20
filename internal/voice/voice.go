package voice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// InputMode represents the active input interaction paradigm
type InputMode string

const (
	ModeText   InputMode = "text"
	ModeVoice  InputMode = "voice"
	ModeHybrid InputMode = "hybrid"
)

// ActiveRecording represents an ongoing recording that can be stopped at any time
type ActiveRecording struct {
	AudioPath string
	Cmd       *exec.Cmd
	IsTermux  bool
	stopped   bool
	mu        sync.Mutex
}

// Stop stops the ongoing recording process cleanly
func (r *ActiveRecording) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopped {
		return nil
	}
	r.stopped = true

	if r.IsTermux {
		stopCmd := exec.Command("termux-microphone-record", "-q")
		_ = stopCmd.Run()
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	if r.Cmd != nil && r.Cmd.Process != nil {
		// Send SIGINT so ffmpeg/arecord flushes WAV headers properly
		_ = r.Cmd.Process.Signal(syscall.SIGINT)
		done := make(chan error, 1)
		go func() {
			done <- r.Cmd.Wait()
		}()
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			_ = r.Cmd.Process.Kill()
		}
	}
	return nil
}

// Engine defines speech capture & transcription abstraction
type Engine interface {
	Name() string
	IsAvailable() (bool, string)
	// StartRecording starts recording audio until stopped or timeout (up to 300 seconds)
	StartRecording(ctx context.Context, onStatus func(string)) (*ActiveRecording, error)
	// Transcribe converts recorded audio into text
	Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error)
	// RecordAndTranscribe records and transcribes in one call (for backward compatibility)
	RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error)
}

// TermuxAPIEngine uses Termux:API tools
type TermuxAPIEngine struct{}

func (t *TermuxAPIEngine) Name() string {
	return "Termux:API Speech Recognition"
}

func (t *TermuxAPIEngine) IsAvailable() (bool, string) {
	if _, err := exec.LookPath("termux-speech-to-text"); err == nil {
		return true, "termux-speech-to-text binary found"
	}
	if _, err := exec.LookPath("termux-microphone-record"); err == nil {
		return true, "termux-microphone-record binary found"
	}
	if _, err := exec.LookPath("termux-dialog"); err == nil {
		return true, "termux-dialog binary found"
	}
	return false, "Termux:API not installed (pkg install termux-api)"
}

func (t *TermuxAPIEngine) StartRecording(ctx context.Context, onStatus func(string)) (*ActiveRecording, error) {
	if _, err := exec.LookPath("termux-microphone-record"); err == nil {
		tmpAudio := fmt.Sprintf("/tmp/tertic_voice_%d.wav", time.Now().UnixNano())
		// termux-microphone-record -f <file> -l 300 (records up to 5 mins until stopped with -q)
		cmd := exec.CommandContext(ctx, "termux-microphone-record", "-f", tmpAudio, "-l", "300")
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start termux recording: %w", err)
		}
		if onStatus != nil {
			onStatus("Recording audio (Speak now - Press [ENTER] when done)...")
		}
		return &ActiveRecording{
			AudioPath: tmpAudio,
			Cmd:       cmd,
			IsTermux:  true,
		}, nil
	}

	return nil, fmt.Errorf("termux-microphone-record not available")
}

func (t *TermuxAPIEngine) Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if onStatus != nil {
		onStatus("Transcribing audio...")
	}
	// Fall back to cloud STT or whisper for audio file
	return TranscribeAudioFile(ctx, audioPath, onStatus)
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

	rec, err := t.StartRecording(ctx, onStatus)
	if err != nil {
		return "", err
	}
	// Default wait if caller doesn't manually stop
	time.Sleep(5 * time.Second)
	_ = rec.Stop()
	defer os.Remove(rec.AudioPath)
	return t.Transcribe(ctx, rec.AudioPath, onStatus)
}

// SystemAudioEngine records via arecord / ffmpeg / sox and transcribes via Whisper or Google Speech API
type SystemAudioEngine struct{}

func (s *SystemAudioEngine) Name() string {
	return "System Audio & Speech Engine"
}

func (s *SystemAudioEngine) IsAvailable() (bool, string) {
	recTool := findRecordTool()
	if recTool != "" {
		return true, fmt.Sprintf("Recording utility '%s' ready", recTool)
	}
	return false, "No recording utility found (arecord, ffmpeg, or sox)"
}

func findRecordTool() string {
	if _, err := exec.LookPath("arecord"); err == nil {
		return "arecord"
	}
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return "ffmpeg"
	}
	if _, err := exec.LookPath("sox"); err == nil {
		return "sox"
	}
	return ""
}

func (s *SystemAudioEngine) StartRecording(ctx context.Context, onStatus func(string)) (*ActiveRecording, error) {
	tmpAudio := fmt.Sprintf("/tmp/tertic_voice_%d.wav", time.Now().UnixNano())
	tool := findRecordTool()
	if tool == "" {
		return nil, fmt.Errorf("no audio recording utility found")
	}

	var cmd *exec.Cmd
	switch tool {
	case "arecord":
		// arecord -d 300 (allows up to 5 min, terminates cleanly on SIGINT)
		cmd = exec.CommandContext(ctx, "arecord", "-d", "300", "-f", "cd", "-t", "wav", tmpAudio)
	case "ffmpeg":
		cmd = exec.CommandContext(ctx, "ffmpeg", "-y", "-f", "pulse", "-i", "default", "-t", "300", tmpAudio)
	case "sox":
		cmd = exec.CommandContext(ctx, "rec", "-r", "16000", "-c", "1", tmpAudio, "trim", "0", "300")
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start audio recording: %w", err)
	}

	if onStatus != nil {
		onStatus("Recording audio (Speak now - Press [ENTER] when done)...")
	}

	return &ActiveRecording{
		AudioPath: tmpAudio,
		Cmd:       cmd,
		IsTermux:  false,
	}, nil
}

func (s *SystemAudioEngine) Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	return TranscribeAudioFile(ctx, audioPath, onStatus)
}

func (s *SystemAudioEngine) RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error) {
	rec, err := s.StartRecording(ctx, onStatus)
	if err != nil {
		return "", err
	}
	time.Sleep(5 * time.Second)
	_ = rec.Stop()
	defer os.Remove(rec.AudioPath)
	return s.Transcribe(ctx, rec.AudioPath, onStatus)
}

// TranscribeAudioFile transcribes a WAV audio file using local Whisper CLI or Google Speech Recognition API
func TranscribeAudioFile(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return "", fmt.Errorf("audio file not found: %s", audioPath)
	}

	// 1. Try local Whisper if installed
	whisperBin := ""
	if _, err := exec.LookPath("whisper"); err == nil {
		whisperBin = "whisper"
	} else if _, err := exec.LookPath("whisper-cpp"); err == nil {
		whisperBin = "whisper-cpp"
	}

	if whisperBin != "" {
		if onStatus != nil {
			onStatus("Transcribing with local Whisper...")
		}
		wCmd := exec.CommandContext(ctx, whisperBin, audioPath)
		out, err := wCmd.Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return strings.TrimSpace(string(out)), nil
		}
	}

	// 2. Transcribe via Google Speech Recognition API (FLAC/WAV standard protocol used by SpeechRecognition)
	if onStatus != nil {
		onStatus("Transcribing with Speech Recognition API...")
	}

	// Convert audio to 16kHz mono FLAC or WAV for best recognition compatibility
	flacPath := fmt.Sprintf("/tmp/tertic_flac_%d.flac", time.Now().UnixNano())
	defer os.Remove(flacPath)

	convertCmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", flacPath)
	if err := convertCmd.Run(); err != nil {
		// If ffmpeg fails, try using raw file
		flacPath = audioPath
	}

	audioBytes, err := os.ReadFile(flacPath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	// Google Speech Recognition v1 endpoint (public key standard used by mobile clients)
	url := "https://www.google.com/speech-api/v2/recognize?output=json&lang=en-US&key=AIzaSyA_Z9D22v_bX2gW8sYF6P5-m8qQv0X2b7c"
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(audioBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "audio/x-flac; rate=16000")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("speech API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read API response: %w", err)
	}

	// Google speech-api returns chunks of JSON lines
	lines := strings.Split(string(body), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		var parsed struct {
			Result []struct {
				Alternative []struct {
					Transcript string  `json:"transcript"`
					Confidence float64 `json:"confidence"`
				} `json:"alternative"`
			} `json:"result"`
		}
		if err := json.Unmarshal([]byte(l), &parsed); err == nil && len(parsed.Result) > 0 {
			if len(parsed.Result[0].Alternative) > 0 {
				return parsed.Result[0].Alternative[0].Transcript, nil
			}
		}
	}

	return "", fmt.Errorf("no speech recognized")
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
			&SystemAudioEngine{},
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
