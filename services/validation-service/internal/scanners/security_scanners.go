package scanners

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"QLP/services/validation-service/pkg/contracts"
)

type SecurityScanner interface {
	Scan(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.SecurityResult, error)
	GetSupportedLanguages() []string
}

type SecurityScannerRegistry struct {
	scanners map[string]SecurityScanner
}

func NewSecurityScannerRegistry() *SecurityScannerRegistry {
	registry := &SecurityScannerRegistry{
		scanners: make(map[string]SecurityScanner),
	}
	
	// Register built-in scanners
	registry.Register("fast", NewFastSecurityScanner())
	registry.Register("standard", NewStandardSecurityScanner())
	registry.Register("comprehensive", NewComprehensiveSecurityScanner())
	
	return registry
}

func (r *SecurityScannerRegistry) Register(name string, scanner SecurityScanner) {
	r.scanners[strings.ToLower(name)] = scanner
}

func (r *SecurityScannerRegistry) GetScanner(name string) SecurityScanner {
	return r.scanners[strings.ToLower(name)]
}

// Fast Security Scanner
type FastSecurityScanner struct{}

func NewFastSecurityScanner() *FastSecurityScanner {
	return &FastSecurityScanner{}
}

func (s *FastSecurityScanner) Scan(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.SecurityResult, error) {
	result := &contracts.SecurityResult{
		Score:           85, // Start with good baseline
		RiskLevel:       contracts.SecurityRiskLevelLow,
		Vulnerabilities: []contracts.SecurityIssue{},
		Warnings:        []contracts.SecurityIssue{},
		Passed:          true,
		ScannedBy:       []string{"fast-security-scanner"},
	}
	
	contentLower := strings.ToLower(content)
	lines := strings.Split(content, "\n")
	
	// Quick security pattern checks
	s.scanForSecrets(content, lines, result)
	s.scanForInjectionPatterns(content, lines, result)
	s.scanForInsecurePractices(contentLower, lines, result)
	s.scanForCryptographyIssues(contentLower, lines, result)
	s.scanLanguageSpecificIssues(content, language, lines, result)
	
	// Calculate final risk level
	s.calculateRiskLevel(result)
	
	result.Passed = result.Score >= 70
	return result, nil
}

func (s *FastSecurityScanner) GetSupportedLanguages() []string {
	return []string{"go", "python", "javascript", "typescript", "java", "c", "cpp", "php", "ruby", "shell", "bash"}
}

func (s *FastSecurityScanner) scanForSecrets(content string, lines []string, result *contracts.SecurityResult) {
	// Common secret patterns
	secretPatterns := map[string]string{
		"api_key":           `(?i)(api[_-]?key|apikey)\s*[:=]\s*["\']?[a-zA-Z0-9]{16,}["\']?`,
		"secret_key":        `(?i)(secret[_-]?key|secretkey)\s*[:=]\s*["\']?[a-zA-Z0-9]{16,}["\']?`,
		"private_key":       `-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`,
		"aws_access_key":    `AKIA[0-9A-Z]{16}`,
		"jwt_token":         `ey[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`,
		"password_in_url":   `(?i)(https?://[^:/\s]+):([^@/\s]+)@`,
		"github_token":      `ghp_[a-zA-Z0-9]{36}`,
		"slack_token":       `xox[baprs]-([0-9a-zA-Z]{10,48})`,
	}
	
	for vulnType, pattern := range secretPatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)
		
		for _, match := range matches {
			// Find line number
			lineNum := s.findLineNumber(content, match[0])
			
			severity := contracts.SeverityCritical
			score_penalty := 25
			
			if vulnType == "password_in_url" || vulnType == "jwt_token" {
				severity = contracts.SeverityError
				score_penalty = 15
			}
			
			result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
				ID:          generateIssueID(vulnType, lineNum),
				Type:        "secret-exposure",
				Severity:    severity,
				Title:       "Hardcoded Secret Detected",
				Description: fmt.Sprintf("Potential %s found in code", strings.ReplaceAll(vulnType, "_", " ")),
				Line:        lineNum,
				CWE:         "CWE-798",
				CVSS:        8.5,
				References:  []string{"https://cwe.mitre.org/data/definitions/798.html"},
				Remediation: "Remove hardcoded secrets and use environment variables or secure vault",
				Metadata: map[string]string{
					"pattern_type": vulnType,
					"confidence":   "high",
				},
			})
			
			result.Score -= score_penalty
		}
	}
}

