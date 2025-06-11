# Comprehensive Logging System for QLP

## Overview

QLP uses **Zap** (go.uber.org/zap) for high-performance structured logging, providing enterprise-grade logging capabilities for agent orchestration, validation, and system monitoring.

## Features

✅ **High Performance**: Zero-allocation JSON encoder  
✅ **Structured Logging**: Consistent field-based logging  
✅ **Multiple Formats**: JSON (production) and Console (development)  
✅ **Log Levels**: DEBUG, INFO, WARN, ERROR, PANIC, FATAL  
✅ **Context-Aware**: Component, agent, task, and intent contexts  
✅ **Performance Metrics**: Built-in performance and metrics logging  
✅ **Environment Configuration**: Configurable via environment variables  

## Configuration

### Environment Variables

```bash
# Log Level (debug, info, warn, error, panic, fatal)
QLP_LOG_LEVEL=info

# Output Format (console, json)
QLP_LOG_FORMAT=console

# Output Destination (stdout, stderr, or file path)
QLP_LOG_OUTPUT=stdout

# Include caller information (true, false)
QLP_LOG_CALLER=true

# Include stack traces on errors (true, false)
QLP_LOG_STACKTRACE=true
```

### Development Configuration

```bash
QLP_LOG_LEVEL=debug
QLP_LOG_FORMAT=console
QLP_LOG_OUTPUT=stdout
QLP_LOG_CALLER=true
QLP_LOG_STACKTRACE=true
```

### Production Configuration

```bash
QLP_LOG_LEVEL=info
QLP_LOG_FORMAT=json
QLP_LOG_OUTPUT=/var/log/qlp/app.log
QLP_LOG_CALLER=false
QLP_LOG_STACKTRACE=false
```

## Usage Examples

### Basic Logging

```go
import "QLP/internal/logger"

// Initialize logger (done automatically in main.go)
logger.InitFromEnv()

// Basic logging
logger.Logger.Info("System started")
logger.Logger.Debug("Debug information")
logger.Logger.Warn("Warning message")
logger.Logger.Error("Error occurred")
```

### Context-Aware Logging

```go
// Component-specific logging
compLogger := logger.WithComponent("orchestrator")
compLogger.Info("Component initialized")

// Agent-specific logging
agentLogger := logger.WithAgent("QLD-AGT-123456")
agentLogger.Info("Agent created", zap.String("task_type", "codegen"))

// Task-specific logging
taskLogger := logger.WithTask("QL-DEV-001")
taskLogger.Info("Task started", zap.Duration("estimated_duration", time.Minute*2))

// Multiple contexts
logger.WithExecution("QLD-AGT-123", "QL-DEV-001").Info("Task execution started")
```

### Performance Logging

```go
// Performance metrics
logger.LogPerformance("agent_execution", 2500, true) // operation, duration_ms, success

// Agent metrics
logger.LogAgentMetrics("QLD-AGT-123", "QL-DEV-001", 2500, 85, true)
// agentID, taskID, execution_time_ms, validation_score, success

// Intent metrics
logger.LogIntentMetrics("QLI-12345", 8, 45000, 92)
// intentID, task_count, total_time_ms, overall_score

// Validation metrics
logger.LogValidationMetrics("QL-DEV-001", 90, 75, 88, 84, true)
// taskID, syntax_score, security_score, quality_score, overall_score, passed
```

### Error Logging

```go
// Simple error logging
logger.WithError(err).Error("Database connection failed")

// Structured error logging
logger.LogError("database_connection", err, map[string]interface{}{
    "database": "postgresql",
    "host":     "localhost",
    "port":     5432,
    "retry_count": 3,
})

// Critical errors
logger.LogCriticalError("system_failure", err, map[string]interface{}{
    "component": "orchestrator",
    "intent_id": "QLI-12345",
})
```

## Log Output Examples

### Console Format (Development)

```
2025/06/11 12:45:43	INFO	orchestrator/orchestrator.go:145	Intent processing started	{"intent_id": "QLI-12345", "task_count": 8}
2025/06/11 12:45:44	INFO	agents/dynamic_agent.go:94	Agent execution started	{"agent_id": "QLD-AGT-123", "task_id": "QL-DEV-001"}
2025/06/11 12:45:46	INFO	validation/validator.go:134	Validation completed	{"task_id": "QL-DEV-001", "overall_score": 85, "passed": true}
```

