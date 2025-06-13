package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"QLP/internal/llm"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// InfrastructureValidator provides comprehensive validation for infrastructure code
type InfrastructureValidator struct {
	llmClient llm.Client
}

// InfraValidationResult represents comprehensive infrastructure validation results
type InfraValidationResult struct {
	OverallScore        int                       `json:"overall_score"`
	TerraformResult     *TerraformValidationResult `json:"terraform_result,omitempty"`
	KubernetesResult    *KubernetesValidationResult `json:"kubernetes_result,omitempty"`
	SecurityResult      *SecurityValidationResult  `json:"security_result"`
	CostEstimation      *CostEstimation           `json:"cost_estimation"`
	ComplianceResult    *ComplianceValidationResult `json:"compliance_result"`
	DeploymentRisk      RiskLevel                 `json:"deployment_risk"`
	ValidationPassed    bool                      `json:"validation_passed"`
	CriticalIssues      []ValidationIssue         `json:"critical_issues"`
	Recommendations     []string                  `json:"recommendations"`
	EstimatedDeployTime time.Duration             `json:"estimated_deploy_time"`
	ValidatedAt         time.Time                 `json:"validated_at"`
}

// TerraformValidationResult contains Terraform-specific validation results
type TerraformValidationResult struct {
	SyntaxValid         bool                 `json:"syntax_valid"`
	PlanValid           bool                 `json:"plan_valid"`
	SecurityScore       int                  `json:"security_score"`
	BestPracticeScore   int                  `json:"best_practice_score"`
	ResourceCount       int                  `json:"resource_count"`
	EstimatedCost       float64              `json:"estimated_cost"`
	SecurityIssues      []SecurityIssue      `json:"security_issues"`
	PolicyViolations    []PolicyViolation    `json:"policy_violations"`
	ResourceEfficiency  int                  `json:"resource_efficiency"`
	Optimizations       []CostOptimization   `json:"optimizations"`
}

// KubernetesValidationResult contains Kubernetes-specific validation results
type KubernetesValidationResult struct {
	ManifestsValid      bool                `json:"manifests_valid"`
	APIVersionsValid    bool                `json:"api_versions_valid"`
	ResourceLimitsSet   bool                `json:"resource_limits_set"`
	SecurityContextSet  bool                `json:"security_context_set"`
	ProductionReadiness int                 `json:"production_readiness"`
	ScalabilityScore    int                 `json:"scalability_score"`
	SecurityScore       int                 `json:"security_score"`
	PolicyCompliance    int                 `json:"policy_compliance"`
	HealthChecksSet     bool                `json:"health_checks_set"`
	NetworkPoliciesSet  bool                `json:"network_policies_set"`
	Issues              []KubernetesIssue   `json:"issues"`
	Recommendations     []string            `json:"recommendations"`
}

// SecurityValidationResult contains infrastructure security analysis
type SecurityValidationResult struct {
	CISCompliance       int                 `json:"cis_compliance"`
	SecurityPosture     int                 `json:"security_posture"`
	VulnerabilityCount  int                 `json:"vulnerability_count"`
	CriticalFindings    []SecurityFinding   `json:"critical_findings"`
	EncryptionEnabled   bool                `json:"encryption_enabled"`
	AccessControlValid  bool                `json:"access_control_valid"`
	NetworkSecuritySet  bool                `json:"network_security_set"`
	AuditLoggingEnabled bool                `json:"audit_logging_enabled"`
	SecurityRecommendations []string        `json:"security_recommendations"`
}

// CostEstimation provides detailed cost analysis for infrastructure
type CostEstimation struct {
	MonthlyCost         float64              `json:"monthly_cost"`
	YearlyCost          float64              `json:"yearly_cost"`
	ResourceBreakdown   map[string]float64   `json:"resource_breakdown"`
	CostOptimizations   []CostOptimization   `json:"cost_optimizations"`
	CostRisk            RiskLevel            `json:"cost_risk"`
	CostEfficiencyScore int                  `json:"cost_efficiency_score"`
	BudgetRecommendation string              `json:"budget_recommendation"`
}

// ComplianceValidationResult contains enterprise compliance validation
type ComplianceValidationResult struct {
	SOC2Compliance      int                 `json:"soc2_compliance"`
	GDPRCompliance      int                 `json:"gdpr_compliance"`
	HIPAACompliance     int                 `json:"hipaa_compliance"`
	PolicyCompliance    int                 `json:"policy_compliance"`
	ComplianceIssues    []ComplianceIssue   `json:"compliance_issues"`
	RequiredActions     []string            `json:"required_actions"`
	CertificationReady  bool                `json:"certification_ready"`
}

// Supporting types
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

