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
	w := checkMissingLogger(nodes, map[string][]string{})
	if len(w) != 1 || w[0].Rule != "missing_observability" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckMissingLogger_2Services_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "Svc1"},
		"s2": {ID: "s2", ComponentType: "service", Label: "Svc2"},
	}
	w := checkMissingLogger(nodes, map[string][]string{})
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
	outgoing := map[string][]string{}
	w := checkMissingLogger(nodes, outgoing)
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
	outgoing := map[string][]string{"s1": {"log"}}
	w := checkMissingLogger(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkIncompleteObservability ---

func TestCheckIncompleteObservability_MetricsOnly(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "Svc1"},
		{ID: "s2", ComponentType: "service", Label: "Svc2"},
		{ID: "s3", ComponentType: "service", Label: "Svc3"},
		{ID: "log", ComponentType: "logger", Label: "Prometheus", Properties: map[string]interface{}{
			"logType": "metrics", "alerting": true,
		}},
	}
	w := checkIncompleteObservability(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for metrics-only, got %d", len(w))
	}
}

func TestCheckIncompleteObservability_AllType_NoWarning(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "Svc1"},
		{ID: "s2", ComponentType: "service", Label: "Svc2"},
		{ID: "s3", ComponentType: "service", Label: "Svc3"},
		{ID: "log", ComponentType: "logger", Label: "Datadog", Properties: map[string]interface{}{
			"logType": "all",
		}},
	}
	w := checkIncompleteObservability(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for all type, got %d", len(w))
	}
}

func TestCheckIncompleteObservability_LogsOnly(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "Svc1"},
		{ID: "s2", ComponentType: "service", Label: "Svc2"},
		{ID: "s3", ComponentType: "service", Label: "Svc3"},
		{ID: "log", ComponentType: "logger", Label: "ELK", Properties: map[string]interface{}{
			"logType": "logs",
		}},
	}
	w := checkIncompleteObservability(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for logs-only, got %d", len(w))
	}
}

func TestCheckIncompleteObservability_TracesOnly(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "Svc1"},
		{ID: "s2", ComponentType: "service", Label: "Svc2"},
		{ID: "s3", ComponentType: "service", Label: "Svc3"},
		{ID: "log", ComponentType: "logger", Label: "Jaeger", Properties: map[string]interface{}{
			"logType": "traces",
		}},
	}
	w := checkIncompleteObservability(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for traces-only, got %d", len(w))
	}
}

// --- checkAlertingDisabled ---

func TestCheckAlertingDisabled_Off(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "log", ComponentType: "logger", Label: "ELK", Properties: map[string]interface{}{
			"logType": "all", "alerting": false,
		}},
	}
	w := checkAlertingDisabled(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckAlertingDisabled_On_NoWarning(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "log", ComponentType: "logger", Label: "ELK", Properties: map[string]interface{}{
			"logType": "all", "alerting": true,
		}},
	}
	w := checkAlertingDisabled(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}
