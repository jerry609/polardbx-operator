package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"polardbx-ui-backend/pkg/api/domain/xstores/k8srepo"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbxv1xstore "github.com/alibaba/polardbx-operator/api/v1/xstore"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// ==================== Followers Retry/Cancel Tests ====================

// mockFollowerRepo is a mock repository for testing Retry and Cancel
type mockFollowerRepo struct {
	getFollowerErr    error
	deleteFollowerErr error
	createFollowerErr error
	updateFollowerErr error

	follower *polardbxv1.XStoreFollower
}

func (m *mockFollowerRepo) List(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.XStore, error) {
	return nil, nil
}
func (m *mockFollowerRepo) Create(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	return nil, nil
}
func (m *mockFollowerRepo) Get(ctx context.Context, cli client.Client, namespace, name string) (*polardbxv1.XStore, error) {
	return nil, nil
}
func (m *mockFollowerRepo) Update(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	return nil, nil
}
func (m *mockFollowerRepo) Delete(ctx context.Context, cli client.Client, namespace, name string) error {
	return nil
}
func (m *mockFollowerRepo) ListPods(ctx context.Context, cli client.Client, namespace, xstoreName string) ([]corev1.Pod, error) {
	return nil, nil
}
func (m *mockFollowerRepo) ListBackups(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.XStoreBackup, error) {
	return nil, nil
}
func (m *mockFollowerRepo) CreateBackup(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	return nil, nil
}
func (m *mockFollowerRepo) GetBackup(ctx context.Context, cli client.Client, namespace, name string) (*polardbxv1.XStoreBackup, error) {
	return nil, nil
}
func (m *mockFollowerRepo) UpdateBackup(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	return nil, nil
}
func (m *mockFollowerRepo) DeleteBackup(ctx context.Context, cli client.Client, namespace, name string) error {
	return nil
}
func (m *mockFollowerRepo) ListFollowers(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.XStoreFollower, error) {
	return nil, nil
}
func (m *mockFollowerRepo) CreateFollower(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	if m.createFollowerErr != nil {
		return nil, m.createFollowerErr
	}
	return obj, nil
}
func (m *mockFollowerRepo) GetFollower(ctx context.Context, cli client.Client, namespace, name string) (*polardbxv1.XStoreFollower, error) {
	if m.getFollowerErr != nil {
		return nil, m.getFollowerErr
	}
	if m.follower != nil {
		return m.follower, nil
	}
	return &polardbxv1.XStoreFollower{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: polardbxv1.XStoreFollowerStatus{
			Phase: polardbxv1xstore.FollowerPhaseFailed,
		},
	}, nil
}
func (m *mockFollowerRepo) UpdateFollower(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	if m.updateFollowerErr != nil {
		return nil, m.updateFollowerErr
	}
	return obj, nil
}
func (m *mockFollowerRepo) DeleteFollower(ctx context.Context, cli client.Client, namespace, name string) error {
	return m.deleteFollowerErr
}

func setupFollowersRetryRouter(t *testing.T, repo k8srepo.XStoreRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()

	scheme := runtime.NewScheme()
	_ = polardbxv1.AddToScheme(scheme)
	cli := crfake.NewClientBuilder().WithScheme(scheme).Build()

	r.Use(func(c *gin.Context) {
		c.Set("k8sClient", cli)
		c.Next()
	})

	svc := &FollowersService{repo: repo}
	r.POST("/followers/:namespace/:name/retry", svc.Retry)
	r.POST("/followers/:namespace/:name/cancel", svc.Cancel)

	return r
}

func TestFollowersService_Retry_Success(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseFailed,
			},
		},
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/retry", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowersService_Retry_GetError(t *testing.T) {
	repo := &mockFollowerRepo{
		getFollowerErr: errors.New("get error"),
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/retry", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code) // K8s connection error returns 502
}

func TestFollowersService_Retry_NotFailed(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseBackup, // Not failed
			},
		},
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/retry", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFollowersService_Retry_DeleteError(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseFailed,
			},
		},
		deleteFollowerErr: errors.New("delete error"),
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/retry", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code) // K8s connection error returns 502
}

func TestFollowersService_Retry_CreateError(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseFailed,
			},
		},
		createFollowerErr: errors.New("create error"),
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/retry", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code) // K8s connection error returns 502
}

func TestFollowersService_Cancel_Success(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseBackup, // In progress
			},
		},
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowersService_Cancel_GetError(t *testing.T) {
	repo := &mockFollowerRepo{
		getFollowerErr: errors.New("get error"),
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code) // K8s connection error returns 502
}

func TestFollowersService_Cancel_AlreadyCompleted(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseSuccess, // Already completed
			},
		},
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFollowersService_Cancel_DeleteError(t *testing.T) {
	repo := &mockFollowerRepo{
		follower: &polardbxv1.XStoreFollower{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-follower",
				Namespace: "default",
			},
			Status: polardbxv1.XStoreFollowerStatus{
				Phase: polardbxv1xstore.FollowerPhaseBackup, // In progress
			},
		},
		deleteFollowerErr: errors.New("delete error"),
	}
	router := setupFollowersRetryRouter(t, repo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test-follower/cancel", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code) // K8s connection error returns 502
}

func TestFollowersService_Retry_NoK8sClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No k8sClient middleware
	svc := &FollowersService{repo: &mockFollowerRepo{}}
	r.POST("/followers/:namespace/:name/retry", svc.Retry)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test/retry", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code) // No K8s client returns 401
}

func TestFollowersService_Cancel_NoK8sClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// No k8sClient middleware
	svc := &FollowersService{repo: &mockFollowerRepo{}}
	r.POST("/followers/:namespace/:name/cancel", svc.Cancel)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/followers/default/test/cancel", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code) // No K8s client returns 401
}

func TestNewFollowersService(t *testing.T) {
	svc := NewFollowersService()
	require.NotNil(t, svc)
}
