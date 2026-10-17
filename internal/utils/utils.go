package utils

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GetExecutableBaseName extracts the application name from the executable path.
// It handles both Windows and Unix path separators and returns both the name
// with and without the file extension.
func GetExecutableBaseName() (withExt, withoutExt string) {
	if len(os.Args) == 0 {
		return "texto-vista", "texto-vista"
	}

	execPath := os.Args[0]

	// Handle both Windows and Unix path separators
	var baseName string
	if strings.Contains(execPath, "\\") {
		// Windows path - handle manually to work cross-platform
		parts := strings.Split(execPath, "\\")
		baseName = parts[len(parts)-1]
	} else {
		// Unix path or no path separators
		baseName = filepath.Base(execPath)
	}

	// Use fallback if empty or invalid
	if baseName == "" || baseName == "." {
		return "texto-vista", "texto-vista"
	}

	ext := filepath.Ext(baseName)
	if ext != "" {
		return baseName, strings.TrimSuffix(baseName, ext)
	}
	return baseName, baseName
}

// GetValueFromMap performs a case-insensitive search for a key in a map.
func GetValueFromMap(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}

	lowerKey := strings.ToLower(key)
	if v, ok := m[lowerKey]; ok {
		return v
	}

	upperKey := strings.ToUpper(key)
	if v, ok := m[upperKey]; ok {
		return v
	}

	caser := cases.Title(language.English, cases.Compact)
	titleKey := caser.String(key)
	if v, ok := m[titleKey]; ok {
		return v
	}

	return ""
}

// NormalizeMapKeys takes a map[string]any and converts all keys to lowercase recursively,
// returning a new map with normalized keys and the original values. It handles nested maps
// by recursively normalizing their keys as well.
func NormalizeMapKeys(input map[string]any) (map[string]any, error) {
	if input == nil {
		return nil, nil
	}

	normalized := make(map[string]any, len(input))
	for key, value := range input {
		lowerKey := strings.ToLower(key)

		// Check if the value is itself a map[string]any and recursively normalize it
		if nestedMap, ok := value.(map[string]any); ok {
			normalizedNested, err := NormalizeMapKeys(nestedMap)
			if err != nil {
				return nil, err
			}
			normalized[lowerKey] = normalizedNested
		} else {
			// For non-map values, just copy the value as-is
			normalized[lowerKey] = value
		}
	}

	return normalized, nil
}

// Generic helper function to retrieve a value from a map[string]any with a specific key.
// Usage: getValue[YourType](yourMap, "yourKey")
//   if title, ok := getValue[string](docFrontmatter, "title"); ok {
//       fmDocTitle = title
//   }
//   if date, ok := getValue[time.Time](docFrontmatter, "date"); ok {
//       fmDocDate = date.String()
//   }
func GetValue[T any](m map[string]any, key string) (T, bool) {
    var zero T
    if m == nil {
        return zero, false
    }
    if val, ok := m[key].(T); ok {
        return val, true
    }
    return zero, false
}



