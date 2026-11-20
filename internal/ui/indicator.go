package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Ashed2127/tertic/internal/theme"
)

// Indicator is an enhanced smart loading and status indicator
type Indicator struct {
	frames     []string
	msg        string
	stop       chan struct{}
	mu         sync.Mutex
	wg         sync.WaitGroup
	enabled    bool
	isVoice    bool
	colorTheme theme.Theme
	stopped    bool
}

// NewIndicator creates an indicator with adaptive animated glyphs
func NewIndicator(msg string) *Indicator {
	cleanMsg := msg
	if len(cleanMsg) > 65 {
		cleanMsg = cleanMsg[:62] + "..."
	}
	th := theme.Current()

	// High-elegance dots/sine pulsing spinner
	frames := []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
	}

	return &Indicator{
		frames:     frames,
		msg:        cleanMsg,
		stop:       make(chan struct{}),
		enabled:    true,
		isVoice:    false,
		colorTheme: th,
	}
}

// NewVoiceIndicator creates a dedicated glowing recording indicator
func NewVoiceIndicator(msg string) *Indicator {
	ind := NewIndicator(msg)
	ind.isVoice = true
	ind.frames = []string{
		"●    ", " ●   ", "  ●  ", "   ● ", "    ●", "   ● ", "  ●  ", " ●   ",
	}
	return ind
}

// Start begins animating on stdout using in-place line rewriting
func (ind *Indicator) Start() {
	if !ind.enabled {
		return
	}
	ind.wg.Add(1)
	go func() {
		defer ind.wg.Done()
		i := 0
		for {
			select {
			case <-ind.stop:
				fmt.Print("\r\033[2K")
				return
			default:
				ind.mu.Lock()
				m := ind.msg
				isV := ind.isVoice
				th := ind.colorTheme
				ind.mu.Unlock()

				frame := ind.frames[i%len(ind.frames)]
				if isV {
					fmt.Printf("\r\033[2K%s%s%s %s%s%s", th.RecordColor, frame, th.Reset, th.Bold, m, th.Reset)
				} else {
					fmt.Printf("\r\033[2K%s%s%s %s%s%s", th.Primary, frame, th.Reset, th.Muted, m, th.Reset)
				}
				i++
				time.Sleep(65 * time.Millisecond)
			}
		}
	}()
}

// Stop stops the animation and outputs an optional final message cleanly without line duplication
func (ind *Indicator) Stop(finalMsg string) {
	if !ind.enabled {
		return
	}
	ind.mu.Lock()
	if ind.stopped {
		ind.mu.Unlock()
		return
	}
	ind.stopped = true
	ind.mu.Unlock()

	close(ind.stop)
	ind.wg.Wait()

	if finalMsg != "" {
		fmt.Printf("\r\033[2K%s\n", finalMsg)
	} else {
		fmt.Print("\r\033[2K")
	}
}

// StopSuccess prints a clean green checkmark indicator
func (ind *Indicator) StopSuccess(msg string) {
	th := theme.Current()
	ind.Stop(fmt.Sprintf("%s✔%s %s%s%s", th.Success, th.Reset, th.Bold, msg, th.Reset))
}

// StopFailure prints a clean red cross indicator
func (ind *Indicator) StopFailure(msg string) {
	th := theme.Current()
	ind.Stop(fmt.Sprintf("%s✖%s %s%s%s", th.Error, th.Reset, th.Bold, msg, th.Reset))
}

// Update updates the in-place display message seamlessly without printing new lines
func (ind *Indicator) Update(newMsg string) {
	if !ind.enabled {
		return
	}
	cleanMsg := newMsg
	if len(cleanMsg) > 65 {
		cleanMsg = cleanMsg[:62] + "..."
	}
	ind.mu.Lock()
	ind.msg = cleanMsg
	ind.mu.Unlock()
}

// FormatActionHumanReadable formats tool actions into clean human descriptions
func FormatActionHumanReadable(toolName string, args map[string]interface{}) string {
	cmd := ""
	if c, ok := args["command"].(string); ok {
		cmd = strings.ToLower(c)
	}
	path := ""
	if p, ok := args["path"].(string); ok {
		path = filepath.Base(p)
	}

	switch toolName {
	case "execute_bash":
		if strings.Contains(cmd, "install") || strings.Contains(cmd, "npm i") || strings.Contains(cmd, "pip i") || strings.Contains(cmd, "pkg in") || strings.Contains(cmd, "go get") {
			return "Installing packages..."
		}
		if strings.Contains(cmd, "termux-open") || strings.Contains(cmd, "xdg-open") || strings.Contains(cmd, "am start") {
			return "Opening application..."
		}
		if strings.Contains(cmd, "build") || strings.Contains(cmd, "compile") || strings.Contains(cmd, "make") {
			return "Building project..."
		}
		if strings.Contains(cmd, "test") || strings.Contains(cmd, "check") {
			return "Testing & verifying..."
		}
		if strings.Contains(cmd, "git") {
			return "Running git operation..."
		}
		if cmd != "" {
			trimmedCmd := strings.TrimSpace(cmd)
			if len(trimmedCmd) > 30 {
				trimmedCmd = trimmedCmd[:27] + "..."
			}
			return fmt.Sprintf("Running: %s", trimmedCmd)
		}
		return "Executing command..."

	case "write_file":
		if path != "" {
			return fmt.Sprintf("Writing: %s", path)
		}
		return "Writing code..."

	case "edit_file":
		if path != "" {
			return fmt.Sprintf("Editing: %s", path)
		}
		return "Refactoring code..."

	case "read_file":
		if path != "" {
			return fmt.Sprintf("Reading: %s", path)
		}
		return "Reading code..."

	case "list_dir":
		if path != "" {
			return fmt.Sprintf("Inspecting directory: %s", path)
		}
		return "Inspecting directory..."

	default:
		return fmt.Sprintf("Executing %s...", toolName)
	}
}
