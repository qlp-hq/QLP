package main

import (
	"context"
	"log"

	"QLP/internal/validation"
)

func main() {
	log.Println("🔧 INFRASTRUCTURE VALIDATION TEST")
	log.Println("=================================")

	// Test comprehensive infrastructure validation
	testTerraformValidation()
	testKubernetesValidation()
	testUniversalValidation()
	
	log.Println("✅ INFRASTRUCTURE VALIDATION TEST COMPLETED!")
}

func testTerraformValidation() {
	log.Println("\n🏗️ Testing Terraform Infrastructure Validation")
	log.Println("----------------------------------------------")
	
	// Create infrastructure validator
	validator := validation.NewInfrastructureValidator()
	ctx := context.Background()
	
	// Sample Terraform code with various validation scenarios
	terraformCode := `terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "infrastructure/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

# VPC Configuration
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  
  tags = {
    Name        = "${var.environment}-vpc"
    Environment = var.environment
  }
}

# Security Group with some issues for testing
resource "aws_security_group" "web" {
  name_prefix = "${var.environment}-web-"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]  # Security issue: overly permissive
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-web-sg"
    Environment = var.environment
  }
}

# EC2 Instances
resource "aws_instance" "web" {
  count                  = 2
  ami                    = "ami-0c55b159cbfafe1d0"
  instance_type          = "t3.medium"
  subnet_id              = aws_subnet.web[count.index].id
  vpc_security_group_ids = [aws_security_group.web.id]
  
  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    environment = var.environment
  }))

  tags = {
    Name        = "${var.environment}-web-${count.index + 1}"
    Environment = var.environment
    Type        = "web-server"
  }
}

# RDS Database
resource "aws_db_instance" "main" {
  identifier           = "${var.environment}-database"
  engine              = "mysql"
  engine_version      = "8.0"
  instance_class      = "db.t3.medium"
  allocated_storage   = 20
  max_allocated_storage = 100
  
  db_name  = "myapp"
  username = "admin"
  password = var.db_password  # Good: using variable
  
  vpc_security_group_ids = [aws_security_group.db.id]
  db_subnet_group_name   = aws_db_subnet_group.main.name
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  encryption_at_rest_enabled = true  # Good: encryption enabled
  
  tags = {
    Name        = "${var.environment}-database"
    Environment = var.environment
  }
}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.environment}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets           = aws_subnet.web[*].id

  enable_deletion_protection = true

  tags = {
    Name        = "${var.environment}-alb"
    Environment = var.environment
  }
}

# CloudWatch for monitoring
resource "aws_cloudwatch_log_group" "app" {
  name              = "/aws/ec2/${var.environment}"
  retention_in_days = 30

  tags = {
    Environment = var.environment
  }
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "load_balancer_dns" {
  description = "DNS name of the load balancer"
  value       = aws_lb.main.dns_name
}`

	// Validate Terraform infrastructure
	log.Printf("🔍 Validating Terraform infrastructure code...")
	result, err := validator.ValidateInfrastructure(ctx, terraformCode, "terraform")
	if err != nil {
		log.Fatalf("Terraform validation failed: %v", err)
	}

	// Display comprehensive results
	log.Printf("✅ Terraform validation completed!")
	log.Printf("📊 Overall Score: %d/100", result.OverallScore)
	log.Printf("🚨 Deployment Risk: %s", result.DeploymentRisk)
	log.Printf("✔️ Validation Passed: %v", result.ValidationPassed)
	
	if result.TerraformResult != nil {
		tf := result.TerraformResult
		log.Printf("\n🏗️ Terraform Specific Results:")
		log.Printf("   📝 Syntax Valid: %v", tf.SyntaxValid)
		log.Printf("   📋 Plan Valid: %v", tf.PlanValid)
		log.Printf("   🔒 Security Score: %d/100", tf.SecurityScore)
		log.Printf("   🎯 Best Practice Score: %d/100", tf.BestPracticeScore)
		log.Printf("   📊 Resource Count: %d", tf.ResourceCount)
		log.Printf("   💰 Estimated Cost: $%.2f/month", tf.EstimatedCost)
		log.Printf("   ⚡ Resource Efficiency: %d/100", tf.ResourceEfficiency)
		
		if len(tf.SecurityIssues) > 0 {
			log.Printf("   🚨 Security Issues Found: %d", len(tf.SecurityIssues))
			for _, issue := range tf.SecurityIssues {
				log.Printf("      - [%s] %s: %s", issue.Severity, issue.Title, issue.Description)
			}
		}
		
		if len(tf.Optimizations) > 0 {
			log.Printf("   💡 Cost Optimizations: %d", len(tf.Optimizations))
			for _, opt := range tf.Optimizations {
				log.Printf("      - %s: Save $%.2f/month (%s)", opt.Resource, opt.Savings, opt.Recommendation)
			}
		}
	}

	// Display security analysis
	if result.SecurityResult != nil {
		sec := result.SecurityResult
		log.Printf("\n🔒 Security Analysis:")
		log.Printf("   🛡️ Security Posture: %d/100", sec.SecurityPosture)
		log.Printf("   📋 CIS Compliance: %d/100", sec.CISCompliance)
		log.Printf("   🚨 Vulnerabilities: %d", sec.VulnerabilityCount)
		log.Printf("   🔐 Encryption Enabled: %v", sec.EncryptionEnabled)
		log.Printf("   🎫 Access Control Valid: %v", sec.AccessControlValid)
		log.Printf("   🌐 Network Security Set: %v", sec.NetworkSecuritySet)
		log.Printf("   📝 Audit Logging Enabled: %v", sec.AuditLoggingEnabled)
	}

	// Display cost estimation
	if result.CostEstimation != nil {
		cost := result.CostEstimation
		log.Printf("\n💰 Cost Analysis:")
		log.Printf("   📊 Monthly Cost: $%.2f", cost.MonthlyCost)
		log.Printf("   📈 Yearly Cost: $%.2f", cost.YearlyCost)
		log.Printf("   ⚠️ Cost Risk: %s", cost.CostRisk)
		log.Printf("   ⚡ Cost Efficiency: %d/100", cost.CostEfficiencyScore)
		log.Printf("   💡 Budget Recommendation: %s", cost.BudgetRecommendation)
	}

	// Display compliance status
	if result.ComplianceResult != nil {
		comp := result.ComplianceResult
		log.Printf("\n📋 Compliance Analysis:")
		log.Printf("   🏢 SOC2 Compliance: %d/100", comp.SOC2Compliance)
		log.Printf("   🇪🇺 GDPR Compliance: %d/100", comp.GDPRCompliance)
		log.Printf("   🏥 HIPAA Compliance: %d/100", comp.HIPAACompliance)
		log.Printf("   📖 Policy Compliance: %d/100", comp.PolicyCompliance)
		log.Printf("   ✅ Certification Ready: %v", comp.CertificationReady)
	}

	// Display critical issues
	if len(result.CriticalIssues) > 0 {
		log.Printf("\n🚨 Critical Issues (%d):", len(result.CriticalIssues))
		for _, issue := range result.CriticalIssues {
			log.Printf("   - [%s] %s: %s", issue.Severity, issue.Category, issue.Message)
			log.Printf("     💡 Remediation: %s", issue.Remediation)
		}
	}

	// Display recommendations
	if len(result.Recommendations) > 0 {
		log.Printf("\n💡 Recommendations (%d):", len(result.Recommendations))
		for _, rec := range result.Recommendations {
			log.Printf("   - %s", rec)
		}
	}

	log.Printf("\n⏱️ Estimated Deployment Time: %v", result.EstimatedDeployTime)
}