### JSON Format (Production)

```json
{"level":"info","timestamp":"2025-06-11T12:45:43.123Z","caller":"orchestrator/orchestrator.go:145","msg":"Intent processing started","intent_id":"QLI-12345","task_count":8}
{"level":"info","timestamp":"2025-06-11T12:45:44.456Z","caller":"agents/dynamic_agent.go:94","msg":"Agent execution started","agent_id":"QLD-AGT-123","task_id":"QL-DEV-001"}
{"level":"info","timestamp":"2025-06-11T12:45:46.789Z","caller":"validation/validator.go:134","msg":"Validation completed","task_id":"QL-DEV-001","overall_score":85,"passed":true}
```

## Integration with Monitoring

### Log Aggregation

**ELK Stack (Elasticsearch, Logstash, Kibana)**
```bash
# Ship JSON logs to Logstash
QLP_LOG_FORMAT=json
QLP_LOG_OUTPUT=/var/log/qlp/app.log
```

**Prometheus Metrics**
```go
// Custom metrics from logs
logger.Logger.Info("metric",
    zap.String("metric_type", "counter"),
    zap.String("metric_name", "tasks_completed"),
    zap.Float64("value", 1.0),
)
```

**Grafana Dashboards**
- Agent performance metrics
- Task completion rates
- Validation score trends
- Error rates by component

### Alerting

**Critical Error Alerts**
```go
logger.LogCriticalError("agent_failure", err, map[string]interface{}{
    "alert": "immediate",
    "severity": "high",
})
```

## Best Practices

### 1. Use Structured Fields
```go
// Good
logger.Logger.Info("Task completed",
    zap.String("task_id", taskID),
    zap.Int("score", score),
    zap.Duration("duration", duration),
)

// Avoid
logger.Logger.Infof("Task %s completed with score %d in %v", taskID, score, duration)
```

### 2. Use Appropriate Log Levels
- **DEBUG**: Detailed execution flow, variable values
- **INFO**: Important events, state changes, metrics
- **WARN**: Recoverable errors, degraded performance
- **ERROR**: Operation failures, exceptions
- **PANIC/FATAL**: System-critical failures

### 3. Include Context
```go
// Always include relevant context
logger.WithTask(taskID).WithAgent(agentID).Info("Execution started")
```

### 4. Log Performance Metrics
```go
start := time.Now()
// ... do work ...
logger.LogPerformance("operation_name", time.Since(start).Milliseconds(), success)
```

## Security Considerations

### Sensitive Data
```go
// Never log sensitive information
logger.Logger.Info("User authenticated",
    zap.String("user_id", userID),
    // DON'T: zap.String("password", password),
    // DON'T: zap.String("api_key", apiKey),
)
```

### Log Sanitization
```go
func sanitizeForLogging(data string) string {
    // Remove or mask sensitive patterns
    re := regexp.MustCompile(`(?i)(password|key|token)=\S+`)
    return re.ReplaceAllString(data, "${1}=***")
}
```

## Deployment Configurations

### Docker Compose
```yaml
services:
  qlp:
    environment:
      - QLP_LOG_LEVEL=info
      - QLP_LOG_FORMAT=json
      - QLP_LOG_OUTPUT=/app/logs/qlp.log
    volumes:
      - ./logs:/app/logs
```

### Kubernetes
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: qlp-logging-config
data:
  QLP_LOG_LEVEL: "info"
  QLP_LOG_FORMAT: "json"
  QLP_LOG_OUTPUT: "stdout"
```

## Troubleshooting

### High Disk Usage
```bash
# Rotate logs
QLP_LOG_OUTPUT=/var/log/qlp/app-$(date +%Y%m%d).log

# Or use log rotation tools
logrotate /etc/logrotate.d/qlp
```

### Performance Impact
```bash
# Reduce log level in production
QLP_LOG_LEVEL=warn

# Disable caller information for performance
QLP_LOG_CALLER=false
```

### Debug Mode
```bash
# Enable debug logging for troubleshooting
QLP_LOG_LEVEL=debug
QLP_LOG_FORMAT=console
QLP_LOG_CALLER=true
```