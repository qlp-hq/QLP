# 📖 QLP Deployment Playbook

## 🎯 Overview

This playbook provides step-by-step instructions for deploying QuantumLayer (QLP) to Azure in production, including all procedures, scripts, and safety checks needed for reliable deployments.

---

## 📋 Quick Reference

### **Deployment Types**
- **🟢 Standard Deployment**: Regular feature releases
- **🔵 Hotfix Deployment**: Critical bug fixes
- **🟡 Database Migration**: Schema changes
- **🔴 Emergency Deployment**: Security or critical issues

### **Deployment Windows**
```yaml
Standard Deployments:
- Preferred: Tuesday/Wednesday 10:00-14:00 UTC
- Avoid: Fridays, Mondays, Holiday weeks
- Duration: 30-60 minutes planned window

Emergency Deployments:
- Available: 24/7 with proper authorization
- Duration: 15-30 minutes critical window
```

---

## 🚀 Standard Deployment Process

### **Phase 1: Pre-Deployment (T-24 hours)**

#### **1.1 Planning & Communication**
```bash
# Send deployment notification
./scripts/notify-deployment.sh \
    --type "standard" \
    --version "v1.2.3" \
    --window "2024-01-15 10:00 UTC" \
    --duration "45 minutes"

# Expected recipients:
# - Engineering team
# - Customer success
# - Leadership (for major releases)
```

#### **1.2 Code Preparation**
```bash
# Verify release branch
git checkout release/v1.2.3
git pull origin release/v1.2.3

# Run comprehensive tests
make test-all
make security-scan
make performance-test

# Build and tag release
git tag v1.2.3
git push origin v1.2.3
```

#### **1.3 Infrastructure Validation**
```bash
# Validate current infrastructure state
./scripts/validate-infrastructure.sh

# Check resource capacity
./scripts/check-resource-capacity.sh

# Verify backup status
./scripts/verify-backups.sh --age-limit 6h
```

### **Phase 2: Pre-Deployment (T-2 hours)**

#### **2.1 Final Validation**
```bash
#!/bin/bash
# pre-deployment-final-check.sh

echo "🔍 Final pre-deployment validation..."

# 1. Verify system health
HEALTH_STATUS=$(curl -s -w "%{http_code}" $PRODUCTION_URL/health -o /dev/null)
if [ $HEALTH_STATUS -ne 200 ]; then
    echo "❌ Production health check failed: $HEALTH_STATUS"
    exit 1
fi

# 2. Check recent error rates
ERROR_RATE=$(az monitor metrics list \
    --resource $APP_INSIGHTS_RESOURCE \
    --metric "requests/failed" \
    --interval PT1H \
    --query 'value[0].timeseries[0].data[-1].total // 0')

if [ $ERROR_RATE -gt 10 ]; then
    echo "⚠️ High error rate detected: $ERROR_RATE errors/hour"
    echo "Consider postponing deployment"
fi

# 3. Verify CI/CD pipeline
PIPELINE_STATUS=$(gh run list --workflow=ci-cd.yml --limit=1 --json conclusion -q '.[0].conclusion')
if [ "$PIPELINE_STATUS" != "success" ]; then
    echo "❌ Latest CI/CD run failed: $PIPELINE_STATUS"
    exit 1
fi

# 4. Check database performance
DB_SLOW_QUERIES=$(az monitor metrics list \
    --resource $DB_RESOURCE_ID \
    --metric "slow_queries" \
    --interval PT1H \
    --query 'value[0].timeseries[0].data[-1].average // 0')

if [ $DB_SLOW_QUERIES -gt 5 ]; then
    echo "⚠️ High slow query count: $DB_SLOW_QUERIES"
fi

echo "✅ All pre-deployment checks passed!"
```

#### **2.2 Staging Deployment Test**
```bash
# Deploy to staging environment
./scripts/deploy-staging.sh v1.2.3

# Run staging smoke tests
./scripts/staging-smoke-tests.sh

# Performance comparison test
./scripts/compare-performance.sh \
    --baseline-env production \
    --test-env staging \
    --duration 300s
```

### **Phase 3: Production Deployment (T-0)**

