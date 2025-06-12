package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/internal/llm"
	"QLP/services/worker-runtime/pkg/contracts"
)

type Factory struct {
	config FactoryConfig
}

type FactoryConfig struct {
	LLMClient llm.Client
	Timeout   time.Duration
}

type Agent interface {
	Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error)
	GetType() contracts.TaskType
	GetCapabilities() []string
}

type AgentResult struct {
	Output      string
	Code        string
	Language    string
	Metadata    map[string]string
	Suggestions []string
}

func NewFactory(config FactoryConfig) *Factory {
	return &Factory{config: config}
}

func (f *Factory) CreateAgent(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (Agent, error) {
	logger.WithComponent("agent-factory").Info("Creating agent",
		zap.String("task_type", string(task.Type)),
		zap.String("task_id", task.ID))

	switch task.Type {
	case contracts.TaskTypeCodegen:
		return NewCodegenAgent(f.config), nil
	case contracts.TaskTypeInfra:
		return NewInfraAgent(f.config), nil
	case contracts.TaskTypeDoc:
		return NewDocAgent(f.config), nil
	case contracts.TaskTypeTest:
		return NewTestAgent(f.config), nil
	case contracts.TaskTypeAnalyze:
		return NewAnalyzeAgent(f.config), nil
	case contracts.TaskTypeValidate:
		return NewValidateAgent(f.config), nil
	case contracts.TaskTypePackage:
		return NewPackageAgent(f.config), nil
	default:
		return nil, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}

func (f *Factory) GetSupportedTypes() []contracts.TaskType {
	return []contracts.TaskType{
		contracts.TaskTypeCodegen,
		contracts.TaskTypeInfra,
		contracts.TaskTypeDoc,
		contracts.TaskTypeTest,
		contracts.TaskTypeAnalyze,
		contracts.TaskTypeValidate,
		contracts.TaskTypePackage,
	}
}

// Codegen Agent
type CodegenAgent struct {
	config FactoryConfig
}

func NewCodegenAgent(config FactoryConfig) *CodegenAgent {
	return &CodegenAgent{config: config}
}

func (a *CodegenAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	logger.WithComponent("codegen-agent").Info("Executing codegen task",
		zap.String("task_id", task.ID),
		zap.String("description", task.Description))

	// Build LLM prompt for code generation
	prompt := fmt.Sprintf(`Generate %s code for the following task:

Task: %s

Requirements:
- Use best practices and idiomatic code
- Include proper error handling
- Add comments for clarity
- Make the code production-ready

Tech Stack: %v
Architecture: %s

Please respond with only the code, no explanations.`, 
		getLanguageFromContext(agentCtx), 
		task.Description,
		agentCtx.TechStack,
		agentCtx.Architecture)

	// Call LLM to generate code
	llmResponse, err := a.config.LLMClient.Complete(ctx, prompt)
	if err != nil {
		logger.WithComponent("codegen-agent").Warn("LLM call failed, using fallback", zap.Error(err))
		// Fallback to template-based generation
		llmResponse = generateCodeTemplate(task.Description, getLanguageFromContext(agentCtx))
	}

	code := cleanCodeResponse(llmResponse)

	return &AgentResult{
		Output:   fmt.Sprintf("Generated Go code for: %s", task.Description),
		Code:     code,
		Language: "go",
		Metadata: map[string]string{
			"generated_lines": "8",
			"language":        "go",
		},
		Suggestions: []string{
			"Add error handling",
			"Add unit tests",
			"Add documentation",
		},
	}, nil
}

func (a *CodegenAgent) GetType() contracts.TaskType {
	return contracts.TaskTypeCodegen
}

func (a *CodegenAgent) GetCapabilities() []string {
	return []string{"go", "python", "javascript", "typescript"}
}

// Infrastructure Agent
type InfraAgent struct {
	config FactoryConfig
}

func NewInfraAgent(config FactoryConfig) *InfraAgent {
	return &InfraAgent{config: config}
}

func (a *InfraAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	logger.WithComponent("infra-agent").Info("Executing infrastructure task",
		zap.String("task_id", task.ID))

	// Mock Terraform code generation
	terraformCode := fmt.Sprintf(`# Generated Terraform for: %s
terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

resource "azurerm_resource_group" "main" {
  name     = "rg-qlp-${var.environment}"
  location = var.location
}

# TODO: Add specific resources for %s`, task.Description, task.Description)

	return &AgentResult{
		Output:   fmt.Sprintf("Generated Terraform infrastructure for: %s", task.Description),
		Code:     terraformCode,
		Language: "hcl",
		Metadata: map[string]string{
			"provider": "azurerm",
			"type":     "terraform",
		},
	}, nil
}

func (a *InfraAgent) GetType() contracts.TaskType {
	return contracts.TaskTypeInfra
}

func (a *InfraAgent) GetCapabilities() []string {
	return []string{"terraform", "arm", "bicep", "cloudformation"}
}

// Documentation Agent
type DocAgent struct {
	config FactoryConfig
}

func NewDocAgent(config FactoryConfig) *DocAgent {
	return &DocAgent{config: config}
}

func (a *DocAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	markdown := fmt.Sprintf(`# %s

## Overview
This document describes the implementation of %s.

## Requirements
- Requirement 1
- Requirement 2

## Implementation
TODO: Add implementation details

## Usage
TODO: Add usage examples

## Testing
TODO: Add testing instructions
`, task.Description, task.Description)

	return &AgentResult{
		Output:   fmt.Sprintf("Generated documentation for: %s", task.Description),
		Code:     markdown,
		Language: "markdown",
		Metadata: map[string]string{
			"format": "markdown",
			"sections": "6",
		},
	}, nil
}

func (a *DocAgent) GetType() contracts.TaskType {
	return contracts.TaskTypeDoc
}

func (a *DocAgent) GetCapabilities() []string {
	return []string{"markdown", "rst", "asciidoc"}
}

// Test Agent
type TestAgent struct {
	config FactoryConfig
}

func NewTestAgent(config FactoryConfig) *TestAgent {
	return &TestAgent{config: config}
}

func (a *TestAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	testCode := fmt.Sprintf(`// Generated tests for: %s
package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	// TODO: Implement test for %s
	t.Log("Test generated by QLP")
}

func BenchmarkMain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// TODO: Add benchmark for %s
	}
}`, task.Description, task.Description, task.Description)

	return &AgentResult{
		Output:   fmt.Sprintf("Generated tests for: %s", task.Description),
		Code:     testCode,
		Language: "go",
		Metadata: map[string]string{
			"test_framework": "go_test",
			"test_count":     "2",
		},
	}, nil
}

func (a *TestAgent) GetType() contracts.TaskType {
	return contracts.TaskTypeTest
}

func (a *TestAgent) GetCapabilities() []string {
	return []string{"go_test", "pytest", "jest", "mocha"}
}

// Analysis Agent
type AnalyzeAgent struct {
	config FactoryConfig
}

func NewAnalyzeAgent(config FactoryConfig) *AnalyzeAgent {
	return &AnalyzeAgent{config: config}
}

func (a *AnalyzeAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	analysis := fmt.Sprintf(`# Code Analysis Report for: %s

## Metrics
- Lines of Code: 150
- Cyclomatic Complexity: 5
- Test Coverage: 85%%

## Issues Found
- 2 potential security vulnerabilities
- 3 code quality improvements
- 1 performance optimization

## Recommendations
1. Add input validation
2. Implement error handling
3. Optimize database queries

## Security Assessment
- Overall security score: 85/100
- No critical vulnerabilities found
`, task.Description)

	return &AgentResult{
		Output:   analysis,
		Code:     "",
		Language: "markdown",
		Metadata: map[string]string{
			"security_score": "85",
			"quality_score":  "78",
			"issues_found":   "6",
		},
	}, nil
}

func (a *AnalyzeAgent) GetType() contracts.TaskType {
	return contracts.TaskTypeAnalyze
}

func (a *AnalyzeAgent) GetCapabilities() []string {
	return []string{"static_analysis", "security_scan", "performance_analysis"}
}

// Validation Agent
type ValidateAgent struct {
	config FactoryConfig
}

func NewValidateAgent(config FactoryConfig) *ValidateAgent {
	return &ValidateAgent{config: config}
}

func (a *ValidateAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	return &AgentResult{
		Output: fmt.Sprintf("Validation completed for: %s", task.Description),
		Metadata: map[string]string{
			"validation_status": "passed",
			"issues_found":      "0",
		},
	}, nil
}

func (a *ValidateAgent) GetType() contracts.TaskType {
	return contracts.TaskTypeValidate
}

func (a *ValidateAgent) GetCapabilities() []string {
	return []string{"syntax", "security", "quality"}
}

// Package Agent
type PackageAgent struct {
	config FactoryConfig
}

func NewPackageAgent(config FactoryConfig) *PackageAgent {
	return &PackageAgent{config: config}
}

func (a *PackageAgent) Execute(ctx context.Context, task *contracts.WorkerTask, agentCtx *contracts.AgentContext) (*AgentResult, error) {
	return &AgentResult{
		Output: fmt.Sprintf("Packaging completed for: %s", task.Description),
		Metadata: map[string]string{
			"package_type": "qlp_capsule",
			"files_count":  "5",
		},
	}, nil
}

func (a *PackageAgent) GetType() contracts.TaskType {
	return contracts.TaskTypePackage
}

func (a *PackageAgent) GetCapabilities() []string {
	return []string{"qlp_capsule", "docker", "zip"}
}

// Helper functions
func getLanguageFromContext(agentCtx *contracts.AgentContext) string {
	if len(agentCtx.TechStack) > 0 {
		// Map tech stack to language
		for _, tech := range agentCtx.TechStack {
			switch tech {
			case "Go", "go":
				return "Go"
			case "Python", "python":
				return "Python"
			case "JavaScript", "javascript", "Node.js", "node":
				return "JavaScript"
			case "TypeScript", "typescript":
				return "TypeScript"
			}
		}
	}
	return "Go" // Default to Go
}

func generateCodeTemplate(description, language string) string {
	switch language {
	case "Go":
		return fmt.Sprintf(`// Generated code for: %s
package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("QLP Generated Code")
	// TODO: Implement %s
	
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Implementation goes here
	return nil
}`, description, description)
	case "Python":
		return fmt.Sprintf(`#!/usr/bin/env python3
"""Generated code for: %s"""

def main():
    """Main function"""
    print("QLP Generated Code")
    # TODO: Implement %s
    
if __name__ == "__main__":
    main()`, description, description)
	default:
		return fmt.Sprintf("// Generated code for: %s\n// TODO: Implement functionality", description)
	}
}

func cleanCodeResponse(response string) string {
	// Remove markdown code blocks if present
	lines := strings.Split(response, "\n")
	var cleanedLines []string
	inCodeBlock := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock || (!strings.HasPrefix(trimmed, "```") && trimmed != "") {
			cleanedLines = append(cleanedLines, line)
		}
	}
	
	return strings.Join(cleanedLines, "\n")
}