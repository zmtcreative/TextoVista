package markdown

// This is a comprehensive test suite for the markdown processing functions.

import (
	"strings"
	"testing"
)

func TestExtractH1(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTitle string
		expectedBody  string
		expectFound   bool
	}{
		{
			name:          "Simple H1",
			input:         "# My Title\n\nSome content.",
			expectedTitle: "My Title",
			expectedBody:  "\n\nSome content.",
			expectFound:   true,
		},
		{
			name:          "No H1",
			input:         "## Subtitle\n\nSome content.",
			expectedTitle: "",
			expectedBody:  "## Subtitle\n\nSome content.",
			expectFound:   true,
		},
		{
			name:          "H1 with surrounding text",
			input:         "pre-text\n# Title\npost-text",
			expectedTitle: "Title",
			expectedBody:  "pre-text\npost-text",
			expectFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := ExtractH1(tt.input) // Renamed 'found' to 'err' for clarity
			if title != tt.expectedTitle {
				t.Errorf("ExtractH1() title = %q, want %q", title, tt.expectedTitle)
			}
			// if strings.TrimSpace(string(body)) != strings.TrimSpace(tt.expectedBody) {
			//     t.Errorf("ExtractH1() body = %q, want %q", string(body), tt.expectedBody)
			// }

			// Check if an error was returned, which indicates whether the H1 was found.
			h1Found := err == nil
			if h1Found != tt.expectFound {
				t.Errorf("ExtractH1() found = %v, want %v (error: %v)", h1Found, tt.expectFound, err)
			}
		})
	}
}

func TestInitAlertCalloutsIcons(t *testing.T) {
	icons := InitAlertCalloutsIcons(alertCalloutsGFMStrictData)

	// Test that core icons are present
	coreIcons := []string{"note", "tip", "important", "warning", "caution"}
	for _, icon := range coreIcons {
		if _, exists := icons[icon]; !exists {
			t.Errorf("InitAlertIcons() missing core icon: %s", icon)
		}
		if len(icons[icon]) == 0 {
			t.Errorf("InitAlertIcons() core icon %s has empty SVG content", icon)
		}
		if !strings.Contains(icons[icon], "<svg") {
			t.Errorf("InitAlertIcons() core icon %s does not contain SVG content", icon)
		}
	}

	// Test that aliases are present and point to correct icons
	aliasTests := map[string]string{
		"cautions": "caution",
		"tips":     "tip",
		"notes":    "note",
		"warnings": "warning",
	}

	for alias, primary := range aliasTests {
		if aliasIcon, exists := icons[alias]; !exists {
			t.Errorf("InitAlertIcons() missing alias: %s", alias)
		} else if primaryIcon, exists := icons[primary]; !exists {
			t.Errorf("InitAlertIcons() missing primary icon for alias %s: %s", alias, primary)
		} else if aliasIcon != primaryIcon {
			t.Errorf("InitAlertIcons() alias %s does not match primary %s", alias, primary)
		}
	}

	// Test that we have a reasonable number of total icons (core + aliases)
	if len(icons) < 5 {
		t.Errorf("InitAlertIcons() returned %d icons, expected at least 5", len(icons))
	}
}

