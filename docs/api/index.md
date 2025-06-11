# 🔧 API Reference

**Complete API documentation for enterprise QuantumLayer integration**

---

## 🎯 **API Overview**

QuantumLayer provides comprehensive REST APIs for enterprise integration, enabling you to programmatically access the 3-layer validation system and achieve 94/100 confidence scores in your applications.

**Base URL**: `https://api.qlp-hq.com/v1`  
**Authentication**: Bearer token or API key  
**Format**: JSON  
**Rate Limits**: 1000 requests/hour (Enterprise: unlimited)

---

## 🔑 **Authentication**

### **API Key Authentication**
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     https://api.qlp-hq.com/v1/deployments
```

### **Environment Variables**
```bash
export QLP_API_KEY="your-api-key"
export QLP_API_ENDPOINT="https://api.qlp-hq.com/v1"
```

---

## 🚀 **Core Endpoints**

### **🎯 Intent Processing**

#### **POST /intents**
Process natural language intent and generate deployment plan

```bash
curl -X POST https://api.qlp-hq.com/v1/intents \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "intent": "Create a secure REST API for user management with JWT authentication",
    "validation_level": "enterprise",
    "compliance_frameworks": ["SOC2", "GDPR"],
    "quality_threshold": 90
  }'
```

**Response**:
```json
{
  "intent_id": "QLI-1749632410195788000",
  "tasks": [
    {
      "id": "QL-DEV-001",
      "type": "codegen",
      "description": "Set up secure authentication system",
      "dependencies": [],
      "estimated_duration": "2m"
    }
  ],
  "validation_plan": {
    "layers": ["static", "dynamic", "enterprise"],
    "estimated_confidence": 94,
    "compliance_checks": ["SOC2", "GDPR"]
  }
}
```

#### **GET /intents/{intent_id}**
Get intent processing status and results

```bash
curl https://api.qlp-hq.com/v1/intents/QLI-1749632410195788000 \
  -H "Authorization: Bearer YOUR_API_KEY"
```

---

### **🤖 Agent Execution**

#### **POST /executions**
Execute task graph with dynamic agents

```bash
curl -X POST https://api.qlp-hq.com/v1/executions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "intent_id": "QLI-1749632410195788000",
    "execution_mode": "parallel",
    "agent_config": {
      "llm_provider": "azure_openai",
      "sandbox_enabled": true,
      "validation_enabled": true
    }
  }'
```

**Response**:
```json
{
  "execution_id": "QLE-001",
  "status": "running",
  "progress": {
    "completed_tasks": 2,
    "total_tasks": 12,
    "current_phase": "dynamic_validation"
  },
  "agents": [
    {
      "agent_id": "QLD-AGT-001",
      "task_id": "QL-DEV-001", 
      "status": "completed",
      "confidence_score": 88
    }
  ]
}
```

#### **GET /executions/{execution_id}**
Monitor execution progress and results

---

### **🛡️ Validation System**

#### **POST /validation/static**
Run Layer 1 static validation

```bash
curl -X POST https://api.qlp-hq.com/v1/validation/static \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "package main\n\nfunc main() { ... }",
    "language": "go",
    "validation_rules": {
      "security_checks": ["owasp_top10", "cwe_top25"],
      "quality_checks": ["complexity", "maintainability"],
      "compliance_frameworks": ["SOC2"]
    }
  }'
```

**Response**:
```json
{
  "validation_id": "QLV-001",
  "overall_score": 88,
  "security_score": 92,
  "quality_score": 85,
  "compliance_score": 90,
  "findings": [
    {
      "type": "security",
      "severity": "medium",
      "description": "Potential SQL injection vulnerability",
      "location": "line 42",
      "recommendation": "Use parameterized queries"
    }
  ],
  "deployment_ready": true
}
```

#### **POST /validation/dynamic**
Run Layer 2 dynamic deployment testing

```bash
curl -X POST https://api.qlp-hq.com/v1/validation/dynamic \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "quantum_capsule_id": "QL-CAP-001",
    "test_config": {
      "load_testing": true,
      "security_scanning": true,
      "performance_benchmarking": true
    }
  }'
