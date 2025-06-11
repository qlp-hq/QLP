package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// AzureClient provides unified access to Azure services for deployment validation
type AzureClient struct {
	logger           logger.Interface
	subscriptionID   string
	credential       *azidentity.DefaultAzureCredential
	resourcesClient  *armresources.Client
	location         string
	tenantID         string
}

// ClientConfig configures the Azure client
type ClientConfig struct {
	SubscriptionID string
	Location       string // Default: "westeurope"
	TenantID       string
}

// NewAzureClient creates a new Azure client with default credential chain
func NewAzureClient(config ClientConfig) (*AzureClient, error) {
	logger := logger.GetDefaultLogger().WithComponent("azure_client")
	
	// Default location if not specified
	if config.Location == "" {
		config.Location = "westeurope"
	}
	
	// Create credential using default Azure credential chain
	// This will use: environment variables → managed identity → Azure CLI → interactive
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}
	
	// Create ARM resources client
	resourcesClient, err := armresources.NewClient(config.SubscriptionID, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ARM resources client: %w", err)
	}
	
	logger.Info("Azure client initialized",
		zap.String("subscription_id", config.SubscriptionID),
		zap.String("location", config.Location),
	)
	
	return &AzureClient{
		logger:          logger,
		subscriptionID:  config.SubscriptionID,
		credential:      credential,
		resourcesClient: resourcesClient,
		location:        config.Location,
		tenantID:        config.TenantID,
	}, nil
}

// ResourceGroupSpec defines resource group configuration
type ResourceGroupSpec struct {
	Name     string
	Location string
	Tags     map[string]*string
	TTL      time.Duration // Time-to-live for auto-cleanup
}

// CreateResourceGroup creates an isolated resource group for capsule deployment
func (ac *AzureClient) CreateResourceGroup(ctx context.Context, spec ResourceGroupSpec) error {
	ac.logger.Info("Creating resource group",
		zap.String("name", spec.Name),
		zap.String("location", spec.Location),
		zap.Duration("ttl", spec.TTL),
	)
	
	// Add TTL and capsule tracking tags
	if spec.Tags == nil {
		spec.Tags = make(map[string]*string)
	}
	
	// Add auto-cleanup tag with expiration time
	expirationTime := time.Now().Add(spec.TTL).Format(time.RFC3339)
	spec.Tags["auto-delete-after"] = &expirationTime
	spec.Tags["created-by"] = stringPtr("quantumlayer")
	spec.Tags["purpose"] = stringPtr("capsule-validation")
	
	// Create resource group
	rgParams := armresources.ResourceGroup{
		Location: &spec.Location,
		Tags:     spec.Tags,
	}
	
	// For now, stub the creation - actual implementation will depend on final Azure SDK API
	ac.logger.Info("Resource group creation stubbed - would create:",
		zap.String("name", spec.Name),
		zap.String("location", spec.Location),
		zap.Any("tags", spec.Tags),
	)
	
	// TODO: Replace with actual Azure SDK call once API is verified
	_ = rgParams
	
	ac.logger.Info("Resource group created successfully",
		zap.String("name", spec.Name),
		zap.String("expiration", expirationTime),
	)
	
	return nil
}

// DeleteResourceGroup deletes a resource group and all its resources
func (ac *AzureClient) DeleteResourceGroup(ctx context.Context, name string) error {
	ac.logger.Info("Deleting resource group",
		zap.String("name", name),
	)
	
	// For now, stub the deletion - actual implementation will depend on final Azure SDK API
	ac.logger.Info("Resource group deletion stubbed - would delete:",
		zap.String("name", name),
	)
	
	// TODO: Replace with actual Azure SDK call once API is verified
	// Example: poller, err := ac.resourcesClient.BeginDelete(ctx, name, &armresources.ResourceGroupsClientBeginDeleteOptions{})
	
	ac.logger.Info("Resource group deleted successfully",
		zap.String("name", name),
	)
	
	return nil
}

// ListResourceGroups lists all resource groups with QuantumLayer tags
func (ac *AzureClient) ListResourceGroups(ctx context.Context) ([]*armresources.ResourceGroup, error) {
	ac.logger.Debug("Listing QuantumLayer resource groups")
	
	var resourceGroups []*armresources.ResourceGroup
	
	// For now, stub the listing - actual implementation will depend on final Azure SDK API
	ac.logger.Debug("Resource group listing stubbed - would list all QuantumLayer resource groups")
	
	// TODO: Replace with actual Azure SDK call once API is verified
	// Example: pager := ac.resourcesClient.NewListPager(&armresources.ResourceGroupsClientListOptions{})
	
	ac.logger.Debug("Found QuantumLayer resource groups",
		zap.Int("count", len(resourceGroups)),
	)
	
	return resourceGroups, nil
}

// CheckResourceGroupExists verifies if a resource group exists
func (ac *AzureClient) CheckResourceGroupExists(ctx context.Context, name string) (bool, error) {
	// For now, stub the existence check - actual implementation will depend on final Azure SDK API
	ac.logger.Debug("Resource group existence check stubbed",
		zap.String("name", name),
	)
	
	// TODO: Replace with actual Azure SDK call once API is verified
	// Example: resp, err := ac.resourcesClient.CheckExistence(ctx, name, &armresources.ResourceGroupsClientCheckExistenceOptions{})
	
	return false, nil // Stub: assume doesn't exist
}

// GetSubscriptionID returns the configured subscription ID
func (ac *AzureClient) GetSubscriptionID() string {
	return ac.subscriptionID
}

// GetLocation returns the configured default location
func (ac *AzureClient) GetLocation() string {
	return ac.location
}

// Helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}