type ValidationIssue struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Message     string `json:"message"`
	Resource    string `json:"resource,omitempty"`
	Remediation string `json:"remediation"`
}

type SecurityIssue struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Remediation string    `json:"remediation"`
}

type SecurityFinding struct {
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	Resource    string `json:"resource"`
	Finding     string `json:"finding"`
	Remediation string `json:"remediation"`
}

type PolicyViolation struct {
	Policy      string `json:"policy"`
	Violation   string `json:"violation"`
	Resource    string `json:"resource"`
	Severity    string `json:"severity"`
	Action      string `json:"action"`
}

type CostOptimization struct {
	Resource        string  `json:"resource"`
	CurrentCost     float64 `json:"current_cost"`
	OptimizedCost   float64 `json:"optimized_cost"`
	Savings         float64 `json:"savings"`
	Recommendation  string  `json:"recommendation"`
	Impact          string  `json:"impact"`
}

type KubernetesIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Resource    string `json:"resource"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion"`
}

type ComplianceIssue struct {
	Framework   string `json:"framework"`
	Control     string `json:"control"`
	Status      string `json:"status"`
	Finding     string `json:"finding"`
	Remediation string `json:"remediation"`
}

// NewInfrastructureValidator creates a new infrastructure validator
func NewInfrastructureValidator() *InfrastructureValidator {
	return &InfrastructureValidator{
		llmClient: llm.NewLLMClient(),
	}
}

// ValidateInfrastructure performs comprehensive infrastructure validation
func (iv *InfrastructureValidator) ValidateInfrastructure(ctx context.Context, infrastructureCode string, infraType string) (*InfraValidationResult, error) {
	logger.WithComponent("validation").Info("Starting infrastructure validation",
		zap.String("infrastructure_type", infraType))
	
	result := &InfraValidationResult{
		ValidatedAt:      time.Now(),
		CriticalIssues:   make([]ValidationIssue, 0),
		Recommendations:  make([]string, 0),
	}
	
	// Determine infrastructure type and validate accordingly
	switch strings.ToLower(infraType) {
	case "terraform", "tf":
		terraformResult, err := iv.validateTerraform(ctx, infrastructureCode)
		if err != nil {
			return nil, fmt.Errorf("terraform validation failed: %w", err)
		}
		result.TerraformResult = terraformResult
		
	case "kubernetes", "k8s":
		kubernetesResult, err := iv.validateKubernetes(ctx, infrastructureCode)
		if err != nil {
			return nil, fmt.Errorf("kubernetes validation failed: %w", err)
		}
		result.KubernetesResult = kubernetesResult
		
	default:
		// Try to auto-detect infrastructure type
		detectedType := iv.detectInfrastructureType(infrastructureCode)
		switch detectedType {
		case "terraform":
			terraformResult, _ := iv.validateTerraform(ctx, infrastructureCode)
			result.TerraformResult = terraformResult
		case "kubernetes":
			kubernetesResult, _ := iv.validateKubernetes(ctx, infrastructureCode)
			result.KubernetesResult = kubernetesResult
		}
	}
	
	// Universal validations (apply to all infrastructure types)
	securityResult := iv.validateSecurity(infrastructureCode)
	result.SecurityResult = securityResult
	
	// Cost estimation
	costResult := iv.estimateCosts(infrastructureCode, infraType)
	result.CostEstimation = costResult
	
	// Compliance validation
	complianceResult := iv.validateCompliance(infrastructureCode)
	result.ComplianceResult = complianceResult
	
	// Calculate overall scores and risk assessment
	result.OverallScore = iv.calculateOverallScore(result)
	result.DeploymentRisk = iv.assessDeploymentRisk(result)
	result.ValidationPassed = iv.determineValidationStatus(result)
	result.EstimatedDeployTime = iv.estimateDeploymentTime(result)
	
	// Aggregate critical issues and recommendations
	result.CriticalIssues = iv.aggregateCriticalIssues(result)
	result.Recommendations = iv.generateRecommendations(result)
	
	logger.WithComponent("validation").Info("Infrastructure validation completed",
		zap.Int("overall_score", result.OverallScore),
		zap.String("deployment_risk", string(result.DeploymentRisk)))
	
	return result, nil
}

