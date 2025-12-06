package errors

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// ============================================================================
// Global Error Handler Middleware
// ============================================================================

// Handler returns a Gin handler that properly formats API errors
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for errors after handler execution
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			HandleError(c, err)
		}
	}
}

// RecoveryHandler returns a recovery middleware that converts panics to APIError
func RecoveryHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace (internal only)
				log.Printf("PANIC recovered: %v\nStack trace:\n%s", r, debug.Stack())

				// Return generic error to client (no internal details)
				apiErr := &APIError{
					Code:       ErrInternal,
					Message:    "An unexpected error occurred",
					HTTPStatus: http.StatusInternalServerError,
				}
				c.AbortWithStatusJSON(apiErr.HTTPStatusCode(), apiErr.ToResponse())
			}
		}()
		c.Next()
	}
}

// ============================================================================
// Error Handling Functions
// ============================================================================

// HandleError handles an error and sends appropriate response
// It performs the following:
// 1. Converts known error types to APIError
// 2. Logs internal details (not exposed to client)
// 3. Returns sanitized response to client
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var apiErr *APIError

	// Try to convert to APIError
	switch e := err.(type) {
	case *APIError:
		apiErr = e
	default:
		// Try context errors first
		if ctxErr := FromContextError(err); ctxErr != nil {
			apiErr = ctxErr
		} else if k8sErr := FromK8sError(err); k8sErr != nil && k8sErr.Code != ErrK8sConnection {
			apiErr = k8sErr
		} else if bindErr := FromBindError(err); bindErr != nil {
			apiErr = bindErr
		} else {
			// Fallback to generic internal error
			apiErr = Internal("An error occurred while processing your request")
			apiErr.Cause = err
		}
	}

	// Log the full error internally (with context)
	logError(c, apiErr)

	// Send sanitized response to client
	sendErrorResponse(c, apiErr)
}

// logError logs the error with request context for debugging
func logError(c *gin.Context, apiErr *APIError) {
	// Build log context
	user := c.GetString("k8sUser")
	if user == "" {
		user = "anonymous"
	}

	requestID := c.GetString("requestId")
	if requestID == "" {
		requestID = c.GetHeader("X-Request-ID")
	}

	// Log with context (internal details)
	logMsg := "[ERROR] code=%s user=%s request_id=%s method=%s path=%s message=%s"
	logArgs := []interface{}{
		apiErr.Code,
		user,
		requestID,
		c.Request.Method,
		c.Request.URL.Path,
		apiErr.Message,
	}

	if apiErr.Cause != nil {
		logMsg += " cause=%v"
		logArgs = append(logArgs, apiErr.Cause)
	}

	log.Printf(logMsg, logArgs...)
}

// sendErrorResponse sends sanitized error response to client
func sendErrorResponse(c *gin.Context, apiErr *APIError) {
	// Build response (sanitized, no internal details)
	resp := apiErr.ToResponse()

	// Add retry headers if applicable
	if details, ok := apiErr.Details.(map[string]interface{}); ok {
		if retryAfter, ok := details["retry_after_seconds"].(int); ok && retryAfter > 0 {
			c.Header("Retry-After", string(rune(retryAfter)))
		}
	}

	c.JSON(apiErr.HTTPStatusCode(), resp)
}

// ============================================================================
// Abort Functions (stop request processing)
// ============================================================================

// Abort aborts the request with an API error
func Abort(c *gin.Context, apiErr *APIError) {
	logError(c, apiErr)
	c.AbortWithStatusJSON(apiErr.HTTPStatusCode(), apiErr.ToResponse())
}

// AbortWithError aborts with any error (converts to APIError first)
func AbortWithError(c *gin.Context, err error) {
	if apiErr, ok := err.(*APIError); ok {
		Abort(c, apiErr)
		return
	}

	// Try to convert
	if ctxErr := FromContextError(err); ctxErr != nil {
		Abort(c, ctxErr)
		return
	}
	if k8sErr := FromK8sError(err); k8sErr != nil {
		Abort(c, k8sErr)
		return
	}
	if bindErr := FromBindError(err); bindErr != nil {
		Abort(c, bindErr)
		return
	}

	// Fallback
	apiErr := Internal("An error occurred")
	apiErr.Cause = err
	Abort(c, apiErr)
}

// AbortNotFound aborts with a not found error
func AbortNotFound(c *gin.Context, resource, name string) {
	Abort(c, NotFound(resource, name))
}

// AbortValidation aborts with a validation error
func AbortValidation(c *gin.Context, message string) {
	Abort(c, Validation(message))
}

// AbortUnauthorized aborts with an unauthorized error
func AbortUnauthorized(c *gin.Context, message string) {
	Abort(c, Unauthorized(message))
}

// AbortForbidden aborts with a forbidden error
func AbortForbidden(c *gin.Context, message string) {
	Abort(c, Forbidden(message))
}

// AbortInternal aborts with an internal server error
func AbortInternal(c *gin.Context, message string) {
	Abort(c, Internal(message))
}

// AbortK8sError aborts with a Kubernetes error (sanitized for client)
func AbortK8sError(c *gin.Context, operation string, err error) {
	apiErr := FromK8sError(err)
	if apiErr == nil {
		apiErr = Internal("Kubernetes operation failed")
	}

	// Override message to be more user-friendly
	switch {
	case k8serrors.IsNotFound(err):
		apiErr.Message = "Resource not found"
	case k8serrors.IsForbidden(err):
		apiErr.Message = "Permission denied for this operation"
	case k8serrors.IsConflict(err):
		apiErr.Message = "Resource conflict - please retry"
	case k8serrors.IsTimeout(err), k8serrors.IsServerTimeout(err):
		apiErr.Message = "Operation timed out - please retry"
	default:
		apiErr.Message = "Failed to " + operation
	}

	apiErr.Cause = err
	Abort(c, apiErr)
}

// ============================================================================
// Response Helpers
// ============================================================================

// Response sends a success response
func Response(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, data interface{}) {
	Response(c, http.StatusOK, data)
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	Response(c, http.StatusCreated, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Accepted sends a 202 Accepted response (for async operations)
func Accepted(c *gin.Context, data interface{}) {
	Response(c, http.StatusAccepted, data)
}

// ============================================================================
// Request ID Middleware
// ============================================================================

// RequestIDMiddleware adds a request ID to the context for tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
