package app

import (
	"fmt"
	"html"
	"texto-vista/internal/utils"
	"sort"
	"strings"
)

// Format type constants
const (
	formatYAML = "YAML"
	formatTOML = "TOML"
)

// Delimiter constants
const (
	yamlDelimiter = "---"
	tomlDelimiter = "+++"
)

// FrontmatterHTMLFormatter converts frontmatter map to HTML with YAML-style syntax highlighting
type FrontmatterHTMLFormatter struct{}

// NewFrontmatterHTMLFormatter creates a new formatter instance
func NewFrontmatterHTMLFormatter() *FrontmatterHTMLFormatter {
	return &FrontmatterHTMLFormatter{}
}

// FormatAsHTML converts a map[string]any frontmatter to HTML with format-specific syntax highlighting
func (f *FrontmatterHTMLFormatter) FormatAsHTML(frontmatter map[string]any) string {
	emptyFrontmatter := `<div class="frontmatter-container"><div class="frontmatter-header">No frontmatter</div></div>`
	commentedFrontmatter := `<div class="frontmatter-container"><div class="frontmatter-header">No ACTIVE frontmatter (all commented out)</div></div>`
	// Note-to-Self:
	// Since we're now storing the key '__ABBR__' in the frontmatter map, we need to account for it:
	//      1. if UseAbbreviations is set, but no frontmatter is present, len(frontmatter) will be 1
	//      2. if UseAbbreviations is not set, but frontmatter is present, len(frontmatter) will at least 2 (1 for __FMTYPE__ and at least one actual frontmatter key)
	//      3. if UseAbbreviations is not set AND there is no frontmatter, len(frontmatter) should be 0
	//      edge-case: if #2 is true, but frontmatter is all commented out, len(frontmatter) should be 1 (i.e., __FMTYPE__ will still be set)
    //                 (but since all frontmatter is commented out, technically there is "No frontmatter", so essentially true :smile: )
	fmlen := len(frontmatter)
	if _, ok := frontmatter["__ABBR__"]; ok {
		if fmlen <= 1 {
			return emptyFrontmatter
		} else if fmlen == 2 {
			return commentedFrontmatter
		}
	} else {
		if fmlen <= 1 {
			return commentedFrontmatter
		}
	}

	fmType := f.extractAndNormalizeFMType(frontmatter)

	// Route to appropriate formatting method based on type
	switch fmType {
	case formatTOML:
		return f.formatWithContainer(frontmatter, fmType, tomlDelimiter, f.formatTOMLKeyValue)
	default: // YAML or any other type defaults to YAML formatting
		return f.formatWithContainer(frontmatter, fmType, yamlDelimiter, f.formatKeyValue)
	}
}

// extractAndNormalizeFMType extracts the frontmatter type and normalizes it
func (f *FrontmatterHTMLFormatter) extractAndNormalizeFMType(frontmatter map[string]any) string {
	var fmType string
	if fmtype, ok := utils.GetValue[string](frontmatter, "__FMTYPE__"); ok {
		fmType = strings.ToUpper(strings.TrimSpace(fmtype))
	}

	// Default to YAML if fmType is empty
	if fmType == "" {
		fmType = formatYAML
	}

	return fmType
}

