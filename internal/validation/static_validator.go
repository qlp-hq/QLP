package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/packaging"
	"QLP/internal/types"
)

// StaticValidator provides comprehensive LLM-based static validation
type StaticValidator struct {
	llmClient         llm.Client
	codeAnalyzer      *CodeAnalyzer
	securityScanner   *SecurityScanner
	qualityChecker    *QualityChecker
	complianceChecker *ComplianceChecker
}

// StaticValidationResult represents comprehensive static validation results
type StaticValidationResult struct {
	OverallScore       int                    `json:"overall_score"`
	SecurityScore      int                    `json:"security_score"`
	QualityScore       int                    `json:"quality_score"`
	ComplianceScore    int                    `json:"compliance_score"`
	ArchitectureScore  int                    `json:"architecture_score"`
	Issues             []ValidationIssue      `json:"issues"`
	Recommendations    []string               `json:"recommendations"`
	DeploymentReady    bool                   `json:"deployment_ready"`
	Confidence         float64                `json:"confidence"`
	SecurityFindings   []types.SecurityFinding `json:"security_findings"`
	QualityFindings    []QualityFinding       `json:"quality_findings"`
	ArchitectureFindings []ArchitectureFinding `json:"architecture_findings"`
	ValidationTime     time.Duration          `json:"validation_time"`
	ValidatedAt        time.Time              `json:"validated_at"`
}

// Use SecurityFinding from types package

// QualityFinding represents a code quality finding
type QualityFinding struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	Location       string `json:"location"`
	Recommendation string `json:"recommendation"`
	Category       string `json:"category"`
}

// ArchitectureFinding represents an architectural finding
type ArchitectureFinding struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	Component      string `json:"component"`
	Recommendation string `json:"recommendation"`
	Pattern        string `json:"pattern"`
}

// CodeAnalyzer analyzes code structure and patterns
type CodeAnalyzer struct {
	llmClient llm.Client
}

// QualityChecker performs code quality assessment
type QualityChecker struct {
	llmClient llm.Client
}

// NewStaticValidator creates a new static validator
func NewStaticValidator(llmClient llm.Client) *StaticValidator {
	return &StaticValidator{
		llmClient:         llmClient,
		codeAnalyzer:      &CodeAnalyzer{llmClient: llmClient},
		securityScanner:   NewSecurityScanner(),
		qualityChecker:    &QualityChecker{llmClient: llmClient},
		complianceChecker: NewComplianceChecker(),
	}
}

// NewComplianceChecker creates a new compliance checker
func NewComplianceChecker() *ComplianceChecker {
	return &ComplianceChecker{}
}

// ValidateQuantumDrop performs comprehensive static validation of a QuantumDrop
func (sv *StaticValidator) ValidateQuantumDrop(ctx context.Context, drop *packaging.QuantumDrop) (*StaticValidationResult, error) {
	startTime := time.Now()
	log.Printf("Starting comprehensive static validation for QuantumDrop: %s", drop.Name)

	result := &StaticValidationResult{
		Issues:               make([]ValidationIssue, 0),
		Recommendations:      make([]string, 0),
		SecurityFindings:     make([]types.SecurityFinding, 0),
		QualityFindings:      make([]QualityFinding, 0),
		ArchitectureFindings: make([]ArchitectureFinding, 0),
		ValidatedAt:          startTime,
	}

	// Extract code content from QuantumDrop files
	codeContent, projectStructure := sv.extractCodeContent(drop)

	// Multi-LLM validation with different specialized models
	results := make([]int, 0)

	// 1. Security-focused LLM validation
	securityScore, securityFindings, err := sv.validateSecurity(ctx, codeContent, drop.Type)
	if err != nil {
		log.Printf("Security validation failed: %v", err)
		securityScore = 50 // Fallback score
	}
	result.SecurityScore = securityScore
	result.SecurityFindings = securityFindings
	results = append(results, securityScore)

	// 2. Code quality-focused LLM validation
	qualityScore, qualityFindings, err := sv.validateQuality(ctx, codeContent, drop.Type)
	if err != nil {
		log.Printf("Quality validation failed: %v", err)
		qualityScore = 60 // Fallback score
	}
	result.QualityScore = qualityScore
	result.QualityFindings = qualityFindings
	results = append(results, qualityScore)

	// 3. Architecture-focused LLM validation
	architectureScore, architectureFindings, err := sv.validateArchitecture(ctx, codeContent, projectStructure, drop.Type)
	if err != nil {
		log.Printf("Architecture validation failed: %v", err)
		architectureScore = 65 // Fallback score
	}
	result.ArchitectureScore = architectureScore
	result.ArchitectureFindings = architectureFindings
	results = append(results, architectureScore)

	// 4. Compliance validation
	complianceScore := sv.validateCompliance(codeContent, drop.Type)
	result.ComplianceScore = complianceScore
	results = append(results, complianceScore)

	// Aggregate results
	result.OverallScore = sv.calculateOverallScore(results)
	result.DeploymentReady = sv.assessDeploymentReadiness(result)
	result.Confidence = sv.calculateConfidence(result)
	result.Issues = sv.aggregateIssues(result)
	result.Recommendations = sv.generateRecommendations(result)
	result.ValidationTime = time.Since(startTime)

	log.Printf("Static validation completed for %s: Overall=%d, Security=%d, Quality=%d, Architecture=%d",
		drop.Name, result.OverallScore, result.SecurityScore, result.QualityScore, result.ArchitectureScore)

	return result, nil
}