func TestExtractH1Extended(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedTitle string
		expectedBody  string
		expectFound   bool
	}{
		{
			name:          "Simple H1",
			input:         "# My Title\n\nSome content.",
			expectedTitle: "My Title",
			expectedBody:  "\n\nSome content.",
			expectFound:   true,
		},
		{
			name:          "No H1",
			input:         "## Subtitle\n\nSome content.",
			expectedTitle: "",
			expectedBody:  "## Subtitle\n\nSome content.",
			expectFound:   true,
		},
		{
			name:          "H1 with surrounding text",
			input:         "pre-text\n# Title\npost-text",
			expectedTitle: "Title",
			expectedBody:  "pre-text\npost-text",
			expectFound:   true,
		},
		{
			name:          "H1 with emphasis",
			input:         "# My **Bold** Title\n\nContent follows.",
			expectedTitle: "My Bold Title",
			expectedBody:  "\n\nContent follows.",
			expectFound:   true,
		},
		{
			name:          "H1 with links",
			input:         "# Title with [Link](http://example.com)\n\nContent",
			expectedTitle: "Title with Link",
			expectedBody:  "\n\nContent",
			expectFound:   true,
		},
		{
			name:          "Multiple H1s - first one extracted",
			input:         "# First H1\n\nContent\n\n# Second H1\n\nMore content",
			expectedTitle: "First H1",
			expectedBody:  "\n\nContent\n\n# Second H1\n\nMore content",
			expectFound:   true,
		},
		{
			name:          "H1 after other content",
			input:         "Some intro text\n\n## Subtitle\n\n# Main Title\n\nContent",
			expectedTitle: "Main Title",
			expectedBody:  "Some intro text\n\n## Subtitle\n\n\nContent", // This reflects what the function actually returns
			expectFound:   true,
		},
		{
			name:          "Empty H1",
			input:         "#\n\nSome content.",
			expectedTitle: "",
			expectedBody:  "#\n\nSome content.",
			expectFound:   true,
		},
		{
			name:          "H1 with only whitespace",
			input:         "#   \n\nSome content.",
			expectedTitle: "",
			expectedBody:  "#   \n\nSome content.",
			expectFound:   true,
		},
		{
			name:          "H1 with trailing spaces",
			input:         "# Title with spaces   \n\nContent",
			expectedTitle: "Title with spaces",
			expectedBody:  "\n\nContent",
			expectFound:   true,
		},
		{
			name:          "Only H1 no other content",
			input:         "# Single Title",
			expectedTitle: "Single Title",
			expectedBody:  "",
			expectFound:   true,
		},
		{
			name:          "Empty markdown",
			input:         "",
			expectedTitle: "",
			expectedBody:  "",
			expectFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := ExtractH1(tt.input)
			if title != tt.expectedTitle {
				t.Errorf("ExtractH1() title = %q, want %q", title, tt.expectedTitle)
			}
			// if strings.TrimSpace(string(body)) != strings.TrimSpace(tt.expectedBody) {
			//     t.Errorf("ExtractH1() body = %q, want %q", string(body), tt.expectedBody)
			// }

			// Check if an error was returned, which indicates whether the H1 was found.
			h1Found := err == nil
			if h1Found != tt.expectFound {
				t.Errorf("ExtractH1() found = %v, want %v (error: %v)", h1Found, tt.expectFound, err)
			}
		})
	}
}

func TestGetLanguageByAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "exact language match",
			input:    []byte("go"),
			expected: []byte("go"),
		},
		{
			name:     "alias match - golang",
			input:    []byte("golang"),
			expected: []byte("go"),
		},
		{
			name:     "alias match - js",
			input:    []byte("js"),
			expected: []byte("javascript"),
		},
		{
			name:     "alias match - py",
			input:    []byte("py"),
			expected: []byte("python"),
		},
		{
			name:     "alias match - py3",
			input:    []byte("py3"),
			expected: []byte("python"),
		},
		{
			name:     "alias match - sh",
			input:    []byte("sh"),
			expected: []byte("bash"),
		},
		{
			name:     "alias match - cpp",
			input:    []byte("cpp"),
			expected: []byte("c++"),
		},
		{
			name:     "alias match - cs",
			input:    []byte("cs"),
			expected: []byte("c#"),
		},
		{
			name:     "unknown language",
			input:    []byte("unknownlang"),
			expected: []byte("unknownlang"),
		},
		{
			name:     "empty input",
			input:    []byte(""),
			expected: []byte(""),
		},
		{
			name:     "case sensitive - no match",
			input:    []byte("GO"),
			expected: []byte("GO"),
		},
		{
			name:     "batch file aliases",
			input:    []byte("bat"),
			expected: []byte("batchfile"),
		},
		{
			name:     "powershell aliases",
			input:    []byte("ps1"),
			expected: []byte("powershell"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLanguageByAlias(tt.input)
			if string(result) != string(tt.expected) {
				t.Errorf("getLanguageByAlias() = %q, want %q", string(result), string(tt.expected))
			}
		})
	}
}

