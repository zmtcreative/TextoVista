package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// FontManager handles system font enumeration
type FontManager struct {
	mu                  sync.RWMutex
	availableFonts      []string
	knownMonospaceFonts map[string]bool
	configManager       *ConfigManager
	// Caching fields for performance
	fontCache       map[string]bool // Cache for font existence checks
	monoCache       map[string]bool // Cache for monospace detection results
	systemFontCache []string        // Cache for system fonts
}

// NewFontManager creates a new FontManager
func NewFontManager(configManager *ConfigManager) *FontManager {
	return &FontManager{
		availableFonts:  []string{},
		configManager:   configManager,
		fontCache:       make(map[string]bool),
		monoCache:       make(map[string]bool),
		systemFontCache: []string{},
		knownMonospaceFonts: map[string]bool{
			"Consolas":         true,
			"Courier New":      true,
			"Courier":          true,
			"Monaco":           true,
			"DejaVu Sans Mono": true,
			"Liberation Mono":  true,
			"Lucida Console":   true,
			"Menlo":            true,
			"Source Code Pro":  true,
			"Fira Code":        true,
			"JetBrains Mono":   true,
			"Cascadia Code":    true,
			"Ubuntu Mono":      true,
			"Roboto Mono":      true,
			"SF Mono":          true,
			"Inconsolata":      true,
			"Anonymous Pro":    true,
			"Hack":             true,
			"IBM Plex Mono":    true,
			"Space Mono":       true,
		},
	}
}

// GetSystemFonts returns a list of available system fonts
func (fm *FontManager) GetSystemFonts() []string {
	// Return cached result if available
	fm.mu.RLock()
	if len(fm.systemFontCache) > 0 {
		result := make([]string, len(fm.systemFontCache))
		copy(result, fm.systemFontCache)
		fm.mu.RUnlock()
		return result
	}
	fm.mu.RUnlock()

	var result []string
	switch runtime.GOOS {
	case "windows":
		result = fm.getWindowsFonts()
	case "linux":
		result = fm.getLinuxFonts()
	case "darwin":
		result = fm.getMacOSFonts()
	default:
		result = fm.getDefaultFonts()
	}

	// Cache the result under lock
	fm.mu.Lock()
	fm.systemFontCache = result
	out := make([]string, len(result))
	copy(out, result)
	fm.mu.Unlock()
	return out
}

// getWindowsFonts retrieves fonts on Windows
func (fm *FontManager) getWindowsFonts() []string {
	fonts := make([]string, 0)

	// Try registry method first (more comprehensive)
	registryFonts := fm.getWindowsFontsFromRegistry()
	if len(registryFonts) > 0 {
		fonts = append(fonts, registryFonts...)
	}

	// If registry method fails, fall back to default fonts
	if len(fonts) == 0 {
		fonts = fm.getDefaultFonts()
	}

	// Always include web-safe and common fonts
	webSafeFonts := []string{
		"Arial, sans-serif",
		"Helvetica, sans-serif",
		"Times New Roman, serif",
		"Georgia, serif",
		"Verdana, sans-serif",
		"Courier New, monospace",
		"Comic Sans MS, cursive",
		"Impact, fantasy",
		"Lucida Console, monospace",
		"Tahoma, sans-serif",
		"Trebuchet MS, sans-serif",
	}

	// Merge and deduplicate
	fontMap := make(map[string]bool)
	result := make([]string, 0)

	// Add web-safe fonts first
	for _, font := range webSafeFonts {
		if !fontMap[font] {
			fontMap[font] = true
			result = append(result, font)
		}
	}

	// Add system fonts
	for _, font := range fonts {
		fontName := strings.TrimSpace(font)
		if fontName != "" && !fontMap[fontName] {
			fontMap[fontName] = true
			result = append(result, fontName)
		}
	}

	return result
}

// getLinuxFonts retrieves fonts on Linux
func (fm *FontManager) getLinuxFonts() []string {
	fonts := fm.getDefaultFonts()

	// Try to get actual system fonts
	systemFonts := fm.getLinuxSystemFonts()
	if len(systemFonts) > 0 {
		// Merge system fonts with defaults
		fontMap := make(map[string]bool)
		result := make([]string, 0)

		// Add default fonts first
		for _, font := range fonts {
			if !fontMap[font] {
				fontMap[font] = true
				result = append(result, font)
			}
		}

		// Add system fonts
		for _, font := range systemFonts {
			fontName := strings.TrimSpace(font)
			if fontName != "" && !fontMap[fontName] {
				fontMap[fontName] = true
				result = append(result, fontName)
			}
		}

		return result
	}

	return fonts
}

