package handler

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
	outgoing := map[string][]string{"lb1": {"s1"}}
	w := checkSPOF(nodes, outgoing)
	if len(w) != 1 || w[0].Rule != "spof" {
		t.Errorf("expected 1 spof warning, got %d", len(w))
	}
}

func TestCheckSPOF_SingleServiceMultipleReplicas(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{"replicas": float64(3), "stateless": true}},
	}
	outgoing := map[string][]string{"lb1": {"s1"}}
	w := checkSPOF(nodes, outgoing)
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
	outgoing := map[string][]string{"lb1": {"s1", "s2"}}
	w := checkSPOF(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multiple services, got %d", len(w))
	}
}

func TestCheckSPOF_ReverseProxyAlsoChecked(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{"replicas": float64(1)}},
	}
	outgoing := map[string][]string{"rp1": {"s1"}}
	w := checkSPOF(nodes, outgoing)
	if len(w) != 1 {
		t.Errorf("expected 1 spof warning for reverse proxy, got %d", len(w))
	}
}

// --- checkLBSPOF ---

func TestCheckLBSPOF_SingleLBNoReplica(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "lb1", ComponentType: "load_balancer", Label: "LB", Properties: map[string]interface{}{}},
	}
	w := checkLBSPOF(nodes)
	if len(w) != 1 || w[0].Rule != "lb_spof" {
		t.Errorf("expected 1 lb_spof warning, got %d", len(w))
	}
}

func TestCheckLBSPOF_SingleLBWithReplicas(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "lb1", ComponentType: "load_balancer", Label: "LB", Properties: map[string]interface{}{"replicas": float64(2)}},
	}
	w := checkLBSPOF(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

func TestCheckLBSPOF_MultipleLBs(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "lb1", ComponentType: "load_balancer", Label: "LB-1", Properties: map[string]interface{}{}},
		{ID: "lb2", ComponentType: "load_balancer", Label: "LB-2", Properties: map[string]interface{}{}},
	}
	w := checkLBSPOF(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multiple LBs, got %d", len(w))
	}
}

// --- checkNoAutoScalingSingle ---

func TestCheckNoAutoScalingSingle_Triggered(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas": float64(1), "autoScaling": false,
		}},
	}
	w := checkNoAutoScalingSingle(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckNoAutoScalingSingle_AutoScalingOn(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas": float64(1), "autoScaling": true,
		}},
	}
	w := checkNoAutoScalingSingle(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

func TestCheckNoAutoScalingSingle_MultipleReplicas(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas": float64(3), "autoScaling": false,
		}},
	}
	w := checkNoAutoScalingSingle(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for multi-replica, got %d", len(w))
	}
}

// --- checkServerlessReplicas ---

func TestCheckServerlessReplicas_Triggered(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "Lambda", Properties: map[string]interface{}{
			"computeType": "serverless", "replicas": float64(3),
		}},
	}
	w := checkServerlessReplicas(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckServerlessReplicas_SingleReplica(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "s1", ComponentType: "service", Label: "Lambda", Properties: map[string]interface{}{
			"computeType": "serverless", "replicas": float64(1),
		}},
	}
	w := checkServerlessReplicas(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkNoHealthCheckBehindLB ---

func TestCheckNoHealthCheckBehindLB_NoHealthCheck(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		{ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"healthCheck": false, "replicas": float64(1),
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "lb1", Target: "s1", ConnectionType: "sync"},
	}
	w := checkNoHealthCheckBehindLB(nodes, edges)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckNoHealthCheckBehindLB_WithHealthCheck(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		{ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"healthCheck": true, "replicas": float64(1),
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "lb1", Target: "s1", ConnectionType: "sync"},
	}
	w := checkNoHealthCheckBehindLB(nodes, edges)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}
