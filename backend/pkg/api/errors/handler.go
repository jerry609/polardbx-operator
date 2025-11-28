package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

// HandleError handles an error and sends appropriate response
func HandleError(c *gin.Context, err error) {
	if apiErr, ok := err.(*APIError); ok {
		c.JSON(apiErr.HTTPStatusCode(), apiErr.ToResponse())
		return
	}

	// Try to convert K8s errors
	if apiErr := FromK8sError(err); apiErr != nil && apiErr.Code != ErrK8sConnection {
		c.JSON(apiErr.HTTPStatusCode(), apiErr.ToResponse())
		return
	}

	// Fallback to generic internal error
	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error": map[string]interface{}{
			"code":    ErrInternal,
			"message": "Internal server error",
		},
	})
}

// Abort aborts the request with an API error
func Abort(c *gin.Context, apiErr *APIError) {
	c.AbortWithStatusJSON(apiErr.HTTPStatusCode(), apiErr.ToResponse())
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

// AbortInternal aborts with an internal server error
func AbortInternal(c *gin.Context, message string) {
	Abort(c, Internal(message))
}

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