```

#### **POST /validation/enterprise**
Run Layer 3 enterprise compliance validation

```bash
curl -X POST https://api.qlp-hq.com/v1/validation/enterprise \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "quantum_capsule_id": "QL-CAP-001",
    "compliance_requirements": {
      "frameworks": ["SOC2", "GDPR", "HIPAA"],
      "security_level": "high",
      "performance_targets": {
        "response_time_ms": 200,
        "throughput_rps": 1000,
        "availability_percent": 99.9
      }
    }
  }'
```

---

### **🤖 HITL Decision Engine**

#### **POST /hitl/decisions**
Submit for HITL decision processing

```bash
curl -X POST https://api.qlp-hq.com/v1/hitl/decisions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "quantum_drops": [
      {
        "id": "QD-001",
        "type": "codebase",
        "confidence_score": 94,
        "validation_results": {...}
      }
    ],
    "decision_criteria": {
      "auto_approve_threshold": 90,
      "require_review_below": 70,
      "compliance_required": true
    }
  }'
```

**Response**:
```json
{
  "decision_id": "QLD-001",
  "action": "approve",
  "confidence": 0.94,
  "auto_approved": true,
  "review_required": false,
  "recommendations": [
    "Consider implementing additional monitoring dashboards",
    "Plan for HIPAA compliance if healthcare clients are targeted"
  ],
  "approved_drops": ["QD-001", "QD-002"],
  "rejected_drops": []
}
```

#### **GET /hitl/decisions/{decision_id}**
Get HITL decision status and history

---

### **📦 QuantumCapsule Management**

#### **POST /capsules**
Generate QuantumCapsule from approved QuantumDrops

```bash
curl -X POST https://api.qlp-hq.com/v1/capsules \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "intent_id": "QLI-001",
    "approved_drops": ["QD-001", "QD-002", "QD-003"],
    "capsule_config": {
      "include_source": true,
      "include_tests": true,
      "include_documentation": true,
      "format": "qlcapsule"
    }
  }'
```

**Response**:
```json
{
  "capsule_id": "QL-CAP-e3122edf4be08930",
  "download_url": "https://api.qlp-hq.com/v1/capsules/QL-CAP-e3122edf4be08930/download",
  "metadata": {
    "overall_score": 94,
    "security_risk": "low",
    "quality_score": 92,
    "enterprise_ready": true,
    "compliance_status": {
      "soc2_compliant": true,
      "gdpr_compliant": true,
      "hipaa_compliant": false
    }
  },
  "artifacts": [
    {
      "name": "unified_project",
      "type": "application_code",
      "file_count": 12,
      "size_bytes": 45678
    }
  ]
}
```

#### **GET /capsules/{capsule_id}**
Get QuantumCapsule metadata and status

#### **GET /capsules/{capsule_id}/download**
Download complete QuantumCapsule package

#### **GET /capsules/{capsule_id}/reports**
Get detailed validation and compliance reports

---

## 📊 **Webhooks**

### **Configure Webhooks**

```bash
curl -X POST https://api.qlp-hq.com/v1/webhooks \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/qlp",
    "events": [
      "execution.completed",
      "validation.failed", 
      "capsule.generated",
      "hitl.decision.required"
    ],
    "secret": "your-webhook-secret"
  }'
```

### **Webhook Events**

#### **execution.completed**
```json
{
  "event": "execution.completed",
  "execution_id": "QLE-001",
  "intent_id": "QLI-001",
  "overall_score": 94,
  "status": "success",
  "capsule_id": "QL-CAP-001",
  "timestamp": "2025-06-11T10:00:00Z"
}
```

#### **validation.failed**
```json
{
  "event": "validation.failed", 
  "validation_id": "QLV-001",
  "layer": "static",
  "score": 65,
  "critical_issues": [
    {
      "type": "security",
      "severity": "high",
      "description": "SQL injection vulnerability detected"
    }
  ],
  "timestamp": "2025-06-11T10:00:00Z"
}
```

---

## 🔍 **Query & Analytics**

### **GET /analytics/confidence**
Get confidence score analytics and trends

```bash
curl "https://api.qlp-hq.com/v1/analytics/confidence?period=30d&granularity=daily" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### **GET /analytics/compliance**
Get compliance status across deployments