#### **3.1 Deployment Execution**
```bash
#!/bin/bash
# production-deployment.sh

set -e
VERSION=${1:-latest}
DEPLOYMENT_ID="deploy-$(date +%Y%m%d-%H%M%S)"

echo "🚀 Starting production deployment: $VERSION"
echo "📝 Deployment ID: $DEPLOYMENT_ID"

# Initialize deployment tracking
./scripts/track-deployment.sh start $DEPLOYMENT_ID $VERSION

# 1. Create deployment snapshot
echo "📸 Creating pre-deployment snapshot..."
SNAPSHOT_NAME="pre-deploy-$DEPLOYMENT_ID"
./scripts/create-snapshot.sh $SNAPSHOT_NAME

# 2. Enable maintenance mode (if required)
if [ "$MAINTENANCE_MODE" = "true" ]; then
    echo "🚧 Enabling maintenance mode..."
    ./scripts/maintenance-mode.sh enable
fi

# 3. Deploy new version
echo "🐳 Deploying container version: $VERSION"
./scripts/blue-green-deploy.sh $VERSION

# 4. Database migrations (if any)
if [ -f "migrations/v${VERSION}.sql" ]; then
    echo "🗄️ Running database migrations..."
    ./scripts/run-migrations.sh $VERSION
fi

# 5. Health check verification
echo "🏥 Verifying deployment health..."
./scripts/health-check-extended.sh --timeout 300

# 6. Disable maintenance mode
if [ "$MAINTENANCE_MODE" = "true" ]; then
    echo "🟢 Disabling maintenance mode..."
    ./scripts/maintenance-mode.sh disable
fi

# 7. Post-deployment tests
echo "🧪 Running post-deployment tests..."
./scripts/post-deployment-tests.sh

# 8. Update deployment tracking
./scripts/track-deployment.sh complete $DEPLOYMENT_ID success

echo "🎉 Deployment completed successfully!"
echo "📊 Version $VERSION is now live"
```

#### **3.2 Blue-Green Deployment Script**
```bash
#!/bin/bash
# blue-green-deploy.sh

VERSION=${1:-latest}
TIMEOUT=${2:-600}

echo "🔵🟢 Starting blue-green deployment for version: $VERSION"

# Determine current and target slots
CURRENT_SLOT=$(./scripts/get-current-slot.sh)
if [ "$CURRENT_SLOT" = "blue" ]; then
    TARGET_SLOT="green"
else
    TARGET_SLOT="blue"
fi

echo "📍 Current slot: $CURRENT_SLOT"
echo "🎯 Target slot: $TARGET_SLOT"

# Deploy to target slot
echo "📦 Deploying to $TARGET_SLOT slot..."
az container create \
    --resource-group $RESOURCE_GROUP \
    --name "qlp-${TARGET_SLOT}" \
    --image "${ACR_NAME}.azurecr.io/qlp:${VERSION}" \
    --cpu 2 \
    --memory 8 \
    --restart-policy Always \
    --environment-variables \
        QLP_MODE=production \
        SLOT_NAME=$TARGET_SLOT \
        DATABASE_URL="@Microsoft.KeyVault(SecretUri=${KV_DB_SECRET})" \
        AZURE_OPENAI_API_KEY="@Microsoft.KeyVault(SecretUri=${KV_OPENAI_SECRET})" \
    --ports 8080 \
    --subnet $CONTAINER_SUBNET_ID

# Wait for target slot to be ready
echo "⏳ Waiting for $TARGET_SLOT slot to become ready..."
end_time=$(($(date +%s) + $TIMEOUT))

while [ $(date +%s) -lt $end_time ]; do
    CONTAINER_STATE=$(az container show \
        --resource-group $RESOURCE_GROUP \
        --name "qlp-${TARGET_SLOT}" \
        --query 'containers[0].instanceView.currentState.state' -o tsv)
    
    if [ "$CONTAINER_STATE" = "Running" ]; then
        # Get container IP
        TARGET_IP=$(az container show \
            --resource-group $RESOURCE_GROUP \
            --name "qlp-${TARGET_SLOT}" \
            --query 'ipAddress.ip' -o tsv)
        
        # Test health endpoint
        if curl -f --max-time 10 "http://${TARGET_IP}:8080/health" >/dev/null 2>&1; then
            echo "✅ $TARGET_SLOT slot is healthy!"
            break
        fi
    fi
    
    echo "⏳ $TARGET_SLOT slot status: $CONTAINER_STATE, waiting..."
    sleep 10
done

# Verify deployment didn't timeout
if [ $(date +%s) -ge $end_time ]; then
    echo "❌ Deployment timeout! $TARGET_SLOT slot failed to become ready"
    # Cleanup failed deployment
    az container delete \
        --resource-group $RESOURCE_GROUP \
        --name "qlp-${TARGET_SLOT}" \
        --yes
    exit 1
fi

# Run smoke tests on target slot
echo "🧪 Running smoke tests on $TARGET_SLOT slot..."
./scripts/smoke-tests.sh "http://${TARGET_IP}:8080"

# Switch traffic to target slot
echo "🔄 Switching traffic to $TARGET_SLOT slot..."
az network front-door routing-rule update \
    --front-door-name $FRONTDOOR_NAME \
    --resource-group $RESOURCE_GROUP \
    --name default-routing-rule \
    --backend-pool "qlp-${TARGET_SLOT}-pool"

# Monitor new slot for stability
echo "📊 Monitoring $TARGET_SLOT slot for 5 minutes..."
./scripts/monitor-deployment.sh \
    --duration 300 \
    --error-threshold 5 \
    --response-time-threshold 2000

# If monitoring passes, cleanup old slot
echo "🧹 Cleaning up $CURRENT_SLOT slot..."
az container delete \
    --resource-group $RESOURCE_GROUP \
    --name "qlp-${CURRENT_SLOT}" \
    --yes

# Update slot tracking
echo $TARGET_SLOT > /tmp/current-slot

echo "🎉 Blue-green deployment completed successfully!"
echo "📊 Traffic now serving from $TARGET_SLOT slot"
```

