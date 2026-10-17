package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"texto-vista/internal/utils"

	"github.com/spf13/viper"
)

var osUserHomeDir = os.UserHomeDir

func defaultConfig() *Config {
	return &Config{
		Application: ApplicationOptions{
			UseInlineHTML:            true,
			UseSanitize:              true,
			UseStripH1:               true,
			UseFrontmatterTitle:      true,
			UseAutoRefresh:           true,
			FontFamily:               "Verdana, Arial, Helvetica, Tahoma, Geneva, sans-serif",
			FontSize:                 16.0,
			FontFamilyMono:           "Consolas, Monaco, DejaVu Sans Mono, Liberation Mono, Courier New, Courier, monospace",
			FontSizeMono:             14.0,
			UseAdvancedFontDetection: true,
		},
		Markdown: MarkdownOptions{
			UseGFM:           true,
			UsePHPMDExt:      true,
			UseEmoji:         true,
			UseMermaid:       true,
			UseFigure:        true,
			UseAnchor:        true,
			UseFences:        true,
			UseSections:      true,
			UseHighlighting:  true,
			UseFancyLists:    true,
			UseAttributes:    true,
			UseTypographic:   true,
			UseAbbreviations: false,
			UseKatex:         true,
			UseD2Diagrams:    true,
		},
		AlertCallouts: AlertCalloutOptions{
			UseAlertCallouts:  true,
			AlertCalloutStyle: "GFMPlus",
		},
	}
}

// Config represents the tabs in the settings dialog
type Config struct {
	Application   ApplicationOptions  `mapstructure:"application" json:"application"`       // Application
	Markdown      MarkdownOptions     `mapstructure:"markdown" json:"markdown"`             // Markdown Features
	AlertCallouts AlertCalloutOptions `mapstructure:"alert_callouts" json:"alert_callouts"` // Alert Callouts
}

// Application-Specific Settings
type ApplicationOptions struct {
	UseInlineHTML            bool    `mapstructure:"use_inline_html" json:"use_inline_html"`                         // Inline HTML support (allow inline HTML in Markdown)
	UseSanitize              bool    `mapstructure:"use_sanitize_html" json:"use_sanitize_html"`                     // HTML Sanitization (remove unsafe elements and links)
	UseStripH1               bool    `mapstructure:"use_strip_h1" json:"use_strip_h1"`                               // Strip First H1
	UseFrontmatterTitle      bool    `mapstructure:"use_frontmatter_title" json:"use_frontmatter_title"`             // Always Use Frontmatter Title if Available
	UseAutoRefresh           bool    `mapstructure:"use_auto_refresh" json:"use_auto_refresh"`                       // Automatically reload the current file when it changes on disk
	FontFamily               string  `mapstructure:"font_family" json:"font_family"`                                 // Selected font family
	FontSize                 float64 `mapstructure:"font_size" json:"font_size"`                                     // Selected font size in pixels
	FontFamilyMono           string  `mapstructure:"font_family_mono" json:"font_family_mono"`                       // Selected monospace font family
	FontSizeMono             float64 `mapstructure:"font_size_mono" json:"font_size_mono"`                           // Selected monospace font size in pixels
	UseAdvancedFontDetection bool    `mapstructure:"use_advanced_font_detection" json:"use_advanced_font_detection"` // Use advanced monospace font detection
}

// Markdown-Specific Settings
type MarkdownOptions struct {
	UseGFM           bool `mapstructure:"use_gfm" json:"use_gfm"`                     // GitHub Flavored Markdown
	UsePHPMDExt      bool `mapstructure:"use_php_md_ext" json:"use_php_md_ext"`       // PHP Markdown Extensions Support
	UseEmoji         bool `mapstructure:"use_emoji" json:"use_emoji"`                 // Emoji Support
	UseMermaid       bool `mapstructure:"use_mermaid" json:"use_mermaid"`             // Mermaid Diagrams Support
	UseFigure        bool `mapstructure:"use_figure" json:"use_figure"`               // Image Figure Wrapping Support
	UseAnchor        bool `mapstructure:"use_anchor" json:"use_anchor"`               // Anchor Links on Headings
	UseFences        bool `mapstructure:"use_fences" json:"use_fences"`               // Fenced DIVs
	UseSections      bool `mapstructure:"use_sections" json:"use_sections"`           // Wrap Headings in SECTION Elements
	UseHighlighting  bool `mapstructure:"use_highlighting" json:"use_highlighting"`   // Fenced Code Highlighting
	UseFancyLists    bool `mapstructure:"use_fancylists" json:"use_fancylists"`       // Allow Pandoc-Style Fancy Lists
	UseAttributes    bool `mapstructure:"use_attributes" json:"use_attributes"`       // Allow Custom Attributes (using '{.myclass}' syntax)
	UseAbbreviations bool `mapstructure:"use_abbreviations" json:"use_abbreviations"` // Allow Abbreviations (using `*[ABBR]: Abbreviations` syntax)
	UseTypographic   bool `mapstructure:"use_typographic" json:"use_typographic"`     // Typographic Extensions to Use Fancy Quotes
	UseKatex         bool `mapstructure:"use_katex" json:"use_katex"`                 // KaTeX Support for Mathematical Expressions
	UseD2Diagrams    bool `mapstructure:"use_d2_diagrams" json:"use_d2_diagrams"`     // D2 Diagrams Support
}

