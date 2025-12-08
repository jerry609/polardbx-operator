package services

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
)

type stubXRepo struct {
	listErr   error
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	podErr    error

	backupListErr   error
	backupCreateErr error
	backupGetErr    error
	backupUpdateErr error
	backupDeleteErr error

	followerListErr   error
	followerCreateErr error
	followerGetErr    error
	followerUpdateErr error
	followerDeleteErr error
}

func (s *stubXRepo) List(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.XStore, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return []polardbxv1.XStore{{}}, nil
}

func (s *stubXRepo) Create(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	if s.createErr != nil {
		return nil, s.createErr
	}
	return obj, nil
}

func (s *stubXRepo) Get(ctx context.Context, cli client.Client, namespace, name string) (*polardbxv1.XStore, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	return &polardbxv1.XStore{}, nil
}

func (s *stubXRepo) Update(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStore) (*polardbxv1.XStore, error) {
	if s.updateErr != nil {
		return nil, s.updateErr
	}
	return obj, nil
}

func (s *stubXRepo) Delete(ctx context.Context, cli client.Client, namespace, name string) error {
	return s.deleteErr
}

func (s *stubXRepo) ListPods(ctx context.Context, cli client.Client, namespace, xstoreName string) ([]corev1.Pod, error) {
	if s.podErr != nil {
		return nil, s.podErr
	}
	return []corev1.Pod{{}}, nil
}

// Unused methods for backups/followers to satisfy interface (noop).
func (s *stubXRepo) ListBackups(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.XStoreBackup, error) {
	if s.backupListErr != nil {
		return nil, s.backupListErr
	}
	return []polardbxv1.XStoreBackup{{}}, nil
}
func (s *stubXRepo) CreateBackup(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	if s.backupCreateErr != nil {
		return nil, s.backupCreateErr
	}
	return obj, nil
}
func (s *stubXRepo) GetBackup(ctx context.Context, cli client.Client, namespace, name string) (*polardbxv1.XStoreBackup, error) {
	if s.backupGetErr != nil {
		return nil, s.backupGetErr
	}
	return &polardbxv1.XStoreBackup{}, nil
}
func (s *stubXRepo) UpdateBackup(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreBackup) (*polardbxv1.XStoreBackup, error) {
	if s.backupUpdateErr != nil {
		return nil, s.backupUpdateErr
	}
	return obj, nil
}
func (s *stubXRepo) DeleteBackup(ctx context.Context, cli client.Client, namespace, name string) error {
	return s.backupDeleteErr
}
func (s *stubXRepo) ListFollowers(ctx context.Context, cli client.Client, namespace string) ([]polardbxv1.XStoreFollower, error) {
	if s.followerListErr != nil {
		return nil, s.followerListErr
	}
	return []polardbxv1.XStoreFollower{{}}, nil
}
func (s *stubXRepo) CreateFollower(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	if s.followerCreateErr != nil {
		return nil, s.followerCreateErr
	}
	return obj, nil
}
func (s *stubXRepo) GetFollower(ctx context.Context, cli client.Client, namespace, name string) (*polardbxv1.XStoreFollower, error) {
	if s.followerGetErr != nil {
		return nil, s.followerGetErr
	}
	return &polardbxv1.XStoreFollower{}, nil
}
func (s *stubXRepo) UpdateFollower(ctx context.Context, cli client.Client, namespace string, obj *polardbxv1.XStoreFollower) (*polardbxv1.XStoreFollower, error) {
	if s.followerUpdateErr != nil {
		return nil, s.followerUpdateErr
	}
	return obj, nil
}
func (s *stubXRepo) DeleteFollower(ctx context.Context, cli client.Client, namespace, name string) error {
	return s.followerDeleteErr
}

func ginCtx(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	buf := bytes.NewBuffer(nil)
	if body != nil {
		buf = bytes.NewBuffer(body)
	}
	c.Request = httptest.NewRequest(method, path, buf)
	if body != nil {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return c, w
}

func fakeClient() client.Client {
	scheme := runtime.NewScheme()
	_ = polardbxv1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).Build()
}