// getMacOSFonts retrieves fonts on macOS
func (fm *FontManager) getMacOSFonts() []string {
	fonts := fm.getDefaultFonts()

	// TODO: Implement Core Text API or system font directory scanning for macOS

	return fonts
}

// getDefaultFonts returns a list of common web-safe and system fonts
func (fm *FontManager) getDefaultFonts() []string {
	return []string{
		"Arial, sans-serif",
		"Helvetica, sans-serif",
		"Times New Roman, serif",
		"Georgia, serif",
		"Verdana, sans-serif",
		"Tahoma, sans-serif",
		"Trebuchet MS, sans-serif",
		"Comic Sans MS, cursive",
		"Courier New, monospace",
		"Lucida Console, monospace",
		"Impact, fantasy",
		"Palatino, serif",
		"Book Antiqua, serif",
		"Century Gothic, sans-serif",
		"Calibri, sans-serif",
		"Cambria, serif",
		"Consolas, monospace",
		"Segoe UI, sans-serif",
	}
}

// GetFontFamilyName extracts just the font family name from a font string
func (fm *FontManager) GetFontFamilyName(fontString string) string {
	// Extract the first font name before any comma
	if strings.Contains(fontString, ",") {
		return strings.TrimSpace(strings.Split(fontString, ",")[0])
	}
	return strings.TrimSpace(fontString)
}

// ValidateFontFamily checks if a font family is in the available list
func (fm *FontManager) ValidateFontFamily(fontFamily string) bool {
	fonts := fm.GetSystemFonts()
	for _, font := range fonts {
		if strings.EqualFold(fm.GetFontFamilyName(font), fm.GetFontFamilyName(fontFamily)) {
			return true
		}
	}
	return false
}

// IsMonospaceFont checks if a font is known to be monospace
func (fm *FontManager) IsMonospaceFont(fontName string) bool {
	// Clean font name and check against known list
	cleanName := fm.GetFontFamilyName(fontName)
	return fm.knownMonospaceFonts[cleanName]
}

// GetMonospaceFonts returns a list of available monospace fonts using the configured detection method
func (fm *FontManager) GetMonospaceFonts() []string {
	if fm.configManager != nil && fm.configManager.GetUseAdvancedFontDetection() {
		return fm.AdvancedMonospaceDetection()
	}
	return fm.BasicMonospaceDetection()
}

// BasicMonospaceDetection uses the original heuristic approach
func (fm *FontManager) BasicMonospaceDetection() []string {
	allFonts := fm.GetSystemFonts()
	monospaceFonts := make([]string, 0)

	for _, font := range allFonts {
		if fm.IsMonospaceFont(font) {
			monospaceFonts = append(monospaceFonts, font)
		}
	}

	return fm.deduplicateAndAddFallbacks(monospaceFonts)
} // AdvancedMonospaceDetection performs more sophisticated monospace font detection
func (fm *FontManager) AdvancedMonospaceDetection() []string {
	allFonts := fm.GetSystemFonts()
	advancedMonospaceFonts := make([]string, 0)

	for _, font := range allFonts {
		cleanName := fm.GetFontFamilyName(font)

		// First check against known list for quick identification
		if fm.IsMonospaceFont(font) {
			advancedMonospaceFonts = append(advancedMonospaceFonts, font)
			continue
		}

		// Check cache first under read lock
		fm.mu.RLock()
		cached, exists := fm.monoCache[cleanName]
		fm.mu.RUnlock()
		if exists {
			if cached {
				advancedMonospaceFonts = append(advancedMonospaceFonts, font)
			}
			continue
		}

		// Advanced detection methods
		isMonospace := fm.detectMonospaceByName(cleanName)
		if !isMonospace {
			// Only do expensive file operations if name-based detection fails
			// and we haven't already checked this font
			isMonospace = fm.detectMonospaceByFontPathOptimized(cleanName)
		}

		// Cache the result under write lock
		fm.mu.Lock()
		fm.monoCache[cleanName] = isMonospace
		fm.mu.Unlock()

		if isMonospace {
			advancedMonospaceFonts = append(advancedMonospaceFonts, font)
		}
	}

	return fm.deduplicateAndAddFallbacks(advancedMonospaceFonts)
}

// detectMonospaceByName uses advanced naming pattern analysis
func (fm *FontManager) detectMonospaceByName(fontName string) bool {
	lowerName := strings.ToLower(fontName)

	// Common monospace indicators in font names
	monoIndicators := []string{
		"mono", "code", "console", "terminal", "typewriter",
		"courier", "fixed", "programming", "dev", "source",
		"pt mono", "nerd font", "powerline", "cascadia",
	}

	for _, indicator := range monoIndicators {
		if strings.Contains(lowerName, indicator) {
			// Additional validation to avoid false positives
			if !fm.isLikelyProportionalFont(lowerName) {
				return true
			}
		}
	}

	return false
}

