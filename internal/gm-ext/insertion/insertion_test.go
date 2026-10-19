package insertion

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"texto-vista/internal/gm-ext/shortcuts"
)

func TestInsertionExtension(t *testing.T) {
	md := goldmark.New(
		goldmark.WithExtensions(
			NewExtender(),
		),
	)

	tests := []struct {
		name            string
		input           string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "basic insertion",
			input:        "This is ++inserted text++ in a paragraph.",
			wantContains: []string{"<ins>inserted text</ins>"},
		},
		{
			name:         "multiple insertions",
			input:        "Compare ++first insertion++ with ++second insertion++.",
			wantContains: []string{"<ins>first insertion</ins>", "<ins>second insertion</ins>"},
		},
		{
			name:         "insertion combined with bold and italic",
			input:        "++**bold insertion**++ and ***++bold italic insertion++***",
			wantContains: []string{"<ins><strong>bold insertion</strong></ins>"},
		},
		{
			name:            "single plus remains literal",
			input:           "Expression 2 + 2 = 4 and +isolated plus+ text.",
			wantContains:    []string{"2 + 2 = 4", "+isolated plus+"},
			wantNotContains: []string{"<ins>"},
		},
		{
			name:            "unmatched delimiters remain literal",
			input:           "This has ++unmatched delimiters in text.",
			wantContains:    []string{"++unmatched delimiters"},
			wantNotContains: []string{"<ins>"},
		},
		{
			name:         "code span preserves literal delimiters",
			input:        "Use `++inserted text++` in markdown.",
			wantContains: []string{"<code>++inserted text++</code>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Convert() failed: %v", err)
			}
			out := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("Convert(%q) = %q; want to contain %q", tt.input, out, want)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(out, notWant) {
					t.Errorf("Convert(%q) = %q; did not want to contain %q", tt.input, out, notWant)
				}
			}
		})
	}
}

func TestInsertionAndShortcutsCoexistence(t *testing.T) {
	md := goldmark.New(
		goldmark.WithExtensions(
			shortcuts.NewExtender(),
			NewExtender(),
		),
	)

	input := "Tolerance +- 5% with ++approved text and +- 10% tolerance++."
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("Convert() failed: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "&plusmn; 5%") {
		t.Errorf("Expected out to contain &plusmn; 5%%, got: %s", out)
	}
	if !strings.Contains(out, "<ins>approved text and &plusmn; 10% tolerance</ins>") {
		t.Errorf("Expected out to contain <ins>approved text and &plusmn; 10%% tolerance</ins>, got: %s", out)
	}
}
