package markdown

// This file was created to expand test coverage without adding more tests to markdown_test.go.

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type markdownTestConfig struct {
	useInlineHTML         bool
	useSanitize           bool
	useStripH1            bool
	useFrontmatter        bool
	useGFM                bool
	usePHPMDExt           bool
	useEmoji              bool
	useMermaid            bool
	useFigure             bool
	useAnchor             bool
	useFences             bool
	useSections           bool
	useHighlighting       bool
	useFancyLists         bool
	useAttributes         bool
	useTypographic        bool
	useAbbreviations      bool
	useKatex              bool
	useD2                 bool
	useSuperSubScript     bool
	useInlineHighlighting bool
	useTextInsertion      bool
	useAlertCallouts      bool
	alertStyle            string
}

func defaultMarkdownTestConfig() markdownTestConfig {
	return markdownTestConfig{
		useInlineHTML:         true,
		useSanitize:           false,
		useStripH1:            false,
		useFrontmatter:        true,
		useGFM:                true,
		usePHPMDExt:           true,
		useEmoji:              true,
		useMermaid:            true,
		useFigure:             true,
		useAnchor:             true,
		useFences:             true,
		useSections:           true,
		useHighlighting:       true,
		useFancyLists:         true,
		useAttributes:         true,
		useTypographic:        true,
		useAbbreviations:      true,
		useKatex:              true,
		useD2:                 true,
		useSuperSubScript:     true,
		useInlineHighlighting: true,
		useTextInsertion:      true,
		useAlertCallouts:      true,
		alertStyle:            "GFMPlus",
	}
}

func (c markdownTestConfig) GetApplicationConfig() (bool, bool) {
	return c.useInlineHTML, c.useSanitize
}
func (c markdownTestConfig) GetAlertCalloutConfig() string { return c.alertStyle }
func (c markdownTestConfig) UseInlineHTML() bool           { return c.useInlineHTML }
func (c markdownTestConfig) UseSanitize() bool             { return c.useSanitize }
func (c markdownTestConfig) UseStripH1() bool              { return c.useStripH1 }
func (c markdownTestConfig) UseFrontmatterTitle() bool     { return c.useFrontmatter }
func (c markdownTestConfig) UseGFM() bool                  { return c.useGFM }
func (c markdownTestConfig) UsePHPMDExt() bool             { return c.usePHPMDExt }
func (c markdownTestConfig) UseEmoji() bool                { return c.useEmoji }
func (c markdownTestConfig) UseMermaid() bool              { return c.useMermaid }
func (c markdownTestConfig) UseFigure() bool               { return c.useFigure }
func (c markdownTestConfig) UseAnchor() bool               { return c.useAnchor }
func (c markdownTestConfig) UseFences() bool               { return c.useFences }
func (c markdownTestConfig) UseSections() bool             { return c.useSections }
func (c markdownTestConfig) UseHighlighting() bool         { return c.useHighlighting }
func (c markdownTestConfig) UseFancyLists() bool           { return c.useFancyLists }
func (c markdownTestConfig) UseAttributes() bool           { return c.useAttributes }
func (c markdownTestConfig) UseTypographic() bool          { return c.useTypographic }
func (c markdownTestConfig) UseAbbreviations() bool        { return c.useAbbreviations }
func (c markdownTestConfig) UseKatex() bool                { return c.useKatex }
func (c markdownTestConfig) UseD2Diagrams() bool           { return c.useD2 }
func (c markdownTestConfig) UseSuperSubScript() bool       { return c.useSuperSubScript }
func (c markdownTestConfig) UseInlineHighlighting() bool   { return c.useInlineHighlighting }
func (c markdownTestConfig) UseTextInsertion() bool        { return c.useTextInsertion }
func (c markdownTestConfig) UseAlertCallouts() bool        { return c.useAlertCallouts }
func (c markdownTestConfig) AlertCalloutStyle() string     { return c.alertStyle }