### **Phase 4: Post-Deployment (T+30 minutes)**

#### **4.1 Extended Monitoring**
```bash
#!/bin/bash
# post-deployment-monitoring.sh

MONITORING_DURATION=${1:-1800}  # 30 minutes default
echo "📊 Starting extended post-deployment monitoring for ${MONITORING_DURATION}s..."

START_TIME=$(date +%s)
END_TIME=$((START_TIME + MONITORING_DURATION))

# Monitoring metrics
ERROR_COUNT=0
WARNING_COUNT=0
ALERT_THRESHOLD=10

while [ $(date +%s) -lt $END_TIME ]; do
    CURRENT_TIME=$(date +%s)
    ELAPSED=$((CURRENT_TIME - START_TIME))
    REMAINING=$((END_TIME - CURRENT_TIME))
    
    echo "⏱️  Monitoring progress: ${ELAPSED}s elapsed, ${REMAINING}s remaining"
    
    # 1. Health check
    HEALTH_STATUS=$(curl -s -w "%{http_code}" $PRODUCTION_URL/health -o /dev/null)
    if [ $HEALTH_STATUS -ne 200 ]; then
        ERROR_COUNT=$((ERROR_COUNT + 1))
        echo "❌ Health check failed: $HEALTH_STATUS (Error #$ERROR_COUNT)"
        
        if [ $ERROR_COUNT -ge $ALERT_THRESHOLD ]; then
            echo "🚨 ERROR THRESHOLD EXCEEDED! Triggering rollback..."
            ./scripts/emergency-rollback.sh "Health check failures: $ERROR_COUNT"
            exit 1
        fi
    fi
    
    # 2. Response time check
    RESPONSE_TIME=$(curl -s -w "%{time_total}" $PRODUCTION_URL/api/v1/status -o /dev/null)
    RESPONSE_MS=$(echo "$RESPONSE_TIME * 1000" | bc)
    
    if (( $(echo "$RESPONSE_MS > 2000" | bc -l) )); then
        WARNING_COUNT=$((WARNING_COUNT + 1))
        echo "⚠️ Slow response time: ${RESPONSE_MS}ms (Warning #$WARNING_COUNT)"
    fi
    
    # 3. Error rate check
    ERROR_RATE=$(az monitor metrics list \
        --resource $APP_INSIGHTS_RESOURCE \
        --metric "requests/failed" \
        --interval PT5M \
        --query 'value[0].timeseries[0].data[-1].total // 0')
    
    if [ $ERROR_RATE -gt 5 ]; then
        WARNING_COUNT=$((WARNING_COUNT + 1))
        echo "⚠️ Elevated error rate: $ERROR_RATE errors/5min (Warning #$WARNING_COUNT)"
    fi
    
    # 4. Database performance
    DB_RESPONSE_TIME=$(az monitor metrics list \
        --resource $DB_RESOURCE_ID \
        --metric "average_query_time_ms" \
        --interval PT5M \
        --query 'value[0].timeseries[0].data[-1].average // 0')
    
    if (( $(echo "$DB_RESPONSE_TIME > 100" | bc -l) )); then
        WARNING_COUNT=$((WARNING_COUNT + 1))
        echo "⚠️ Slow database queries: ${DB_RESPONSE_TIME}ms avg (Warning #$WARNING_COUNT)"
    fi
    
    sleep 30
done

echo "✅ Post-deployment monitoring completed!"
echo "📊 Summary: $ERROR_COUNT errors, $WARNING_COUNT warnings"

if [ $ERROR_COUNT -gt 0 ] || [ $WARNING_COUNT -gt 20 ]; then
    echo "⚠️ Consider investigating issues before next deployment"
    exit 1
fi
```