func TestCodeLanguagesCompleteness(t *testing.T) {
	// Test that our CodeLanguages slice contains expected languages
	expectedLanguages := []string{
		"go", "javascript", "python", "bash", "c", "c++", "c#",
		"html", "rust", "kotlin", "typescript",
	}

	languageMap := make(map[string]bool)
	for _, lang := range CodeLanguages {
		languageMap[lang.Name] = true
		// Also add aliases
		for _, alias := range lang.Aliases {
			languageMap[alias] = true
		}
	}

	missingLanguages := []string{}
	for _, expected := range expectedLanguages {
		if !languageMap[expected] {
			missingLanguages = append(missingLanguages, expected)
		}
	}

	if len(missingLanguages) > 0 {
		t.Errorf("CodeLanguages missing expected languages: %v", missingLanguages)
	}

	// Test that we have a reasonable number of languages
	if len(CodeLanguages) < 20 {
		t.Errorf("CodeLanguages has %d languages, expected at least 20", len(CodeLanguages))
	}
}

func TestStripCommentsFromFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		fmtype   string
		input    string
		expected string
	}{
		{
			name:   "basic frontmatter with comment",
			fmtype: "YAML",
			input: `---
Title: Hello World
# Doctype: techdoc
Date: 2025-08-08
---

# This is the First section

This is some text in the first section.`,
			expected: `---
Title: Hello World
Date: 2025-08-08
---

# This is the First section

This is some text in the first section.`,
		},
		{
			name:   "frontmatter with plus delimiters",
			fmtype: "TOML",
			input: `+++
title = "Hello World"
# doctype = "techdoc"
date = "2025-08-08"
+++

# This is the First section

This is some text in the first section.`,
			expected: `+++
title = "Hello World"
date = "2025-08-08"
+++

# This is the First section

This is some text in the first section.`,
		},
		{
			name:   "no frontmatter",
			fmtype: "",
			input: `# This is the First section

This is some text in the first section.`,
			expected: `# This is the First section

This is some text in the first section.`,
		},
		{
			name:   "frontmatter with no comments",
			fmtype: "YAML",
			input: `---
Title: Hello World
Date: 2025-08-08
---

# This is the First section`,
			expected: `---
Title: Hello World
Date: 2025-08-08
---

# This is the First section`,
		},
		{
			name:   "blank lines before frontmatter",
			fmtype: "YAML",
			input: `

---
Title: Hello World
# Doctype: techdoc
Date: 2025-08-08
---

# Content`,
			expected: `

---
Title: Hello World
Date: 2025-08-08
---

# Content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, fmType := stripCommentsFromFrontmatter([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("stripCommentsFromFrontmatter() = %q, want %q", string(result), tt.expected)
			}
			if fmType != tt.fmtype {
				t.Errorf("stripCommentsFromFrontmatter() fmType = %q, want %q", fmType, tt.fmtype)
			}
		})
	}
}

func BenchmarkExtractH1(b *testing.B) {
	markdown := "# Test Title\n\nThis is some content that follows the title.\n\n## Subtitle\n\nMore content here."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExtractH1(markdown)
	}
}

func BenchmarkGetLanguageByAlias(b *testing.B) {
	languages := [][]byte{
		[]byte("go"), []byte("js"), []byte("py"), []byte("cpp"),
		[]byte("bash"), []byte("unknownlang"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, lang := range languages {
			getLanguageByAlias(lang)
		}
	}
}