// validateTerraform performs Terraform-specific validation
func (iv *InfrastructureValidator) validateTerraform(ctx context.Context, terraformCode string) (*TerraformValidationResult, error) {
	logger.WithComponent("validation").Info("Validating Terraform configuration")
	
	result := &TerraformValidationResult{
		SecurityIssues:   make([]SecurityIssue, 0),
		PolicyViolations: make([]PolicyViolation, 0),
		Optimizations:    make([]CostOptimization, 0),
	}
	
	// Syntax validation
	result.SyntaxValid = iv.validateTerraformSyntax(terraformCode)
	
	// Best practices validation using LLM
	bestPracticeScore, err := iv.validateTerraformBestPractices(ctx, terraformCode)
	if err != nil {
		logger.WithComponent("validation").Warn("Best practices validation failed",
			zap.Error(err))
		bestPracticeScore = 50 // Default fallback score
	}
	result.BestPracticeScore = bestPracticeScore
	
	// Security analysis
	securityScore, securityIssues := iv.analyzeTerraformSecurity(terraformCode)
	result.SecurityScore = securityScore
	result.SecurityIssues = securityIssues
	
	// Resource analysis
	result.ResourceCount = iv.countTerraformResources(terraformCode)
	result.ResourceEfficiency = iv.calculateResourceEfficiency(terraformCode)
	
	// Cost estimation
	estimatedCost := iv.estimateTerraformCosts(terraformCode)
	result.EstimatedCost = estimatedCost
	
	// Plan validation (dry-run simulation)
	result.PlanValid = iv.validateTerraformPlan(terraformCode)
	
	logger.WithComponent("validation").Info("Terraform validation completed",
		zap.Int("security_score", result.SecurityScore),
		zap.Int("best_practice_score", result.BestPracticeScore),
		zap.Float64("estimated_cost", result.EstimatedCost))
	
	return result, nil
}

// validateTerraformBestPractices uses LLM to validate Terraform best practices
func (iv *InfrastructureValidator) validateTerraformBestPractices(ctx context.Context, terraformCode string) (int, error) {
	prompt := fmt.Sprintf(`You are a senior DevOps engineer and Terraform expert reviewing infrastructure code for enterprise production deployment.

TERRAFORM BEST PRACTICES EVALUATION:

1. Resource Naming Conventions (consistent, descriptive names)
2. Variable and Output Usage (proper parameterization)
3. State Management (backend configuration, locking)
4. Provider Version Constraints (pinned versions)
5. Resource Dependencies (explicit dependencies)
6. Security Configuration (encryption, access controls)
7. Monitoring and Logging (CloudWatch, logging setup)
8. Backup and Disaster Recovery (backup policies)
9. Scalability and Availability (multi-AZ, auto-scaling)
10. Cost Optimization (right-sizing, reserved instances)

TERRAFORM CODE TO EVALUATE:
%s

PROVIDE ASSESSMENT:
- Overall best practices score (0-100)
- Security configuration score (0-100)
- Scalability design score (0-100)
- Cost efficiency score (0-100)
- Maintainability score (0-100)
- Critical issues that must be addressed
- Optimization recommendations
- Enterprise deployment readiness assessment

RESPOND WITH JSON:
{
  "best_practices_score": 85,
  "security_score": 90,
  "scalability_score": 80,
  "cost_efficiency_score": 75,
  "maintainability_score": 88,
  "critical_issues": ["Issue 1", "Issue 2"],
  "recommendations": ["Recommendation 1", "Recommendation 2"],
  "enterprise_ready": true,
  "confidence_level": 90
}`, terraformCode)

	response, err := iv.llmClient.Complete(ctx, prompt)
	if err != nil {
		return 0, fmt.Errorf("LLM validation failed: %w", err)
	}

	// Parse JSON response
	var assessment struct {
		BestPracticesScore int      `json:"best_practices_score"`
		SecurityScore      int      `json:"security_score"`
		ScalabilityScore   int      `json:"scalability_score"`
		CostEfficiencyScore int     `json:"cost_efficiency_score"`
		MaintainabilityScore int    `json:"maintainability_score"`
		CriticalIssues     []string `json:"critical_issues"`
		Recommendations    []string `json:"recommendations"`
		EnterpriseReady    bool     `json:"enterprise_ready"`
		ConfidenceLevel    int      `json:"confidence_level"`
	}

	if err := json.Unmarshal([]byte(response), &assessment); err != nil {
		logger.WithComponent("validation").Warn("Failed to parse LLM response, using fallback analysis")
		return iv.fallbackTerraformAnalysis(terraformCode), nil
	}

	return assessment.BestPracticesScore, nil
}