#### **4.2 Performance Baseline Update**
```bash
#!/bin/bash
# update-performance-baseline.sh

echo "📊 Updating performance baselines post-deployment..."

# Collect performance metrics
RESPONSE_TIME_P95=$(az monitor metrics list \
    --resource $APP_INSIGHTS_RESOURCE \
    --metric "requests/duration" \
    --interval PT1H \
    --aggregation Percentile95 \
    --query 'value[0].timeseries[0].data[-1].percentile95')

DB_CONNECTION_TIME=$(az monitor metrics list \
    --resource $DB_RESOURCE_ID \
    --metric "connections_active" \
    --interval PT1H \
    --query 'value[0].timeseries[0].data[-1].average')

MEMORY_USAGE=$(az monitor metrics list \
    --resource $CONTAINER_RESOURCE_ID \
    --metric "memory_usage" \
    --interval PT1H \
    --query 'value[0].timeseries[0].data[-1].average')

# Update baseline file
cat > performance-baseline.json <<EOF
{
    "version": "$DEPLOYMENT_VERSION",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "metrics": {
        "response_time_p95_ms": $RESPONSE_TIME_P95,
        "db_connection_time_ms": $DB_CONNECTION_TIME,
        "memory_usage_mb": $MEMORY_USAGE
    }
}
EOF

# Commit baseline to repository
git add performance-baseline.json
git commit -m "Update performance baseline for $DEPLOYMENT_VERSION"
git push origin main

echo "✅ Performance baseline updated for version: $DEPLOYMENT_VERSION"
```

---

## 🔥 Emergency Deployment Process

### **Emergency Criteria**
```yaml
Security Issues:
- Critical security vulnerabilities (CVE 9.0+)
- Active security breaches
- Data exposure incidents

Critical Bugs:
- Service unavailability
- Data corruption issues
- Payment processing failures
- Major feature failures affecting >50% users

Compliance Issues:
- Regulatory requirement violations
- Legal compliance mandates
- Audit finding remediations
```

