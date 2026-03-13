package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkClientToDB ---

func TestCheckClientToDB_DirectConnection(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"db1": {ID: "db1", ComponentType: "database", Label: "MySQL"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "db1", ConnectionType: "sync"},
	}
	w := checkClientToDB(nodes, edges)
	if len(w) != 1 || w[0].Rule != "client_direct_db" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckClientToDB_ViaService_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "API"},
		"db1": {ID: "db1", ComponentType: "database", Label: "MySQL"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "s1", ConnectionType: "sync"},
		{ID: "e2", Source: "s1", Target: "db1", ConnectionType: "sync"},
	}
	w := checkClientToDB(nodes, edges)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkClientToCache ---

func TestCheckClientToCache_DirectConnection(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1": {ID: "c1", ComponentType: "client", Label: "Browser"},
		"r1": {ID: "r1", ComponentType: "cache", Label: "Redis"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "r1", ConnectionType: "sync"},
	}
	w := checkClientToCache(nodes, edges)
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
	outgoing := map[string][]string{"c1": {"lb1"}}
	w := checkMissingFirewall(nodes, outgoing)
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
	outgoing := map[string][]string{"c1": {"lb1"}}
	w := checkMissingFirewall(nodes, outgoing)
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
	outgoing := map[string][]string{"c1": {"fw1"}, "fw1": {"lb1"}}
	w := checkMissingFirewall(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkFirewallMonitorMode ---

func TestCheckFirewallMonitorMode_Monitor(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "fw1", ComponentType: "firewall", Label: "WAF", Properties: map[string]interface{}{
			"mode": "monitor", "layer": "l7",
		}},
	}
	w := checkFirewallMonitorMode(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckFirewallMonitorMode_Inline_NoWarning(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "fw1", ComponentType: "firewall", Label: "WAF", Properties: map[string]interface{}{
			"mode": "inline", "layer": "l7",
		}},
	}
	w := checkFirewallMonitorMode(nodes)
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
	w := checkFirewallL3Only(nodes)
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
	w := checkFirewallL3Only(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings without API gateway, got %d", len(w))
	}
}
