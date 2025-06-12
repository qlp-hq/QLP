# QLP Azure Deployment Plan - Production Ready

## 🏗️ **Azure Architecture Overview**

```
┌─────────────────────────────────────────────────────────────┐
│                    Azure Front Door                         │
│                  (Global CDN + WAF)                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│              Azure Container Instances                     │
│              (Multi-Container Groups)                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              QLP Application                        │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │         Go Binary (main.go)                 │    │   │
│  │  │  • Intent Processing                        │    │   │
│  │  │  • Task Orchestration                       │    │   │
│  │  │  • WebSocket Endpoints                      │    │   │
│  │  │  • REST API                                 │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  │  ┌─────────────────────────────────────────────┐    │   │
│  │  │       Docker-in-Docker Sandbox              │    │   │
│  │  │  • Container Execution                      │    │   │
│  │  │  • Resource Isolation                       │    │   │
│  │  │  • Security Constraints                     │    │   │
│  │  └─────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│        Azure Database for PostgreSQL (Flexible)            │
│                   with pgvector                            │
│  • Intent Storage                                          │
│  • Task Execution History                                  │
│  • Vector Embeddings (1536 dimensions)                    │
│  • User Sessions                                           │
│  • Audit Logs                                              │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Azure Key Vault                             │
│  • OpenAI API Keys                                         │
│  • Database Credentials                                    │
│  • JWT Secrets                                             │
│  • SSL Certificates                                        │
│  • Container Registry Credentials                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│               Supporting Services                           │
├─────────────────────────────────────────────────────────────┤
│  Azure Container Registry  │  Application Insights         │
│  • QLP Docker Images       │  • Performance Monitoring     │
│  • Vulnerability Scanning  │  • Error Tracking             │
│  • Multi-arch Support      │  • Custom Metrics             │
├─────────────────────────────────────────────────────────────┤
│  Azure Storage Account     │  Azure Log Analytics           │
│  • QuantumCapsule Files    │  • Centralized Logging        │
│  • Static Assets           │  • Query & Analysis            │
│  • Backup Storage          │  • Alerting Rules              │
│  • Container Volumes       │  • Workbooks & Dashboards     │
└─────────────────────────────────────────────────────────────┘
```

## 📁 **Project Structure**

```
QLP/
├── infrastructure/
│   ├── terraform/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── provider.tf
│   │   ├── locals.tf
│   │   ├── modules/
│   │   │   ├── container_instances/
│   │   │   ├── database/
│   │   │   ├── key_vault/
│   │   │   ├── container_registry/
│   │   │   ├── monitoring/
│   │   │   ├── networking/
│   │   │   └── storage/
│   │   └── environments/
│   │       ├── dev.tfvars
│   │       ├── staging.tfvars
│   │       └── prod.tfvars
├── deployment/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── .dockerignore
│   ├── azure/
│   │   ├── container-group.yaml
│   │   ├── app-config.yaml
│   │   └── secrets.yaml
│   └── scripts/
│       ├── build.sh
│       ├── deploy.sh
│       ├── health-check.sh
│       ├── backup.sh
│       └── rollback.sh
├── .github/
│   └── workflows/
│       ├── ci-cd.yml
│       ├── infrastructure.yml
│       ├── security-scan.yml
│       └── backup.yml
├── config/
│   ├── production.env
│   ├── staging.env
│   ├── azure-config.go
│   └── database/
│       ├── schema.sql
│       ├── migrations/
│       └── seed-data.sql
└── docs/
    ├── AZURE_DEPLOYMENT_PLAN.md
    ├── ARCHITECTURE.md
    ├── SECURITY.md
    └── OPERATIONS.md
```

## 🎯 **Key Architecture Decisions**

### **Why Azure Container Instances over App Service?**

Based on your QLP requirements for Docker-in-Docker capabilities:

✅ **Container Instances Advantages:**
- Native Docker-in-Docker support
- Full container control and privileged access
- Custom networking and storage configurations
- No restrictions on container runtimes
- Better isolation for sandbox execution

❌ **App Service Limitations:**
- Restricted Docker-in-Docker capabilities
- Limited privileged container access
- Less control over container runtime
- Potential security restrictions for QLP's sandbox needs

### **PostgreSQL with pgvector Configuration**

Enhanced database setup for vector operations:

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create tables with vector support
CREATE TABLE intents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    content TEXT NOT NULL,
    embedding vector(1536), -- OpenAI ada-002 embedding size
    -- ... other fields
);

-- Create vector index for similarity search
CREATE INDEX idx_intents_embedding 
ON intents USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

## 🏗️ **Production Terraform Infrastructure**

### **Main Infrastructure (infrastructure/terraform/main.tf)**

