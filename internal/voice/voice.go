package voice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

// RecordingState tracks the lifecycle of voice recording
type RecordingState string

const (
	StateIdle         RecordingState = "IDLE"
	StateRecording    RecordingState = "RECORDING"
	StateStopping     RecordingState = "STOPPING"
	StateRecorded     RecordingState = "RECORDED"
	StateTranscribing RecordingState = "TRANSCRIBING"
	StateCompleted    RecordingState = "COMPLETED"
	StateFailed       RecordingState = "FAILED"
)

// GetRecordingsDir returns the dedicated writable directory for audio recordings.
// Uses $HOME/.tertic/recordings/ rather than /tmp to ensure reliability on Termux & Android.
func GetRecordingsDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.Getenv("HOME")
	}
	if home == "" {
		home = "."
	}
	dir := filepath.Join(home, ".tertic", "recordings")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

// GenerateAudioPath returns a unique audio file path inside the safe recordings directory.
func GenerateAudioPath() string {
	dir := GetRecordingsDir()
	return filepath.Join(dir, fmt.Sprintf("voice_%d.wav", time.Now().UnixNano()))
}

// ActiveRecording represents an ongoing recording that can be stopped at any time
type ActiveRecording struct {
	AudioPath string
	Cmd       *exec.Cmd
	IsTermux  bool
	stopped   bool
	mu        sync.Mutex
	StartTime time.Time
}

// Stop stops the ongoing recording process cleanly, waiting for the process to exit
// and flushes audio buffers/headers completely before returning.
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

		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if fi, err := os.Stat(r.AudioPath); err == nil && fi.Size() > 0 {
				time.Sleep(200 * time.Millisecond)
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		return nil
	}

	if r.Cmd != nil && r.Cmd.Process != nil {
		_ = r.Cmd.Process.Signal(syscall.SIGINT)
		done := make(chan error, 1)
		go func() {
			done <- r.Cmd.Wait()
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = r.Cmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(100 * time.Millisecond)
			_ = r.Cmd.Process.Kill()
		}
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}

// ValidateAudioFile verifies that the audio file exists, is non-empty, and has valid header/size
func ValidateAudioFile(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("audio file was not generated: %s", path)
		}
		return fmt.Errorf("cannot access audio file: %w", err)
	}

	if fi.Size() < 44 {
		return fmt.Errorf("audio file is empty or corrupted (size: %d bytes)", fi.Size())
	}

	return nil
}

// Engine defines speech capture & transcription abstraction
type Engine interface {
	Name() string
	IsAvailable() (bool, string)
	HasLiveCapture() bool
	LiveCapture(ctx context.Context, onStatus func(string)) (string, error)
	StartRecording(ctx context.Context, onStatus func(string)) (*ActiveRecording, error)
	Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error)
	RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error)
}

// TermuxAPIEngine uses Android's native speech recognition or microphone capture in Termux
type TermuxAPIEngine struct{}

func (t *TermuxAPIEngine) Name() string {
	return "Android Native Speech Recognition (Termux:API)"
}

func (t *TermuxAPIEngine) IsAvailable() (bool, string) {
	if _, err := exec.LookPath("termux-speech-to-text"); err == nil {
		return true, "termux-speech-to-text ready (Android SpeechRecognizer)"
	}
	if _, err := exec.LookPath("termux-dialog"); err == nil {
		return true, "termux-dialog ready (Android Speech Dialog)"
	}
	if _, err := exec.LookPath("termux-microphone-record"); err == nil {
		return true, "termux-microphone-record ready"
	}
	return false, "Termux:API not installed. Run 'pkg install termux-api' and install the Termux:API app from F-Droid."
}

func (t *TermuxAPIEngine) HasLiveCapture() bool {
	// Native Android speech recognition can capture live speech directly!
	if _, err := exec.LookPath("termux-speech-to-text"); err == nil {
		return true
	}
	if _, err := exec.LookPath("termux-dialog"); err == nil {
		return true
	}
	return false
}

