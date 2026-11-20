package ui

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/Ashed2127/tertic/internal/theme"
	"github.com/Ashed2127/tertic/internal/voice"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// GetTerminalWidth returns the column width of the terminal (default 80 if cannot detect)
func GetTerminalWidth() int {
	ws := &winsize{}
	ret, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(ret) == 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	return 80
}

// RenderBanner renders the modern minimalist Tertic header
func RenderBanner(version, provider, modelName, workDir string, inputMode voice.InputMode) {
	th := theme.Current()
	width := GetTerminalWidth()
	if width > 78 {
		width = 78
	}
	if width < 40 {
		width = 40
	}

	divLine := strings.Repeat("─", width)

	// Clean micro ASCII logo with gradient/accent
	logo := `  ████████╗███████╗██████╗ ████████╗██╗ ██████╗
  ╚══██╔══╝██╔════╝██╔══██╗╚══██╔══╝██║██╔════╝
     ██║   █████╗  ██████╔╝   ██║   ██║██║     
     ██║   ██╔══╝  ██╔══██╗   ██║   ██║██║     
     ██║   ███████╗██║  ██║   ██║   ██║╚██████╗
     ╚═╝   ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚═╝ ╚═════╝`

	// On narrow mobile screens (<50 cols), use ultra-compact logo
	if width < 54 {
		logo = "  TERTIC // Autonomous AI Coding Agent"
	}

	fmt.Printf("\n%s%s%s%s\n", th.Primary, th.Bold, logo, th.Reset)
	fmt.Printf("%s%s%s\n", th.Border, divLine, th.Reset)

	// Session Info Grid
	modeBadge := strings.ToUpper(string(inputMode))
	if inputMode == voice.ModeVoice {
		modeBadge = fmt.Sprintf("%s%s[VOICE ACTIVE]%s", th.RecordColor, th.Bold, th.Reset)
	} else if inputMode == voice.ModeHybrid {
		modeBadge = fmt.Sprintf("%s%s[HYBRID (VOICE+TEXT)]%s", th.Secondary, th.Bold, th.Reset)
	} else {
		modeBadge = fmt.Sprintf("%s%s[TEXT]%s", th.Primary, th.Bold, th.Reset)
	}

	shortDir := workDir
	if len(shortDir) > 28 {
		shortDir = "..." + shortDir[len(shortDir)-25:]
	}

	fmt.Printf(" %sModel:%s %s%-18s%s │ %sProvider:%s %s%-12s%s\n",
		th.Bold, th.Reset, th.Primary, modelName, th.Reset,
		th.Bold, th.Reset, th.Secondary, provider, th.Reset)
	fmt.Printf(" %sWorkspace:%s %-14s │ %sInput:%s %s\n",
		th.Muted, th.Reset, shortDir,
		th.Muted, th.Reset, modeBadge)
	fmt.Printf(" %sCommands:%s %s/help%s, %s/model%s, %s/voice%s, %s/theme%s, %s/clear%s, %s/exit%s\n",
		th.Muted, th.Reset,
		th.Primary, th.Reset,
		th.Primary, th.Reset,
		th.Warning, th.Reset,
		th.Primary, th.Reset,
		th.Primary, th.Reset,
		th.Muted, th.Reset)

	fmt.Printf("%s%s%s\n\n", th.Border, divLine, th.Reset)
}

// RenderHelp prints a modern categorized help table
func RenderHelp() {
	th := theme.Current()
	fmt.Printf("\n%s%sTertic Interactive Command Reference:%s\n\n", th.Bold, th.Primary, th.Reset)

	commands := []struct {
		cmd  string
		desc string
	}{
		{"/help", "Display this interactive command manual"},
		{"/model", "Switch between Cloud and Local Offline models"},
		{"/models", "List currently configured models"},
		{"/agents", "Display specialized autonomous agents & roles"},
		{"/status", "Inspect multi-agent runtime & tool mesh status"},
		{"/voice", "Activate voice input mode (speak instructions)"},
		{"/text", "Switch to keyboard text input mode"},
		{"/input", "Show and configure input modes (Text/Voice/Hybrid)"},
		{"/listen", "Start recording voice prompt immediately"},
		{"/stop", "Stop active voice recording and transcribe"},
		{"/theme", "Change terminal color theme (dark/light/mono/contrast)"},
		{"/verbose", "Toggle detailed tool execution output & previews"},
		{"/clear", "Clear terminal screen and redraw welcome view"},
		{"/history", "Display shared session interaction history"},
		{"/context", "Display project directory structure and context"},
		{"/config", "Display active model and system configuration"},
		{"/exit", "Safely exit Tertic session"},
	}

	for _, c := range commands {
		fmt.Printf("  %s%-12s%s %s%s%s\n", th.Primary, c.cmd, th.Reset, th.Muted, c.desc, th.Reset)
	}
	fmt.Println()
}

// RenderInputPrompt renders a distinctive, clean input prompt line
func RenderInputPrompt(mode voice.InputMode) string {
	th := theme.Current()
	switch mode {
	case voice.ModeVoice:
		return fmt.Sprintf("%s%s[voice] ❯%s ", th.RecordColor, th.Bold, th.Reset)
	case voice.ModeHybrid:
		return fmt.Sprintf("%s%s[hybrid] ❯%s ", th.Secondary, th.Bold, th.Reset)
	default:
		return fmt.Sprintf("%s%strtc ❯%s ", th.Primary, th.Bold, th.Reset)
	}
}
