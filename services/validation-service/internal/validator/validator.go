package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/models"
)

const (
	maxRefinementCycles   = 3
	passingScoreThreshold = 95
)

// Validator is responsible for validating artifacts.
type Validator struct {
	llmClient     llm.Client
	patternEngine *PatternEngine
}

// New creates a new Validator.
func New(llmClient llm.Client) *Validator {
	return &Validator{
		llmClient:     llmClient,
		patternEngine: NewPatternEngine(),
	}
}

// Validate performs a validation check on a given artifact, including an iterative refinement loop.
func (v *Validator) Validate(ctx context.Context, artifact *models.Artifact) *models.ValidationResult {
	currentContent := artifact.Content
	var lastResult *models.ValidationResult

	for i := 0; i < maxRefinementCycles; i++ {
		result := v.runValidationCycle(ctx, artifact, currentContent)
		lastResult = result

		if result.Passed {
			break // Exit the loop if validation passes
		}

		if i < maxRefinementCycles-1 {
			// If not the last cycle, attempt to refine the code
			refinementPrompt, err := v.generateRefinementPrompt(result)
			if err != nil {
				// Cannot generate a fix, so no point continuing.
				break
			}

			fixedCode, err := v.refineCode(ctx, currentContent, refinementPrompt)
			if err != nil {
				// LLM failed to fix the code, so we stop.
				break
			}
			currentContent = fixedCode          // Use the refined code for the next cycle
			result.Artifact.Content = fixedCode // Update the artifact content in the result
		}
	}

	return lastResult
}

func (v *Validator) runValidationCycle(ctx context.Context, artifact *models.Artifact, content string) *models.ValidationResult {
	result := &models.ValidationResult{
		Artifact:         *artifact,
		ValidatedAt:      time.Now(),
		ComponentScores:  make(map[string]int),
		Issues:           []models.Issue{},
		SecurityFindings: []models.SecurityFinding{},
	}
	result.Artifact.Content = content // Ensure the result has the content that was validated

	matches := v.patternEngine.Match(artifact.Task.Language, content)
	for _, match := range matches {
		issue := models.Issue{
			ID:          match.Pattern.ID,
			Title:       match.Pattern.Description,
			Description: match.Pattern.Description,
			Severity:    match.Pattern.Severity,
			Suggestion:  match.Pattern.Suggestion,
			Location:    &match.Location,
		}
		result.Issues = append(result.Issues, issue)
	}

	// Calculate scores based on findings
	score := 100
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "critical":
			score -= 50
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 2
		}
	}
	if score < 0 {
		score = 0
	}

	result.OverallScore = score
	result.ComponentScores["static_analysis"] = score
	result.Passed = result.OverallScore >= passingScoreThreshold

	return result
}

func (v *Validator) generateRefinementPrompt(result *models.ValidationResult) (string, error) {
	if len(result.Issues) == 0 {
		return "", fmt.Errorf("no issues found to generate refinement prompt")
	}

	var promptBuilder strings.Builder
	promptBuilder.WriteString("The following code has been reviewed and found to have issues. Please fix the code to address the following points.\n")
	promptBuilder.WriteString("Provide only the complete, corrected code. Do not provide any commentary, explanations, or markdown formatting.\n\n")
	promptBuilder.WriteString("ISSUES TO FIX:\n")

	for _, issue := range result.Issues {
		promptBuilder.WriteString(fmt.Sprintf("- **File:** %s, Line: %d\n", issue.Location.FilePath, issue.Location.Line))
		promptBuilder.WriteString(fmt.Sprintf("  - **Issue:** %s (Severity: %s)\n", issue.Title, issue.Severity))
		promptBuilder.WriteString(fmt.Sprintf("  - **Suggestion:** %s\n\n", issue.Suggestion))
	}

	return promptBuilder.String(), nil
}

func (v *Validator) refineCode(ctx context.Context, originalCode string, prompt string) (string, error) {
	fullPrompt := fmt.Sprintf("%s\n--- ORIGINAL CODE ---\n%s", prompt, originalCode)

	// Make the LLM call to get the fixed code
	fixedCode, err := v.llmClient.Complete(ctx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("LLM refinement call failed: %w", err)
	}

	// The response from the LLM is expected to be just the code.
	// We might need to strip markdown ``` wrappers.
	fixedCode = strings.TrimSpace(fixedCode)
	if strings.HasPrefix(fixedCode, "```") {
		lines := strings.SplitN(fixedCode, "\n", 2)
		if len(lines) > 1 {
			fixedCode = lines[1]
		}
		fixedCode = strings.TrimSuffix(fixedCode, "```")
	}

	return fixedCode, nil
}
