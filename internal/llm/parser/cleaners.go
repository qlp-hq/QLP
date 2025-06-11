package parser

import (
	"context"
	"regexp"
	"strings"
	"unicode"
)

// MarkdownCodeBlockCleaner removes markdown code block markers
type MarkdownCodeBlockCleaner struct{}

func (mcbc *MarkdownCodeBlockCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	// Remove markdown code block markers
	patterns := []string{
		"```json",
		"```JSON", 
		"```",
		"`json",
		"`JSON",
		"`",
	}

	cleaned := rawResponse
	for _, pattern := range patterns {
		cleaned = strings.ReplaceAll(cleaned, pattern, "")
	}

	// Remove common markdown artifacts
	cleaned = strings.ReplaceAll(cleaned, "**", "")
	cleaned = strings.ReplaceAll(cleaned, "__", "")

	return strings.TrimSpace(cleaned), nil
}

func (mcbc *MarkdownCodeBlockCleaner) GetPriority() int {
	return 100 // High priority - should run first
}

// WhitespaceCleaner normalizes whitespace and removes unnecessary spacing
type WhitespaceCleaner struct{}

func (wc *WhitespaceCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	// Normalize line endings
	normalized := strings.ReplaceAll(rawResponse, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Remove excessive whitespace but preserve JSON structure
	lines := strings.Split(normalized, "\n")
	cleanedLines := make([]string, 0, len(lines))

	for _, line := range lines {
		// Trim trailing whitespace but preserve leading whitespace for JSON structure
		trimmed := strings.TrimRightFunc(line, unicode.IsSpace)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	// Join back with single newlines
	cleaned := strings.Join(cleanedLines, "\n")
	
	// Remove excessive spaces (but not within JSON strings)
	spaceRegex := regexp.MustCompile(`\s{2,}`)
	cleaned = spaceRegex.ReplaceAllStringFunc(cleaned, func(match string) string {
		// If we're inside quotes, preserve the spacing
		// This is a simple heuristic - more sophisticated parsing could be added
		return " "
	})

	return strings.TrimSpace(cleaned), nil
}

func (wc *WhitespaceCleaner) GetPriority() int {
	return 90 // High priority
}

// InvalidCharacterCleaner removes or replaces invalid characters
type InvalidCharacterCleaner struct{}

func (icc *InvalidCharacterCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	// Remove null bytes and other control characters that can break JSON
	cleaned := strings.ReplaceAll(rawResponse, "\x00", "")
	cleaned = strings.ReplaceAll(cleaned, "\x08", "") // Backspace
	cleaned = strings.ReplaceAll(cleaned, "\x0C", "") // Form feed
	
	// Replace problematic Unicode characters
	cleaned = strings.ReplaceAll(cleaned, "\u0000", "")
	cleaned = strings.ReplaceAll(cleaned, "\uFEFF", "") // BOM
	
	// Fix common JSON escape issues
	cleaned = icc.fixJSONEscaping(cleaned)

	return cleaned, nil
}

func (icc *InvalidCharacterCleaner) fixJSONEscaping(input string) string {
	// Fix unescaped quotes within JSON strings
	result := input
	
	// Basic heuristic to fix common JSON escaping issues
	// This is simplified - a full JSON parser would be more robust
	result = regexp.MustCompile(`"([^"\\]*[^\\])"`).ReplaceAllStringFunc(result, func(match string) string {
		inner := match[1 : len(match)-1] // Remove surrounding quotes
		// Escape any unescaped quotes
		escaped := strings.ReplaceAll(inner, `"`, `\"`)
		return `"` + escaped + `"`
	})

	return result
}

func (icc *InvalidCharacterCleaner) GetPriority() int {
	return 80 // Medium-high priority
}

// JSONStructureCleaner fixes common JSON structure issues
type JSONStructureCleaner struct{}

func (jsc *JSONStructureCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	cleaned := rawResponse

	// Fix trailing commas in JSON objects and arrays
	cleaned = jsc.fixTrailingCommas(cleaned)
	
	// Fix missing commas between JSON properties
	cleaned = jsc.fixMissingCommas(cleaned)
	
	// Fix malformed property names (ensure they're quoted)
	cleaned = jsc.fixUnquotedKeys(cleaned)

	return cleaned, nil
}

func (jsc *JSONStructureCleaner) fixTrailingCommas(input string) string {
	// Remove trailing commas before closing braces and brackets
	trailingCommaRegex := regexp.MustCompile(`,\s*([}\]])`)
	return trailingCommaRegex.ReplaceAllString(input, "$1")
}

func (jsc *JSONStructureCleaner) fixMissingCommas(input string) string {
	// Add missing commas between JSON properties
	// This is a simplified implementation
	missingCommaRegex := regexp.MustCompile(`"([^"]+)"\s*:\s*([^,\}\]]+)\s*"([^"]+)"\s*:`)
	return missingCommaRegex.ReplaceAllString(input, `"$1": $2, "$3":`)
}

func (jsc *JSONStructureCleaner) fixUnquotedKeys(input string) string {
	// Quote unquoted property names
	unquotedKeyRegex := regexp.MustCompile(`([{\s,])\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
	return unquotedKeyRegex.ReplaceAllString(input, `$1"$2":`)
}

func (jsc *JSONStructureCleaner) GetPriority() int {
	return 70 // Medium priority
}

// ResponsePrefixCleaner removes common LLM response prefixes
type ResponsePrefixCleaner struct{}

func (rpc *ResponsePrefixCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	// Common LLM response prefixes to remove
	prefixes := []string{
		"Here's the analysis:",
		"Here is the analysis:",
		"The validation result is:",
		"Here's the validation:",
		"Response:",
		"Result:",
		"Analysis:",
		"Based on the code review:",
		"After analyzing the code:",
		"The security assessment shows:",
	}

	cleaned := rawResponse
	for _, prefix := range prefixes {
		// Remove prefix (case insensitive)
		pattern := `(?i)^\s*` + regexp.QuoteMeta(prefix) + `\s*`
		re := regexp.MustCompile(pattern)
		cleaned = re.ReplaceAllString(cleaned, "")
	}

	// Remove explanatory text before JSON
	jsonStart := strings.Index(cleaned, "{")
	if jsonStart > 0 {
		beforeJSON := cleaned[:jsonStart]
		// If there's explanatory text before JSON, remove it
		if len(beforeJSON) > 50 && strings.Contains(beforeJSON, "JSON") {
			cleaned = cleaned[jsonStart:]
		}
	}

	return strings.TrimSpace(cleaned), nil
}

func (rpc *ResponsePrefixCleaner) GetPriority() int {
	return 95 // Very high priority - should run early
}

// ResponseSuffixCleaner removes common LLM response suffixes
type ResponseSuffixCleaner struct{}

func (rsc *ResponseSuffixCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	// Common LLM response suffixes to remove
	suffixes := []string{
		"This analysis provides a comprehensive assessment.",
		"The validation is complete.",
		"Please let me know if you need more details.",
		"I hope this helps!",
		"Let me know if you have questions.",
		"Is there anything else you'd like me to analyze?",
	}

	cleaned := rawResponse
	
	// Find the last closing brace (end of JSON)
	lastBrace := strings.LastIndex(cleaned, "}")
	if lastBrace == -1 {
		return cleaned, nil
	}

	// Check if there's explanatory text after JSON
	afterJSON := cleaned[lastBrace+1:]
	if len(afterJSON) > 20 {
		// Remove common suffixes
		for _, suffix := range suffixes {
			pattern := `(?i)\s*` + regexp.QuoteMeta(suffix) + `\s*$`
			re := regexp.MustCompile(pattern)
			afterJSON = re.ReplaceAllString(afterJSON, "")
		}
		
		// If most of the suffix was explanatory, remove it
		if len(strings.TrimSpace(afterJSON)) < len(afterJSON)/2 {
			cleaned = cleaned[:lastBrace+1]
		}
	}

	return strings.TrimSpace(cleaned), nil
}

func (rsc *ResponseSuffixCleaner) GetPriority() int {
	return 85 // High priority
}

// MultiJSONCleaner handles responses with multiple JSON objects
type MultiJSONCleaner struct{}

func (mjc *MultiJSONCleaner) Clean(ctx context.Context, rawResponse string) (string, error) {
	// Find all JSON objects in the response
	jsonObjects := mjc.findJSONObjects(rawResponse)
	
	if len(jsonObjects) == 0 {
		return rawResponse, nil
	}
	
	if len(jsonObjects) == 1 {
		return jsonObjects[0], nil
	}

	// Multiple JSON objects found - try to merge or select the best one
	return mjc.selectBestJSON(jsonObjects), nil
}

func (mjc *MultiJSONCleaner) findJSONObjects(input string) []string {
	var objects []string
	
	for i := 0; i < len(input); i++ {
		if input[i] == '{' {
			// Found potential start of JSON object
			braceCount := 0
			start := i
			
			for j := i; j < len(input); j++ {
				switch input[j] {
				case '{':
					braceCount++
				case '}':
					braceCount--
					if braceCount == 0 {
						// Found complete JSON object
						candidate := input[start : j+1]
						if isValidJSON(candidate) {
							objects = append(objects, candidate)
						}
						i = j // Skip to end of this object
						break
					}
				}
			}
		}
	}
	
	return objects
}

func (mjc *MultiJSONCleaner) selectBestJSON(objects []string) string {
	// Select the largest valid JSON object
	best := ""
	maxLength := 0
	
	for _, obj := range objects {
		if len(obj) > maxLength && isValidJSON(obj) {
			best = obj
			maxLength = len(obj)
		}
	}
	
	if best == "" && len(objects) > 0 {
		return objects[0] // Fallback to first object
	}
	
	return best
}

func (mjc *MultiJSONCleaner) GetPriority() int {
	return 60 // Medium priority - run after basic cleaning
}