### **Emergency Deployment Script**
```bash
#!/bin/bash
# emergency-deployment.sh

EMERGENCY_TYPE=${1:-"security"}
VERSION=${2:-"emergency-$(date +%Y%m%d-%H%M%S)"}
APPROVER=${3:-""}

echo "🚨 EMERGENCY DEPLOYMENT INITIATED"
echo "📝 Type: $EMERGENCY_TYPE"
echo "📦 Version: $VERSION" 
echo "👤 Approver: $APPROVER"

# Validate emergency approval
if [ -z "$APPROVER" ]; then
    echo "❌ Emergency deployments require approver name"
    echo "Usage: $0 <type> <version> <approver>"
    exit 1
fi

# Send emergency notifications
./scripts/notify-emergency.sh \
    --type "$EMERGENCY_TYPE" \
    --version "$VERSION" \
    --approver "$APPROVER"

# Skip normal validation for true emergencies
if [ "$EMERGENCY_TYPE" = "security" ] || [ "$EMERGENCY_TYPE" = "outage" ]; then
    echo "⚡ Skipping extended validation for emergency type: $EMERGENCY_TYPE"
    SKIP_VALIDATION=true
else
    SKIP_VALIDATION=false
fi

# Create emergency backup
echo "💾 Creating emergency backup..."
EMERGENCY_BACKUP="emergency-backup-$(date +%Y%m%d-%H%M%S)"
./scripts/create-snapshot.sh $EMERGENCY_BACKUP

# Fast-track deployment
if [ "$SKIP_VALIDATION" = "true" ]; then
    echo "🚀 Executing fast-track deployment..."
    ./scripts/fast-deploy.sh $VERSION
else
    echo "🚀 Executing standard emergency deployment..."
    ./scripts/blue-green-deploy.sh $VERSION 300  # 5 min timeout
fi

# Minimal post-deployment verification
echo "✅ Running minimal verification..."
timeout 60 bash -c 'until curl -f $PRODUCTION_URL/health; do sleep 5; done'

# Log emergency deployment
echo "📝 Logging emergency deployment..."
cat >> emergency-deployments.log <<EOF
$(date -u +%Y-%m-%dT%H:%M:%SZ) | $EMERGENCY_TYPE | $VERSION | $APPROVER
EOF

echo "🎉 Emergency deployment completed!"
echo "📊 Version $VERSION is now live"
echo "⚠️ Schedule post-emergency review within 24 hours"
```

---

## 🗄️ Database Migration Playbook

### **Migration Types**
```yaml
Schema Changes:
- Table creation/modification
- Index creation/removal
- Column additions/modifications
- Constraint changes

Data Migrations:
- Data transformations
- Data cleanup
- Bulk data imports
- Archive operations

Performance Migrations:
- Index optimizations
- Query performance improvements
- Partitioning implementation
```

### **Migration Deployment Process**
```bash
#!/bin/bash
# database-migration-deployment.sh

MIGRATION_VERSION=${1:-""}
MAINTENANCE_MODE=${2:-"false"}

if [ -z "$MIGRATION_VERSION" ]; then
    echo "❌ Migration version required"
    echo "Usage: $0 <migration_version> [maintenance_mode]"
    exit 1
fi

echo "🗄️ Starting database migration deployment: $MIGRATION_VERSION"

# 1. Validate migration scripts
echo "🔍 Validating migration scripts..."
if [ ! -f "migrations/v${MIGRATION_VERSION}.sql" ]; then
    echo "❌ Migration file not found: migrations/v${MIGRATION_VERSION}.sql"
    exit 1
fi

# Validate SQL syntax
if ! psql --dry-run -f "migrations/v${MIGRATION_VERSION}.sql" >/dev/null 2>&1; then
    echo "❌ Migration SQL syntax validation failed"
    exit 1
fi

# 2. Create pre-migration backup
echo "💾 Creating pre-migration backup..."
BACKUP_NAME="pre-migration-v${MIGRATION_VERSION}-$(date +%Y%m%d-%H%M%S)"
./scripts/create-database-backup.sh $BACKUP_NAME

# 3. Test migration on staging
echo "🧪 Testing migration on staging..."
./scripts/test-migration-staging.sh $MIGRATION_VERSION

# 4. Enable maintenance mode if required
if [ "$MAINTENANCE_MODE" = "true" ]; then
    echo "🚧 Enabling maintenance mode..."
    ./scripts/maintenance-mode.sh enable
    
    # Wait for connections to drain
    echo "⏳ Waiting for active connections to drain..."
    sleep 30
fi

# 5. Execute migration
echo "🔄 Executing database migration..."
MIGRATION_START=$(date +%s)

psql $DATABASE_URL -f "migrations/v${MIGRATION_VERSION}.sql" 2>&1 | tee migration-log-${MIGRATION_VERSION}.txt

MIGRATION_END=$(date +%s)
MIGRATION_DURATION=$((MIGRATION_END - MIGRATION_START))

# 6. Verify migration success
echo "✅ Verifying migration success..."
MIGRATION_STATUS=$(psql $DATABASE_URL -t -c "SELECT version FROM schema_migrations WHERE version = '$MIGRATION_VERSION';" | xargs)

if [ "$MIGRATION_STATUS" != "$MIGRATION_VERSION" ]; then
    echo "❌ Migration verification failed!"
    
    if [ "$MAINTENANCE_MODE" = "true" ]; then
        echo "🔄 Rolling back migration..."
        ./scripts/rollback-migration.sh $MIGRATION_VERSION
        ./scripts/maintenance-mode.sh disable
    fi
    exit 1
fi

# 7. Update application if needed
if [ -f "deployment/migration-${MIGRATION_VERSION}-app-update.sh" ]; then
    echo "🚀 Executing application update for migration..."
    ./deployment/migration-${MIGRATION_VERSION}-app-update.sh
fi

# 8. Disable maintenance mode
if [ "$MAINTENANCE_MODE" = "true" ]; then
    echo "🟢 Disabling maintenance mode..."
    ./scripts/maintenance-mode.sh disable
fi

# 9. Post-migration verification
echo "🧪 Running post-migration tests..."
./scripts/post-migration-tests.sh $MIGRATION_VERSION

echo "🎉 Database migration completed successfully!"
echo "📊 Migration $MIGRATION_VERSION completed in ${MIGRATION_DURATION}s"
```