// Alert Callouts Settings
type AlertCalloutOptions struct {
	UseAlertCallouts  bool   `mapstructure:"use_alertcallouts" json:"use_alertcallouts"`   // GitHub and/or Obsidian Alert/Callouts
	AlertCalloutStyle string `mapstructure:"alertcallout_style" json:"alertcallout_style"` // Select Alert Callout Style
}

// ConfigManager handles configuration loading and saving
type ConfigManager struct {
	config     *Config
	configPath string
	viper      *viper.Viper
}

// AlertCalloutStyles defines the available alert callout styles
// (Note to self: Do NOT change "GFMStrict" if you can help it -- this is being used as the fallback default)
var AlertCalloutStyles = map[string]string{
	"GFMStrict":      "Strict GFM Alerts (Standard 5 Alert Types)",
	"GFMWithAliases": "GFM Alerts + Aliases (GFM + Aliases [e.g., notes = note])",
	"GFMPlus":        "GFM Alerts Plus (GFM + Some Obsidian-Style Callouts)",
	"Obsidian":       "Obsidian-Style (Obsidian Icons and Callout Names)",
}

// getAppNameFromExecutable extracts the application name from the executable path
// without the file extension. This is a thin wrapper around the shared utility.
func getAppNameFromExecutable() string {
	_, appName := utils.GetExecutableBaseName()
	return appName
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	v := viper.New()

	// Get appName
	appName := getAppNameFromExecutable()

	// Set configuration file name and type
	v.SetConfigName(appName)
	v.SetConfigType("json")

	// Set default values for Application section
	v.SetDefault("application.use_inline_html", true)
	v.SetDefault("application.use_sanitize_html", true)
	v.SetDefault("application.use_strip_h1", true)
	v.SetDefault("application.use_frontmatter_title", true)
	v.SetDefault("application.use_auto_refresh", true)
	v.SetDefault("application.font_family", "Verdana, Arial, Helvetica, Tahoma, Geneva, sans-serif")
	v.SetDefault("application.font_size", 16.0)
	v.SetDefault("application.font_family_mono", "Consolas, Monaco, DejaVu Sans Mono, Liberation Mono, Courier New, Courier, monospace")
	v.SetDefault("application.font_size_mono", 14.0)
	v.SetDefault("application.use_advanced_font_detection", true)

	// Set default values for Markdown section
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

	// Set default values for Alert Callouts section
	v.SetDefault("alert_callouts.use_alertcallouts", true)
	v.SetDefault("alert_callouts.alertcallout_style", "GFMPlus")

	// Get configuration directory
	configDir := getConfigDir(appName)
	v.AddConfigPath(configDir)

	configPath := filepath.Join(configDir, appName+".json")

	cm := &ConfigManager{
		config:     &Config{},
		configPath: configPath,
		viper:      v,
	}

	// Load existing configuration
	cm.loadConfig()

	return cm
}

// getConfigDir returns the appropriate configuration directory for the OS
func getConfigDir(appName string) string {
	var configDir string

	if configPath := os.Getenv("XDG_CONFIG_HOME"); configPath != "" {
		// Linux XDG standard
		configDir = filepath.Join(configPath, appName)
	} else if homeDir, err := osUserHomeDir(); err == nil {
		// Windows, macOS, and Linux fallback
		switch {
		case os.Getenv("OS") == "Windows_NT" || filepath.Separator == '\\':
			configDir = filepath.Join(homeDir, "AppData", "Roaming", appName)
		case os.Getenv("HOME") != "":
			configDir = filepath.Join(homeDir, ".config", appName)
		default:
			configDir = filepath.Join(homeDir, ".config", appName)
		}
	} else {
		// Fallback to current directory
		configDir = "."
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("##> LOG: Warning: could not create config directory %q: %v", configDir, err)
	}

	return configDir
}