func TestXStoreService_UnauthorizedWhenNoClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{}}

	c, w := ginCtx(http.MethodGet, "/", nil)
	svc.List(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	c, w = ginCtx(http.MethodPost, "/", []byte(`{}`))
	svc.Create(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestXStoreService_List_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{listErr: assert.AnError}}
	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	svc.List(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	svc.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestXStoreService_Create_ValidationAndErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{createErr: assert.AnError}}
	// invalid JSON
	c, w := ginCtx(http.MethodPost, "/", []byte("{"))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// create error
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"metadata":{"name":"x1"}}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// success
	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"metadata":{"name":"x1"}}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestXStoreService_Get_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{getErr: assert.AnError}}
	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "x1"}}
	svc.Get(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "x1"}}
	svc.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestXStoreService_Update_ValidationErrorAndRepoErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{updateErr: assert.AnError}}
	// invalid json
	c, w := ginCtx(http.MethodPut, "/", []byte("{"))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// repo error
	c, w = ginCtx(http.MethodPut, "/", []byte(`{"metadata":{"name":"x1"}}`))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// success
	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodPut, "/", []byte(`{"metadata":{"name":"x1"}}`))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestXStoreService_Delete_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{deleteErr: assert.AnError}}
	c, w := ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "x1"}}
	svc.Delete(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "x1"}}
	svc.Delete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestXStoreService_ListPods_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &XStoreService{repo: &stubXRepo{podErr: assert.AnError}}
	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "x1"}}
	svc.ListPods(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "x1"}}
	svc.ListPods(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- Backups (service level) ----

func TestBackupsService_List_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupListErr: assert.AnError}}

	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	svc.List(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	svc.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupsService_Create_Validation_Error_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupCreateErr: assert.AnError}}

	// invalid json
	c, w := ginCtx(http.MethodPost, "/", []byte("{"))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// repo error
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"metadata":{"name":"b1"}}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// success
	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"metadata":{"name":"b1"}}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBackupsService_Get_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupGetErr: assert.AnError}}
	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.Get(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupsService_Update_Validation_Error_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupUpdateErr: assert.AnError}}

	c, w := ginCtx(http.MethodPut, "/", []byte("{"))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	c, w = ginCtx(http.MethodPut, "/", []byte(`{"metadata":{"name":"b1"}}`))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodPut, "/", []byte(`{"metadata":{"name":"b1"}}`))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupsService_Delete_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupDeleteErr: assert.AnError}}

	c, w := ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.Delete(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.Delete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupsService_ForceDelete_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupGetErr: assert.AnError}}

	// get error
	c, w := ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.ForceDelete(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// update error
	svc.repo = &stubXRepo{backupUpdateErr: assert.AnError}
	c, w = ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.ForceDelete(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// success
	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.ForceDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBackupsService_RemoteInfo_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &BackupsService{repo: &stubXRepo{backupGetErr: assert.AnError}}

	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.RemoteInfo(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "b1"}}
	svc.RemoteInfo(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---- Followers (service level) ----

func TestFollowersService_List_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &FollowersService{repo: &stubXRepo{followerListErr: assert.AnError}}

	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	svc.List(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	svc.List(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowersService_Create_ValidationAndErrorSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &FollowersService{repo: &stubXRepo{followerCreateErr: assert.AnError}}

	// invalid json
	c, w := ginCtx(http.MethodPost, "/", []byte("{"))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// missing name
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"spec":{"xStoreName":""}}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// missing xStoreName
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"name":"f1"}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// repo error
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"name":"f1","xStoreName":"x1"}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	// success
	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodPost, "/", []byte(`{"name":"f1","xStoreName":"x1"}`))
	c.Set("k8sClient", fakeClient())
	svc.Create(c)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFollowersService_Get_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &FollowersService{repo: &stubXRepo{followerGetErr: assert.AnError}}

	c, w := ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "f1"}}
	svc.Get(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodGet, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "f1"}}
	svc.Get(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowersService_Update_ValidationError_Error_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &FollowersService{repo: &stubXRepo{followerUpdateErr: assert.AnError}}

	c, w := ginCtx(http.MethodPut, "/", []byte("{"))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	c, w = ginCtx(http.MethodPut, "/", []byte(`{"metadata":{"name":"f1"}}`))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodPut, "/", []byte(`{"metadata":{"name":"f1"}}`))
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}}
	svc.Update(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFollowersService_Delete_ErrorAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &FollowersService{repo: &stubXRepo{followerDeleteErr: assert.AnError}}

	c, w := ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "f1"}}
	svc.Delete(c)
	assert.Equal(t, http.StatusBadGateway, w.Code)

	svc.repo = &stubXRepo{}
	c, w = ginCtx(http.MethodDelete, "/", nil)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "f1"}}
	svc.Delete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}