func (t *TermuxAPIEngine) LiveCapture(ctx context.Context, onStatus func(string)) (string, error) {
	if _, err := exec.LookPath("termux-speech-to-text"); err == nil {
		if onStatus != nil {
			onStatus("Listening for speech (Android SpeechRecognizer)...")
		}
		cmd := exec.CommandContext(ctx, "termux-speech-to-text")
		cmd.Stdin = nil
		out, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("Android SpeechRecognizer error: %w (check microphone permission in Android settings)", err)
		}
		res := strings.TrimSpace(string(out))
		if res == "" {
			return "", fmt.Errorf("no speech recognized by Android SpeechRecognizer")
		}
		return res, nil
	}

	if _, err := exec.LookPath("termux-dialog"); err == nil {
		if onStatus != nil {
			onStatus("Opening Android speech dialog...")
		}
		cmd := exec.CommandContext(ctx, "termux-dialog", "speech", "-t", "Speak your instruction to Tertic")
		cmd.Stdin = nil
		out, err := cmd.Output()
		if err == nil {
			var parsed struct {
				Code int    `json:"code"`
				Text string `json:"text"`
			}
			if jErr := json.Unmarshal(out, &parsed); jErr == nil && parsed.Text != "" {
				return parsed.Text, nil
			}
		}
		return "", fmt.Errorf("no speech input received from Android dialog")
	}

	return "", fmt.Errorf("native Android live speech capture tools not found")
}

func (t *TermuxAPIEngine) StartRecording(ctx context.Context, onStatus func(string)) (*ActiveRecording, error) {
	if _, err := exec.LookPath("termux-microphone-record"); err == nil {
		audioPath := GenerateAudioPath()
		cmd := exec.CommandContext(ctx, "termux-microphone-record", "-f", audioPath, "-l", "300")
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start Termux microphone: %w (grant mic permission to Termux:API)", err)
		}
		return &ActiveRecording{
			AudioPath: audioPath,
			Cmd:       cmd,
			IsTermux:  true,
			StartTime: time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("termux-microphone-record binary not found (run 'pkg install termux-api')")
}

func (t *TermuxAPIEngine) Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if err := ValidateAudioFile(audioPath); err != nil {
		return "", err
	}
	return TranscribeAudioFile(ctx, audioPath, onStatus)
}

func (t *TermuxAPIEngine) RecordAndTranscribe(ctx context.Context, onStatus func(string)) (string, error) {
	if t.HasLiveCapture() {
		return t.LiveCapture(ctx, onStatus)
	}
	rec, err := t.StartRecording(ctx, onStatus)
	if err != nil {
		return "", err
	}
	time.Sleep(5 * time.Second)
	_ = rec.Stop()
	defer os.Remove(rec.AudioPath)
	return t.Transcribe(ctx, rec.AudioPath, onStatus)
}

// SystemAudioEngine records via arecord / ffmpeg / sox and transcribes via Whisper or Gemini Cloud
type SystemAudioEngine struct{}

func (s *SystemAudioEngine) Name() string {
	return "System Audio & Cloud Transcription"
}

func (s *SystemAudioEngine) IsAvailable() (bool, string) {
	recTool := findRecordTool()
	if recTool != "" {
		return true, fmt.Sprintf("Recording utility '%s' ready", recTool)
	}
	return false, "No recording utility found (arecord, ffmpeg, or sox)"
}

func (s *SystemAudioEngine) HasLiveCapture() bool {
	return false
}

