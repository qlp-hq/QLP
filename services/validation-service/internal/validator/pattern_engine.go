package validator

import (
	"regexp"
	"strings"
	"sync"

	"QLP/internal/models"
)

// Pattern represents a regex-based pattern to find in code.
type Pattern struct {
	ID          string
	Regex       string
	Description string
	Severity    string
	Suggestion  string
}

// Match represents a single occurrence of a matched pattern.
type Match struct {
	Pattern    Pattern
	Location   models.Location
	Confidence float64
}

// PatternEngine is responsible for finding pattern matches in code.
type PatternEngine struct {
	patterns      map[string][]Pattern // language -> patterns
	compiledRegex map[string]*regexp.Regexp
	mu            sync.RWMutex
}

// NewPatternEngine creates a new pattern engine with a default set of patterns.
func NewPatternEngine() *PatternEngine {
	engine := &PatternEngine{
		patterns:      make(map[string][]Pattern),
		compiledRegex: make(map[string]*regexp.Regexp),
	}
	engine.initializePatterns()
	return engine
}

// Match finds all pattern matches in the given content for a specific language.
func (pe *PatternEngine) Match(language, content string) []Match {
	var matches []Match
	lines := strings.Split(content, "\n")

	// Get generic patterns and language-specific patterns
	patterns := pe.patterns["generic"]
	if langPatterns, ok := pe.patterns[language]; ok {
		patterns = append(patterns, langPatterns...)
	}

	for _, pattern := range patterns {
		regex := pe.getCompiledRegex(pattern.Regex)
		if regex == nil {
			continue
		}

		for lineNum, line := range lines {
			if regex.MatchString(line) {
				allMatches := regex.FindAllStringIndex(line, -1)
				for _, matchIndices := range allMatches {
					matches = append(matches, Match{
						Pattern: pattern,
						Location: models.Location{
							FilePath: "artifact", // File path is not known here, use a placeholder
							Line:     lineNum + 1,
							Column:   matchIndices[0] + 1,
						},
						Confidence: 0.9, // Simplified confidence
					})
				}
			}
		}
	}
	return matches
}

func (pe *PatternEngine) getCompiledRegex(pattern string) *regexp.Regexp {
	pe.mu.RLock()
	if compiled, exists := pe.compiledRegex[pattern]; exists {
		pe.mu.RUnlock()
		return compiled
	}
	pe.mu.RUnlock()

	pe.mu.Lock()
	defer pe.mu.Unlock()
	// Double-check in case another goroutine compiled it while we waited for the lock
	if compiled, exists := pe.compiledRegex[pattern]; exists {
		return compiled
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// Log this error in a real system
		return nil
	}
	pe.compiledRegex[pattern] = compiled
	return compiled
}

// initializePatterns loads the default set of patterns into the engine.
func (pe *PatternEngine) initializePatterns() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Generic Patterns (apply to all languages)
	pe.patterns["generic"] = []Pattern{
		{
			ID:          "todo_comment",
			Regex:       `(?i)(TODO|FIXME|HACK|XXX)`,
			Description: "A TODO, FIXME, HACK, or XXX comment was found.",
			Severity:    "low",
			Suggestion:  "Resolve the comment or create a work item to track it.",
		},
		{
			ID:          "long_line",
			Regex:       `.{121,}`,
			Description: "Line exceeds 120 characters.",
			Severity:    "low",
			Suggestion:  "Consider breaking the line into multiple lines for readability.",
		},
		{
			ID:          "trailing_whitespace",
			Regex:       `\s+$`,
			Description: "Trailing whitespace detected at the end of a line.",
			Severity:    "low",
			Suggestion:  "Remove trailing whitespace to maintain clean code.",
		},
	}

	// Security Patterns (generic)
	pe.patterns["security"] = []Pattern{
		{
			ID:          "generic_hardcoded_secret",
			Regex:       `(?i)(password|secret|apikey|token|auth_key|access_key|private_key)\s*[=:]\s*['"](.+)['"]`,
			Description: "A hardcoded secret (password, API key, etc.) was found.",
			Severity:    "critical",
			Suggestion:  "Use a secret management system or environment variables instead of hardcoding secrets.",
		},
		{
			ID:          "aws_access_key",
			Regex:       `(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`,
			Description: "Potential AWS Access Key ID found.",
			Severity:    "critical",
			Suggestion:  "Ensure this key is not checked into version control. Use IAM roles where possible.",
		},
	}
	pe.patterns["generic"] = append(pe.patterns["generic"], pe.patterns["security"]...)

	// Go-specific patterns
	pe.patterns["go"] = []Pattern{
		{
			ID:          "go_error_check",
			Regex:       `if err != nil {`,
			Description: "A standard Go error check. (This is for demonstration and is not an issue).",
			Severity:    "info",
			Suggestion:  "N/A",
		},
		{
			ID:          "go_magic_numbers",
			Regex:       `\b(if|case|return|==|!=|>|<|>=|<=)\s+\d+\b`,
			Description: "Magic number detected. A raw number is being used in a comparison or return statement.",
			Severity:    "medium",
			Suggestion:  "Consider defining the number as a constant with a descriptive name.",
		},
	}

	// Python-specific patterns
	pe.patterns["python"] = []Pattern{
		{
			ID:          "python_print_statement",
			Regex:       `\bprint\(`,
			Description: "A `print()` statement was found.",
			Severity:    "low",
			Suggestion:  "Use a structured logger instead of `print()` for application logging.",
		},
	}
}
