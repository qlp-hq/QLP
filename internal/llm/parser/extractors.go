package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// JSONExtractor extracts JSON content from mixed text responses
type JSONExtractor struct{}

func (je *JSONExtractor) Extract(ctx context.Context, rawResponse string) (string, error) {
	// Find JSON block in response
	start := strings.Index(rawResponse, "{")
	if start == -1 {
		return "{}", fmt.Errorf("no JSON object found in response")
	}

	// Find matching closing brace
	braceCount := 0
	end := start
	for i := start; i < len(rawResponse); i++ {
		switch rawResponse[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount == 0 {
				end = i
				break
			}
		}
	}

	if braceCount != 0 {
		// Fallback to simple last brace
		end = strings.LastIndex(rawResponse, "}")
		if end == -1 || end <= start {
			return "{}", fmt.Errorf("malformed JSON in response")
		}
	}

	extracted := rawResponse[start : end+1]
	
	// Basic validation
	if !isValidJSON(extracted) {
		return "{}", fmt.Errorf("extracted content is not valid JSON")
	}

	return extracted, nil
}

func (je *JSONExtractor) GetType() ResponseType {
	return ResponseTypeJSON
}

func (je *JSONExtractor) GetPriority() int {
	return 100 // High priority for JSON extraction
}

// ValidationExtractor specifically extracts validation response JSON
type ValidationExtractor struct{}

func (ve *ValidationExtractor) Extract(ctx context.Context, rawResponse string) (string, error) {
	// First try standard JSON extraction
	jsonExtractor := &JSONExtractor{}
	result, err := jsonExtractor.Extract(ctx, rawResponse)
	if err == nil {
		return result, nil
	}

	// Look for validation-specific patterns
	validationPatterns := []string{
		`"overall_score":\s*\d+`,
		`"security_score":\s*\d+`,
		`"quality_score":\s*\d+`,
		`"confidence":\s*\d*\.?\d+`,
	}

	// Check if response contains validation fields
	hasValidationFields := false
	for _, pattern := range validationPatterns {
		if matched, _ := regexp.MatchString(pattern, rawResponse); matched {
			hasValidationFields = true
			break
		}
	}

	if !hasValidationFields {
		return "{}", fmt.Errorf("no validation fields found in response")
	}

	// Try to construct a valid JSON from fragments
	return ve.constructValidationJSON(rawResponse)
}

func (ve *ValidationExtractor) constructValidationJSON(response string) (string, error) {
	// Extract scores using regex
	scoreRegex := regexp.MustCompile(`"(\w+_score)":\s*(\d+)`)
	confidenceRegex := regexp.MustCompile(`"confidence":\s*(\d*\.?\d+)`)
	
	scores := scoreRegex.FindAllStringSubmatch(response, -1)
	confidenceMatch := confidenceRegex.FindStringSubmatch(response)

	if len(scores) == 0 {
		return "{}", fmt.Errorf("no scores found in response")
	}

	// Build JSON
	jsonParts := []string{"{"}
	
	for _, score := range scores {
		jsonParts = append(jsonParts, fmt.Sprintf(`"%s": %s,`, score[1], score[2]))
	}

	if len(confidenceMatch) > 1 {
		jsonParts = append(jsonParts, fmt.Sprintf(`"confidence": %s,`, confidenceMatch[1]))
	} else {
		jsonParts = append(jsonParts, `"confidence": 0.8,`)
	}

	// Remove trailing comma and close JSON
	if len(jsonParts) > 1 {
		lastIdx := len(jsonParts) - 1
		jsonParts[lastIdx] = strings.TrimSuffix(jsonParts[lastIdx], ",")
	}
	jsonParts = append(jsonParts, "}")

	constructed := strings.Join(jsonParts, "")
	
	if !isValidJSON(constructed) {
		return "{}", fmt.Errorf("constructed JSON is invalid")
	}

	return constructed, nil
}

func (ve *ValidationExtractor) GetType() ResponseType {
	return ResponseTypeValidation
}

func (ve *ValidationExtractor) GetPriority() int {
	return 90 // High priority for validation extraction
}

// AnalysisExtractor extracts analysis response data
type AnalysisExtractor struct{}

func (ae *AnalysisExtractor) Extract(ctx context.Context, rawResponse string) (string, error) {
	// First try standard JSON extraction
	jsonExtractor := &JSONExtractor{}
	result, err := jsonExtractor.Extract(ctx, rawResponse)
	if err == nil && ae.isAnalysisResponse(result) {
		return result, nil
	}

	// Look for analysis-specific patterns
	analysisPatterns := []string{
		`"analysis_type":\s*"[^"]+`,
		`"findings":\s*\[`,
		`"recommendations":\s*\[`,
		`"summary":\s*"[^"]+`,
	}

	hasAnalysisFields := false
	for _, pattern := range analysisPatterns {
		if matched, _ := regexp.MatchString(pattern, rawResponse); matched {
			hasAnalysisFields = true
			break
		}
	}

	if !hasAnalysisFields {
		return "{}", fmt.Errorf("no analysis fields found in response")
	}

	// Try to extract analysis from structured text
	return ae.extractFromStructuredText(rawResponse)
}