func testKubernetesValidation() {
	log.Println("\n☸️ Testing Kubernetes Infrastructure Validation")
	log.Println("---------------------------------------------")
	
	validator := validation.NewInfrastructureValidator()
	ctx := context.Background()
	
	// Sample Kubernetes manifests with various scenarios
	kubernetesManifests := `# Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    environment: production
    tier: application

---
# ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  database_host: "db.example.com"
  cache_timeout: "300"
  log_level: "info"

---
# Secret
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: production
type: Opaque
data:
  database_password: cGFzc3dvcmQxMjM=  # password123 base64 encoded
  api_key: YWJjZGVmZ2hpams=

---
# Deployment with some production readiness issues
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: production
  labels:
    app: web-app
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
        version: v1.0.0
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: web-app
        image: myregistry/web-app:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: DATABASE_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: database_host
        - name: DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database_password
        # Missing resource limits and requests - production issue
        # Missing health checks - production issue
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          capabilities:
            drop:
            - ALL

---
# Service
apiVersion: v1
kind: Service
metadata:
  name: web-app-service
  namespace: production
  labels:
    app: web-app
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: web-app

---
# Ingress
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-app-ingress
  namespace: production
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - app.example.com
    secretName: web-app-tls
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-app-service
            port:
              number: 80

---
# HorizontalPodAutoscaler
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-app-hpa
  namespace: production
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80

---
# PodDisruptionBudget
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: web-app-pdb
  namespace: production
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: web-app

---
# NetworkPolicy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: web-app-netpol
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: web-app
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: database
    ports:
    - protocol: TCP
      port: 5432`

	// Validate Kubernetes infrastructure
	log.Printf("🔍 Validating Kubernetes manifests...")
	result, err := validator.ValidateInfrastructure(ctx, kubernetesManifests, "kubernetes")
	if err != nil {
		log.Fatalf("Kubernetes validation failed: %v", err)
	}

	// Display comprehensive results
	log.Printf("✅ Kubernetes validation completed!")
	log.Printf("📊 Overall Score: %d/100", result.OverallScore)
	log.Printf("🚨 Deployment Risk: %s", result.DeploymentRisk)
	log.Printf("✔️ Validation Passed: %v", result.ValidationPassed)
	
	if result.KubernetesResult != nil {
		k8s := result.KubernetesResult
		log.Printf("\n☸️ Kubernetes Specific Results:")
		log.Printf("   📝 Manifests Valid: %v", k8s.ManifestsValid)
		log.Printf("   🔧 API Versions Valid: %v", k8s.APIVersionsValid)
		log.Printf("   🏭 Production Readiness: %d/100", k8s.ProductionReadiness)
		log.Printf("   📈 Scalability Score: %d/100", k8s.ScalabilityScore)
		log.Printf("   🔒 Security Score: %d/100", k8s.SecurityScore)
		log.Printf("   📋 Policy Compliance: %d/100", k8s.PolicyCompliance)
		log.Printf("   📊 Resource Limits Set: %v", k8s.ResourceLimitsSet)
		log.Printf("   ❤️ Health Checks Set: %v", k8s.HealthChecksSet)
		log.Printf("   🔐 Security Context Set: %v", k8s.SecurityContextSet)
		log.Printf("   🌐 Network Policies Set: %v", k8s.NetworkPoliciesSet)
		
		if len(k8s.Issues) > 0 {
			log.Printf("   🚨 Issues Found: %d", len(k8s.Issues))
			for _, issue := range k8s.Issues {
				log.Printf("      - [%s] %s: %s", issue.Severity, issue.Type, issue.Message)
			}
		}
		
		if len(k8s.Recommendations) > 0 {
			log.Printf("   💡 Recommendations: %d", len(k8s.Recommendations))
			for _, rec := range k8s.Recommendations {
				log.Printf("      - %s", rec)
			}
		}
	}

	log.Printf("\n⏱️ Estimated Deployment Time: %v", result.EstimatedDeployTime)
}

