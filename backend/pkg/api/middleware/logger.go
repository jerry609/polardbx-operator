package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
const RequestIDKey = "request_id"

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestLogger 请求日志中间件
func RequestLogger(config ...RequestLogConfig) gin.HandlerFunc {
	cfg := DefaultLogConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 跳过特定路径
		path := c.Request.URL.Path
		for _, skip := range cfg.SkipPaths {
			if strings.HasPrefix(path, skip) {
				c.Next()
				return
			}
		}

		// 生成请求 ID
		requestID := uuid.New().String()[:8]
		c.Set(RequestIDKey, requestID)

		// 记录开始时间
		startTime := time.Now()

		// 获取请求信息
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 读取请求体（如果需要）
		var requestBody string
		if cfg.LogRequestBody && c.Request.Body != nil && c.Request.ContentLength > 0 {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > cfg.MaxBodyLogSize {
				requestBody = string(bodyBytes[:cfg.MaxBodyLogSize]) + "...(truncated)"
			} else {
				requestBody = maskSensitiveFields(string(bodyBytes), cfg.SensitiveFields)
			}
		}

		// 捕获响应体（如果需要）
		var blw *bodyLogWriter
		if cfg.LogResponseBody {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
		}

		// 记录请求开始
		log.Printf("[%s] ▶ %s %s from %s | User-Agent: %s",
			requestID, method, path, clientIP, truncateString(userAgent, 50))

		if requestBody != "" && method != "GET" {
			log.Printf("[%s] 📤 Request Body: %s", requestID, truncateString(requestBody, 500))
		}

		// 执行请求
		c.Next()

		// 计算耗时
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 获取用户身份（如果有）
		user := c.GetString("k8sUser")
		context := c.GetString("k8sContext")

		// 构建日志消息
		logLevel := "INFO"
		emoji := "✓"
		if statusCode >= 400 && statusCode < 500 {
			logLevel = "WARN"
			emoji = "⚠"
		} else if statusCode >= 500 {
			logLevel = "ERROR"
			emoji = "✗"
		}

		// 记录请求完成
		identityInfo := ""
		if user != "" {
			identityInfo = fmt.Sprintf(" | user=%s context=%s", user, context)
		}

		log.Printf("[%s] %s %s %s %s | %d | %v%s",
			requestID, emoji, logLevel, method, path,
			statusCode, latency, identityInfo)

		// 记录响应体（如果是错误响应）
		if cfg.LogResponseBody && blw != nil && statusCode >= 400 {
			responseBody := blw.body.String()
			if len(responseBody) > cfg.MaxBodyLogSize {
				responseBody = responseBody[:cfg.MaxBodyLogSize] + "...(truncated)"
			}
			log.Printf("[%s] 📥 Response Body: %s", requestID, responseBody)
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[%s] ❌ Error: %s", requestID, e.Error())
			}
		}
	}
}

// K8sOperationLogger 记录 K8s 操作的辅助函数
func K8sOperationLogger(c *gin.Context, operation, resource, namespace, name string) func(err error) {
	requestID := c.GetString(RequestIDKey)
	startTime := time.Now()

	log.Printf("[%s] 🔄 K8s %s %s/%s/%s starting",
		requestID, operation, resource, namespace, name)

	return func(err error) {
		latency := time.Since(startTime)
		if err != nil {
			log.Printf("[%s] ❌ K8s %s %s/%s/%s failed after %v: %v",
				requestID, operation, resource, namespace, name, latency, err)
		} else {
			log.Printf("[%s] ✓ K8s %s %s/%s/%s completed in %v",
				requestID, operation, resource, namespace, name, latency)
		}
	}
}

// BusinessLogger 业务日志记录器
type BusinessLogger struct {
	RequestID string
	Component string
}

// NewBusinessLogger 从上下文创建业务日志记录器
func NewBusinessLogger(c *gin.Context, component string) *BusinessLogger {
	requestID := c.GetString(RequestIDKey)
	if requestID == "" {
		requestID = "no-req-id"
	}
	return &BusinessLogger{
		RequestID: requestID,
		Component: component,
	}
}

// Info 记录信息日志
func (l *BusinessLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] ℹ️ %s", l.RequestID, l.Component, msg)
}

// Warn 记录警告日志
func (l *BusinessLogger) Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] ⚠️ %s", l.RequestID, l.Component, msg)
}

// Error 记录错误日志
func (l *BusinessLogger) Error(err error, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] ❌ %s: %v", l.RequestID, l.Component, msg, err)
}

// Debug 记录调试日志
func (l *BusinessLogger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] 🔍 %s", l.RequestID, l.Component, msg)
}

// WithField 添加额外字段
func (l *BusinessLogger) WithField(key string, value interface{}) *BusinessLogger {
	return &BusinessLogger{
		RequestID: l.RequestID,
		Component: fmt.Sprintf("%s.%s=%v", l.Component, key, value),
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

// StructuredLog 结构化日志
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
		log.Println(string(data))
	}
}