func (ae *AnalysisExtractor) isAnalysisResponse(jsonStr string) bool {
	analysisFields := []string{
		"analysis_type",
		"findings",
		"recommendations",
		"summary",
	}

	for _, field := range analysisFields {
		if strings.Contains(jsonStr, fmt.Sprintf(`"%s"`, field)) {
			return true
		}
	}
	return false
}

func (ae *AnalysisExtractor) extractFromStructuredText(response string) (string, error) {
	// Extract analysis type
	analysisType := ae.extractField(response, "Analysis Type:", "Type:")
	if analysisType == "" {
		analysisType = "general"
	}

	// Extract summary
	summary := ae.extractField(response, "Summary:", "Conclusion:")
	if summary == "" {
		summary = "Analysis completed"
	}

	// Extract findings (look for bullet points or numbered lists)
	findings := ae.extractList(response, "Findings:", "Issues:", "Problems:")
	
	// Extract recommendations
	recommendations := ae.extractList(response, "Recommendations:", "Suggestions:", "Actions:")

	// Construct JSON
	jsonStr := fmt.Sprintf(`{
		"analysis_type": "%s",
		"summary": "%s",
		"findings": %s,
		"recommendations": %s,
		"confidence": 0.8,
		"timestamp": "%s"
	}`, analysisType, summary, findings, recommendations, "2025-06-11T14:58:40Z")

	if !isValidJSON(jsonStr) {
		return "{}", fmt.Errorf("constructed analysis JSON is invalid")
	}

	return jsonStr, nil
}

func (ae *AnalysisExtractor) extractField(text string, labels ...string) string {
	for _, label := range labels {
		pattern := fmt.Sprintf(`%s\s*(.+?)(?:\n|$)`, regexp.QuoteMeta(label))
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func (ae *AnalysisExtractor) extractList(text string, labels ...string) string {
	items := []string{}
	
	for _, label := range labels {
		// Look for lists after the label
		labelIndex := strings.Index(text, label)
		if labelIndex == -1 {
			continue
		}

		// Extract text after label
		afterLabel := text[labelIndex+len(label):]
		lines := strings.Split(afterLabel, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			
			// Check if it's a list item
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || 
			   strings.HasPrefix(line, "•") || regexp.MustCompile(`^\d+\.`).MatchString(line) {
				
				// Clean the item
				item := regexp.MustCompile(`^[-*•\d\.\s]+`).ReplaceAllString(line, "")
				item = strings.TrimSpace(item)
				if item != "" {
					items = append(items, fmt.Sprintf(`{"description": "%s"}`, item))
				}
			} else if strings.Contains(line, ":") && len(items) < 10 {
				// Stop if we hit a new section
				break
			}
		}
		
		if len(items) > 0 {
			break
		}
	}

	if len(items) == 0 {
		return "[]"
	}

	return "[" + strings.Join(items, ",") + "]"
}

func (ae *AnalysisExtractor) GetType() ResponseType {
	return ResponseTypeAnalysis
}

func (ae *AnalysisExtractor) GetPriority() int {
	return 85 // Medium-high priority
}

// StructuredExtractor handles structured but non-JSON responses
type StructuredExtractor struct{}

func (se *StructuredExtractor) Extract(ctx context.Context, rawResponse string) (string, error) {
	// Convert structured text to JSON format
	lines := strings.Split(rawResponse, "\n")
	jsonData := make(map[string]interface{})
	
	currentSection := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for section headers
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			currentSection = strings.TrimSuffix(line, ":")
			continue
		}

		// Check for key-value pairs
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				if currentSection != "" {
					if jsonData[currentSection] == nil {
						jsonData[currentSection] = make(map[string]string)
					}
					if sectionMap, ok := jsonData[currentSection].(map[string]string); ok {
						sectionMap[key] = value
					}
				} else {
					jsonData[key] = value
				}
			}
		}
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return "{}", fmt.Errorf("failed to convert structured data to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

func (se *StructuredExtractor) GetType() ResponseType {
	return ResponseTypeStructured
}

func (se *StructuredExtractor) GetPriority() int {
	return 70 // Medium priority
}

// TextExtractor handles plain text responses
type TextExtractor struct{}

func (te *TextExtractor) Extract(ctx context.Context, rawResponse string) (string, error) {
	// Wrap plain text in JSON structure
	escaped := strings.ReplaceAll(rawResponse, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	
	jsonStr := fmt.Sprintf(`{
		"content": "%s",
		"type": "text",
		"length": %d,
		"timestamp": "2025-06-11T14:58:40Z"
	}`, escaped, len(rawResponse))

	return jsonStr, nil
}

func (te *TextExtractor) GetType() ResponseType {
	return ResponseTypeText
}

func (te *TextExtractor) GetPriority() int {
	return 50 // Lower priority - fallback option
}