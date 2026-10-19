package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTextInsertionProcessing(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		enableInsertion bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "basic text insertion",
			input:           "This is ++inserted text++ in a paragraph.",
			enableInsertion: true,
			wantContains:    []string{"<ins>inserted text</ins>"},
		},
		{
			name:            "multiple insertions in one sentence",
			input:           "Compare ++first insertion++ with ++second insertion++.",
			enableInsertion: true,
			wantContains:    []string{"<ins>first insertion</ins>", "<ins>second insertion</ins>"},
		},
		{
			name:            "insertion combined with bold and italic",
			input:           "++**bold insertion**++ and ***++bold italic insertion++***",
			enableInsertion: true,
			wantContains:    []string{"<ins><strong>bold insertion</strong></ins>"},
		},
		{
			name:            "insertion combined with subscript and superscript",
			input:           "Formula ++H~2~O++ and power ++2^10^++ inserted.",
			enableInsertion: true,
			wantContains:    []string{"<ins>H<sub>2</sub>O</ins>", "<ins>2<sup>10</sup></ins>"},
		},
		{
			name:            "insertion combined with highlighting",
			input:           "Combined ==++highlighted insertion++== test.",
			enableInsertion: true,
			wantContains:    []string{"<mark><ins>highlighted insertion</ins></mark>"},
		},
		{
			name:            "insertion disabled renders as literal text",
			input:           "This is ++not inserted++ text.",
			enableInsertion: false,
			wantNotContains: []string{"<ins>", "</ins>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultMarkdownTestConfig()
			cfg.useTextInsertion = tt.enableInsertion
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

func TestTextInsertionSampleDoc(t *testing.T) {
	cfg := defaultMarkdownTestConfig()
	cfg.useTextInsertion = true

	docPath := filepath.Join("..", "..", "assets", "docs", "sample.md")
	content, err := os.ReadFile(docPath)
	if err != nil {
		t.Skipf("Skipping sample doc test, could not read %s: %v", docPath, err)
	}

	html, _, _ := convertMarkdownWithConfig(t, cfg, string(content))
	if !strings.Contains(html, "<ins>Text Insertion</ins>") {
		t.Errorf("HTML missing <ins>Text Insertion</ins> from sample.md\nHTML: %s", html)
	}
}