// validateKubernetes performs Kubernetes-specific validation
func (iv *InfrastructureValidator) validateKubernetes(ctx context.Context, kubernetesManifests string) (*KubernetesValidationResult, error) {
	logger.WithComponent("validation").Info("Validating Kubernetes manifests")
	
	result := &KubernetesValidationResult{
		Issues:          make([]KubernetesIssue, 0),
		Recommendations: make([]string, 0),
	}
	
	// Manifest syntax validation
	result.ManifestsValid = iv.validateKubernetesManifests(kubernetesManifests)
	
	// API version validation
	result.APIVersionsValid = iv.validateKubernetesAPIVersions(kubernetesManifests)
	
	// Production readiness assessment using LLM
	productionReadiness, err := iv.validateKubernetesProductionReadiness(ctx, kubernetesManifests)
	if err != nil {
		logger.WithComponent("validation").Warn("Production readiness validation failed",
			zap.Error(err))
		productionReadiness = 50
	}
	result.ProductionReadiness = productionReadiness
	
	// Security configuration analysis
	result.SecurityScore = iv.analyzeKubernetesSecurity(kubernetesManifests)
	
	// Resource configuration validation
	result.ResourceLimitsSet = iv.checkKubernetesResourceLimits(kubernetesManifests)
	result.HealthChecksSet = iv.checkKubernetesHealthChecks(kubernetesManifests)
	result.SecurityContextSet = iv.checkKubernetesSecurityContext(kubernetesManifests)
	result.NetworkPoliciesSet = iv.checkKubernetesNetworkPolicies(kubernetesManifests)
	
	// Scalability assessment
	result.ScalabilityScore = iv.assessKubernetesScalability(kubernetesManifests)
	
	// Policy compliance
	result.PolicyCompliance = iv.checkKubernetesPolicyCompliance(kubernetesManifests)
	
	// Generate issues and recommendations
	result.Issues = iv.generateKubernetesIssues(kubernetesManifests)
	result.Recommendations = iv.generateKubernetesRecommendations(kubernetesManifests)
	
	logger.WithComponent("validation").Info("Kubernetes validation completed",
		zap.Int("production_readiness", result.ProductionReadiness),
		zap.Int("security_score", result.SecurityScore))
	
	return result, nil
}

// Helper methods for validation logic
func (iv *InfrastructureValidator) detectInfrastructureType(code string) string {
	// Simple heuristic detection
	if strings.Contains(code, "resource \"") || strings.Contains(code, "terraform {") {
		return "terraform"
	}
	if strings.Contains(code, "apiVersion:") || strings.Contains(code, "kind:") {
		return "kubernetes"
	}
	return "unknown"
}

func (iv *InfrastructureValidator) validateTerraformSyntax(code string) bool {
	// Basic syntax validation using regex patterns
	resourcePattern := regexp.MustCompile(`resource\s+"[^"]+"\s+"[^"]+"\s+{`)
	if !resourcePattern.MatchString(code) && strings.Contains(code, "resource") {
		return false
	}
	
	// Check for balanced braces
	braceCount := strings.Count(code, "{") - strings.Count(code, "}")
	return braceCount == 0
}

func (iv *InfrastructureValidator) analyzeTerraformSecurity(code string) (int, []SecurityIssue) {
	issues := make([]SecurityIssue, 0)
	score := 100
	
	// Check for hardcoded secrets
	if strings.Contains(strings.ToLower(code), "password") && 
	   (strings.Contains(code, "=") && !strings.Contains(code, "var.")) {
		issues = append(issues, SecurityIssue{
			ID:          "SEC-001",
			Severity:    "HIGH",
			Title:       "Hardcoded Password Detected",
			Description: "Password appears to be hardcoded in configuration",
			Remediation: "Use variables or AWS Secrets Manager",
		})
		score -= 20
	}
	
	// Check for public access
	if strings.Contains(code, "0.0.0.0/0") {
		issues = append(issues, SecurityIssue{
			ID:          "SEC-002", 
			Severity:    "MEDIUM",
			Title:       "Overly Permissive Network Access",
			Description: "Resources allow access from any IP address",
			Remediation: "Restrict CIDR blocks to specific networks",
		})
		score -= 15
	}
	
	// Check for encryption
	if !strings.Contains(strings.ToLower(code), "encrypt") {
		issues = append(issues, SecurityIssue{
			ID:          "SEC-003",
			Severity:    "MEDIUM", 
			Title:       "Missing Encryption Configuration",
			Description: "No encryption configuration found",
			Remediation: "Enable encryption for storage and data transit",
		})
		score -= 10
	}
	
	return score, issues
}

func (iv *InfrastructureValidator) countTerraformResources(code string) int {
	resourcePattern := regexp.MustCompile(`resource\s+"[^"]+"\s+"[^"]+"\s+{`)
	matches := resourcePattern.FindAllString(code, -1)
	return len(matches)
}

func (iv *InfrastructureValidator) calculateResourceEfficiency(code string) int {
	// Simple heuristic: check for resource optimization patterns
	efficiency := 70 // Base efficiency score
	
	// Bonus for auto-scaling configuration
	if strings.Contains(strings.ToLower(code), "autoscaling") {
		efficiency += 10
	}
	
	// Bonus for monitoring setup
	if strings.Contains(strings.ToLower(code), "cloudwatch") || 
	   strings.Contains(strings.ToLower(code), "monitoring") {
		efficiency += 10
	}
	
	// Bonus for tagging
	if strings.Contains(strings.ToLower(code), "tags") {
		efficiency += 10
	}
	
	return efficiency
}

