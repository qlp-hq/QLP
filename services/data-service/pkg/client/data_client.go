package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"QLP/services/data-service/pkg/contracts"
)

// DataClient provides a client interface for the data service
type DataClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewDataClient creates a new data service client
func NewDataClient(baseURL string) *DataClient {
	return &DataClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateIntent creates a new intent for a tenant
func (dc *DataClient) CreateIntent(ctx context.Context, tenantID string, req *contracts.CreateIntentRequest) (*contracts.Intent, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/intents", dc.baseURL, tenantID)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := dc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response contracts.CreateIntentResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response.Intent, nil
}

// GetIntent retrieves an intent by ID for a tenant
func (dc *DataClient) GetIntent(ctx context.Context, tenantID, intentID string) (*contracts.Intent, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/intents/%s", dc.baseURL, tenantID, intentID)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := dc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("intent not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var intent contracts.Intent
	if err := json.NewDecoder(resp.Body).Decode(&intent); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &intent, nil
}

// ListIntents retrieves intents for a tenant with optional filtering
func (dc *DataClient) ListIntents(ctx context.Context, tenantID string, req *contracts.ListIntentsRequest) (*contracts.ListIntentsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/intents", dc.baseURL, tenantID)
	
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := httpReq.URL.Query()
	if req.Status != "" {
		q.Add("status", req.Status)
	}
	if req.Limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", req.Limit))
	}
	if req.Offset > 0 {
		q.Add("offset", fmt.Sprintf("%d", req.Offset))
	}
	httpReq.URL.RawQuery = q.Encode()

	resp, err := dc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response contracts.ListIntentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// UpdateIntent updates an existing intent
func (dc *DataClient) UpdateIntent(ctx context.Context, tenantID, intentID string, req *contracts.UpdateIntentRequest) (*contracts.Intent, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/intents/%s", dc.baseURL, tenantID, intentID)
	
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := dc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("intent not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var intent contracts.Intent
	if err := json.NewDecoder(resp.Body).Decode(&intent); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &intent, nil
}

// DeleteIntent deletes an intent
func (dc *DataClient) DeleteIntent(ctx context.Context, tenantID, intentID string) error {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/intents/%s", dc.baseURL, tenantID, intentID)
	
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := dc.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("intent not found")
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}