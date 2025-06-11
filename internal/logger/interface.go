package logger

import "go.uber.org/zap"

// Interface defines logging methods to avoid conflict with global Logger variable
type Interface interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	WithComponent(component string) Interface
}

// ZapLogger wraps zap.Logger to implement Interface
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger
func NewZapLogger(zapLogger *zap.Logger) Interface {
	return &ZapLogger{logger: zapLogger}
}

// Info logs an info message
func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

// Warn logs a warning message
func (z *ZapLogger) Warn(msg string, fields ...zap.Field) {
	z.logger.Warn(msg, fields...)
}

// Error logs an error message
func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}

// Debug logs a debug message
func (z *ZapLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}

// WithComponent adds a component field to the logger
func (z *ZapLogger) WithComponent(component string) Interface {
	return &ZapLogger{
		logger: z.logger.With(zap.String("component", component)),
	}
}

// GetDefaultLogger returns a logger using the global zap logger
func GetDefaultLogger() Interface {
	return NewZapLogger(Logger)
}