func (s *FastSecurityScanner) scanForInjectionPatterns(content string, lines []string, result *contracts.SecurityResult) {
	// SQL Injection patterns
	sqlPatterns := []string{
		`(?i)select\s+.*\s+from\s+.*\s+where\s+.*=\s*['"]?\s*\+`,
		`(?i)insert\s+into\s+.*\s+values\s*\(\s*['"]?\s*\+`,
		`(?i)update\s+.*\s+set\s+.*=\s*['"]?\s*\+`,
		`(?i)delete\s+from\s+.*\s+where\s+.*=\s*['"]?\s*\+`,
		`(?i)query\s*\(\s*['"].*\+.*['"]`,
	}
	
	for _, pattern := range sqlPatterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindAllString(content, -1); len(matches) > 0 {
			for _, match := range matches {
				lineNum := s.findLineNumber(content, match)
				
				result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
					ID:          generateIssueID("sql_injection", lineNum),
					Type:        "injection",
					Severity:    contracts.SeverityCritical,
					Title:       "Potential SQL Injection",
					Description: "SQL query construction using string concatenation",
					Line:        lineNum,
					CWE:         "CWE-89",
					CVSS:        9.3,
					References:  []string{"https://cwe.mitre.org/data/definitions/89.html"},
					Remediation: "Use parameterized queries or prepared statements",
					Metadata: map[string]string{
						"injection_type": "sql",
						"confidence":     "medium",
					},
				})
				
				result.Score -= 20
			}
		}
	}
	
	// Command Injection patterns
	cmdPatterns := []string{
		`(?i)exec\s*\(\s*['"].*\+.*['"]`,
		`(?i)system\s*\(\s*['"].*\+.*['"]`,
		`(?i)shell_exec\s*\(\s*['"].*\+.*['"]`,
		`(?i)subprocess\.call\s*\(\s*['"].*\+.*['"]`,
	}
	
	for _, pattern := range cmdPatterns {
		regex := regexp.MustCompile(pattern)
		if matches := regex.FindAllString(content, -1); len(matches) > 0 {
			for _, match := range matches {
				lineNum := s.findLineNumber(content, match)
				
				result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
					ID:          generateIssueID("cmd_injection", lineNum),
					Type:        "injection",
					Severity:    contracts.SeverityCritical,
					Title:       "Potential Command Injection",
					Description: "Command execution with user input concatenation",
					Line:        lineNum,
					CWE:         "CWE-78",
					CVSS:        8.8,
					References:  []string{"https://cwe.mitre.org/data/definitions/78.html"},
					Remediation: "Validate and sanitize input, use safe APIs",
					Metadata: map[string]string{
						"injection_type": "command",
						"confidence":     "medium",
					},
				})
				
				result.Score -= 20
			}
		}
	}
}

func (s *FastSecurityScanner) scanForInsecurePractices(contentLower string, lines []string, result *contracts.SecurityResult) {
	// Insecure practices
	insecurePractices := map[string]struct {
		pattern     string
		title       string
		description string
		cwe         string
		cvss        float64
		penalty     int
		severity    contracts.Severity
	}{
		"hardcoded_password": {
			pattern:     `(?i)password\s*[:=]\s*["\'][^"\']{3,}["\']`,
			title:       "Hardcoded Password",
			description: "Password appears to be hardcoded in source code",
			cwe:         "CWE-259",
			cvss:        7.5,
			penalty:     20,
			severity:    contracts.SeverityCritical,
		},
		"weak_crypto": {
			pattern:     `(?i)(md5|sha1)\s*\(`,
			title:       "Weak Cryptographic Algorithm",
			description: "Use of weak cryptographic hash functions",
			cwe:         "CWE-327",
			cvss:        5.9,
			penalty:     10,
			severity:    contracts.SeverityWarning,
		},
		"insecure_random": {
			pattern:     `(?i)(math\.random|random\.random)\s*\(`,
			title:       "Insecure Random Number Generation",
			description: "Use of predictable random number generator",
			cwe:         "CWE-338",
			cvss:        5.3,
			penalty:     8,
			severity:    contracts.SeverityWarning,
		},
		"debug_enabled": {
			pattern:     `(?i)(debug\s*[:=]\s*true|debug\s*=\s*1)`,
			title:       "Debug Mode Enabled",
			description: "Debug mode should not be enabled in production",
			cwe:         "CWE-489",
			cvss:        4.3,
			penalty:     5,
			severity:    contracts.SeverityWarning,
		},
	}
	
	for practiceType, practice := range insecurePractices {
		regex := regexp.MustCompile(practice.pattern)
		matches := regex.FindAllString(contentLower, -1)
		
		for _, match := range matches {
			lineNum := s.findLineNumber(contentLower, match)
			
			result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
				ID:          generateIssueID(practiceType, lineNum),
				Type:        "insecure-practice",
				Severity:    practice.severity,
				Title:       practice.title,
				Description: practice.description,
				Line:        lineNum,
				CWE:         practice.cwe,
				CVSS:        practice.cvss,
				Remediation: getRemediationForPractice(practiceType),
				Metadata: map[string]string{
					"practice_type": practiceType,
					"confidence":    "medium",
				},
			})
			
			result.Score -= practice.penalty
		}
	}
}

