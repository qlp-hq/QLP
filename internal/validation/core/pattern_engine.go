package core

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// DefaultPatternEngine implements the PatternEngine interface
type DefaultPatternEngine struct {
	patterns         map[string][]Pattern
	securityPatterns map[string][]SecurityPattern
	qualityPatterns  map[string][]QualityPattern
	compiledRegex    map[string]*regexp.Regexp
	cache            map[string][]Match
	mu               sync.RWMutex
}

// NewDefaultPatternEngine creates a new pattern engine with default patterns
func NewDefaultPatternEngine() *DefaultPatternEngine {
	engine := &DefaultPatternEngine{
		patterns:         make(map[string][]Pattern),
		securityPatterns: make(map[string][]SecurityPattern),
		qualityPatterns:  make(map[string][]QualityPattern),
		compiledRegex:    make(map[string]*regexp.Regexp),
		cache:            make(map[string][]Match),
	}
	
	engine.initializePatterns()
	return engine
}

// MatchPatterns finds all pattern matches in the given content
func (dpe *DefaultPatternEngine) MatchPatterns(content string, patterns []Pattern) []Match {
	// Generate cache key
	cacheKey := fmt.Sprintf("%d_%d", len(content), len(patterns))
	
	dpe.mu.RLock()
	if cached, exists := dpe.cache[cacheKey]; exists {
		dpe.mu.RUnlock()
		return cached
	}
	dpe.mu.RUnlock()

	var matches []Match
	lines := strings.Split(content, "\n")

	for _, pattern := range patterns {
		regex := dpe.getCompiledRegex(pattern.Regex)
		if regex == nil {
			continue
		}

		for lineNum, line := range lines {
			if regex.MatchString(line) {
				// Find all matches in the line
				allMatches := regex.FindAllStringIndex(line, -1)
				for _, matchIndices := range allMatches {
					match := Match{
						Pattern: pattern,
						Location: Location{
							Line:      lineNum + 1,
							Column:    matchIndices[0] + 1,
							StartLine: lineNum + 1,
							EndLine:   lineNum + 1,
						},
						Confidence: dpe.calculateConfidence(pattern, line),
						Context:    dpe.extractContext(lines, lineNum),
					}
					matches = append(matches, match)
				}
			}
		}
	}

	// Cache results
	dpe.mu.Lock()
	dpe.cache[cacheKey] = matches
	dpe.mu.Unlock()

	return matches
}

// GetPatternsForLanguage returns language-specific patterns
func (dpe *DefaultPatternEngine) GetPatternsForLanguage(language string) []Pattern {
	dpe.mu.RLock()
	defer dpe.mu.RUnlock()
	
	if patterns, exists := dpe.patterns[language]; exists {
		return patterns
	}
	
	// Return generic patterns if language-specific not found
	if genericPatterns, exists := dpe.patterns["generic"]; exists {
		return genericPatterns
	}
	
	return []Pattern{}
}

// GetSecurityPatterns returns all security patterns
func (dpe *DefaultPatternEngine) GetSecurityPatterns() []SecurityPattern {
	dpe.mu.RLock()
	defer dpe.mu.RUnlock()
	
	var allPatterns []SecurityPattern
	for _, patterns := range dpe.securityPatterns {
		allPatterns = append(allPatterns, patterns...)
	}
	
	return allPatterns
}

// GetQualityPatterns returns all quality patterns
func (dpe *DefaultPatternEngine) GetQualityPatterns() []QualityPattern {
	dpe.mu.RLock()
	defer dpe.mu.RUnlock()
	
	var allPatterns []QualityPattern
	for _, patterns := range dpe.qualityPatterns {
		allPatterns = append(allPatterns, patterns...)
	}
	
	return allPatterns
}

