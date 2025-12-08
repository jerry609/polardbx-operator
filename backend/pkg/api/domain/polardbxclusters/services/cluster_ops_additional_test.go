package services

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func fakeClient() client.Client {
	scheme := runtime.NewScheme()
	_ = polardbxv1.AddToScheme(scheme)
	return crfake.NewClientBuilder().WithScheme(scheme).Build()
}

func TestUpdateLogConfig_NoClientUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns"},
		{Key: "name", Value: "demo"},
		{Key: "nodeType", Value: "cn"},
	}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_ = NewClusterService().UpdateLogConfig(c, c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateLogConfig_InvalidNodeType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return &polardbxv1.PolarDBXCluster{}, nil
	}
	defer func() { patchClusterJSON = orig }()
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns"},
		{Key: "name", Value: "demo"},
		{Key: "nodeType", Value: "invalid"},
	}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_ = NewClusterService().UpdateLogConfig(c, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScale_NoReplicaChanges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{
		{Key: "namespace", Value: "ns"},
		{Key: "name", Value: "demo"},
	}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_ = NewClusterService().Scale(c, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestScale_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{`))
	c.Request.Header.Set("Content-Type", "application/json")

	_ = NewClusterService().Scale(c, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpgrade_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{`))
	c.Request.Header.Set("Content-Type", "application/json")

	_ = NewClusterService().Upgrade(c, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateLogConfig_PatchFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}, {Key: "nodeType", Value: "cn"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"logLevel":"info"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return nil, assert.AnError
	}
	defer func() { patchClusterJSON = orig }()

	_ = NewClusterService().UpdateLogConfig(c, c)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestUpdateLogConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}, {Key: "nodeType", Value: "cn"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"logLevel":"info"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return &polardbxv1.PolarDBXCluster{}, nil
	}
	defer func() { patchClusterJSON = orig }()

	_ = NewClusterService().UpdateLogConfig(c, c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestScale_PatchFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"cnReplicas":2}`))
	c.Request.Header.Set("Content-Type", "application/json")

	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return nil, assert.AnError
	}
	defer func() { patchClusterJSON = orig }()

	_ = NewClusterService().Scale(c, c)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestScale_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"cnReplicas":2}`))
	c.Request.Header.Set("Content-Type", "application/json")

	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return &polardbxv1.PolarDBXCluster{}, nil
	}
	defer func() { patchClusterJSON = orig }()

	_ = NewClusterService().Scale(c, c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpgrade_MissingTargetVersion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	_ = NewClusterService().Upgrade(c, c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpgrade_PatchFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"targetVersion":"5.4.19"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return nil, assert.AnError
	}
	defer func() { patchClusterJSON = orig }()

	_ = NewClusterService().Upgrade(c, c)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestUpgrade_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", fakeClient())
	c.Params = gin.Params{{Key: "namespace", Value: "ns"}, {Key: "name", Value: "demo"}}
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"targetVersion":"5.4.19"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	orig := patchClusterJSON
	patchClusterJSON = func(ctx context.Context, cli client.Client, namespace, name string, patch []byte) (*polardbxv1.PolarDBXCluster, error) {
		return &polardbxv1.PolarDBXCluster{}, nil
	}
	defer func() { patchClusterJSON = orig }()

	_ = NewClusterService().Upgrade(c, c)

	assert.Equal(t, http.StatusOK, w.Code)
}