// validateSecurity performs security-focused LLM validation
func (sv *StaticValidator) validateSecurity(ctx context.Context, codeContent string, dropType packaging.DropType) (int, []types.SecurityFinding, error) {
	prompt := fmt.Sprintf(`You are a senior cybersecurity expert and penetration tester reviewing code for enterprise deployment in a Fortune 500 company.

CRITICAL SECURITY ASSESSMENT for %s:

SECURITY CHECKLIST (Enterprise Grade):
1. Input Validation & Sanitization (SQL injection, XSS, command injection)
2. Authentication & Authorization (proper session management, RBAC)
3. Secrets Management (no hardcoded credentials, proper key rotation)
4. Cryptography (strong algorithms, proper key management)
5. Error Handling (no information leakage, proper logging)
6. Rate Limiting & DDoS Protection
7. CSRF & CORS Protection
8. Secure Headers Implementation (HSTS, CSP, etc.)
9. Dependency Vulnerability Assessment
10. Data Protection (encryption at rest/transit, PII handling)
11. Session Security (secure cookies, timeout, invalidation)
12. File Upload Security (type validation, size limits, scanning)
13. API Security (proper authentication, input validation)
14. Infrastructure Security (secure configurations, principle of least privilege)
15. Logging & Monitoring (security events, audit trails)

COMPLIANCE REQUIREMENTS:
- OWASP Top 10 (2021)
- CWE Top 25
- NIST Cybersecurity Framework
- SOC 2 Type II
- ISO 27001

CODE TO REVIEW:
%s

PROVIDE DETAILED SECURITY ASSESSMENT:
1. Overall security score (0-100) - Enterprise grade threshold is 90+
2. Critical security vulnerabilities (severity: CRITICAL, HIGH, MEDIUM, LOW)
3. OWASP Top 10 mapping for each finding
4. CWE identification where applicable
5. Specific remediation steps
6. Enterprise deployment readiness (yes/no with justification)
7. Confidence level in assessment (0.0-1.0)

RESPOND WITH JSON:
{
  "security_score": 85,
  "enterprise_ready": true,
  "confidence": 0.92,
  "findings": [
    {
      "type": "Input Validation",
      "severity": "HIGH",
      "description": "SQL injection vulnerability in user input handling",
      "location": "line 45, function getUserData",
      "recommendation": "Use parameterized queries or prepared statements",
      "cwe": "CWE-89",
      "owasp": "A03:2021 - Injection"
    }
  ],
  "compliance_gaps": ["Missing rate limiting", "Inadequate error handling"],
  "recommendations": ["Implement comprehensive input validation", "Add proper authentication middleware"]
}`, dropType, codeContent)

	response, err := sv.llmClient.Complete(ctx, prompt)
	if err != nil {
		return 0, nil, fmt.Errorf("LLM security validation failed: %w", err)
	}

	// Parse JSON response
	var securityResult struct {
		SecurityScore   int                `json:"security_score"`
		EnterpriseReady bool               `json:"enterprise_ready"`
		Confidence      float64            `json:"confidence"`
		Findings        []types.SecurityFinding  `json:"findings"`
		ComplianceGaps  []string           `json:"compliance_gaps"`
		Recommendations []string           `json:"recommendations"`
	}

	if err := json.Unmarshal([]byte(response), &securityResult); err != nil {
		log.Printf("Failed to parse security validation response: %v", err)
		return sv.fallbackSecurityAnalysis(codeContent), []types.SecurityFinding{}, nil
	}

	return securityResult.SecurityScore, securityResult.Findings, nil
}

