package cloud

import (
	"time"
)

// Region represents a cloud provider region
type Region struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     CloudProvider     `json:"provider"`
	Location     string            `json:"location"`
	Available    bool              `json:"available"`
	Capabilities []string          `json:"capabilities"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Resource represents a cloud resource
type Resource struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         ResourceType      `json:"type"`
	Provider     CloudProvider     `json:"provider"`
	Region       string            `json:"region"`
	TenantID     string            `json:"tenant_id"`
	Status       ResourceStatus    `json:"status"`
	Configuration map[string]interface{} `json:"configuration"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ResourceStatus string

const (
	ResourceStatusCreating ResourceStatus = "creating"
	ResourceStatusRunning  ResourceStatus = "running"
	ResourceStatusStopped  ResourceStatus = "stopped"
	ResourceStatusDeleting ResourceStatus = "deleting"
	ResourceStatusError    ResourceStatus = "error"
	ResourceStatusUnknown  ResourceStatus = "unknown"
)

// CreateResourceRequest for creating resources
type CreateResourceRequest struct {
	Name          string                 `json:"name"`
	Type          ResourceType           `json:"type"`
	Region        string                 `json:"region"`
	TenantID      string                 `json:"tenant_id"`
	Configuration map[string]interface{} `json:"configuration"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// UpdateResourceRequest for updating resources
type UpdateResourceRequest struct {
	Name          string                 `json:"name,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// ResourceFilters for filtering resources
type ResourceFilters struct {
	Type     ResourceType `json:"type,omitempty"`
	TenantID string       `json:"tenant_id,omitempty"`
	Region   string       `json:"region,omitempty"`
	Status   ResourceStatus `json:"status,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Limit    int          `json:"limit,omitempty"`
	Offset   int          `json:"offset,omitempty"`
}

// Tenant provisioning models
type TenantProvisionRequest struct {
	TenantID     string                 `json:"tenant_id"`
	Model        string                 `json:"model"` // pooled, silo, bridge
	Tier         string                 `json:"tier"`  // free, standard, premium, enterprise
	Region       string                 `json:"region"`
	Requirements TenantRequirements     `json:"requirements"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

type TenantRequirements struct {
	Compute    ComputeRequirements `json:"compute"`
	Storage    StorageRequirements `json:"storage"`
	Database   DatabaseRequirements `json:"database"`
	Network    NetworkRequirements `json:"network"`
	Security   SecurityRequirements `json:"security"`
}

type ComputeRequirements struct {
	CPUCores     int    `json:"cpu_cores"`
	MemoryGB     int    `json:"memory_gb"`
	InstanceType string `json:"instance_type,omitempty"`
	Dedicated    bool   `json:"dedicated"`
}

type StorageRequirements struct {
	SizeGB      int    `json:"size_gb"`
	Type        string `json:"type"` // standard, premium, ultra
	Encryption  bool   `json:"encryption"`
	Backup      bool   `json:"backup"`
	Replication string `json:"replication,omitempty"`
}

type DatabaseRequirements struct {
	Engine      string `json:"engine"` // postgresql, mysql, mongodb
	Version     string `json:"version,omitempty"`
	SizeGB      int    `json:"size_gb"`
	Dedicated   bool   `json:"dedicated"`
	HighAvailability bool `json:"high_availability"`
	BackupRetentionDays int `json:"backup_retention_days"`
}

type NetworkRequirements struct {
	VirtualNetwork bool     `json:"virtual_network"`
	Subnets        []string `json:"subnets,omitempty"`
	LoadBalancer   bool     `json:"load_balancer"`
	Firewall       bool     `json:"firewall"`
	DNS            bool     `json:"dns"`
}

type SecurityRequirements struct {
	EncryptionAtRest    bool     `json:"encryption_at_rest"`
	EncryptionInTransit bool     `json:"encryption_in_transit"`
	KeyManagement       string   `json:"key_management,omitempty"`
	Compliance          []string `json:"compliance,omitempty"`
	NetworkIsolation    bool     `json:"network_isolation"`
	AuditLogging        bool     `json:"audit_logging"`
}

type TenantProvisionResult struct {
	TenantID    string     `json:"tenant_id"`
	Status      string     `json:"status"`
	Resources   []*Resource `json:"resources"`
	Endpoints   map[string]string `json:"endpoints"`
	Credentials map[string]string `json:"credentials,omitempty"`
	Message     string     `json:"message,omitempty"`
	ProvisionedAt time.Time `json:"provisioned_at"`
}

// Storage models
type Bucket struct {
	Name         string            `json:"name"`
	Provider     CloudProvider     `json:"provider"`
	Region       string            `json:"region"`
	TenantID     string            `json:"tenant_id"`
	Encryption   bool              `json:"encryption"`
	Versioning   bool              `json:"versioning"`
	PublicAccess bool              `json:"public_access"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type CreateBucketRequest struct {
	Name         string            `json:"name"`
	Region       string            `json:"region"`
	TenantID     string            `json:"tenant_id"`
	Encryption   bool              `json:"encryption"`
	Versioning   bool              `json:"versioning"`
	PublicAccess bool              `json:"public_access"`
	Tags         map[string]string `json:"tags,omitempty"`
}

type FileInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ContentType  string    `json:"content_type"`
	ETag         string    `json:"etag,omitempty"`
}

// Compute models
type Instance struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     CloudProvider     `json:"provider"`
	Region       string            `json:"region"`
	TenantID     string            `json:"tenant_id"`
	InstanceType string            `json:"instance_type"`
	Status       InstanceStatus    `json:"status"`
	PublicIP     string            `json:"public_ip,omitempty"`
	PrivateIP    string            `json:"private_ip,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type InstanceStatus string

const (
	InstanceStatusPending    InstanceStatus = "pending"
	InstanceStatusRunning    InstanceStatus = "running"
	InstanceStatusStopping   InstanceStatus = "stopping"
	InstanceStatusStopped    InstanceStatus = "stopped"
	InstanceStatusTerminating InstanceStatus = "terminating"
	InstanceStatusTerminated InstanceStatus = "terminated"
)

type CreateInstanceRequest struct {
	Name         string            `json:"name"`
	InstanceType string            `json:"instance_type"`
	Region       string            `json:"region"`
	TenantID     string            `json:"tenant_id"`
	ImageID      string            `json:"image_id"`
	SubnetID     string            `json:"subnet_id,omitempty"`
	SecurityGroups []string        `json:"security_groups,omitempty"`
	UserData     string            `json:"user_data,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

type InstanceFilters struct {
	TenantID     string            `json:"tenant_id,omitempty"`
	Region       string            `json:"region,omitempty"`
	InstanceType string            `json:"instance_type,omitempty"`
	Status       InstanceStatus    `json:"status,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	Limit        int               `json:"limit,omitempty"`
	Offset       int               `json:"offset,omitempty"`
}

// Database models
type Database struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Provider        CloudProvider     `json:"provider"`
	Region          string            `json:"region"`
	TenantID        string            `json:"tenant_id"`
	Engine          string            `json:"engine"`
	Version         string            `json:"version"`
	Status          DatabaseStatus    `json:"status"`
	Endpoint        string            `json:"endpoint"`
	Port            int               `json:"port"`
	SizeGB          int               `json:"size_gb"`
	HighAvailability bool             `json:"high_availability"`
	Tags            map[string]string `json:"tags,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type DatabaseStatus string

const (
	DatabaseStatusCreating   DatabaseStatus = "creating"
	DatabaseStatusAvailable  DatabaseStatus = "available"
	DatabaseStatusModifying  DatabaseStatus = "modifying"
	DatabaseStatusDeleting   DatabaseStatus = "deleting"
	DatabaseStatusFailed     DatabaseStatus = "failed"
)

type CreateDatabaseRequest struct {
	Name                string            `json:"name"`
	Engine              string            `json:"engine"`
	Version             string            `json:"version,omitempty"`
	Region              string            `json:"region"`
	TenantID            string            `json:"tenant_id"`
	SizeGB              int               `json:"size_gb"`
	InstanceClass       string            `json:"instance_class"`
	Username            string            `json:"username"`
	Password            string            `json:"password"`
	HighAvailability    bool              `json:"high_availability"`
	BackupRetentionDays int               `json:"backup_retention_days"`
	SubnetGroupName     string            `json:"subnet_group_name,omitempty"`
	SecurityGroups      []string          `json:"security_groups,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
}

type DatabaseFilters struct {
	TenantID string            `json:"tenant_id,omitempty"`
	Region   string            `json:"region,omitempty"`
	Engine   string            `json:"engine,omitempty"`
	Status   DatabaseStatus    `json:"status,omitempty"`
	Tags     map[string]string `json:"tags,omitempty"`
	Limit    int               `json:"limit,omitempty"`
	Offset   int               `json:"offset,omitempty"`
}

// Backup models
type Backup struct {
	ID           string            `json:"id"`
	DatabaseID   string            `json:"database_id"`
	Name         string            `json:"name"`
	Status       BackupStatus      `json:"status"`
	SizeGB       int               `json:"size_gb"`
	Type         BackupType        `json:"type"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiresAt    *time.Time        `json:"expires_at,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

type BackupStatus string

const (
	BackupStatusCreating  BackupStatus = "creating"
	BackupStatusAvailable BackupStatus = "available"
	BackupStatusDeleting  BackupStatus = "deleting"
	BackupStatusFailed    BackupStatus = "failed"
)

type BackupType string

const (
	BackupTypeManual    BackupType = "manual"
	BackupTypeAutomatic BackupType = "automatic"
)

type BackupRequest struct {
	Name string            `json:"name"`
	Type BackupType        `json:"type"`
	Tags map[string]string `json:"tags,omitempty"`
}

type RestoreRequest struct {
	DatabaseName   string `json:"database_name"`
	InstanceClass  string `json:"instance_class,omitempty"`
	Region         string `json:"region,omitempty"`
	SubnetGroupName string `json:"subnet_group_name,omitempty"`
}

type ScaleRequest struct {
	InstanceClass string `json:"instance_class,omitempty"`
	SizeGB        int    `json:"size_gb,omitempty"`
}

// Time range for metrics and costs
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Metrics models
type ResourceMetrics struct {
	ResourceID string                   `json:"resource_id"`
	TimeRange  TimeRange                `json:"time_range"`
	Metrics    map[string][]MetricPoint `json:"metrics"`
}

type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit,omitempty"`
}

type TenantUsage struct {
	TenantID    string                      `json:"tenant_id"`
	TimeRange   TimeRange                   `json:"time_range"`
	Resources   map[ResourceType]UsageStats `json:"resources"`
	TotalCost   float64                     `json:"total_cost"`
	Currency    string                      `json:"currency"`
}

type UsageStats struct {
	ResourceCount int     `json:"resource_count"`
	UsageHours    float64 `json:"usage_hours"`
	Cost          float64 `json:"cost"`
	Unit          string  `json:"unit"`
}

// Cost models
type ResourceCosts struct {
	ResourceID string      `json:"resource_id"`
	TimeRange  TimeRange   `json:"time_range"`
	TotalCost  float64     `json:"total_cost"`
	Currency   string      `json:"currency"`
	Breakdown  []CostItem  `json:"breakdown"`
}

type TenantCosts struct {
	TenantID    string                      `json:"tenant_id"`
	TimeRange   TimeRange                   `json:"time_range"`
	TotalCost   float64                     `json:"total_cost"`
	Currency    string                      `json:"currency"`
	ByProvider  map[CloudProvider]float64   `json:"by_provider"`
	ByResource  map[ResourceType]float64    `json:"by_resource"`
	Breakdown   []CostItem                  `json:"breakdown"`
	Projections *CostProjection            `json:"projections,omitempty"`
}

type CostItem struct {
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Unit        string    `json:"unit"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Timestamp   time.Time `json:"timestamp"`
}

type CostProjection struct {
	Daily   float64 `json:"daily"`
	Weekly  float64 `json:"weekly"`
	Monthly float64 `json:"monthly"`
	Yearly  float64 `json:"yearly"`
}

type CostOptimizationPlan struct {
	TenantID        string                    `json:"tenant_id"`
	CurrentCost     float64                   `json:"current_cost"`
	OptimizedCost   float64                   `json:"optimized_cost"`
	PotentialSavings float64                  `json:"potential_savings"`
	Recommendations []OptimizationRecommendation `json:"recommendations"`
	GeneratedAt     time.Time                 `json:"generated_at"`
}

type OptimizationRecommendation struct {
	Type        string  `json:"type"`
	ResourceID  string  `json:"resource_id"`
	Action      string  `json:"action"`
	Description string  `json:"description"`
	Savings     float64 `json:"savings"`
	Impact      string  `json:"impact"`
	Confidence  float64 `json:"confidence"`
}

type CrossCloudUsage struct {
	TenantID  string                         `json:"tenant_id"`
	TimeRange TimeRange                      `json:"time_range"`
	ByProvider map[CloudProvider]TenantUsage `json:"by_provider"`
	TotalCost float64                        `json:"total_cost"`
	Currency  string                         `json:"currency"`
}

// Security models
type SecurityPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Rules       []SecurityRule         `json:"rules"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

type SecurityRule struct {
	Protocol    string `json:"protocol"`
	Port        string `json:"port"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Action      string `json:"action"`
	Priority    int    `json:"priority"`
}

type ComplianceStatus struct {
	ResourceID    string                    `json:"resource_id"`
	OverallStatus ComplianceStatusLevel    `json:"overall_status"`
	Standards     []ComplianceStandardResult `json:"standards"`
	LastChecked   time.Time                 `json:"last_checked"`
	NextCheck     time.Time                 `json:"next_check"`
}

type ComplianceStatusLevel string

const (
	ComplianceStatusCompliant    ComplianceStatusLevel = "compliant"
	ComplianceStatusNonCompliant ComplianceStatusLevel = "non_compliant"
	ComplianceStatusUnknown      ComplianceStatusLevel = "unknown"
)

type ComplianceStandardResult struct {
	Standard    string                `json:"standard"`
	Version     string                `json:"version"`
	Status      ComplianceStatusLevel `json:"status"`
	Score       float64               `json:"score"`
	Issues      []ComplianceIssue     `json:"issues,omitempty"`
	LastChecked time.Time             `json:"last_checked"`
}

type ComplianceIssue struct {
	ID          string `json:"id"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
	Reference   string `json:"reference,omitempty"`
}

// Container and Kubernetes models
type Cluster struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Provider        CloudProvider     `json:"provider"`
	Region          string            `json:"region"`
	TenantID        string            `json:"tenant_id"`
	Status          ClusterStatus     `json:"status"`
	NodeCount       int               `json:"node_count"`
	KubernetesVersion string          `json:"kubernetes_version"`
	Endpoint        string            `json:"endpoint"`
	Tags            map[string]string `json:"tags,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type ClusterStatus string

const (
	ClusterStatusProvisioning ClusterStatus = "provisioning"
	ClusterStatusRunning     ClusterStatus = "running"
	ClusterStatusStopping    ClusterStatus = "stopping"
	ClusterStatusStopped     ClusterStatus = "stopped"
	ClusterStatusDeleting    ClusterStatus = "deleting"
	ClusterStatusFailed      ClusterStatus = "failed"
)

type CreateClusterRequest struct {
	Name              string            `json:"name"`
	Region            string            `json:"region"`
	TenantID          string            `json:"tenant_id"`
	NodeCount         int               `json:"node_count"`
	NodeInstanceType  string            `json:"node_instance_type"`
	KubernetesVersion string            `json:"kubernetes_version,omitempty"`
	SubnetIDs         []string          `json:"subnet_ids,omitempty"`
	SecurityGroups    []string          `json:"security_groups,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

type Deployment struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ClusterID   string            `json:"cluster_id"`
	TenantID    string            `json:"tenant_id"`
	Status      DeploymentStatus  `json:"status"`
	Image       string            `json:"image"`
	Replicas    int               `json:"replicas"`
	Ports       []int             `json:"ports,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusStopped   DeploymentStatus = "stopped"
)

type DeploymentRequest struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Replicas    int               `json:"replicas"`
	Ports       []int             `json:"ports,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Resources   ResourceRequests  `json:"resources,omitempty"`
}

type ResourceRequests struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// Network models
type VirtualNetwork struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Provider     CloudProvider     `json:"provider"`
	Region       string            `json:"region"`
	TenantID     string            `json:"tenant_id"`
	CIDR         string            `json:"cidr"`
	Status       NetworkStatus     `json:"status"`
	Subnets      []Subnet          `json:"subnets,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

type NetworkStatus string

const (
	NetworkStatusCreating  NetworkStatus = "creating"
	NetworkStatusAvailable NetworkStatus = "available"
	NetworkStatusDeleting  NetworkStatus = "deleting"
	NetworkStatusFailed    NetworkStatus = "failed"
)

type CreateVNetRequest struct {
	Name     string            `json:"name"`
	Region   string            `json:"region"`
	TenantID string            `json:"tenant_id"`
	CIDR     string            `json:"cidr"`
	Tags     map[string]string `json:"tags,omitempty"`
}

type Subnet struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	VNetID           string            `json:"vnet_id"`
	CIDR             string            `json:"cidr"`
	AvailabilityZone string            `json:"availability_zone,omitempty"`
	Status           NetworkStatus     `json:"status"`
	Tags             map[string]string `json:"tags,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}

type CreateSubnetRequest struct {
	Name             string            `json:"name"`
	CIDR             string            `json:"cidr"`
	AvailabilityZone string            `json:"availability_zone,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
}

type SecurityGroup struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	VNetID      string            `json:"vnet_id,omitempty"`
	TenantID    string            `json:"tenant_id"`
	Rules       []SecurityRule    `json:"rules"`
	Tags        map[string]string `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type CreateSecurityGroupRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	VNetID      string            `json:"vnet_id,omitempty"`
	TenantID    string            `json:"tenant_id"`
	Rules       []SecurityRule    `json:"rules,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type LoadBalancer struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    CloudProvider     `json:"provider"`
	Region      string            `json:"region"`
	TenantID    string            `json:"tenant_id"`
	Type        string            `json:"type"` // internal, external
	DNSName     string            `json:"dns_name"`
	Status      NetworkStatus     `json:"status"`
	Listeners   []LoadBalancerListener `json:"listeners"`
	Tags        map[string]string `json:"tags,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type LoadBalancerListener struct {
	Port       int    `json:"port"`
	Protocol   string `json:"protocol"`
	TargetPort int    `json:"target_port"`
	TargetGroup string `json:"target_group,omitempty"`
}

type CreateLoadBalancerRequest struct {
	Name        string                   `json:"name"`
	Type        string                   `json:"type"`
	Region      string                   `json:"region"`
	TenantID    string                   `json:"tenant_id"`
	SubnetIDs   []string                 `json:"subnet_ids"`
	Listeners   []LoadBalancerListener   `json:"listeners"`
	Tags        map[string]string        `json:"tags,omitempty"`
}