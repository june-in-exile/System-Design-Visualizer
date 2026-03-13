package handler

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
				Message:  fmt.Sprintf("🧊 快取失效策略提醒：%q 未配置適當的失效演算法。", node.Label),
				Solution: "請設定 Eviction Policy (如 LRU, LFU)，以確保在記憶體用罄時能正確處理舊數據。",
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
				Message: fmt.Sprintf("⚡ 快取一致性權衡：Service %q 同時連接 Cache 與 Database。",
					node.Label),
				Solution: "明確快取更新策略（如 Cache-aside），並設定合理的 TTL 以防數據過期。",
				NodeIDs:  append([]string{id}, involvedIDs...),
			})
		}
	}
	return warnings
}
