package logic

import (
	"testing"

	"github.com/architectmind/backend/model"
)

func TestHighQPSNoCache(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "svc1", ComponentType: "service", Label: "Service", Properties: map[string]interface{}{"replicas": 2}},
		},
		Params: &model.SystemParams{PeakQPS: 10000},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkHighQPSNoCache(ctx)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
	if warnings[0].Rule != "high_qps_no_cache" {
		t.Errorf("expected rule high_qps_no_cache, got %s", warnings[0].Rule)
	}
}

func TestHighQPSWithCache(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "svc1", ComponentType: "service", Label: "Service", Properties: map[string]interface{}{}},
			{ID: "cache1", ComponentType: "cache", Label: "Cache", Properties: map[string]interface{}{}},
		},
		Params: &model.SystemParams{PeakQPS: 10000},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkHighQPSNoCache(ctx)
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(warnings))
	}
}

func TestHighQPSSingleLB(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "lb1", ComponentType: "load_balancer", Label: "LB", Properties: map[string]interface{}{"replicas": 1}},
		},
		Params: &model.SystemParams{PeakQPS: 15000},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkHighQPSSingleLB(ctx)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
}

func TestNoParamsSkipsCapacity(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "svc1", ComponentType: "service", Label: "Service", Properties: map[string]interface{}{}},
		},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkCapacity(ctx)
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings when params is nil, got %d", len(warnings))
	}
}

func TestHighAvailabilityInsufficientReplicas(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "svc1", ComponentType: "service", Label: "Service", Properties: map[string]interface{}{"replicas": 1}},
		},
		Params: &model.SystemParams{Availability: "99.99%"},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkHighAvailabilityInsufficientReplicas(ctx)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
}

func TestHighDAUNoCDN(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "svc1", ComponentType: "service", Label: "Service", Properties: map[string]interface{}{}},
		},
		Params: &model.SystemParams{DAU: 500000},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkHighDAUNoCDN(ctx)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
}

func TestReadHeavyNoReplica(t *testing.T) {
	topo := model.SystemTopology{
		ID: "test", Name: "test", Version: 1,
		Nodes: []model.SystemNode{
			{ID: "db1", ComponentType: "database", Label: "DB", Properties: map[string]interface{}{"dbType": "sql"}},
		},
		Params: &model.SystemParams{ReadWriteRatio: 0.9},
	}
	ctx := model.NewTopologyContext(topo)
	warnings := checkReadHeavyNoReadReplica(ctx)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
}
