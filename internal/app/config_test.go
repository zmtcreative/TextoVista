package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func newTestConfigManager(t *testing.T, args ...string) *ConfigManager {
	t.Helper()

	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })

	if len(args) == 0 {
		os.Args = []string{"texto-vista"}
	} else {
		os.Args = args
	}

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	return NewConfigManager()
}

func TestNewConfigManager(t *testing.T) {
	cm := newTestConfigManager(t)
	if cm == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	// Test that defaults are set correctly
	config := cm.GetConfig()
	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Test Application defaults
	if !config.Application.UseInlineHTML {
		t.Error("Default UseInlineHTML should be true")
	}
	if !config.Application.UseSanitize {
		t.Error("Default UseSanitize should be true")
	}
	if !config.Application.UseStripH1 {
		t.Error("Default StripH1 should be true")
	}
	if !config.Application.UseFrontmatterTitle {
		t.Error("Default UseFrontmatterTitle should be true")
	}

	// Test Markdown defaults
	if !config.Markdown.UseGFM {
		t.Error("Default UseGFM should be true")
	}
	if !config.Markdown.UseEmoji {
		t.Error("Default UseEmoji should be true")
	}
	if !config.Markdown.UseMermaid {
		t.Error("Default UseMermaid should be true")
	}

	// Test Alert Callouts defaults
	if !config.AlertCallouts.UseAlertCallouts {
		t.Error("Default UseAlertCallouts should be true")
	}
	if config.AlertCallouts.AlertCalloutStyle != "GFMPlus" {
		t.Errorf("Default AlertCalloutStyle = %q, want %q", config.AlertCallouts.AlertCalloutStyle, "GFMPlus")
	}
}

func TestConfigManagerGettersAndSetters(t *testing.T) {
	cm := newTestConfigManager(t)

	// Test initial values
	if !cm.UseInlineHTML() {
		t.Error("UseInlineHTML() should initially be true")
	}
	if !cm.UseSanitize() {
		t.Error("UseSanitize() should initially be true")
	}
	if !cm.UseGFM() {
		t.Error("UseGFM() should initially be true")
	}

	// Create new config and set it
	newConfig := &Config{
		Application: ApplicationOptions{
			UseInlineHTML:       false,
			UseSanitize:         false,
			UseStripH1:          false,
			UseFrontmatterTitle: false,
		},
		Markdown: MarkdownOptions{
			UseGFM:          false,
			UseEmoji:        false,
			UseMermaid:      false,
			UseFigure:       false,
			UseAnchor:       false,
			UseFences:       false,
			UseSections:     false,
			UseHighlighting: false,
			UseFancyLists:   false,
			UseAttributes:   false,
			UseTypographic:  false,
		},
		AlertCallouts: AlertCalloutOptions{
			UseAlertCallouts:  false,
			AlertCalloutStyle: "GFMStrict",
		},
	}

	cm.SetConfig(newConfig)

	// Test that values changed
	if cm.UseInlineHTML() {
		t.Error("UseInlineHTML() should be false after SetConfig")
	}
	if cm.UseSanitize() {
		t.Error("UseSanitize() should be false after SetConfig")
	}
	if cm.UseGFM() {
		t.Error("UseGFM() should be false after SetConfig")
	}
	if cm.AlertCalloutStyle() != "GFMStrict" {
		t.Errorf("AlertCalloutStyle() = %q, want %q", cm.AlertCalloutStyle(), "GFMStrict")
	}
}