// GetSecurityPatternsForLanguage returns language-specific security patterns
func (dpe *DefaultPatternEngine) GetSecurityPatternsForLanguage(language string) []SecurityPattern {
	dpe.mu.RLock()
	defer dpe.mu.RUnlock()
	
	if patterns, exists := dpe.securityPatterns[language]; exists {
		return patterns
	}
	
	// Fall back to generic security patterns
	if genericPatterns, exists := dpe.securityPatterns["generic"]; exists {
		return genericPatterns
	}
	
	return []SecurityPattern{}
}

// GetQualityPatternsForLanguage returns language-specific quality patterns
func (dpe *DefaultPatternEngine) GetQualityPatternsForLanguage(language string) []QualityPattern {
	dpe.mu.RLock()
	defer dpe.mu.RUnlock()
	
	if patterns, exists := dpe.qualityPatterns[language]; exists {
		return patterns
	}
	
	// Fall back to generic quality patterns
	if genericPatterns, exists := dpe.qualityPatterns["generic"]; exists {
		return genericPatterns
	}
	
	return []QualityPattern{}
}

// AddPattern adds a new pattern to the engine
func (dpe *DefaultPatternEngine) AddPattern(language string, pattern Pattern) error {
	// Validate regex
	if _, err := regexp.Compile(pattern.Regex); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	dpe.mu.Lock()
	defer dpe.mu.Unlock()
	
	if dpe.patterns[language] == nil {
		dpe.patterns[language] = make([]Pattern, 0)
	}
	
	dpe.patterns[language] = append(dpe.patterns[language], pattern)
	
	// Clear cache as patterns have changed
	dpe.cache = make(map[string][]Match)
	
	return nil
}

// AddSecurityPattern adds a new security pattern
func (dpe *DefaultPatternEngine) AddSecurityPattern(language string, pattern SecurityPattern) error {
	// Validate regex
	if _, err := regexp.Compile(pattern.Regex); err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	
	dpe.mu.Lock()
	defer dpe.mu.Unlock()
	
	if dpe.securityPatterns[language] == nil {
		dpe.securityPatterns[language] = make([]SecurityPattern, 0)
	}
	
	dpe.securityPatterns[language] = append(dpe.securityPatterns[language], pattern)
	
	return nil
}

// Helper methods

func (dpe *DefaultPatternEngine) getCompiledRegex(pattern string) *regexp.Regexp {
	dpe.mu.RLock()
	if compiled, exists := dpe.compiledRegex[pattern]; exists {
		dpe.mu.RUnlock()
		return compiled
	}
	dpe.mu.RUnlock()

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	dpe.mu.Lock()
	dpe.compiledRegex[pattern] = compiled
	dpe.mu.Unlock()

	return compiled
}

func (dpe *DefaultPatternEngine) calculateConfidence(pattern Pattern, line string) float64 {
	baseConfidence := 0.8
	
	// Adjust confidence based on pattern specificity
	regexComplexity := len(pattern.Regex)
	if regexComplexity > 50 {
		baseConfidence += 0.1
	}
	
	// Adjust based on context
	if strings.Contains(line, "TODO") || strings.Contains(line, "FIXME") {
		baseConfidence -= 0.2
	}
	
	if strings.Contains(line, "test") || strings.Contains(line, "Test") {
		baseConfidence -= 0.1
	}
	
	// Ensure confidence is within valid range
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	if baseConfidence < 0.0 {
		baseConfidence = 0.0
	}
	
	return baseConfidence
}

func (dpe *DefaultPatternEngine) extractContext(lines []string, lineNum int) string {
	start := lineNum - 2
	end := lineNum + 2
	
	if start < 0 {
		start = 0
	}
	if end >= len(lines) {
		end = len(lines) - 1
	}
	
	contextLines := lines[start:end+1]
	return strings.Join(contextLines, "\n")
}

func (dpe *DefaultPatternEngine) initializePatterns() {
	// Initialize security patterns
	dpe.initializeSecurityPatterns()
	
	// Initialize quality patterns
	dpe.initializeQualityPatterns()
	
	// Initialize generic patterns
	dpe.initializeGenericPatterns()
}

