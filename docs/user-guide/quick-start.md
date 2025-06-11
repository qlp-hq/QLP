# 🚀 Quick Start Guide

**Get QuantumLayer running with enterprise-grade validation in 5 minutes**

---

## ⏱️ **5-Minute Enterprise Deployment**

### **Step 1: Prerequisites** (30 seconds)

Ensure you have:
- ✅ **Go 1.21+** installed
- ✅ **Git** for cloning
- ✅ **Azure OpenAI access** (or Ollama for local)

```bash
# Verify Go version
go version  # Should show 1.21+
```

### **Step 2: Clone & Setup** (1 minute)

```bash
# Clone the repository
git clone https://github.com/qlp-hq/QLP.git
cd QLP

# Download dependencies
go mod tidy
```

### **Step 3: Configure Environment** (1 minute)

```bash
# Option A: Azure OpenAI (Recommended for enterprise)
export AZURE_OPENAI_API_KEY="your-azure-openai-key"
export AZURE_OPENAI_ENDPOINT="https://your-instance.openai.azure.com"

# Option B: Local Ollama (Development)
# ollama serve &
# ollama pull llama3
```

### **Step 4: Build & Test** (1 minute)

```bash
# Build QuantumLayer
go build -o qlp ./main.go

# Verify installation
./qlp --version
```

### **Step 5: Your First Enterprise Deployment** (2 minutes)

```bash
# Deploy a secure REST API with full validation
./qlp "Create a secure REST API for user management with JWT authentication"
```

🎉 **You'll see:**
```
🔄 Processing intent: Create a secure REST API...
📋 Parsed 12 tasks from intent
🤖 Executing task graph with 12 real agents
💧 Generated 3 QuantumDrops
🤔 Processing HITL decisions...
📦 Generating final QuantumCapsule...
🎯 QuantumCapsule generated: QL-CAP-xyz
   📊 Overall Score: 94/100
   🔒 Security Risk: LOW
   📈 Quality Score: 92/100
   ⏱️ Execution Time: 2.3s
```

---

## 🎯 **Understanding Your Results**

### **🏆 What Just Happened?**

1. **🔍 Layer 1: Static Validation**
   - LLM analyzed code for security vulnerabilities
   - Quality metrics calculated for maintainability
   - Architecture patterns validated for scalability

2. **🚀 Layer 2: Dynamic Testing**
   - Code built and deployed in sandbox
   - Load testing performed automatically
   - Security scanning completed

3. **🏢 Layer 3: Enterprise Readiness**
   - SOC2, GDPR compliance checked
   - Production deployment risks assessed
   - Enterprise certifications validated

4. **🤖 HITL Decision Engine**
   - AI-powered quality gates evaluated
   - Automatic approval at 90%+ confidence
   - Human review triggered only if needed

### **📊 Your Confidence Score Breakdown**

```
📊 OVERALL CONFIDENCE: 94/100 (ENTERPRISE GRADE)
├── 🔒 Security Score: 88/100
├── 🎯 Quality Score: 92/100  
├── 🏗️ Architecture Score: 96/100
├── 📋 Compliance Score: 94/100
└── ⚡ Performance Score: 89/100
```

### **📦 Your QuantumCapsule Contains**

```
📦 QL-CAP-xyz.qlcapsule
├── 📄 manifest.json          # Deployment metadata
├── 🗂️ project/              # Complete application code
│   ├── main.go              # REST API server
│   ├── auth/                # JWT authentication
│   ├── handlers/            # API endpoints
│   └── tests/               # Comprehensive tests
├── 📊 reports/              # Validation reports
│   ├── security_report.json
│   ├── quality_report.json
│   └── compliance_report.json
└── 📋 README.md             # Deployment instructions
```

---

## 🎪 **Try More Examples**

### **🏗️ Infrastructure Automation**
```bash
./qlp "Build Kubernetes infrastructure for a microservices deployment"
```

### **📊 Data Pipeline**
```bash
./qlp "Create a real-time data processing pipeline with Apache Kafka"
```

### **🛡️ Security Audit**
```bash
./qlp "Perform security audit and penetration testing on existing API"
```

---

## 📈 **Next Steps**

### **🔧 Customize Your Validation**
Learn to configure validation rules for your specific needs:
```bash
# Configure compliance frameworks
export QLP_COMPLIANCE_FRAMEWORKS="SOC2,GDPR,HIPAA"

# Set quality thresholds
export QLP_MIN_QUALITY_SCORE="85"
export QLP_MIN_SECURITY_SCORE="90"
```

### **🏢 Enterprise Integration**
Connect QuantumLayer to your enterprise stack:
- 📊 [Monitoring Integration](/user-guide/integrations/monitoring/)
- 🔄 [CI/CD Pipeline Setup](/user-guide/integrations/cicd/)
- 📧 [Slack/Teams Notifications](/user-guide/integrations/notifications/)

### **🎯 Advanced Workflows**
Master complex deployment scenarios:
- 🌐 [Multi-Environment Deployments](/user-guide/workflows/multi-env/)
- 🔄 [Blue-Green Deployments](/user-guide/workflows/blue-green/)
- 📊 [A/B Testing Setup](/user-guide/workflows/ab-testing/)

---

## 🆘 **Troubleshooting**

### **Common Issues & Solutions**

#### **❌ "LLM client failed"**
```bash
# Check environment variables
echo $AZURE_OPENAI_API_KEY
echo $AZURE_OPENAI_ENDPOINT

# Test connection
curl -H "api-key: $AZURE_OPENAI_API_KEY" $AZURE_OPENAI_ENDPOINT/openai/deployments
```

#### **❌ "Build failed"**
```bash
# Check Go version
go version  # Ensure 1.21+

# Clean and rebuild
go clean -cache
go mod tidy
go build -o qlp ./main.go
```

#### **❌ "Low confidence score"**
```bash
# Enable all validation layers
export QLP_ENABLE_ALL_LAYERS="true"

# Use more detailed prompts
./qlp "Create a highly secure, scalable REST API with comprehensive error handling and monitoring"
```

---

## 🎖️ **Success Metrics**

**🎯 You're successful when you see:**
- ✅ **90%+ confidence scores** consistently
- ✅ **Enterprise-grade compliance** validation
- ✅ **Auto-approved HITL decisions** at scale
- ✅ **Production-ready QuantumCapsules** every time

---

## 🔗 **What's Next?**

### **📚 Deep Dive Learning**
- 🛡️ [Understanding 3-Layer Validation](/user-guide/validation-layers/)
- 🤖 [Mastering HITL Decision Engine](/user-guide/hitl-engine/)
- 📊 [Optimizing Confidence Scores](/user-guide/confidence-scoring/)

### **🏢 Enterprise Features**
- 💼 [Enterprise Pricing & Licensing](/enterprise/pricing/)
- 🔒 [Security & Compliance](/enterprise/compliance/)
- 📞 [Professional Support](/enterprise/support/)

---

**🚀 Congratulations! You've just deployed with enterprise-grade confidence!**

Ready to transform your entire development workflow? 

[➡️ Continue to Full User Guide](/user-guide/) | [🏢 Explore Enterprise Features](/enterprise/)