package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"QLP/internal/models"
	"QLP/internal/sandbox"
	"QLP/internal/types"
)

type SecurityScanner struct {
	patterns      map[types.SecurityRiskLevel][]SecurityPattern
	cveDatabase   *CVEDatabase
	complianceChecker *ComplianceChecker
}

type SecurityPattern struct {
	Name        string
	Pattern     *regexp.Regexp
	Description string
	Mitigation  string
	Category    string
}

type CVEDatabase struct {
	vulnerabilities map[string]CVEInfo
}

type CVEInfo struct {
	ID          string
	Score       float64
	Description string
	Affected    []string
}

type ComplianceChecker struct {
	standards map[string]ComplianceStandard
}

type ComplianceStandard struct {
	Name  string
	Rules []ComplianceRule
}

type ComplianceRule struct {
	ID          string
	Description string
	Check       func(string) bool
	Severity    types.SecurityRiskLevel
}

func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		patterns:          initializeSecurityPatterns(),
		cveDatabase:       initializeCVEDatabase(),
		complianceChecker: initializeComplianceChecker(),
	}
}

func (ss *SecurityScanner) ScanOutput(ctx context.Context, task models.Task, output string, sandboxResult *sandbox.SandboxExecutionResult) (*types.SecurityResult, error) {
	result := &types.SecurityResult{
		Score:             100,
		RiskLevel:         types.SecurityRiskLevelNone,
		Vulnerabilities:   []types.SecurityIssue{},
		SandboxViolations: []string{},
		Passed:            true,
	}

	// Add compliance score as a separate field
	_ = 100 // complianceScore - TODO: implement compliance scoring

	// 1. Pattern-based vulnerability scanning
	vulnerabilities := ss.scanForVulnerabilities(output)
	result.Vulnerabilities = append(result.Vulnerabilities, vulnerabilities...)

	// 2. CVE database lookup
	cveIssues := ss.checkCVEDatabase(output)
	result.Vulnerabilities = append(result.Vulnerabilities, cveIssues...)

	// 3. Compliance checking
	_ = ss.checkCompliance(task, output) // TODO: use compliance score

	// 4. Sandbox violation analysis
	if sandboxResult != nil {
		violations := ss.analyzeSandboxViolations(sandboxResult)
		result.SandboxViolations = violations
	}

	// Calculate overall security score
	result.Score = ss.calculateSecurityScore(result)
	result.RiskLevel = ss.determineRiskLevel(result.Score, result.Vulnerabilities)
	result.Passed = result.Score >= 70

	return result, nil
}

