package cli

import (
	"os"
	"strings"
	"testing"
)

func TestGetArgs(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedFile   *string
		expectedHelp   *bool
		expectError    bool
	}{
		{
			name:         "no arguments",
			args:         []string{"texto-vista"},
			expectedFile: nil,
			expectedHelp: nil,
		},
		{
			name:         "file flag with value",
			args:         []string{"texto-vista", "-f", "test.md"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: nil,
		},
		{
			name:         "file flag long form with value",
			args:         []string{"texto-vista", "--file", "test.md"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: nil,
		},
		{
			name:         "positional file argument",
			args:         []string{"texto-vista", "test.md"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: nil,
		},
		{
			name:         "help flag short form",
			args:         []string{"texto-vista", "-h"},
			expectedFile: nil,
			expectedHelp: boolPtr(true),
		},
		{
			name:         "help flag long form",
			args:         []string{"texto-vista", "--help"},
			expectedFile: nil,
			expectedHelp: boolPtr(true),
		},
		{
			name:         "file flag and help flag",
			args:         []string{"texto-vista", "-f", "test.md", "-h"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: boolPtr(true),
		},
		{
			name:         "positional argument takes precedence when both specified",
			args:         []string{"texto-vista", "-f", "flag.md", "positional.md"},
			expectedFile: stringPtr("positional.md"),
			expectedHelp: nil,
		},
		{
			name:         "multiple positional arguments - first one wins",
			args:         []string{"texto-vista", "first.md", "second.md"},
			expectedFile: stringPtr("first.md"),
			expectedHelp: nil,
		},
		{
			name:         "unknown long flag is ignored",
			args:         []string{"texto-vista", "--unknown"},
			expectedFile: nil,
			expectedHelp: nil,
		},
		{
			name:         "unknown flag before file flag is ignored",
			args:         []string{"texto-vista", "--unknown", "--file", "test.md"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: nil,
		},
		{
			name:         "unknown short flag before positional file is ignored",
			args:         []string{"texto-vista", "-z", "test.md"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: nil,
		},
		{
			name:         "unknown equals flag is ignored",
			args:         []string{"texto-vista", "--unknown=value", "--file", "test.md"},
			expectedFile: stringPtr("test.md"),
			expectedHelp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			args, err := GetArgs()

			if tt.expectError && err == nil {
				t.Errorf("GetArgs() expected error, got nil")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("GetArgs() unexpected error: %v", err)
				return
			}

			// Test InitialFile
			if tt.expectedFile == nil && args.InitialFile != nil {
				t.Errorf("GetArgs() InitialFile = %v, want nil", *args.InitialFile)
			} else if tt.expectedFile != nil {
				if args.InitialFile == nil {
					t.Errorf("GetArgs() InitialFile = nil, want %s", *tt.expectedFile)
				} else if *args.InitialFile != *tt.expectedFile {
					t.Errorf("GetArgs() InitialFile = %s, want %s", *args.InitialFile, *tt.expectedFile)
				}
			}

			// Test ShowHelp
			if tt.expectedHelp == nil && args.ShowHelp != nil {
				t.Errorf("GetArgs() ShowHelp = %v, want nil", *args.ShowHelp)
			} else if tt.expectedHelp != nil {
				if args.ShowHelp == nil {
					t.Errorf("GetArgs() ShowHelp = nil, want %v", *tt.expectedHelp)
				} else if *args.ShowHelp != *tt.expectedHelp {
					t.Errorf("GetArgs() ShowHelp = %v, want %v", *args.ShowHelp, *tt.expectedHelp)
				}
			}
		})
	}
}

func TestGetArgsAppNames(t *testing.T) {
	tests := []struct {
		name                       string
		args                       []string
		expectedAppProgName        string
		expectedAppProgNameWithExt string
	}{
		{
			name:                       "standard executable name",
			args:                       []string{"texto-vista"},
			expectedAppProgName:        "texto-vista",
			expectedAppProgNameWithExt: "texto-vista",
		},
		{
			name:                       "executable with extension",
			args:                       []string{"texto-vista.exe"},
			expectedAppProgName:        "texto-vista",
			expectedAppProgNameWithExt: "texto-vista.exe",
		},
		{
			name:                       "full path executable",
			args:                       []string{"/usr/local/bin/texto-vista"},
			expectedAppProgName:        "texto-vista",
			expectedAppProgNameWithExt: "texto-vista",
		},
		{
			name:                       "windows path with extension",
			args:                       []string{"C:\\Program Files\\texto-vista\\texto-vista.exe"},
			expectedAppProgName:        "texto-vista",
			expectedAppProgNameWithExt: "texto-vista.exe",
		},
		{
			name:                       "different executable name",
			args:                       []string{"markdown-viewer.exe"},
			expectedAppProgName:        "markdown-viewer",
			expectedAppProgNameWithExt: "markdown-viewer.exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = tt.args

			args, err := GetArgs()
			if err != nil {
				t.Fatalf("GetArgs() unexpected error: %v", err)
			}

			if args.AppProgName == nil || *args.AppProgName != tt.expectedAppProgName {
				var got string
				if args.AppProgName != nil {
					got = *args.AppProgName
				}
				t.Errorf("GetArgs() AppProgName = %q, want %q", got, tt.expectedAppProgName)
			}

			if args.AppProgNameWithExt == nil || *args.AppProgNameWithExt != tt.expectedAppProgNameWithExt {
				var got string
				if args.AppProgNameWithExt != nil {
					got = *args.AppProgNameWithExt
				}
				t.Errorf("GetArgs() AppProgNameWithExt = %q, want %q", got, tt.expectedAppProgNameWithExt)
			}
		})
	}
}

func TestGetArgsCmdlineOptions(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"texto-vista"}

	args, err := GetArgs()
	if err != nil {
		t.Fatalf("GetArgs() unexpected error: %v", err)
	}

	if args.CmdlineOptions == nil {
		t.Fatal("GetArgs() CmdlineOptions = nil, want non-nil")
	}

	cmdlineText := *args.CmdlineOptions

	// Check that the usage text contains expected elements
	expectedElements := []string{
		"<pre>",
		"Usage:",
		"texto-vista",
		"[options]",
		"[filepath]",
		"--file",
		"</pre>",
	}

	for _, element := range expectedElements {
		if !strings.Contains(cmdlineText, element) {
			t.Errorf("GetArgs() CmdlineOptions missing expected element: %q", element)
		}
	}
}

func TestGetArgsEmptyOsArgs(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test edge case where os.Args is empty (should not happen in practice)
	os.Args = []string{}

	args, err := GetArgs()
	if err != nil {
		t.Fatalf("GetArgs() unexpected error: %v", err)
	}

	// Should have reasonable defaults
	if args.AppProgName == nil || *args.AppProgName != "texto-vista" {
		var got string
		if args.AppProgName != nil {
			got = *args.AppProgName
		}
		t.Errorf("GetArgs() AppProgName = %q, want %q", got, "texto-vista")
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
