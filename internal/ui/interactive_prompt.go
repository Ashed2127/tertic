package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Ashed2127/tertic/internal/commands"
	"github.com/Ashed2127/tertic/internal/theme"
)

// InteractiveSlashReader handles full character-by-character interactive prompt input
// with real-time slash command discovery, arrow navigation, TAB completion, and ESC closing.
type InteractiveSlashReader struct {
	promptPrefix string
	state        commands.AppState
	registry     *commands.Registry
}

// NewInteractiveSlashReader creates a new reader
func NewInteractiveSlashReader(promptPrefix string, state commands.AppState) *InteractiveSlashReader {
	return &InteractiveSlashReader{
		promptPrefix: promptPrefix,
		state:        state,
		registry:     commands.Default(),
	}
}

// ReadLine performs line reading. If the user begins typing '/', it renders
// the interactive command suggestions overlay dynamically without duplicating lines.
// If stdin is not a terminal, it safely falls back to standard line reading.
func (r *InteractiveSlashReader) ReadLine(fallbackReader *bufio.Reader) (string, error) {
	fd := int(os.Stdin.Fd())
	var oldState syscall.Termios
	ret, _, _ := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
	if int(ret) != 0 {
		// Non-terminal fallback (e.g. pipes, redirect)
		fmt.Print(r.promptPrefix)
		line, err := fallbackReader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	// Terminal raw non-canonical mode
	raw := oldState
	raw.Lflag &^= (syscall.ICANON | syscall.ECHO)
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&raw)), 0, 0, 0)

	defer func() {
		// Always restore terminal mode
		_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
	}()

	var input []rune
	selectedIndex := 0
	menuOpen := false
	renderedMenuLines := 0

	clearMenu := func() {
		if renderedMenuLines > 0 {
			for i := 0; i < renderedMenuLines; i++ {
				// Move down, clear line
				fmt.Print("\033[1B\r\033[2K")
			}
			// Move cursor back up to input line
			fmt.Printf("\033[%dA", renderedMenuLines)
			renderedMenuLines = 0
		}
	}

	renderLineAndMenu := func(matches []commands.Command) {
		th := theme.Current()
		termWidth := GetTerminalWidth()
		if termWidth < 40 {
			termWidth = 40
		}

		// 1. Clear any existing rendered menu below
		clearMenu()

		// 2. Render input prompt + current typed text
		// Move to start of line, clear it, write prompt and input
		fmt.Printf("\r\033[2K%s%s", r.promptPrefix, string(input))

		// 3. Render suggestions below if menu is open
		if menuOpen && len(matches) > 0 {
			if selectedIndex >= len(matches) {
				selectedIndex = len(matches) - 1
			}
			if selectedIndex < 0 {
				selectedIndex = 0
			}

			// We will collect menu lines
			var menuLines []string
			menuLines = append(menuLines, "")
			menuLines = append(menuLines, fmt.Sprintf("  %s%sCOMMANDS%s", th.Bold, th.Border, th.Reset))
			menuLines = append(menuLines, "")

			// Max visible items to fit nicely on mobile screens
			maxItems := 8
			if len(matches) < maxItems {
				maxItems = len(matches)
			}

			// Calculate scroll window around selectedIndex
			startIdx := 0
			if selectedIndex >= maxItems {
				startIdx = selectedIndex - maxItems + 1
			}
			endIdx := startIdx + maxItems
			if endIdx > len(matches) {
				endIdx = len(matches)
			}

			for i := startIdx; i < endIdx; i++ {
				cmd := matches[i]
				desc := cmd.Description
				// Truncate description if terminal is narrow
				maxDescLen := termWidth - 20
				if maxDescLen < 15 {
					maxDescLen = 15
				}
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}

				if i == selectedIndex {
					// Highlighted row
					menuLines = append(menuLines, fmt.Sprintf("  \033[7m %-12s %s \033[0m", cmd.Name, desc))
				} else {
					menuLines = append(menuLines, fmt.Sprintf("  %s%-12s%s %s%s%s", th.Primary, cmd.Name, th.Reset, th.Muted, desc, th.Reset))
				}
			}

			menuLines = append(menuLines, "")
			menuLines = append(menuLines, fmt.Sprintf("  %s↑↓ Navigate  TAB Complete  ENTER Select  ESC Close%s", th.Muted, th.Reset))

			// Print each line of menu below the input line
			for _, mLine := range menuLines {
				fmt.Printf("\n\r\033[2K%s", mLine)
			}
			renderedMenuLines = len(menuLines)

			// Move cursor back to input line, and position at end of input
			fmt.Printf("\033[%dA\r%s%s", renderedMenuLines, r.promptPrefix, string(input))
		}
	}

	// Initial render
	fmt.Printf("\r\033[2K%s", r.promptPrefix)

	buf := make([]byte, 16)
	for {
		curStr := string(input)
		isSlash := strings.HasPrefix(curStr, "/")
		var matches []commands.Command
		if isSlash && menuOpen {
			matches = r.registry.Filter(curStr, r.state)
		}

		n, err := os.Stdin.Read(buf)
		if err != nil {
			clearMenu()
			return curStr, err
		}
		if n == 0 {
			continue
		}

		// Single byte key
		if n == 1 {
			b := buf[0]

			// ENTER
			if b == '\n' || b == '\r' {
				clearMenu()
				if menuOpen && len(matches) > 0 && selectedIndex >= 0 && selectedIndex < len(matches) {
					// Execute the selected command directly
					sel := matches[selectedIndex].Name
					fmt.Printf("\r\033[2K%s%s\n", r.promptPrefix, sel)
					return sel, nil
				}
				fmt.Printf("\r\033[2K%s%s\n", r.promptPrefix, curStr)
				return curStr, nil
			}

			// TAB: autocomplete selected command
			if b == '\t' {
				if menuOpen && len(matches) > 0 && selectedIndex >= 0 && selectedIndex < len(matches) {
					sel := matches[selectedIndex].Name
					input = []rune(sel)
					// Keep menu updated with completed command
					matches = r.registry.Filter(string(input), r.state)
					renderLineAndMenu(matches)
				}
				continue
			}

			// Ctrl+C
			if b == 3 {
				clearMenu()
				fmt.Printf("\r\033[2K%s\n", r.promptPrefix)
				return "", nil
			}

			// Ctrl+D / EOF
			if b == 4 {
				clearMenu()
				if len(input) == 0 {
					return "/exit", nil
				}
				continue
			}

			// Backspace / Delete (127 or 8)
			if b == 127 || b == 8 {
				if len(input) > 0 {
					input = input[:len(input)-1]
					curStr = string(input)
					if strings.HasPrefix(curStr, "/") {
						menuOpen = true
						matches = r.registry.Filter(curStr, r.state)
					} else {
						menuOpen = false
						clearMenu()
					}
					renderLineAndMenu(matches)
				}
				continue
			}

			// ESC (27)
			if b == 27 {
				if menuOpen {
					menuOpen = false
					clearMenu()
					renderLineAndMenu(nil)
				}
				continue
			}

			// Regular printable character
			if b >= 32 && b <= 126 {
				input = append(input, rune(b))
				curStr = string(input)
				if strings.HasPrefix(curStr, "/") {
					menuOpen = true
					selectedIndex = 0
					matches = r.registry.Filter(curStr, r.state)
				} else {
					if menuOpen {
						menuOpen = false
						clearMenu()
					}
				}
				renderLineAndMenu(matches)
				continue
			}
		}

		// Multi-byte ANSI sequences (Arrow keys, etc.)
		if n >= 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 'A': // UP arrow
				if menuOpen && len(matches) > 0 {
					if selectedIndex > 0 {
						selectedIndex--
					} else {
						selectedIndex = len(matches) - 1
					}
					renderLineAndMenu(matches)
				}
			case 'B': // DOWN arrow
				if menuOpen && len(matches) > 0 {
					if selectedIndex < len(matches)-1 {
						selectedIndex++
					} else {
						selectedIndex = 0
					}
					renderLineAndMenu(matches)
				}
			}
			continue
		}
	}
}
