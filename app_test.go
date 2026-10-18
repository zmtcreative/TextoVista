package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"texto-vista/internal/cli"
)

func TestBoolFromPtr(t *testing.T) {
	tests := []struct {
		name         string
		ptr          *bool
		defaultValue bool
		expected     bool
	}{
		{
			name:         "nil pointer with true default",
			ptr:          nil,
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "nil pointer with false default",
			ptr:          nil,
			defaultValue: false,
			expected:     false,
		},
		{
			name:         "non-nil pointer with true value",
			ptr:          boolPtr(true),
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "non-nil pointer with false value",
			ptr:          boolPtr(false),
			defaultValue: true,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := boolFromPtr(tt.ptr, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("boolFromPtr() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStringFromPtr(t *testing.T) {
	tests := []struct {
		name         string
		ptr          *string
		defaultValue string
		expected     string
	}{
		{
			name:         "nil pointer with default",
			ptr:          nil,
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "nil pointer with empty default",
			ptr:          nil,
			defaultValue: "",
			expected:     "",
		},
		{
			name:         "non-nil pointer with value",
			ptr:          stringPtr("test"),
			defaultValue: "default",
			expected:     "test",
		},
		{
			name:         "non-nil pointer with empty value",
			ptr:          stringPtr(""),
			defaultValue: "default",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringFromPtr(tt.ptr, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("stringFromPtr() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSetAbout(t *testing.T) {
	tests := []struct {
		name                string
		appProgNameWithExt  string
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:               "standard app name",
			appProgNameWithExt: "texto-vista.exe",
			expectedContains: []string{
				"texto-vista.exe",
				"Copyright 2026",
				"Version",
				Date, // Use the actual variable instead of "BuildDate"
			},
			expectedNotContains: []string{},
		},
		{
			name:               "custom app name",
			appProgNameWithExt: "markdown-viewer",
			expectedContains: []string{
				"markdown-viewer",
				"Copyright 2026",
			},
			expectedNotContains: []string{},
		},
		{
			name:               "empty app name",
			appProgNameWithExt: "",
			expectedContains: []string{
				"Copyright 2026",
				"Version",
			},
			expectedNotContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := setAbout(tt.appProgNameWithExt)

			// Check that no error occurred
			if err != nil {
				t.Errorf("setAbout() returned unexpected error: %v", err)
			}

			// Check that result is not empty
			if result == "" {
				t.Error("setAbout() returned empty string")
			}

			// Check expected content
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("setAbout() result missing expected content: %q", expected)
				}
			}

			// Check content that should not be present
			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("setAbout() result contains unexpected content: %q", notExpected)
				}
			}

			// Check that it looks like HTML (basic sanity check)
			if !strings.Contains(result, "<") || !strings.Contains(result, ">") {
				t.Error("setAbout() result does not appear to be HTML")
			}
		})
	}
}

func TestNewApp(t *testing.T) {
	tests := []struct {
		name     string
		cliArgs  *cli.CliArgs
		validate func(t *testing.T, app *App)
	}{
		{
			name: "default values",
			cliArgs: &cli.CliArgs{
				InitialFile:         nil,
				ShowHelp:            nil,
				CmdlineOptions:      stringPtr("test options"),
				AppProgName:         stringPtr("texto-vista"),
				AppProgNameWithExt:  stringPtr("texto-vista.exe"),
			},
			validate: func(t *testing.T, app *App) {
				if app.currentFile != "" {
					t.Errorf("NewApp() currentFile = %q, want empty string", app.currentFile)
				}
				if app.showHelp != false {
					t.Errorf("NewApp() showHelp = %v, want false", app.showHelp)
				}
				if app.appProgName != "texto-vista" {
					t.Errorf("NewApp() appProgName = %q, want %q", app.appProgName, "texto-vista")
				}
				if app.appProgNameWithExt != "texto-vista.exe" {
					t.Errorf("NewApp() appProgNameWithExt = %q, want %q", app.appProgNameWithExt, "texto-vista.exe")
				}
				if app.cmdlineOptions != "test options" {
					t.Errorf("NewApp() cmdlineOptions = %q, want %q", app.cmdlineOptions, "test options")
				}
			},
		},
		{
			name: "with initial file and help",
			cliArgs: &cli.CliArgs{
				InitialFile:         stringPtr("test.md"),
				ShowHelp:            boolPtr(true),
				CmdlineOptions:      stringPtr("help text"),
				AppProgName:         stringPtr("custom-reader"),
				AppProgNameWithExt:  stringPtr("custom-reader"),
			},
			validate: func(t *testing.T, app *App) {
				if app.currentFile != "test.md" {
					t.Errorf("NewApp() currentFile = %q, want %q", app.currentFile, "test.md")
				}
				if app.showHelp != true {
					t.Errorf("NewApp() showHelp = %v, want true", app.showHelp)
				}
				if app.appProgName != "custom-reader" {
					t.Errorf("NewApp() appProgName = %q, want %q", app.appProgName, "custom-reader")
				}
			},
		},
		{
			name: "nil values",
			cliArgs: &cli.CliArgs{
				InitialFile:         nil,
				ShowHelp:            nil,
				CmdlineOptions:      nil,
				AppProgName:         nil,
				AppProgNameWithExt:  nil,
			},
			validate: func(t *testing.T, app *App) {
				if app.currentFile != "" {
					t.Errorf("NewApp() currentFile = %q, want empty string", app.currentFile)
				}
				if app.showHelp != false {
					t.Errorf("NewApp() showHelp = %v, want false", app.showHelp)
				}
				if app.appProgName != "texto-vista" {
					t.Errorf("NewApp() appProgName = %q, want %q", app.appProgName, "texto-vista")
				}
				if app.appProgNameWithExt != "texto-vista" {
					t.Errorf("NewApp() appProgNameWithExt = %q, want %q", app.appProgNameWithExt, "texto-vista")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.cliArgs)

			if app == nil {
				t.Fatal("NewApp() returned nil")
			}

			// Check that configManager is initialized
			if app.configManager == nil {
				t.Error("NewApp() configManager is nil")
			}

			// Check that versionInfo is not empty
			if app.versionInfo == "" {
				t.Error("NewApp() versionInfo is empty")
			}

			// Run specific validation
			if tt.validate != nil {
				tt.validate(t, app)
			}
		})
	}
}

func TestAppConfigDelegation(t *testing.T) {
	// Create a basic app instance for testing delegation methods
	cliArgs := &cli.CliArgs{
		AppProgName:        stringPtr("test-app"),
		AppProgNameWithExt: stringPtr("test-app.exe"),
	}

	app := NewApp(cliArgs)
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	// Test GetSettings delegation
	settings := app.GetSettings()
	if settings == nil {
		t.Error("GetSettings() returned nil")
	}

	// Test GetAlertCalloutStyles delegation
	styles := app.GetAlertCalloutStyles()
	if styles == nil {
		t.Error("GetAlertCalloutStyles() returned nil")
	}

	// Should contain expected alert callout styles
	expectedKeys := []string{"GFMStrict", "GFMWithAliases", "GFMPlus", "Obsidian"}
	for _, key := range expectedKeys {
		if _, exists := styles[key]; !exists {
			t.Errorf("GetAlertCalloutStyles() missing key: %s", key)
		}
	}

	// Skip GetTheme() test since it requires context initialization
	// which happens in the startup() method, not in NewApp()
}

func TestAppVersionInfo(t *testing.T) {
	cliArgs := &cli.CliArgs{
		AppProgName:        stringPtr("test-texto-vista"),
		AppProgNameWithExt: stringPtr("test-texto-vista.exe"),
	}

	app := NewApp(cliArgs)
	if app == nil {
		t.Fatal("NewApp() returned nil")
	}

	versionInfo := app.versionInfo
	if versionInfo == "" {
		t.Error("App versionInfo is empty")
	}

	// Check that version info contains expected elements
	expectedElements := []string{
		"test-texto-vista.exe",
		"Copyright 2026",
	}

	for _, element := range expectedElements {
		if !strings.Contains(versionInfo, element) {
			t.Errorf("versionInfo missing expected element: %q", element)
		}
	}
}

func TestInitialFileLoadErrorMessage(t *testing.T) {
	err := initialFileLoadErrorMessage(assertError("file not found: missing.md"))

	if err != "Failed to load initial file: file not found: missing.md" {
		t.Fatalf("initialFileLoadErrorMessage() = %q", err)
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

type assertError string

func (e assertError) Error() string {
	return string(e)
}

func TestAppSaveSettingsNilSafety(t *testing.T) {
	app := NewApp(&cli.CliArgs{})
	if err := app.SaveSettings(nil); err == nil {
		t.Fatal("SaveSettings(nil) expected error, got nil")
	}
	if err := app.SaveSettingsSessionOnly(nil); err == nil {
		t.Fatal("SaveSettingsSessionOnly(nil) expected error, got nil")
	}
}

func TestAppLoadAndDisplayMarkdownValidation(t *testing.T) {
	app := NewApp(&cli.CliArgs{})

	// Test empty path
	if err := app.LoadAndDisplayMarkdown(""); err == nil {
		t.Fatal("LoadAndDisplayMarkdown(\"\") expected error, got nil")
	}

	// Test non-existent file
	if err := app.LoadAndDisplayMarkdown("non_existent_file_xyz_123.md"); err == nil {
		t.Fatal("LoadAndDisplayMarkdown(missing) expected error, got nil")
	}

	// Test directory path
	tempDir := t.TempDir()
	if err := app.LoadAndDisplayMarkdown(tempDir); err == nil {
		t.Fatal("LoadAndDisplayMarkdown(directory) expected error, got nil")
	}

	// Test path escaping parent directory
	subDir := filepath.Join(tempDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	baseFile := filepath.Join(subDir, "base.md")
	if err := os.WriteFile(baseFile, []byte("# Base"), 0644); err != nil {
		t.Fatal(err)
	}
	app.setCurrentFile(baseFile)

	// Try relative path escaping parent directory
	if err := app.LoadAndDisplayMarkdown("../../outside.md"); err == nil {
		t.Fatal("LoadAndDisplayMarkdown(path escaping baseDir) expected error, got nil")
	}
}
