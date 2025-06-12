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

	"QLP/services/packaging-service/pkg/contracts"
)

// PackagingClient provides a client for the packaging service
type PackagingClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPackagingClient creates a new packaging service client
func NewPackagingClient(baseURL string) *PackagingClient {
	return &PackagingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateCapsule creates a new QL capsule
func (pc *PackagingClient) CreateCapsule(ctx context.Context, tenantID string, req *contracts.CreateCapsuleRequest) (*contracts.CreateCapsuleResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/capsules", pc.baseURL, tenantID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := pc.httpClient.Do(httpReq)
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

	var response contracts.CreateCapsuleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// CreateQuantumDrops creates quantum drops for an intent
func (pc *PackagingClient) CreateQuantumDrops(ctx context.Context, tenantID string, req *contracts.CreateQuantumDropRequest) (*contracts.CreateQuantumDropResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/quantum-drops", pc.baseURL, tenantID)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := pc.httpClient.Do(httpReq)
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

	var response contracts.CreateQuantumDropResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// GetCapsule retrieves a capsule by ID
func (pc *PackagingClient) GetCapsule(ctx context.Context, tenantID, capsuleID string) (*contracts.GetCapsuleResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/capsules/%s", pc.baseURL, tenantID, capsuleID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pc.httpClient.Do(httpReq)
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

	var response contracts.GetCapsuleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// ListCapsules lists capsules for a tenant
func (pc *PackagingClient) ListCapsules(ctx context.Context, tenantID string, page, pageSize int) (*contracts.ListCapsulesResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/capsules", pc.baseURL, tenantID)

	// Add query parameters
	params := url.Values{}
	if page > 0 {
		params.Add("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		params.Add("page_size", strconv.Itoa(pageSize))
	}

	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pc.httpClient.Do(httpReq)
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

	var response contracts.ListCapsulesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// DownloadCapsule downloads a capsule as a binary file
func (pc *PackagingClient) DownloadCapsule(ctx context.Context, tenantID, capsuleID string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/capsules/%s/download", pc.baseURL, tenantID, capsuleID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pc.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// GetQuantumDrop retrieves a quantum drop by ID
func (pc *PackagingClient) GetQuantumDrop(ctx context.Context, tenantID, dropID string) (*contracts.QuantumDrop, error) {
	endpoint := fmt.Sprintf("%s/api/v1/tenants/%s/quantum-drops/%s", pc.baseURL, tenantID, dropID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pc.httpClient.Do(httpReq)
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

	var response contracts.QuantumDrop
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}