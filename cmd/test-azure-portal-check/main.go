package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"QLP/internal/deployment/azure"
	"QLP/internal/logger"
	"go.uber.org/zap"
)

// Test that creates Azure resources and waits for manual verification
func main() {
	ctx := context.Background()
	
	// Initialize logger
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger.Logger = zapLogger
	
	agentLogger := logger.GetDefaultLogger().WithComponent("azure_portal_check_test")
	agentLogger.Info("🔍 Starting Azure Portal Verification Test")
	
	// Get Azure config
	azureConfig, err := getAzureConfig()
	if err != nil {
		agentLogger.Error("Failed to get Azure configuration", zap.Error(err))
		os.Exit(1)
	}
	
	fmt.Println("🔥 QUANTUMLAYER AZURE PORTAL VERIFICATION TEST")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📍 Subscription: %s\n", maskValue(azureConfig.SubscriptionID))
	fmt.Printf("🌍 Region: %s\n", azureConfig.Location)
	fmt.Printf("🏷️  Tenant: %s\n", maskValue(azureConfig.TenantID))
	fmt.Println()
	
	// Create unique resource group name
	resourceGroupName := fmt.Sprintf("quantumlayer-test-%d", time.Now().Unix())
	
	fmt.Printf("🚀 Creating Resource Group: %s\n", resourceGroupName)
	fmt.Println("⏳ This will create a REAL Azure resource group...")
	
	// Confirm before creating
	fmt.Print("Continue? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response != "y" && response != "yes" {
		fmt.Println("❌ Test cancelled by user")
		return
	}
	
	// Create resource group
	agentLogger.Info("Creating resource group",
		zap.String("name", resourceGroupName),
		zap.String("location", azureConfig.Location),
	)
	
	if err := createResourceGroupWithDetails(ctx, resourceGroupName, azureConfig.Location, agentLogger); err != nil {
		agentLogger.Error("Failed to create resource group", zap.Error(err))
		os.Exit(1)
	}
	
	// Show resource group details
	fmt.Println("\n✅ RESOURCE GROUP CREATED SUCCESSFULLY!")
	fmt.Println(strings.Repeat("=", 60))
	
	if err := showResourceGroupDetails(ctx, resourceGroupName, agentLogger); err != nil {
		agentLogger.Warn("Failed to get resource group details", zap.Error(err))
	}
	
	// Wait for manual verification
	fmt.Println("\n🔍 MANUAL VERIFICATION INSTRUCTIONS:")
	fmt.Println("1. Open the Azure Portal: https://portal.azure.com")
	fmt.Println("2. Navigate to Resource Groups")
	fmt.Printf("3. Look for resource group: %s\n", resourceGroupName)
	fmt.Println("4. Verify it exists and shows the tags")
	fmt.Println()
	fmt.Printf("🌐 Direct link: https://portal.azure.com/#@%s/resource/subscriptions/%s/resourceGroups/%s/overview\n", 
		azureConfig.TenantID, azureConfig.SubscriptionID, resourceGroupName)
	fmt.Println()
	
	// Wait for user confirmation
	fmt.Println("⏰ Resource group will remain active for verification...")
	fmt.Print("Press ENTER when you've verified the resource group in the portal: ")
	reader.ReadString('\n')
	
	// List all QuantumLayer resource groups
	fmt.Println("\n📋 LISTING ALL QUANTUMLAYER RESOURCE GROUPS:")
	if err := listQuantumLayerResourceGroups(ctx, agentLogger); err != nil {
		agentLogger.Warn("Failed to list resource groups", zap.Error(err))
	}
	
	// Cleanup confirmation
	fmt.Printf("\n🧹 Ready to delete resource group: %s\n", resourceGroupName)
	fmt.Print("Delete the resource group? (Y/n): ")
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response == "n" || response == "no" {
		fmt.Printf("⚠️  Resource group %s left active for manual cleanup\n", resourceGroupName)
		fmt.Printf("To delete later: az group delete --name %s --yes\n", resourceGroupName)
		return
	}
	
	// Delete resource group
	agentLogger.Info("Deleting resource group", zap.String("name", resourceGroupName))
	
	if err := deleteResourceGroup(ctx, resourceGroupName, agentLogger); err != nil {
		agentLogger.Error("Failed to delete resource group", zap.Error(err))
		fmt.Printf("❌ Failed to delete resource group. Delete manually: az group delete --name %s --yes\n", resourceGroupName)
		os.Exit(1)
	}
	
	fmt.Println("\n🎉 AZURE PORTAL VERIFICATION TEST COMPLETED!")
	fmt.Println("✅ Resource group created and verified")
	fmt.Println("✅ Resource group deleted successfully")
	fmt.Println("🔥 QuantumLayer Azure integration is LIVE!")
}