func (iv *InfrastructureValidator) estimateTerraformCosts(code string) float64 {
	// Simple cost estimation based on resource patterns
	cost := 0.0
	
	// EC2 instances
	if strings.Contains(code, "aws_instance") {
		instanceCount := strings.Count(code, "aws_instance")
		cost += float64(instanceCount) * 72.0 // ~$72/month per t3.medium
	}
	
	// RDS instances
	if strings.Contains(code, "aws_db_instance") {
		dbCount := strings.Count(code, "aws_db_instance")
		cost += float64(dbCount) * 150.0 // ~$150/month per db.t3.micro
	}
	
	// Load balancers
	if strings.Contains(code, "aws_lb") || strings.Contains(code, "aws_alb") {
		lbCount := strings.Count(code, "aws_lb") + strings.Count(code, "aws_alb")
		cost += float64(lbCount) * 22.0 // ~$22/month per ALB
	}
	
	return cost
}

func (iv *InfrastructureValidator) validateTerraformPlan(code string) bool {
	// Simulate terraform plan validation
	// In a real implementation, this would run actual terraform plan
	
	// Check for basic required configuration
	hasProvider := strings.Contains(code, "provider")
	hasResource := strings.Contains(code, "resource")
	
	return hasProvider && hasResource
}

func (iv *InfrastructureValidator) fallbackTerraformAnalysis(code string) int {
	score := 50 // Base score
	
	// Check for best practices indicators
	if strings.Contains(code, "variable") {
		score += 10
	}
	if strings.Contains(code, "output") {
		score += 10
	}
	if strings.Contains(code, "tags") {
		score += 10
	}
	if strings.Contains(code, "backend") {
		score += 10
	}
	if strings.Contains(code, "required_version") {
		score += 10
	}
	
	return score
}

// Kubernetes validation helper methods
func (iv *InfrastructureValidator) validateKubernetesManifests(manifests string) bool {
	// Check for required Kubernetes fields
	hasAPIVersion := strings.Contains(manifests, "apiVersion:")
	hasKind := strings.Contains(manifests, "kind:")
	hasMetadata := strings.Contains(manifests, "metadata:")
	
	return hasAPIVersion && hasKind && hasMetadata
}

func (iv *InfrastructureValidator) validateKubernetesAPIVersions(manifests string) bool {
	// Check for deprecated API versions
	deprecatedAPIs := []string{
		"extensions/v1beta1",
		"apps/v1beta1",
		"apps/v1beta2",
	}
	
	for _, api := range deprecatedAPIs {
		if strings.Contains(manifests, api) {
			return false
		}
	}
	
	return true
}

func (iv *InfrastructureValidator) validateKubernetesProductionReadiness(ctx context.Context, manifests string) (int, error) {
	prompt := fmt.Sprintf(`You are a Kubernetes expert reviewing manifests for production deployment.

KUBERNETES PRODUCTION READINESS CHECKLIST:

1. Resource Requests and Limits (CPU/memory)
2. Health Checks (liveness/readiness probes) 
3. Security Context (runAsNonRoot, securityContext)
4. Pod Disruption Budgets (availability during updates)
5. Horizontal Pod Autoscaling (scaling configuration)
6. Network Policies (network security)
7. Service Configuration (proper service types)
8. Persistent Volume Configuration (storage)
9. ConfigMaps and Secrets (configuration management)
10. RBAC Configuration (access control)

KUBERNETES MANIFESTS:
%s

EVALUATE PRODUCTION READINESS:
- Overall production readiness score (0-100)
- Security configuration score (0-100)
- Scalability readiness score (0-100)
- High availability score (0-100)
- Operational readiness score (0-100)
- Critical issues for production
- Optimization recommendations

RESPOND WITH JSON:
{
  "production_readiness_score": 85,
  "security_score": 90,
  "scalability_score": 80,
  "high_availability_score": 75,
  "operational_score": 88,
  "critical_issues": ["Issue 1", "Issue 2"],
  "recommendations": ["Recommendation 1", "Recommendation 2"],
  "production_ready": true
}`, manifests)

	response, err := iv.llmClient.Complete(ctx, prompt)
	if err != nil {
		return 0, fmt.Errorf("LLM validation failed: %w", err)
	}

	var assessment struct {
		ProductionReadinessScore int      `json:"production_readiness_score"`
		SecurityScore           int      `json:"security_score"`
		ScalabilityScore        int      `json:"scalability_score"`
		HighAvailabilityScore   int      `json:"high_availability_score"`
		OperationalScore        int      `json:"operational_score"`
		CriticalIssues          []string `json:"critical_issues"`
		Recommendations         []string `json:"recommendations"`
		ProductionReady         bool     `json:"production_ready"`
	}

	if err := json.Unmarshal([]byte(response), &assessment); err != nil {
		logger.WithComponent("validation").Warn("Failed to parse LLM response, using fallback analysis")
		return iv.fallbackKubernetesAnalysis(manifests), nil
	}

	return assessment.ProductionReadinessScore, nil
}

