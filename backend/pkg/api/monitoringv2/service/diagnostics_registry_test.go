package service

import (
	"context"
	"testing"
)

func TestListDefaultDiagnosticProbes(t *testing.T) {
	ids := ListDefaultDiagnosticProbes()
	expected := []string{
		"grafana-connectivity",
		"k8s-events",
		"operator-events",
		"prometheus-health",
	}
	if len(ids) != len(expected) {
		t.Fatalf("unexpected length: got %d want %d", len(ids), len(expected))
	}
	for i, id := range ids {
		if id != expected[i] {
			t.Fatalf("unexpected id at %d: got %s want %s", i, id, expected[i])
		}
	}
}

func TestResolveDiagnosticProbePrometheus(t *testing.T) {
	probe := ResolveDiagnosticProbe("prometheus-health")
	if probe == nil {
		t.Fatal("expected probe instance")
	}
	if probe.ID() != "prometheus-health" {
		t.Fatalf("unexpected id: %s", probe.ID())
	}
	if _, err := probe.Run(context.Background(), ProbeInput{}); err == nil {
		t.Fatal("expected error when running without kubernetes client")
	}
}

func TestResolveDiagnosticProbeUnknown(t *testing.T) {
	if probe := ResolveDiagnosticProbe("unknown-probe"); probe != nil {
		t.Fatalf("expected nil probe for unknown id")
	}
}