func (s *FastSecurityScanner) scanForCryptographyIssues(contentLower string, lines []string, result *contracts.SecurityResult) {
	// Cryptography-related issues
	cryptoIssues := map[string]struct {
		pattern     string
		title       string
		description string
		penalty     int
		severity    contracts.Severity
	}{
		"weak_cipher": {
			pattern:     `(?i)(des|3des|rc4|blowfish)\s*[(\[]`,
			title:       "Weak Encryption Algorithm",
			description: "Use of weak or deprecated encryption algorithm",
			penalty:     15,
			severity:    contracts.SeverityError,
		},
		"no_ssl_verify": {
			pattern:     `(?i)(ssl[_-]?verify|verify[_-]?ssl)\s*[:=]\s*false`,
			title:       "SSL Verification Disabled",
			description: "SSL certificate verification is disabled",
			penalty:     20,
			severity:    contracts.SeverityCritical,
		},
		"weak_key_size": {
			pattern:     `(?i)(rsa|key[_-]?size)\s*[:=]\s*(512|1024)`,
			title:       "Weak Key Size",
			description: "Cryptographic key size is too small",
			penalty:     10,
			severity:    contracts.SeverityWarning,
		},
	}
	
	for issueType, issue := range cryptoIssues {
		regex := regexp.MustCompile(issue.pattern)
		matches := regex.FindAllString(contentLower, -1)
		
		for _, match := range matches {
			lineNum := s.findLineNumber(contentLower, match)
			
			result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
				ID:          generateIssueID(issueType, lineNum),
				Type:        "cryptography",
				Severity:    issue.severity,
				Title:       issue.title,
				Description: issue.description,
				Line:        lineNum,
				CWE:         "CWE-327",
				CVSS:        6.5,
				Remediation: getCryptoRemediation(issueType),
				Metadata: map[string]string{
					"crypto_issue": issueType,
					"confidence":   "high",
				},
			})
			
			result.Score -= issue.penalty
		}
	}
}

func (s *FastSecurityScanner) scanLanguageSpecificIssues(content, language string, lines []string, result *contracts.SecurityResult) {
	switch strings.ToLower(language) {
	case "go":
		s.scanGoSecurityIssues(content, lines, result)
	case "python":
		s.scanPythonSecurityIssues(content, lines, result)
	case "javascript", "typescript":
		s.scanJavaScriptSecurityIssues(content, lines, result)
	case "java":
		s.scanJavaSecurityIssues(content, lines, result)
	}
}

func (s *FastSecurityScanner) scanGoSecurityIssues(content string, lines []string, result *contracts.SecurityResult) {
	// Go-specific security issues
	goIssues := []struct {
		pattern     string
		title       string
		description string
		penalty     int
	}{
		{
			pattern:     `(?i)sql\.Open\s*\(\s*[^,]+,\s*[^)]*\+`,
			title:       "SQL Query Concatenation",
			description: "Database query built with string concatenation",
			penalty:     15,
		},
		{
			pattern:     `(?i)exec\.Command\s*\([^)]*\+`,
			title:       "Command Injection Risk",
			description: "Command execution with concatenated arguments",
			penalty:     18,
		},
		{
			pattern:     `(?i)template\.HTML\s*\(`,
			title:       "Unescaped HTML Template",
			description: "HTML template without proper escaping",
			penalty:     12,
		},
	}
	
	for _, issue := range goIssues {
		regex := regexp.MustCompile(issue.pattern)
		matches := regex.FindAllString(content, -1)
		
		for _, match := range matches {
			lineNum := s.findLineNumber(content, match)
			
			result.Warnings = append(result.Warnings, contracts.SecurityIssue{
				ID:          generateIssueID("go_specific", lineNum),
				Type:        "language-specific",
				Severity:    contracts.SeverityWarning,
				Title:       issue.title,
				Description: issue.description,
				Line:        lineNum,
				Remediation: "Use parameterized queries and input validation",
				Metadata: map[string]string{
					"language":   "go",
					"confidence": "medium",
				},
			})
			
			result.Score -= issue.penalty
		}
	}
}