func convertMarkdownWithConfig(t *testing.T, cfg markdownTestConfig, input string) (string, map[string]any, string) {
	t.Helper()

	md := CreateGoldmarkInstance(cfg)
	html, meta, title, err := ConvertMarkdownToHTML(md, []byte(input), cfg)
	if err != nil {
		t.Fatalf("ConvertMarkdownToHTML() error = %v", err)
	}
	return string(html), meta, title
}

func TestCreateGoldmarkInstanceFeatureMatrix(t *testing.T) {
	tests := []struct {
		name            string
		cfg             markdownTestConfig
		input           string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "inline html disabled escapes raw html",
			cfg: func() markdownTestConfig {
				cfg := defaultMarkdownTestConfig()
				cfg.useInlineHTML = false
				cfg.useSanitize = false
				return cfg
			}(),
			input:           `<span class="raw">inline</span>`,
			wantNotContains: []string{`<span class="raw">inline</span>`},
		},
		{
			name:         "inline html enabled preserves raw html",
			cfg:          defaultMarkdownTestConfig(),
			input:        `<span class="raw">inline</span>`,
			wantContains: []string{`<span class="raw">inline</span>`},
		},
		{
			name:         "gfm task list renders checkbox",
			cfg:          defaultMarkdownTestConfig(),
			input:        "- [x] done",
			wantContains: []string{`type="checkbox"`, `checked=""`},
		},
		{
			name:         "php markdown footnotes render references",
			cfg:          defaultMarkdownTestConfig(),
			input:        "note[^1]\n\n[^1]: footnote body",
			wantContains: []string{`fn:1`, `footnote body`},
		},
		{
			name:         "emoji extension renders unicode emoji",
			cfg:          defaultMarkdownTestConfig(),
			input:        `Emoji :smile:`,
			wantContains: []string{"&#x1f604;"},
		},
		{
			name:         "mermaid extension marks mermaid blocks",
			cfg:          defaultMarkdownTestConfig(),
			input:        "```mermaid\ngraph TD\nA-->B\n```",
			wantContains: []string{"mermaid", "graph TD"},
		},
		{
			name:         "figure extension wraps image",
			cfg:          defaultMarkdownTestConfig(),
			input:        "![Figure caption](image.png \"Figure caption\")",
			wantContains: []string{`<img src="image.png"`, "Figure caption"},
		},
		{
			name:         "anchor extension adds heading anchors",
			cfg:          defaultMarkdownTestConfig(),
			input:        "# Heading",
			wantContains: []string{`href="#heading"`},
		},
		{
			name:         "fences extension preserves custom fenced div content",
			cfg:          defaultMarkdownTestConfig(),
			input:        "::: note\ninside fenced block\n:::",
			wantContains: []string{"inside fenced block"},
		},
		{
			name:         "sections extension wraps heading sections",
			cfg:          defaultMarkdownTestConfig(),
			input:        "# Heading\n\nBody",
			wantContains: []string{"<section", `class="section-h1 h1"`},
		},
		{
			name:            "highlighting and attributes render code wrapper attributes",
			cfg:             defaultMarkdownTestConfig(),
			input:           "```golang {data-test=\"yes\" label=\"Go\" ignored=\"no\"}\nfmt.Println(1)\n```",
			wantContains:    []string{`<pre language="go"`, `data-test="yes"`, `label="Go"`, `class="chroma"`},
			wantNotContains: []string{`ignored="no"`},
		},
		{
			name:         "fancy lists render alphabetic ordered list",
			cfg:          defaultMarkdownTestConfig(),
			input:        "a. alpha\nb. beta",
			wantContains: []string{`<ol class="fancy fl-lcalpha" type="a"`, "alpha", "beta"},
		},
		{
			name:         "attributes extension applies block classes",
			cfg:          defaultMarkdownTestConfig(),
			input:        "Paragraph\n{.callout}",
			wantContains: []string{`class="callout"`},
		},
		{
			name:            "typographic extension converts punctuation and shortcuts",
			cfg:             defaultMarkdownTestConfig(),
			input:           `"quoted" -- text ... and Copyright (c) TextoVista(tm) Registered(r) Section (p) Service(sm) +- 1/2 -> done => end`,
			wantContains:    []string{"&ldquo;quoted&rdquo;", "&hellip;", "&ndash;", "&copy;", "&trade;", "&reg;", "&sect;", "&#8480;", "&plusmn;", "&frac12;", "&rarr;", "&rArr;"},
			wantNotContains: []string{"..."},
		},
		{
			name: "typographic extension disabled leaves shortcuts and punctuation literal",
			cfg: func() markdownTestConfig {
				cfg := defaultMarkdownTestConfig()
				cfg.useTypographic = false
				return cfg
			}(),
			input:           `Copyright (c) 2026 TextoVista(tm) Registered(r) (p) (sm) +- 1/2 -> =>`,
			wantContains:    []string{"(c)", "(tm)", "(r)", "(p)", "(sm)", "+-", "1/2", "-&gt;", "=&gt;"},
			wantNotContains: []string{"&copy;", "&trade;", "&reg;", "&sect;", "&#8480;", "&plusmn;", "&frac12;", "&rarr;", "&rArr;"},
		},
		{
			name:         "katex extension renders katex markup",
			cfg:          defaultMarkdownTestConfig(),
			input:        `$x^2$`,
			wantContains: []string{`\(x^2\)`},
		},
		{
			name:         "d2 extension renders svg output",
			cfg:          defaultMarkdownTestConfig(),
			input:        "```d2\na -> b\n```",
			wantContains: []string{"<svg"},
		},
		{
			name: "sanitize extension filters dangerous raw html",
			cfg: func() markdownTestConfig {
				cfg := defaultMarkdownTestConfig()
				cfg.useSanitize = true
				return cfg
			}(),
			input:           `<script>alert('xss')</script><p>safe</p>`,
			wantContains:    []string{"safe"},
			wantNotContains: []string{"<script>"},
		},
		{
			name:         "superscript and subscript enabled render sup and sub tags",
			cfg:          defaultMarkdownTestConfig(),
			input:        "Water is H~2~O and power is 2^10^.",
			wantContains: []string{"H<sub>2</sub>O", "2<sup>10</sup>"},
		},
		{
			name: "superscript and subscript disabled do not render sup and sub tags",
			cfg: func() markdownTestConfig {
				cfg := defaultMarkdownTestConfig()
				cfg.useSuperSubScript = false
				return cfg
			}(),
			input:           "Water is H~2~O and power is 2^10^.",
			wantNotContains: []string{"<sub>2</sub>", "<sup>10</sup>"},
		},
		{
			name:         "inline text highlighting enabled renders mark tags",
			cfg:          defaultMarkdownTestConfig(),
			input:        "This is ==highlighted text== in markdown.",
			wantContains: []string{"<mark>highlighted text</mark>"},
		},
		{
			name: "inline text highlighting disabled does not render mark tags",
			cfg: func() markdownTestConfig {
				cfg := defaultMarkdownTestConfig()
				cfg.useInlineHighlighting = false
				return cfg
			}(),
			input:           "This is ==highlighted text== in markdown.",
			wantNotContains: []string{"<mark>"},
		},
		{
			name:         "text insertion enabled renders ins tags",
			cfg:          defaultMarkdownTestConfig(),
			input:        "This is ++inserted text++ in markdown.",
			wantContains: []string{"<ins>inserted text</ins>"},
		},
		{
			name: "text insertion disabled does not render ins tags",
			cfg: func() markdownTestConfig {
				cfg := defaultMarkdownTestConfig()
				cfg.useTextInsertion = false
				return cfg
			}(),
			input:           "This is ++inserted text++ in markdown.",
			wantNotContains: []string{"<ins>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, _, _ := convertMarkdownWithConfig(t, tt.cfg, tt.input)
			for _, expected := range tt.wantContains {
				if !strings.Contains(html, expected) {
					t.Fatalf("HTML missing %q\nHTML: %s", expected, html)
				}
			}
			for _, forbidden := range tt.wantNotContains {
				if strings.Contains(html, forbidden) {
					t.Fatalf("HTML unexpectedly contains %q\nHTML: %s", forbidden, html)
				}
			}
		})
	}
}

