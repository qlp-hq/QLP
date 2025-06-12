package factories

import (
	"fmt"

	"QLP/internal/cloud"
	"QLP/internal/cloud/azure"
)

// AzureProviderFactory creates Azure cloud managers
type AzureProviderFactory struct{}

// NewAzureProviderFactory creates a new Azure provider factory
func NewAzureProviderFactory() *AzureProviderFactory {
	return &AzureProviderFactory{}
}

// CreateProvider creates a cloud manager for the specified provider
func (f *AzureProviderFactory) CreateProvider(provider cloud.CloudProvider, config cloud.ProviderConfig) (cloud.CloudManager, error) {
	switch provider {
	case cloud.CloudProviderAzure:
		return f.createAzureProvider(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (f *AzureProviderFactory) createAzureProvider(config cloud.ProviderConfig) (cloud.CloudManager, error) {
	// Extract Azure-specific configuration
	azureConfig := &azure.Config{
		Environment: "production",
	}

	if subscriptionID, ok := config.Config["subscription_id"].(string); ok {
		azureConfig.SubscriptionID = subscriptionID
	} else {
		return nil, fmt.Errorf("Azure subscription_id is required")
	}

	if resourceGroup, ok := config.Config["resource_group"].(string); ok {
		azureConfig.ResourceGroup = resourceGroup
	}

	if location, ok := config.Config["location"].(string); ok {
		azureConfig.Location = location
	}

	if tenantID, ok := config.Config["tenant_id"].(string); ok {
		azureConfig.TenantID = tenantID
	}

	// Add tags from config
	if len(config.Config) > 0 {
		if azureConfig.Tags == nil {
			azureConfig.Tags = make(map[string]string)
		}
		for k, v := range config.Config {
			if str, ok := v.(string); ok && k != "subscription_id" && k != "resource_group" && k != "location" && k != "tenant_id" {
				azureConfig.Tags[k] = str
			}
		}
	}

	return azure.NewAzureManager(azureConfig)
}