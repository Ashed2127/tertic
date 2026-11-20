package theme

import (
	"fmt"
	"os"
	"strings"
)

// Theme defines the color and style palette for Tertic
type Theme struct {
	Name        string
	Primary     string // Accent color (Cyan/Blue/Green)
	Secondary   string // Sub-accent (Purple/Yellow)
	Success     string // Positive / OK
	Warning     string // Warnings / Attention
	Error       string // Errors / Failures
	Muted       string // Dim text / Secondary info
	Text        string // Standard text
	Bold        string // Bold prefix
	Reset       string // Reset code
	Border      string // Separator lines
	Highlight   string // Highlighted selection
	RecordColor string // Voice recording indicator
}

// Global active theme
var currentTheme = DarkTheme()

func init() {
	if os.Getenv("NO_COLOR") != "" {
		currentTheme = MonochromeTheme()
	}
}

// Current returns the active theme
func Current() Theme {
	return currentTheme
}

// SetTheme sets the active theme by name
func SetTheme(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "dark":
		currentTheme = DarkTheme()
		return true
	case "light":
		currentTheme = LightTheme()
		return true
	case "monochrome", "mono":
		currentTheme = MonochromeTheme()
		return true
	case "high-contrast", "contrast":
		currentTheme = HighContrastTheme()
		return true
	default:
		return false
	}
}

// AvailableThemes returns list of theme names
func AvailableThemes() []string {
	return []string{"dark", "light", "monochrome", "high-contrast"}
}

// DarkTheme (Default modern DevTools aesthetic)
func DarkTheme() Theme {
	return Theme{
		Name:        "dark",
		Primary:     "\033[38;2;64;196;255m",  // Electric Cyan
		Secondary:   "\033[38;2;180;140;255m", // Soft Purple
		Success:     "\033[38;2;76;209;55m",   // Bright Green
		Warning:     "\033[38;2;251;197;49m",  // Warm Amber
		Error:       "\033[38;2;232;65;24m",   // Red
		Muted:       "\033[38;2;120;144;156m", // Slate Dim
		Text:        "\033[38;2;236;240;241m", // Off-white
		Bold:        "\033[1m",
		Reset:       "\033[0m",
		Border:      "\033[38;2;55;71;79m",    // Subtle divider
		Highlight:   "\033[48;2;38;50;56m\033[38;2;255;255;255m",
		RecordColor: "\033[38;2;255;82;82m",   // Vibrant Red for recording
	}
}

// LightTheme
func LightTheme() Theme {
	return Theme{
		Name:        "light",
		Primary:     "\033[38;5;27m",  // Deep Blue
		Secondary:   "\033[38;5;98m",  // Purple
		Success:     "\033[38;5;28m",  // Dark Green
		Warning:     "\033[38;5;172m", // Dark Yellow/Orange
		Error:       "\033[38;5;160m", // Crimson
		Muted:       "\033[38;5;244m", // Gray
		Text:        "\033[38;5;234m", // Near black
		Bold:        "\033[1m",
		Reset:       "\033[0m",
		Border:      "\033[38;5;250m", // Light gray
		Highlight:   "\033[48;5;254m\033[38;5;232m",
		RecordColor: "\033[38;5;196m",
	}
}

// MonochromeTheme (Respects NO_COLOR or basic terminals)
func MonochromeTheme() Theme {
	return Theme{
		Name:        "monochrome",
		Primary:     "",
		Secondary:   "",
		Success:     "",
		Warning:     "",
		Error:       "",
		Muted:       "",
		Text:        "",
		Bold:        "\033[1m",
		Reset:       "\033[0m",
		Border:      "",
		Highlight:   "",
		RecordColor: "",
	}
}

// HighContrastTheme
func HighContrastTheme() Theme {
	return Theme{
		Name:        "high-contrast",
		Primary:     "\033[1;36m", // Bold Bright Cyan
		Secondary:   "\033[1;35m", // Bold Magenta
		Success:     "\033[1;32m", // Bold Bright Green
		Warning:     "\033[1;33m", // Bold Bright Yellow
		Error:       "\033[1;31m", // Bold Bright Red
		Muted:       "\033[0;37m", // White / Silver
		Text:        "\033[1;37m", // Bold White
		Bold:        "\033[1m",
		Reset:       "\033[0m",
		Border:      "\033[1;30m", // Bold Gray
		Highlight:   "\033[7m",    // Invert video
		RecordColor: "\033[1;41;37m",
	}
}

// Helper formatters
func FmtPrimary(s string) string   { return fmt.Sprintf("%s%s%s", currentTheme.Primary, s, currentTheme.Reset) }
func FmtBold(s string) string      { return fmt.Sprintf("%s%s%s", currentTheme.Bold, s, currentTheme.Reset) }
func FmtSuccess(s string) string   { return fmt.Sprintf("%s%s%s", currentTheme.Success, s, currentTheme.Reset) }
func FmtWarning(s string) string   { return fmt.Sprintf("%s%s%s", currentTheme.Warning, s, currentTheme.Reset) }
func FmtError(s string) string     { return fmt.Sprintf("%s%s%s", currentTheme.Error, s, currentTheme.Reset) }
func FmtMuted(s string) string     { return fmt.Sprintf("%s%s%s", currentTheme.Muted, s, currentTheme.Reset) }
func FmtBorder(s string) string    { return fmt.Sprintf("%s%s%s", currentTheme.Border, s, currentTheme.Reset) }
