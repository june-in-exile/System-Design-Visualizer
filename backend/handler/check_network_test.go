package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkCDNUsage ---

func TestCheckCDNUsage_ClientNoCDN(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1": {ID: "c1", ComponentType: "client", Label: "Browser"},
		"s1": {ID: "s1", ComponentType: "service", Label: "API"},
	}
	outgoing := map[string][]string{"c1": {"s1"}}
	w := checkCDNUsage(nodes, outgoing)
	if len(w) != 1 || w[0].Rule != "cdn_usage" {
		t.Errorf("expected 1 cdn_usage warning, got %d", len(w))
	}
}

func TestCheckCDNUsage_ClientAndCDN_Connected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":   {ID: "c1", ComponentType: "client", Label: "Browser"},
		"cdn1": {ID: "cdn1", ComponentType: "cdn", Label: "CloudFront"},
	}
	outgoing := map[string][]string{"c1": {"cdn1"}}
	w := checkCDNUsage(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

func TestCheckCDNUsage_ClientAndCDN_NotConnected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":   {ID: "c1", ComponentType: "client", Label: "Browser"},
		"cdn1": {ID: "cdn1", ComponentType: "cdn", Label: "CloudFront"},
		"s1":   {ID: "s1", ComponentType: "service", Label: "API"},
	}
	outgoing := map[string][]string{"c1": {"s1"}}
	w := checkCDNUsage(nodes, outgoing)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for disconnected CDN, got %d", len(w))
	}
}

func TestCheckCDNUsage_NoClient_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API"},
	}
	w := checkCDNUsage(nodes, map[string][]string{})
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkReverseProxySSL (補充現有測試) ---

func TestCheckReverseProxySSL_HTTPNoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":  {ID: "c1", ComponentType: "client", Label: "Browser"},
		"rp1": {ID: "rp1", ComponentType: "reverse_proxy", Label: "Nginx", Properties: map[string]interface{}{"sslTermination": false}},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "rp1", ConnectionType: "sync", Protocol: "http"},
	}
	w := checkReverseProxySSL(nodes, edges)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for HTTP (not HTTPS), got %d", len(w))
	}
}
