package app

import (
	"errors"
	"io"
	"os"
	"unicode/utf8"
)

// BinaryDetector handles binary file detection
type BinaryDetector struct{}

// NewBinaryDetector creates a new BinaryDetector
func NewBinaryDetector() *BinaryDetector {
	return &BinaryDetector{}
}

// IsBinaryFile checks if a file is binary by reading the first 8192 bytes
// and using multiple detection methods including UTF-8 validation
func (bd *BinaryDetector) IsBinaryFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 8KB to check for binary content (increased from 512 bytes)
	buffer := make([]byte, 8192)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	if n == 0 {
		return false, nil // Empty file is considered text
	}

	// Trim buffer to actual read size
	buffer = buffer[:n]

	// Check for UTF-16/UTF-32 BOMs first. If a BOM is present, it's a text file.
	if (n >= 2 && ((buffer[0] == 0xFE && buffer[1] == 0xFF) || (buffer[0] == 0xFF && buffer[1] == 0xFE))) ||
		(n >= 4 && ((buffer[0] == 0x00 && buffer[1] == 0x00 && buffer[2] == 0xFE && buffer[3] == 0xFF) ||
			(buffer[0] == 0xFF && buffer[1] == 0xFE && buffer[2] == 0x00 && buffer[3] == 0x00))) {
		return false, nil
	}

	// Check for null bytes (but allow some exceptions for UTF-16/UTF-32)
	nullCount := 0
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			nullCount++
		}
	}

	// If more than 1% null bytes, likely binary (unless it's UTF-16/UTF-32)
	if float64(nullCount)/float64(n) > 0.01 {
		// Check if it might be UTF-16 or UTF-32 by looking for BOM or patterns
		if !bd.isLikelyUTF16or32(buffer) {
			return true, nil
		}
	}

	// Check if the content is valid UTF-8
	if !bd.isValidUTF8(buffer) {
		// If not valid UTF-8, check if it might be other text encodings
		if !bd.isLikelyTextEncoding(buffer) {
			return true, nil
		}
	}

	// Check for high percentage of control characters (excluding common whitespace)
	controlChars := 0
	for i := 0; i < n; i++ {
		b := buffer[i]
		// Count control characters but exclude common text characters:
		// 9 (tab), 10 (LF), 13 (CR), and anything >= 32 (printable ASCII)
		if b < 32 && b != 9 && b != 10 && b != 13 {
			controlChars++
		}
	}

	// If more than 5% control characters, likely binary
	if n > 0 && float64(controlChars)/float64(n) > 0.05 {
		return true, nil
	}

	// Check for common binary file signatures/magic numbers
	if bd.hasBinarySignature(buffer) {
		return true, nil
	}

	return false, nil
}

// isValidUTF8 checks if the buffer contains valid UTF-8 text
func (bd *BinaryDetector) isValidUTF8(buffer []byte) bool {
	return utf8.Valid(buffer)
}

// isLikelyUTF16or32 checks for UTF-16 or UTF-32 patterns
func (bd *BinaryDetector) isLikelyUTF16or32(buffer []byte) bool {
	if len(buffer) < 4 {
		return false
	}

	// Check for UTF-16 BOM (Byte Order Mark)
	if (buffer[0] == 0xFF && buffer[1] == 0xFE) || // UTF-16 LE
		(buffer[0] == 0xFE && buffer[1] == 0xFF) { // UTF-16 BE
		return true
	}

	// Check for UTF-32 BOM
	if len(buffer) >= 4 {
		if (buffer[0] == 0xFF && buffer[1] == 0xFE && buffer[2] == 0x00 && buffer[3] == 0x00) || // UTF-32 LE
			(buffer[0] == 0x00 && buffer[1] == 0x00 && buffer[2] == 0xFE && buffer[3] == 0xFF) { // UTF-32 BE
			return true
		}
	}

	// Look for UTF-16 patterns (every other byte might be null for ASCII in UTF-16)
	if len(buffer) >= 100 { // Need reasonable sample size
		evenNulls, oddNulls := 0, 0
		for i := 0; i < len(buffer)-1; i += 2 {
			if buffer[i] == 0 {
				evenNulls++
			}
			if buffer[i+1] == 0 {
				oddNulls++
			}
		}

		// If predominantly even or odd positioned nulls, might be UTF-16
		total := len(buffer) / 2
		if total > 0 {
			evenRatio := float64(evenNulls) / float64(total)
			oddRatio := float64(oddNulls) / float64(total)
			// Lowered threshold to 0.4 to catch more cases of mixed ASCII/non-ASCII
			if evenRatio > 0.4 || oddRatio > 0.4 {
				return true
			}
		}
	}

	return false
}

// isLikelyTextEncoding checks for other common text encodings
func (bd *BinaryDetector) isLikelyTextEncoding(buffer []byte) bool {
	// Check for common text file patterns even if not UTF-8
	textIndicators := 0
	totalChars := len(buffer)

	if totalChars == 0 {
		return true
	}

	for _, b := range buffer {
		// Count characters that are likely to appear in text files
		if (b >= 32 && b <= 126) || // Printable ASCII
			b == 9 || b == 10 || b == 13 || // Tab, LF, CR
			(b >= 128) { // Extended ASCII/Latin-1
			textIndicators++
		}
	}

	// If more than 70% of characters look like text, consider it text
	return float64(textIndicators)/float64(totalChars) > 0.70
}

// hasBinarySignature checks for common binary file magic numbers/signatures
func (bd *BinaryDetector) hasBinarySignature(buffer []byte) bool {
	if len(buffer) < 4 {
		return false
	}

	// Common binary file signatures
	signatures := [][]byte{
		{0x89, 0x50, 0x4E, 0x47}, // PNG
		{0xFF, 0xD8, 0xFF},       // JPEG
		{0x47, 0x49, 0x46},       // GIF
		{0x25, 0x50, 0x44, 0x46}, // PDF
		{0x50, 0x4B, 0x03, 0x04}, // ZIP/DOCX/etc
		{0x50, 0x4B, 0x05, 0x06}, // ZIP (empty)
		{0x50, 0x4B, 0x07, 0x08}, // ZIP (spanned)
		{0x52, 0x61, 0x72, 0x21}, // RAR
		{0x7F, 0x45, 0x4C, 0x46}, // ELF (Linux executable)
		{0x4D, 0x5A},             // Windows PE executable
		{0xCA, 0xFE, 0xBA, 0xBE}, // Java class file
		{0xFE, 0xED, 0xFA},       // Mach-O binary (macOS)
		{0x89, 0x48, 0x44, 0x46}, // HDF5
	}

	for _, sig := range signatures {
		if len(buffer) >= len(sig) {
			match := true
			for i, b := range sig {
				if buffer[i] != b {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}

	return false
}
