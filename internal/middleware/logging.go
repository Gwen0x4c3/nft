package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoggingConfig defines configuration for request logging middleware
type LoggingConfig struct {
	Logger       *zap.Logger
	SkipPaths    []string          // Paths to skip logging (e.g., health checks)
	SkipHeaders  []string          // Headers to exclude from logs
	LogBody      bool              // Whether to log request/response bodies
	MaxBodySize  int64             // Maximum body size to log (bytes)
	LogFields    []string          // Additional fields to log
	Formatter    LogFormatter      // Custom log formatter
	SensitiveHeaders []string      // Headers to redact in logs
}

// LogFormatter defines the interface for custom log formatting
type LogFormatter interface {
	Format(*LogEntry) map[string]interface{}
}

// LogEntry contains all the information about a request/response
type LogEntry struct {
	RequestID     string            `json:"request_id"`
	Timestamp     time.Time         `json:"timestamp"`
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Query         string            `json:"query,omitempty"`
	ClientIP      string            `json:"client_ip"`
	UserAgent     string            `json:"user_agent,omitempty"`
	StatusCode    int               `json:"status_code"`
	Latency       time.Duration     `json:"latency"`
	RequestSize   int64             `json:"request_size"`
	ResponseSize  int64             `json:"response_size"`
	RequestHeaders map[string]string `json:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	RequestBody   string            `json:"request_body,omitempty"`
	ResponseBody  string            `json:"response_body,omitempty"`
	UserID        *uint             `json:"user_id,omitempty"`
	Username      string            `json:"username,omitempty"`
	Error         string            `json:"error,omitempty"`
	TraceID       string            `json:"trace_id,omitempty"`
}

// ResponseWriter wraps gin.ResponseWriter to capture response data
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	size int64
}

// Write captures the response body
func (w *ResponseWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		w.body.Write(b)
	}
	w.size += int64(len(b))
	return w.ResponseWriter.Write(b)
}

// DefaultLogFormatter implements a standard JSON log format
type DefaultLogFormatter struct{}

func (f *DefaultLogFormatter) Format(entry *LogEntry) map[string]interface{} {
	fields := map[string]interface{}{
		"request_id":    entry.RequestID,
		"timestamp":     entry.Timestamp.Format(time.RFC3339Nano),
		"method":        entry.Method,
		"path":          entry.Path,
		"client_ip":     entry.ClientIP,
		"status_code":   entry.StatusCode,
		"latency_ms":    entry.Latency.Milliseconds(),
		"request_size":  entry.RequestSize,
		"response_size": entry.ResponseSize,
	}

	if entry.Query != "" {
		fields["query"] = entry.Query
	}
	if entry.UserAgent != "" {
		fields["user_agent"] = entry.UserAgent
	}
	if entry.UserID != nil {
		fields["user_id"] = *entry.UserID
	}
	if entry.Username != "" {
		fields["username"] = entry.Username
	}
	if entry.Error != "" {
		fields["error"] = entry.Error
	}
	if entry.TraceID != "" {
		fields["trace_id"] = entry.TraceID
	}
	if len(entry.RequestHeaders) > 0 {
		fields["request_headers"] = entry.RequestHeaders
	}
	if len(entry.ResponseHeaders) > 0 {
		fields["response_headers"] = entry.ResponseHeaders
	}
	if entry.RequestBody != "" {
		fields["request_body"] = entry.RequestBody
	}
	if entry.ResponseBody != "" {
		fields["response_body"] = entry.ResponseBody
	}

	return fields
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig(logger *zap.Logger) LoggingConfig {
	return LoggingConfig{
		Logger: logger,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		SkipHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Csrf-Token",
		},
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Csrf-Token",
			"X-Api-Key",
		},
		LogBody:     false,
		MaxBodySize: 4096, // 4KB
		Formatter:   &DefaultLogFormatter{},
	}
}

// ProductionLoggingConfig returns a production-ready logging configuration
func ProductionLoggingConfig(logger *zap.Logger) LoggingConfig {
	return LoggingConfig{
		Logger: logger,
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
			"/robots.txt",
		},
		SkipHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Csrf-Token",
			"X-Api-Key",
			"X-Forwarded-For",
			"X-Real-Ip",
		},
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Csrf-Token",
			"X-Api-Key",
			"X-Session-Token",
		},
		LogBody:     false,
		MaxBodySize: 1024, // 1KB
		Formatter:   &DefaultLogFormatter{},
	}
}

// DevelopmentLoggingConfig returns a development logging configuration with more verbose output
func DevelopmentLoggingConfig(logger *zap.Logger) LoggingConfig {
	return LoggingConfig{
		Logger: logger,
		SkipPaths: []string{
			"/favicon.ico",
		},
		SkipHeaders:      []string{},
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Csrf-Token",
		},
		LogBody:     true,
		MaxBodySize: 8192, // 8KB
		Formatter:   &DefaultLogFormatter{},
	}
}

// LoggingWithConfig creates a logging middleware with custom configuration
func LoggingWithConfig(config LoggingConfig) gin.HandlerFunc {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}
	if config.Formatter == nil {
		config.Formatter = &DefaultLogFormatter{}
	}
	if config.MaxBodySize <= 0 {
		config.MaxBodySize = 4096
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		path := c.Request.URL.Path
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Create log entry
		entry := &LogEntry{
			RequestID: requestID,
			Timestamp: time.Now(),
			Method:    c.Request.Method,
			Path:      path,
			Query:     c.Request.URL.RawQuery,
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}

		// Extract trace ID if available
		if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
			entry.TraceID = traceID
		}

		// Get user information if authenticated
		if user, exists := GetUserFromContext(c); exists {
			entry.UserID = &user.UserID
			entry.Username = user.Username
		}

		// Log request headers
		if !config.LogBody || len(config.SkipHeaders) > 0 {
			entry.RequestHeaders = filterHeaders(c.Request.Header, config.SkipHeaders, config.SensitiveHeaders)
		}

		// Log request body
		var requestBody []byte
		if config.LogBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			if len(requestBody) > int(config.MaxBodySize) {
				requestBody = requestBody[:config.MaxBodySize]
			}
			entry.RequestSize = int64(len(requestBody))
			entry.RequestBody = string(requestBody)
			
			// Restore request body for further processing
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		} else if c.Request.Body != nil {
			// Still need to calculate request size
			if c.Request.ContentLength > 0 {
				entry.RequestSize = c.Request.ContentLength
			}
		}

		// Wrap response writer to capture response data
		var responseWriter *ResponseWriter
		if config.LogBody {
			responseWriter = &ResponseWriter{
				ResponseWriter: c.Writer,
				body:          &bytes.Buffer{},
			}
			c.Writer = responseWriter
		}

		// Process request
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		// Complete log entry
		entry.StatusCode = c.Writer.Status()
		entry.Latency = latency

		// Log response data
		if responseWriter != nil {
			entry.ResponseSize = responseWriter.size
			if responseWriter.body.Len() > 0 {
				responseBody := responseWriter.body.Bytes()
				if len(responseBody) > int(config.MaxBodySize) {
					responseBody = responseBody[:config.MaxBodySize]
				}
				entry.ResponseBody = string(responseBody)
			}
		} else {
			entry.ResponseSize = int64(c.Writer.Size())
		}

		// Log response headers
		if !config.LogBody || len(config.SkipHeaders) > 0 {
			responseHeaders := make(http.Header)
			for k, v := range c.Writer.Header() {
				responseHeaders[k] = v
			}
			entry.ResponseHeaders = filterHeaders(responseHeaders, config.SkipHeaders, config.SensitiveHeaders)
		}

		// Get error information if any
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Format and log
		logFields := config.Formatter.Format(entry)
		
		// Determine log level based on status code
		switch {
		case entry.StatusCode >= 500:
			config.Logger.Error("HTTP Request", zap.Any("request", logFields))
		case entry.StatusCode >= 400:
			config.Logger.Warn("HTTP Request", zap.Any("request", logFields))
		case entry.Latency > 5*time.Second:
			config.Logger.Warn("Slow HTTP Request", zap.Any("request", logFields))
		default:
			config.Logger.Info("HTTP Request", zap.Any("request", logFields))
		}
	}
}

// Logging creates a basic logging middleware
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return LoggingWithConfig(DefaultLoggingConfig(logger))
}

// filterHeaders filters out sensitive or unwanted headers
func filterHeaders(headers http.Header, skipHeaders, sensitiveHeaders []string) map[string]string {
	filtered := make(map[string]string)
	
	for name, values := range headers {
		// Skip headers that shouldn't be logged
		skip := false
		for _, skipHeader := range skipHeaders {
			if strings.EqualFold(name, skipHeader) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		
		// Redact sensitive headers
		sensitive := false
		for _, sensitiveHeader := range sensitiveHeaders {
			if strings.EqualFold(name, sensitiveHeader) {
				sensitive = true
				break
			}
		}
		
		if sensitive {
			filtered[name] = "[REDACTED]"
		} else {
			filtered[name] = strings.Join(values, ", ")
		}
	}
	
	return filtered
}

// MetricsLogging creates a lightweight logging middleware for metrics collection
func MetricsLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		c.Next()
		
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		
		// This could be extended to send metrics to a metrics collector
		// For now, we'll just log basic metrics
		if statusCode >= 400 {
			fmt.Printf("[METRICS] %s %s %d %v\n", method, path, statusCode, latency)
		}
	}
}

// AccessLogFormatter provides Apache-style access log formatting
type AccessLogFormatter struct{}

func (f *AccessLogFormatter) Format(entry *LogEntry) map[string]interface{} {
	// Apache Common Log Format: IP - - [timestamp] "method path protocol" status size
	logLine := fmt.Sprintf(`%s - - [%s] "%s %s HTTP/1.1" %d %d`,
		entry.ClientIP,
		entry.Timestamp.Format("02/Jan/2006:15:04:05 -0700"),
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.ResponseSize,
	)
	
	return map[string]interface{}{
		"access_log": logLine,
		"latency_ms": entry.Latency.Milliseconds(),
	}
}

// StructuredLogging creates a middleware with structured JSON logging
func StructuredLogging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("ip", clientIP),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
		}
		
		if raw != "" {
			fields = append(fields, zap.String("query", raw))
		}
		
		if user, exists := GetUserFromContext(c); exists {
			fields = append(fields, 
				zap.Uint("user_id", user.UserID),
				zap.String("username", user.Username),
			)
		}
		
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		switch {
		case statusCode >= 500:
			logger.Error("Request completed", fields...)
		case statusCode >= 400:
			logger.Warn("Request completed", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}
