# 🔧 QuantumLayer Platform - Complete API Reference

**Enterprise-Grade API Documentation for AI Agent Orchestration**

---

## 📋 Table of Contents

1. [API Overview](#api-overview)
2. [Authentication & Security](#authentication--security)
3. [Core Endpoints](#core-endpoints)
4. [Intent Processing API](#intent-processing-api)
5. [QuantumCapsule API](#quantumcapsule-api)
6. [Agent Management API](#agent-management-api)
7. [Validation API](#validation-api)
8. [System Management API](#system-management-api)
9. [Webhook API](#webhook-api)
10. [SDK Documentation](#sdk-documentation)
11. [Error Handling](#error-handling)
12. [Rate Limiting](#rate-limiting)

---

## 🌐 API Overview

The QuantumLayer Platform exposes a comprehensive REST API for integrating AI agent orchestration into enterprise systems. The API follows RESTful principles with JSON payloads and standard HTTP status codes.

### Base Configuration

```yaml
# API Configuration
api:
  base_url: "https://api.quantumlayer.com/v1"
  version: "1.0"
  content_type: "application/json"
  timeout: 30s
  rate_limit: 1000/hour
  
# Local Development
dev:
  base_url: "http://localhost:8080/api/v1"
  
# Enterprise On-Premises
enterprise:
  base_url: "https://qlp.yourcompany.com/api/v1"
```

### API Versioning

| **Version** | **Status** | **Support** | **Features** |
|-------------|------------|-------------|--------------|
| `v1` | **Current** | Full Support | Complete feature set |
| `v2` | Beta | Preview | Enhanced validation |
| `v0` | Deprecated | Security fixes only | Legacy support |

### Global Request/Response Format

```json
{
  "success": true,
  "data": {},
  "metadata": {
    "request_id": "req_123abc456def",
    "timestamp": "2024-06-13T12:34:56Z",
    "version": "v1",
    "processing_time_ms": 150
  },
  "errors": []
}
```

---

## 🔐 Authentication & Security

### API Key Authentication

```bash
# Header-based authentication
curl -H "Authorization: Bearer qlp_live_1234567890abcdef" \
     -H "Content-Type: application/json" \
     https://api.quantumlayer.com/v1/intents
```

### Authentication Methods

| **Method** | **Header** | **Format** | **Use Case** |
|------------|------------|------------|--------------|
| **API Key** | `Authorization: Bearer <key>` | `qlp_live_*` | Production |
| **Development** | `Authorization: Bearer <key>` | `qlp_dev_*` | Development |
| **Enterprise** | `Authorization: Bearer <token>` | JWT Token | SSO Integration |

### API Key Management

```json
POST /api/v1/auth/keys
{
  "name": "Production Integration",
  "permissions": ["intents:create", "capsules:read"],
  "expires_at": "2024-12-31T23:59:59Z",
  "ip_whitelist": ["192.168.1.0/24"],
  "rate_limit": 10000
}

Response:
{
  "success": true,
  "data": {
    "key_id": "key_abc123",
    "api_key": "qlp_live_xyz789...",
    "created_at": "2024-06-13T12:34:56Z",
    "permissions": ["intents:create", "capsules:read"]
  }
}
```

---

## 🎯 Core Endpoints

### System Information

```bash
GET /api/v1/system/info
```

```json
{
  "success": true,
  "data": {
    "system": {
      "name": "QuantumLayer Platform",
      "version": "2.1.0",
      "build": "20240613-1234",
      "environment": "production"
    },
    "capabilities": {
      "max_concurrent_intents": 100,
      "supported_languages": ["go", "python", "javascript", "java"],
      "validation_layers": 3,
      "compliance_frameworks": ["SOC2", "GDPR", "HIPAA"]
    },
    "status": {
      "healthy": true,
      "agents_active": 15,
      "queue_depth": 3
    }
  }
}
```

### Health Check

```bash
GET /api/v1/health
```

```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-06-13T12:34:56Z",
    "checks": {
      "database": "healthy",
      "llm_providers": "healthy",
      "agent_factory": "healthy",
      "validation_engine": "healthy"
    },
    "metrics": {
      "uptime_seconds": 86400,
      "processed_intents": 1337,
      "success_rate": 0.94
    }
  }
}
```

---

## 🎯 Intent Processing API

### Create Intent

Submit a natural language intent for processing.

```bash
POST /api/v1/intents
```

**Request Body:**
```json
{
  "text": "Create a secure REST API for user management with JWT authentication",
  "options": {
    "validation_level": "enterprise",
    "compliance_frameworks": ["SOC2", "GDPR"],
    "target_confidence": 90,
    "include_tests": true,
    "include_documentation": true,
    "deployment_target": "kubernetes"
  },
  "metadata": {
    "project_name": "UserAPI",
    "organization": "ACME Corp",
    "environment": "production"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "intent_id": "QLI-1749814983847612000",
    "status": "processing",
    "created_at": "2024-06-13T12:34:56Z",
    "estimated_completion": "2024-06-13T12:37:56Z",
    "tasks": [
      {
        "id": "QL-DEV-20240613-001",
        "type": "codegen",
        "description": "Set up Go project structure with REST API framework",
        "status": "pending",
        "dependencies": []
      },
      {
        "id": "QL-DEV-20240613-002", 
        "type": "codegen",
        "description": "Implement JWT authentication middleware",
        "status": "pending",
        "dependencies": ["QL-DEV-20240613-001"]
      }
    ]
  },
  "metadata": {
    "request_id": "req_abc123",
    "processing_time_ms": 250
  }
}
```

### Get Intent Status

```bash
GET /api/v1/intents/{intent_id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "intent_id": "QLI-1749814983847612000",
    "text": "Create a secure REST API for user management with JWT authentication",
    "status": "completed",
    "created_at": "2024-06-13T12:34:56Z",
    "completed_at": "2024-06-13T12:37:23Z",
    "processing_time": "2m27s",
    "progress": {
      "total_tasks": 15,
      "completed_tasks": 15,
      "failed_tasks": 0,
      "percentage": 100
    },
    "validation": {
      "overall_score": 94,
      "security_score": 100,
      "quality_score": 91,
      "compliance_score": 96,
      "enterprise_ready": true
    },
    "capsule": {
      "id": "QL-CAP-34da40246951338d",
      "status": "ready",
      "download_url": "/api/v1/capsules/QL-CAP-34da40246951338d/download",
      "size_bytes": 35612
    }
  }
}
```

### List Intents

```bash
GET /api/v1/intents?limit=20&offset=0&status=completed&sort=created_at:desc
```

**Query Parameters:**
- `limit`: Number of results (default: 20, max: 100)
- `offset`: Pagination offset
- `status`: Filter by status (`pending`, `processing`, `completed`, `failed`)
- `sort`: Sort order (`created_at:asc|desc`, `completion_time:asc|desc`)
- `search`: Search in intent text
- `from_date`: Filter from date (ISO 8601)
- `to_date`: Filter to date (ISO 8601)

**Response:**
```json
{
  "success": true,
  "data": {
    "intents": [
      {
        "intent_id": "QLI-1749814983847612000",
        "text": "Create a secure REST API...",
        "status": "completed",
        "created_at": "2024-06-13T12:34:56Z",
        "validation_score": 94,
        "capsule_id": "QL-CAP-34da40246951338d"
      }
    ],
    "pagination": {
      "total": 157,
      "limit": 20,
      "offset": 0,
      "has_more": true
    }
  }
}
```

### Cancel Intent

```bash
DELETE /api/v1/intents/{intent_id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "intent_id": "QLI-1749814983847612000",
    "status": "cancelled",
    "cancelled_at": "2024-06-13T12:35:30Z",
    "reason": "User requested cancellation"
  }
}
```

---

## 📦 QuantumCapsule API

### Get QuantumCapsule

```bash
GET /api/v1/capsules/{capsule_id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "QL-CAP-34da40246951338d",
    "intent_id": "QLI-1749814983847612000",
    "created_at": "2024-06-13T12:37:23Z",
    "metadata": {
      "name": "user-api",
      "description": "Secure REST API for user management with JWT authentication",
      "version": "1.0.0",
      "technologies": ["go", "gin", "jwt", "postgresql"],
      "security_risk": "none",
      "quality_score": 91,
      "compliance": {
        "SOC2": true,
        "GDPR": true,
        "HIPAA": false
      }
    },
    "projects": {
      "user-api": {
        "name": "user-api",
        "language": "go",
        "framework": "gin",
        "files": [
          {
            "path": "main.go",
            "type": "source",
            "size": 1024
          },
          {
            "path": "handlers/auth.go", 
            "type": "source",
            "size": 2048
          },
          {
            "path": "Dockerfile",
            "type": "config",
            "size": 512
          }
        ],
        "dependencies": [
          "github.com/gin-gonic/gin v1.9.1",
          "github.com/golang-jwt/jwt/v5 v5.0.0"
        ]
      }
    },
    "validation": {
      "overall_score": 94,
      "layers": {
        "static_analysis": 96,
        "dynamic_testing": 92,
        "enterprise_compliance": 94
      },
      "security_scan": {
        "vulnerabilities": 0,
        "warnings": 1,
        "score": 100
      }
    },
    "size_bytes": 35612,
    "download_urls": {
      "zip": "/api/v1/capsules/QL-CAP-34da40246951338d/download?format=zip",
      "tar_gz": "/api/v1/capsules/QL-CAP-34da40246951338d/download?format=tar.gz"
    }
  }
}
```

### Download QuantumCapsule

```bash
GET /api/v1/capsules/{capsule_id}/download?format=zip
```

**Query Parameters:**
- `format`: Download format (`zip`, `tar.gz`, `json`)
- `include_source`: Include source files (default: true)
- `include_tests`: Include test files (default: true)
- `include_docs`: Include documentation (default: true)

**Response:** Binary file download with appropriate headers.

### List QuantumCapsules

```bash
GET /api/v1/capsules?limit=20&sort=created_at:desc&min_score=90
```

**Response:**
```json
{
  "success": true,
  "data": {
    "capsules": [
      {
        "id": "QL-CAP-34da40246951338d",
        "intent_id": "QLI-1749814983847612000",
        "name": "user-api",
        "created_at": "2024-06-13T12:37:23Z",
        "quality_score": 91,
        "size_bytes": 35612,
        "technologies": ["go", "gin", "jwt"]
      }
    ],
    "pagination": {
      "total": 89,
      "limit": 20,
      "offset": 0
    }
  }
}
```

### Delete QuantumCapsule

```bash
DELETE /api/v1/capsules/{capsule_id}
```

---

## 🤖 Agent Management API

### List Active Agents

```bash
GET /api/v1/agents
```

**Response:**
```json
{
  "success": true,
  "data": {
    "agents": [
      {
        "id": "QLD-AGT-124303-000",
        "type": "codegen",
        "status": "active",
        "task_id": "QL-DEV-20240613-001",
        "intent_id": "QLI-1749814983847612000",
        "created_at": "2024-06-13T12:35:00Z",
        "progress": 75,
        "llm_requests": 3,
        "execution_time": "45s"
      }
    ],
    "summary": {
      "total_agents": 5,
      "active_agents": 3,
      "idle_agents": 2
    }
  }
}
```

### Get Agent Details

```bash
GET /api/v1/agents/{agent_id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "QLD-AGT-124303-000",
    "type": "codegen",
    "status": "active",
    "task": {
      "id": "QL-DEV-20240613-001",
      "description": "Set up Go project structure with REST API framework",
      "progress": 75
    },
    "capabilities": {
      "languages": ["go", "python"],
      "frameworks": ["gin", "fastapi"],
      "specializations": ["api_development", "authentication"]
    },
    "metrics": {
      "llm_requests": 3,
      "execution_time": "45s",
      "memory_usage": "128MB",
      "success_rate": 0.95
    },
    "current_operation": {
      "step": "code_generation",
      "estimated_completion": "2024-06-13T12:36:00Z"
    }
  }
}
```

### Terminate Agent

```bash
DELETE /api/v1/agents/{agent_id}
```

---

## ✅ Validation API

### Validate Code

Submit code for validation outside of the normal intent flow.

```bash
POST /api/v1/validation/code
```

**Request:**
```json
{
  "code": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello World\")\n}",
  "language": "go",
  "validation_level": "enterprise",
  "compliance_frameworks": ["SOC2"],
  "context": {
    "project_type": "api",
    "security_requirements": "high"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "validation_id": "VAL-abc123def456",
    "overall_score": 87,
    "layers": {
      "static_analysis": {
        "score": 90,
        "issues": [
          {
            "type": "warning",
            "line": 5,
            "message": "Consider adding error handling",
            "severity": "medium"
          }
        ]
      },
      "security_scan": {
        "score": 100,
        "vulnerabilities": [],
        "compliance": {
          "SOC2": true
        }
      }
    },
    "recommendations": [
      "Add comprehensive error handling",
      "Include input validation",
      "Add logging for audit trails"
    ]
  }
}
```

### Get Validation Result

```bash
GET /api/v1/validation/{validation_id}
```

### Validate Project

```bash
POST /api/v1/validation/project
```

**Request:** Multipart form with project files.

---

## ⚙️ System Management API

### Get System Metrics

```bash
GET /api/v1/system/metrics
```

**Response:**
```json
{
  "success": true,
  "data": {
    "performance": {
      "intents_per_minute": 15.7,
      "average_processing_time": "2m34s",
      "success_rate": 0.94,
      "error_rate": 0.02
    },
    "resources": {
      "cpu_usage": 45.2,
      "memory_usage": 2.1,
      "disk_usage": 67.8,
      "active_connections": 23
    },
    "business": {
      "total_intents": 15437,
      "total_capsules": 14562,
      "average_confidence": 91.3,
      "enterprise_compliance": 96.8
    }
  }
}
```

### System Configuration

```bash
GET /api/v1/system/config
PUT /api/v1/system/config
```

### Log Management

```bash
GET /api/v1/system/logs?level=error&from=2024-06-13T00:00:00Z&limit=100
```

---

## 🔔 Webhook API

### Create Webhook

```bash
POST /api/v1/webhooks
```

**Request:**
```json
{
  "url": "https://api.yourcompany.com/qlp-webhooks",
  "events": ["intent.completed", "capsule.generated", "validation.failed"],
  "secret": "webhook_secret_key",
  "headers": {
    "Authorization": "Bearer your-token"
  },
  "retry_config": {
    "max_attempts": 3,
    "backoff_seconds": 60
  }
}
```

### Webhook Events

| **Event** | **Description** | **Payload** |
|-----------|-----------------|-------------|
| `intent.created` | New intent submitted | Intent details |
| `intent.processing` | Intent processing started | Progress update |
| `intent.completed` | Intent processing finished | Full results |
| `intent.failed` | Intent processing failed | Error details |
| `capsule.generated` | QuantumCapsule created | Capsule metadata |
| `validation.completed` | Validation finished | Validation results |
| `agent.created` | New agent spawned | Agent details |
| `system.alert` | System alert triggered | Alert information |

### Webhook Payload Example

```json
{
  "event": "intent.completed",
  "timestamp": "2024-06-13T12:37:23Z",
  "data": {
    "intent_id": "QLI-1749814983847612000",
    "status": "completed",
    "validation_score": 94,
    "capsule_id": "QL-CAP-34da40246951338d",
    "processing_time": "2m27s"
  },
  "signature": "sha256=abc123..."
}
```

---

## 🛠️ SDK Documentation

### Go SDK

```go
package main

import (
    "context"
    "fmt"
    "github.com/quantumlayer/qlp-go-sdk"
)

func main() {
    client := qlp.NewClient("qlp_live_your_api_key")
    
    // Submit intent
    intent, err := client.Intents.Create(context.Background(), &qlp.IntentRequest{
        Text: "Create a secure REST API for user management",
        Options: &qlp.IntentOptions{
            ValidationLevel: "enterprise",
            IncludeTests:   true,
        },
    })
    if err != nil {
        panic(err)
    }
    
    // Wait for completion
    result, err := client.Intents.WaitForCompletion(context.Background(), intent.ID)
    if err != nil {
        panic(err)
    }
    
    // Download capsule
    capsule, err := client.Capsules.Download(context.Background(), result.CapsuleID)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Downloaded capsule: %d bytes\n", len(capsule))
}
```

### Python SDK

```python
from quantumlayer import QLPClient

client = QLPClient(api_key="qlp_live_your_api_key")

# Submit intent
intent = client.intents.create(
    text="Create a secure REST API for user management",
    options={
        "validation_level": "enterprise",
        "include_tests": True
    }
)

# Wait for completion
result = client.intents.wait_for_completion(intent.id)

# Download capsule
capsule_data = client.capsules.download(result.capsule_id)
print(f"Downloaded capsule: {len(capsule_data)} bytes")
```

### JavaScript/Node.js SDK

```javascript
const { QLPClient } = require('@quantumlayer/qlp-js');

const client = new QLPClient({ apiKey: 'qlp_live_your_api_key' });

async function processIntent() {
    // Submit intent
    const intent = await client.intents.create({
        text: 'Create a secure REST API for user management',
        options: {
            validationLevel: 'enterprise',
            includeTests: true
        }
    });
    
    // Wait for completion
    const result = await client.intents.waitForCompletion(intent.id);
    
    // Download capsule
    const capsuleData = await client.capsules.download(result.capsuleId);
    console.log(`Downloaded capsule: ${capsuleData.length} bytes`);
}

processIntent().catch(console.error);
```

---

## 🚨 Error Handling

### Standard Error Response

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Intent validation failed due to security concerns",
    "details": {
      "validation_score": 45,
      "issues": [
        {
          "type": "security",
          "severity": "high",
          "message": "Potential SQL injection vulnerability detected"
        }
      ]
    },
    "request_id": "req_abc123",
    "timestamp": "2024-06-13T12:34:56Z"
  }
}
```

### Error Codes

| **Code** | **HTTP Status** | **Description** |
|----------|----------------|-----------------|
| `INVALID_REQUEST` | 400 | Malformed request payload |
| `UNAUTHORIZED` | 401 | Invalid or missing API key |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `RATE_LIMITED` | 429 | Rate limit exceeded |
| `VALIDATION_FAILED` | 422 | Content validation failed |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `LLM_UNAVAILABLE` | 503 | LLM provider unavailable |
| `TIMEOUT` | 504 | Request timeout |

### Retry Guidelines

```javascript
// Exponential backoff with jitter
function retryWithBackoff(apiCall, maxRetries = 3) {
    return new Promise((resolve, reject) => {
        function attempt(retryCount) {
            apiCall()
                .then(resolve)
                .catch(error => {
                    if (retryCount >= maxRetries || error.status < 500) {
                        reject(error);
                        return;
                    }
                    
                    const backoffMs = Math.min(1000 * Math.pow(2, retryCount), 30000);
                    const jitter = Math.random() * 1000;
                    
                    setTimeout(() => attempt(retryCount + 1), backoffMs + jitter);
                });
        }
        
        attempt(0);
    });
}
```

---

## 🚦 Rate Limiting

### Rate Limit Headers

All API responses include rate limiting headers:

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 987
X-RateLimit-Reset: 1623456789
X-RateLimit-Window: 3600
```

### Rate Limit Tiers

| **Tier** | **Requests/Hour** | **Concurrent Intents** | **Burst Limit** |
|----------|-------------------|------------------------|------------------|
| **Development** | 100 | 1 | 10 |
| **Professional** | 1,000 | 5 | 50 |
| **Enterprise** | 10,000 | 20 | 200 |
| **Enterprise+** | 100,000 | 100 | 1,000 |

### Rate Limit Handling

```python
import time
from quantumlayer import QLPClient, RateLimitError

client = QLPClient(api_key="your_api_key")

def make_request_with_retry():
    max_retries = 3
    retry_count = 0
    
    while retry_count < max_retries:
        try:
            return client.intents.create(text="Your intent here")
        except RateLimitError as e:
            if retry_count >= max_retries - 1:
                raise
            
            # Wait for rate limit reset
            wait_time = int(e.headers.get('X-RateLimit-Reset', 60))
            time.sleep(wait_time)
            retry_count += 1
```

---

## 🎯 Best Practices

### 1. API Key Security

- Store API keys in environment variables
- Use different keys for development and production
- Rotate keys regularly
- Monitor key usage for anomalies

### 2. Request Optimization

- Use appropriate timeout values
- Implement exponential backoff for retries
- Cache responses when appropriate
- Use webhooks instead of polling

### 3. Error Handling

- Always check the `success` field in responses
- Log error details for debugging
- Implement graceful degradation
- Use specific error codes for different handling

### 4. Performance

- Process intents asynchronously
- Use pagination for large result sets
- Implement request batching where possible
- Monitor API response times

---

## 🔗 Additional Resources

- **API Status Page**: [status.quantumlayer.com](https://status.quantumlayer.com)
- **Postman Collection**: [Download Collection](https://www.postman.com/quantumlayer/workspace)
- **OpenAPI Specification**: [Download OpenAPI](https://api.quantumlayer.com/openapi.json)
- **SDK GitHub Repositories**: [github.com/quantumlayer](https://github.com/quantumlayer)

---

**🎖️ Transform your development integration from impressive to absolutely bulletproof with the QuantumLayer API!**

*For enterprise support and custom integrations, contact [enterprise@qlp-hq.com](mailto:enterprise@qlp-hq.com)*