func (iv *InfrastructureValidator) analyzeKubernetesSecurity(manifests string) int {
	score := 100
	
	// Check for security context
	if !strings.Contains(manifests, "securityContext") {
		score -= 20
	}
	
	// Check for non-root user
	if !strings.Contains(manifests, "runAsNonRoot") {
		score -= 15
	}
	
	// Check for resource limits
	if !strings.Contains(manifests, "limits:") {
		score -= 15
	}
	
	// Check for network policies
	if !strings.Contains(manifests, "NetworkPolicy") {
		score -= 10
	}
	
	return score
}

func (iv *InfrastructureValidator) checkKubernetesResourceLimits(manifests string) bool {
	return strings.Contains(manifests, "limits:") && strings.Contains(manifests, "requests:")
}

func (iv *InfrastructureValidator) checkKubernetesHealthChecks(manifests string) bool {
	return strings.Contains(manifests, "livenessProbe:") || strings.Contains(manifests, "readinessProbe:")
}

func (iv *InfrastructureValidator) checkKubernetesSecurityContext(manifests string) bool {
	return strings.Contains(manifests, "securityContext:")
}

func (iv *InfrastructureValidator) checkKubernetesNetworkPolicies(manifests string) bool {
	return strings.Contains(manifests, "kind: NetworkPolicy")
}

func (iv *InfrastructureValidator) assessKubernetesScalability(manifests string) int {
	score := 70 // Base score
	
	// Check for HPA
	if strings.Contains(manifests, "HorizontalPodAutoscaler") {
		score += 20
	}
	
	// Check for resource requests (required for HPA)
	if strings.Contains(manifests, "requests:") {
		score += 10
	}
	
	return score
}

func (iv *InfrastructureValidator) checkKubernetesPolicyCompliance(manifests string) int {
	score := 80 // Base compliance score
	
	// Deduct for missing best practices
	if !strings.Contains(manifests, "limits:") {
		score -= 20
	}
	if !strings.Contains(manifests, "livenessProbe:") {
		score -= 15
	}
	if !strings.Contains(manifests, "securityContext:") {
		score -= 10
	}
	
	return score
}

func (iv *InfrastructureValidator) generateKubernetesIssues(manifests string) []KubernetesIssue {
	issues := make([]KubernetesIssue, 0)
	
	// Check for common issues
	if !strings.Contains(manifests, "limits:") {
		issues = append(issues, KubernetesIssue{
			Type:       "ResourceManagement",
			Severity:   "HIGH",
			Resource:   "Container",
			Message:    "Missing resource limits on containers",
			Suggestion: "Add resource limits to prevent containers from consuming excessive resources",
		})
	}
	
	if !strings.Contains(manifests, "livenessProbe:") && !strings.Contains(manifests, "readinessProbe:") {
		issues = append(issues, KubernetesIssue{
			Type:       "HealthChecks",
			Severity:   "HIGH",
			Resource:   "Pod",
			Message:    "Missing health check probes",
			Suggestion: "Add liveness and readiness probes to ensure proper health monitoring",
		})
	}
	
	return issues
}

func (iv *InfrastructureValidator) generateKubernetesRecommendations(manifests string) []string {
	recommendations := make([]string, 0)
	
	// Resource management recommendations
	if !strings.Contains(manifests, "requests:") {
		recommendations = append(recommendations, "Add resource requests to help Kubernetes scheduler make better placement decisions")
	}
	
	// Scalability recommendations
	if !strings.Contains(manifests, "HorizontalPodAutoscaler") {
		recommendations = append(recommendations, "Consider adding Horizontal Pod Autoscaler for automatic scaling based on metrics")
	}
	
	return recommendations
}

func (iv *InfrastructureValidator) fallbackKubernetesAnalysis(manifests string) int {
	score := 50 // Base score
	
	// Check for production readiness indicators
	if strings.Contains(manifests, "limits:") {
		score += 15
	}
	if strings.Contains(manifests, "livenessProbe:") {
		score += 15
	}
	if strings.Contains(manifests, "securityContext:") {
		score += 10
	}
	if strings.Contains(manifests, "replicas:") {
		score += 10
	}
	
	return score
}

