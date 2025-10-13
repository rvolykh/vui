package common

import "unicode/utf8"

// PadRight pads a string to the right with spaces, accounting for tview color tags
func PadRight(s string, width int) string {
	visibleLen := utf8.RuneCountInString(s) - CountColorTags(s)
	if visibleLen >= width {
		return s
	}
	return s + RepeatString(" ", width-visibleLen)
}

// RepeatString repeats a string n times
func RepeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

// CountColorTags counts the number of characters used by tview color tags
func CountColorTags(s string) int {
	count := 0
	inTag := false
	for _, c := range s {
		if c == '[' {
			inTag = true
		}
		if inTag {
			count++
		}
		if c == ']' {
			inTag = false
		}
	}
	return count
}

// TruncateString truncates a string to the specified length and adds ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// VisibleLength returns the visible length of a string, excluding color tags
func VisibleLength(s string) int {
	return utf8.RuneCountInString(s) - CountColorTags(s)
}
