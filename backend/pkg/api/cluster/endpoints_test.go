package cluster

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	polardbxv1 "github.com/alibaba/polardbx-operator/api/v1"
	polardbx "github.com/alibaba/polardbx-operator/api/v1/polardbx"

	"github.com/gin-gonic/gin"
)

func hmacHex(secret, plain string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(plain))
	return hex.EncodeToString(h.Sum(nil))
}

func buildContextWithClient(t *testing.T, objs ...runtime.Object) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = polardbxv1.AddToScheme(scheme)
	cli := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("k8sClient", cli)
	return c, w
}

func TestScale_PrecheckEnforced_MissingToken(t *testing.T) {
	ns, name := "ns", "pxc"
	cluster := &polardbxv1.PolarDBXCluster{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, Spec: polardbxv1.PolarDBXClusterSpec{Topology: polardbx.Topology{Version: "8.0.18"}}}
	cfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "polardbx-ui-backend-config", Namespace: "polardbx-operator-system"}, Data: map[string]string{"prechange.enforce": "true"}}
	c, w := buildContextWithClient(t, cluster, cfg)

	body, _ := json.Marshal(map[string]any{"cnReplicas": 2})
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/clusters/ns/pxc/scale", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "namespace", Value: ns}, {Key: "name", Value: name}}

	Scale(c)
	if w.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412, got %d", w.Code)
	}
}

func TestScale_PrecheckEnforced_WithTokenAndSignature(t *testing.T) {
	ns, name := "ns", "pxc"
	secret := "unit-test-secret"
	cluster := &polardbxv1.PolarDBXCluster{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, Spec: polardbxv1.PolarDBXClusterSpec{Topology: polardbx.Topology{Version: "8.0.18"}}}
	cfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "polardbx-ui-backend-config", Namespace: "polardbx-operator-system"}, Data: map[string]string{"prechange.enforce": "true", "precheck.secret": secret}}
	c, w := buildContextWithClient(t, cluster, cfg)

	body, _ := json.Marshal(map[string]any{"cnReplicas": 2})
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/clusters/ns/pxc/scale", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "namespace", Value: ns}, {Key: "name", Value: name}}

	// Build token and signature
	plain := fmt.Sprintf("%d:%s:%s/%s", time.Now().UnixNano(), "scale", ns, name)
	sig := hmacHex(secret, plain)
	c.Request.Header.Set("X-Precheck-Token", plain)
	c.Request.Header.Set("X-Precheck-Token-Signature", sig)

	Scale(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestScale_PrecheckEnforced_BadSignature(t *testing.T) {
	ns, name := "ns", "pxc"
	secret := "unit-test-secret"
	cluster := &polardbxv1.PolarDBXCluster{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, Spec: polardbxv1.PolarDBXClusterSpec{Topology: polardbx.Topology{Version: "8.0.18"}}}
	cfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "polardbx-ui-backend-config", Namespace: "polardbx-operator-system"}, Data: map[string]string{"prechange.enforce": "true", "precheck.secret": secret}}
	c, w := buildContextWithClient(t, cluster, cfg)

	body, _ := json.Marshal(map[string]any{"cnReplicas": 2})
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/clusters/ns/pxc/scale", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "namespace", Value: ns}, {Key: "name", Value: name}}

	plain := fmt.Sprintf("%d:%s:%s/%s", time.Now().UnixNano(), "scale", ns, name)
	// wrong signature
	c.Request.Header.Set("X-Precheck-Token", plain)
	c.Request.Header.Set("X-Precheck-Token-Signature", "badsignature")

	Scale(c)
	if w.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 for bad signature, got %d", w.Code)
	}
}

func TestScale_PrecheckEnforced_ExpiredToken(t *testing.T) {
	ns, name := "ns", "pxc"
	secret := "unit-test-secret"
	cluster := &polardbxv1.PolarDBXCluster{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, Spec: polardbxv1.PolarDBXClusterSpec{Topology: polardbx.Topology{Version: "8.0.18"}}}
	cfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "polardbx-ui-backend-config", Namespace: "polardbx-operator-system"}, Data: map[string]string{"prechange.enforce": "true", "precheck.secret": secret}}
	c, w := buildContextWithClient(t, cluster, cfg)

	body, _ := json.Marshal(map[string]any{"cnReplicas": 2})
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/clusters/ns/pxc/scale", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "namespace", Value: ns}, {Key: "name", Value: name}}

	// build expired token (-11m)
	ts := time.Now().Add(-11 * time.Minute).UnixNano()
	plain := fmt.Sprintf("%d:%s:%s/%s", ts, "scale", ns, name)
	sig := hmacHex(secret, plain)
	c.Request.Header.Set("X-Precheck-Token", plain)
	c.Request.Header.Set("X-Precheck-Token-Signature", sig)

	Scale(c)
	if w.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 for expired token, got %d", w.Code)
	}
}

func TestUpgrade_PrecheckEnforced_WrongOpToken(t *testing.T) {
	ns, name := "ns", "pxc"
	secret := "unit-test-secret"
	cluster := &polardbxv1.PolarDBXCluster{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}, Spec: polardbxv1.PolarDBXClusterSpec{Topology: polardbx.Topology{Version: "8.0.18"}}}
	cfg := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "polardbx-ui-backend-config", Namespace: "polardbx-operator-system"}, Data: map[string]string{"prechange.enforce": "true", "precheck.secret": secret}}
	c, w := buildContextWithClient(t, cluster, cfg)

	body, _ := json.Marshal(map[string]any{"targetVersion": "8.0.19"})
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/clusters/ns/pxc/upgrade", bytes.NewReader(body))
	c.Params = gin.Params{{Key: "namespace", Value: ns}, {Key: "name", Value: name}}

	// token with scale op but calling upgrade => should fail
	plain := fmt.Sprintf("%d:%s:%s/%s", time.Now().UnixNano(), "scale", ns, name)
	sig := hmacHex(secret, plain)
	c.Request.Header.Set("X-Precheck-Token", plain)
	c.Request.Header.Set("X-Precheck-Token-Signature", sig)

	Upgrade(c)
	if w.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 for wrong op token, got %d", w.Code)
	}
}
