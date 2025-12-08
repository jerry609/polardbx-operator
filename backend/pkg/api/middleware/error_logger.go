package middleware

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// ErrorContext 错误上下文
type ErrorContext struct {
	RequestID  string
	Component  string
	Operation  string
	Resource   string
	Namespace  string
	Name       string
	User       string
	StackTrace string
}

// LogError 记录带上下文的错误
func LogError(c *gin.Context, component, operation string, err error, extra ...string) {
	requestID := c.GetString(RequestIDKey)
	user := c.GetString("k8sUser")

	// 获取调用栈
	stackTrace := getStackTrace(3)

	// 构建错误上下文
	ctx := ErrorContext{
		RequestID:  requestID,
		Component:  component,
		Operation:  operation,
		User:       user,
		StackTrace: stackTrace,
	}

	// 解析 K8s 错误
	errType := "UNKNOWN"
	errDetail := err.Error()
	if k8serrors.IsNotFound(err) {
		errType = "NOT_FOUND"
	} else if k8serrors.IsAlreadyExists(err) {
		errType = "ALREADY_EXISTS"
	} else if k8serrors.IsConflict(err) {
		errType = "CONFLICT"
	} else if k8serrors.IsUnauthorized(err) {
		errType = "UNAUTHORIZED"
	} else if k8serrors.IsForbidden(err) {
		errType = "FORBIDDEN"
	} else if k8serrors.IsTimeout(err) {
		errType = "TIMEOUT"
	} else if k8serrors.IsServerTimeout(err) {
		errType = "SERVER_TIMEOUT"
	} else if k8serrors.IsServiceUnavailable(err) {
		errType = "SERVICE_UNAVAILABLE"
	} else if k8serrors.IsBadRequest(err) {
		errType = "BAD_REQUEST"
	} else if k8serrors.IsInvalid(err) {
		errType = "INVALID"
	}

	// 构建额外信息
	extraInfo := ""
	if len(extra) > 0 {
		extraInfo = " | " + strings.Join(extra, " | ")
	}

	log.Printf("[%s] ❌ ERROR [%s] %s.%s failed | type=%s | user=%s%s\n  Error: %s\n  Stack: %s",
		ctx.RequestID,
		errType,
		ctx.Component,
		ctx.Operation,
		errType,
		ctx.User,
		extraInfo,
		errDetail,
		ctx.StackTrace,
	)
}

// LogK8sError 记录 K8s API 错误
func LogK8sError(c *gin.Context, operation, resourceType, namespace, name string, err error) {
	requestID := c.GetString(RequestIDKey)
	user := c.GetString("k8sUser")

	// 解析 K8s 错误类型和状态码
	errType := "UNKNOWN"
	statusCode := 500
	reason := ""

	if statusErr, ok := err.(*k8serrors.StatusError); ok {
		statusCode = int(statusErr.ErrStatus.Code)
		reason = string(statusErr.ErrStatus.Reason)
		errType = reason
	} else if k8serrors.IsNotFound(err) {
		errType = "NOT_FOUND"
		statusCode = 404
	} else if k8serrors.IsAlreadyExists(err) {
		errType = "ALREADY_EXISTS"
		statusCode = 409
	} else if k8serrors.IsForbidden(err) {
		errType = "FORBIDDEN"
		statusCode = 403
	} else if k8serrors.IsUnauthorized(err) {
		errType = "UNAUTHORIZED"
		statusCode = 401
	}

	log.Printf("[%s] ❌ K8s ERROR | op=%s | resource=%s/%s/%s | status=%d | type=%s | user=%s | error=%v",
		requestID,
		operation,
		resourceType,
		namespace,
		name,
		statusCode,
		errType,
		user,
		err,
	)
}

// LogValidationError 记录验证错误
func LogValidationError(c *gin.Context, component string, field, message string) {
	requestID := c.GetString(RequestIDKey)
	log.Printf("[%s] ⚠️ VALIDATION [%s] field=%s | %s",
		requestID, component, field, message)
}

// LogSecurityEvent 记录安全事件
func LogSecurityEvent(c *gin.Context, eventType, message string) {
	requestID := c.GetString(RequestIDKey)
	clientIP := c.ClientIP()
	user := c.GetString("k8sUser")
	context := c.GetString("k8sContext")

	log.Printf("[%s] 🔒 SECURITY [%s] | ip=%s | user=%s | context=%s | %s",
		requestID, eventType, clientIP, user, context, message)
}

// LogAudit 记录审计日志
func LogAudit(c *gin.Context, action, resourceType, namespace, name string, success bool) {
	requestID := c.GetString(RequestIDKey)
	user := c.GetString("k8sUser")
	clientIP := c.ClientIP()

	status := "SUCCESS"
	emoji := "✓"
	if !success {
		status = "FAILURE"
		emoji = "✗"
	}

	log.Printf("[%s] %s AUDIT [%s] %s %s/%s/%s | user=%s | ip=%s",
		requestID, emoji, status, action, resourceType, namespace, name, user, clientIP)
}

// getStackTrace 获取调用栈
func getStackTrace(skip int) string {
	var pcs [10]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		// 跳过运行时和标准库
		if strings.Contains(frame.File, "runtime/") ||
			strings.Contains(frame.File, "net/http/") {
			if !more {
				break
			}
			continue
		}
		sb.WriteString(fmt.Sprintf("\n    %s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	return sb.String()
}

// RecoveryWithLogger Panic 恢复中间件，带日志
func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString(RequestIDKey)
				stack := getStackTrace(3)

				log.Printf("[%s] 💥 PANIC RECOVERED | path=%s | error=%v\n  Stack: %s",
					requestID, c.Request.URL.Path, err, stack)

				c.AbortWithStatusJSON(500, gin.H{
					"error":      "Internal Server Error",
					"requestId": requestID,
				})
			}
		}()
		c.Next()
	}
}
