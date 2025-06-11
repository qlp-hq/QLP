package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

// LogLevel represents available log levels
type LogLevel string

const (
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
	PANIC LogLevel = "panic"
	FATAL LogLevel = "fatal"
)

// LogFormat represents output formats
type LogFormat string

const (
	JSON    LogFormat = "json"
	CONSOLE LogFormat = "console"
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel  `json:"level"`
	Format     LogFormat `json:"format"`
	OutputPath string    `json:"output_path"`
	Caller     bool      `json:"caller"`
	Stacktrace bool      `json:"stacktrace"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      INFO,
		Format:     CONSOLE,
		OutputPath: "stdout",
		Caller:     true,
		Stacktrace: true,
	}
}

// InitLogger initializes the global logger with configuration
func InitLogger(config Config) error {
	// Determine log level
	var level zapcore.Level
	switch config.Level {
	case DEBUG:
		level = zapcore.DebugLevel
	case INFO:
		level = zapcore.InfoLevel
	case WARN:
		level = zapcore.WarnLevel
	case ERROR:
		level = zapcore.ErrorLevel
	case PANIC:
		level = zapcore.PanicLevel
	case FATAL:
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// Configure encoder
	var encoderConfig zapcore.EncoderConfig
	var encoder zapcore.Encoder

	if config.Format == JSON {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05")
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	if config.OutputPath == "stdout" || config.OutputPath == "" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	} else {
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Build logger with options
	var options []zap.Option
	if config.Caller {
		options = append(options, zap.AddCaller())
		options = append(options, zap.AddCallerSkip(1))
	}
	if config.Stacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	Logger = zap.New(core, options...)
	Sugar = Logger.Sugar()

	return nil
}

// InitFromEnv initializes logger from environment variables
func InitFromEnv() error {
	config := DefaultConfig()

	// Override from environment
	if level := os.Getenv("QLP_LOG_LEVEL"); level != "" {
		config.Level = LogLevel(strings.ToLower(level))
	}
	if format := os.Getenv("QLP_LOG_FORMAT"); format != "" {
		config.Format = LogFormat(strings.ToLower(format))
	}
	if output := os.Getenv("QLP_LOG_OUTPUT"); output != "" {
		config.OutputPath = output
	}
	if caller := os.Getenv("QLP_LOG_CALLER"); caller == "false" {
		config.Caller = false
	}
	if stacktrace := os.Getenv("QLP_LOG_STACKTRACE"); stacktrace == "false" {
		config.Stacktrace = false
	}

	return InitLogger(config)
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Context-aware logging helpers

// WithComponent adds component context to logger
func WithComponent(component string) *zap.Logger {
	return Logger.With(zap.String("component", component))
}

// WithAgent adds agent context to logger
func WithAgent(agentID string) *zap.Logger {
	return Logger.With(zap.String("agent_id", agentID))
}

// WithTask adds task context to logger
func WithTask(taskID string) *zap.Logger {
	return Logger.With(zap.String("task_id", taskID))
}

// WithIntent adds intent context to logger
func WithIntent(intentID string) *zap.Logger {
	return Logger.With(zap.String("intent_id", intentID))
}

// WithValidation adds validation context to logger
func WithValidation(taskID string, score int) *zap.Logger {
	return Logger.With(
		zap.String("task_id", taskID),
		zap.Int("validation_score", score),
	)
}

// WithExecution adds execution context to logger
func WithExecution(agentID, taskID string) *zap.Logger {
	return Logger.With(
		zap.String("agent_id", agentID),
		zap.String("task_id", taskID),
	)
}

// WithError adds error context to logger
func WithError(err error) *zap.Logger {
	return Logger.With(zap.Error(err))
}

// Performance logging helpers

// LogPerformance logs performance metrics
func LogPerformance(operation string, duration int64, success bool) {
	Logger.Info("Performance metric",
		zap.String("operation", operation),
		zap.Int64("duration_ms", duration),
		zap.Bool("success", success),
	)
}

// LogAgentMetrics logs agent execution metrics
func LogAgentMetrics(agentID, taskID string, executionTime int64, validationScore int, success bool) {
	Logger.Info("Agent execution completed",
		zap.String("agent_id", agentID),
		zap.String("task_id", taskID),
		zap.Int64("execution_time_ms", executionTime),
		zap.Int("validation_score", validationScore),
		zap.Bool("success", success),
	)
}

// LogIntentMetrics logs intent processing metrics
func LogIntentMetrics(intentID string, taskCount int, totalTime int64, overallScore int) {
	Logger.Info("Intent processing completed",
		zap.String("intent_id", intentID),
		zap.Int("task_count", taskCount),
		zap.Int64("total_time_ms", totalTime),
		zap.Int("overall_score", overallScore),
	)
}

// LogValidationMetrics logs validation metrics
func LogValidationMetrics(taskID string, syntaxScore, securityScore, qualityScore, overallScore int, passed bool) {
	Logger.Info("Validation completed",
		zap.String("task_id", taskID),
		zap.Int("syntax_score", syntaxScore),
		zap.Int("security_score", securityScore),
		zap.Int("quality_score", qualityScore),
		zap.Int("overall_score", overallScore),
		zap.Bool("passed", passed),
	)
}

// Structured error logging

// LogError logs structured error information
func LogError(operation string, err error, context map[string]interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Error(err),
	}
	
	for key, value := range context {
		fields = append(fields, zap.Any(key, value))
	}
	
	Logger.Error("Operation failed", fields...)
}

// LogCriticalError logs critical system errors
func LogCriticalError(operation string, err error, context map[string]interface{}) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.Error(err),
		zap.String("severity", "critical"),
	}
	
	for key, value := range context {
		fields = append(fields, zap.Any(key, value))
	}
	
	Logger.Error("Critical system error", fields...)
}