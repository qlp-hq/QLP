# 🆘 Troubleshooting Guide

**Complete troubleshooting guide for QuantumLayer enterprise deployments**

---

## 🎯 **Quick Diagnostics**

### **Health Check Commands**
```bash
# Basic health check
./qlp health-check

# Comprehensive system check
./qlp health-check --verbose --all-components

# Check specific component
./qlp health-check --component=llm
./qlp health-check --component=validation
./qlp health-check --component=agents
```

### **Log Analysis**
```bash
# View real-time logs
sudo journalctl -u qlp-orchestrator -f

# Search for errors
sudo journalctl -u qlp-orchestrator | grep -i error

# Check last 100 lines
sudo journalctl -u qlp-orchestrator -n 100

# Export logs for support
sudo journalctl -u qlp-orchestrator --since="1 hour ago" > qlp-logs.txt
```

---

## 🚨 **Common Issues & Solutions**

### **1. LLM Connection Issues**

#### **❌ Problem**: "LLM client failed: timeout"
```
Error: LLM client 1 failed: timeout waiting for response
```

#### **✅ Solutions**:
```bash
# Check API key and endpoint
echo $AZURE_OPENAI_API_KEY
echo $AZURE_OPENAI_ENDPOINT

# Test direct connection
curl -H "api-key: $AZURE_OPENAI_API_KEY" \
  "$AZURE_OPENAI_ENDPOINT/openai/deployments" \
  --connect-timeout 10

# Increase timeout
export QLP_LLM_TIMEOUT="120s"
export QLP_LLM_RETRY_COUNT="5"

# Check firewall/proxy
telnet your-endpoint.openai.azure.com 443
```

#### **❌ Problem**: "Model 'llama3' not found"
```
Error: Ollama returned status 404: {"error":"model 'llama3' not found"}
```

#### **✅ Solutions**:
```bash
# Install Ollama model
ollama pull llama3

# Start Ollama service
ollama serve &

# Check available models
ollama list

# Use different model
export QLP_OLLAMA_MODEL="codellama"
```

### **2. Validation Failures**

#### **❌ Problem**: Low confidence scores (< 80)
```
Validation completed: Score=65, Passed=false
```

#### **✅ Solutions**:
```bash
# Use more detailed prompts
./qlp "Create a highly secure, scalable REST API with comprehensive error handling, input validation, rate limiting, and monitoring"

# Enable all validation layers
export QLP_ENABLE_ALL_LAYERS="true"
export QLP_VALIDATION_LEVEL="enterprise"

# Adjust thresholds temporarily for testing
export QLP_MIN_CONFIDENCE_SCORE="70"

# Check validation logs
./qlp validate --debug --verbose
```

#### **❌ Problem**: Security validation failing
```
Security Score: 45/100 - Multiple high-severity issues found
```

#### **✅ Solutions**:
```bash
# Check security configuration
export QLP_SECURITY_LEVEL="high"
export QLP_COMPLIANCE_FRAMEWORKS="SOC2,GDPR"

# Review security findings
./qlp validate --security-only --detailed-report

# Update security rules
./qlp config update --security-rules=latest

# Check for hardcoded secrets
grep -r "password\|key\|secret" ./code/
```

### **3. Performance Issues**

#### **❌ Problem**: Slow execution times (> 5 minutes)
```
Execution completed in 8m 45s (expected < 2m)
```

#### **✅ Solutions**:
```bash
# Check system resources
htop
free -h
df -h

# Monitor QLPprocesses
ps aux | grep qlp
netstat -tulpn | grep qlp

# Optimize agent configuration
export QLP_MAX_CONCURRENT_AGENTS="10"
export QLP_AGENT_TIMEOUT="120s"

# Enable caching
export QLP_VALIDATION_CACHE_ENABLED="true"
export QLP_VALIDATION_CACHE_TTL="3600s"

# Use faster LLM provider
export QLP_PREFERRED_LLM="azure_openai"
```

#### **❌ Problem**: High memory usage
```
Memory usage: 8GB / 8GB (100% - OOM risk)
```

#### **✅ Solutions**:
```bash
# Reduce concurrent agents
export QLP_MAX_CONCURRENT_AGENTS="5"

# Limit validation scope
export QLP_VALIDATION_SAMPLE_SIZE="50"

# Check for memory leaks
valgrind ./qlp "simple test"

# Restart service
sudo systemctl restart qlp-orchestrator
```

### **4. HITL Decision Engine Issues**

#### **❌ Problem**: All tasks requiring manual review
```
HITL Decision: review_required=true (low confidence)
```

