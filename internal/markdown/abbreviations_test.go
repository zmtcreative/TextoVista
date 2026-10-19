package markdown

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
)

func TestAbbreviationsProcessing(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name: "Basic abbreviation definition and usage",
			markdown: `This refers to ANSI standards.

*[ANSI]: American National Standards Institute`,
			expected: `<p>This refers to <abbr title="American National Standards Institute">ANSI</abbr> standards.</p>`,
		},
		{
			name: "Multiple abbreviations",
			markdown: `This refers to ANSI standards.

The letters in this sentence use the ASCII standard, but could be stored using the UTF-8 standard.

HTML is a very common markup language.

*[ANSI]: American National Standards Institute
*[ASCII]: American Standard Code for Information Interchange
*[UTF-8]: 8-Bit UCS Transformation Format
*[HTML]: HyperText Markup Language`,
			expected: `<p>This refers to <abbr title="American National Standards Institute">ANSI</abbr> standards.</p>
<p>The letters in this sentence use the <abbr title="American Standard Code for Information Interchange">ASCII</abbr> standard, but could be stored using the <abbr title="8-Bit UCS Transformation Format">UTF-8</abbr> standard.</p>
<p><abbr title="HyperText Markup Language">HTML</abbr> is a very common markup language.</p>`,
		},
		{
			name: "Abbreviation with spaces in key",
			markdown: `The API KEY is important.

*[API KEY]: Application Programming Interface Key`,
			expected: `<p>The <abbr title="Application Programming Interface Key">API KEY</abbr> is important.</p>`,
		},
		{
			name: "Multiple line value (should be truncated)",
			markdown: `BAZ is an example of value truncation.

*[BAZ]: This definition has more than one line
listed and the second line will be ignored and rendered as normal text`,
			expected: `<p><abbr title="This definition has more than one line">BAZ</abbr> is an example of value truncation.</p>
<p>listed and the second line will be ignored and rendered as normal text</p>`,
		},
		{
			name: "Abbreviation definitions with leading whitespace",
			markdown: `This refers to ANSI standards.

	*[ANSI]: American National Standards Institute
    *[ASCII]: American Standard Code for Information Interchange`,
			expected: `<p>This refers to <abbr title="American National Standards Institute">ANSI</abbr> standards.</p>`,
		},
		{
			name: "Invalid definitions should not be processed",
			markdown: `FOO and BAR should not be abbreviations.

*[[FOO]]: Cannot have nested square brackets.
*[BAR]:Must have a space between colon and text.`,
			expected: `<p>FOO and BAR should not be abbreviations.</p>
<ul>
<li><code>[[FOO]]</code>  [(<strong>INVALID ABBR DEF:</strong> Cannot have nested brackets)]{.invalid}</li>
<li><code>*[BAR]:Must have a space between colon and text.</code>  [(<strong>INVALID ABBR DEF:</strong> Must have whitespace between <code>:</code> and the value)]{.invalid}</li>
</ul>`,
		},
		{
			name:     "No abbreviations defined",
			markdown: `This is just plain text with no abbreviations.`,
			expected: `<p>This is just plain text with no abbreviations.</p>`,
		},
		{
			name: "Abbreviations in code blocks should be ignored",
			markdown: `Here is some code:

` + "```" + `
ANSI standards are important
` + "```" + `

But ANSI should work here.

*[ANSI]: American National Standards Institute`,
			expected: `<p>Here is some code:</p>
<pre><code>ANSI standards are important
</code></pre>
<p>But <abbr title="American National Standards Institute">ANSI</abbr> should work here.</p>`,
		},
		{
			name: "Inline code should be ignored",
			markdown: `Here is ` + "`ANSI`" + ` in code, but ANSI should work here.

*[ANSI]: American National Standards Institute`,
			expected: `<p>Here is <code>ANSI</code> in code, but <abbr title="American National Standards Institute">ANSI</abbr> should work here.</p>`,
		},
		{
			name: "Case sensitive abbreviations",
			markdown: `Both HTML and html should be different.

*[HTML]: HyperText Markup Language
*[html]: hypertext markup language`,
			expected: `<p>Both <abbr title="HyperText Markup Language">HTML</abbr> and <abbr title="hypertext markup language">html</abbr> should be different.</p>`,
		},
		{
			name: "Special characters in abbreviation value",
			markdown: `This uses XML.

*[XML]: eXtensible Markup Language (with "quotes" & ampersands)`,
			expected: `<p>This uses <abbr title="eXtensible Markup Language (with &#34;quotes&#34; &amp; ampersands)">XML</abbr>.</p>`,
		},
		{
			name: "Multiple occurrences of same abbreviation",
			markdown: `HTML is great. I love HTML. HTML rocks!

*[HTML]: HyperText Markup Language`,
			expected: `<p><abbr title="HyperText Markup Language">HTML</abbr> is great. I love <abbr title="HyperText Markup Language">HTML</abbr>. <abbr title="HyperText Markup Language">HTML</abbr> rocks!</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// // Create base Goldmark processor without abbreviations
			baseMarkdown := goldmark.New()

			abbrDefs, cleanMarkdown := GetMarkdownAbbreviations([]byte(tt.markdown))

			// Process the cleaned markdown with the base Goldmark processor
			var buf strings.Builder
			if err := baseMarkdown.Convert([]byte(cleanMarkdown), &buf); err != nil {
				t.Fatalf("Failed to convert markdown: %v", err)
			}

			html := buf.String()

			// Replace abbreviations in the HTML output
			result := []byte(html)
			if len(abbrDefs) > 0 {
				result = ReplaceAbbreviationsInHTML(result, abbrDefs)
			}

			resultStr := strings.TrimSpace(string(result))
			expected := strings.TrimSpace(tt.expected)

			if resultStr != expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, resultStr)
			}
		})
	}
}

func TestAbbreviationRegexPatterns(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "Valid basic definition",
			line:     "*[HTML]: HyperText Markup Language",
			expected: true,
		},
		{
			name:     "Valid definition with leading spaces",
			line:     "    *[HTML]: HyperText Markup Language",
			expected: true,
		},
		{
			name:     "Valid definition with leading tabs",
			line:     "\t*[HTML]: HyperText Markup Language",
			expected: true,
		},
		{
			name:     "Valid definition with multiple spaces after colon",
			line:     "*[HTML]:   HyperText Markup Language",
			expected: true,
		},
		{
			name:     "Invalid definition starts with non-alphanumeric character",
			line:     "*[_HTML]: HyperText Markup Language",
			expected: false,
		},
		{
			name:     "Invalid definition must be at least 2 characters",
			line:     "*[A]: Abbreviation",
			expected: false,
		},
		{
			name:     "Invalid definition with nested brackets",
			line:     "*[[HTML]]: HyperText Markup Language",
			expected: false,
		},
		{
			name:     "Invalid definition with extra bracket",
			line:     "*[HTML]]: HyperText Markup Language",
			expected: false,
		},
		{
			name:     "Invalid definition without space after colon",
			line:     "*[HTML]:HyperText Markup Language",
			expected: false,
		},
		{
			name:     "Invalid definition without value",
			line:     "*[HTML]: ",
			expected: false,
		},
		{
			name:     "Not an abbreviation definition",
			line:     "This is just regular text",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := abbreviationDefinitionPattern.FindStringSubmatch(tt.line)
			result := matches != nil
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for line: %s", tt.expected, result, tt.line)
			}
		})
	}
}

func TestAbbreviationExtraction(t *testing.T) {
	tests := []struct {
		name                string
		markdown            string
		expectedDefinitions map[string]string
		expectedCleanText   string
	}{
		{
			name: "Basic abbreviation extraction",
			markdown: `This is text.

*[HTML]: HyperText Markup Language

More text.`,
			expectedDefinitions: map[string]string{
				"HTML": "HyperText Markup Language",
			},
			expectedCleanText: `This is text.


More text.`,
		},
		{
			name: "Multiple abbreviations",
			markdown: `Text here.

*[HTML]: HyperText Markup Language
*[CSS]: Cascading Style Sheets

More text.`,
			expectedDefinitions: map[string]string{
				"HTML": "HyperText Markup Language",
				"CSS":  "Cascading Style Sheets",
			},
			expectedCleanText: `Text here.


More text.`,
		},
		{
			name: "No abbreviations",
			markdown: `Just regular text.

No abbreviations here.`,
			expectedDefinitions: map[string]string{},
			expectedCleanText: `Just regular text.

No abbreviations here.`,
		},
		{
			name: "Invalid definitions ignored",
			markdown: `Text here.

*[[HTML]]: Invalid nested brackets
*[CSS]:No space after colon

More text.`,
			expectedDefinitions: map[string]string{},
			expectedCleanText: `Text here.

- ` + "`[[HTML]]`" + ` &nbsp;[(**INVALID ABBR DEF:** Cannot have nested brackets)]{.invalid}
- ` + "`*[CSS]:No space after colon`" + ` &nbsp;[(**INVALID ABBR DEF:** Must have whitespace between ` + "`:`" + ` and the value)]{.invalid}

More text.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			definitions, cleanText := extractAbbreviations(tt.markdown)

			// Check definitions
			if len(definitions) != len(tt.expectedDefinitions) {
				t.Errorf("Expected %d definitions, got %d", len(tt.expectedDefinitions), len(definitions))
			}

			for key, value := range tt.expectedDefinitions {
				if definitions[key] != value {
					t.Errorf("Expected definition[%s] = '%s', got '%s'", key, value, definitions[key])
				}
			}

			// Check clean text
			if cleanText != tt.expectedCleanText {
				t.Errorf("Expected clean text:\n%s\n\nGot:\n%s", tt.expectedCleanText, cleanText)
			}
		})
	}
}

// // func TestCreateAbbreviationProcessor(t *testing.T) {
// 	baseMarkdown := goldmark.New()
// 	processor := CreateAbbreviationProcessor(baseMarkdown)

// 	markdown := []byte(`This refers to ANSI standards.

// *[ANSI]: American National Standards Institute`)

// 	result, err := processor(markdown)
// 	if err != nil {
// 		t.Fatalf("Failed to process markdown: %v", err)
// 	}

// 	expected := `<p>This refers to <abbr title="American National Standards Institute">ANSI</abbr> standards.</p>`
// 	resultStr := strings.TrimSpace(string(result))

// 	if resultStr != expected {
// 		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, resultStr)
// 	}
// }