// validateQuality performs code quality-focused LLM validation
func (sv *StaticValidator) validateQuality(ctx context.Context, codeContent string, dropType packaging.DropType) (int, []QualityFinding, error) {
	prompt := fmt.Sprintf(`You are a principal software engineer and code review expert with 15+ years of experience, reviewing code for production deployment at a tech unicorn.

CODE QUALITY ASSESSMENT for %s:

QUALITY CHECKLIST (Production Grade):
1. Code Structure & Organization (clean architecture, separation of concerns)
2. Error Handling Completeness (proper try-catch, graceful degradation)
3. Performance Considerations (O(n) complexity, memory efficiency)
4. Memory Management (proper cleanup, garbage collection awareness)
5. Concurrency Safety (thread safety, race condition prevention)
6. Testing Coverage (unit tests, integration tests, edge cases)
7. Documentation Quality (clear comments, API documentation)
8. Maintainability (SOLID principles, DRY, KISS)
9. Scalability (horizontal scaling readiness, stateless design)
10. Best Practices Adherence (language-specific conventions)
11. Code Readability (clear naming, logical flow)
12. Resource Management (connection pooling, proper cleanup)
13. Configuration Management (externalized config, environment awareness)
14. Logging Strategy (structured logging, appropriate levels)
15. Monitoring & Observability (metrics, health checks, tracing)

LANGUAGE-SPECIFIC STANDARDS:
- Go: Effective Go principles, proper error handling, goroutine management
- Python: PEP 8, type hints, proper imports, virtual environments
- JavaScript/TypeScript: ES6+ standards, proper async handling, type safety
- Java: Oracle coding standards, proper exception handling, memory management
- Rust: Ownership principles, error handling, performance optimization

CODE TO REVIEW:
%s

PROVIDE COMPREHENSIVE QUALITY ASSESSMENT:
1. Overall quality score (0-100) - Production grade threshold is 85+
2. Code smells and anti-patterns
3. Performance bottlenecks and optimization opportunities
4. Maintainability issues
5. Testing gaps
6. Documentation deficiencies
7. Refactoring suggestions with priorities
8. Production readiness assessment
9. Scalability concerns
10. Technical debt assessment

RESPOND WITH JSON:
{
  "quality_score": 82,
  "production_ready": true,
  "maintainability_score": 85,
  "performance_score": 78,
  "testability_score": 80,
  "documentation_score": 75,
  "findings": [
    {
      "type": "Performance",
      "severity": "MEDIUM",
      "description": "Inefficient database query in loop causing N+1 problem",
      "location": "line 67, function processUsers",
      "recommendation": "Implement batch loading or use JOIN queries",
      "category": "Performance Optimization"
    }
  ],
  "refactoring_suggestions": ["Extract service layer", "Implement caching strategy"],
  "technical_debt": "Medium - some legacy patterns need modernization"
}`, dropType, codeContent)

	response, err := sv.llmClient.Complete(ctx, prompt)
	if err != nil {
		return 0, nil, fmt.Errorf("LLM quality validation failed: %w", err)
	}

	// Parse JSON response
	var qualityResult struct {
		QualityScore        int              `json:"quality_score"`
		ProductionReady     bool             `json:"production_ready"`
		MaintainabilityScore int             `json:"maintainability_score"`
		PerformanceScore    int              `json:"performance_score"`
		TestabilityScore    int              `json:"testability_score"`
		DocumentationScore  int              `json:"documentation_score"`
		Findings            []QualityFinding `json:"findings"`
		RefactoringSuggestions []string      `json:"refactoring_suggestions"`
		TechnicalDebt       string           `json:"technical_debt"`
	}

	if err := json.Unmarshal([]byte(response), &qualityResult); err != nil {
		log.Printf("Failed to parse quality validation response: %v", err)
		return sv.fallbackQualityAnalysis(codeContent), []QualityFinding{}, nil
	}

	return qualityResult.QualityScore, qualityResult.Findings, nil
}