#### **✅ Solutions**:
```bash
# Lower auto-approval threshold
export QLP_AUTO_APPROVE_THRESHOLD="85"

# Check HITL configuration
./qlp config show --section=hitl

# Review quality gates
./qlp hitl analyze --last-decisions=10

# Update decision model
./qlp hitl update-model --training-data=recent
```

### **5. QuantumCapsule Generation Issues**

#### **❌ Problem**: Capsule generation failing
```
Error: Failed to package capsule: invalid project structure
```

#### **✅ Solutions**:
```bash
# Check project structure
./qlp validate --project-structure-only

# Clean temporary files
rm -rf /tmp/qlp_validation/*

# Check disk space
df -h /tmp

# Manual capsule generation
./qlp package --debug --verbose \
  --intent-id=QLI-123 \
  --output-dir=/tmp/manual-capsule
```

---

## 🔧 **Advanced Diagnostics**

### **System Performance Analysis**
```bash
# CPU usage analysis
sar -u 1 10

# Memory usage analysis  
sar -r 1 10

# I/O analysis
sar -b 1 10

# Network analysis
sar -n DEV 1 10

# QLPspecific metrics
curl http://localhost:9090/metrics | grep qlp_
```

### **Database Diagnostics**
```bash
# Check database connections
sudo -u postgres psql -c "SELECT * FROM pg_stat_activity WHERE application_name LIKE '%qlp%';"

# Check database performance
sudo -u postgres psql -c "SELECT query, mean_exec_time, calls FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# Database size
sudo -u postgres psql -c "SELECT pg_size_pretty(pg_database_size('qlp_production'));"
```

### **Network Diagnostics**
```bash
# Test LLM endpoint connectivity
dig your-endpoint.openai.azure.com
nslookup your-endpoint.openai.azure.com
ping your-endpoint.openai.azure.com

# Check ports
netstat -tulpn | grep :8080
ss -tulpn | grep :8080

# Test API endpoints
curl -v http://localhost:8080/health
curl -v http://localhost:8080/metrics
```

---

## 📊 **Monitoring & Alerting**

### **Setting Up Alerts**
```yaml
# prometheus-alerts.yml
groups:
  - name: qlp-alerts
    rules:
      - alert: QLPServiceDown
        expr: up{job="qlp-orchestrator"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "QuantumLayer service is down"
          
      - alert: QLPLowConfidenceScore
        expr: qlp_avg_confidence_score < 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Average confidence score below threshold"
          
      - alert: QLPHighMemoryUsage
        expr: qlp_memory_usage_percent > 90
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "QuantumLayer memory usage critical"
```

### **Grafana Dashboards**
```json
{
  "dashboard": {
    "title": "QuantumLayer Troubleshooting",
    "panels": [
      {
        "title": "Error Rate",
        "targets": [{"expr": "rate(qlp_errors_total[5m])"}]
      },
      {
        "title": "Response Times",
        "targets": [{"expr": "histogram_quantile(0.95, qlp_request_duration_seconds_bucket)"}]
      },
      {
        "title": "LLM Success Rate", 
        "targets": [{"expr": "rate(qlp_llm_requests_successful[5m])"}]
      }
    ]
  }
}
```

---

## 🔍 **Debug Mode**

### **Enable Debug Logging**
```bash
# Temporary debug mode
export QLP_LOG_LEVEL="debug"
export QLP_DEBUG_MODE="true"
./qlp "test command"

# Persistent debug mode
sudo tee /etc/qlp/debug.conf << EOF
log_level = "debug"
debug_mode = true
trace_requests = true
log_llm_requests = true
log_validation_details = true
EOF

sudo systemctl restart qlp-orchestrator
```

### **Trace Mode**
```bash
# Enable request tracing
export QLP_TRACE_ENABLED="true"
export QLP_TRACE_OUTPUT="/tmp/qlp-trace.log"

# View trace output
tail -f /tmp/qlp-trace.log

# Analyze trace
./qlp trace analyze --input=/tmp/qlp-trace.log
```

---

## 🛠️ **Configuration Validation**

### **Validate Configuration**
```bash
# Check configuration syntax
./qlp config validate

# Test configuration
./qlp config test --dry-run

# Show effective configuration
./qlp config show --all

# Check environment variables
./qlp config env-check
```

### **Reset Configuration**
```bash
# Backup current config
sudo cp /etc/qlp/config.yaml /etc/qlp/config.yaml.backup

# Reset to defaults
sudo ./qlp config reset --confirm

# Restore from backup
sudo cp /etc/qlp/config.yaml.backup /etc/qlp/config.yaml
sudo systemctl restart qlp-orchestrator
```

---

## 🔄 **Recovery Procedures**

