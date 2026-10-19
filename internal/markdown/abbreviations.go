package markdown

import (
	"bufio"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// abbreviationDefinitionPattern matches abbreviation definitions
var abbreviationDefinitionPattern = regexp.MustCompile(`^(?:[\t ]*|\s*[\*\+-]\s+)\*\[([a-zA-Z0-9][^\[\]]+)]:[\t ]+(.+)$`) // Valid ABBR Pattern
var abbrInvalidPattern1 = regexp.MustCompile(`^(?:[\t ]*|\s*[\*\+-]\s+)(\*\[(?:[^a-zA-Z0-9][^\[\]]+)]:[\t ]*(?:.+))$`)    // Invalid ABBR Pattern (non-alphanumeric starting character)
var abbrInvalidPattern2 = regexp.MustCompile(`^(?:[\t ]*|\s*[\*\+-]\s+)(\*\[(?:[^\[\]])]:[\t ]*(?:.+))$`)                 // Invalid ABBR Pattern (cannot have only a single character)
var abbrInvalidPattern3 = regexp.MustCompile(`^(?:[\t ]*|\s*[\*\+-]\s+)(\*\[(?:[^\[\]]+)]:[^\t ]+(?:.+))$`)               // Invalid ABBR Pattern (no whitespace between `:` and the value)
var abbrInvalidPattern4 = regexp.MustCompile(`^(?:[\t ]*|\s*[\*\+-]\s+)\*\[(.*(?:\[*[^\]\[]*\])+.*)]:[\t ]+(.+)$`)        // Invalid ABBR Pattern (no nested brackets)

func GetMarkdownAbbreviations(markdown []byte) (map[string]string, []byte) {
	// Extract abbreviation definitions from the markdown
	definitions, cleanMarkdown := extractAbbreviations(string(markdown))

	return definitions, []byte(cleanMarkdown)
}

// extractAbbreviations extracts abbreviation definitions from markdown text
func extractAbbreviations(markdown string) (map[string]string, string) {
	definitions := make(map[string]string)
	var cleanedLines []string

	// bufio defaults to a 64K buffer, which will be too small for some documents
	bufSize := len(markdown) + 1024
	scanner := bufio.NewScanner(strings.NewReader(markdown))
	buffer := make([]byte, 0, bufSize)
	scanner.Buffer(buffer, bufSize)
	// Now that the buffer is adjusted, start the scan
	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line is an abbreviation definition
		if matches := abbreviationDefinitionPattern.FindStringSubmatch(line); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			if key != "" && value != "" {
				definitions[key] = value
				// Skip this line (don't add to cleaned markdown)
				continue
			}
		}
		// Check for invalid abbreviation patterns and change them to bullet list items wrapped in a bracketed span
		if matches := abbrInvalidPattern1.FindStringSubmatch(line); matches != nil {
			commentLine := fmt.Sprintf("- `%s` &nbsp;[(**INVALID ABBR DEF:** Must start with an alpha-numeric character)]{.invalid}", strings.TrimSpace(matches[1]))
			cleanedLines = append(cleanedLines, commentLine)
			continue
		}
		if matches := abbrInvalidPattern2.FindStringSubmatch(line); matches != nil {
			commentLine := fmt.Sprintf("- `%s` &nbsp;[(**INVALID ABBR DEF:** Cannot be a single character)]{.invalid}", strings.TrimSpace(matches[1]))
			cleanedLines = append(cleanedLines, commentLine)
			continue
		}
		if matches := abbrInvalidPattern3.FindStringSubmatch(line); matches != nil {
			commentLine := fmt.Sprintf("- `%s` &nbsp;[(**INVALID ABBR DEF:** Must have whitespace between `:` and the value)]{.invalid}", strings.TrimSpace(matches[1]))
			cleanedLines = append(cleanedLines, commentLine)
			continue
		}
		if matches := abbrInvalidPattern4.FindStringSubmatch(line); matches != nil {
			commentLine := fmt.Sprintf("- `[%s]` &nbsp;[(**INVALID ABBR DEF:** Cannot have nested brackets)]{.invalid}", strings.TrimSpace(matches[1]))
			cleanedLines = append(cleanedLines, commentLine)
			continue
		}

		// Add non-abbreviation lines to cleaned markdown
		cleanedLines = append(cleanedLines, line)
	}

	return definitions, strings.Join(cleanedLines, "\n")
}