func (s *FastSecurityScanner) scanPythonSecurityIssues(content string, lines []string, result *contracts.SecurityResult) {
	// Python-specific security issues
	pythonIssues := []struct {
		pattern     string
		title       string
		description string
		penalty     int
	}{
		{
			pattern:     `(?i)eval\s*\(`,
			title:       "Dangerous Use of eval()",
			description: "eval() can execute arbitrary code",
			penalty:     25,
		},
		{
			pattern:     `(?i)exec\s*\(`,
			title:       "Dangerous Use of exec()",
			description: "exec() can execute arbitrary code",
			penalty:     25,
		},
		{
			pattern:     `(?i)pickle\.loads?\s*\(`,
			title:       "Unsafe Deserialization",
			description: "pickle.load can execute arbitrary code",
			penalty:     20,
		},
		{
			pattern:     `(?i)subprocess\.shell\s*=\s*True`,
			title:       "Shell Injection Risk",
			description: "subprocess with shell=True is dangerous",
			penalty:     15,
		},
	}
	
	for _, issue := range pythonIssues {
		regex := regexp.MustCompile(issue.pattern)
		matches := regex.FindAllString(content, -1)
		
		for _, match := range matches {
			lineNum := s.findLineNumber(content, match)
			
			result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
				ID:          generateIssueID("python_specific", lineNum),
				Type:        "language-specific",
				Severity:    contracts.SeverityCritical,
				Title:       issue.title,
				Description: issue.description,
				Line:        lineNum,
				CWE:         "CWE-94",
				CVSS:        9.8,
				Remediation: "Avoid dynamic code execution, use safer alternatives",
				Metadata: map[string]string{
					"language":   "python",
					"confidence": "high",
				},
			})
			
			result.Score -= issue.penalty
		}
	}
}

func (s *FastSecurityScanner) scanJavaScriptSecurityIssues(content string, lines []string, result *contracts.SecurityResult) {
	// JavaScript-specific security issues
	jsIssues := []struct {
		pattern     string
		title       string
		description string
		penalty     int
	}{
		{
			pattern:     `(?i)eval\s*\(`,
			title:       "Dangerous Use of eval()",
			description: "eval() can execute arbitrary JavaScript",
			penalty:     25,
		},
		{
			pattern:     `(?i)innerHTML\s*=.*\+`,
			title:       "XSS Risk with innerHTML",
			description: "innerHTML with concatenated content can lead to XSS",
			penalty:     18,
		},
		{
			pattern:     `(?i)document\.write\s*\(.*\+`,
			title:       "XSS Risk with document.write",
			description: "document.write with concatenated content can lead to XSS",
			penalty:     18,
		},
		{
			pattern:     `(?i)setTimeout\s*\(\s*["\'][^"\']*\+`,
			title:       "Code Injection in setTimeout",
			description: "setTimeout with string concatenation can execute arbitrary code",
			penalty:     20,
		},
	}
	
	for _, issue := range jsIssues {
		regex := regexp.MustCompile(issue.pattern)
		matches := regex.FindAllString(content, -1)
		
		for _, match := range matches {
			lineNum := s.findLineNumber(content, match)
			
			result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
				ID:          generateIssueID("js_specific", lineNum),
				Type:        "language-specific",
				Severity:    contracts.SeverityError,
				Title:       issue.title,
				Description: issue.description,
				Line:        lineNum,
				CWE:         "CWE-79",
				CVSS:        7.5,
				Remediation: "Use safe DOM manipulation methods and input validation",
				Metadata: map[string]string{
					"language":   "javascript",
					"confidence": "medium",
				},
			})
			
			result.Score -= issue.penalty
		}
	}
}

func (s *FastSecurityScanner) scanJavaSecurityIssues(content string, lines []string, result *contracts.SecurityResult) {
	// Java-specific security issues
	javaIssues := []struct {
		pattern     string
		title       string
		description string
		penalty     int
	}{
		{
			pattern:     `(?i)Runtime\.getRuntime\(\)\.exec\s*\(`,
			title:       "Command Execution Risk",
			description: "Runtime.exec can lead to command injection",
			penalty:     20,
		},
		{
			pattern:     `(?i)ObjectInputStream\s*\(`,
			title:       "Unsafe Deserialization",
			description: "ObjectInputStream can lead to remote code execution",
			penalty:     25,
		},
		{
			pattern:     `(?i)Class\.forName\s*\(.*\+`,
			title:       "Dynamic Class Loading",
			description: "Dynamic class loading with user input is dangerous",
			penalty:     18,
		},
	}
	
	for _, issue := range javaIssues {
		regex := regexp.MustCompile(issue.pattern)
		matches := regex.FindAllString(content, -1)
		
		for _, match := range matches {
			lineNum := s.findLineNumber(content, match)
			
			result.Vulnerabilities = append(result.Vulnerabilities, contracts.SecurityIssue{
				ID:          generateIssueID("java_specific", lineNum),
				Type:        "language-specific",
				Severity:    contracts.SeverityCritical,
				Title:       issue.title,
				Description: issue.description,
				Line:        lineNum,
				CWE:         "CWE-94",
				CVSS:        9.8,
				Remediation: "Validate input and use safer alternatives",
				Metadata: map[string]string{
					"language":   "java",
					"confidence": "high",
				},
			})
			
			result.Score -= issue.penalty
		}
	}
}

