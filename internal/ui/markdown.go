package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ashed2127/tertic/internal/theme"
)

// RenderMarkdown parses basic Markdown and outputs formatted ANSI text
// Handles code blocks, inline code, headings, lists, bold, and URLs
func RenderMarkdown(content string) {
	th := theme.Current()
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	codeBlockLang := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code block toggle ```
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// End of code block
				inCodeBlock = false
				fmt.Printf("  %s──────────────────────────────────────────────────%s\n", th.Border, th.Reset)
			} else {
				// Start of code block
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
				if codeBlockLang == "" {
					codeBlockLang = "code"
				}
				fmt.Printf("\n  %s┌── %s%s%s %s%s\n", th.Border, th.Secondary, codeBlockLang, th.Reset, strings.Repeat("─", 40), th.Reset)
			}
			continue
		}

		if inCodeBlock {
			fmt.Printf("  %s│%s %s%s%s\n", th.Border, th.Reset, th.Text, line, th.Reset)
			continue
		}

		// Headings
		if strings.HasPrefix(line, "# ") {
			fmt.Printf("\n%s%s%s%s\n\n", th.Bold, th.Primary, strings.TrimPrefix(line, "# "), th.Reset)
			continue
		}
		if strings.HasPrefix(line, "## ") {
			fmt.Printf("\n%s%s%s%s\n\n", th.Bold, th.Secondary, strings.TrimPrefix(line, "## "), th.Reset)
			continue
		}
		if strings.HasPrefix(line, "### ") {
			fmt.Printf("\n%s%s%s%s\n", th.Bold, th.Text, strings.TrimPrefix(line, "### "), th.Reset)
			continue
		}

		// Bullet lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.TrimPrefix(item, "* ")
			fmt.Printf("  %s•%s %s\n", th.Primary, th.Reset, formatInlineMarkdown(item, th))
			continue
		}

		// Numbered lists
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.' {
			fmt.Printf("  %s%s%s\n", th.Secondary, trimmed[:2], th.Reset)
			fmt.Printf("    %s\n", formatInlineMarkdown(strings.TrimSpace(trimmed[2:]), th))
			continue
		}

		// Horizontal rule
		if trimmed == "---" || trimmed == "___" || trimmed == "***" {
			fmt.Printf("%s%s%s\n", th.Border, strings.Repeat("─", 50), th.Reset)
			continue
		}

		// Standard text
		fmt.Println(formatInlineMarkdown(line, th))
	}
}

// formatInlineMarkdown formats bold (**text**), backticks (`code`), and clickable URLs
func formatInlineMarkdown(text string, th theme.Theme) string {
	if text == "" {
		return ""
	}

	// Format backticks `code`
	parts := strings.Split(text, "`")
	if len(parts) > 1 {
		var sb strings.Builder
		for i, p := range parts {
			if i%2 == 1 {
				sb.WriteString(fmt.Sprintf("%s%s%s", th.Secondary, p, th.Reset))
			} else {
				sb.WriteString(p)
			}
		}
		text = sb.String()
	}

	// Format bold **text**
	bParts := strings.Split(text, "**")
	if len(bParts) > 1 {
		var sb strings.Builder
		for i, p := range bParts {
			if i%2 == 1 {
				sb.WriteString(fmt.Sprintf("%s%s%s", th.Bold, p, th.Reset))
			} else {
				sb.WriteString(p)
			}
		}
		text = sb.String()
	}

	return text
}

// StreamTypewriter streams text with smooth typing animation
func StreamTypewriter(text string) {
	if text == "" {
		return
	}
	th := theme.Current()

	// If contains markdown codeblocks, render markdown directly for clean aesthetics
	if strings.Contains(text, "```") || strings.Contains(text, "# ") {
		RenderMarkdown(text)
		return
	}

	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		fmt.Printf("%s%s%s", th.Text, string(runes[i]), th.Reset)
		if runes[i] == '\n' {
			time.Sleep(12 * time.Millisecond)
		} else if runes[i] == '.' || runes[i] == '!' || runes[i] == '?' {
			time.Sleep(10 * time.Millisecond)
		} else {
			time.Sleep(3 * time.Millisecond)
		}
	}
	fmt.Println()
}