// isLikelyProportionalFont helps avoid false positives
func (fm *FontManager) isLikelyProportionalFont(fontName string) bool {
	proportionalIndicators := []string{
		"display", "text", "serif", "sans", "script",
		"decorative", "handwriting", "calligraphy",
	}

	for _, indicator := range proportionalIndicators {
		if strings.Contains(fontName, indicator) {
			return true
		}
	}

	return false
}

// detectMonospaceByFontPathOptimized is an optimized version that limits expensive operations
func (fm *FontManager) detectMonospaceByFontPathOptimized(fontName string) bool {
	// Skip file system operations for very common non-monospace fonts
	lowerName := strings.ToLower(fontName)
	commonProportionalFonts := []string{
		"arial", "helvetica", "times", "georgia", "verdana", "tahoma",
		"trebuchet", "comic", "impact", "palatino", "calibri", "cambria",
	}

	for _, proportional := range commonProportionalFonts {
		if strings.Contains(lowerName, proportional) {
			return false
		}
	}

	// Only proceed with file operations if the font name suggests it might be monospace
	if !fm.couldBeMonospaceByName(fontName) {
		return false
	}

	// Get potential font file paths based on OS
	fontPaths := fm.getFontDirectories()

	// Limit to fewer extensions and files to check
	extensions := []string{".ttf", ".otf"}

	for _, basePath := range fontPaths {
		// Check if the directory exists first to avoid unnecessary operations
		if !fm.fileExistsWithCache(basePath) {
			continue
		}

		for _, ext := range extensions {
			// Generate only the most likely filenames
			possibleFiles := fm.generateLimitedFontFilenames(fontName, ext)

			for _, filename := range possibleFiles {
				fullPath := filepath.Join(basePath, filename)
				if fm.fileExistsWithCache(fullPath) {
					// Found a font file, try to analyze it
					if fm.analyzeFontFile(fullPath) {
						return true
					}
				}
			}
		}
	}

	return false
}

// couldBeMonospaceByName does a quick check if a font name suggests monospace
func (fm *FontManager) couldBeMonospaceByName(fontName string) bool {
	lowerName := strings.ToLower(fontName)

	// Quick indicators that suggest monospace
	monoHints := []string{
		"mono", "code", "console", "terminal", "courier", "fixed",
		"typewriter", "programming", "dev", "source", "nerd", "cascadia",
	}

	for _, hint := range monoHints {
		if strings.Contains(lowerName, hint) {
			return true
		}
	}

	return false
}

// generateLimitedFontFilenames creates fewer, more likely font filenames
func (fm *FontManager) generateLimitedFontFilenames(fontName string, ext string) []string {
	var filenames []string

	// Most common patterns only
	filenames = append(filenames, fontName+ext)

	// Replace spaces with the most common substitute
	noSpaces := strings.ReplaceAll(fontName, " ", "")
	if noSpaces != fontName {
		filenames = append(filenames, noSpaces+ext)
	}

	// Add Regular suffix (most common)
	filenames = append(filenames, fontName+"Regular"+ext)
	filenames = append(filenames, noSpaces+"Regular"+ext)

	return filenames
}

// fileExistsWithCache checks file existence with caching
func (fm *FontManager) fileExistsWithCache(filename string) bool {
	fm.mu.RLock()
	cached, exists := fm.fontCache[filename]
	fm.mu.RUnlock()
	if exists {
		return cached
	}

	_, err := os.Stat(filename)
	result := !os.IsNotExist(err)

	// Cache the result under write lock
	fm.mu.Lock()
	fm.fontCache[filename] = result
	fm.mu.Unlock()
	return result
}

// // detectMonospaceByFontPath attempts to find and analyze font files
// func (fm *FontManager) detectMonospaceByFontPath(fontName string) bool {
// 	// Use the optimized version for better performance
// 	return fm.detectMonospaceByFontPathOptimized(fontName)
// }

// getFontDirectories returns OS-specific font directories
func (fm *FontManager) getFontDirectories() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			filepath.Join(os.Getenv("WINDIR"), "Fonts"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Fonts"),
		}
	case "darwin":
		return []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
			filepath.Join(os.Getenv("HOME"), "Library", "Fonts"),
		}
	case "linux":
		return []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(os.Getenv("HOME"), ".fonts"),
			filepath.Join(os.Getenv("HOME"), ".local", "share", "fonts"),
		}
	default:
		return []string{}
	}
}

// // generateFontFilenames creates possible font filenames from font name
// func (fm *FontManager) generateFontFilenames(fontName string, ext string) []string {
// 	var filenames []string