// Universal validation methods
func (iv *InfrastructureValidator) validateSecurity(code string) *SecurityValidationResult {
	result := &SecurityValidationResult{
		CriticalFindings:        make([]SecurityFinding, 0),
		SecurityRecommendations: make([]string, 0),
	}
	
	// Basic security posture assessment
	result.SecurityPosture = 80 // Base score
	
	// Check encryption
	result.EncryptionEnabled = strings.Contains(strings.ToLower(code), "encrypt")
	if !result.EncryptionEnabled {
		result.SecurityPosture -= 15
		result.SecurityRecommendations = append(result.SecurityRecommendations, 
			"Enable encryption for data at rest and in transit")
	}
	
	// Check access control
	result.AccessControlValid = strings.Contains(strings.ToLower(code), "iam") || 
		strings.Contains(strings.ToLower(code), "rbac")
	if !result.AccessControlValid {
		result.SecurityPosture -= 10
	}
	
	// Check network security
	result.NetworkSecuritySet = strings.Contains(strings.ToLower(code), "security") ||
		strings.Contains(strings.ToLower(code), "network")
	
	// Check audit logging
	result.AuditLoggingEnabled = strings.Contains(strings.ToLower(code), "log") ||
		strings.Contains(strings.ToLower(code), "audit")
	
	// CIS compliance basic check
	result.CISCompliance = result.SecurityPosture
	
	return result
}

func (iv *InfrastructureValidator) estimateCosts(code string, infraType string) *CostEstimation {
	result := &CostEstimation{
		ResourceBreakdown:   make(map[string]float64),
		CostOptimizations:   make([]CostOptimization, 0),
	}
	
	// Basic cost estimation
	switch strings.ToLower(infraType) {
	case "terraform", "tf":
		result.MonthlyCost = iv.estimateTerraformCosts(code)
	case "kubernetes", "k8s":
		result.MonthlyCost = iv.estimateKubernetesCosts(code)
	default:
		result.MonthlyCost = 100.0 // Default estimation
	}
	
	result.YearlyCost = result.MonthlyCost * 12
	
	// Assess cost risk
	if result.MonthlyCost > 1000 {
		result.CostRisk = RiskHigh
	} else if result.MonthlyCost > 500 {
		result.CostRisk = RiskMedium
	} else {
		result.CostRisk = RiskLow
	}
	
	// Calculate cost efficiency
	result.CostEfficiencyScore = 75 // Base efficiency
	if strings.Contains(strings.ToLower(code), "spot") {
		result.CostEfficiencyScore += 15
	}
	
	// Budget recommendation
	if result.MonthlyCost < 100 {
		result.BudgetRecommendation = "Low cost infrastructure, monitor for unexpected growth"
	} else if result.MonthlyCost < 500 {
		result.BudgetRecommendation = "Moderate cost, consider setting up billing alerts"
	} else {
		result.BudgetRecommendation = "High cost infrastructure, requires careful cost management"
	}
	
	return result
}

func (iv *InfrastructureValidator) estimateKubernetesCosts(code string) float64 {
	cost := 0.0
	
	// EKS cluster
	if strings.Contains(code, "EKS") || strings.Contains(code, "eks") {
		cost += 72.0 // EKS control plane ~$72/month
	}
	
	// Estimate based on resource requests
	if strings.Contains(code, "requests:") {
		cost += 50.0 // Basic container costs
	}
	
	return cost
}

func (iv *InfrastructureValidator) validateCompliance(code string) *ComplianceValidationResult {
	result := &ComplianceValidationResult{
		ComplianceIssues: make([]ComplianceIssue, 0),
		RequiredActions:  make([]string, 0),
	}
	
	// Basic compliance checks
	result.SOC2Compliance = 75
	result.GDPRCompliance = 75
	result.HIPAACompliance = 75
	
	// Check encryption for compliance
	if !strings.Contains(strings.ToLower(code), "encrypt") {
		result.SOC2Compliance -= 20
		result.GDPRCompliance -= 25
		result.HIPAACompliance -= 30
	}
	
	// Overall policy compliance
	result.PolicyCompliance = (result.SOC2Compliance + result.GDPRCompliance + result.HIPAACompliance) / 3
	result.CertificationReady = result.PolicyCompliance >= 80
	
	return result
}

// Overall assessment methods
func (iv *InfrastructureValidator) calculateOverallScore(result *InfraValidationResult) int {
	scores := make([]int, 0)
	
	if result.TerraformResult != nil {
		scores = append(scores, result.TerraformResult.BestPracticeScore)
		scores = append(scores, result.TerraformResult.SecurityScore)
	}
	
	if result.KubernetesResult != nil {
		scores = append(scores, result.KubernetesResult.ProductionReadiness)
		scores = append(scores, result.KubernetesResult.SecurityScore)
	}
	
	if result.SecurityResult != nil {
		scores = append(scores, result.SecurityResult.SecurityPosture)
	}
	
	if result.ComplianceResult != nil {
		scores = append(scores, result.ComplianceResult.PolicyCompliance)
	}
	
	if len(scores) == 0 {
		return 0
	}
	
	total := 0
	for _, score := range scores {
		total += score
	}
	
	return total / len(scores)
}

