package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewBinaryDetector(t *testing.T) {
	bd := NewBinaryDetector()
	if bd == nil {
		t.Fatal("NewBinaryDetector() returned nil")
	}
}

func TestIsBinaryFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		content        []byte
		filename       string
		expectedBinary bool
	}{
		{
			name:           "empty file",
			content:        []byte(""),
			filename:       "empty.txt",
			expectedBinary: false,
		},
		{
			name:           "simple text file",
			content:        []byte("Hello, World!\nThis is a text file."),
			filename:       "text.txt",
			expectedBinary: false,
		},
		{
			name:           "markdown file",
			content:        []byte("# Heading\n\nThis is **markdown** content.\n\n* List item 1\n* List item 2"),
			filename:       "test.md",
			expectedBinary: false,
		},
		{
			name:           "file with null bytes",
			content:        []byte("Text\x00with\x00null\x00bytes"),
			filename:       "binary.dat",
			expectedBinary: true,
		},
		{
			name:           "file with many null bytes",
			content:        append([]byte("Start"), make([]byte, 100)...), // 100 null bytes
			filename:       "manynulls.dat",
			expectedBinary: true,
		},
		{
			name:           "file with control characters",
			content:        []byte("Text\x01\x02\x03\x04\x05\x06\x07\x08control"),
			filename:       "control.dat",
			expectedBinary: true,
		},
		{
			name:           "file with valid UTF-8",
			content:        []byte("Hello 世界 🌍 Café naïve résumé"),
			filename:       "utf8.txt",
			expectedBinary: false,
		},
		{
			name:           "file with tabs and newlines",
			content:        []byte("Line 1\nLine 2\n\tIndented line\r\nWindows line ending"),
			filename:       "whitespace.txt",
			expectedBinary: false,
		},
		{
			name:           "JSON file",
			content:        []byte(`{"name": "test", "value": 123, "array": [1, 2, 3]}`),
			filename:       "data.json",
			expectedBinary: false,
		},
		{
			name:           "XML file",
			content:        []byte(`<?xml version="1.0"?><root><item>value</item></root>`),
			filename:       "data.xml",
			expectedBinary: false,
		},
		{
			name:           "UTF-16 BOM file",
			content:        []byte{0xFF, 0xFE, 0x48, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00}, // "Hello" in UTF-16 LE
			filename:       "utf16.txt",
			expectedBinary: false, // BOM indicates text file
		},
		{
			name:           "UTF-16 BE BOM file",
			content:        []byte{0xFE, 0xFF, 0x00, 0x48, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F}, // "Hello" in UTF-16 BE
			filename:       "utf16be.txt",
			expectedBinary: false, // BOM indicates text file
		},
		{
			name:           "very long text line",
			content:        []byte("This is a very long line of text that keeps going and going without any newlines to test handling of long text content"),
			filename:       "longline.txt",
			expectedBinary: false,
		},
	}

	bd := NewBinaryDetector()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(tempDir, tt.filename)
			err := os.WriteFile(filePath, tt.content, 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test binary detection
			isBinary, err := bd.IsBinaryFile(filePath)
			if err != nil {
				t.Fatalf("IsBinaryFile() error = %v", err)
			}

			if isBinary != tt.expectedBinary {
				t.Errorf("IsBinaryFile() = %v, want %v for file %s with content: %q",
					isBinary, tt.expectedBinary, tt.filename, string(tt.content))
			}
		})
	}
}

func TestIsBinaryFileErrors(t *testing.T) {
	bd := NewBinaryDetector()

	// Test non-existent file
	_, err := bd.IsBinaryFile("nonexistent_file.txt")
	if err == nil {
		t.Error("IsBinaryFile() expected error for non-existent file, got nil")
	}

	// Test directory instead of file
	tempDir := t.TempDir()
	_, err = bd.IsBinaryFile(tempDir)
	if err == nil {
		t.Error("IsBinaryFile() expected error for directory, got nil")
	}
}

func TestIsValidUTF8(t *testing.T) {
	bd := NewBinaryDetector()

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "valid ASCII",
			content:  []byte("Hello, World!"),
			expected: true,
		},
		{
			name:     "valid UTF-8",
			content:  []byte("Hello 世界 🌍"),
			expected: true,
		},
		{
			name:     "invalid UTF-8 sequence",
			content:  []byte{0xFF, 0xFE, 0xFD},
			expected: false,
		},
		{
			name:     "partial UTF-8 sequence",
			content:  []byte{0xC0}, // incomplete UTF-8
			expected: false,
		},
		{
			name:     "empty content",
			content:  []byte{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bd.isValidUTF8(tt.content)
			if result != tt.expected {
				t.Errorf("isValidUTF8() = %v, want %v for content: %v", result, tt.expected, tt.content)
			}
		})
	}
}

func TestIsLikelyUTF16or32(t *testing.T) {
	bd := NewBinaryDetector()

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "ASCII text",
			content:  []byte("Hello, World!"),
			expected: false,
		},
		{
			name:     "UTF-16 LE BOM",
			content:  []byte{0xFF, 0xFE, 0x48, 0x00, 0x65, 0x00}, // UTF-16 LE BOM + "He"
			expected: true,
		},
		{
			name:     "UTF-16 BE BOM",
			content:  []byte{0xFE, 0xFF, 0x00, 0x48, 0x00, 0x65}, // UTF-16 BE BOM + "He"
			expected: true,
		},
		{
			name:     "UTF-32 LE BOM",
			content:  []byte{0xFF, 0xFE, 0x00, 0x00, 0x48, 0x00, 0x00, 0x00}, // UTF-32 LE BOM + "H"
			expected: true,
		},
		{
			name:     "UTF-16 pattern large enough",
			content:  append([]byte{0x48, 0x00, 0x65, 0x00, 0x6C, 0x00}, make([]byte, 100)...), // "Hel" + padding
			expected: true,
		},
		{
			name:     "mixed content",
			content:  []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}, // "Hello" in ASCII
			expected: false,
		},
		{
			name:     "short content",
			content:  []byte{0x48},
			expected: false,
		},
		{
			name:     "empty content",
			content:  []byte{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bd.isLikelyUTF16or32(tt.content)
			if result != tt.expected {
				t.Errorf("isLikelyUTF16or32() = %v, want %v for content: %v", result, tt.expected, tt.content)
			}
		})
	}
}

func BenchmarkIsBinaryFile(b *testing.B) {
	// Create a temporary file for benchmarking
	tempDir := b.TempDir()
	filePath := filepath.Join(tempDir, "benchmark.txt")

	// Create a reasonably large text file
	content := make([]byte, 8192)
	for i := range content {
		content[i] = byte('A' + (i % 26))
	}

	err := os.WriteFile(filePath, content, 0644)
	if err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}

	bd := NewBinaryDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bd.IsBinaryFile(filePath)
		if err != nil {
			b.Fatalf("IsBinaryFile() error: %v", err)
		}
	}
}