func (ss *SecurityScanner) scanForVulnerabilities(output string) []types.SecurityIssue {
	var issues []types.SecurityIssue

	for riskLevel, patterns := range ss.patterns {
		for _, pattern := range patterns {
			matches := pattern.Pattern.FindAllStringSubmatch(output, -1)
			for _, match := range matches {
				issue := types.SecurityIssue{
					Type:        pattern.Category,
					Severity:    string(riskLevel),
					Description: fmt.Sprintf("%s: %s", pattern.Name, pattern.Description),
					Location:    ss.findLocation(output, match[0]),
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

func (ss *SecurityScanner) checkCVEDatabase(output string) []types.SecurityIssue {
	var issues []types.SecurityIssue

	// Check for known vulnerable patterns or dependencies
	for cveID, cveInfo := range ss.cveDatabase.vulnerabilities {
		for _, affected := range cveInfo.Affected {
			if strings.Contains(strings.ToLower(output), strings.ToLower(affected)) {
				severity := types.SecurityRiskMediumStr
				if cveInfo.Score >= 7.0 {
					severity = types.SecurityRiskHighStr
				}
				if cveInfo.Score >= 9.0 {
					severity = types.SecurityRiskCriticalStr
				}

				issue := types.SecurityIssue{
					Type:        "CVE",
					Severity:    severity,
					Description: fmt.Sprintf("CVE %s: %s", cveID, cveInfo.Description),
					Location:    ss.findLocation(output, affected),
				}
				issues = append(issues, issue)
			}
		}
	}

	return issues
}

func (ss *SecurityScanner) checkCompliance(task models.Task, output string) int {
	score := 100
	
	standards := []string{"OWASP", "CIS", "NIST"}
	
	for _, standardName := range standards {
		if standard, exists := ss.complianceChecker.standards[standardName]; exists {
			violations := 0
			
			for _, rule := range standard.Rules {
				if !rule.Check(output) {
					violations++
					switch rule.Severity {
					case types.SecurityRiskLevelCritical:
						score -= 20
					case types.SecurityRiskLevelHigh:
						score -= 10
					case types.SecurityRiskLevelMedium:
						score -= 5
					case types.SecurityRiskLevelLow:
						score -= 2
					}
				}
			}
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (ss *SecurityScanner) analyzeSandboxViolations(sandboxResult *sandbox.SandboxExecutionResult) []string {
	var violations []string

	// Check if sandbox execution was successful
	if !sandboxResult.Success {
		violations = append(violations, "Sandbox execution failed")
	}

	// Check security score from sandbox
	if sandboxResult.SecurityScore < 80 {
		violations = append(violations, fmt.Sprintf("Low sandbox security score: %d", sandboxResult.SecurityScore))
	}

	// Analyze execution results for violations
	for _, result := range sandboxResult.Results {
		if result.ExitCode != 0 {
			violations = append(violations, fmt.Sprintf("Command failed with exit code %d: %s", result.ExitCode, result.Command))
		}

		// Check for suspicious stderr output
		if ss.containsSuspiciousOutput(result.Stderr) {
			violations = append(violations, fmt.Sprintf("Suspicious stderr output in command: %s", result.Command))
		}
	}

	return violations
}

func (ss *SecurityScanner) containsSuspiciousOutput(output string) bool {
	suspiciousPatterns := []string{
		"permission denied",
		"access denied", 
		"unauthorized",
		"segmentation fault",
		"buffer overflow",
		"stack smashing",
	}

	lowercaseOutput := strings.ToLower(output)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowercaseOutput, pattern) {
			return true
		}
	}

	return false
}

func (ss *SecurityScanner) calculateSecurityScore(result *types.SecurityResult) int {
	score := 100

	// Deduct points for vulnerabilities
	for _, vuln := range result.Vulnerabilities {
		switch vuln.Severity {
		case types.SecurityRiskCriticalStr:
			score -= 30
		case types.SecurityRiskHighStr:
			score -= 20
		case types.SecurityRiskMediumStr:
			score -= 10
		case types.SecurityRiskLowStr:
			score -= 5
		}
	}

	// Factor in compliance score (use local variable)
	// score = (score + complianceScore) / 2

	// Deduct for sandbox violations
	score -= len(result.SandboxViolations) * 5

	if score < 0 {
		score = 0
	}

	return score
}

func (ss *SecurityScanner) determineRiskLevel(score int, vulnerabilities []types.SecurityIssue) types.SecurityRiskLevel {
	// Check for critical vulnerabilities first
	for _, vuln := range vulnerabilities {
		if vuln.Severity == types.SecurityRiskCriticalStr {
			return types.SecurityRiskLevelCritical
		}
	}

	// Determine by score
	switch {
	case score >= 90:
		return types.SecurityRiskLevelNone
	case score >= 70:
		return types.SecurityRiskLevelLow
	case score >= 50:
		return types.SecurityRiskLevelMedium
	case score >= 30:
		return types.SecurityRiskLevelHigh
	default:
		return types.SecurityRiskLevelCritical
	}
}

func (ss *SecurityScanner) findLocation(text, pattern string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return fmt.Sprintf("Line %d", i+1)
		}
	}
	return "Unknown"
}

func initializeSecurityPatterns() map[types.SecurityRiskLevel][]SecurityPattern {
	return map[types.SecurityRiskLevel][]SecurityPattern{
		types.SecurityRiskLevelCritical: {
			{
				Name:        "SQL Injection",
				Pattern:     regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE).*\+.*\$`),
				Description: "Potential SQL injection vulnerability",
				Mitigation:  "Use parameterized queries or prepared statements",
				Category:    "Injection",
			},
			{
				Name:        "Command Injection",
				Pattern:     regexp.MustCompile(`(?i)(exec|system|eval|os\.system)\s*\(.*\+`),
				Description: "Potential command injection vulnerability",
				Mitigation:  "Sanitize input and use safe execution methods",
				Category:    "Injection",
			},
			{
				Name:        "Hardcoded Secrets",
				Pattern:     regexp.MustCompile(`(?i)(password|secret|key|token)\s*=\s*["'][^"']+["']`),
				Description: "Hardcoded credentials detected",
				Mitigation:  "Use environment variables or secure credential storage",
				Category:    "Secrets",
			},
		},
		types.SecurityRiskLevelHigh: {
			{
				Name:        "Unsafe Deserialization",
				Pattern:     regexp.MustCompile(`(?i)(pickle|yaml|json)\.loads?\s*\(`),
				Description: "Potentially unsafe deserialization",
				Mitigation:  "Validate and sanitize input before deserialization",
				Category:    "Deserialization",
			},
			{
				Name:        "Path Traversal",
				Pattern:     regexp.MustCompile(`\.\./`),
				Description: "Potential path traversal vulnerability",
				Mitigation:  "Validate and sanitize file paths",
				Category:    "FileSystem",
			},
		},
		types.SecurityRiskLevelMedium: {
			{
				Name:        "Weak Cryptography",
				Pattern:     regexp.MustCompile(`(?i)(md5|sha1|des|rc4)`),
				Description: "Use of weak cryptographic algorithms",
				Mitigation:  "Use strong cryptographic algorithms (AES, SHA-256+)",
				Category:    "Cryptography",
			},
			{
				Name:        "Insecure Random",
				Pattern:     regexp.MustCompile(`(?i)math\.random|Random\(\)`),
				Description: "Use of insecure random number generation",
				Mitigation:  "Use cryptographically secure random number generators",
				Category:    "Random",
			},
		},
		types.SecurityRiskLevelLow: {
			{
				Name:        "TODO Security",
				Pattern:     regexp.MustCompile(`(?i)TODO.*security`),
				Description: "Security-related TODO comments found",
				Mitigation:  "Complete security implementations",
				Category:    "Documentation",
			},
		},
	}
}

func initializeCVEDatabase() *CVEDatabase {
	return &CVEDatabase{
		vulnerabilities: map[string]CVEInfo{
			"CVE-2023-44487": {
				ID:          "CVE-2023-44487",
				Score:       7.5,
				Description: "HTTP/2 Rapid Reset vulnerability",
				Affected:    []string{"http2", "grpc"},
			},
			"CVE-2023-39325": {
				ID:          "CVE-2023-39325",
				Score:       7.5,
				Description: "golang.org/x/net/http2 vulnerability",
				Affected:    []string{"golang.org/x/net/http2"},
			},
		},
	}
}

func initializeComplianceChecker() *ComplianceChecker {
	return &ComplianceChecker{
		standards: map[string]ComplianceStandard{
			"OWASP": {
				Name: "OWASP Top 10",
				Rules: []ComplianceRule{
					{
						ID:          "OWASP-A01",
						Description: "Broken Access Control",
						Check: func(output string) bool {
							return !strings.Contains(strings.ToLower(output), "admin") || 
								   strings.Contains(strings.ToLower(output), "authorization")
						},
						Severity: types.SecurityRiskLevelHigh,
					},
					{
						ID:          "OWASP-A02", 
						Description: "Cryptographic Failures",
						Check: func(output string) bool {
							return !regexp.MustCompile(`(?i)(md5|sha1|des)`).MatchString(output)
						},
						Severity: types.SecurityRiskLevelMedium,
					},
				},
			},
			"CIS": {
				Name: "CIS Controls",
				Rules: []ComplianceRule{
					{
						ID:          "CIS-3.1",
						Description: "Secure Configuration",
						Check: func(output string) bool {
							return !strings.Contains(strings.ToLower(output), "default_password")
						},
						Severity: types.SecurityRiskLevelHigh,
					},
				},
			},
		},
	}
}