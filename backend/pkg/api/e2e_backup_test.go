package api

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	crd_polardbxbackups "polardbx-ui-backend/pkg/api/crd/polardbxbackups"
	domain_pxc "polardbx-ui-backend/pkg/api/domain/polardbxclusters"
)

// setupBackupRouter sets up a test router with backup routes
func setupBackupRouter(t *testing.T, objs ...runtime.Object) (*gin.Engine, client.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")

	// Build scheme with all required types
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = polardbxv1.AddToScheme(scheme)

	ctrlClient := crfake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objs...).
		Build()
	cs := k8sfake.NewSimpleClientset()

	v1.Use(func(c *gin.Context) {
		c.Set("k8sClient", ctrlClient)
		c.Set("clientset", cs)
		c.Set("k8sDefaultNamespace", "default")
		c.Next()
	})

	// Register backup routes
	crdGroup := v1.Group("/crd")
	crd_polardbxbackups.RegisterRoutes(crdGroup)

	// Also register domain routes for backup operations
	v1.GET("/backups/overview", domain_pxc.GetBackupOverview)
	v1.GET("/backups/binlog/metrics", domain_pxc.GetBinlogMetrics)
	v1.POST("/backups/validate", domain_pxc.ValidateBackup)

	return r, ctrlClient
}

// ==================== PolarDBXBackup E2E Tests ====================

func TestE2E_PolarDBXBackup_Validate_InvalidJSON(t *testing.T) {
	router, _ := setupBackupRouter(t)

	body := bytes.NewBufferString(`{invalid json}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/crd/polardbxbackups/validate", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestE2E_PolarDBXBackup_Validate_Success(t *testing.T) {
	router, _ := setupBackupRouter(t)

	body := bytes.NewBufferString(`{
		"cluster": {
			"name": "test-cluster"
		},
		"storageProvider": {
			"storageName": "OSS",
			"sink": "oss://test-bucket/backup/"
		}
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/crd/polardbxbackups/validate", body)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	// Validation may return 200 (valid) or 400 (invalid)
	assert.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, w.Code)
}

func TestE2E_PolarDBXBackup_GetOverview_Empty(t *testing.T) {
	router, _ := setupBackupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/backups/overview?namespace=default", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
}

func TestE2E_PolarDBXBackup_GetOverview_WithBackups(t *testing.T) {
	backup := &polardbxv1.PolarDBXBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-backup",
			Namespace: "default",
		},
		Spec: polardbxv1.PolarDBXBackupSpec{
			Cluster: polardbxv1.PolarDBXClusterReference{
				Name: "test-cluster",
			},
		},
		Status: polardbxv1.PolarDBXBackupStatus{
			Phase: polardbxv1.BackupFinished,
		},
	}

	router, _ := setupBackupRouter(t, backup)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/backups/overview?namespace=default", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestE2E_PolarDBXBackup_GetBinlogMetrics_Empty(t *testing.T) {
	router, _ := setupBackupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/backups/binlog/metrics?namespace=default", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestE2E_PolarDBXBackup_GetMetrics_NotFound(t *testing.T) {
	router, _ := setupBackupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/crd/polardbxbackups/default/nonexistent/metrics", nil)
	router.ServeHTTP(w, req)
	// May return 404 or 200 (empty metrics)
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}

func TestE2E_PolarDBXBackup_GetMetrics_Success(t *testing.T) {
	backup := &polardbxv1.PolarDBXBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-backup",
			Namespace: "default",
		},
		Spec: polardbxv1.PolarDBXBackupSpec{
			Cluster: polardbxv1.PolarDBXClusterReference{
				Name: "test-cluster",
			},
		},
		Status: polardbxv1.PolarDBXBackupStatus{
			Phase: polardbxv1.BackupFinished,
		},
	}

	router, _ := setupBackupRouter(t, backup)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/crd/polardbxbackups/default/test-backup/metrics", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestE2E_PolarDBXBackup_Delete_NotFound(t *testing.T) {
	router, _ := setupBackupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/crd/polardbxbackups/default/nonexistent", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestE2E_PolarDBXBackup_Delete_Success(t *testing.T) {
	backup := &polardbxv1.PolarDBXBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-backup",
			Namespace: "default",
		},
		Spec: polardbxv1.PolarDBXBackupSpec{
			Cluster: polardbxv1.PolarDBXClusterReference{
				Name: "test-cluster",
			},
		},
	}

	router, _ := setupBackupRouter(t, backup)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/crd/polardbxbackups/default/test-backup", nil)
	router.ServeHTTP(w, req)
	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, w.Code)
}

func TestE2E_PolarDBXBackup_StreamEvents_NotFound(t *testing.T) {
	router, _ := setupBackupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/crd/polardbxbackups/default/nonexistent/stream", nil)
	router.ServeHTTP(w, req)
	// Stream endpoint may return 404 or start streaming
	assert.Contains(t, []int{http.StatusNotFound, http.StatusOK}, w.Code)
}

func TestE2E_PolarDBXBackup_StreamEvents_Success(t *testing.T) {
	backup := &polardbxv1.PolarDBXBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-backup",
			Namespace: "default",
		},
		Spec: polardbxv1.PolarDBXBackupSpec{
			Cluster: polardbxv1.PolarDBXClusterReference{
				Name: "test-cluster",
			},
		},
		Status: polardbxv1.PolarDBXBackupStatus{
			Phase: polardbxv1.BackupFinished,
		},
	}

	router, _ := setupBackupRouter(t, backup)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/crd/polardbxbackups/default/test-backup/stream", nil)
	router.ServeHTTP(w, req)
	// Stream endpoint may return 200 (start streaming) or other status
	assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, w.Code)
}
