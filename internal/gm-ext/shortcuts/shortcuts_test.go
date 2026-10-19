package shortcuts

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestShortcutsExtension(t *testing.T) {
	md := goldmark.New(
		goldmark.WithExtensions(
			NewExtender(),
		),
	)

	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:         "copyright lowercase",
			input:        "Copyright (c) 2026",
			wantContains: []string{"Copyright &copy; 2026"},
		},
		{
			name:         "copyright uppercase",
			input:        "Copyright (C) 2026",
			wantContains: []string{"Copyright &copy; 2026"},
		},
		{
			name:         "registered trademark lowercase",
			input:        "Registered(r)",
			wantContains: []string{"Registered&reg;"},
		},
		{
			name:         "registered trademark uppercase",
			input:        "Registered(R)",
			wantContains: []string{"Registered&reg;"},
		},
		{
			name:         "trademark variations",
			input:        "Acme(tm), Acme(TM), Acme(Tm), Acme(tM)",
			wantContains: []string{"Acme&trade;, Acme&trade;, Acme&trade;, Acme&trade;"},
		},
		{
			name:         "all shortcuts in a paragraph",
			input:        "Copyright (c) 2026 TextoVista(tm). All rights reserved(r).",
			wantContains: []string{"Copyright &copy; 2026 TextoVista&trade;. All rights reserved&reg;."},
		},
		{
			name:         "non-matching parenthesized text remains untouched",
			input:        "Care of (c/o), cat (cat), empty (), number (1), abbreviation (xyz)",
			wantContains: []string{"(c/o)", "(cat)", "()", "(1)", "(xyz)"},
		},
		{
			name:         "code span preserves literal shortcut",
			input:        "Use `(c)` or `(tm)` or `(r)` in markdown.",
			wantContains: []string{"<code>(c)</code>", "<code>(tm)</code>", "<code>(r)</code>"},
		},
		{
			name:         "parenthesized link destination untouched",
			input:        "[Link](https://example.com/c)",
			wantContains: []string{`href="https://example.com/c"`},
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
		})
	}
}