```hcl
terraform {
  required_version = ">= 1.6.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.80.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~>3.6.0"
    }
  }
  
  backend "azurerm" {
    resource_group_name  = "qlp-terraform-state"
    storage_account_name = "qlpterraformstate"
    container_name      = "tfstate"
    key                 = "qlp.terraform.tfstate"
  }
}

provider "azurerm" {
  features {
    key_vault {
      purge_soft_delete_on_destroy    = true
      recover_soft_deleted_key_vaults = true
    }
  }
}

# Resource Group
resource "azurerm_resource_group" "main" {
  name     = "${var.environment}-qlp-rg"
  location = var.location
  tags     = local.common_tags
}

# Container Registry
resource "azurerm_container_registry" "main" {
  name                = "${var.environment}qlpacr"
  resource_group_name = azurerm_resource_group.main.name
  location           = azurerm_resource_group.main.location
  sku                = var.acr_sku
  admin_enabled      = true
  
  trust_policy {
    enabled = true
  }
  
  tags = local.common_tags
}

# Virtual Network
resource "azurerm_virtual_network" "main" {
  name                = "${var.environment}-qlp-vnet"
  address_space       = var.vnet_address_space
  location           = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags               = local.common_tags
}

# Subnets
resource "azurerm_subnet" "container" {
  name                 = "${var.environment}-qlp-container-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = var.subnet_address_spaces.container_subnet

  delegation {
    name = "container_delegation"
    service_delegation {
      name = "Microsoft.ContainerInstance/containerGroups"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
        "Microsoft.Network/virtualNetworks/subnets/prepareNetworkPolicies/action"
      ]
    }
  }
}

resource "azurerm_subnet" "database" {
  name                 = "${var.environment}-qlp-db-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = var.subnet_address_spaces.database_subnet

  delegation {
    name = "postgres_delegation"
    service_delegation {
      name = "Microsoft.DBforPostgreSQL/flexibleServers"
      actions = [
        "Microsoft.Network/virtualNetworks/subnets/join/action",
      ]
    }
  }
}

# Key Vault
resource "azurerm_key_vault" "main" {
  name                = "${var.environment}-qlp-kv-${random_string.suffix.result}"
  location           = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id          = data.azurerm_client_config.current.tenant_id
  sku_name           = "standard"
  
  enabled_for_disk_encryption     = true
  enabled_for_deployment          = true
  enabled_for_template_deployment = true
  purge_protection_enabled       = var.environment == "prod"
  
  network_acls {
    default_action = var.environment == "prod" ? "Deny" : "Allow"
    bypass         = "AzureServices"
    virtual_network_subnet_ids = [azurerm_subnet.container.id]
  }
  
  tags = local.common_tags
}

# PostgreSQL Flexible Server
resource "azurerm_postgresql_flexible_server" "main" {
  name                   = "${var.environment}-qlp-postgres"
  resource_group_name    = azurerm_resource_group.main.name
  location              = azurerm_resource_group.main.location
  version               = var.db_version
  delegated_subnet_id   = azurerm_subnet.database.id
  private_dns_zone_id   = azurerm_private_dns_zone.postgres.id
  administrator_login    = var.db_admin_username
  administrator_password = random_password.db_password.result
  zone                  = "1"
  
  storage_mb = var.db_storage_mb
  sku_name   = var.db_sku_name
  
  backup_retention_days        = var.db_backup_retention_days
  geo_redundant_backup_enabled = var.environment == "prod"
  
  dynamic "high_availability" {
    for_each = var.environment == "prod" ? [1] : []
    content {
      mode                      = "ZoneRedundant"
      standby_availability_zone = "2"
    }
  }
  
  maintenance_window {
    day_of_week  = 0
    start_hour   = 3
    start_minute = 0
  }
  
  tags = local.common_tags
  depends_on = [azurerm_private_dns_zone_virtual_network_link.postgres]
}

# PostgreSQL Configuration for pgvector
resource "azurerm_postgresql_flexible_server_configuration" "shared_preload_libraries" {
  name      = "shared_preload_libraries"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = "vector"
}

resource "azurerm_postgresql_flexible_server_configuration" "max_connections" {
  name      = "max_connections"
  server_id = azurerm_postgresql_flexible_server.main.id
  value     = "200"
}

# Database
resource "azurerm_postgresql_flexible_server_database" "main" {
  name      = "qlp_db"
  server_id = azurerm_postgresql_flexible_server.main.id
  collation = "en_US.utf8"
  charset   = "utf8"
}

# Container Instance
resource "azurerm_container_group" "main" {
  name                = "${var.environment}-qlp-containers"
  location           = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  ip_address_type    = "Private"
  subnet_ids         = [azurerm_subnet.container.id]
  os_type            = "Linux"
  restart_policy     = "Always"
  
  identity {
    type = "SystemAssigned"
  }

  # Main