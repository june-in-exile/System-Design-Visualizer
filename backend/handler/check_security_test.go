package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkClientDirectAccess ---

func TestCheckClientDirectAccess_DB_DirectConnection(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"db1": {ID: "db1", ComponentType: "database", Label: "MySQL"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "db1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkClientDirectAccess(ctx, "database", "client_direct_db", "🚫 安全風險：禁止從 %q 直接連線至 %q。", "Solution")
	if len(w) != 1 || w[0].Rule != "client_direct_db" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckClientDirectAccess_Cache_DirectConnection(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1": {ID: "c1", ComponentType: "client", Label: "Browser"},
		"r1": {ID: "r1", ComponentType: "cache", Label: "Redis"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "r1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkClientDirectAccess(ctx, "cache", "client_direct_cache", "🧊 暴露風險：不建議從 %q 直接連線至 %q。", "Solution")
	if len(w) != 1 || w[0].Rule != "client_direct_cache" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

// --- checkMissingFirewall ---

func TestCheckMissingFirewall_NoFirewall(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "lb1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkMissingFirewall(ctx)
	if len(w) != 1 || w[0].Rule != "missing_firewall" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckMissingFirewall_FirewallNotConnected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"fw1": {ID: "fw1", ComponentType: "firewall", Label: "WAF"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "lb1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkMissingFirewall(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for disconnected firewall, got %d", len(w))
	}
}

func TestCheckMissingFirewall_FirewallConnected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"lb1": {ID: "lb1", ComponentType: "load_balancer", Label: "LB"},
		"fw1": {ID: "fw1", ComponentType: "firewall", Label: "WAF"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "fw1", ConnectionType: "sync"},
		{ID: "e2", Source: "fw1", Target: "lb1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkMissingFirewall(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkFirewallMonitorMode ---

func TestCheckFirewallMonitorMode_Monitor(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"fw1": {ID: "fw1", ComponentType: "firewall", Label: "WAF", Properties: map[string]interface{}{
			"mode": "monitor", "layer": "l7",
		}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkFirewallMonitorMode(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckFirewallMonitorMode_Inline_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"fw1": {ID: "fw1", ComponentType: "firewall", Label: "WAF", Properties: map[string]interface{}{
			"mode": "inline", "layer": "l7",
		}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkFirewallMonitorMode(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkFirewallL3Only ---

func TestCheckFirewallL3Only_WithAPIGateway(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"fw1": {ID: "fw1", ComponentType: "firewall", Label: "FW", Properties: map[string]interface{}{
			"layer": "l3",
		}},
		"ag1": {ID: "ag1", ComponentType: "api_gateway", Label: "Gateway"},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkFirewallL3Only(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckFirewallL3Only_NoAPIGateway_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"fw1": {ID: "fw1", ComponentType: "firewall", Label: "FW", Properties: map[string]interface{}{
			"layer": "l3",
		}},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkFirewallL3Only(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings without API gateway, got %d", len(w))
	}
}