---

## 🔄 Rollback Procedures

### **Rollback Decision Matrix**
```yaml
Automatic Rollback Triggers:
- Health check failures (>50% in 5 minutes)
- Error rate spike (>10x baseline)
- Critical functionality broken
- Database connection failures

Manual Rollback Scenarios:
- Performance degradation (>2x baseline)
- User-reported critical issues
- Business logic errors
- Integration failures

Rollback Approval Required:
- Database schema rollbacks
- Data loss potential
- Complex dependency changes
```

### **Automatic Rollback Script**
```bash
#!/bin/bash
# automatic-rollback.sh

ROLLBACK_REASON=${1:-"Automatic rollback triggered"}
PREVIOUS_VERSION=${2:-$(cat /tmp/previous-version 2>/dev/null || echo "unknown")}

echo "🚨 AUTOMATIC ROLLBACK INITIATED"
echo "📝 Reason: $ROLLBACK_REASON"
echo "🔙 Target version: $PREVIOUS_VERSION"

# 1. Immediate traffic switch
echo "⚡ Switching traffic to previous version..."
CURRENT_SLOT=$(cat /tmp/current-slot 2>/dev/null || echo "blue")
if [ "$CURRENT_SLOT" = "blue" ]; then
    ROLLBACK_SLOT="green"
else
    ROLLBACK_SLOT="blue"
fi

# Check if rollback slot exists and is healthy
ROLLBACK_EXISTS=$(az container show \
    --resource-group $RESOURCE_GROUP \
    --name "qlp-${ROLLBACK_SLOT}" \
    --query 'name' -o tsv 2>/dev/null || echo "")

if [ -n "$ROLLBACK_EXISTS" ]; then
    echo "✅ Rollback slot found: $ROLLBACK_SLOT"
    
    # Switch traffic
    az network front-door routing-rule update \
        --front-door-name $FRONTDOOR_NAME \
        --resource-group $RESOURCE_GROUP \
        --name default-routing-rule \
        --backend-pool "qlp-${ROLLBACK_SLOT}-pool"
    
    # Update slot tracking
    echo $ROLLBACK_SLOT > /tmp/current-slot
    
else
    echo "⚠️ No rollback slot available, deploying previous version..."
    ./scripts/blue-green-deploy.sh $PREVIOUS_VERSION 300
fi

# 2. Verify rollback success
echo "🔍 Verifying rollback..."
timeout 120 bash -c 'until curl -f $PRODUCTION_URL/health; do sleep 5; done'

if [ $? -eq 0 ]; then
    echo "✅ Rollback verification successful"
else
    echo "❌ Rollback verification failed!"
    # Try emergency deployment of known good version
    echo "🚨 Attempting emergency deployment of last known good version..."
    ./scripts/emergency-deployment.sh "rollback-failure" "last-known-good" "automatic-rollback"
fi

# 3. Send notifications
echo "📧 Sending rollback notifications..."
./scripts/notify-rollback.sh \
    --reason "$ROLLBACK_REASON" \
    --version "$PREVIOUS_VERSION" \
    --success "true"

echo "✅ Automatic rollback completed!"
echo "📋 Post-rollback investigation required"
```