func testUniversalValidation() {
	log.Println("\n🌐 Testing Universal Infrastructure Validation")
	log.Println("--------------------------------------------")
	
	validator := validation.NewInfrastructureValidator()
	ctx := context.Background()
	
	// Test with mixed/unknown infrastructure type
	mixedInfraCode := `# Mixed infrastructure example
terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_eks_cluster" "main" {
  name     = "production-cluster"
  role_arn = aws_iam_role.cluster.arn
  
  vpc_config {
    subnet_ids = aws_subnet.cluster[*].id
  }
}

---
# This will be detected as Kubernetes
apiVersion: v1
kind: Service
metadata:
  name: mixed-service
spec:
  type: LoadBalancer
  ports:
  - port: 80`

	log.Printf("🔍 Validating mixed infrastructure (auto-detection)...")
	result, err := validator.ValidateInfrastructure(ctx, mixedInfraCode, "auto")
	if err != nil {
		log.Fatalf("Mixed validation failed: %v", err)
	}

	log.Printf("✅ Mixed infrastructure validation completed!")
	log.Printf("📊 Overall Score: %d/100", result.OverallScore)
	log.Printf("🚨 Deployment Risk: %s", result.DeploymentRisk)
	log.Printf("✔️ Validation Passed: %v", result.ValidationPassed)
	
	// Show what was detected and validated
	if result.TerraformResult != nil {
		log.Printf("🏗️ Terraform components detected and validated")
	}
	if result.KubernetesResult != nil {
		log.Printf("☸️ Kubernetes components detected and validated")
	}
	
	log.Printf("\n🎯 Universal validation provides multi-layer analysis regardless of infrastructure type!")
	log.Printf("🛡️ This enables bulletproof confidence for any deployment scenario.")
}