func TestCreateGoldmarkInstanceAlertCalloutStyles(t *testing.T) {
	styles := []string{"GFMStrict", "GFMWithAliases", "GFMPlus", "Obsidian", "UnexpectedStyle"}
	input := "> [!NOTE]\n> Pay attention."

	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			cfg := defaultMarkdownTestConfig()
			cfg.alertStyle = style
			html, _, _ := convertMarkdownWithConfig(t, cfg, input)
			if !strings.Contains(strings.ToLower(html), "note") {
				t.Fatalf("HTML for style %q did not render a note callout\nHTML: %s", style, html)
			}
			if strings.TrimSpace(html) == "" {
				t.Fatalf("HTML for style %q is empty", style)
			}
		})
	}
}

func TestConvertMarkdownToHTMLReturnsMetadataTitleAndAbbreviations(t *testing.T) {
	cfg := defaultMarkdownTestConfig()
	input := `---
Title: Configured Title
# doctype: commented out
Date: 2026-06-01
---

*[HTML]: Hyper Text Markup Language

#   My   Heading

HTML body text.
`

	html, meta, title := convertMarkdownWithConfig(t, cfg, input)

	if title != " My Heading" && title != "My Heading" {
		t.Fatalf("title = %q, want normalized H1 title", title)
	}
	if got, ok := meta["title"].(string); !ok || got != "Configured Title" {
		t.Fatalf("meta[title] = %#v, want %q", meta["title"], "Configured Title")
	}
	if got, ok := meta["__FMTYPE__"].(string); !ok || got != "YAML" {
		t.Fatalf("meta[__FMTYPE__] = %#v, want %q", meta["__FMTYPE__"], "YAML")
	}
	abbrs, ok := meta["__ABBR__"].(map[string]string)
	if !ok {
		t.Fatalf("meta[__ABBR__] type = %T, want map[string]string", meta["__ABBR__"])
	}
	if abbrs["HTML"] != "Hyper Text Markup Language" {
		t.Fatalf("abbrs[HTML] = %q, want %q", abbrs["HTML"], "Hyper Text Markup Language")
	}
	if strings.Contains(html, "*[HTML]:") {
		t.Fatalf("HTML still contains abbreviation definition line: %s", html)
	}
	if !strings.Contains(html, "HTML body text.") {
		t.Fatalf("HTML missing body text: %s", html)
	}
}

func TestRemoveNodeFromSourceRemovesFirstH1(t *testing.T) {
	source := []byte("before\n# Heading\n\nafter\n")
	doc := defaultMarkdownParserForTest(source)

	var h1Node ast.Node
	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || h1Node != nil {
			return ast.WalkContinue, nil
		}
		heading, ok := node.(*ast.Heading)
		if ok && heading.Level == 1 {
			h1Node = node
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})

	if h1Node == nil {
		t.Fatal("failed to locate H1 node")
	}

	result := RemoveNodeFromSource(source, h1Node)
	if result != "before\n\nafter\n" {
		t.Fatalf("RemoveNodeFromSource() = %q, want %q", result, "before\\n\\nafter\\n")
	}
}

func defaultMarkdownParserForTest(source []byte) ast.Node {
	reader := text.NewReader(source)
	return CreateGoldmarkInstance(defaultMarkdownTestConfig()).Parser().Parse(reader)
}
