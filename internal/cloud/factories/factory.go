package factories

import (
	"fmt"

	"QLP/internal/cloud"
)

// DefaultProviderFactory provides a unified factory for all cloud providers
type DefaultProviderFactory struct {
	azureFactory *AzureProviderFactory
}

// NewDefaultProviderFactory creates a new default provider factory
func NewDefaultProviderFactory() *DefaultProviderFactory {
	return &DefaultProviderFactory{
		azureFactory: NewAzureProviderFactory(),
	}
}

// CreateProvider creates a cloud manager for the specified provider
func (f *DefaultProviderFactory) CreateProvider(provider cloud.CloudProvider, config cloud.ProviderConfig) (cloud.CloudManager, error) {
	switch provider {
	case cloud.CloudProviderAzure:
		return f.azureFactory.CreateProvider(provider, config)
	case cloud.CloudProviderAWS:
		return nil, fmt.Errorf("AWS provider not yet implemented")
	case cloud.CloudProviderGCP:
		return nil, fmt.Errorf("GCP provider not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}