// ReplaceAbbreviationsInHTML replaces abbreviations in HTML output
func ReplaceAbbreviationsInHTML(htmlContent []byte, definitions map[string]string) []byte {
	if len(definitions) == 0 {
		return htmlContent
	}

	// Sort keys by length (longest first) to handle overlapping abbreviations correctly
	keys := make([]string, 0, len(definitions))
	for key := range definitions {
		keys = append(keys, key)
	}

	// Simple bubble sort by length (descending)
	for i := 0; i < len(keys)-1; i++ {
		for j := 0; j < len(keys)-i-1; j++ {
			if len(keys[j]) < len(keys[j+1]) {
				keys[j], keys[j+1] = keys[j+1], keys[j]
			}
		}
	}

	result := string(htmlContent)

	for _, key := range keys {
		value := definitions[key]
		result = replaceAbbreviationInHTML(result, key, value)
	}

	return []byte(result)
}

// replaceAbbreviationInHTML replaces a single abbreviation in HTML content
func replaceAbbreviationInHTML(htmlContent, key, value string) string {
	// Pattern to match the key as a whole word, but not inside HTML tags or code blocks
	pattern := `\b` + regexp.QuoteMeta(key) + `\b`
	re := regexp.MustCompile(pattern)

	// Find all matches
	matches := re.FindAllStringIndex(htmlContent, -1)

	// Process matches from right to left to avoid index shifting
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		start, end := match[0], match[1]

		// Check if this match is inside a code block or HTML tag
		if isInsideCodeOrTag(htmlContent, start, end) {
			continue
		} else if isInsideAbbreviation(htmlContent, start, end) {
			continue
		}

		// Replace with abbr tag
		replacement := `<abbr title="` + html.EscapeString(value) + `">` + html.EscapeString(key) + `</abbr>`
		htmlContent = htmlContent[:start] + replacement + htmlContent[end:]
	}

	return htmlContent
}

// isInsideAbbreviation checks if a text position is inside an abbreviation element
func isInsideAbbreviation(htmlContent string, start, end int) bool {
	// Check for <abbr> tags before the match
	beforeText := htmlContent[:start]
	_ = end
	// afterText := htmlContent[end:]

	abbrCount := strings.Count(beforeText, "<abbr")
	abbrCloseCount := strings.Count(beforeText, "</abbr>")

	// If we're inside an <abbr> element, skip this match
	if abbrCount > abbrCloseCount {
		return true
	}

	return false
}

// isInsideCodeOrTag checks if a text position is inside a code block or HTML tag
func isInsideCodeOrTag(htmlContent string, start, end int) bool {
	// Check for code blocks
	beforeText := htmlContent[:start]
	afterText := htmlContent[end:]

	// Count <pre> and </pre> tags before the match
	preCount := strings.Count(beforeText, "<pre")
	preCloseCount := strings.Count(beforeText, "</pre>")

	// If we're inside a <pre> block, skip this match
	if preCount > preCloseCount {
		return true
	}

	// Count <code> and </code> tags before the match
	codeCount := strings.Count(beforeText, "<code")
	codeCloseCount := strings.Count(beforeText, "</code>")

	// If we're inside a <code> block, skip this match
	if codeCount > codeCloseCount {
		return true
	}

	// Check if we're inside an HTML tag (between < and >)
	lastOpenTag := strings.LastIndex(beforeText, "<")
	lastCloseTag := strings.LastIndex(beforeText, ">")

	if lastOpenTag > lastCloseTag {
		// We might be inside an HTML tag
		nextCloseTag := strings.Index(afterText, ">")
		if nextCloseTag != -1 {
			return true
		}
	}

	return false
}
