package uname

import (
	"testing"
)

func TestNormalizeUnameS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Darwin",
			input:    "Darwin",
			expected: "darwin",
		},
		{
			name:     "darwin",
			input:    "darwin",
			expected: "darwin",
		},
		{
			name:     "Linux",
			input:    "Linux",
			expected: "linux",
		},
		{
			name:     "linux",
			input:    "linux",
			expected: "linux",
		},
		{
			name:     "unknown kernel",
			input:    "windows",
			expected: "linux",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeUnameS(tt.input); got != tt.expected {
				t.Errorf("NormalizeUnameS() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNormalizeUnameM(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "arm64",
			input:    "arm64",
			expected: "arm64",
		},
		{
			name:     "ARM64",
			input:    "ARM64",
			expected: "arm64",
		},
		{
			name:     "aarch64",
			input:    "aarch64",
			expected: "arm64",
		},
		{
			name:     "arm",
			input:    "arm",
			expected: "arm64",
		},
		{
			name:     "amd64",
			input:    "amd64",
			expected: "amd64",
		},
		{
			name:     "x86_64",
			input:    "x86_64",
			expected: "amd64",
		},
		{
			name:     "x64",
			input:    "x64",
			expected: "amd64",
		},
		{
			name:     "unknown architecture",
			input:    "i386",
			expected: "amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeUnameM(tt.input); got != tt.expected {
				t.Errorf("NormalizeUnameM() = %v, want %v", got, tt.expected)
			}
		})
	}
}
