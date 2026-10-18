package htmlsanitize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	gohtml "golang.org/x/net/html"
)

func renderSanitizedMarkdown(t *testing.T, input string) string {
	t.Helper()

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&SanitizeHTMLExtension{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("md.Convert() error = %v", err)
	}
	return buf.String()
}

func TestHTMLSanitizeAsExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic superscript",
			input:    `<sup>10</sup>`,
			expected: `<sup>10</sup>`,
		},
		{
			name:     "Basic subscript",
			input:    `<sub>2</sub>`,
			expected: `<sub>2</sub>`,
		},
		{
			name:     "Superscript in text",
			input:    `This line contains a superscript<sup>10</sup> test.`,
			expected: `This line contains a superscript<sup>10</sup> test.`,
		},
		{
			name:     "Subscript in text",
			input:    `This line contains a subscript<sub>2</sub> test.`,
			expected: `This line contains a subscript<sub>2</sub> test.`,
		},
		{
			name:     "Multiple inline elements",
			input:    `H<sub>2</sub>O and E=mc<sup>2</sup>`,
			expected: `H<sub>2</sub>O and E=mc<sup>2</sup>`,
		},
		{
			name:     "Nested elements",
			input:    `<strong>Bold<sup>10</sup></strong>`,
			expected: `<strong>Bold<sup>10</sup></strong>`,
		},
		{
			name:     "URL A HREF Python script should be disallowed",
			input:    `<a href="https://www.example.com/myapp.py">Python Script</a>`,
			expected: `<a bad-href="https://www.example.com/myapp.py" class="disallowed bad-href" title="(Link to 'https://www.example.com/myapp.py' is disallowed by policy)">Python Script</a>`,
		},
		{
			name:     "URL A HREF Python script in GET string should be disallowed",
			input:    `<a href="https://www.example.com/?prog=myapp.py&arg1=hello&arg2=world">Python Script</a>`,
			expected: `<a bad-href="https://www.example.com/?prog=myapp.py&arg1=hello&arg2=world" class="disallowed bad-href" title="(Link to 'https://www.example.com/?prog=myapp.py&arg1=hello&arg2=world' is disallowed by policy)">Python Script</a>`,
		},
		{
			name:     "URL A HREF Markdown document should be allowed",
			input:    `<a href="https://www.example.com/myapp.md">MyApp.md</a>`,
			expected: `<a href="https://www.example.com/myapp.md" target="_blank">MyApp.md</a>`,
		},
		{
			name:     "URL A HREF with GET parameters should be allowed",
			input:    `<a href="https://www.example.com/myapp?q=42&s=hello">MyApp.md</a>`,
			expected: `<a href="https://www.example.com/myapp?q=42&s=hello" target="_blank">MyApp.md</a>`,
		},
		{
			name:     "URL IMG SRC with executable link should be disallowed",
			input:    `<img src="https://www.example.com/myapp.dll">`,
			expected: `<img bad-src="https://www.example.com/myapp.dll" class="disallowed bad-src" title="(Link to 'https://www.example.com/myapp.dll' is disallowed by policy)">`,
		},
		{
			name:     "URL IMG SRC with image should be allowed",
			input:    `<img src="https://www.example.com/img/foo.png">`,
			expected: `<img src="https://www.example.com/img/foo.png">`,
		},
		{
			name:     "URL IMG SRC with GET parameters should be allowed",
			input:    `<img src="https://www.example.com/myapp?q=42&s=hello">`,
			expected: `<img src="https://www.example.com/myapp?q=42&s=hello">`,
		},
		{
			name:     "URL IMG SRC Check that self-closing tags work properly",
			input:    `<img src="https://www.example.com/img/foo.png"/>`,
			expected: `<img src="https://www.example.com/img/foo.png"/>`,
		},
		{
			name:     "BR - Check that self-closing tags work (no space)",
			input:    `<br/>`,
			expected: `<br/>`,
		},
		{
			name:     "BR - Check that self-closing tags work (extra space removed)",
			input:    `<br />`,
			expected: `<br/>`,
		},
		{
			name:     "Remove FORM element",
			input:    `<form><input type="text"></form>`,
			expected: `<!-- Removed by HTML sanitizer: form -->`,
		},
		{
			name:     "Leave SCRIPT element in HEAD",
			input:    `<head><script>alert("XSS");</script></head>`,
			expected: ``,
		},
		{
			name:     "Remove SCRIPT element in BODY",
			input:    `<body><script>alert("XSS");</script></body>`,
			expected: `<!-- Removed by HTML sanitizer: script -->`,
		},
		{
			name:     "Remove DIALOG element",
			input:    `<dialog><input></dialog>`,
			expected: `<!-- Removed by HTML sanitizer: dialog -->`,
		},
		{
			name:     "Keep allowed elements",
			input:    `<div><span>text</span></div>`,
			expected: `<div><span>text</span></div>`,
		},
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&SanitizeHTMLExtension{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := md.Convert([]byte(tt.input), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			result := buf.String()

			// Remove the paragraph tags for inline tests since Goldmark wraps in <p>
			if strings.HasPrefix(result, "<p>") && strings.HasSuffix(result, "</p>\n") {
				result = strings.TrimSuffix(strings.TrimPrefix(result, "<p>"), "</p>\n")
			}

			if result != tt.expected {
				t.Errorf("Expected: %s", tt.expected)
				t.Errorf("Got:      %s", result)
			}
		})
	}
}

func TestFilterHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple superscript",
			input:    `<sup>10</sup>`,
			expected: `<sup>10</sup>`,
		},
		{
			name:     "Simple subscript",
			input:    `<sub>2</sub>`,
			expected: `<sub>2</sub>`,
		},
		{
			name:     "Remove form element",
			input:    `<form><input type="text"></form>`,
			expected: `<!-- Removed by HTML sanitizer: form -->`,
		},
		{
			name:     "Keep allowed elements",
			input:    `<div><span>text</span></div>`,
			expected: `<div><span>text</span></div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHTML(tt.input)
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain: %s", tt.expected)
				t.Errorf("Got: %s", result)
			}
		})
	}
}

func TestDisallowedLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		shouldContain string
	}{
		{
			name:     "Python script should be disallowed",
			input:    `<a href="https://www.example.com/myapp.py">Python Script</a>`,
			shouldContain: `class="disallowed bad-href"`,
		},
		{
			name:     "Markdown document should be allowed",
			input:    `<a href="https://www.example.com/myapp.md">MyApp.md</a>`,
			shouldContain: `href="https://www.example.com/myapp.md"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterHTML(tt.input)
			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("Expected result to contain: %s", tt.shouldContain)
				t.Errorf("Got: %s", result)
			}
		})
	}
}