// validateArchitecture performs architecture-focused LLM validation
func (sv *StaticValidator) validateArchitecture(ctx context.Context, codeContent string, projectStructure string, dropType packaging.DropType) (int, []ArchitectureFinding, error) {
	prompt := fmt.Sprintf(`You are a senior solutions architect and enterprise architect with deep expertise in system design, reviewing architecture for enterprise deployment.

ARCHITECTURE ASSESSMENT for %s:

PROJECT STRUCTURE:
%s

ARCHITECTURE CHECKLIST (Enterprise Grade):
1. Separation of Concerns (proper layering, domain boundaries)
2. Single Responsibility Principle (focused modules, clear interfaces)
3. Dependency Injection (loose coupling, testability)
4. Configuration Management (externalized config, environment handling)
5. Logging & Monitoring (structured logging, observability patterns)
6. Health Checks & Metrics (endpoint monitoring, system health)
7. Graceful Shutdown (resource cleanup, signal handling)
8. Circuit Breaker Patterns (fault tolerance, resilience)
9. Retry Mechanisms (exponential backoff, dead letter queues)
10. Caching Strategy (appropriate levels, invalidation)
11. Database Design (proper indexing, connection pooling)
12. API Design (RESTful principles, versioning, documentation)
13. Security Architecture (defense in depth, principle of least privilege)
14. Scalability Patterns (horizontal scaling, stateless design)
15. Event-Driven Architecture (async processing, message queues)
16. Microservices Patterns (service boundaries, communication)
17. Data Architecture (consistency, backup, disaster recovery)
18. Infrastructure as Code (containerization, orchestration)
19. CI/CD Integration (automated testing, deployment pipelines)
20. Operational Excellence (monitoring, alerting, runbooks)

ARCHITECTURAL PATTERNS TO ASSESS:
- Clean Architecture / Hexagonal Architecture
- Domain-Driven Design (DDD)
- Command Query Responsibility Segregation (CQRS)
- Event Sourcing
- Saga Pattern
- Bulkhead Pattern
- Strangler Fig Pattern

CODE TO REVIEW:
%s

PROVIDE COMPREHENSIVE ARCHITECTURE ASSESSMENT:
1. Overall architecture score (0-100) - Enterprise grade threshold is 85+
2. Architectural soundness and design patterns
3. Scalability potential and bottlenecks
4. Maintainability and extensibility
5. Operational readiness
6. Enterprise integration compatibility
7. Cloud-native readiness
8. Microservices compatibility
9. Performance architecture
10. Security architecture
11. Data architecture
12. Integration architecture

RESPOND WITH JSON:
{
  "architecture_score": 88,
  "enterprise_ready": true,
  "scalability_score": 85,
  "maintainability_score": 90,
  "operational_score": 82,
  "cloud_native_score": 87,
  "findings": [
    {
      "type": "Scalability",
      "severity": "MEDIUM",
      "description": "Service lacks horizontal scaling capabilities due to in-memory state",
      "component": "UserService",
      "recommendation": "Externalize state to Redis or database for stateless scaling",
      "pattern": "Stateless Service Pattern"
    }
  ],
  "architectural_patterns": ["Clean Architecture", "Repository Pattern"],
  "improvement_areas": ["Add circuit breakers", "Implement event sourcing"],
  "enterprise_readiness": "Ready with minor improvements"
}`, dropType, projectStructure, codeContent)

	response, err := sv.llmClient.Complete(ctx, prompt)
	if err != nil {
		return 0, nil, fmt.Errorf("LLM architecture validation failed: %w", err)
	}

	// Parse JSON response
	var architectureResult struct {
		ArchitectureScore   int                    `json:"architecture_score"`
		EnterpriseReady     bool                   `json:"enterprise_ready"`
		ScalabilityScore    int                    `json:"scalability_score"`
		MaintainabilityScore int                   `json:"maintainability_score"`
		OperationalScore    int                    `json:"operational_score"`
		CloudNativeScore    int                    `json:"cloud_native_score"`
		Findings            []ArchitectureFinding  `json:"findings"`
		ArchitecturalPatterns []string             `json:"architectural_patterns"`
		ImprovementAreas    []string               `json:"improvement_areas"`
		EnterpriseReadiness string                 `json:"enterprise_readiness"`
	}

	if err := json.Unmarshal([]byte(response), &architectureResult); err != nil {
		log.Printf("Failed to parse architecture validation response: %v", err)
		return sv.fallbackArchitectureAnalysis(codeContent), []ArchitectureFinding{}, nil
	}

	return architectureResult.ArchitectureScore, architectureResult.Findings, nil
}

