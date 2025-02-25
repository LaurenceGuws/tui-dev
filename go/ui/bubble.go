package ui

import (
	"fmt"
	"strings"
)

// wrapText wraps text at a given width.
func wrapText(text string, width int) []string {
	words := strings.Fields(text)
	var lines []string
	var current string
	for _, w := range words {
		if len(current)+len(w)+1 > width {
			lines = append(lines, current)
			current = w
		} else {
			if current == "" {
				current = w
			} else {
				current += " " + w
			}
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// createBubble returns an ASCII bubble with a title around the text.
func createBubble(title, text, color string) string {
	const wrapWidth = 50
	lines := wrapText(text, wrapWidth)
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	titleFormatted := fmt.Sprintf(" %s ", title)
	bubbleWidth := maxWidth + 2
	if len(titleFormatted) > bubbleWidth {
		bubbleWidth = len(titleFormatted)
	}
	titlePadding := (bubbleWidth - len(titleFormatted)) / 2
	topBorder := color + "┌" + strings.Repeat("─", titlePadding) + titleFormatted +
		strings.Repeat("─", bubbleWidth-len(titleFormatted)-titlePadding) + "┐" + "[white]\n"
	var middleLines string
	for _, line := range lines {
		spaces := bubbleWidth - len(line)
		middleLines += color + "│ " + line + strings.Repeat(" ", spaces-1) + "│" + "[white]\n"
	}
	bottomBorder := color + "└" + strings.Repeat("─", bubbleWidth) + "┘" + "[white]"
	return topBorder + middleLines + bottomBorder
}