// getSortedKeys returns sorted keys from frontmatter, excluding __FMTYPE__
func (f *FrontmatterHTMLFormatter) getSortedKeys(frontmatter map[string]any) []string {
	keys := make([]string, 0, len(frontmatter))
	for k := range frontmatter {
		if k == "__FMTYPE__" {  // Ignore frontmatter type key
			continue
		} else if k == "__ABBR__" {  // Ignore frontmatter abbreviations key
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// formatWithContainer provides the common container structure for both YAML and TOML
func (f *FrontmatterHTMLFormatter) formatWithContainer(frontmatter map[string]any, fmType, delimiter string, keyValueFormatter func(string, any) string) string {
	var htmlParts []string
	htmlParts = append(htmlParts, `<div class="frontmatter-container">`)
	htmlParts = append(htmlParts, fmt.Sprintf(`<div class="frontmatter-header"><abbr title="%s">%s</abbr></div>`, fmType, delimiter))
	// htmlParts = append(htmlParts, fmt.Sprintf(`<div class="frontmatter-header">%s %s</div>`, delimiter, html.EscapeString(fmType)))

	keys := f.getSortedKeys(frontmatter)

	// Process each key-value pair
	for _, key := range keys {
		value := frontmatter[key]
		htmlLine := keyValueFormatter(key, value)
		htmlParts = append(htmlParts, htmlLine)
	}

	htmlParts = append(htmlParts, fmt.Sprintf(`<div class="frontmatter-footer">%s</div>`, delimiter))
	htmlParts = append(htmlParts, `</div>`)

	return strings.Join(htmlParts, "\n")
}

// formatValue formats a value with the specified format style
func (f *FrontmatterHTMLFormatter) formatValue(value any, formatStyle string) string {
	if value == nil {
		return `<span class="fm-null">null</span>`
	}

	_ = formatStyle // Currently unused, but can be used for future format-specific handling

	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`<span class="fm-string">"%s"</span>`, html.EscapeString(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf(`<span class="fm-number">%v</span>`, v)
	case float32, float64:
		return fmt.Sprintf(`<span class="fm-number">%v</span>`, v)
	case bool:
		return fmt.Sprintf(`<span class="fm-boolean">%v</span>`, v)
	default:
		// Handle other types by converting to string
		valueStr := fmt.Sprintf("%v", v)
		// Check if it looks like a date/time
		if strings.Contains(valueStr, "-") && (strings.Contains(valueStr, ":") || strings.Contains(valueStr, "T")) {
			return fmt.Sprintf(`<span class="fm-datetime">%s</span>`, html.EscapeString(valueStr))
		}
		return fmt.Sprintf(`<span class="fm-string">"%s"</span>`, html.EscapeString(valueStr))
	}
}

// formatKeyValue formats a single key-value pair as HTML with YAML syntax
func (f *FrontmatterHTMLFormatter) formatKeyValue(key string, value any) string {
	return f.formatKeyValueWithSyntax(key, value, ":")
}

// formatTOMLKeyValue formats a single key-value pair as HTML with TOML syntax
func (f *FrontmatterHTMLFormatter) formatTOMLKeyValue(key string, value any) string {
	return f.formatKeyValueWithSyntax(key, value, " =")
}

// formatKeyValueWithSyntax formats a key-value pair with the specified syntax separator
func (f *FrontmatterHTMLFormatter) formatKeyValueWithSyntax(key string, value any, separator string) string {
	if value == nil {
		if separator == ":" {
			return fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s:</span> <span class="fm-null">null</span></div>`,
				html.EscapeString(key))
		}
		return fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s</span> = <span class="fm-null">null</span></div>`,
			html.EscapeString(key))
	}

	switch v := value.(type) {
	case []interface{}:
		if separator == " =" { // TOML array
			return f.formatTOMLArray(key, v)
		}
		return f.formatArray(key, v) // YAML array
	case map[string]interface{}:
		if separator == " =" { // TOML map
			return f.formatTOMLMap(key, v)
		}
		return f.formatMap(key, v) // YAML map
	default:
		valueHTML := f.formatValue(value, "")
		if separator == ":" {
			return fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s:</span> %s</div>`,
				html.EscapeString(key), valueHTML)
		}
		return fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s</span> = %s</div>`,
			html.EscapeString(key), valueHTML)
	}
}

// formatArray formats an array value
func (f *FrontmatterHTMLFormatter) formatArray(key string, arr []interface{}) string {
	var parts []string
	parts = append(parts, fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s:</span></div>`, html.EscapeString(key)))

	for _, item := range arr {
		itemHTML := f.formatArrayItem(item)
		parts = append(parts, fmt.Sprintf(`<div class="frontmatter-line frontmatter-indent">%s</div>`, itemHTML))
	}

	return strings.Join(parts, "\n")
}

// formatArrayItem formats a single array item with optional YAML-style marker
func (f *FrontmatterHTMLFormatter) formatArrayItem(value any) string {
	return f.formatArrayItemWithMarker(value, `<span class="fm-array-marker">-</span> `)
}

// formatTOMLArrayItem formats a single array item for TOML syntax
func (f *FrontmatterHTMLFormatter) formatTOMLArrayItem(value any) string {
	return f.formatArrayItemWithMarker(value, "")
}

// formatArrayItemWithMarker formats a single array item with optional marker prefix
func (f *FrontmatterHTMLFormatter) formatArrayItemWithMarker(value any, marker string) string {
	if value == nil {
		return marker + `<span class="fm-null">null</span>`
	}

	valueHTML := f.formatValue(value, "")
	return marker + valueHTML
}

// formatMap formats a nested map value with YAML syntax
func (f *FrontmatterHTMLFormatter) formatMap(key string, m map[string]interface{}) string {
	return f.formatMapWithSyntax(key, m, ":", f.formatKeyValue)
}

// formatTOMLArray formats an array value with TOML syntax
func (f *FrontmatterHTMLFormatter) formatTOMLArray(key string, arr []interface{}) string {
	// TOML arrays are formatted as [item1, item2, item3]
	var items []string
	for _, item := range arr {
		items = append(items, f.formatTOMLArrayItem(item))
	}

	arrayContent := strings.Join(items, ", ")
	return fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s</span> = [%s]</div>`,
		html.EscapeString(key), arrayContent)
}

// formatTOMLMap formats a nested map value with TOML table syntax
func (f *FrontmatterHTMLFormatter) formatTOMLMap(key string, m map[string]interface{}) string {
	return f.formatMapWithSyntax(key, m, "", f.formatTOMLKeyValue, fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">[%s]</span></div>`, html.EscapeString(key)))
}

// formatMapWithSyntax formats a nested map with the specified syntax and optional header
func (f *FrontmatterHTMLFormatter) formatMapWithSyntax(key string, m map[string]interface{}, separator string, keyValueFormatter func(string, any) string, header ...string) string {
	var parts []string

	if len(header) > 0 {
		// TOML table header
		parts = append(parts, header[0])
	} else {
		// YAML map header
		parts = append(parts, fmt.Sprintf(`<div class="frontmatter-line"><span class="fm-key">%s%s</span></div>`, html.EscapeString(key), separator))
	}

	keys := f.getSortedKeysFromMap(m)

	for _, nestedKey := range keys {
		nestedValue := m[nestedKey]
		nestedHTML := keyValueFormatter(nestedKey, nestedValue)
		// Add indentation to nested content
		indentedHTML := strings.ReplaceAll(nestedHTML, `class="frontmatter-line"`, `class="frontmatter-line frontmatter-indent"`)
		parts = append(parts, indentedHTML)
	}

	return strings.Join(parts, "\n")
}

// getSortedKeysFromMap returns sorted keys from a map[string]interface{}
func (f *FrontmatterHTMLFormatter) getSortedKeysFromMap(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
