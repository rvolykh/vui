package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			width:    10,
			expected: "hello     ",
		},
		{
			name:     "string with color tags",
			input:    "[yellow]hello[white]",
			width:    10,
			expected: "[yellow]hello[white]     ",
		},
		{
			name:     "string already at width",
			input:    "hello",
			width:    5,
			expected: "hello",
		},
		{
			name:     "string exceeds width",
			input:    "hello world",
			width:    5,
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			width:    5,
			expected: "     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadRight(tt.input, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRepeatString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		count    int
		expected string
	}{
		{
			name:     "repeat dash",
			input:    "-",
			count:    5,
			expected: "-----",
		},
		{
			name:     "repeat word",
			input:    "ab",
			count:    3,
			expected: "ababab",
		},
		{
			name:     "repeat zero times",
			input:    "test",
			count:    0,
			expected: "",
		},
		{
			name:     "repeat negative times",
			input:    "test",
			count:    -1,
			expected: "",
		},
		{
			name:     "repeat empty string",
			input:    "",
			count:    5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RepeatString(tt.input, tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountColorTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "no tags",
			input:    "hello world",
			expected: 0,
		},
		{
			name:     "single color tag",
			input:    "[yellow]hello[white]",
			expected: 15, // [yellow] = 8, [white] = 7
		},
		{
			name:     "multiple tags",
			input:    "[red]error[white] [blue]info[white]",
			expected: 25, // [red] + [white] + [blue] + [white] (counted via actual implementation)
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "nested style tags",
			input:    "[yellow::b]bold text[::-]",
			expected: 16, // [yellow::b] + [::-] (counted via actual implementation)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountColorTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string equal to max",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "string longer than max",
			input:    "hello world this is a long string",
			maxLen:   15,
			expected: "hello world ...",
		},
		{
			name:     "very short max length",
			input:    "hello",
			maxLen:   2,
			expected: "he",
		},
		{
			name:     "max length of 3",
			input:    "hello world",
			maxLen:   3,
			expected: "hel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVisibleLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "plain text",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "text with color tags",
			input:    "[yellow]hello[white]",
			expected: 5,
		},
		{
			name:     "text with multiple tags",
			input:    "[red]error[white]: [blue]message[white]",
			expected: 14, // "error: message" (counted via actual implementation)
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VisibleLength(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
