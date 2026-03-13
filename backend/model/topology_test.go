package model

import "testing"

func TestGetEffectiveRoles_WithRoles(t *testing.T) {
	node := SystemNode{ID: "1", ComponentType: "reverse_proxy", Roles: []string{"reverse_proxy", "load_balancer"}}
	roles := GetEffectiveRoles(node)
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestGetEffectiveRoles_FallbackToComponentType(t *testing.T) {
	node := SystemNode{ID: "1", ComponentType: "database"}
	roles := GetEffectiveRoles(node)
	if len(roles) != 1 || roles[0] != "database" {
		t.Errorf("expected [database], got %v", roles)
	}
}

func TestNodeHasRole_SimpleMatch(t *testing.T) {
	node := SystemNode{ID: "1", ComponentType: "cache"}
	if !NodeHasRole(node, "cache") {
		t.Error("expected true")
	}
	if NodeHasRole(node, "database") {
		t.Error("expected false")
	}
}

func TestNodeHasRole_MergedNode(t *testing.T) {
	node := SystemNode{ID: "1", ComponentType: "reverse_proxy", Roles: []string{"reverse_proxy", "load_balancer"}}
	if !NodeHasRole(node, "load_balancer") {
		t.Error("expected true for load_balancer role")
	}
	if !NodeHasRole(node, "reverse_proxy") {
		t.Error("expected true for reverse_proxy role")
	}
	if NodeHasRole(node, "service") {
		t.Error("expected false for service role")
	}
}
