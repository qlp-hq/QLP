package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"QLP/internal/logger"
	"go.uber.org/zap"
)

// UnifiedResponseParser provides centralized LLM response parsing capabilities
type UnifiedResponseParser struct {
	logger     logger.Interface
	extractors map[ResponseType]ResponseExtractor
	validators map[ResponseType]ResponseValidator
	cleaners   []ResponseCleaner
}

// ResponseType enum for different response formats
type ResponseType string

const (
	ResponseTypeJSON       ResponseType = "json"
	ResponseTypeStructured ResponseType = "structured"
	ResponseTypeText       ResponseType = "text"
	ResponseTypeValidation ResponseType = "validation"
	ResponseTypeAnalysis   ResponseType = "analysis"
)

// ParsedResponse represents a successfully parsed LLM response
type ParsedResponse struct {
	Type       ResponseType    `json:"type"`
	Data       interface{}     `json:"data"`
	Confidence float64         `json:"confidence"`
	ParsedAt   time.Time       `json:"parsed_at"`
	Metadata   ResponseMetadata `json:"metadata"`
}

// ResponseMetadata contains parsing information
type ResponseMetadata struct {
	OriginalLength int           `json:"original_length"`
	CleanedLength  int           `json:"cleaned_length"`
	ParseDuration  time.Duration `json:"parse_duration"`
	ExtractorUsed  string        `json:"extractor_used"`
	ValidatorUsed  string        `json:"validator_used"`
	Warnings       []string      `json:"warnings,omitempty"`
}

// ResponseExtractor interface for extracting structured data from raw responses
type ResponseExtractor interface {
	Extract(ctx context.Context, rawResponse string) (string, error)
	GetType() ResponseType
	GetPriority() int
}

// ResponseValidator interface for validating parsed responses
type ResponseValidator interface {
	Validate(ctx context.Context, response interface{}) error
	GetType() ResponseType
}

// ResponseCleaner interface for preprocessing raw responses
type ResponseCleaner interface {
	Clean(ctx context.Context, rawResponse string) (string, error)
	GetPriority() int
}

// NewUnifiedResponseParser creates a new unified parser with default configurations
func NewUnifiedResponseParser(logger logger.Interface) *UnifiedResponseParser {
	parser := &UnifiedResponseParser{
		logger:     logger.WithComponent("llm_parser"),
		extractors: make(map[ResponseType]ResponseExtractor),
		validators: make(map[ResponseType]ResponseValidator),
		cleaners:   make([]ResponseCleaner, 0),
	}

	parser.initializeDefaultComponents()
	return parser
}

// Parse processes a raw LLM response and returns structured data
func (urp *UnifiedResponseParser) Parse(ctx context.Context, rawResponse string, expectedType ResponseType) (*ParsedResponse, error) {
	startTime := time.Now()
	
	urp.logger.Debug("Starting LLM response parsing",
		zap.String("expected_type", string(expectedType)),
		zap.Int("raw_length", len(rawResponse)),
	)

	// Phase 1: Clean the response
	cleanedResponse, warnings, err := urp.cleanResponse(ctx, rawResponse)
	if err != nil {
		urp.logger.Error("Failed to clean response", zap.Error(err))
		return nil, fmt.Errorf("response cleaning failed: %w", err)
	}

	// Phase 2: Extract structured data
	extractor, exists := urp.extractors[expectedType]
	if !exists {
		urp.logger.Warn("No extractor found for type, using JSON fallback",
			zap.String("expected_type", string(expectedType)),
		)
		extractor = urp.extractors[ResponseTypeJSON]
	}

	extractedData, err := extractor.Extract(ctx, cleanedResponse)
	if err != nil {
		urp.logger.Error("Failed to extract data", 
			zap.String("extractor_type", string(extractor.GetType())),
			zap.Error(err),
		)
		return nil, fmt.Errorf("data extraction failed: %w", err)
	}

	// Phase 3: Parse into target structure
	var parsedData interface{}
	switch expectedType {
	case ResponseTypeValidation:
		parsedData = &ValidationResponse{}
	case ResponseTypeAnalysis:
		parsedData = &AnalysisResponse{}
	case ResponseTypeJSON:
		parsedData = make(map[string]interface{})
	default:
		parsedData = &GenericResponse{}
	}

	if err := json.Unmarshal([]byte(extractedData), parsedData); err != nil {
		urp.logger.Error("Failed to unmarshal extracted data", zap.Error(err))
		return nil, fmt.Errorf("JSON unmarshaling failed: %w", err)
	}

	// Phase 4: Validate the parsed data
	if validator, exists := urp.validators[expectedType]; exists {
		if err := validator.Validate(ctx, parsedData); err != nil {
			urp.logger.Warn("Response validation failed", zap.Error(err))
			warnings = append(warnings, fmt.Sprintf("Validation failed: %v", err))
		}
	}

	// Create result
	result := &ParsedResponse{
		Type:       expectedType,
		Data:       parsedData,
		Confidence: urp.calculateParseConfidence(rawResponse, extractedData, warnings),
		ParsedAt:   time.Now(),
		Metadata: ResponseMetadata{
			OriginalLength: len(rawResponse),
			CleanedLength:  len(cleanedResponse),
			ParseDuration:  time.Since(startTime),
			ExtractorUsed:  string(extractor.GetType()),
			ValidatorUsed:  string(expectedType),
			Warnings:       warnings,
		},
	}

	urp.logger.Info("LLM response parsing completed",
		zap.String("type", string(expectedType)),
		zap.Float64("confidence", result.Confidence),
		zap.Duration("duration", result.Metadata.ParseDuration),
		zap.Int("warnings", len(warnings)),
	)

	return result, nil
}