func (s *FastSecurityScanner) calculateRiskLevel(result *contracts.SecurityResult) {
	criticalCount := 0
	errorCount := 0
	
	for _, vuln := range result.Vulnerabilities {
		switch vuln.Severity {
		case contracts.SeverityCritical:
			criticalCount++
		case contracts.SeverityError:
			errorCount++
		}
	}
	
	if criticalCount > 0 {
		result.RiskLevel = contracts.SecurityRiskLevelCritical
	} else if errorCount > 2 {
		result.RiskLevel = contracts.SecurityRiskLevelHigh
	} else if errorCount > 0 || result.Score < 60 {
		result.RiskLevel = contracts.SecurityRiskLevelMedium
	} else if result.Score < 80 {
		result.RiskLevel = contracts.SecurityRiskLevelLow
	} else {
		result.RiskLevel = contracts.SecurityRiskLevelNone
	}
}

func (s *FastSecurityScanner) findLineNumber(content, pattern string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return i + 1
		}
	}
	return 1
}

// Helper functions
func generateIssueID(issueType string, lineNum int) string {
	return fmt.Sprintf("%s_%d", issueType, lineNum)
}

func getRemediationForPractice(practiceType string) string {
	remediations := map[string]string{
		"hardcoded_password":  "Use environment variables or secure configuration management",
		"weak_crypto":         "Use stronger hash functions like SHA-256 or bcrypt",
		"insecure_random":     "Use cryptographically secure random number generators",
		"debug_enabled":       "Disable debug mode in production environments",
	}
	
	if remediation, exists := remediations[practiceType]; exists {
		return remediation
	}
	return "Follow security best practices"
}

func getCryptoRemediation(issueType string) string {
	remediations := map[string]string{
		"weak_cipher":    "Use AES with proper key sizes (128, 192, or 256 bits)",
		"no_ssl_verify":  "Enable SSL certificate verification",
		"weak_key_size":  "Use RSA keys of at least 2048 bits or equivalent",
	}
	
	if remediation, exists := remediations[issueType]; exists {
		return remediation
	}
	return "Use modern cryptographic standards"
}

// Standard Security Scanner (placeholder for more comprehensive scanning)
type StandardSecurityScanner struct {
	fastScanner *FastSecurityScanner
}

func NewStandardSecurityScanner() *StandardSecurityScanner {
	return &StandardSecurityScanner{
		fastScanner: NewFastSecurityScanner(),
	}
}

func (s *StandardSecurityScanner) Scan(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.SecurityResult, error) {
	// Start with fast scan
	result, err := s.fastScanner.Scan(ctx, content, language, taskType)
	if err != nil {
		return nil, err
	}
	
	// Add standard-level checks
	// TODO: Implement more sophisticated scanning
	
	return result, nil
}

func (s *StandardSecurityScanner) GetSupportedLanguages() []string {
	return s.fastScanner.GetSupportedLanguages()
}

// Comprehensive Security Scanner (placeholder for advanced scanning)
type ComprehensiveSecurityScanner struct {
	standardScanner *StandardSecurityScanner
}

func NewComprehensiveSecurityScanner() *ComprehensiveSecurityScanner {
	return &ComprehensiveSecurityScanner{
		standardScanner: NewStandardSecurityScanner(),
	}
}

func (s *ComprehensiveSecurityScanner) Scan(ctx context.Context, content, language string, taskType contracts.TaskType) (*contracts.SecurityResult, error) {
	// Start with standard scan
	result, err := s.standardScanner.Scan(ctx, content, language, taskType)
	if err != nil {
		return nil, err
	}
	
	// Add comprehensive-level checks
	// TODO: Implement advanced scanning (SAST integration, CVE database lookup, etc.)
	
	return result, nil
}

func (s *ComprehensiveSecurityScanner) GetSupportedLanguages() []string {
	return s.standardScanner.GetSupportedLanguages()
}