// loadConfig loads the configuration from file
func (cm *ConfigManager) loadConfig() {
	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			log.Printf("##> LOG: Config file not found, using defaults: %s", cm.configPath)
		} else {
			// Config file found but another error occurred
			log.Printf("##> LOG: Error reading config file: %v", err)
			cm.config = defaultConfig()
			return
		}
	}

	// Unmarshal configuration into struct
	if err := cm.viper.Unmarshal(cm.config); err != nil {
		log.Printf("##> LOG: Error unmarshaling config: %v", err)
		// Use defaults if unmarshal fails
		cm.config = defaultConfig()
	}

	// Validate and enforce security constraint: inline HTML must always be sanitized
	if cm.config.Application.UseInlineHTML && !cm.config.Application.UseSanitize {
		log.Printf("##> LOG: Security: Enabling HTML sanitization (inline HTML requires sanitization)")
		cm.config.Application.UseSanitize = true
	}
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetConfig updates the configuration with validation
func (cm *ConfigManager) SetConfig(newConfig *Config) {
	// Validate and enforce security constraint: inline HTML must always be sanitized
	if newConfig.Application.UseInlineHTML && !newConfig.Application.UseSanitize {
		// Force sanitization to true when inline HTML is enabled to prevent XSS
		newConfig.Application.UseSanitize = true
	}
	cm.config = newConfig
}