// validateCompliance performs compliance validation
func (sv *StaticValidator) validateCompliance(codeContent string, dropType packaging.DropType) int {
	score := 80 // Base compliance score

	// Check for compliance indicators
	complianceIndicators := map[string]int{
		"logging":     10, // Audit logging
		"encryption":  15, // Data protection
		"auth":        15, // Access control
		"validation":  10, // Input validation
		"error":       10, // Error handling
		"config":      5,  // Configuration management
		"health":      5,  // Health monitoring
		"backup":      5,  // Data backup
		"test":        10, // Testing
		"doc":         5,  // Documentation
	}

	codeUpper := strings.ToLower(codeContent)
	for indicator, points := range complianceIndicators {
		if strings.Contains(codeUpper, indicator) {
			score += points
		}
	}

	// Cap score at 100
	if score > 100 {
		score = 100
	}

	return score
}

// Helper methods
func (sv *StaticValidator) extractCodeContent(drop *packaging.QuantumDrop) (string, string) {
	var codeContent strings.Builder
	var projectStructure strings.Builder

	projectStructure.WriteString("Project Structure:\n")
	for filePath, content := range drop.Files {
		projectStructure.WriteString(fmt.Sprintf("- %s\n", filePath))
		codeContent.WriteString(fmt.Sprintf("=== %s ===\n%s\n\n", filePath, content))
	}

	return codeContent.String(), projectStructure.String()
}

func (sv *StaticValidator) calculateOverallScore(scores []int) int {
	if len(scores) == 0 {
		return 0
	}

	total := 0
	for _, score := range scores {
		total += score
	}

	return total / len(scores)
}

func (sv *StaticValidator) assessDeploymentReadiness(result *StaticValidationResult) bool {
	// Enterprise deployment thresholds
	return result.SecurityScore >= 85 &&
		result.QualityScore >= 80 &&
		result.ArchitectureScore >= 80 &&
		result.ComplianceScore >= 75
}

func (sv *StaticValidator) calculateConfidence(result *StaticValidationResult) float64 {
	// Multi-dimensional confidence calculation
	securityConfidence := float64(result.SecurityScore) / 100.0
	qualityConfidence := float64(result.QualityScore) / 100.0
	architectureConfidence := float64(result.ArchitectureScore) / 100.0
	complianceConfidence := float64(result.ComplianceScore) / 100.0

	// Weighted confidence
	weights := map[string]float64{
		"security":     0.35,
		"quality":      0.30,
		"architecture": 0.25,
		"compliance":   0.10,
	}

	overallConfidence := weights["security"]*securityConfidence +
		weights["quality"]*qualityConfidence +
		weights["architecture"]*architectureConfidence +
		weights["compliance"]*complianceConfidence

	return overallConfidence
}