### **Manual Rollback Script**
```bash
#!/bin/bash
# manual-rollback.sh

ROLLBACK_VERSION=${1:-""}
ROLLBACK_REASON=${2:-"Manual rollback requested"}
APPROVER=${3:-""}

if [ -z "$ROLLBACK_VERSION" ] || [ -z "$APPROVER" ]; then
    echo "❌ Missing required parameters"
    echo "Usage: $0 <version> <reason> <approver>"
    exit 1
fi

echo "🔄 MANUAL ROLLBACK INITIATED"
echo "📦 Target version: $ROLLBACK_VERSION"
echo "📝 Reason: $ROLLBACK_REASON"
echo "👤 Approved by: $APPROVER"

# Confirm rollback
read -p "⚠️ Confirm rollback to $ROLLBACK_VERSION? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "❌ Rollback cancelled"
    exit 1
fi

# Create rollback snapshot
echo "📸 Creating rollback snapshot..."
ROLLBACK_SNAPSHOT="rollback-snapshot-$(date +%Y%m%d-%H%M%S)"
./scripts/create-snapshot.sh $ROLLBACK_SNAPSHOT

# Execute rollback deployment
echo "🚀 Executing rollback deployment..."
./scripts/blue-green-deploy.sh $ROLLBACK_VERSION

# Post-rollback verification
echo "🧪 Running post-rollback verification..."
./scripts/post-deployment-tests.sh

# Log rollback
echo "📝 Logging rollback..."
cat >> rollback-log.txt <<EOF
$(date -u +%Y-%m-%dT%H:%M:%SZ) | $ROLLBACK_VERSION | $ROLLBACK_REASON | $APPROVER
EOF

echo "✅ Manual rollback completed successfully!"
```

---

## 📊 Monitoring & Alerting

### **Deployment Monitoring Dashboard**
```yaml
Key Metrics:
- Deployment success rate
- Average deployment time
- Rollback frequency
- Error rate during deployments
- Performance impact post-deployment

Alerts:
- Deployment failures
- Extended deployment times (>60 minutes)
- High error rates post-deployment
- Performance degradation
- Rollback events
```

### **Deployment Health Check Script**
```bash
#!/bin/bash
# deployment-health-check.sh

DEPLOYMENT_ID=${1:-"current"}
echo "🏥 Running deployment health check for: $DEPLOYMENT_ID"

# 1. Service availability
echo "🔍 Checking service availability..."
for endpoint in "/health" "/api/v1/status" "/api/v1/intents"; do
    STATUS=$(curl -s -w "%{http_code}" "$PRODUCTION_URL$endpoint" -o /dev/null)
    if [ $STATUS -ne 200 ]; then
        echo "❌ $endpoint returned $STATUS"
        exit 1
    else
        echo "✅ $endpoint: OK"
    fi
done

# 2. Database connectivity
echo "🗄️ Checking database connectivity..."
DB_STATUS=$(curl -s "$PRODUCTION_URL/api/v1/health/database" | jq -r '.status')
if [ "$DB_STATUS" != "healthy" ]; then
    echo "❌ Database health check failed: $DB_STATUS"
    exit 1
else
    echo "✅ Database: OK"
fi

# 3. Vector search functionality
echo "🔍 Testing vector search..."
SEARCH_RESPONSE=$(curl -s -X POST "$PRODUCTION_URL/api/v1/search" \
    -H "Content-Type: application/json" \
    -d '{"query": "test search"}')

SEARCH_STATUS=$(echo $SEARCH_RESPONSE | jq -r '.status // "error"')
if [ "$SEARCH_STATUS" != "success" ]; then
    echo "❌ Vector search failed: $SEARCH_STATUS"
    exit 1
else
    echo "✅ Vector search: OK"
fi

# 4. Performance check
echo "📊 Performance check..."
RESPONSE_TIME=$(curl -s -w "%{time_total}" "$PRODUCTION_URL/health" -o /dev/null)
RESPONSE_MS=$(echo "$RESPONSE_TIME * 1000" | bc)

if (( $(echo "$RESPONSE_MS > 1000" | bc -l) )); then
    echo "⚠️ Slow response time: ${RESPONSE_MS}ms"
else
    echo "✅ Response time: ${RESPONSE_MS}ms"
fi

echo "✅ All health checks passed!"
```