func createResourceGroupWithDetails(ctx context.Context, name, location string, logger logger.Interface) error {
	logger.Info("Creating resource group with detailed tracking",
		zap.String("name", name),
		zap.String("location", location),
	)
	
	// Create with comprehensive tags
	cmd := exec.CommandContext(ctx, "az", "group", "create",
		"--name", name,
		"--location", location,
		"--tags",
		"created-by=quantumlayer",
		"purpose=portal-verification-test",
		"test-mode=true",
		fmt.Sprintf("created-at=%s", time.Now().Format(time.RFC3339)),
		fmt.Sprintf("expires-at=%s", time.Now().Add(1*time.Hour).Format(time.RFC3339)),
		"--output", "json")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create resource group: %w, output: %s", err, string(output))
	}
	
	// Parse and display creation details
	var rgInfo map[string]interface{}
	if err := json.Unmarshal(output, &rgInfo); err == nil {
		fmt.Printf("📍 Resource Group ID: %v\n", rgInfo["id"])
		fmt.Printf("📍 Location: %v\n", rgInfo["location"])
		fmt.Printf("📍 Provisioning State: %v\n", rgInfo["properties"].(map[string]interface{})["provisioningState"])
		
		if tags, ok := rgInfo["tags"].(map[string]interface{}); ok {
			fmt.Println("🏷️  Tags:")
			for key, value := range tags {
				fmt.Printf("   %s: %v\n", key, value)
			}
		}
	}
	
	logger.Info("Resource group created successfully",
		zap.String("name", name),
		zap.Int("output_bytes", len(output)),
	)
	
	return nil
}

func showResourceGroupDetails(ctx context.Context, name string, logger logger.Interface) error {
	cmd := exec.CommandContext(ctx, "az", "group", "show",
		"--name", name,
		"--output", "json")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get resource group details: %w", err)
	}
	
	var rgInfo map[string]interface{}
	if err := json.Unmarshal(output, &rgInfo); err != nil {
		return fmt.Errorf("failed to parse resource group info: %w", err)
	}
	
	fmt.Println("📊 RESOURCE GROUP DETAILS:")
	fmt.Printf("   Name: %v\n", rgInfo["name"])
	fmt.Printf("   ID: %v\n", rgInfo["id"])
	fmt.Printf("   Location: %v\n", rgInfo["location"])
	fmt.Printf("   Provisioning State: %v\n", 
		rgInfo["properties"].(map[string]interface{})["provisioningState"])
	
	return nil
}

func listQuantumLayerResourceGroups(ctx context.Context, logger logger.Interface) error {
	cmd := exec.CommandContext(ctx, "az", "group", "list",
		"--tag", "created-by=quantumlayer",
		"--output", "table")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to list resource groups: %w", err)
	}
	
	fmt.Println(string(output))
	return nil
}

func deleteResourceGroup(ctx context.Context, name string, logger logger.Interface) error {
	logger.Info("Deleting resource group", zap.String("name", name))
	
	// Delete with progress tracking
	cmd := exec.CommandContext(ctx, "az", "group", "delete",
		"--name", name,
		"--yes",
		"--verbose")
	
	fmt.Printf("🗑️  Deleting resource group %s...\n", name)
	fmt.Println("⏳ This may take a few minutes...")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete resource group: %w, output: %s", err, string(output))
	}
	
	logger.Info("Resource group deleted successfully", zap.String("name", name))
	fmt.Printf("✅ Resource group %s deleted successfully\n", name)
	
	return nil
}

func getAzureConfig() (azure.ClientConfig, error) {
	// Try environment variables first
	if subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID"); subscriptionID != "" {
		return azure.ClientConfig{
			SubscriptionID: subscriptionID,
			Location:       getEnvOrDefault("AZURE_LOCATION", "uksouth"),
			TenantID:       os.Getenv("AZURE_TENANT_ID"),
		}, nil
	}
	
	// Fall back to Azure CLI
	cmd := exec.Command("az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return azure.ClientConfig{}, fmt.Errorf("failed to get Azure account info: %w", err)
	}
	
	var accountInfo struct {
		ID       string `json:"id"`
		TenantID string `json:"tenantId"`
	}
	
	if err := json.Unmarshal(output, &accountInfo); err != nil {
		return azure.ClientConfig{}, fmt.Errorf("failed to parse Azure account info: %w", err)
	}
	
	return azure.ClientConfig{
		SubscriptionID: accountInfo.ID,
		Location:       getEnvOrDefault("AZURE_LOCATION", "uksouth"),
		TenantID:       accountInfo.TenantID,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskValue(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}