### **Service Recovery**
```bash
# Graceful restart
sudo systemctl restart qlp-orchestrator

# Force restart
sudo systemctl kill qlp-orchestrator
sudo systemctl start qlp-orchestrator

# Check service status
sudo systemctl status qlp-orchestrator --no-pager -l

# Emergency stop
sudo pkill -9 -f qlp-orchestrator
```

### **Database Recovery**
```bash
# Check database status
sudo systemctl status postgresql

# Restart database
sudo systemctl restart postgresql

# Check connections
sudo -u postgres psql -c "SELECT count(*) FROM pg_stat_activity;"

# Vacuum and analyze
sudo -u postgres psql qlp_production -c "VACUUM ANALYZE;"
```

### **Cache Recovery**
```bash
# Clear validation cache
rm -rf /var/lib/qlp/cache/*

# Clear temporary files
rm -rf /tmp/qlp_*

# Reset file permissions
sudo chown -R qlp:qlp /var/lib/qlp
sudo chmod -R 755 /var/lib/qlp
```

---

## 📞 **Getting Support**

### **Preparing Support Requests**
```bash
# Generate support bundle
./qlp support bundle --output=/tmp/qlp-support-bundle.tar.gz

# Include system information
uname -a > /tmp/system-info.txt
free -h >> /tmp/system-info.txt
df -h >> /tmp/system-info.txt

# Collect recent logs
sudo journalctl -u qlp-orchestrator --since="24 hours ago" > /tmp/qlp-recent-logs.txt
```

### **Support Channels**

#### **Enterprise Support (24/7)**
- 📞 **Emergency Hotline**: +1-800-QLP-HELP
- 📧 **Critical Issues**: [emergency@qlp-hq.com](mailto:emergency@qlp-hq.com)
- 💬 **Enterprise Slack**: [#qlp-emergency](https://qlp-hq.slack.com)

#### **Technical Support**
- 📧 **General Support**: [support@qlp-hq.com](mailto:support@qlp-hq.com)
- 💬 **Community Discord**: [Discord Server](https://discord.gg/qlp)
- 📖 **GitHub Issues**: [Report Bug](https://github.com/qlp-hq/QLP/issues)

#### **Professional Services**
- 📧 **Consulting**: [consulting@qlp-hq.com](mailto:consulting@qlp-hq.com)
- 📞 **Architecture Review**: [Schedule Call](https://calendly.com/qlp-architecture)

---

## 📚 **Knowledge Base**

### **FAQ - Frequently Asked Questions**

#### **Q: Why is my confidence score always low?**
A: Low confidence scores usually indicate:
- Insufficient detail in prompts
- Missing security or quality patterns
- Outdated validation rules
- LLM connectivity issues

#### **Q: How do I speed up validation?**
A: Performance optimization steps:
- Enable validation caching
- Reduce concurrent agents if memory-limited
- Use Azure OpenAI instead of local Ollama
- Enable parallel processing

#### **Q: What compliance frameworks are supported?**
A: Currently supported:
- SOC 2 Type II
- GDPR
- HIPAA
- PCI DSS
- ISO 27001
- Custom frameworks available for Enterprise+

#### **Q: How do I customize validation rules?**
A: Enterprise customers can:
- Define custom security rules
- Configure compliance frameworks
- Set quality thresholds
- Create custom approval workflows

### **Best Practices**
- ✅ Always use detailed, specific prompts
- ✅ Enable all validation layers for production
- ✅ Monitor confidence score trends
- ✅ Set up proper alerting and monitoring
- ✅ Regularly update validation rules
- ✅ Use enterprise-grade LLM providers
- ✅ Implement proper backup and recovery
- ✅ Follow security hardening guidelines

---

## 📈 **Performance Tuning Guide**

### **Optimization Checklist**
```bash
# System optimization
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
echo 'fs.file-max=65536' | sudo tee -a /etc/sysctl.conf

# QLPoptimization
export QLP_MAX_CONCURRENT_AGENTS="20"
export QLP_VALIDATION_CACHE_TTL="3600s"
export QLP_LLM_CONNECTION_POOL_SIZE="10"

# Database optimization
sudo -u postgres psql qlp_production -c "
ALTER SYSTEM SET shared_buffers = '1GB';
ALTER SYSTEM SET effective_cache_size = '3GB';
ALTER SYSTEM SET maintenance_work_mem = '256MB';
"
```

---

**🎖️ Need immediate help? Our enterprise support team is standing by 24/7!**

[📞 Emergency Support](tel:+1-800-QLP-HELP) | [📧 Support Email](mailto:support@qlp-hq.com) | [💬 Live Chat](https://support.qlp-hq.com)