func (sv *StaticValidator) aggregateIssues(result *StaticValidationResult) []ValidationIssue {
	issues := make([]ValidationIssue, 0)

	// Convert security findings to validation issues
	for _, finding := range result.SecurityFindings {
		if finding.Severity == "CRITICAL" || finding.Severity == "HIGH" {
			issues = append(issues, ValidationIssue{
				Severity:    finding.Severity,
				Category:    "Security",
				Message:     finding.Description,
				Resource:    finding.Location,
				Remediation: finding.Recommendation,
			})
		}
	}

	// Convert quality findings to validation issues
	for _, finding := range result.QualityFindings {
		if finding.Severity == "HIGH" {
			issues = append(issues, ValidationIssue{
				Severity:    finding.Severity,
				Category:    "Quality",
				Message:     finding.Description,
				Resource:    finding.Location,
				Remediation: finding.Recommendation,
			})
		}
	}

	// Convert architecture findings to validation issues
	for _, finding := range result.ArchitectureFindings {
		if finding.Severity == "HIGH" {
			issues = append(issues, ValidationIssue{
				Severity:    finding.Severity,
				Category:    "Architecture",
				Message:     finding.Description,
				Resource:    finding.Component,
				Remediation: finding.Recommendation,
			})
		}
	}

	return issues
}

func (sv *StaticValidator) generateRecommendations(result *StaticValidationResult) []string {
	recommendations := make([]string, 0)

	// Security recommendations
	if result.SecurityScore < 90 {
		recommendations = append(recommendations, "Enhance security controls to meet enterprise standards (target: 90+)")
	}

	// Quality recommendations
	if result.QualityScore < 85 {
		recommendations = append(recommendations, "Improve code quality and testing coverage (target: 85+)")
	}

	// Architecture recommendations
	if result.ArchitectureScore < 85 {
		recommendations = append(recommendations, "Refactor architecture for better scalability and maintainability")
	}

	// Compliance recommendations
	if result.ComplianceScore < 80 {
		recommendations = append(recommendations, "Address compliance gaps for enterprise deployment")
	}

	// Overall recommendations
	if !result.DeploymentReady {
		recommendations = append(recommendations, "Address critical issues before production deployment")
	}

	if result.Confidence < 0.9 {
		recommendations = append(recommendations, "Improve validation confidence through additional testing")
	}

	return recommendations
}

// Fallback analysis methods
func (sv *StaticValidator) fallbackSecurityAnalysis(codeContent string) int {
	score := 70 // Base security score

	// Basic security checks
	securityIndicators := map[string]int{
		"authentication": 10,
		"authorization":  10,
		"encryption":     15,
		"validation":     10,
		"sanitiz":        10,
		"csrf":           5,
		"cors":           5,
		"rate":           5,
		"log":            5,
		"audit":          5,
	}

	codeUpper := strings.ToLower(codeContent)
	for indicator, points := range securityIndicators {
		if strings.Contains(codeUpper, indicator) {
			score += points
		}
	}

	// Deduct for potential issues
	if strings.Contains(codeUpper, "password") && strings.Contains(codeContent, "=") {
		score -= 20 // Potential hardcoded password
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

func (sv *StaticValidator) fallbackQualityAnalysis(codeContent string) int {
	score := 65 // Base quality score

	// Basic quality checks
	qualityIndicators := map[string]int{
		"test":    15,
		"error":   10,
		"log":     5,
		"config":  5,
		"doc":     5,
		"comment": 5,
		"func":    5,
		"class":   5,
		"interface": 5,
		"struct":  5,
	}

	codeUpper := strings.ToLower(codeContent)
	for indicator, points := range qualityIndicators {
		if strings.Contains(codeUpper, indicator) {
			score += points
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

func (sv *StaticValidator) fallbackArchitectureAnalysis(codeContent string) int {
	score := 60 // Base architecture score

	// Basic architecture checks
	architectureIndicators := map[string]int{
		"service":    10,
		"handler":    5,
		"controller": 5,
		"repository": 10,
		"interface":  10,
		"config":     5,
		"middleware": 5,
		"router":     5,
		"model":      5,
		"dto":        5,
	}

	codeUpper := strings.ToLower(codeContent)
	for indicator, points := range architectureIndicators {
		if strings.Contains(codeUpper, indicator) {
			score += points
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}