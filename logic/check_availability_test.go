package logic

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkSPOF ---

func TestCheckSPOF_SingleServiceBehindLB(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{"replicas": float64(1), "stateless": true}},
	}
	edges := []model.SystemEdge{{ID: "e1", Source: "lb1", Target: "s1"}}
	w := checkSPOF(makeCtx(nodes, edges))
	if len(w) != 1 || w[0].Rule != "spof" {
		t.Errorf("expected 1 spof warning, got %d", len(w))
	}
}

func TestCheckSPOF_SingleServiceMultipleReplicas(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{"replicas": float64(3), "stateless": true}},
	}
	edges := []model.SystemEdge{{ID: "e1", Source: "lb1", Target: "s1"}}
	w := checkSPOF(makeCtx(nodes, edges))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multi-replica service, got %d", len(w))
	}
}

func TestCheckSPOF_MultipleServices(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API-1", Properties: map[string]interface{}{"replicas": float64(1)}},
		"s2":  {ID: "s2", ComponentType: "service", Label: "API-2", Properties: map[string]interface{}{"replicas": float64(1)}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "lb1", Target: "s1"},
		{ID: "e2", Source: "lb1", Target: "s2"},
	}
	w := checkSPOF(makeCtx(nodes, edges))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multiple services, got %d", len(w))
	}
}

func TestCheckSPOF_ReverseProxyAlsoChecked(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{"replicas": float64(1)}},
	}
	edges := []model.SystemEdge{{ID: "e1", Source: "rp1", Target: "s1"}}
	w := checkSPOF(makeCtx(nodes, edges))
	if len(w) != 1 {
		t.Errorf("expected 1 spof warning for reverse proxy, got %d", len(w))
	}
}

// --- checkEntryPointSPOF ---

func TestCheckEntryPointSPOF_LB_SingleNoReplica(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB", Properties: map[string]interface{}{}},
	}
	w := checkEntryPointSPOF(makeCtx(nodes, nil), "load_balancer", "lb_spof", "⚖️", "Load Balancer")
	if len(w) != 1 || w[0].Rule != "lb_spof" {
		t.Errorf("expected 1 lb_spof warning, got %d", len(w))
	}
}

func TestCheckEntryPointSPOF_LB_SingleWithReplicas(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB", Properties: map[string]interface{}{"replicas": float64(2)}},
	}
	w := checkEntryPointSPOF(makeCtx(nodes, nil), "load_balancer", "lb_spof", "⚖️", "Load Balancer")
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

func TestCheckEntryPointSPOF_LB_Multiple(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB-1", Properties: map[string]interface{}{}},
		"lb2": {ID: "lb2", ComponentType: "load_balancer", Label: "LB-2", Properties: map[string]interface{}{}},
	}
	w := checkEntryPointSPOF(makeCtx(nodes, nil), "load_balancer", "lb_spof", "⚖️", "Load Balancer")
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multiple LBs, got %d", len(w))
	}
}

// --- checkNoAutoScalingSingle ---

func TestCheckNoAutoScalingSingle_Triggered(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas": float64(1), "autoScaling": false,
		}},
	}
	w := checkNoAutoScalingSingle(makeCtx(nodes, nil))
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckNoAutoScalingSingle_AutoScalingOn(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas": float64(1), "autoScaling": true,
		}},
	}
	w := checkNoAutoScalingSingle(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

func TestCheckNoAutoScalingSingle_MultipleReplicas(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas": float64(3), "autoScaling": false,
		}},
	}
	w := checkNoAutoScalingSingle(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multi-replica, got %d", len(w))
	}
}

// --- checkServerlessReplicas ---

func TestCheckServerlessReplicas_Triggered(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "Lambda", Properties: map[string]interface{}{
			"computeType": "serverless", "replicas": float64(3),
		}},
	}
	w := checkServerlessReplicas(makeCtx(nodes, nil))
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckServerlessReplicas_SingleReplica(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "Lambda", Properties: map[string]interface{}{
			"computeType": "serverless", "replicas": float64(1),
		}},
	}
	w := checkServerlessReplicas(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkNoHealthCheckBehindLB ---

func TestCheckNoHealthCheckBehindLB_NoHealthCheck(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"s1": {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"healthCheck": false, "replicas": float64(1),
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "lb1", Target: "s1", ConnectionType: "sync"},
	}
	w := checkNoHealthCheckBehindLB(makeCtx(nodes, edges))
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckNoHealthCheckBehindLB_WithHealthCheck(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"s1": {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"healthCheck": true, "replicas": float64(1),
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "lb1", Target: "s1", ConnectionType: "sync"},
	}
	w := checkNoHealthCheckBehindLB(makeCtx(nodes, edges))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

