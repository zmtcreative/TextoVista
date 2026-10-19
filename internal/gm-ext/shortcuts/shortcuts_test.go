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
		name            string
		input           string
		wantContains    []string
		wantNotContains []string
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
			name:         "section sign variations",
			input:        "Section (p) and (P)",
			wantContains: []string{"Section &sect; and &sect;"},
		},
		{
			name:         "service mark variations",
			input:        "Service(sm), Service(SM), Service(Sm), Service(sM)",
			wantContains: []string{"Service&#8480;, Service&#8480;, Service&#8480;, Service&#8480;"},
		},
		{
			name:         "plus minus symbol",
			input:        "Tolerance: 5V +- 5%",
			wantContains: []string{"Tolerance: 5V &plusmn; 5%"},
		},
		{
			name:         "fractions in prose",
			input:        "Add 1/2 cup sugar, 1/4 tsp salt, and 3/4 cup milk.",
			wantContains: []string{"Add &frac12; cup sugar, &frac14; tsp salt, and &frac34; cup milk."},
		},
		{
			name:            "fraction exclusions for dates and urls",
			input:           "Date is 1/2/2026 or 3/4/2026 at http://example.com/1/2 and 1/23 items.",
			wantContains:    []string{"1/2/2026", "3/4/2026", "http://example.com/1/2", "1/23"},
			wantNotContains: []string{"&frac12;", "&frac34;"},
		},
		{
			name:         "ascii arrows simple",
			input:        "a -> b and c <- d",
			wantContains: []string{"a &rarr; b and c &larr; d"},
		},
		{
			name:         "ascii arrows long and bidirectional",
			input:        "a --> b, c <-- d, e <-> f, g <--> h",
			wantContains: []string{"a &rarr; b, c &larr; d, e &harr; f, g &harr; h"},
		},
		{
			name:         "ascii double arrows",
			input:        "a => b, c ==> d, e <== f, g <=> h",
			wantContains: []string{"a &rArr; b, c &rArr; d, e &lArr; f, g &hArr; h"},
		},
		{
			name:            "arrow exclusions for comparisons and comments",
			input:           "Check if x <= y and x >= y <!-- comment -->",
			wantContains:    []string{"x &lt;= y", "x &gt;= y"},
			wantNotContains: []string{"&larr;", "&rarr;", "&lArr;"},
		},
		{
			name:         "all shortcuts combined in a paragraph",
			input:        "Copyright (c) 2026 TextoVista(tm). Service(sm) (p) 12: +- 1/2 inch -> goal => done.",
			wantContains: []string{"Copyright &copy; 2026 TextoVista&trade;. Service&#8480; &sect; 12: &plusmn; &frac12; inch &rarr; goal &rArr; done."},
		},
		{
			name:         "non-matching parenthesized text remains untouched",
			input:        "Care of (c/o), cat (cat), empty (), number (1), abbreviation (xyz)",
			wantContains: []string{"(c/o)", "(cat)", "()", "(1)", "(xyz)"},
		},
		{
			name:         "code span preserves literal shortcut",
			input:        "Use `(c)`, `(p)`, `(sm)`, `+-`, `1/2`, `->`, `=>` in markdown.",
			wantContains: []string{"<code>(c)</code>", "<code>(p)</code>", "<code>(sm)</code>", "<code>+-</code>", "<code>1/2</code>", "<code>-&gt;</code>", "<code>=&gt;</code>"},
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
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(out, notWant) {
					t.Errorf("Convert(%q) = %q; did not want to contain %q", tt.input, out, notWant)
				}
			}
		})
	}
}