func (dpe *DefaultPatternEngine) initializeSecurityPatterns() {
	// Generic security patterns (apply to all languages)
	genericSecurityPatterns := []SecurityPattern{
		{
			Pattern: Pattern{
				ID:          "hardcoded_password",
				Type:        PatternTypeSecurity,
				Regex:       `(?i)(password|pwd|pass)\s*[=:]\s*["'][^"']{3,}["']`,
				Description: "Hardcoded password detected",
				Severity:    SeverityHigh,
				Category:    "credential_exposure",
			},
			CWE:   "CWE-798",
			OWASP: "A07:2021 – Identification and Authentication Failures",
			References: []string{
				"https://cwe.mitre.org/data/definitions/798.html",
			},
		},
		{
			Pattern: Pattern{
				ID:          "hardcoded_api_key",
				Type:        PatternTypeSecurity,
				Regex:       `(?i)(api[_-]?key|apikey|access[_-]?token)\s*[=:]\s*["'][A-Za-z0-9+/=]{20,}["']`,
				Description: "Hardcoded API key detected",
				Severity:    SeverityHigh,
				Category:    "credential_exposure",
			},
			CWE:   "CWE-798",
			OWASP: "A07:2021 – Identification and Authentication Failures",
		},
		{
			Pattern: Pattern{
				ID:          "sql_injection_risk",
				Type:        PatternTypeSecurity,
				Regex:       `(?i)(SELECT|INSERT|UPDATE|DELETE).*\+.*["'].*["']`,
				Description: "Potential SQL injection vulnerability",
				Severity:    SeverityCritical,
				Category:    "injection",
			},
			CWE:   "CWE-89",
			OWASP: "A03:2021 – Injection",
		},
		{
			Pattern: Pattern{
				ID:          "command_injection_risk",
				Type:        PatternTypeSecurity,
				Regex:       `(?i)(exec|system|shell_exec|passthru)\s*\([^)]*\$`,
				Description: "Potential command injection vulnerability",
				Severity:    SeverityCritical,
				Category:    "injection",
			},
			CWE:   "CWE-78",
			OWASP: "A03:2021 – Injection",
		},
		{
			Pattern: Pattern{
				ID:          "weak_crypto",
				Type:        PatternTypeSecurity,
				Regex:       `(?i)(md5|sha1|des|rc4)\s*\(`,
				Description: "Weak cryptographic algorithm detected",
				Severity:    SeverityMedium,
				Category:    "cryptography",
			},
			CWE:   "CWE-327",
			OWASP: "A02:2021 – Cryptographic Failures",
		},
	}
	
	// Language-specific security patterns
	goSecurityPatterns := []SecurityPattern{
		{
			Pattern: Pattern{
				ID:          "go_unsafe_pointer",
				Type:        PatternTypeSecurity,
				Regex:       `unsafe\.Pointer`,
				Description: "Use of unsafe.Pointer can lead to memory safety issues",
				Severity:    SeverityMedium,
				Category:    "memory_safety",
			},
			CWE: "CWE-119",
		},
		{
			Pattern: Pattern{
				ID:          "go_sql_injection",
				Type:        PatternTypeSecurity,
				Regex:       `(?i)Query\s*\([^?]*\+`,
				Description: "Potential SQL injection in database query",
				Severity:    SeverityHigh,
				Category:    "injection",
			},
			CWE:   "CWE-89",
			OWASP: "A03:2021 – Injection",
		},
	}
	
	pythonSecurityPatterns := []SecurityPattern{
		{
			Pattern: Pattern{
				ID:          "python_eval_risk",
				Type:        PatternTypeSecurity,
				Regex:       `eval\s*\(`,
				Description: "Use of eval() can lead to code injection",
				Severity:    SeverityCritical,
				Category:    "injection",
			},
			CWE:   "CWE-95",
			OWASP: "A03:2021 – Injection",
		},
		{
			Pattern: Pattern{
				ID:          "python_pickle_risk",
				Type:        PatternTypeSecurity,
				Regex:       `pickle\.loads?\s*\(`,
				Description: "Unsafe deserialization with pickle",
				Severity:    SeverityHigh,
				Category:    "deserialization",
			},
			CWE:   "CWE-502",
			OWASP: "A08:2021 – Software and Data Integrity Failures",
		},
	}
	
	// Store patterns
	dpe.securityPatterns["generic"] = genericSecurityPatterns
	dpe.securityPatterns["go"] = goSecurityPatterns
	dpe.securityPatterns["python"] = pythonSecurityPatterns
}