// SaveConfig saves the current configuration to file
func (cm *ConfigManager) SaveConfig() error {
	// Update viper with current config values for Application section
	cm.viper.Set("application.use_inline_html", cm.config.Application.UseInlineHTML)
	cm.viper.Set("application.use_sanitize_html", cm.config.Application.UseSanitize)
	cm.viper.Set("application.use_strip_h1", cm.config.Application.UseStripH1)
	cm.viper.Set("application.use_frontmatter_title", cm.config.Application.UseFrontmatterTitle)
	cm.viper.Set("application.use_auto_refresh", cm.config.Application.UseAutoRefresh)
	cm.viper.Set("application.font_family", cm.config.Application.FontFamily)
	cm.viper.Set("application.font_size", cm.config.Application.FontSize)
	cm.viper.Set("application.font_family_mono", cm.config.Application.FontFamilyMono)
	cm.viper.Set("application.font_size_mono", cm.config.Application.FontSizeMono)
	cm.viper.Set("application.use_advanced_font_detection", cm.config.Application.UseAdvancedFontDetection)

	// Update viper with current config values for Markdown section
	cm.viper.Set("markdown.use_gfm", cm.config.Markdown.UseGFM)
	cm.viper.Set("markdown.use_php_md_ext", cm.config.Markdown.UsePHPMDExt)
	cm.viper.Set("markdown.use_emoji", cm.config.Markdown.UseEmoji)
	cm.viper.Set("markdown.use_mermaid", cm.config.Markdown.UseMermaid)
	cm.viper.Set("markdown.use_figure", cm.config.Markdown.UseFigure)
	cm.viper.Set("markdown.use_anchor", cm.config.Markdown.UseAnchor)
	cm.viper.Set("markdown.use_fences", cm.config.Markdown.UseFences)
	cm.viper.Set("markdown.use_sections", cm.config.Markdown.UseSections)
	cm.viper.Set("markdown.use_highlighting", cm.config.Markdown.UseHighlighting)
	cm.viper.Set("markdown.use_fancylists", cm.config.Markdown.UseFancyLists)
	cm.viper.Set("markdown.use_attributes", cm.config.Markdown.UseAttributes)
	cm.viper.Set("markdown.use_typographic", cm.config.Markdown.UseTypographic)
	cm.viper.Set("markdown.use_abbreviations", cm.config.Markdown.UseAbbreviations)
	cm.viper.Set("markdown.use_katex", cm.config.Markdown.UseKatex)
	cm.viper.Set("markdown.use_d2_diagrams", cm.config.Markdown.UseD2Diagrams)

	// Update viper with current config values for Alert Callouts section
	cm.viper.Set("alert_callouts.use_alertcallouts", cm.config.AlertCallouts.UseAlertCallouts)
	cm.viper.Set("alert_callouts.alertcallout_style", cm.config.AlertCallouts.AlertCalloutStyle)

	// Write configuration to file
	if err := cm.viper.WriteConfigAs(cm.configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ApplyCliOverrides applies command-line argument overrides to the configuration
// DEPRECATED: We no longer provide command line overrides (keeping this here for reference for now)
func (cm *ConfigManager) ApplyCliOverrides(allowInlineHTML, sanitizeHTML, stripH1 *bool) {
	if allowInlineHTML != nil {
		cm.config.Application.UseInlineHTML = *allowInlineHTML
	}
	if sanitizeHTML != nil {
		cm.config.Application.UseSanitize = *sanitizeHTML
	}
	if stripH1 != nil {
		cm.config.Application.UseStripH1 = *stripH1
	}

	// Validate and enforce security constraint: inline HTML must always be sanitized
	if cm.config.Application.UseInlineHTML && !cm.config.Application.UseSanitize {
		cm.config.Application.UseSanitize = true
	}
}

// ValidateAlertCalloutStyle validates and returns a valid alert callout style
func (cm *ConfigManager) ValidateAlertCalloutStyle(style string) string {
	if _, exists := AlertCalloutStyles[style]; exists {
		return style
	}
	// Return default if invalid
	return "GFMPlus"
}

// GetApplicationConfig returns application configuration for markdown processing
func (cm *ConfigManager) GetApplicationConfig() (useInlineHTML, useSanitize bool) {
	return cm.config.Application.UseInlineHTML, cm.config.Application.UseSanitize
}

// GetAlertCalloutConfig returns the alert callout style configuration
func (cm *ConfigManager) GetAlertCalloutConfig() string {
	return cm.config.AlertCallouts.AlertCalloutStyle
}

// Application-specific configuration getters
func (cm *ConfigManager) UseInlineHTML() bool {
	return cm.config.Application.UseInlineHTML
}

func (cm *ConfigManager) UseSanitize() bool {
	return cm.config.Application.UseSanitize
}

func (cm *ConfigManager) UseStripH1() bool {
	return cm.config.Application.UseStripH1
}

func (cm *ConfigManager) UseFrontmatterTitle() bool {
	return cm.config.Application.UseFrontmatterTitle
}

func (cm *ConfigManager) UseAutoRefresh() bool {
	return cm.config.Application.UseAutoRefresh
}

// Markdown-specific configuration getters
func (cm *ConfigManager) UseGFM() bool {
	return cm.config.Markdown.UseGFM
}

func (cm *ConfigManager) UsePHPMDExt() bool {
	return cm.config.Markdown.UsePHPMDExt
}

func (cm *ConfigManager) UseEmoji() bool {
	return cm.config.Markdown.UseEmoji
}

func (cm *ConfigManager) UseMermaid() bool {
	return cm.config.Markdown.UseMermaid
}

func (cm *ConfigManager) UseFigure() bool {
	return cm.config.Markdown.UseFigure
}

func (cm *ConfigManager) UseAnchor() bool {
	return cm.config.Markdown.UseAnchor
}

func (cm *ConfigManager) UseFences() bool {
	return cm.config.Markdown.UseFences
}

func (cm *ConfigManager) UseSections() bool {
	return cm.config.Markdown.UseSections
}

func (cm *ConfigManager) UseHighlighting() bool {
	return cm.config.Markdown.UseHighlighting
}

func (cm *ConfigManager) UseFancyLists() bool {
	return cm.config.Markdown.UseFancyLists
}

func (cm *ConfigManager) UseAttributes() bool {
	return cm.config.Markdown.UseAttributes
}

func (cm *ConfigManager) UseTypographic() bool {
	return cm.config.Markdown.UseTypographic
}

func (cm *ConfigManager) UseAbbreviations() bool {
	return cm.config.Markdown.UseAbbreviations
}

func (cm *ConfigManager) UseKatex() bool {
	return cm.config.Markdown.UseKatex
}

func (cm *ConfigManager) UseD2Diagrams() bool {
	return cm.config.Markdown.UseD2Diagrams
}

// Alert callouts configuration getters
func (cm *ConfigManager) UseAlertCallouts() bool {
	return cm.config.AlertCallouts.UseAlertCallouts
}

func (cm *ConfigManager) AlertCalloutStyle() string {
	return cm.config.AlertCallouts.AlertCalloutStyle
}

// Font configuration getters and setters
func (cm *ConfigManager) GetFontFamily() string {
	return cm.config.Application.FontFamily
}

func (cm *ConfigManager) SetFontFamily(fontFamily string) {
	cm.config.Application.FontFamily = fontFamily
}

func (cm *ConfigManager) GetFontSize() float64 {
	return cm.config.Application.FontSize
}

func (cm *ConfigManager) SetFontSize(fontSize float64) {
	cm.config.Application.FontSize = fontSize
}

func (cm *ConfigManager) GetFontFamilyMono() string {
	return cm.config.Application.FontFamilyMono
}

func (cm *ConfigManager) SetFontFamilyMono(fontFamily string) {
	cm.config.Application.FontFamilyMono = fontFamily
}

func (cm *ConfigManager) GetFontSizeMono() float64 {
	return cm.config.Application.FontSizeMono
}

func (cm *ConfigManager) SetFontSizeMono(fontSize float64) {
	cm.config.Application.FontSizeMono = fontSize
}

func (cm *ConfigManager) GetUseAdvancedFontDetection() bool {
	return cm.config.Application.UseAdvancedFontDetection
}

func (cm *ConfigManager) SetUseAdvancedFontDetection(useAdvanced bool) {
	cm.config.Application.UseAdvancedFontDetection = useAdvanced
}

func (cm *ConfigManager) SetUseAutoRefresh(useAutoRefresh bool) {
	cm.config.Application.UseAutoRefresh = useAutoRefresh
}
