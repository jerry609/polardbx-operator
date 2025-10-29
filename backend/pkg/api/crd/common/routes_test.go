package common

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Mock handlers for testing
var (
	mockList   = func(c *gin.Context) { c.String(http.StatusOK, "list") }
	mockCreate = func(c *gin.Context) { c.String(http.StatusOK, "create") }
	mockGet    = func(c *gin.Context) { c.String(http.StatusOK, "get") }
	mockUpdate = func(c *gin.Context) { c.String(http.StatusOK, "update") }
	mockDelete = func(c *gin.Context) { c.String(http.StatusOK, "delete") }
)

func TestRegisterNamespaceScopedCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	crd := router.Group("/crd")

	handlers := CRUDHandlers{
		List:   mockList,
		Create: mockCreate,
		Get:    mockGet,
		Update: mockUpdate,
		Delete: mockDelete,
	}

	RegisterNamespaceScopedCRUD(crd, "testresource", handlers)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"List", "GET", "/crd/testresource", http.StatusOK, "list"},
		{"Create", "POST", "/crd/testresource", http.StatusOK, "create"},
		{"Get", "GET", "/crd/testresource/ns1/item1", http.StatusOK, "get"},
		{"Update", "PUT", "/crd/testresource/ns1/item1", http.StatusOK, "update"},
		{"Delete", "DELETE", "/crd/testresource/ns1/item1", http.StatusOK, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestRegisterClusterScopedCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	crd := router.Group("/crd")

	handlers := CRUDHandlers{
		List:   mockList,
		Create: mockCreate,
		Get:    mockGet,
		Update: mockUpdate,
		Delete: mockDelete,
	}

	RegisterClusterScopedCRUD(crd, "clusterresource", handlers)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"List", "GET", "/crd/clusterresource", http.StatusOK, "list"},
		{"Create", "POST", "/crd/clusterresource", http.StatusOK, "create"},
		{"Get", "GET", "/crd/clusterresource/item1", http.StatusOK, "get"},
		{"Update", "PUT", "/crd/clusterresource/item1", http.StatusOK, "update"},
		{"Delete", "DELETE", "/crd/clusterresource/item1", http.StatusOK, "delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}

func TestRegisterNamespaceScopedCRUD_ReturnsRouterGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	crd := router.Group("/crd")

	handlers := CRUDHandlers{
		List:   mockList,
		Create: mockCreate,
		Get:    mockGet,
		Update: mockUpdate,
		Delete: mockDelete,
	}

	// The function should return a router group for additional endpoints
	r := RegisterNamespaceScopedCRUD(crd, "testresource", handlers)
	assert.NotNil(t, r, "RegisterNamespaceScopedCRUD should return a router group")

	// Add an additional endpoint to the returned group
	r.GET("/extra", func(c *gin.Context) { c.String(http.StatusOK, "extra") })

	// Test the extra endpoint
	req, _ := http.NewRequest("GET", "/crd/testresource/extra", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "extra", w.Body.String())
}
