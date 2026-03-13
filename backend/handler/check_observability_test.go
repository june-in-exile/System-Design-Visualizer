package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkMissingLogger ---

func TestCheckMissingLogger_3Services_NoLogger(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2": {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3": {ID: "s3", ComponentType: "service", Label: "Svc3"},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkMissingLogger(ctx)
	if len(w) != 1 || w[0].Rule != "missing_observability" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckMissingLogger_2Services_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2": {ID: "s2", ComponentType: "service", Label: "Svc2"},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkMissingLogger(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for <3 services, got %d", len(w))
	}
}

func TestCheckMissingLogger_LoggerNotConnected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3":  {ID: "s3", ComponentType: "service", Label: "Svc3"},
		"log": {ID: "log", ComponentType: "logger", Label: "ELK"},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkMissingLogger(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for disconnected logger, got %d", len(w))
	}
}

func TestCheckMissingLogger_LoggerConnected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3":  {ID: "s3", ComponentType: "service", Label: "Svc3"},
		"log": {ID: "log", ComponentType: "logger", Label: "ELK"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "s1", Target: "log", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkMissingLogger(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkIncompleteObservability ---

func TestCheckIncompleteObservability_MetricsOnly(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3":  {ID: "s3", ComponentType: "service", Label: "Svc3"},
		"log": {ID: "log", ComponentType: "logger", Label: "Prometheus", Properties: map[string]interface{}{"logType": "metrics", "alerting": true}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkIncompleteObservability(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for metrics-only, got %d", len(w))
	}
}

func TestCheckIncompleteObservability_AllType_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3":  {ID: "s3", ComponentType: "service", Label: "Svc3"},
		"log": {ID: "log", ComponentType: "logger", Label: "Datadog", Properties: map[string]interface{}{"logType": "all"}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkIncompleteObservability(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for all type, got %d", len(w))
	}
}

func TestCheckIncompleteObservability_LogsOnly(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3":  {ID: "s3", ComponentType: "service", Label: "Svc3"},
		"log": {ID: "log", ComponentType: "logger", Label: "ELK", Properties: map[string]interface{}{"logType": "logs"}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkIncompleteObservability(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for logs-only, got %d", len(w))
	}
}

func TestCheckIncompleteObservability_TracesOnly(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Svc2"},
		"s3":  {ID: "s3", ComponentType: "service", Label: "Svc3"},
		"log": {ID: "log", ComponentType: "logger", Label: "Jaeger", Properties: map[string]interface{}{"logType": "traces"}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkIncompleteObservability(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for traces-only, got %d", len(w))
	}
}

// --- checkAlertingDisabled ---

func TestCheckAlertingDisabled_Off(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"log": {ID: "log", ComponentType: "logger", Label: "ELK", Properties: map[string]interface{}{"logType": "all", "alerting": false}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkAlertingDisabled(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckAlertingDisabled_On_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"log": {ID: "log", ComponentType: "logger", Label: "ELK", Properties: map[string]interface{}{"logType": "all", "alerting": true}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkAlertingDisabled(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}