func TestMarkdownLinksHandleAnchorsTargetsAndDisallowedURLs(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "internal anchor does not get target blank",
			input:           `[Jump](#section-1)`,
			wantContains:    []string{`href="#section-1"`},
			wantNotContains: []string{`target="_blank"`},
		},
		{
			name:         "external safe link gets target blank",
			input:        `[Docs](https://example.com/readme.md)`,
			wantContains: []string{`href="https://example.com/readme.md"`, `target="_blank"`},
		},
		{
			name:  "disallowed link gets warning title and broken file svg",
			input: `[Installer](https://example.com/setup.exe "Download")`,
			wantContains: []string{
				`class="disallowed bad-href"`,
				`bad-href="https://example.com/setup.exe"`,
				`title="Download (Link to 'https://example.com/setup.exe' is disallowed by policy)"`,
				`<svg`,
			},
			wantNotContains: []string{`target="_blank"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderSanitizedMarkdown(t, tt.input)
			for _, expected := range tt.wantContains {
				if !strings.Contains(result, expected) {
					t.Fatalf("result missing %q\nresult: %s", expected, result)
				}
			}
			for _, forbidden := range tt.wantNotContains {
				if strings.Contains(result, forbidden) {
					t.Fatalf("result unexpectedly contains %q\nresult: %s", forbidden, result)
				}
			}
		})
	}
}

func TestMarkdownImagesRenderAltTitleAndDisallowedState(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:  "allowed image preserves src alt and title",
			input: `![Architecture Diagram](https://example.com/diagram.png "System Diagram")`,
			wantContains: []string{
				`src="https://example.com/diagram.png"`,
				`alt="Architecture Diagram"`,
				`title="System Diagram"`,
			},
		},
		{
			name:  "disallowed image rewrites src and preserves alt",
			input: `![Binary Payload](https://example.com/payload.dll "Payload")`,
			wantContains: []string{
				`class="disallowed bad-src"`,
				`bad-src="https://example.com/payload.dll"`,
				`alt="Binary Payload"`,
				`title="Payload"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderSanitizedMarkdown(t, tt.input)
			for _, expected := range tt.wantContains {
				if !strings.Contains(result, expected) {
					t.Fatalf("result missing %q\nresult: %s", expected, result)
				}
			}
		})
	}
}

func TestAddTargetBlank(t *testing.T) {
	tests := []struct {
		name          string
		node          *gohtml.Node
		wantTarget    bool
		wantTargetVal string
	}{
		{
			name: "external link gets target blank",
			node: &gohtml.Node{Type: gohtml.ElementNode, Data: "a", Attr: []gohtml.Attribute{{Key: attrHref, Val: "https://example.com"}}},
			wantTarget: true,
			wantTargetVal: attrTargetBlank,
		},
		{
			name: "internal anchor does not get target",
			node: &gohtml.Node{Type: gohtml.ElementNode, Data: "a", Attr: []gohtml.Attribute{{Key: attrHref, Val: "#section"}}},
			wantTarget: false,
		},
		{
			name: "existing target is preserved",
			node: &gohtml.Node{Type: gohtml.ElementNode, Data: "a", Attr: []gohtml.Attribute{{Key: attrHref, Val: "https://example.com"}, {Key: attrTarget, Val: "_self"}}},
			wantTarget: true,
			wantTargetVal: "_self",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addTargetBlank(tt.node)
			target := ""
			for _, attr := range tt.node.Attr {
				if attr.Key == attrTarget {
					target = attr.Val
					break
				}
			}
			if tt.wantTarget && target != tt.wantTargetVal {
				t.Fatalf("target = %q, want %q", target, tt.wantTargetVal)
			}
			if !tt.wantTarget && target != "" {
				t.Fatalf("target = %q, want empty", target)
			}
		})
	}
}

func TestExtractExtension(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBase string
		wantExt  string
	}{
		{
			name:     "simple file path",
			input:    "https://example.com/files/readme.md",
			wantBase: "readme.md",
			wantExt:  ".md",
		},
		{
			name:     "query string stays in base but extension remains empty",
			input:    "https://example.com/download?file=setup.exe",
			wantBase: "download?file=setup.exe",
			wantExt:  ".exe",
		},
		{
			name:     "fragment remains part of base when passed directly",
			input:    "manual.pdf#page-2",
			wantBase: "manual.pdf#page-2",
			wantExt:  ".pdf#page-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, ext := extractExtension(tt.input)
			if base != tt.wantBase || ext != tt.wantExt {
				t.Fatalf("extractExtension() = (%q, %q), want (%q, %q)", base, ext, tt.wantBase, tt.wantExt)
			}
		})
	}
}

func TestIsValidFragment(t *testing.T) {
	tests := []struct {
		fragment string
		want     bool
	}{
		{fragment: "", want: true},
		{fragment: "section-1", want: true},
		{fragment: "anchor_name", want: true},
		{fragment: "bad fragment", want: false},
		{fragment: "bad/slash", want: false},
		{fragment: "bad?query", want: false},
	}

	for _, tt := range tests {
		if got := isValidFragment(tt.fragment); got != tt.want {
			t.Fatalf("isValidFragment(%q) = %v, want %v", tt.fragment, got, tt.want)
		}
	}
}

func TestResizeSVG(t *testing.T) {
	original := `<svg width="24" height="24" viewBox="0 0 24 24"></svg>`

	resized := resizeSVG(original, 16, 20)
	if !strings.Contains(resized, `width="16px"`) || !strings.Contains(resized, `height="20px"`) {
		t.Fatalf("resizeSVG() explicit resize = %q", resized)
	}

	square := resizeSVG(original, 0, 18)
	if !strings.Contains(square, `width="18px"`) || !strings.Contains(square, `height="18px"`) {
		t.Fatalf("resizeSVG() square resize = %q", square)
	}
}

func TestHTMLSanitizeSecurityHardening(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		disallow string
		strip    string
	}{
		{
			name:     "Javascript scheme link rewritten to bad-href",
			input:    `<a href="javascript:alert(1)">Click Me</a>`,
			disallow: "bad-href",
		},
		{
			name:     "Data URI scheme link rewritten to bad-href",
			input:    `<a href="data:text/html,<script>alert(1)</script>">Click Me</a>`,
			disallow: "bad-href",
		},
		{
			name:     "Inline onerror handler stripped from img",
			input:    `<img src="https://www.example.com/img/foo.png" onerror="alert(1)"/>`,
			strip:    "onerror",
		},
		{
			name:     "Object element removed",
			input:    `<object data="test.swf"></object>`,
			disallow: "Removed by HTML sanitizer: object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderSanitizedMarkdown(t, tt.input)
			if tt.disallow != "" && !strings.Contains(output, tt.disallow) {
				t.Fatalf("renderSanitizedMarkdown(%q) = %q, expected to contain %q", tt.input, output, tt.disallow)
			}
			if tt.strip != "" && strings.Contains(output, tt.strip) {
				t.Fatalf("renderSanitizedMarkdown(%q) = %q, expected to NOT contain %q", tt.input, output, tt.strip)
			}
		})
	}
}

func TestHTMLSanitizeDataURIs(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSrc    string
		wantBadSrc bool
	}{
		{
			name:       "Markdown image with base64 PNG data URI",
			input:      `![Status](data:image/png;base64,iVBORw0KGgo= "Status")`,
			wantSrc:    `src="data:image/png;base64,iVBORw0KGgo="`,
			wantBadSrc: false,
		},
		{
			name:       "HTML img with base64 PNG data URI",
			input:      `<img src="data:image/png;base64,iVBORw0KGgo=" alt="Status"/>`,
			wantSrc:    `src="data:image/png;base64,iVBORw0KGgo="`,
			wantBadSrc: false,
		},
		{
			name:       "HTML img with base64 JPEG data URI",
			input:      `<img src="data:image/jpeg;base64,/9j/4AAQSkZJRg==" alt="Photo"/>`,
			wantSrc:    `src="data:image/jpeg;base64,/9j/4AAQSkZJRg=="`,
			wantBadSrc: false,
		},
		{
			name:       "Markdown image with data:text/html is disallowed",
			input:      `![XSS](data:text/html,<script>alert(1)</script>)`,
			wantBadSrc: true,
		},
		{
			name:       "HTML img with data:text/html is disallowed",
			input:      `<img src="data:text/html,<script>alert(1)</script>"/>`,
			wantBadSrc: true,
		},
		{
			name:       "HTML img with raw SVG data URI is disallowed",
			input:      `<img src="data:image/svg+xml;utf8,<svg onload=alert(1)>"/>`,
			wantBadSrc: true,
		},
		{
			name:       "Markdown image with SVG data URI is disallowed",
			input:      `![SVG Data](data:image/svg+xml;base64,PHN2Zz4=)`,
			wantBadSrc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderSanitizedMarkdown(t, tt.input)
			if tt.wantBadSrc {
				if !strings.Contains(output, "bad-src=") {
					t.Fatalf("renderSanitizedMarkdown(%q) = %q, expected bad-src attribute", tt.input, output)
				}
			} else {
				if strings.Contains(output, "bad-src=") {
					t.Fatalf("renderSanitizedMarkdown(%q) = %q, expected valid src, got bad-src", tt.input, output)
				}
				if !strings.Contains(output, tt.wantSrc) {
					t.Fatalf("renderSanitizedMarkdown(%q) = %q, expected to contain %q", tt.input, output, tt.wantSrc)
				}
			}
		})
	}
}


