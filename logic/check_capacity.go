package logic

import (
	"fmt"

	"github.com/architectmind/backend/model"
)

func checkCapacity(ctx model.TopologyContext) []Warning {
	if ctx.Params == nil {
		return nil
	}
	var warnings []Warning
	warnings = append(warnings, checkHighQPSNoCache(ctx)...)
	warnings = append(warnings, checkHighQPSSingleLB(ctx)...)
	warnings = append(warnings, checkHighQPSNoAutoScaling(ctx)...)
	warnings = append(warnings, checkHighDAUNoCDN(ctx)...)
	warnings = append(warnings, checkStorageGrowthNoPartitioning(ctx)...)
	warnings = append(warnings, checkHighAvailabilityInsufficientReplicas(ctx)...)
	warnings = append(warnings, checkLatencyLongSyncChain(ctx)...)
	warnings = append(warnings, checkReadHeavyNoReadReplica(ctx)...)
	return warnings
}

func checkHighQPSNoCache(ctx model.TopologyContext) []Warning {
	if ctx.Params.PeakQPS <= 5000 {
		return nil
	}
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "cache") {
			return nil
		}
	}
	return []Warning{{
		Rule:     "high_qps_no_cache",
		Message:  fmt.Sprintf("⚡ 高流量缺少快取層：Peak QPS 為 %d，但架構中沒有 Cache 節點。", ctx.Params.PeakQPS),
		Solution: "建議加入 Cache（如 Redis）以降低資料庫壓力，一般可減少 60-80%% 的 DB 查詢。",
		NodeIDs:  []string{},
	}}
}

