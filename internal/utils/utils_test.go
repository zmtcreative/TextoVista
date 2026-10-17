package utils

import (
	"os"
	"testing"
)

func TestGetValueFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]string
		key      string
		expected string
	}{
		{
			name:     "exact case match",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "lowercase key in mixed case map",
			m:        map[string]string{"Key": "value", "Another": "test"},
			key:      "key",
			expected: "value",
		},
		{
			name:     "uppercase key in mixed case map",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "KEY",
			expected: "value",
		},
		{
			name:     "title case key in lowercase map",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "Key",
			expected: "value",
		},
		{
			name:     "mixed case variations",
			m:        map[string]string{"Testkey": "testvalue", "other": "othervalue"},
			key:      "testkey",
			expected: "testvalue",
		},
		{
			name:     "key not found",
			m:        map[string]string{"key": "value", "another": "test"},
			key:      "notfound",
			expected: "",
		},
		{
			name:     "empty map",
			m:        map[string]string{},
			key:      "key",
			expected: "",
		},
		{
			name:     "empty key",
			m:        map[string]string{"key": "value", "": "empty"},
			key:      "",
			expected: "empty",
		},
		{
			name:     "special characters in key",
			m:        map[string]string{"key-with-dashes": "value", "key_with_underscores": "undervalue"},
			key:      "KEY-WITH-DASHES",
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetValueFromMap(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("GetValueFromMap() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetValueFromMapConsistency(t *testing.T) {
	// Test that the function consistently returns the same value for the same input
	m := map[string]string{"key": "value1", "KEY": "value2", "Key": "value3"}
	key := "key"

	first := GetValueFromMap(m, key)
	for i := 0; i < 10; i++ {
		result := GetValueFromMap(m, key)
		if result != first {
			t.Errorf("GetValueFromMap() returned inconsistent result on iteration %d: got %q, want %q", i, result, first)
		}
	}
}

func BenchmarkGetValueFromMap(b *testing.B) {
	m := map[string]string{
		"key1":     "value1",
		"key2":     "value2",
		"KEY3":     "value3",
		"Key4":     "value4",
		"TestKey5": "value5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetValueFromMap(m, "testkey5")
	}
}

func TestNormalizeMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		validate func(t *testing.T, got map[string]any)
	}{
		{
			name:  "nil input returns nil",
			input: nil,
			validate: func(t *testing.T, got map[string]any) {
				if got != nil {
					t.Fatalf("NormalizeMapKeys() = %#v, want nil", got)
				}
			},
		},
		{
			name: "top-level keys lowercased",
			input: map[string]any{
				"Title": "Example",
				"COUNT": 3,
			},
			validate: func(t *testing.T, got map[string]any) {
				if got["title"] != "Example" {
					t.Fatalf("got[title] = %#v, want %q", got["title"], "Example")
				}
				if got["count"] != 3 {
					t.Fatalf("got[count] = %#v, want %d", got["count"], 3)
				}
				if _, exists := got["Title"]; exists {
					t.Fatal("NormalizeMapKeys() kept original mixed-case key")
				}
			},
		},
		{
			name: "nested maps are normalized recursively",
			input: map[string]any{
				"Meta": map[string]any{
					"Author": "Ada",
					"DETAILS": map[string]any{
						"Reviewer": "Linus",
					},
				},
			},
			validate: func(t *testing.T, got map[string]any) {
				meta, ok := got["meta"].(map[string]any)
				if !ok {
					t.Fatalf("got[meta] type = %T, want map[string]any", got["meta"])
				}
				if meta["author"] != "Ada" {
					t.Fatalf("meta[author] = %#v, want %q", meta["author"], "Ada")
				}
				details, ok := meta["details"].(map[string]any)
				if !ok {
					t.Fatalf("meta[details] type = %T, want map[string]any", meta["details"])
				}
				if details["reviewer"] != "Linus" {
					t.Fatalf("details[reviewer] = %#v, want %q", details["reviewer"], "Linus")
				}
			},
		},
		{
			name: "non-map composite values are preserved",
			input: map[string]any{
				"Items": []string{"one", "two"},
				"Flags": map[string]bool{"A": true},
			},
			validate: func(t *testing.T, got map[string]any) {
				items, ok := got["items"].([]string)
				if !ok {
					t.Fatalf("got[items] type = %T, want []string", got["items"])
				}
				if len(items) != 2 || items[0] != "one" || items[1] != "two" {
					t.Fatalf("items = %#v, want [one two]", items)
				}
				flags, ok := got["flags"].(map[string]bool)
				if !ok {
					t.Fatalf("got[flags] type = %T, want map[string]bool", got["flags"])
				}
				if !flags["A"] {
					t.Fatalf("flags = %#v, want key A preserved", flags)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeMapKeys(tt.input)
			if err != nil {
				t.Fatalf("NormalizeMapKeys() error = %v", err)
			}
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	tests := []struct {
		name     string
		run      func(t *testing.T)
	}{
		{
			name: "nil map returns zero value and false",
			run: func(t *testing.T) {
				got, ok := GetValue[string](nil, "title")
				if ok {
					t.Fatal("GetValue() ok = true, want false")
				}
				if got != "" {
					t.Fatalf("GetValue() = %q, want empty string", got)
				}
			},
		},
		{
			name: "missing key returns zero value and false",
			run: func(t *testing.T) {
				got, ok := GetValue[int](map[string]any{"count": 3}, "missing")
				if ok {
					t.Fatal("GetValue() ok = true, want false")
				}
				if got != 0 {
					t.Fatalf("GetValue() = %d, want 0", got)
				}
			},
		},
		{
			name: "matching type returns value and true",
			run: func(t *testing.T) {
				got, ok := GetValue[string](map[string]any{"title": "Reader"}, "title")
				if !ok {
					t.Fatal("GetValue() ok = false, want true")
				}
				if got != "Reader" {
					t.Fatalf("GetValue() = %q, want %q", got, "Reader")
				}
			},
		},
		{
			name: "type mismatch returns zero value and false",
			run: func(t *testing.T) {
				got, ok := GetValue[string](map[string]any{"count": 3}, "count")
				if ok {
					t.Fatal("GetValue() ok = true, want false")
				}
				if got != "" {
					t.Fatalf("GetValue() = %q, want empty string", got)
				}
			},
		},
		{
			name: "matching map value preserves typed map",
			run: func(t *testing.T) {
				value := map[string]any{"title": "Nested"}
				got, ok := GetValue[map[string]any](map[string]any{"meta": value}, "meta")
				if !ok {
					t.Fatal("GetValue() ok = false, want true")
				}
				if got["title"] != "Nested" {
					t.Fatalf("GetValue() map = %#v, want nested title preserved", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestGetExecutableBaseName(t *testing.T) {
	// Save and restore original os.Args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	tests := []struct {
		name           string
		osArgs         []string
		wantWithExt    string
		wantWithoutExt string
	}{
		{
			name:           "Windows path with extension",
			osArgs:         []string{`C:\Program Files\App\texto-vista.exe`},
			wantWithExt:    "texto-vista.exe",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Windows path without extension",
			osArgs:         []string{`C:\Program Files\App\texto-vista`},
			wantWithExt:    "texto-vista",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Unix path with extension",
			osArgs:         []string{"/usr/local/bin/texto-vista.exe"},
			wantWithExt:    "texto-vista.exe",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Unix path without extension",
			osArgs:         []string{"/usr/local/bin/texto-vista"},
			wantWithExt:    "texto-vista",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Just basename no path",
			osArgs:         []string{"texto-vista.exe"},
			wantWithExt:    "texto-vista.exe",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Empty os.Args uses fallback",
			osArgs:         []string{},
			wantWithExt:    "texto-vista",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Windows relative path",
			osArgs:         []string{`.\texto-vista.exe`},
			wantWithExt:    "texto-vista.exe",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Unix relative path",
			osArgs:         []string{"./texto-vista"},
			wantWithExt:    "texto-vista",
			wantWithoutExt: "texto-vista",
		},
		{
			name:           "Dot fallback",
			osArgs:         []string{"."},
			wantWithExt:    "texto-vista",
			wantWithoutExt: "texto-vista",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.osArgs
			gotWithExt, gotWithoutExt := GetExecutableBaseName()
			if gotWithExt != tt.wantWithExt {
				t.Errorf("GetExecutableBaseName() withExt = %q, want %q", gotWithExt, tt.wantWithExt)
			}
			if gotWithoutExt != tt.wantWithoutExt {
				t.Errorf("GetExecutableBaseName() withoutExt = %q, want %q", gotWithoutExt, tt.wantWithoutExt)
			}
		})
	}
}
