package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineHighlightProcessing(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		enableHighlight bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "basic inline text highlighting",
			input:           "This is ==highlighted text== in a paragraph.",
			enableHighlight: true,
			wantContains:    []string{"<mark>highlighted text</mark>"},
		},
		{
			name:            "multiple highlights in one sentence",
			input:           "Compare ==first highlight== with ==second highlight==.",
			enableHighlight: true,
			wantContains:    []string{"<mark>first highlight</mark>", "<mark>second highlight</mark>"},
		},
		{
			name:            "highlight combined with bold and italic",
			input:           "==**bold highlight**== and ***==bold italic highlight==***",
			enableHighlight: true,
			wantContains:    []string{"<mark><strong>bold highlight</strong></mark>"},
		},
		{
			name:            "highlight combined with subscript and superscript",
			input:           "Formula ==H~2~O== and power ==2^10^== highlighted.",
			enableHighlight: true,
			wantContains:    []string{"<mark>H<sub>2</sub>O</mark>", "<mark>2<sup>10</sup></mark>"},
		},
		{
			name:            "highlight disabled renders as literal text",
			input:           "This is ==not highlighted== text.",
			enableHighlight: false,
			wantNotContains: []string{"<mark>", "</mark>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultMarkdownTestConfig()
			cfg.useInlineHighlighting = tt.enableHighlight
			html, _, _ := convertMarkdownWithConfig(t, cfg, tt.input)

			for _, want := range tt.wantContains {
				if !strings.Contains(html, want) {
					t.Errorf("HTML missing expected substring %q\nHTML: %s", want, html)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(html, notWant) {
					t.Errorf("HTML contains unexpected substring %q\nHTML: %s", notWant, html)
				}
			}
		})
	}
}

func TestInlineHighlightSampleDoc(t *testing.T) {
	cfg := defaultMarkdownTestConfig()
	cfg.useInlineHighlighting = true

	docPath := filepath.Join("..", "..", "assets", "docs", "sample.md")
	content, err := os.ReadFile(docPath)
	if err != nil {
		t.Skipf("Skipping sample doc test, could not read %s: %v", docPath, err)
	}

	html, _, _ := convertMarkdownWithConfig(t, cfg, string(content))
	if !strings.Contains(html, "<mark>Highlighted text</mark>") {
		t.Errorf("HTML missing <mark>Highlighted text</mark> from sample.md\nHTML: %s", html)
	}
}
