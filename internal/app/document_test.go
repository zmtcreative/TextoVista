package app

import (
	"os"
	"strings"
	"testing"
)

func TestStripFirstH1(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple h1 removal",
			input:    `<p>Some content</p><h1>Title</h1><p>More content</p>`,
			expected: `<p>Some content</p><p>More content</p>`,
		},
		{
			name:     "h1 with attributes",
			input:    `<div>Content</div><h1 id="title" class="header">My Title</h1><p>Rest of content</p>`,
			expected: `<div>Content</div><p>Rest of content</p>`,
		},
		{
			name:     "h1 with nested content",
			input:    `<p>Intro</p><h1><span>Title</span> with <strong>bold</strong></h1><p>Body</p>`,
			expected: `<p>Intro</p><p>Body</p>`,
		},
		{
			name:     "case insensitive h1",
			input:    `<p>Content</p><H1>TITLE</H1><p>More</p>`,
			expected: `<p>Content</p><p>More</p>`,
		},
		{
			name:     "multiple h1 elements - only first removed",
			input:    `<h1>First Title</h1><p>Content</p><h1>Second Title</h1><p>More</p>`,
			expected: `<p>Content</p><h1>Second Title</h1><p>More</p>`,
		},
		{
			name:     "h2 before h1 - no removal",
			input:    `<p>Content</p><h2>Subtitle</h2><h1>Main Title</h1><p>Body</p>`,
			expected: `<p>Content</p><h2>Subtitle</h2><h1>Main Title</h1><p>Body</p>`,
		},
		{
			name:     "h3 before h1 - no removal",
			input:    `<h3>Small Header</h3><p>Content</p><h1>Main Title</h1>`,
			expected: `<h3>Small Header</h3><p>Content</p><h1>Main Title</h1>`,
		},
		{
			name:     "h6 before h1 - no removal",
			input:    `<h6>Tiny Header</h6><h1>Big Title</h1><p>Content</p>`,
			expected: `<h6>Tiny Header</h6><h1>Big Title</h1><p>Content</p>`,
		},
		{
			name:     "no headers at all",
			input:    `<p>Just regular content</p><div>No headers here</div>`,
			expected: `<p>Just regular content</p><div>No headers here</div>`,
		},
		{
			name:     "only h1 present",
			input:    `<h1>Only Title</h1>`,
			expected: ``,
		},
		{
			name:     "h1 at the beginning",
			input:    `<h1>First Thing</h1><p>Content follows</p><h2>Subtitle</h2>`,
			expected: `<p>Content follows</p><h2>Subtitle</h2>`,
		},
		{
			name:     "h1 at the end",
			input:    `<p>Content first</p><div>More content</div><h1>Final Title</h1>`,
			expected: `<p>Content first</p><div>More content</div>`,
		},
		{
			name:     "h1 with line breaks",
			input:    "<p>Content</p>\n<h1>Title with\nLine breaks</h1>\n<p>More</p>",
			expected: "<p>Content</p>\n\n<p>More</p>",
		},
		{
			name:     "h1 with multiple line breaks and nested content",
			input:    "<div>Before</div>\n<h1>Complex\n<span>nested</span>\ncontent</h1>\n<p>After</p>",
			expected: "<div>Before</div>\n\n<p>After</p>",
		},
		{
			name:     "h1 with line breaks and attributes",
			input:    "<p>Start</p>\n<h1 class=\"title\" id=\"main\">Multi-line\nTitle</h1>\n<p>End</p>",
			expected: "<p>Start</p>\n\n<p>End</p>",
		},
		{
			name:     "empty h1",
			input:    `<p>Before</p><h1></h1><p>After</p>`,
			expected: `<p>Before</p><p>After</p>`,
		},
		{
			name:     "h1 with special characters",
			input:    `<div>Content</div><h1>Title & "Quotes" < > &amp;</h1><p>Body</p>`,
			expected: `<div>Content</div><p>Body</p>`,
		},
		{
			name:     "header with attributes before h1 - no removal",
			input:    `<h2 class="subtitle" id="sub">Subtitle</h2><h1>Main</h1><p>Content</p>`,
			expected: `<h2 class="subtitle" id="sub">Subtitle</h2><h1>Main</h1><p>Content</p>`,
		},
		{
			name:     "mixed case headers before h1 - no removal",
			input:    `<H3>Mixed Case</H3><p>Content</p><H1>Title</H1>`,
			expected: `<H3>Mixed Case</H3><p>Content</p><H1>Title</H1>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripFirstH1([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("stripFirstH1() = %q, want %q", string(result), tt.expected)
			}
		})
	}
}

func TestProcessDocumentMetadataEscapesUnsafeTitleSources(t *testing.T) {
	tempFile := createTempDocumentFile(t)

	tests := []struct {
		name                string
		useFrontmatterTitle bool
		extractedTitle      string
		frontmatter         map[string]any
		wantContains        string
		wantNotContains     string
	}{
		{
			name:                "frontmatter title is escaped",
			useFrontmatterTitle: true,
			extractedTitle:      "ignored",
			frontmatter: map[string]any{
				"title": `<img src=x onerror="alert('xss')">`,
			},
			wantContains:    `&lt;img src=x onerror=&#34;alert(&#39;xss&#39;)&#34;&gt;`,
			wantNotContains: `<img src=x onerror="alert('xss')">`,
		},
		{
			name:                "extracted title is escaped",
			useFrontmatterTitle: false,
			extractedTitle:      `<b>unsafe heading</b>`,
			frontmatter:         map[string]any{},
			wantContains:        `&lt;b&gt;unsafe heading&lt;/b&gt;`,
			wantNotContains:     `<b>unsafe heading</b>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := &DocumentProcessor{
				configManager: &ConfigManager{
					config: &Config{
						Application: ApplicationOptions{
							UseFrontmatterTitle: tt.useFrontmatterTitle,
						},
					},
				},
			}

			docTitle, _, _ := dp.processDocumentMetadata(tempFile, tt.extractedTitle, tt.frontmatter)

			if !strings.Contains(docTitle, tt.wantContains) {
				t.Fatalf("processDocumentMetadata() title = %q, want substring %q", docTitle, tt.wantContains)
			}
			if strings.Contains(docTitle, tt.wantNotContains) {
				t.Fatalf("processDocumentMetadata() title = %q, should not contain raw HTML %q", docTitle, tt.wantNotContains)
			}
		})
	}
}

func createTempDocumentFile(t *testing.T) string {
	t.Helper()

	filePath := t.TempDir() + `\document.md`
	if err := os.WriteFile(filePath, []byte("# test\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return filePath
}
