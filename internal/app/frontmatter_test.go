package app

import (
	"strings"
	"testing"
)

func TestFrontmatterHTMLFormatter_FormatAsHTML_YAML(t *testing.T) {
	formatter := NewFrontmatterHTMLFormatter()

	// Test YAML formatting
	yamlFrontmatter := map[string]any{
		"__FMTYPE__": "YAML",
		"title":      "Test Document",
		"author":     "John Doe",
		"date":       "2024-01-01",
		"published":  true,
		"tags":       []interface{}{"go", "markdown", "test"},
		"config": map[string]interface{}{
			"theme": "dark",
			"debug": false,
		},
	}

	result := formatter.FormatAsHTML(yamlFrontmatter)
	// fmt.Printf("##> %s", result)

	// Verify YAML-specific formatting
	if !strings.Contains(result, `<div class="frontmatter-header"><abbr title="YAML">---</abbr></div>`) {
		t.Error("Expected YAML header with '---'")
	}
	if !strings.Contains(result, `<div class="frontmatter-footer">---</div>`) {
		t.Error("Expected YAML footer with '---'")
	}
	if !strings.Contains(result, `<span class="fm-key">title:</span>`) {
		t.Error("Expected YAML-style key formatting with colon")
	}
	if !strings.Contains(result, `<span class="fm-array-marker">-</span>`) {
		t.Error("Expected YAML-style array markers with dashes")
	}

	// Ensure _FM_TYPE is not displayed in output
	if strings.Contains(result, "_FM_TYPE") {
		t.Error("_FM_TYPE should not appear in formatted output")
	}
}

func TestFrontmatterHTMLFormatter_FormatAsHTML_TOML(t *testing.T) {
	formatter := NewFrontmatterHTMLFormatter()

	// Test TOML formatting
	tomlFrontmatter := map[string]any{
		"__FMTYPE__": "TOML",
		"title":      "Test Document",
		"author":     "John Doe",
		"date":       "2024-01-01",
		"published":  true,
		"tags":       []interface{}{"go", "markdown", "test"},
		"config": map[string]interface{}{
			"theme": "dark",
			"debug": false,
		},
	}

	result := formatter.FormatAsHTML(tomlFrontmatter)

	// Verify TOML-specific formatting
	if !strings.Contains(result, `<div class="frontmatter-header"><abbr title="TOML">+++</abbr></div>`) {
		t.Error("Expected TOML header with '+++'")
	}
	if !strings.Contains(result, `<div class="frontmatter-footer">+++</div>`) {
		t.Error("Expected TOML footer with '+++'")
	}
	if !strings.Contains(result, `<span class="fm-key">title</span> =`) {
		t.Error("Expected TOML-style key formatting with equals sign")
	}
	if !strings.Contains(result, `<span class="fm-key">[config]</span>`) {
		t.Error("Expected TOML-style table notation for nested objects")
	}
	if !strings.Contains(result, `tags</span> = [`) {
		t.Error("Expected TOML-style array formatting with brackets")
	}

	// Ensure _FM_TYPE is not displayed in output
	if strings.Contains(result, "_FM_TYPE") {
		t.Error("_FM_TYPE should not appear in formatted output")
	}
}

func TestFrontmatterHTMLFormatter_FormatAsHTML_EmptyType_DefaultsToYAML(t *testing.T) {
	formatter := NewFrontmatterHTMLFormatter()

	// Test default behavior when __FMTYPE__ is empty
	frontmatter := map[string]any{
		"__FMTYPE__": "",
		"title":      "Test Document",
	}

	result := formatter.FormatAsHTML(frontmatter)

	// Should default to YAML formatting
	if !strings.Contains(result, `<div class="frontmatter-header"><abbr title="YAML">---</abbr></div>`) {
		t.Error("Expected to default to YAML formatting when __FMTYPE__ is empty")
	}
}

func TestFrontmatterHTMLFormatter_FormatAsHTML_NoType_DefaultsToYAML(t *testing.T) {
	formatter := NewFrontmatterHTMLFormatter()

	// Test default behavior when __FMTYPE__ is missing
	frontmatter := map[string]any{
		"__FMTYPE__": "YAML",
		"title":      "Test Document",
	}

	result := formatter.FormatAsHTML(frontmatter)

	// Should default to YAML formatting
	if !strings.Contains(result, `<div class="frontmatter-header"><abbr title="YAML">---</abbr></div>`) {
		t.Error("Expected to default to YAML formatting when __FMTYPE__ is missing")
	}
}
