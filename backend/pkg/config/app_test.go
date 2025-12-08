package config

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

func TestMaskSecret(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "<empty>"},
		{"abc", "***"},
		{"abcd", "***"},
		{"abcdef", "ab***ef"},
	}

	for _, tt := range tests {
		if got := maskSecret(tt.in); got != tt.want {
			t.Fatalf("maskSecret(%s) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestPrintConfigRedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	origOut := log.Writer()
	origFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(origOut)
		log.SetFlags(origFlags)
	}()

	appCfg := &AppConfig{
		Server: &ServerConfig{
			Port:          8080,
			Mode:          "debug",
			LogLevel:      "info",
			JWTSecret:     "super-secret-token",
			JWTExpiration: 24 * time.Hour,
		},
		Kubernetes: &KubeConfig{ConfigPath: "/tmp/kube", DefaultNamespace: "default"},
		Image:      &ImageRegistryConfig{DefaultRegistry: "docker.m.daocloud.io"},
		AutoFix:    &AutoFixOverlayConfig{Enabled: true, Namespace: "ns"},
	}

	appCfg.PrintConfig()

	out := buf.String()
	if strings.Contains(out, "super-secret-token") {
		t.Fatalf("PrintConfig leaked sensitive value: %s", out)
	}
	if !strings.Contains(out, "JWTSecret=su***en") {
		t.Fatalf("masked secret not present: %s", out)
	}
}
