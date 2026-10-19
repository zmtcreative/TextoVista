package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuperSubScriptProcessing(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		enableSuperSub  bool
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:           "basic superscript",
			input:          "x^2^ + y^2^ = z^2^",
			enableSuperSub: true,
			wantContains:   []string{"x<sup>2</sup>", "y<sup>2</sup>", "z<sup>2</sup>"},
		},
		{
			name:           "basic subscript",
			input:          "Water is H~2~O and carbon dioxide is CO~2~",
			enableSuperSub: true,
			wantContains:   []string{"H<sub>2</sub>O", "CO<sub>2</sub>"},
		},
		{
			name:           "strikethrough interoperability with double tildes",
			input:          "Sub: ~sub~ and Strike: ~~deleted~~",
			enableSuperSub: true,
			wantContains:   []string{"<sub>sub</sub>", "<del>deleted</del>"},
			wantNotContains: []string{"<sub>deleted</sub>", "<del>sub</del>"},
		},
		{
			name:           "consecutive words with mixed sub and strikethrough",
			input:          "Chemical C~2~H~4~ and ~~removed item~~ in sentence.",
			enableSuperSub: true,
			wantContains:   []string{"C<sub>2</sub>H<sub>4</sub>", "<del>removed item</del>"},
		},
		{
			name:           "superscript and subscript disabled",
			input:          "Formula: H~2~O, Power: 2^10^, Strike: ~~deleted~~",
			enableSuperSub: false,
			wantContains:   []string{"<del>deleted</del>"},
			wantNotContains: []string{"<sub>2</sub>", "<sup>10</sup>"},
		},
		{
			name:           "no whitespace allowed inside superscript markup",
			input:          "TeX slip: x^2 + 5^ is not a superscript segment.",
			enableSuperSub: true,
			// Per goldmark-supersubscript rules, whitespace inside delimiters prevents superscript parsing
			wantNotContains: []string{"<sup>2 + 5</sup>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultMarkdownTestConfig()
			cfg.useSuperSubScript = tt.enableSuperSub
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

func TestSuperSubScriptDocsFile(t *testing.T) {
	cfg := defaultMarkdownTestConfig()
	cfg.useSuperSubScript = true

	// Read docs/test-inline-formatting.md
	docPath := filepath.Join("..", "..", "docs", "test-inline-formatting.md")
	content, err := os.ReadFile(docPath)
	if err != nil {
		t.Skipf("Skipping docs file test, could not read %s: %v", docPath, err)
	}

	html, _, _ := convertMarkdownWithConfig(t, cfg, string(content))
	if !strings.Contains(html, "C<sub>2</sub>H<sub>4</sub>") {
		t.Errorf("HTML missing C<sub>2</sub>H<sub>4</sub>\nHTML snippet around C2H4:\n%s", html)
	}
	if !strings.Contains(html, "<sup>10</sup>") {
		t.Errorf("HTML missing <sup>10</sup> from SUP/SUB section")
	}
}