func (iv *InfrastructureValidator) assessDeploymentRisk(result *InfraValidationResult) RiskLevel {
	// High cost = higher risk
	if result.CostEstimation != nil && result.CostEstimation.MonthlyCost > 1000 {
		return RiskHigh
	}
	
	// Low security score = higher risk
	if result.SecurityResult != nil && result.SecurityResult.SecurityPosture < 60 {
		return RiskHigh
	}
	
	// Critical issues = higher risk
	if len(result.CriticalIssues) > 2 {
		return RiskMedium
	}
	
	// Low overall score = medium risk
	if result.OverallScore < 70 {
		return RiskMedium
	}
	
	return RiskLow
}

func (iv *InfrastructureValidator) determineValidationStatus(result *InfraValidationResult) bool {
	// Must pass basic validation thresholds
	if result.OverallScore < 60 {
		return false
	}
	
	if result.SecurityResult != nil && result.SecurityResult.SecurityPosture < 70 {
		return false
	}
	
	if len(result.CriticalIssues) > 0 {
		return false
	}
	
	return true
}

func (iv *InfrastructureValidator) estimateDeploymentTime(result *InfraValidationResult) time.Duration {
	baseTime := 5 * time.Minute
	
	// Add time based on resource count
	if result.TerraformResult != nil {
		resourceTime := time.Duration(result.TerraformResult.ResourceCount) * 30 * time.Second
		baseTime += resourceTime
	}
	
	// Add time for complex Kubernetes deployments
	if result.KubernetesResult != nil && !result.KubernetesResult.ManifestsValid {
		baseTime += 10 * time.Minute
	}
	
	return baseTime
}

func (iv *InfrastructureValidator) aggregateCriticalIssues(result *InfraValidationResult) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	
	// Add Terraform critical issues
	if result.TerraformResult != nil {
		for _, secIssue := range result.TerraformResult.SecurityIssues {
			if secIssue.Severity == "HIGH" || secIssue.Severity == "CRITICAL" {
				issues = append(issues, ValidationIssue{
					Severity:    secIssue.Severity,
					Category:    "Terraform Security",
					Message:     secIssue.Title,
					Resource:    secIssue.Resource,
					Remediation: secIssue.Remediation,
				})
			}
		}
	}
	
	// Add Kubernetes critical issues
	if result.KubernetesResult != nil {
		for _, k8sIssue := range result.KubernetesResult.Issues {
			if k8sIssue.Severity == "HIGH" || k8sIssue.Severity == "CRITICAL" {
				issues = append(issues, ValidationIssue{
					Severity:    k8sIssue.Severity,
					Category:    "Kubernetes",
					Message:     k8sIssue.Message,
					Resource:    k8sIssue.Resource,
					Remediation: k8sIssue.Suggestion,
				})
			}
		}
	}
	
	return issues
}

func (iv *InfrastructureValidator) generateRecommendations(result *InfraValidationResult) []string {
	recommendations := make([]string, 0)
	
	// Cost optimization recommendations
	if result.CostEstimation != nil {
		if result.CostEstimation.MonthlyCost > 500 {
			recommendations = append(recommendations, "Consider using reserved instances for cost savings")
		}
		if result.CostEstimation.CostEfficiencyScore < 80 {
			recommendations = append(recommendations, "Review resource sizing for cost optimization")
		}
	}
	
	// Security recommendations
	if result.SecurityResult != nil {
		if result.SecurityResult.SecurityPosture < 85 {
			recommendations = append(recommendations, "Implement additional security controls")
		}
		if !result.SecurityResult.EncryptionEnabled {
			recommendations = append(recommendations, "Enable encryption for data at rest and in transit")
		}
	}
	
	// Infrastructure-specific recommendations
	if result.TerraformResult != nil {
		if result.TerraformResult.BestPracticeScore < 80 {
			recommendations = append(recommendations, "Follow Terraform best practices for maintainability")
		}
	}
	
	if result.KubernetesResult != nil {
		if !result.KubernetesResult.ResourceLimitsSet {
			recommendations = append(recommendations, "Set resource requests and limits for all containers")
		}
		if !result.KubernetesResult.HealthChecksSet {
			recommendations = append(recommendations, "Configure liveness and readiness probes")
		}
	}
	
	return recommendations
}