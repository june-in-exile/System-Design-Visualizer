package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkMQConsumer ---

func TestCheckMQConsumer_NoOutgoing(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"mq1": {ID: "mq1", ComponentType: "message_queue", Label: "RabbitMQ"},
	}
	outgoing := map[string][]string{}
	w := checkMQConsumer(nodes, outgoing)
	if len(w) != 1 || w[0].Rule != "mq_consumer_missing" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckMQConsumer_WithConsumer(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"mq1": {ID: "mq1", ComponentType: "message_queue", Label: "RabbitMQ"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "Worker"},
	}
	outgoing := map[string][]string{"mq1": {"s1"}}
	w := checkMQConsumer(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkMQDLQ ---

func TestCheckMQDLQ_NoDLQ(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "mq1", ComponentType: "message_queue", Label: "RabbitMQ", Properties: map[string]interface{}{
			"hasDLQ": false,
		}},
	}
	w := checkMQDLQ(nodes)
	if len(w) != 1 || w[0].Rule != "mq_dlq_missing" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckMQDLQ_WithDLQ(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "mq1", ComponentType: "message_queue", Label: "RabbitMQ", Properties: map[string]interface{}{
			"hasDLQ": true,
		}},
	}
	w := checkMQDLQ(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkAsyncDecoupling ---

func TestCheckAsyncDecoupling_SyncCallToHeavyService(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API"},
		"s2": {ID: "s2", ComponentType: "service", Label: "Email Service"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "s1", Target: "s2", ConnectionType: "sync"},
	}
	w := checkAsyncDecoupling(nodes, edges)
	if len(w) != 1 || w[0].Rule != "async_decoupling" {
		t.Errorf("expected 1 async_decoupling warning, got %d", len(w))
	}
}

func TestCheckAsyncDecoupling_AlreadyDecoupled(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"mq1": {ID: "mq1", ComponentType: "message_queue", Label: "RabbitMQ"},
		"s2":  {ID: "s2", ComponentType: "service", Label: "Email Service"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "mq1", Target: "s2", ConnectionType: "async"},
	}
	w := checkAsyncDecoupling(nodes, edges)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings when decoupled via MQ, got %d", len(w))
	}
}

func TestCheckAsyncDecoupling_NonHeavyService_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API"},
		"s2": {ID: "s2", ComponentType: "service", Label: "Auth Service"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "s1", Target: "s2", ConnectionType: "sync"},
	}
	w := checkAsyncDecoupling(nodes, edges)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for non-heavy service, got %d", len(w))
	}
}

// --- checkAsyncPeakShaving ---

func TestCheckAsyncPeakShaving_LBToHighWriteDB(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"db1": {ID: "db1", ComponentType: "database", Label: "WriteDB", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.2,
		}},
	}
	outgoing := map[string][]string{"lb1": {"db1"}}
	w := checkAsyncPeakShaving(nodes, outgoing)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckAsyncPeakShaving_LBToHighReadDB_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"db1": {ID: "db1", ComponentType: "database", Label: "ReadDB", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.8,
		}},
	}
	outgoing := map[string][]string{"lb1": {"db1"}}
	w := checkAsyncPeakShaving(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}