func (s *SystemAudioEngine) LiveCapture(ctx context.Context, onStatus func(string)) (string, error) {
	return "", fmt.Errorf("live capture only supported on Android native engine")
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
	audioPath := GenerateAudioPath()
	tool := findRecordTool()
	if tool == "" {
		return nil, fmt.Errorf("no audio recording utility found (arecord, ffmpeg, or sox)")
	}

	var cmd *exec.Cmd
	switch tool {
	case "arecord":
		cmd = exec.CommandContext(ctx, "arecord", "-q", "-d", "300", "-f", "cd", "-t", "wav", audioPath)
	case "ffmpeg":
		inputDriver := "default"
		if hasPulseAudio() {
			inputDriver = "pulse"
		} else {
			inputDriver = "alsa"
		}
		cmd = exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-loglevel", "quiet", "-y", "-f", inputDriver, "-i", "default", "-t", "300", audioPath)
	case "sox":
		cmd = exec.CommandContext(ctx, "rec", "-q", "-r", "16000", "-c", "1", audioPath, "trim", "0", "300")
	}

	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s recording: %w", tool, err)
	}

	return &ActiveRecording{
		AudioPath: audioPath,
		Cmd:       cmd,
		IsTermux:  false,
		StartTime: time.Now(),
	}, nil
}

func hasPulseAudio() bool {
	if _, err := exec.LookPath("pactl"); err == nil {
		cmd := exec.Command("pactl", "info")
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return false
}

func (s *SystemAudioEngine) Transcribe(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if err := ValidateAudioFile(audioPath); err != nil {
		return "", err
	}
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

// TranscribeAudioFile converts audio to text prioritizing:
// 1. Android native speech recognition (when on Termux)
// 2. Configured Gemini multimodal transcription (fast, accurate, uses existing Gemini config)
// 3. Local Whisper / whisper.cpp
// 4. Google Speech API (v2) if GOOGLE_SPEECH_API_KEY is configured
func TranscribeAudioFile(ctx context.Context, audioPath string, onStatus func(string)) (string, error) {
	if err := ValidateAudioFile(audioPath); err != nil {
		return "", err
	}

	// 1. Check Gemini Cloud Audio Transcription
	geminiKey := resolveGeminiKey()
	if geminiKey != "" {
		if onStatus != nil {
			onStatus("Transcribing with Gemini...")
		}
		audioBytes, err := os.ReadFile(audioPath)
		if err == nil {
			b64Data := base64.StdEncoding.EncodeToString(audioBytes)
			transcript, tErr := transcribeWithGemini(ctx, geminiKey, b64Data)
			if tErr == nil && len(transcript) > 0 {
				return transcript, nil
			}
		}
	}

	// 2. Check Local Whisper binary (whisper or whisper-cpp)
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
		wCmd.Stdin = nil
		out, err := wCmd.Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return strings.TrimSpace(string(out)), nil
		}
	}

	// 3. Google Speech API (v2) if key is configured
	speechKey := os.Getenv("GOOGLE_SPEECH_API_KEY")
	if speechKey != "" {
		if onStatus != nil {
			onStatus("Transcribing with Google Speech API...")
		}
		flacPath := filepath.Join(GetRecordingsDir(), fmt.Sprintf("flac_%d.flac", time.Now().UnixNano()))
		defer os.Remove(flacPath)

		convertCmd := exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-loglevel", "quiet", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", flacPath)
		convertCmd.Stdin = nil
		convertCmd.Stdout = nil
		convertCmd.Stderr = nil
		_ = convertCmd.Run()

		audioBytes, err := os.ReadFile(flacPath)
		if err == nil {
			endpoint := fmt.Sprintf("https://www.google.com/speech-api/v2/recognize?output=json&lang=en-US&key=%s", speechKey)
			req, rErr := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(audioBytes)))
			if rErr == nil {
				req.Header.Set("Content-Type", "audio/x-flac; rate=16000")
				client := &http.Client{Timeout: 20 * time.Second}
				resp, dErr := client.Do(req)
				if dErr == nil {
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)
					for _, l := range strings.Split(string(body), "\n") {
						var parsed struct {
							Result []struct {
								Alternative []struct {
									Transcript string `json:"transcript"`
								} `json:"alternative"`
							} `json:"result"`
						}
						if jErr := json.Unmarshal([]byte(strings.TrimSpace(l)), &parsed); jErr == nil && len(parsed.Result) > 0 {
							if len(parsed.Result[0].Alternative) > 0 {
								return parsed.Result[0].Alternative[0].Transcript, nil
							}
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no transcription backend active.\n" +
		"To transcribe voice in Tertic, choose one:\n" +
		"  1. Android Native: In Termux run 'pkg install termux-api' and install Termux:API app from F-Droid.\n" +
		"  2. Cloud Gemini: Set GEMINI_API_KEY in ~/.tertic/config.env (or export GEMINI_API_KEY=\"...\").\n" +
		"  3. Local Whisper: Install 'whisper' or 'whisper-cpp' on your system.")
}

func resolveGeminiKey() string {
	if k := os.Getenv("GEMINI_API_KEY"); k != "" {
		return k
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		if data, err := os.ReadFile(filepath.Join(home, ".tertic", "config.env")); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "export GEMINI_API_KEY=") {
					return strings.Trim(strings.TrimPrefix(line, "export GEMINI_API_KEY="), `"' `)
				}
			}
		}
	}
	return ""
}

