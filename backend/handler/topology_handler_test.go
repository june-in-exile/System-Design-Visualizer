package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

func TestCheckProtocolMismatch_Unspecified(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"1": {ID: "1", ComponentType: "database", Label: "Master"},
		"2": {ID: "2", ComponentType: "database", Label: "Slave"},
	}
	edges := []model.SystemEdge{
		{
			ID:             "e1",
			Source:         "1",
			Target:         "2",
			ConnectionType: "replication",
			Protocol:       "",
		},
	}

	ctx := makeCtx(nodes, edges)
	warnings := checkProtocolMismatch(ctx)
	if len(warnings) > 0 {
		t.Errorf("Expected no warnings for unspecified protocol, got %v", warnings)
	}
}

func TestCheckProtocolMismatch_Mismatch(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"1": {ID: "1", ComponentType: "service", Label: "Service"},
		"2": {ID: "2", ComponentType: "database", Label: "Database"},
	}
	edges := []model.SystemEdge{
		{
			ID:             "e1",
			Source:         "1",
			Target:         "2",
			ConnectionType: "sync",
			Protocol:       "http",
		},
	}

	ctx := makeCtx(nodes, edges)
	warnings := checkProtocolMismatch(ctx)
	if len(warnings) == 0 {
		t.Error("Expected warning for HTTP protocol on database, got none")
	}
}

func TestCheckConnectionTypeProtocolMismatch_Unspecified(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"1": {ID: "1", ComponentType: "database", Label: "Master"},
		"2": {ID: "2", ComponentType: "database", Label: "Slave"},
	}
	edges := []model.SystemEdge{
		{
			ID:             "e1",
			Source:         "1",
			Target:         "2",
			ConnectionType: "replication",
			Protocol:       "",
		},
	}

	ctx := makeCtx(nodes, edges)
	warnings := checkConnectionTypeProtocolMismatch(ctx)
	if len(warnings) > 0 {
		t.Errorf("Expected no warnings for replication with unspecified protocol, got %v", warnings)
	}
}

func TestCheckEntryPointSPOF_RP_SingleNoReplica(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx", Properties: map[string]interface{}{}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	warnings := checkEntryPointSPOF(ctx, "reverse_proxy", "reverse_proxy_spof", "🔀", "Reverse Proxy")
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning for single reverse proxy without replicas, got %d", len(warnings))
	}
}

func TestCheckEntryPointSPOF_RP_SingleWithReplica(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx", Properties: map[string]interface{}{
			"replicas": float64(3),
		}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	warnings := checkEntryPointSPOF(ctx, "reverse_proxy", "reverse_proxy_spof", "🔀", "Reverse Proxy")
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for reverse proxy with replicas > 1, got %d", len(warnings))
	}
}

func TestCheckReverseProxySSL_HTTPSWithoutTermination(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1": {ID: "c1", ComponentType: "client", Label: "Browser"},
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx", Properties: map[string]interface{}{
			"sslTermination": false,
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "rp1", ConnectionType: "sync", Protocol: "https"},
	}
	ctx := makeCtx(nodes, edges)
	warnings := checkReverseProxySSL(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning for HTTPS without SSL termination, got %d", len(warnings))
	}
}

func TestCheckReverseProxySSL_HTTPSWithTermination(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1": {ID: "c1", ComponentType: "client", Label: "Browser"},
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx", Properties: map[string]interface{}{
			"sslTermination": true,
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "rp1", ConnectionType: "sync", Protocol: "https"},
	}
	ctx := makeCtx(nodes, edges)
	warnings := checkReverseProxySSL(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for HTTPS with SSL termination, got %d", len(warnings))
	}
}

func TestCheckAsyncPeakShaving_ReverseProxy(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx"},
		"db1": {ID: "db1", ComponentType: "database", Label: "WriteDB", Properties: map[string]interface{}{
			"dbType":         "sql",
			"readWriteRatio": 0.2,
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "rp1", Target: "db1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	warnings := checkAsyncPeakShaving(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning for reverse proxy directly connecting to high-write DB, got %d", len(warnings))
	}
}

func TestCheckSPOF_ReverseProxySingleService(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx"},
		"s1": {ID: "s1", ComponentType: "service", Label: "API", Properties: map[string]interface{}{
			"replicas":  float64(1),
			"stateless": true,
		}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "rp1", Target: "s1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	warnings := checkSPOF(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 SPOF warning for reverse proxy with single service, got %d", len(warnings))
	}
}

func TestNodeHasRole(t *testing.T) {
	// Without roles — falls back to ComponentType
	node := model.SystemNode{ID: "1", ComponentType: "load_balancer", Label: "LB"}
	if !model.NodeHasRole(node, "load_balancer") {
		t.Error("Expected NodeHasRole to return true for ComponentType fallback")
	}
	if model.NodeHasRole(node, "reverse_proxy") {
		t.Error("Expected NodeHasRole to return false for unmatched role")
	}

	// With roles — uses Roles field
	merged := model.SystemNode{
		ID: "2", ComponentType: "reverse_proxy", Label: "Nginx+LB",
		Roles: []string{"reverse_proxy", "load_balancer"},
	}
	if !model.NodeHasRole(merged, "load_balancer") {
		t.Error("Expected NodeHasRole to return true for role in Roles slice")
	}
	if !model.NodeHasRole(merged, "reverse_proxy") {
		t.Error("Expected NodeHasRole to return true for role in Roles slice")
	}
	if model.NodeHasRole(merged, "database") {
		t.Error("Expected NodeHasRole to return false for unmatched role")
	}
}
