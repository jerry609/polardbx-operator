package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseBuilder provides a fluent interface for building HTTP responses.
// This implements the Builder Pattern for consistent API responses.
type ResponseBuilder struct {
	statusCode int
	data       interface{}
	message    string
	errorMsg   string
	details    string
}

// NewResponse creates a new response builder.
func NewResponse() *ResponseBuilder {
	return &ResponseBuilder{
		statusCode: http.StatusOK,
	}
}

// Status sets the HTTP status code.
func (r *ResponseBuilder) Status(code int) *ResponseBuilder {
	r.statusCode = code
	return r
}

// Data sets the response data payload.
func (r *ResponseBuilder) Data(data interface{}) *ResponseBuilder {
	r.data = data
	return r
}

// Message sets a success message.
func (r *ResponseBuilder) Message(msg string) *ResponseBuilder {
	r.message = msg
	return r
}

// Error sets an error message.
func (r *ResponseBuilder) Error(errMsg string) *ResponseBuilder {
	r.errorMsg = errMsg
	return r
}

// Details adds additional error details.
func (r *ResponseBuilder) Details(details string) *ResponseBuilder {
	r.details = details
	return r
}

// Send sends the response using the gin context.
func (r *ResponseBuilder) Send(c *gin.Context) {
	response := gin.H{}
	
	if r.data != nil {
		// For successful responses with data, send data directly
		if r.errorMsg == "" && r.message == "" {
			c.JSON(r.statusCode, r.data)
			return
		}
		response["data"] = r.data
	}
	
	if r.message != "" {
		response["message"] = r.message
	}
	
	if r.errorMsg != "" {
		response["error"] = r.errorMsg
	}
	
	if r.details != "" {
		response["details"] = r.details
	}
	
	c.JSON(r.statusCode, response)
}

// Convenience helper functions for common response patterns

// Success sends a successful response with data.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// SuccessMessage sends a successful response with a message.
func SuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// Created sends a 201 Created response with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// BadRequest sends a 400 Bad Request error response.
func BadRequest(c *gin.Context, message string, details ...string) {
	response := gin.H{"error": message}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(http.StatusBadRequest, response)
}

// NotFound sends a 404 Not Found error response.
func NotFound(c *gin.Context, message string, details ...string) {
	response := gin.H{"error": message}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(http.StatusNotFound, response)
}

// Unauthorized sends a 401 Unauthorized error response.
func Unauthorized(c *gin.Context, message string, details ...string) {
	response := gin.H{"error": message}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(http.StatusUnauthorized, response)
}

// Forbidden sends a 403 Forbidden error response.
func Forbidden(c *gin.Context, message string, details ...string) {
	response := gin.H{"error": message}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(http.StatusForbidden, response)
}

// Conflict sends a 409 Conflict error response.
func Conflict(c *gin.Context, message string, details ...string) {
	response := gin.H{"error": message}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(http.StatusConflict, response)
}

// InternalError sends a 500 Internal Server Error response.
func InternalError(c *gin.Context, message string, details ...string) {
	response := gin.H{"error": message}
	if len(details) > 0 && details[0] != "" {
		response["details"] = details[0]
	}
	c.JSON(http.StatusInternalServerError, response)
}
