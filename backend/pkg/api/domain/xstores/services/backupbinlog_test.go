package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// ==================== BackupBinlogService Tests ====================

func setupBackupBinlogRouter(t *testing.T, objs ...runtime.Object) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()

	scheme := runtime.NewScheme()
	_ = polardbxv1.AddToScheme(scheme)
	cli := crfake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	r.Use(func(c *gin.Context) {
		c.Set("k8sClient", cli)
		c.Next()
	})

	svc := NewBackupBinlogService()
	r.GET("/backupbinlogs", svc.List)
	r.POST("/backupbinlogs", svc.Create)
	r.GET("/backupbinlogs/:namespace/:name", svc.Get)
	r.PUT("/backupbinlogs/:namespace/:name", svc.Update)
	r.DELETE("/backupbinlogs/:namespace/:name", svc.Delete)

	return r
}

func TestBackupBinlogService_List_Empty(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/backupbinlogs?namespace=default", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupBinlogService_List_WithData(t *testing.T) {
	binlog := &polardbxv1.XStoreBackupBinlog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-binlog",
			Namespace: "default",
		},
	}
	router := setupBackupBinlogRouter(t, binlog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/backupbinlogs?namespace=default", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupBinlogService_List_AllNamespaces(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/backupbinlogs", nil) // No namespace filter
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupBinlogService_Create_Success(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	body, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string]string{
			"name": "new-binlog",
		},
		"spec": map[string]interface{}{},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/backupbinlogs?namespace=default", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBackupBinlogService_Create_InvalidJSON(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/backupbinlogs?namespace=default", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupBinlogService_Create_DefaultNamespace(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	body, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string]string{
			"name": "new-binlog",
		},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/backupbinlogs", bytes.NewBuffer(body)) // Use default namespace
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBackupBinlogService_Get_Success(t *testing.T) {
	binlog := &polardbxv1.XStoreBackupBinlog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-binlog",
			Namespace: "default",
		},
	}
	router := setupBackupBinlogRouter(t, binlog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/backupbinlogs/default/test-binlog", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupBinlogService_Get_NotFound(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/backupbinlogs/default/nonexistent", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestBackupBinlogService_Update_Success(t *testing.T) {
	binlog := &polardbxv1.XStoreBackupBinlog{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-binlog",
			Namespace:       "default",
			ResourceVersion: "1", // Add resource version for update
		},
	}
	router := setupBackupBinlogRouter(t, binlog)

	body, _ := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":            "test-binlog",
			"namespace":       "default",
			"resourceVersion": "1", // Must match existing resource version
		},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/backupbinlogs/default/test-binlog", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupBinlogService_Update_InvalidJSON(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/backupbinlogs/default/test", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBackupBinlogService_Delete_Success(t *testing.T) {
	binlog := &polardbxv1.XStoreBackupBinlog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-binlog",
			Namespace: "default",
		},
	}
	router := setupBackupBinlogRouter(t, binlog)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/backupbinlogs/default/test-binlog", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupBinlogService_Delete_NotFound(t *testing.T) {
	router := setupBackupBinlogRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/backupbinlogs/default/nonexistent", nil)
	router.ServeHTTP(w, req)

	// Delete of nonexistent may return 404 or succeed silently
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestBackupBinlogService_NoK8sClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// No k8sClient middleware
	svc := NewBackupBinlogService()
	r.GET("/backupbinlogs", svc.List)
	r.POST("/backupbinlogs", svc.Create)
	r.GET("/backupbinlogs/:namespace/:name", svc.Get)
	r.PUT("/backupbinlogs/:namespace/:name", svc.Update)
	r.DELETE("/backupbinlogs/:namespace/:name", svc.Delete)

	tests := []struct {
		method string
		path   string
		body   []byte
	}{
		{"GET", "/backupbinlogs", nil},
		{"POST", "/backupbinlogs", []byte(`{"metadata":{"name":"test"}}`)},
		{"GET", "/backupbinlogs/default/test", nil},
		{"PUT", "/backupbinlogs/default/test", []byte(`{"metadata":{"name":"test"}}`)},
		{"DELETE", "/backupbinlogs/default/test", nil},
	}

	for _, tc := range tests {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			var req *http.Request
			if tc.body != nil {
				req, _ = http.NewRequest(tc.method, tc.path, bytes.NewBuffer(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = http.NewRequest(tc.method, tc.path, nil)
			}
			r.ServeHTTP(w, req)
			// Should return 401 (unauthorized) when k8s client not available
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestNewBackupBinlogService(t *testing.T) {
	svc := NewBackupBinlogService()
	require.NotNil(t, svc)
}