func checkHighQPSSingleLB(ctx model.TopologyContext) []Warning {
	if ctx.Params.PeakQPS <= 10000 {
		return nil
	}
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "load_balancer") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		lbProps, ok := props.(*model.LoadBalancerProperties)
		if !ok {
			continue
		}
		if lbProps.Replicas <= 1 {
			warnings = append(warnings, Warning{
				Rule:     "high_qps_single_lb",
				Message:  fmt.Sprintf("⚖️ 高流量單一 LB：Peak QPS 為 %d，但 %q 僅有 1 個複本。", ctx.Params.PeakQPS, node.Label),
				Solution: "Peak QPS 超過 10,000 時建議 LB 至少 2 個複本，避免入口成為瓶頸。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

func checkHighQPSNoAutoScaling(ctx model.TopologyContext) []Warning {
	if ctx.Params.PeakQPS <= 5000 {
		return nil
	}
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		svcProps, ok := props.(*model.ServiceProperties)
		if !ok {
			continue
		}
		if !svcProps.AutoScaling {
			warnings = append(warnings, Warning{
				Rule:     "high_qps_no_autoscaling",
				Message:  fmt.Sprintf("📈 高流量未自動擴縮：Peak QPS 為 %d，服務 %q 未啟用 Auto Scaling。", ctx.Params.PeakQPS, node.Label),
				Solution: "高流量場景下建議啟用 Auto Scaling，根據 CPU/記憶體/請求數自動調整複本數。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

func checkHighDAUNoCDN(ctx model.TopologyContext) []Warning {
	if ctx.Params.DAU <= 100000 {
		return nil
	}
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "cdn") {
			return nil
		}
	}
	return []Warning{{
		Rule:     "high_dau_no_cdn",
		Message:  fmt.Sprintf("🌐 高用戶量缺少 CDN：DAU 為 %d，但架構中沒有 CDN 節點。", ctx.Params.DAU),
		Solution: "DAU 超過 10 萬時建議加入 CDN，將靜態資源分發到邊緣節點以降低延遲和後端負載。",
		NodeIDs:  []string{},
	}}
}

func checkStorageGrowthNoPartitioning(ctx model.TopologyContext) []Warning {
	if ctx.Params.DailyGrowthGB <= 10 {
		return nil
	}
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "database") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		dbProps, ok := props.(*model.DatabaseProperties)
		if !ok {
			continue
		}
		if dbProps.ScalingStrategy != "horizontal" {
			warnings = append(warnings, Warning{
				Rule:     "storage_growth_no_partitioning",
				Message:  fmt.Sprintf("💾 資料增長但未水平擴展：每日增長 %.0f GB，資料庫 %q 未設定水平擴展策略。", ctx.Params.DailyGrowthGB, node.Label),
				Solution: "每日資料增長超過 10 GB 時，建議考慮分庫分表 (Sharding) 或使用支援水平擴展的資料庫。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

func checkHighAvailabilityInsufficientReplicas(ctx model.TopologyContext) []Warning {
	avail := ctx.Params.Availability
	if avail != "99.99%" && avail != "99.999%" {
		return nil
	}
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		svcProps, ok := props.(*model.ServiceProperties)
		if !ok {
			continue
		}
		if svcProps.Replicas < 3 {
			warnings = append(warnings, Warning{
				Rule:     "high_availability_insufficient_replicas",
				Message:  fmt.Sprintf("🛡️ 高可用性複本不足：目標 %s，但服務 %q 僅有 %d 個複本。", avail, node.Label, svcProps.Replicas),
				Solution: fmt.Sprintf("可用性目標 %s 至少需要 3 個 Service 複本，搭配健康檢查與自動故障轉移。", avail),
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

func checkLatencyLongSyncChain(ctx model.TopologyContext) []Warning {
	target := ctx.Params.LatencyTarget
	if target != "p99 < 100ms" && target != "p95 < 50ms" {
		return nil
	}

	maxDepth := 0
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "client") {
			continue
		}
		depth := bfsSyncDepth(ctx, node.ID)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	if maxDepth > 3 {
		return []Warning{{
			Rule:     "latency_long_sync_chain",
			Message:  fmt.Sprintf("⏱️ 嚴格延遲目標 + 長同步鏈：目標 %s，但同步呼叫鏈深度達 %d 層。", target, maxDepth),
			Solution: "嚴格延遲要求下，建議將部分同步呼叫改為非同步，或使用 Cache 縮短關鍵路徑。",
			NodeIDs:  []string{},
		}}
	}
	return nil
}

func bfsSyncDepth(ctx model.TopologyContext, startID string) int {
	syncAdj := make(map[string][]string)
	for _, edge := range ctx.Edges {
		if edge.ConnectionType == "sync" || edge.ConnectionType == "unspecified" {
			syncAdj[edge.Source] = append(syncAdj[edge.Source], edge.Target)
		}
	}

	visited := map[string]bool{startID: true}
	queue := []struct {
		id    string
		depth int
	}{{startID, 0}}
	maxDepth := 0

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if curr.depth > maxDepth {
			maxDepth = curr.depth
		}
		for _, next := range syncAdj[curr.id] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, struct {
					id    string
					depth int
				}{next, curr.depth + 1})
			}
		}
	}
	return maxDepth
}

func checkReadHeavyNoReadReplica(ctx model.TopologyContext) []Warning {
	if ctx.Params.ReadWriteRatio <= 0.8 {
		return nil
	}

	var dbNodes []model.SystemNode
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "database") {
			dbNodes = append(dbNodes, node)
		}
	}

	if len(dbNodes) <= 1 {
		ids := make([]string, len(dbNodes))
		for i, n := range dbNodes {
			ids[i] = n.ID
		}
		return []Warning{{
			Rule:     "read_heavy_no_read_replica",
			Message:  fmt.Sprintf("📖 讀多寫少但無讀寫分離：讀佔比 %.0f%%，但僅有 %d 個資料庫節點。", ctx.Params.ReadWriteRatio*100, len(dbNodes)),
			Solution: "讀寫比超過 80%% 讀時，建議設定讀寫分離（Read Replica），將讀流量分散到從庫以提升效能。",
			NodeIDs:  ids,
		}}
	}

	return nil
}