func transcribeWithGemini(ctx context.Context, apiKey string, b64Audio string) (string, error) {
	// Use current Gemini models that support multimodal audio
	supportedModels := []string{"gemini-2.5-flash", "gemini-2.0-flash", "gemini-1.5-flash"}

	var lastErr error
	for _, m := range supportedModels {
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", m, apiKey)
		payload := map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"parts": []map[string]interface{}{
						{
							"text": "Transcribe this spoken audio verbatim into plain English text. Return ONLY the transcribed text. Do not add quotes, commentary, or markdown.",
						},
						{
							"inlineData": map[string]string{
								"mimeType": "audio/wav",
								"data":     b64Audio,
							},
						},
					},
				},
			},
		}

		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			lastErr = err
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 25 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		var parsed struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &parsed); err != nil {
			lastErr = err
			continue
		}

		if parsed.Error != nil {
			lastErr = fmt.Errorf("Gemini API (%s): %s", m, parsed.Error.Message)
			continue
		}

		if len(parsed.Candidates) > 0 && len(parsed.Candidates[0].Content.Parts) > 0 {
			res := strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text)
			if res != "" {
				return res, nil
			}
		}
	}

	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("Gemini audio transcription returned empty result")
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

	// Prioritize Termux engine when running on Termux, otherwise system engine
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
	sb.WriteString(fmt.Sprintf("Recordings Cache: %s\n", GetRecordingsDir()))

	gemKey := resolveGeminiKey()
	if gemKey != "" {
		sb.WriteString("Cloud STT Backend: Gemini Audio Ready\n")
	} else {
		sb.WriteString("Cloud STT Backend: Not configured (add GEMINI_API_KEY to ~/.tertic/config.env)\n")
	}

	sb.WriteString("\nAvailable Recognition Backends:\n")
	for _, eng := range m.engines {
		avail, details := eng.IsAvailable()
		status := "NOT AVAILABLE"
		if avail {
			status = "READY"
		}
		sb.WriteString(fmt.Sprintf(" • %-42s [%s] (%s)\n", eng.Name(), status, details))
	}
	if m.activeEngine != nil {
		sb.WriteString(fmt.Sprintf("\nActive Engine: %s\n", m.activeEngine.Name()))
	} else {
		sb.WriteString("\nActive Engine: None (Keyboard text input active). To enable voice:\n")
		sb.WriteString("  1. In Termux: pkg install termux-api && install Termux:API app from F-Droid.\n")
		sb.WriteString("  2. In Linux/Desktop: set GEMINI_API_KEY in ~/.tertic/config.env or install whisper.\n")
	}
	return sb.String()
}
