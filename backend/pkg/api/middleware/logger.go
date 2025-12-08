package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"polardbx-ui-backend/pkg/logger"
)

// RequestLogConfig 请求日志配置
type RequestLogConfig struct {
	// LogRequestBody 是否记录请求体
	LogRequestBody bool
	// LogResponseBody 是否记录响应体
	LogResponseBody bool
	// MaxBodyLogSize 最大记录的 body 大小
	MaxBodyLogSize int
	// SkipPaths 跳过日志的路径
	SkipPaths []string
	// SensitiveFields 需要脱敏的字段
	SensitiveFields []string
}

// DefaultLogConfig 默认日志配置
var DefaultLogConfig = RequestLogConfig{
	LogRequestBody:  true,
	LogResponseBody: false, // 响应体通常较大，默认不记录
	MaxBodyLogSize:  4096,
	SkipPaths:       []string{"/ping", "/health", "/metrics"},
	SensitiveFields: []string{"password", "token", "secret", "kubeconfig", "authorization"},
}

// RequestIDKey 请求 ID 上下文键
const RequestIDKey = "requestId"

// bodyLogWriter 用于捕获响应体
// (仅在 LogResponseBody 启用时使用)
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger 请求日志中间件（zap 结构化）
func RequestLogger(config ...RequestLogConfig) gin.HandlerFunc {
	cfg := DefaultLogConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for _, skip := range cfg.SkipPaths {
			if strings.HasPrefix(path, skip) {
				c.Next()
				return
			}
		}

		requestID := c.GetString(RequestIDKey)
		if requestID == "" {
			requestID = uuid.New().String()[:8]
			c.Set(RequestIDKey, requestID)
		}

		startTime := time.Now()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := truncateString(c.Request.UserAgent(), 80)

		l := logger.L().With(
			zap.String("requestId", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("clientIP", clientIP),
			zap.String("userAgent", userAgent),
		)

		// 读取请求体（如果需要）
		if cfg.LogRequestBody && c.Request.Body != nil && c.Request.ContentLength > 0 {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			body := string(bodyBytes)
			if len(bodyBytes) > cfg.MaxBodyLogSize {
				body = body[:cfg.MaxBodyLogSize] + "...(truncated)"
			}
			body = maskSensitiveFields(body, cfg.SensitiveFields)
			l.Info("request received", zap.String("requestBody", truncateString(body, cfg.MaxBodyLogSize)))
		} else {
			l.Info("request received")
		}

		// 捕获响应体（如果需要）
		var blw *bodyLogWriter
		if cfg.LogResponseBody {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
		}

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		user := c.GetString("k8sUser")
		k8sContext := c.GetString("k8sContext")

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		}
		if user != "" {
			fields = append(fields, zap.String("user", user))
		}
		if k8sContext != "" {
			fields = append(fields, zap.String("k8sContext", k8sContext))
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		if cfg.LogResponseBody && blw != nil && statusCode >= 400 {
			responseBody := blw.body.String()
			if len(responseBody) > cfg.MaxBodyLogSize {
				responseBody = responseBody[:cfg.MaxBodyLogSize] + "...(truncated)"
			}
			fields = append(fields, zap.String("responseBody", responseBody))
		}

		switch {
		case statusCode >= 500:
			l.With(fields...).Error("request completed")
		case statusCode >= 400:
			l.With(fields...).Warn("request completed")
		default:
			l.With(fields...).Info("request completed")
		}
	}
}

// K8sOperationLogger 记录 K8s 操作的辅助函数
func K8sOperationLogger(c *gin.Context, operation, resource, namespace, name string) func(err error) {
	requestID := c.GetString(RequestIDKey)
	startTime := time.Now()
	l := logger.L().With(
		zap.String("requestId", requestID),
		zap.String("operation", operation),
		zap.String("resource", resource),
		zap.String("namespace", namespace),
		zap.String("name", name),
	)

	l.Info("k8s operation starting")

	return func(err error) {
		latency := time.Since(startTime)
		if err != nil {
			l.Error("k8s operation failed", zap.Duration("latency", latency), zap.Error(err))
		} else {
			l.Info("k8s operation completed", zap.Duration("latency", latency))
		}
	}
}

// BusinessLogger 业务日志记录器
// 维持与旧接口兼容，同时输出结构化字段。
type BusinessLogger struct {
	logger *zap.SugaredLogger
}

// NewBusinessLogger 从上下文创建业务日志记录器
func NewBusinessLogger(c *gin.Context, component string) *BusinessLogger {
	requestID := c.GetString(RequestIDKey)
	if requestID == "" {
		requestID = "no-req-id"
	}
	l := logger.L().With(
		zap.String("requestId", requestID),
		zap.String("component", component),
	).Sugar()
	return &BusinessLogger{logger: l}
}

// Info 记录信息日志
func (l *BusinessLogger) Info(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warn 记录警告日志
func (l *BusinessLogger) Warn(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Error 记录错误日志
func (l *BusinessLogger) Error(err error, format string, args ...interface{}) {
	if err != nil {
		args = append(args, err)
		l.logger.With("error", err).Errorf(format, args...)
	} else {
		l.logger.Errorf(format, args...)
	}
}

// Debug 记录调试日志
func (l *BusinessLogger) Debug(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// WithField 添加额外字段
func (l *BusinessLogger) WithField(key string, value interface{}) *BusinessLogger {
	return &BusinessLogger{logger: l.logger.With(key, value)}
}

// StructuredLog 结构化日志
// (保留给需要 JSON 形式的场景)
type StructuredLog struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`
	RequestID   string                 `json:"request_id,omitempty"`
	Component   string                 `json:"component,omitempty"`
	Message     string                 `json:"message"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Latency     string                 `json:"latency,omitempty"`
	ClientIP    string                 `json:"client_ip,omitempty"`
	User        string                 `json:"user,omitempty"`
	Context     string                 `json:"context,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ExtraFields map[string]interface{} `json:"extra,omitempty"`
}

// LogJSON 输出 JSON 格式日志
func LogJSON(l StructuredLog) {
	l.Timestamp = time.Now().UTC().Format(time.RFC3339)
	if data, err := json.Marshal(l); err == nil {
		logger.L().Info(string(data))
	}
}

// maskSensitiveFields 脱敏敏感字段
func maskSensitiveFields(body string, fields []string) string {
	result := body
	for _, field := range fields {
		// JSON 字段脱敏
		patterns := []string{
			fmt.Sprintf(`"%s"\s*:\s*"[^"]*"`, field),
			fmt.Sprintf(`"%s"\s*:\s*'[^']*'`, field),
		}
		for _, pattern := range patterns {
			result = strings.ReplaceAll(result, pattern, fmt.Sprintf(`"%s":"***MASKED***"`, field))
		}
	}
	return result
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
