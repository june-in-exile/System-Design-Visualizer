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
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "s1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkCDNUsage(ctx)
	if len(w) != 1 || w[0].Rule != "cdn_usage" {
		t.Errorf("expected 1 cdn_usage warning, got %d", len(w))
	}
}

func TestCheckCDNUsage_ClientAndCDN_Connected(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"c1":   {ID: "c1", ComponentType: "client", Label: "Browser"},
		"cdn1": {ID: "cdn1", ComponentType: "cdn", Label: "CloudFront"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "cdn1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkCDNUsage(ctx)
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
	edges := []model.SystemEdge{
		{ID: "e1", Source: "c1", Target: "s1", ConnectionType: "sync"},
	}
	ctx := makeCtx(nodes, edges)
	w := checkCDNUsage(ctx)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for disconnected CDN, got %d", len(w))
	}
}

func TestCheckCDNUsage_NoClient_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API"},
	}
	ctx := makeCtx(nodes, []model.SystemEdge{})
	w := checkCDNUsage(ctx)
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
	ctx := makeCtx(nodes, edges)
	w := checkReverseProxySSL(ctx)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for HTTP (not HTTPS), got %d", len(w))
	}
}
