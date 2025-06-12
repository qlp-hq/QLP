package cloud

import (
	"context"
	"time"

	"QLP/internal/tenancy"
)

// CloudProvider represents different cloud providers
type CloudProvider string

const (
	CloudProviderAzure CloudProvider = "azure"
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
)

// ResourceType represents different types of cloud resources
type ResourceType string

// String returns the string representation of ResourceType
func (rt ResourceType) String() string {
	return string(rt)
}

const (
	ResourceTypeCompute    ResourceType = "compute"
	ResourceTypeStorage    ResourceType = "storage"
	ResourceTypeDatabase   ResourceType = "database"
	ResourceTypeNetwork    ResourceType = "network"
	ResourceTypeContainer  ResourceType = "container"
	ResourceTypeKubernetes ResourceType = "kubernetes"
)

// CloudManager provides unified interface for multi-cloud operations
type CloudManager interface {
	// Provider management
	GetProvider() CloudProvider
	IsAvailable(ctx context.Context) bool
	GetRegions(ctx context.Context) ([]Region, error)
	
	// Resource management
	CreateResource(ctx context.Context, req *CreateResourceRequest) (*Resource, error)
	GetResource(ctx context.Context, resourceID string) (*Resource, error)
	UpdateResource(ctx context.Context, resourceID string, req *UpdateResourceRequest) (*Resource, error)
	DeleteResource(ctx context.Context, resourceID string) error
	ListResources(ctx context.Context, filters ResourceFilters) ([]*Resource, error)
	
	// Tenant-aware operations
	ProvisionTenantResources(ctx context.Context, tenantCtx *tenancy.TenantContext, req *TenantProvisionRequest) (*TenantProvisionResult, error)
	DeprovisionTenantResources(ctx context.Context, tenantID string) error
	GetTenantResources(ctx context.Context, tenantID string) ([]*Resource, error)
	
	// Monitoring and metrics
	GetResourceMetrics(ctx context.Context, resourceID string, timeRange TimeRange) (*ResourceMetrics, error)
	GetTenantUsage(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantUsage, error)
	
	// Cost management
	GetResourceCosts(ctx context.Context, resourceID string, timeRange TimeRange) (*ResourceCosts, error)
	GetTenantCosts(ctx context.Context, tenantID string, timeRange TimeRange) (*TenantCosts, error)
	
	// Security and compliance
	ApplySecurityPolicies(ctx context.Context, resourceID string, policies []SecurityPolicy) error
	GetComplianceStatus(ctx context.Context, resourceID string) (*ComplianceStatus, error)
}

// StorageManager provides cloud storage operations
type StorageManager interface {
	CreateBucket(ctx context.Context, req *CreateBucketRequest) (*Bucket, error)
	DeleteBucket(ctx context.Context, bucketName string) error
	UploadFile(ctx context.Context, bucketName, key string, data []byte) error
	DownloadFile(ctx context.Context, bucketName, key string) ([]byte, error)
	DeleteFile(ctx context.Context, bucketName, key string) error
	ListFiles(ctx context.Context, bucketName, prefix string) ([]FileInfo, error)
	GetSignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)
}

// ComputeManager provides cloud compute operations
type ComputeManager interface {
	CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*Instance, error)
	StartInstance(ctx context.Context, instanceID string) error
	StopInstance(ctx context.Context, instanceID string) error
	RestartInstance(ctx context.Context, instanceID string) error
	DeleteInstance(ctx context.Context, instanceID string) error
	GetInstance(ctx context.Context, instanceID string) (*Instance, error)
	ListInstances(ctx context.Context, filters InstanceFilters) ([]*Instance, error)
	ResizeInstance(ctx context.Context, instanceID string, newSize string) error
}

// DatabaseManager provides cloud database operations
type DatabaseManager interface {
	CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*Database, error)
	DeleteDatabase(ctx context.Context, databaseID string) error
	GetDatabase(ctx context.Context, databaseID string) (*Database, error)
	ListDatabases(ctx context.Context, filters DatabaseFilters) ([]*Database, error)
	CreateBackup(ctx context.Context, databaseID string, req *BackupRequest) (*Backup, error)
	RestoreBackup(ctx context.Context, backupID string, req *RestoreRequest) (*Database, error)
	ScaleDatabase(ctx context.Context, databaseID string, req *ScaleRequest) error
}

// ContainerManager provides container orchestration operations
type ContainerManager interface {
	CreateCluster(ctx context.Context, req *CreateClusterRequest) (*Cluster, error)
	DeleteCluster(ctx context.Context, clusterID string) error
	GetCluster(ctx context.Context, clusterID string) (*Cluster, error)
	ScaleCluster(ctx context.Context, clusterID string, nodeCount int) error
	DeployApplication(ctx context.Context, clusterID string, req *DeploymentRequest) (*Deployment, error)
	DeleteDeployment(ctx context.Context, clusterID, deploymentID string) error
	GetDeploymentStatus(ctx context.Context, clusterID, deploymentID string) (*DeploymentStatus, error)
}

// NetworkManager provides network operations
type NetworkManager interface {
	CreateVirtualNetwork(ctx context.Context, req *CreateVNetRequest) (*VirtualNetwork, error)
	DeleteVirtualNetwork(ctx context.Context, vnetID string) error
	CreateSubnet(ctx context.Context, vnetID string, req *CreateSubnetRequest) (*Subnet, error)
	DeleteSubnet(ctx context.Context, subnetID string) error
	CreateSecurityGroup(ctx context.Context, req *CreateSecurityGroupRequest) (*SecurityGroup, error)
	UpdateSecurityRules(ctx context.Context, sgID string, rules []SecurityRule) error
	CreateLoadBalancer(ctx context.Context, req *CreateLoadBalancerRequest) (*LoadBalancer, error)
}

// MultiCloudManager coordinates operations across multiple cloud providers
type MultiCloudManager interface {
	GetProviders() []CloudProvider
	GetPrimaryProvider() CloudProvider
	SetPrimaryProvider(provider CloudProvider) error
	GetManager(provider CloudProvider) (CloudManager, error)
	
	// Cross-cloud operations
	MigrateResource(ctx context.Context, resourceID string, targetProvider CloudProvider) (*Resource, error)
	ReplicateResource(ctx context.Context, resourceID string, targetProviders []CloudProvider) ([]*Resource, error)
	SyncTenantAcrossProviders(ctx context.Context, tenantID string) error
	
	// Cost optimization
	OptimizeTenantCosts(ctx context.Context, tenantID string) (*CostOptimizationPlan, error)
	GetCrossCloudUsage(ctx context.Context, tenantID string, timeRange TimeRange) (*CrossCloudUsage, error)
}

// CloudEvent represents cloud infrastructure events
type CloudEvent struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"`
	Provider     CloudProvider     `json:"provider"`
	ResourceID   string            `json:"resource_id"`
	ResourceType ResourceType      `json:"resource_type"`
	TenantID     string            `json:"tenant_id,omitempty"`
	Action       string            `json:"action"`
	Status       string            `json:"status"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	Timestamp    time.Time         `json:"timestamp"`
	Source       string            `json:"source"`
}

// CloudEventHandler processes cloud events
type CloudEventHandler interface {
	HandleEvent(ctx context.Context, event *CloudEvent) error
	Subscribe(eventTypes []string) error
	Unsubscribe(eventTypes []string) error
}