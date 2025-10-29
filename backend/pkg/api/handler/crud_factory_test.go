package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// MockResource is a simple mock for testing
type MockResource struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// MockResourceOps implements ResourceOperations for testing
type MockResourceOps struct {
	listFunc   func(ctx context.Context, cli client.Client, namespace string) ([]MockResource, error)
	getFunc    func(ctx context.Context, cli client.Client, namespace, name string) (*MockResource, error)
	createFunc func(ctx context.Context, cli client.Client, namespace string, obj *MockResource) (*MockResource, error)
	updateFunc func(ctx context.Context, cli client.Client, namespace string, obj *MockResource) (*MockResource, error)
	deleteFunc func(ctx context.Context, cli client.Client, namespace, name string) error
}

func (m MockResourceOps) List(ctx context.Context, cli client.Client, namespace string) ([]MockResource, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, cli, namespace)
	}
	return []MockResource{}, nil
}

func (m MockResourceOps) Get(ctx context.Context, cli client.Client, namespace, name string) (*MockResource, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, cli, namespace, name)
	}
	return &MockResource{}, nil
}

func (m MockResourceOps) Create(ctx context.Context, cli client.Client, namespace string, obj *MockResource) (*MockResource, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, cli, namespace, obj)
	}
	return obj, nil
}

func (m MockResourceOps) Update(ctx context.Context, cli client.Client, namespace string, obj *MockResource) (*MockResource, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, cli, namespace, obj)
	}
	return obj, nil
}

func (m MockResourceOps) Delete(ctx context.Context, cli client.Client, namespace, name string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, cli, namespace, name)
	}
	return nil
}

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Set up a fake k8s client
	scheme := runtime.NewScheme()
	cli := fake.NewClientBuilder().WithScheme(scheme).Build()
	c.Set("k8sClient", cli)
	c.Set("k8sDefaultNamespace", "default")
	
	return c, w
}

func TestCRUDHandlerFactory_List(t *testing.T) {
	ops := MockResourceOps{
		listFunc: func(ctx context.Context, cli client.Client, namespace string) ([]MockResource, error) {
			return []MockResource{
				{Name: "resource1", Value: "value1"},
				{Name: "resource2", Value: "value2"},
			}, nil
		},
	}
	
	factory := NewCRUDHandlerFactory(ops, "mock resource", func() *MockResource { return &MockResource{} })
	handler := factory.List()
	
	c, w := setupTestContext()
	c.Request = httptest.NewRequest("GET", "/resources", nil)
	
	handler(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "resource1")
	assert.Contains(t, w.Body.String(), "resource2")
}

func TestCRUDHandlerFactory_Get(t *testing.T) {
	ops := MockResourceOps{
		getFunc: func(ctx context.Context, cli client.Client, namespace, name string) (*MockResource, error) {
			return &MockResource{
				Name:  name,
				Value: "test-value",
			}, nil
		},
	}
	
	factory := NewCRUDHandlerFactory(ops, "mock resource", func() *MockResource { return &MockResource{} })
	handler := factory.Get()
	
	c, w := setupTestContext()
	c.Request = httptest.NewRequest("GET", "/resources/ns1/res1", nil)
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns1"},
		{Key: "name", Value: "res1"},
	}
	
	handler(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-value")
	assert.Contains(t, w.Body.String(), "res1")
}

func TestCRUDHandlerFactory_Create(t *testing.T) {
	ops := MockResourceOps{
		createFunc: func(ctx context.Context, cli client.Client, namespace string, obj *MockResource) (*MockResource, error) {
			return obj, nil
		},
	}
	
	factory := NewCRUDHandlerFactory(ops, "mock resource", func() *MockResource { return &MockResource{} })
	handler := factory.Create()
	
	c, w := setupTestContext()
	body := `{"name":"new-resource","value":"created"}`
	c.Request = httptest.NewRequest("POST", "/resources", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	
	handler(c)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "new-resource")
	assert.Contains(t, w.Body.String(), "created")
}

func TestCRUDHandlerFactory_Update(t *testing.T) {
	ops := MockResourceOps{
		updateFunc: func(ctx context.Context, cli client.Client, namespace string, obj *MockResource) (*MockResource, error) {
			return obj, nil
		},
	}
	
	factory := NewCRUDHandlerFactory(ops, "mock resource", func() *MockResource { return &MockResource{} })
	handler := factory.Update()
	
	c, w := setupTestContext()
	body := `{"name":"update-resource","value":"updated"}`
	c.Request = httptest.NewRequest("PUT", "/resources/ns1/res1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns1"},
	}
	
	handler(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "updated")
}

func TestCRUDHandlerFactory_Delete(t *testing.T) {
	ops := MockResourceOps{
		deleteFunc: func(ctx context.Context, cli client.Client, namespace, name string) error {
			return nil
		},
	}
	
	factory := NewCRUDHandlerFactory(ops, "mock resource", func() *MockResource { return &MockResource{} })
	handler := factory.Delete()
	
	c, w := setupTestContext()
	c.Request = httptest.NewRequest("DELETE", "/resources/ns1/res1", nil)
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns1"},
		{Key: "name", Value: "res1"},
	}
	
	handler(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "deleted")
}

func TestCRUDHandlerFactory_ErrorHandling(t *testing.T) {
	ops := MockResourceOps{
		getFunc: func(ctx context.Context, cli client.Client, namespace, name string) (*MockResource, error) {
			return nil, errors.New("resource not found")
		},
	}
	
	factory := NewCRUDHandlerFactory(ops, "mock resource", func() *MockResource { return &MockResource{} })
	handler := factory.Get()
	
	c, w := setupTestContext()
	c.Request = httptest.NewRequest("GET", "/resources/ns1/notfound", nil)
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns1"},
		{Key: "name", Value: "notfound"},
	}
	
	handler(c)
	
	// Should return an error response
	assert.NotEqual(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

func TestClusterScopedCRUDHandlerFactory(t *testing.T) {
	ops := MockResourceOps{
		getFunc: func(ctx context.Context, cli client.Client, namespace, name string) (*MockResource, error) {
			return &MockResource{Name: name, Value: "cluster-resource"}, nil
		},
	}
	
	factory := NewClusterScopedCRUDHandlerFactory(ops, "cluster resource", func() *MockResource { return &MockResource{} })
	handler := factory.Get()
	
	c, w := setupTestContext()
	c.Request = httptest.NewRequest("GET", "/resources/res1", nil)
	c.Params = gin.Params{
		{Key: "name", Value: "res1"},
	}
	
	handler(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "res1")
	assert.Contains(t, w.Body.String(), "cluster-resource")
}