---

## 🎯 Deployment Checklist Templates

### **Standard Deployment Checklist**
```yaml
Pre-Deployment:
- [ ] Code reviewed and approved
- [ ] Tests passing (unit, integration, security)
- [ ] Staging deployment tested
- [ ] Database migration tested (if applicable)
- [ ] Performance impact assessed
- [ ] Rollback plan prepared
- [ ] Team notification sent
- [ ] Change request approved

During Deployment:
- [ ] Pre-deployment backup created
- [ ] Infrastructure validation passed
- [ ] Application deployed successfully
- [ ] Database migrations completed (if applicable)
- [ ] Health checks passing
- [ ] Smoke tests completed
- [ ] Performance monitoring active

Post-Deployment:
- [ ] Extended monitoring completed (30 minutes)
- [ ] Performance baseline updated
- [ ] Customer communication sent (if needed)
- [ ] Documentation updated
- [ ] Deployment retrospective scheduled
- [ ] Success metrics recorded
```

### **Emergency Deployment Checklist**
```yaml
Emergency Authorization:
- [ ] Emergency approval obtained
- [ ] Business impact assessed
- [ ] Technical risk evaluated
- [ ] Communication plan activated

Emergency Deployment:
- [ ] Emergency backup created
- [ ] Fast-track deployment executed
- [ ] Critical functionality verified
- [ ] Emergency notification sent
- [ ] Incident tracking updated

Post-Emergency:
- [ ] Full system verification
- [ ] Performance monitoring extended
- [ ] Post-emergency review scheduled
- [ ] Process improvement items identified
- [ ] Documentation updated
```

---

## 🔧 Deployment Scripts Repository

### **Script Organization**
```
scripts/
├── deployment/
│   ├── blue-green-deploy.sh
│   ├── standard-deploy.sh
│   ├── emergency-deploy.sh
│   └── rollback.sh
├── validation/
│   ├── pre-deployment-check.sh
│   ├── health-check.sh
│   ├── smoke-tests.sh
│   └── performance-test.sh
├── database/
│   ├── run-migrations.sh
│   ├── rollback-migration.sh
│   ├── backup-database.sh
│   └── verify-backup.sh
├── monitoring/
│   ├── deployment-monitor.sh
│   ├── alert-setup.sh
│   └── metrics-collection.sh
└── utilities/
    ├── maintenance-mode.sh
    ├── notify-team.sh
    ├── create-snapshot.sh
    └── cleanup-resources.sh
```

### **Script Configuration**
```bash
# deployment-config.sh
# Source this file before running deployment scripts

# Azure Configuration
export RESOURCE_GROUP="prod-qlp-rg"
export CONTAINER_GROUP_NAME="qlp-containers"
export ACR_NAME="prodqlpacr"
export FRONTDOOR_NAME="qlp-frontdoor"
export DB_SERVER_NAME="prod-qlp-postgres"

# Application Configuration
export PRODUCTION_URL="https://api.qlp-hq.com"
export STAGING_URL="https://staging.qlp-hq.com"
export APP_INSIGHTS_RESOURCE="/subscriptions/.../prod-qlp-insights"
export DB_RESOURCE_ID="/subscriptions/.../prod-qlp-postgres"

# Deployment Settings
export DEPLOYMENT_TIMEOUT=600
export HEALTH_CHECK_TIMEOUT=300
export MONITORING_DURATION=1800
export ERROR_THRESHOLD=10

# Notification Settings
export SLACK_WEBHOOK="https://hooks.slack.com/..."
export PAGERDUTY_SERVICE_KEY="..."
export EMAIL_RECIPIENTS="team@qlp-hq.com"

# Safety Settings
export REQUIRE_APPROVAL=true
export AUTO_ROLLBACK_ENABLED=true
export MAINTENANCE_MODE_TIMEOUT=1800
```

---

*This deployment playbook provides comprehensive procedures for safe, reliable deployments of QLP to Azure production environment with proper safety checks, monitoring, and rollback capabilities.*