// ParseWithFallback attempts parsing with fallback strategies
func (urp *UnifiedResponseParser) ParseWithFallback(ctx context.Context, rawResponse string, preferredType ResponseType, fallbackTypes ...ResponseType) (*ParsedResponse, error) {
	// Try preferred type first
	result, err := urp.Parse(ctx, rawResponse, preferredType)
	if err == nil && result.Confidence >= 0.7 {
		return result, nil
	}

	urp.logger.Warn("Primary parsing failed, trying fallbacks",
		zap.String("preferred_type", string(preferredType)),
		zap.Error(err),
	)

	// Try fallback types
	for _, fallbackType := range fallbackTypes {
		result, err := urp.Parse(ctx, rawResponse, fallbackType)
		if err == nil && result.Confidence >= 0.5 {
			urp.logger.Info("Fallback parsing succeeded",
				zap.String("fallback_type", string(fallbackType)),
				zap.Float64("confidence", result.Confidence),
			)
			return result, nil
		}
	}

	return nil, fmt.Errorf("all parsing attempts failed for response")
}

// Helper methods

func (urp *UnifiedResponseParser) cleanResponse(ctx context.Context, rawResponse string) (string, []string, error) {
	cleaned := rawResponse
	warnings := make([]string, 0)

	for _, cleaner := range urp.cleaners {
		var err error
		cleaned, err = cleaner.Clean(ctx, cleaned)
		if err != nil {
			warning := fmt.Sprintf("Cleaner failed: %v", err)
			warnings = append(warnings, warning)
			urp.logger.Warn("Response cleaner failed", zap.Error(err))
		}
	}

	return cleaned, warnings, nil
}

func (urp *UnifiedResponseParser) calculateParseConfidence(original, extracted string, warnings []string) float64 {
	baseConfidence := 0.8

	// Reduce confidence for warnings
	warningPenalty := float64(len(warnings)) * 0.1
	baseConfidence -= warningPenalty

	// Increase confidence if extraction is clean
	if len(extracted) > 0 && isValidJSON(extracted) {
		baseConfidence += 0.1
	}

	// Reduce confidence for very short extractions
	if len(extracted) < 50 {
		baseConfidence -= 0.2
	}

	// Clamp to valid range
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	if baseConfidence < 0.0 {
		baseConfidence = 0.0
	}

	return baseConfidence
}

func (urp *UnifiedResponseParser) initializeDefaultComponents() {
	// Initialize default extractors
	urp.extractors[ResponseTypeJSON] = &JSONExtractor{}
	urp.extractors[ResponseTypeValidation] = &ValidationExtractor{}
	urp.extractors[ResponseTypeAnalysis] = &AnalysisExtractor{}
	urp.extractors[ResponseTypeStructured] = &StructuredExtractor{}
	urp.extractors[ResponseTypeText] = &TextExtractor{}

	// Initialize default validators
	urp.validators[ResponseTypeValidation] = &ValidationResponseValidator{}
	urp.validators[ResponseTypeAnalysis] = &AnalysisResponseValidator{}
	urp.validators[ResponseTypeJSON] = &JSONResponseValidator{}
	urp.validators[ResponseTypeStructured] = &StructuredResponseValidator{}
	urp.validators[ResponseTypeText] = &TextResponseValidator{}

	// Initialize default cleaners
	urp.cleaners = append(urp.cleaners,
		&MarkdownCodeBlockCleaner{},
		&WhitespaceCleaner{},
		&InvalidCharacterCleaner{},
	)
}

// Helper function
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// Common response types

// ValidationResponse represents validation-specific LLM responses
type ValidationResponse struct {
	OverallScore     int                    `json:"overall_score" validate:"min=0,max=100"`
	SecurityScore    int                    `json:"security_score" validate:"min=0,max=100"`
	QualityScore     int                    `json:"quality_score" validate:"min=0,max=100"`
	PerformanceScore int                    `json:"performance_score" validate:"min=0,max=100"`
	Confidence       float64                `json:"confidence" validate:"min=0,max=1"`
	Issues           []ValidationIssue      `json:"issues"`
	Recommendations  []ValidationRecommendation `json:"recommendations"`
	Summary          string                 `json:"summary"`
	Timestamp        string                 `json:"timestamp"`
}

// AnalysisResponse represents analysis-specific LLM responses  
type AnalysisResponse struct {
	AnalysisType     string                 `json:"analysis_type"`
	Findings         []AnalysisFinding      `json:"findings"`
	Confidence       float64                `json:"confidence" validate:"min=0,max=1"`
	Summary          string                 `json:"summary"`
	Recommendations  []AnalysisRecommendation `json:"recommendations"`
	Metadata         map[string]interface{} `json:"metadata"`
	Timestamp        string                 `json:"timestamp"`
}

// GenericResponse for unstructured responses
type GenericResponse struct {
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp string                 `json:"timestamp"`
}

// Supporting types
type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Location    string `json:"location,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type ValidationRecommendation struct {
	Priority    string `json:"priority"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Effort      string `json:"effort"`
}

type AnalysisFinding struct {
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type AnalysisRecommendation struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Impact      string `json:"impact"`
}