### **GET /analytics/performance**
Get system performance metrics and benchmarks

---

## ⚙️ **Configuration**

### **GET /config/validation-rules**
Get available validation rules and frameworks

### **POST /config/custom-rules**
Define custom validation rules for your organization

```bash
curl -X POST https://api.qlp-hq.com/v1/config/custom-rules \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "rule_name": "company_security_policy",
    "rule_type": "security",
    "description": "Company-specific security requirements",
    "validation_logic": {
      "required_patterns": ["error handling", "input validation"],
      "forbidden_patterns": ["hardcoded secrets", "eval()"],
      "compliance_frameworks": ["SOC2"]
    }
  }'
```

---

## 📈 **Rate Limits & Quotas**

### **Rate Limits**
- **Free Tier**: 100 requests/hour
- **Professional**: 1,000 requests/hour  
- **Enterprise**: 10,000 requests/hour
- **Enterprise+**: Unlimited

### **Quota Management**
```bash
# Check current usage
curl https://api.qlp-hq.com/v1/usage \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Response**:
```json
{
  "current_period": "2025-06",
  "requests_used": 250,
  "requests_limit": 10000,
  "validations_used": 45,
  "validations_limit": 1000,
  "reset_date": "2025-07-01T00:00:00Z"
}
```

---

## 🛠️ **SDKs & Libraries**

### **Official SDKs**
- 🐹 **Go SDK**: `go get github.com/qlp-hq/qlp-go`
- 🟢 **Node.js SDK**: `npm install @qlp-hq/qlp-js`
- 🐍 **Python SDK**: `pip install qlp-python`
- ☕ **Java SDK**: Maven/Gradle available

### **Go SDK Example**
```go
import "github.com/qlp-hq/qlp-go"

client := qlp.NewClient("YOUR_API_KEY")

result, err := client.ProcessIntent(ctx, &qlp.IntentRequest{
    Intent: "Create secure REST API",
    ValidationLevel: qlp.ValidationLevelEnterprise,
    ComplianceFrameworks: []string{"SOC2", "GDPR"},
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Confidence Score: %d/100\n", result.ConfidenceScore)
```

---

## 🆘 **Error Handling**

### **HTTP Status Codes**
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (invalid API key)
- `403` - Forbidden (quota exceeded)
- `404` - Not Found
- `429` - Rate Limited
- `500` - Internal Server Error

### **Error Response Format**
```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Validation confidence below threshold",
    "details": {
      "confidence_score": 65,
      "minimum_required": 80,
      "critical_issues": [...]
    },
    "request_id": "req_123456789"
  }
}
```

---

## 📞 **Support & Resources**

### **Enterprise Support**
- 📧 **Technical Support**: [api-support@qlp-hq.com](mailto:api-support@qlp-hq.com)
- 📞 **Phone Support**: +1-800-QLP-HELP (Enterprise customers)
- 💬 **Slack Integration**: [#qlp-support](https://qlp-hq.slack.com)

### **Developer Resources**
- 📚 **API Documentation**: [https://docs.qlp-hq.com/api/](https://docs.qlp-hq.com/api/)
- 🧪 **API Playground**: [https://api.qlp-hq.com/playground](https://api.qlp-hq.com/playground)
- 📖 **Postman Collection**: [Download Collection](https://api.qlp-hq.com/postman)

---

**🎖️ Build enterprise-grade applications with 94/100 confidence through our comprehensive API platform!**

[➡️ Try API Playground](/api/playground/) | [📚 View SDK Documentation](/api/sdks/)