// 	// Original name
// 	filenames = append(filenames, fontName+ext)

// 	// Replace spaces with common substitutes
// 	replacements := []string{
// 		strings.ReplaceAll(fontName, " ", ""),      // NoSpaces
// 		strings.ReplaceAll(fontName, " ", "-"),     // Dashes
// 		strings.ReplaceAll(fontName, " ", "_"),     // Underscores
// 	}

// 	for _, replacement := range replacements {
// 		if replacement != fontName {
// 			filenames = append(filenames, replacement+ext)
// 		}
// 	}

// 	// Add common style suffixes
// 	styles := []string{"Regular", "Normal", "Medium"}
// 	for _, style := range styles {
// 		filenames = append(filenames, fontName+style+ext)
// 		for _, replacement := range replacements {
// 			filenames = append(filenames, replacement+style+ext)
// 		}
// 	}

// 	return filenames
// }

// // fileExists checks if a file exists
// func (fm *FontManager) fileExists(filename string) bool {
// 	return fm.fileExistsWithCache(filename)
// }

// analyzeFontFile performs basic font file analysis to detect monospace characteristics
func (fm *FontManager) analyzeFontFile(filePath string) bool {
	// Open and read the font file
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 1KB to analyze font headers
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil || n < 100 {
		return false
	}

	// Basic font format detection and analysis
	return fm.analyzeFontHeader(buffer[:n])
}

// analyzeFontHeader analyzes font file headers for monospace indicators
func (fm *FontManager) analyzeFontHeader(data []byte) bool {
	// Check for TrueType/OpenType signatures
	if len(data) < 12 {
		return false
	}

	// Look for font format signatures
	signature := string(data[0:4])
	switch signature {
	case "OTTO", "\x00\x01\x00\x00": // OpenType/TrueType
		return fm.analyzeTrueTypeHeader(data)
	case "wOFF", "wOF2": // Web fonts
		return fm.analyzeWebFontHeader(data)
	}

	// For other formats, fall back to content analysis
	return fm.analyzeGenericFontData(data)
}

// analyzeTrueTypeHeader analyzes TrueType/OpenType font headers
func (fm *FontManager) analyzeTrueTypeHeader(data []byte) bool {
	// This is a simplified analysis - in a full implementation,
	// you would parse the font tables to get accurate metrics

	// Look for monospace-related metadata in the first 1KB
	dataStr := strings.ToLower(string(data))

	// Check for monospace indicators in font metadata
	monoIndicators := []string{
		"monospace", "mono", "fixed", "courier", "consola",
		"terminal", "typewriter",
	}

	for _, indicator := range monoIndicators {
		if strings.Contains(dataStr, indicator) {
			return true
		}
	}

	return false
}

// analyzeWebFontHeader analyzes web font headers
func (fm *FontManager) analyzeWebFontHeader(data []byte) bool {
	// Similar to TrueType analysis but for web fonts
	dataStr := strings.ToLower(string(data))

	monoIndicators := []string{
		"monospace", "mono", "fixed", "courier", "consola",
	}

	for _, indicator := range monoIndicators {
		if strings.Contains(dataStr, indicator) {
			return true
		}
	}

	return false
}

// analyzeGenericFontData performs generic font data analysis
func (fm *FontManager) analyzeGenericFontData(data []byte) bool {
	// Look for patterns in the binary data that might indicate monospace fonts
	// This is a heuristic approach

	dataStr := strings.ToLower(string(data))

	// Check for any monospace-related strings in the font data
	if strings.Contains(dataStr, "monospace") ||
		strings.Contains(dataStr, "fixed") ||
		strings.Contains(dataStr, "courier") {
		return true
	}

	return false
}

// deduplicateAndAddFallbacks ensures unique fonts and adds fallbacks
func (fm *FontManager) deduplicateAndAddFallbacks(fonts []string) []string {
	fontMap := make(map[string]bool)
	result := make([]string, 0)

	// Add detected fonts
	for _, font := range fonts {
		cleanName := fm.GetFontFamilyName(font)
		if !fontMap[cleanName] {
			fontMap[cleanName] = true
			result = append(result, font)
		}
	}

	// Add fallback fonts if not already present
	fallbackFonts := []string{
		"Consolas, monospace",
		"Monaco, monospace",
		"Courier New, monospace",
		"DejaVu Sans Mono, monospace",
		"Liberation Mono, monospace",
		"Lucida Console, monospace",
	}

	for _, font := range fallbackFonts {
		cleanName := fm.GetFontFamilyName(font)
		if !fontMap[cleanName] {
			fontMap[cleanName] = true
			result = append(result, font)
		}
	}

	return result
}
