#!/bin/bash

# Azure Deployment Validation Test Setup Script
# This script helps configure Azure authentication for testing QuantumLayer deployment validation

set -e

echo "🚀 QuantumLayer Azure Deployment Validation Setup"
echo "================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Azure CLI is installed
print_status "Checking Azure CLI installation..."
if ! command -v az &> /dev/null; then
    print_error "Azure CLI is not installed. Please install it first:"
    echo "  macOS: brew install azure-cli"
    echo "  Ubuntu: curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash"
    echo "  Windows: winget install Microsoft.AzureCLI"
    exit 1
else
    print_success "Azure CLI is installed"
fi

# Check if user is logged in
print_status "Checking Azure CLI authentication..."
if ! az account show &> /dev/null; then
    print_warning "Not logged in to Azure CLI. Please login:"
    az login
    if [ $? -ne 0 ]; then
        print_error "Azure login failed"
        exit 1
    fi
else
    print_success "Already logged in to Azure CLI"
fi

# Get current subscription info
SUBSCRIPTION_INFO=$(az account show --output json)
SUBSCRIPTION_ID=$(echo $SUBSCRIPTION_INFO | jq -r '.id')
SUBSCRIPTION_NAME=$(echo $SUBSCRIPTION_INFO | jq -r '.name')
TENANT_ID=$(echo $SUBSCRIPTION_INFO | jq -r '.tenantId')

print_success "Current Azure subscription:"
echo "  Name: $SUBSCRIPTION_NAME"
echo "  ID: $SUBSCRIPTION_ID"
echo "  Tenant: $TENANT_ID"

# Ask user if they want to use current subscription
echo ""
read -p "Use this subscription for testing? (y/n): " -r
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_status "Available subscriptions:"
    az account list --output table
    echo ""
    read -p "Enter subscription ID to use: " -r SUBSCRIPTION_ID
    az account set --subscription "$SUBSCRIPTION_ID"
    print_success "Switched to subscription: $SUBSCRIPTION_ID"
fi

# Set default location
DEFAULT_LOCATION="westeurope"
echo ""
read -p "Enter Azure region for testing (default: $DEFAULT_LOCATION): " -r LOCATION
LOCATION=${LOCATION:-$DEFAULT_LOCATION}

# Check if resource provider is registered
print_status "Checking required resource providers..."
REQUIRED_PROVIDERS=("Microsoft.ContainerInstance" "Microsoft.Resources")

for provider in "${REQUIRED_PROVIDERS[@]}"; do
    STATUS=$(az provider show --namespace "$provider" --query "registrationState" -o tsv 2>/dev/null || echo "NotFound")
    if [ "$STATUS" != "Registered" ]; then
        print_warning "Registering provider: $provider"
        az provider register --namespace "$provider"
    else
        print_success "Provider registered: $provider"
    fi
done

# Create .env file for testing
ENV_FILE=".env.azure-test"
print_status "Creating environment file: $ENV_FILE"

cat > "$ENV_FILE" << EOF
# Azure Configuration for QuantumLayer Deployment Validation Testing
# Generated on $(date)

# Required Azure settings
AZURE_SUBSCRIPTION_ID=$SUBSCRIPTION_ID
AZURE_TENANT_ID=$TENANT_ID
AZURE_LOCATION=$LOCATION

# Optional settings (uncomment and set if using service principal)
# AZURE_CLIENT_ID=your-client-id
# AZURE_CLIENT_SECRET=your-client-secret

# Test configuration
AZURE_TEST_COST_LIMIT=5.00
AZURE_TEST_TTL_MINUTES=30
AZURE_TEST_RESOURCE_GROUP_PREFIX=qlp-test
EOF

print_success "Environment file created: $ENV_FILE"

# Create service principal (optional)
echo ""
read -p "Create a service principal for testing? (y/n): " -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    SP_NAME="qlp-deployment-test-$(date +%s)"
    print_status "Creating service principal: $SP_NAME"
    
    SP_INFO=$(az ad sp create-for-rbac \
        --name "$SP_NAME" \
        --role "Contributor" \
        --scopes "/subscriptions/$SUBSCRIPTION_ID" \
        --output json)
    
    CLIENT_ID=$(echo $SP_INFO | jq -r '.appId')
    CLIENT_SECRET=$(echo $SP_INFO | jq -r '.password')
    
    print_success "Service principal created:"
    echo "  Client ID: $CLIENT_ID"
    echo "  Client Secret: [hidden]"
    
    # Update .env file with service principal
    echo "" >> "$ENV_FILE"
    echo "# Service Principal (created $(date))" >> "$ENV_FILE"
    echo "AZURE_CLIENT_ID=$CLIENT_ID" >> "$ENV_FILE"
    echo "AZURE_CLIENT_SECRET=$CLIENT_SECRET" >> "$ENV_FILE"
    
    print_warning "Service principal credentials added to $ENV_FILE"
    print_warning "Keep these credentials secure and delete the service principal after testing!"
fi

# Test resource group creation
echo ""
read -p "Test resource group creation? (y/n): " -r
if [[ $REPLY =~ ^[Yy]$ ]]; then
    TEST_RG="qlp-test-setup-$(date +%s)"
    print_status "Testing resource group creation: $TEST_RG"
    
    az group create --name "$TEST_RG" --location "$LOCATION" --tags "purpose=qlp-test" "auto-delete=true" > /dev/null
    
    if [ $? -eq 0 ]; then
        print_success "Resource group created successfully"
        
        # Clean up test resource group
        print_status "Cleaning up test resource group..."
        az group delete --name "$TEST_RG" --yes --no-wait > /dev/null
        print_success "Test resource group cleanup initiated"
    else
        print_error "Failed to create test resource group"
        print_error "Please check your permissions and try again"
        exit 1
    fi
fi

echo ""
print_success "Azure setup completed successfully!"
echo ""
echo "📋 Next Steps:"
echo "1. Source the environment file: source $ENV_FILE"
echo "2. Run the deployment test: go run cmd/test-azure-deployment/main.go"
echo ""
echo "📖 Environment Variables Set:"
echo "  AZURE_SUBSCRIPTION_ID: $SUBSCRIPTION_ID"
echo "  AZURE_TENANT_ID: $TENANT_ID"
echo "  AZURE_LOCATION: $LOCATION"
echo ""
echo "🧹 Cleanup:"
echo "  - Delete service principal: az ad sp delete --id <client-id>"
echo "  - Remove environment file: rm $ENV_FILE"
echo ""
print_warning "Remember to clean up any test resources to avoid charges!"