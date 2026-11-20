package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Ashed2127/tertic/internal/theme"
)

// ReadMultiline reads input line-by-line.
// If a line ends with '\' or user enters '"""', it continues multiline input until closed or empty line.
// Regular lines submit immediately on Enter.
// Typing '/' displays contextual slash-command suggestions.
func ReadMultiline(reader *bufio.Reader, promptPrefix string) (string, error) {
	fmt.Print(promptPrefix)

	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	trimmedFirst := strings.TrimRight(firstLine, "\r\n")

	// Check if user initiated slash command suggestions
	if trimmedFirst == "/" {
		ShowSlashSuggestions()
		fmt.Printf("\n%s", promptPrefix)
		secondLine, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		trimmedFirst = strings.TrimRight(secondLine, "\r\n")
	}

	// Multiline delimiter block """
	if strings.HasPrefix(trimmedFirst, `"""`) {
		var lines []string
		content := strings.TrimPrefix(trimmedFirst, `"""`)
		if strings.HasSuffix(content, `"""`) && len(content) >= 3 {
			return strings.TrimSuffix(content, `"""`), nil
		}
		if content != "" {
			lines = append(lines, content)
		}
		th := theme.Current()
		for {
			fmt.Printf("%s... ❯%s ", th.Muted, th.Reset)
			nextLine, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			trimmedNext := strings.TrimRight(nextLine, "\r\n")
			if strings.HasSuffix(trimmedNext, `"""`) {
				lines = append(lines, strings.TrimSuffix(trimmedNext, `"""`))
				break
			}
			lines = append(lines, trimmedNext)
		}
		return strings.Join(lines, "\n"), nil
	}

	// Line continuation using trailing backslash \
	if strings.HasSuffix(trimmedFirst, "\\") {
		var lines []string
		lines = append(lines, strings.TrimSuffix(trimmedFirst, "\\"))
		th := theme.Current()
		for {
			fmt.Printf("%s... ❯%s ", th.Muted, th.Reset)
			nextLine, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			trimmedNext := strings.TrimRight(nextLine, "\r\n")
			if strings.HasSuffix(trimmedNext, "\\") {
				lines = append(lines, strings.TrimSuffix(trimmedNext, "\\"))
			} else {
				lines = append(lines, trimmedNext)
				break
			}
		}
		return strings.Join(lines, "\n"), nil
	}

	return trimmedFirst, nil
}

// ShowSlashSuggestions prints instant command suggestions
func ShowSlashSuggestions() {
	th := theme.Current()
	fmt.Printf("\n%s%sCommand suggestions:%s\n", th.Bold, th.Primary, th.Reset)
	suggestions := []struct {
		cmd  string
		desc string
	}{
		{"/help", "List all available commands"},
		{"/model", "Switch AI model or provider"},
		{"/models", "List configured models"},
		{"/voice", "Activate voice input mode"},
		{"/text", "Switch to keyboard text mode"},
		{"/input", "View and switch input modes"},
		{"/listen", "Record voice instruction now"},
		{"/theme", "Change theme (dark/light/mono/contrast)"},
		{"/verbose", "Toggle detailed execution output"},
		{"/clear", "Clear terminal screen"},
		{"/status", "Check system status"},
		{"/exit", "Exit session"},
	}
	for _, s := range suggestions {
		fmt.Printf("  %s%-10s%s %s%s%s\n", th.Primary, s.cmd, th.Reset, th.Muted, s.desc, th.Reset)
	}
}

// SelectInteractiveWithTheme presents an interactive arrow-key selector respecting active theme
func SelectInteractiveWithTheme(title string, options []string) int {
	fd := int(os.Stdin.Fd())
	var oldState syscall.Termios
	_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
	raw := oldState
	raw.Lflag &^= (syscall.ICANON | syscall.ECHO)
	_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&raw)), 0, 0, 0)

	defer func() {
		_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
		fmt.Print("\033[?25h") // Show cursor
	}()

	fmt.Print("\033[?25l") // Hide cursor
	selected := 0
	th := theme.Current()

	render := func() {
		fmt.Printf("\r\033[%dA", len(options)+1) // move up
		fmt.Printf("%s%s%s\n", th.Bold, title, th.Reset)
		for i, opt := range options {
			if i == selected {
				fmt.Printf("\r\033[K %s%s❯ %s%s\n", th.Bold, th.Primary, opt, th.Reset)
			} else {
				fmt.Printf("\r\033[K   %s%s%s\n", th.Muted, opt, th.Reset)
			}
		}
	}

	fmt.Printf("%s%s%s\n", th.Bold, title, th.Reset)
	for i, opt := range options {
		if i == selected {
			fmt.Printf(" %s%s❯ %s%s\n", th.Bold, th.Primary, opt, th.Reset)
		} else {
			fmt.Printf("   %s%s%s\n", th.Muted, opt, th.Reset)
		}
	}

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n == 1 {
			if buf[0] == '\n' || buf[0] == '\r' {
				break
			}
			if buf[0] == 3 { // Ctrl+C
				os.Exit(0)
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 'A': // UP arrow
				if selected > 0 {
					selected--
					render()
				}
			case 'B': // DOWN arrow
				if selected < len(options)-1 {
					selected++
					render()
				}
			}
		}
	}
	fmt.Println()
	return selected
}
