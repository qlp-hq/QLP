package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"QLP/services/orchestrator-service/pkg/contracts"
)

// OrchestratorClient provides a client for the orchestrator service
type OrchestratorClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOrchestratorClient creates a new orchestrator service client
func NewOrchestratorClient(baseURL string) *OrchestratorClient {
	return &OrchestratorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for workflow operations
		},
	}
}

// ExecuteWorkflow starts execution of a new workflow
func (oc *OrchestratorClient) ExecuteWorkflow(ctx context.Context, tenantID string, req *contracts.ExecuteWorkflowRequest) (*contracts.ExecuteWorkflowResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows", oc.baseURL, tenantID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response contracts.ExecuteWorkflowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetWorkflow retrieves a workflow execution by ID
func (oc *OrchestratorClient) GetWorkflow(ctx context.Context, tenantID, workflowID string) (*contracts.GetWorkflowResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows/%s", oc.baseURL, tenantID, workflowID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response contracts.GetWorkflowResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ListWorkflows lists workflows for a tenant with pagination
func (oc *OrchestratorClient) ListWorkflows(ctx context.Context, tenantID string, page, pageSize int, status string) (*contracts.ListWorkflowsResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows", oc.baseURL, tenantID)

	// Add query parameters
	params := url.Values{}
	if page > 0 {
		params.Add("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		params.Add("page_size", strconv.Itoa(pageSize))
	}
	if status != "" {
		params.Add("status", status)
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response contracts.ListWorkflowsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// PauseWorkflow pauses a running workflow
func (oc *OrchestratorClient) PauseWorkflow(ctx context.Context, tenantID, workflowID string, req *contracts.PauseWorkflowRequest) error {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows/%s/pause", oc.baseURL, tenantID, workflowID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ResumeWorkflow resumes a paused workflow
func (oc *OrchestratorClient) ResumeWorkflow(ctx context.Context, tenantID, workflowID string, req *contracts.ResumeWorkflowRequest) error {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows/%s/resume", oc.baseURL, tenantID, workflowID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// CancelWorkflow cancels a workflow execution
func (oc *OrchestratorClient) CancelWorkflow(ctx context.Context, tenantID, workflowID string, req *contracts.CancelWorkflowRequest) error {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows/%s/cancel", oc.baseURL, tenantID, workflowID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RetryTask retries a failed task
func (oc *OrchestratorClient) RetryTask(ctx context.Context, tenantID, workflowID string, req *contracts.RetryTaskRequest) error {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows/%s/retry", oc.baseURL, tenantID, workflowID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ValidateDAG validates a DAG structure
func (oc *OrchestratorClient) ValidateDAG(ctx context.Context, req *contracts.DAGValidationRequest) (*contracts.DAGValidationResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/dag/validate", oc.baseURL)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response contracts.DAGValidationResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetWorkflowMetrics retrieves metrics for a workflow
func (oc *OrchestratorClient) GetWorkflowMetrics(ctx context.Context, tenantID, workflowID string) (*contracts.WorkflowMetrics, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/workflows/%s/metrics", oc.baseURL, tenantID, workflowID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := oc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response contracts.WorkflowMetrics
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}