func (dpe *DefaultPatternEngine) initializeQualityPatterns() {
	// Generic quality patterns
	genericQualityPatterns := []QualityPattern{
		{
			Pattern: Pattern{
				ID:          "long_line",
				Type:        PatternTypeQuality,
				Regex:       `.{120,}`,
				Description: "Line length exceeds recommended limit",
				Severity:    SeverityLow,
				Category:    "readability",
			},
			BestPractice: "Keep lines under 100-120 characters for better readability",
			Suggestion:   "Break long lines into multiple lines",
		},
		{
			Pattern: Pattern{
				ID:          "todo_comment",
				Type:        PatternTypeQuality,
				Regex:       `(?i)(TODO|FIXME|HACK|XXX)`,
				Description: "TODO/FIXME comment found",
				Severity:    SeverityInfo,
				Category:    "technical_debt",
			},
			BestPractice: "Address TODO comments before production",
			Suggestion:   "Create tickets for TODO items and resolve them",
		},
		{
			Pattern: Pattern{
				ID:          "magic_number",
				Type:        PatternTypeQuality,
				Regex:       `[^a-zA-Z_]\d{2,}[^a-zA-Z_\d]`,
				Description: "Magic number detected",
				Severity:    SeverityLow,
				Category:    "maintainability",
			},
			BestPractice: "Use named constants instead of magic numbers",
			Suggestion:   "Define constants with meaningful names",
		},
		{
			Pattern: Pattern{
				ID:          "commented_code",
				Type:        PatternTypeQuality,
				Regex:       `^\s*//.*[{};()].*$`,
				Description: "Commented out code detected",
				Severity:    SeverityInfo,
				Category:    "cleanliness",
			},
			BestPractice: "Remove commented code before committing",
			Suggestion:   "Use version control to track code history instead",
		},
	}
	
	// Language-specific quality patterns
	goQualityPatterns := []QualityPattern{
		{
			Pattern: Pattern{
				ID:          "go_error_ignored",
				Type:        PatternTypeQuality,
				Regex:       `[^,]\s*,\s*_\s*:?=.*\.(.*Error|.*err)\(`,
				Description: "Error return value ignored",
				Severity:    SeverityMedium,
				Category:    "error_handling",
			},
			BestPractice: "Always handle error return values",
			Suggestion:   "Check and handle the error appropriately",
		},
		{
			Pattern: Pattern{
				ID:          "go_fmt_not_used",
				Type:        PatternTypeQuality,
				Regex:       `func.*\{[^}]*fmt\.Print`,
				Description: "Using fmt.Print in function (consider logging)",
				Severity:    SeverityLow,
				Category:    "logging",
			},
			BestPractice: "Use structured logging instead of fmt.Print",
			Suggestion:   "Replace with proper logging framework",
		},
	}
	
	// Store patterns
	dpe.qualityPatterns["generic"] = genericQualityPatterns
	dpe.qualityPatterns["go"] = goQualityPatterns
}

func (dpe *DefaultPatternEngine) initializeGenericPatterns() {
	// Convert security and quality patterns to generic patterns
	genericPatterns := make([]Pattern, 0)
	
	for _, secPatterns := range dpe.securityPatterns {
		for _, secPattern := range secPatterns {
			genericPatterns = append(genericPatterns, secPattern.Pattern)
		}
	}
	
	for _, qualPatterns := range dpe.qualityPatterns {
		for _, qualPattern := range qualPatterns {
			genericPatterns = append(genericPatterns, qualPattern.Pattern)
		}
	}
	
	dpe.patterns["generic"] = genericPatterns
}