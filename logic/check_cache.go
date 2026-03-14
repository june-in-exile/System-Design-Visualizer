package logic

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkCacheEviction warns if a Cache node lacks an eviction policy.
func checkCacheEviction(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "cache") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		cacheProps, ok := props.(*model.CacheProperties)
		if !ok {
			continue
		}
		if cacheProps.EvictionPolicy == "" || cacheProps.EvictionPolicy == "none" {
			warnings = append(warnings, Warning{
				Rule:     "cache_eviction",
				Message:  fmt.Sprintf("🧊 Cache Eviction Policy Reminder: %q has no appropriate eviction algorithm configured.", node.Label),
				Solution: "Please set an Eviction Policy (e.g., LRU, LFU) to ensure old data is handled correctly when memory is exhausted.",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkCacheConsistency detects services connected to both cache and database.
// The warning is suppressed if at least one connected cache node has a non-zero TTLSeconds configured.
func checkCacheConsistency(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") {
			continue
		}
		targets := ctx.Outgoing[id]
		var involvedIDs []string
		hasCache := false
		hasDB := false
		anyCacheHasTTL := false

		for _, targetID := range targets {
			if target, ok := ctx.NodeByID[targetID]; ok {
				if model.NodeHasRole(target, "cache") {
					hasCache = true
					involvedIDs = append(involvedIDs, targetID)

					// Check if this cache node has a TTL configured
					props, err := model.ParseNodeProperties(target)
					if err == nil {
						if cacheProps, ok := props.(*model.CacheProperties); ok && cacheProps.TTLSeconds > 0 {
							anyCacheHasTTL = true
						}
					}
				}
				if model.NodeHasRole(target, "database") {
					hasDB = true
					involvedIDs = append(involvedIDs, targetID)
				}
			}
		}
		if hasCache && hasDB && !anyCacheHasTTL {
			warnings = append(warnings, Warning{
				Rule: "cache_consistency",
				Message: fmt.Sprintf("⚡ Cache Consistency Tradeoff: Service %q is connected to both Cache and Database.",
					node.Label),
				Solution: "Define a clear cache update strategy (e.g., Cache-aside) and set a reasonable TTL to prevent data staleness.",
				NodeIDs:  append([]string{id}, involvedIDs...),
			})
		}
	}
	return warnings
}

// checkCacheOnly warns if a service connects to Cache but has no path to Database/Storage.
// This handles the "no fallback" issue (C2).
func checkCacheOnly(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		// Skip fan-out services as they often only write to cache
		if labelContains(node.Label, "fanout") || labelContains(node.Label, "fan-out") {
			continue
		}

		targets := ctx.Outgoing[id]
		hasCache := false
		hasDB := false
		var cacheIDs []string

		for _, targetID := range targets {
			target, ok := ctx.NodeByID[targetID]
			if !ok {
				continue
			}
			if model.NodeHasRole(target, "cache") {
				hasCache = true
				cacheIDs = append(cacheIDs, targetID)
			}
			if model.NodeHasRole(target, "database") || model.NodeHasRole(target, "storage") {
				hasDB = true
			}
		}

		if hasCache && !hasDB {
			warnings = append(warnings, Warning{
				Rule:     "cache_no_fallback",
				Message:  fmt.Sprintf("❄️ Missing Cache Fallback: Service %q connects only to Cache without a Database/Storage fallback path.", node.Label),
				Solution: "In a Cache-aside pattern, when a Cache miss or eviction occurs, a database should serve as the ultimate data source (Fallback). Please establish a connection between this Service and a database.",
				NodeIDs:  append([]string{id}, cacheIDs...),
			})
		}
	}
	return warnings
}
