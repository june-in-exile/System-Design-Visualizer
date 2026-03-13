package handler

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkCacheEviction ---

func TestCheckCacheEviction_NoPolicy(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "c1", ComponentType: "cache", Label: "Redis", Properties: map[string]interface{}{
			"evictionPolicy": "none",
		}},
	}
	w := checkCacheEviction(nodes)
	if len(w) != 1 || w[0].Rule != "cache_eviction" {
		t.Errorf("expected 1 cache_eviction warning, got %d", len(w))
	}
}

func TestCheckCacheEviction_EmptyPolicy(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "c1", ComponentType: "cache", Label: "Redis", Properties: map[string]interface{}{
			"evictionPolicy": "",
		}},
	}
	w := checkCacheEviction(nodes)
	if len(w) != 1 {
		t.Errorf("expected 1 warning for empty policy, got %d", len(w))
	}
}

func TestCheckCacheEviction_WithLRU_NoWarning(t *testing.T) {
	nodes := []model.SystemNode{
		{ID: "c1", ComponentType: "cache", Label: "Redis", Properties: map[string]interface{}{
			"evictionPolicy": "lru",
		}},
	}
	w := checkCacheEviction(nodes)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkCacheConsistency ---

func TestCheckCacheConsistency_ServiceToCacheAndDB_NoTTL(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "API"},
		"c1":  {ID: "c1", ComponentType: "cache", Label: "Redis", Properties: map[string]interface{}{"ttlSeconds": float64(0)}},
		"db1": {ID: "db1", ComponentType: "database", Label: "PG"},
	}
	outgoing := map[string][]string{"s1": {"c1", "db1"}}
	w := checkCacheConsistency(nodes, outgoing)
	if len(w) != 1 || w[0].Rule != "cache_consistency" {
		t.Errorf("expected 1 cache_consistency warning, got %d", len(w))
	}
}

func TestCheckCacheConsistency_WithTTL_Suppressed(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1":  {ID: "s1", ComponentType: "service", Label: "API"},
		"c1":  {ID: "c1", ComponentType: "cache", Label: "Redis", Properties: map[string]interface{}{"ttlSeconds": float64(300)}},
		"db1": {ID: "db1", ComponentType: "database", Label: "PG"},
	}
	outgoing := map[string][]string{"s1": {"c1", "db1"}}
	w := checkCacheConsistency(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings with TTL set, got %d", len(w))
	}
}

func TestCheckCacheConsistency_OnlyCacheNoDB_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"s1": {ID: "s1", ComponentType: "service", Label: "API"},
		"c1": {ID: "c1", ComponentType: "cache", Label: "Redis", Properties: map[string]interface{}{}},
	}
	outgoing := map[string][]string{"s1": {"c1"}}
	w := checkCacheConsistency(nodes, outgoing)
	if len(w) != 0 {
		t.Errorf("expected 0 warnings without DB, got %d", len(w))
	}
}