func TestValidateAlertCalloutStyle(t *testing.T) {
	cm := newTestConfigManager(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid GFMStrict",
			input:    "GFMStrict",
			expected: "GFMStrict",
		},
		{
			name:     "valid GFMWithAliases",
			input:    "GFMWithAliases",
			expected: "GFMWithAliases",
		},
		{
			name:     "valid GFMPlus",
			input:    "GFMPlus",
			expected: "GFMPlus",
		},
		{
			name:     "valid Obsidian",
			input:    "Obsidian",
			expected: "Obsidian",
		},
		{
			name:     "invalid style",
			input:    "InvalidStyle",
			expected: "GFMPlus", // default
		},
		{
			name:     "empty string",
			input:    "",
			expected: "GFMPlus", // default
		},
		{
			name:     "case sensitive",
			input:    "gfmstrict", // lowercase
			expected: "GFMPlus",   // default because case matters
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cm.ValidateAlertCalloutStyle(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateAlertCalloutStyle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetApplicationConfig(t *testing.T) {
	cm := newTestConfigManager(t)

	useInlineHTML, useSanitize := cm.GetApplicationConfig()
	if !useInlineHTML {
		t.Error("GetApplicationConfig() useInlineHTML should be true")
	}
	if !useSanitize {
		t.Error("GetApplicationConfig() useSanitize should be true")
	}

	// Change config and test again
	newConfig := &Config{
		Application: ApplicationOptions{
			UseInlineHTML: false,
			UseSanitize:   false,
		},
	}
	cm.SetConfig(newConfig)

	useInlineHTML, useSanitize = cm.GetApplicationConfig()
	if useInlineHTML {
		t.Error("GetApplicationConfig() useInlineHTML should be false after change")
	}
	if useSanitize {
		t.Error("GetApplicationConfig() useSanitize should be false after change")
	}
}

func TestGetAlertCalloutConfig(t *testing.T) {
	cm := newTestConfigManager(t)

	style := cm.GetAlertCalloutConfig()
	if style != "GFMPlus" {
		t.Errorf("GetAlertCalloutConfig() = %q, want %q", style, "GFMPlus")
	}

	// Change config and test again
	newConfig := &Config{
		AlertCallouts: AlertCalloutOptions{
			AlertCalloutStyle: "GFMStrict",
		},
	}
	cm.SetConfig(newConfig)

	style = cm.GetAlertCalloutConfig()
	if style != "GFMStrict" {
		t.Errorf("GetAlertCalloutConfig() = %q, want %q", style, "GFMStrict")
	}
}

func TestApplyCliOverrides(t *testing.T) {
	cm := newTestConfigManager(t)

	// Initial values should be defaults (true)
	if !cm.UseInlineHTML() || !cm.UseSanitize() || !cm.UseStripH1() {
		t.Fatal("Initial configuration should have all true values")
	}

	// Apply CLI overrides
	allowInlineHTML := false
	sanitizeHTML := false
	stripH1 := false

	cm.ApplyCliOverrides(&allowInlineHTML, &sanitizeHTML, &stripH1)

	// Test that overrides were applied
	if cm.UseInlineHTML() {
		t.Error("UseInlineHTML() should be false after CLI override")
	}
	if cm.UseSanitize() {
		t.Error("UseSanitize() should be false after CLI override")
	}
	if cm.UseStripH1() {
		t.Error("UseStripH1() should be false after CLI override")
	}

	// Test with nil values (should not change)
	cm.ApplyCliOverrides(nil, nil, nil)

	// Values should remain false
	if cm.UseInlineHTML() || cm.UseSanitize() || cm.UseStripH1() {
		t.Error("Values should remain false when nil overrides are applied")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	cm := newTestConfigManager(t, "test-texto-vista")

	// Test that we can set and get configuration
	newConfig := &Config{
		Application: ApplicationOptions{
			UseInlineHTML:       false,
			UseSanitize:         false,
			UseStripH1:          false,
			UseFrontmatterTitle: false,
		},
		Markdown: MarkdownOptions{
			UseGFM:   false,
			UseEmoji: false,
		},
		AlertCallouts: AlertCalloutOptions{
			UseAlertCallouts:  false,
			AlertCalloutStyle: "GFMStrict",
		},
	}

	cm.SetConfig(newConfig)

	// Verify that the configuration was set correctly
	if cm.UseInlineHTML() {
		t.Error("Config should have UseInlineHTML = false")
	}
	if cm.UseSanitize() {
		t.Error("Config should have UseSanitize = false")
	}
	if cm.AlertCalloutStyle() != "GFMStrict" {
		t.Errorf("Config AlertCalloutStyle = %q, want %q", cm.AlertCalloutStyle(), "GFMStrict")
	}

	// Test that SaveConfig doesn't return an error (even if it creates files in user directories)
	err := cm.SaveConfig()
	if err != nil {
		t.Errorf("SaveConfig() error = %v", err)
	}

	if _, err := os.Stat(cm.configPath); err != nil {
		t.Fatalf("SaveConfig() did not create config file %q: %v", cm.configPath, err)
	}

	wantConfigPath := filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "test-texto-vista", "test-texto-vista.json")
	if cm.configPath != wantConfigPath {
		t.Fatalf("configPath = %q, want %q", cm.configPath, wantConfigPath)
	}
}

func TestGetAppNameFromExecutable(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple name",
			args:     []string{"texto-vista"},
			expected: "texto-vista",
		},
		{
			name:     "with extension",
			args:     []string{"texto-vista.exe"},
			expected: "texto-vista",
		},
		{
			name:     "full path",
			args:     []string{"/usr/local/bin/texto-vista"},
			expected: "texto-vista",
		},
		{
			name:     "windows path",
			args:     []string{"C:\\Program Files\\texto-vista\\texto-vista.exe"},
			expected: "texto-vista",
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: "texto-vista", // fallback
		},
		{
			name:     "different name",
			args:     []string{"markdown-viewer.exe"},
			expected: "markdown-viewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			result := getAppNameFromExecutable()
			if result != tt.expected {
				t.Errorf("getAppNameFromExecutable() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestAlertCalloutStylesConstant(t *testing.T) {
	// Test that the AlertCalloutStyles map is properly defined
	expectedStyles := map[string]string{
		"GFMStrict":      "Strict GFM Alerts (Standard 5 Alert Types)",
		"GFMWithAliases": "GFM Alerts + Aliases (GFM + Aliases [e.g., notes = note])",
		"GFMPlus":        "GFM Alerts Plus (GFM + Some Obsidian-Style Callouts)",
		"Obsidian":       "Obsidian-Style (Obsidian Icons and Callout Names)",
	}

	for key, expectedDesc := range expectedStyles {
		if desc, exists := AlertCalloutStyles[key]; !exists {
			t.Errorf("AlertCalloutStyles missing key: %s", key)
		} else if desc != expectedDesc {
			t.Errorf("AlertCalloutStyles[%s] = %q, want %q", key, desc, expectedDesc)
		}
	}

	if len(AlertCalloutStyles) != len(expectedStyles) {
		t.Errorf("AlertCalloutStyles length = %d, want %d", len(AlertCalloutStyles), len(expectedStyles))
	}
}

func TestConfigManagerMarkdownFeatureGetters(t *testing.T) {
	cm := newTestConfigManager(t)

	config := &Config{
		Application: ApplicationOptions{
			UseInlineHTML:            true,
			UseSanitize:              false,
			UseStripH1:               true,
			UseFrontmatterTitle:      false,
			FontFamily:               "Atkinson Hyperlegible",
			FontSize:                 18,
			FontFamilyMono:           "Cascadia Code",
			FontSizeMono:             15,
			UseAdvancedFontDetection: false,
		},
		Markdown: MarkdownOptions{
			UseGFM:           false,
			UsePHPMDExt:      true,
			UseEmoji:         false,
			UseMermaid:       true,
			UseFigure:        false,
			UseAnchor:        true,
			UseFences:        false,
			UseSections:      true,
			UseHighlighting:  false,
			UseFancyLists:    true,
			UseAttributes:    false,
			UseTypographic:   true,
			UseAbbreviations: true,
			UseKatex:         false,
			UseD2Diagrams:    true,
		},
		AlertCallouts: AlertCalloutOptions{
			UseAlertCallouts:  false,
			AlertCalloutStyle: "Obsidian",
		},
	}

	cm.SetConfig(config)

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{name: "UseInlineHTML", got: cm.UseInlineHTML(), want: true},
		{name: "UseSanitize", got: cm.UseSanitize(), want: true},
		{name: "UseStripH1", got: cm.UseStripH1(), want: true},
		{name: "UseFrontmatterTitle", got: cm.UseFrontmatterTitle(), want: false},
		{name: "UseGFM", got: cm.UseGFM(), want: false},
		{name: "UsePHPMDExt", got: cm.UsePHPMDExt(), want: true},
		{name: "UseEmoji", got: cm.UseEmoji(), want: false},
		{name: "UseMermaid", got: cm.UseMermaid(), want: true},
		{name: "UseFigure", got: cm.UseFigure(), want: false},
		{name: "UseAnchor", got: cm.UseAnchor(), want: true},
		{name: "UseFences", got: cm.UseFences(), want: false},
		{name: "UseSections", got: cm.UseSections(), want: true},
		{name: "UseHighlighting", got: cm.UseHighlighting(), want: false},
		{name: "UseFancyLists", got: cm.UseFancyLists(), want: true},
		{name: "UseAttributes", got: cm.UseAttributes(), want: false},
		{name: "UseTypographic", got: cm.UseTypographic(), want: true},
		{name: "UseAbbreviations", got: cm.UseAbbreviations(), want: true},
		{name: "UseKatex", got: cm.UseKatex(), want: false},
		{name: "UseD2Diagrams", got: cm.UseD2Diagrams(), want: true},
		{name: "UseAlertCallouts", got: cm.UseAlertCallouts(), want: false},
		{name: "GetUseAdvancedFontDetection", got: cm.GetUseAdvancedFontDetection(), want: false},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s() = %v, want %v", tt.name, tt.got, tt.want)
		}
	}

	if got := cm.AlertCalloutStyle(); got != "Obsidian" {
		t.Fatalf("AlertCalloutStyle() = %q, want %q", got, "Obsidian")
	}
	if got := cm.GetFontFamily(); got != "Atkinson Hyperlegible" {
		t.Fatalf("GetFontFamily() = %q, want %q", got, "Atkinson Hyperlegible")
	}
	if got := cm.GetFontSize(); got != 18.0 {
		t.Fatalf("GetFontSize() = %v, want %v", got, 18.0)
	}
	if got := cm.GetFontFamilyMono(); got != "Cascadia Code" {
		t.Fatalf("GetFontFamilyMono() = %q, want %q", got, "Cascadia Code")
	}
	if got := cm.GetFontSizeMono(); got != 15.0 {
		t.Fatalf("GetFontSizeMono() = %v, want %v", got, 15.0)
	}

	cm.SetFontFamily("Literata")
	cm.SetFontSize(20)
	cm.SetFontFamilyMono("Fira Code")
	cm.SetFontSizeMono(13)
	cm.SetUseAdvancedFontDetection(true)

	if got := cm.GetFontFamily(); got != "Literata" {
		t.Fatalf("GetFontFamily() after SetFontFamily() = %q, want %q", got, "Literata")
	}
	if got := cm.GetFontSize(); got != 20.0 {
		t.Fatalf("GetFontSize() after SetFontSize() = %v, want %v", got, 20.0)
	}
	if got := cm.GetFontFamilyMono(); got != "Fira Code" {
		t.Fatalf("GetFontFamilyMono() after SetFontFamilyMono() = %q, want %q", got, "Fira Code")
	}
	if got := cm.GetFontSizeMono(); got != 13.0 {
		t.Fatalf("GetFontSizeMono() after SetFontSizeMono() = %v, want %v", got, 13.0)
	}
	if !cm.GetUseAdvancedFontDetection() {
		t.Fatal("GetUseAdvancedFontDetection() = false after setter, want true")
	}
}

func TestLoadConfigFallsBackOnInvalidJSON(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "broken.json")
	if err := os.WriteFile(configPath, []byte(`{"application":{"font_size":"bad"}`), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")

	cm := &ConfigManager{
		config:     &Config{},
		configPath: configPath,
		viper:      v,
	}

	cm.loadConfig()

	if !cm.config.Application.UseInlineHTML {
		t.Fatal("UseInlineHTML default fallback = false, want true")
	}
	if !cm.config.Markdown.UsePHPMDExt {
		t.Fatal("UsePHPMDExt default fallback = false, want true")
	}
	if cm.config.Markdown.UseAbbreviations {
		t.Fatal("UseAbbreviations default fallback = true, want false")
	}
	if cm.config.AlertCallouts.AlertCalloutStyle != "GFMPlus" {
		t.Fatalf("AlertCalloutStyle fallback = %q, want %q", cm.config.AlertCallouts.AlertCalloutStyle, "GFMPlus")
	}
	if cm.config.Application.FontFamily == "" || cm.config.Application.FontFamilyMono == "" {
		t.Fatal("fallback font families should not be empty")
	}
}

func TestLoadConfigUsesDefaultsWhenFileMissing(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "missing.json")
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("json")
	v.SetDefault("application.use_inline_html", true)
	v.SetDefault("application.use_sanitize_html", true)
	v.SetDefault("application.use_strip_h1", true)
	v.SetDefault("application.use_frontmatter_title", true)
	v.SetDefault("application.use_auto_refresh", true)
	v.SetDefault("application.font_family", "Verdana")
	v.SetDefault("application.font_size", 16.0)
	v.SetDefault("application.font_family_mono", "Consolas")
	v.SetDefault("application.font_size_mono", 14.0)
	v.SetDefault("application.use_advanced_font_detection", true)
	v.SetDefault("markdown.use_gfm", true)
	v.SetDefault("markdown.use_php_md_ext", true)
	v.SetDefault("markdown.use_emoji", true)
	v.SetDefault("markdown.use_mermaid", true)
	v.SetDefault("markdown.use_figure", true)
	v.SetDefault("markdown.use_anchor", true)
	v.SetDefault("markdown.use_fences", true)
	v.SetDefault("markdown.use_sections", true)
	v.SetDefault("markdown.use_highlighting", true)
	v.SetDefault("markdown.use_fancylists", true)
	v.SetDefault("markdown.use_attributes", true)
	v.SetDefault("markdown.use_typographic", true)
	v.SetDefault("markdown.use_abbreviations", false)
	v.SetDefault("markdown.use_katex", true)
	v.SetDefault("markdown.use_d2_diagrams", true)
	v.SetDefault("alert_callouts.use_alertcallouts", true)
	v.SetDefault("alert_callouts.alertcallout_style", "GFMPlus")

	cm := &ConfigManager{
		config:     &Config{},
		configPath: configPath,
		viper:      v,
	}

	cm.loadConfig()

	if !cm.UseInlineHTML() || !cm.UseSanitize() || !cm.UsePHPMDExt() {
		t.Fatal("missing config file did not preserve default values")
	}
	if !cm.UseAutoRefresh() {
		t.Fatal("missing config file did not preserve auto refresh default value")
	}
	if cm.GetFontFamily() == "" || cm.GetFontFamilyMono() == "" {
		t.Fatal("missing config file did not preserve default font values")
	}
	if cm.AlertCalloutStyle() != "GFMPlus" {
		t.Fatalf("AlertCalloutStyle() = %q, want %q", cm.AlertCalloutStyle(), "GFMPlus")
	}
}

func TestGetConfigDirUsesXDGConfigHome(t *testing.T) {
	xdgHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgHome)

	configDir := getConfigDir("texto-vista")
	want := filepath.Join(xdgHome, "texto-vista")

	if configDir != want {
		t.Fatalf("getConfigDir() = %q, want %q", configDir, want)
	}
	if info, err := os.Stat(configDir); err != nil || !info.IsDir() {
		t.Fatalf("getConfigDir() did not create directory %q: %v", configDir, err)
	}
}

func TestGetConfigDirFallsBackToCurrentDirectoryWhenHomeUnavailable(t *testing.T) {
	originalHomeDir := osUserHomeDir
	osUserHomeDir = func() (string, error) {
		return "", os.ErrNotExist
	}
	defer func() { osUserHomeDir = originalHomeDir }()
	t.Setenv("XDG_CONFIG_HOME", "")

	configDir := getConfigDir("texto-vista")
	if configDir != "." {
		t.Fatalf("getConfigDir() = %q, want %q", configDir, ".")
	}
	if info, err := os.Stat(configDir); err != nil || !info.IsDir() {
		t.Fatalf("getConfigDir() fallback directory check failed: %v", err)
	}
}
