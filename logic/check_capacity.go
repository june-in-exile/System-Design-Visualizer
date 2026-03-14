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
		Message:  fmt.Sprintf("⚡ High Traffic Missing Cache: Peak QPS is %d, but no Cache nodes exist in the architecture.", ctx.Params.PeakQPS),
		Solution: "Consider adding a Cache (e.g., Redis) to reduce database pressure, typically reducing DB queries by 60-80%.",
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
				Message:  fmt.Sprintf("⚖️ High Traffic Single LB: Peak QPS is %d, but %q has only 1 replica.", ctx.Params.PeakQPS, node.Label),
				Solution: "When Peak QPS exceeds 10,000, at least 2 LB replicas are recommended to avoid entry bottlenecks.",
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
				Message:  fmt.Sprintf("📈 High Traffic No Auto Scaling: Peak QPS is %d, and service %q has Auto Scaling disabled.", ctx.Params.PeakQPS, node.Label),
				Solution: "In high traffic scenarios, it is recommended to enable Auto Scaling to automatically adjust replica counts based on CPU/Memory/Requests.",
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
		Message:  fmt.Sprintf("🌐 High User Volume Missing CDN: DAU is %d, but no CDN nodes exist in the architecture.", ctx.Params.DAU),
		Solution: "When DAU exceeds 100,000, it is recommended to add a CDN to distribute static resources to edge nodes, reducing latency and backend load.",
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
				Message:  fmt.Sprintf("💾 Data Growth Missing Horizontal Scaling: Daily growth is %.0f GB, and database %q has no horizontal scaling strategy set.", ctx.Params.DailyGrowthGB, node.Label),
				Solution: "When daily data growth exceeds 10 GB, consider Database Sharding or using a database that supports horizontal scaling.",
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
				Message:  fmt.Sprintf("🛡️ Insufficient Replicas for High Availability: Target is %s, but service %q has only %d replica(s).", avail, node.Label, svcProps.Replicas),
				Solution: fmt.Sprintf("An availability target of %s requires at least 3 service replicas, combined with health checks and automatic failover.", avail),
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
			Message:  fmt.Sprintf("⏱️ Strict Latency Target + Long Sync Chain: Target is %s, but synchronous call chain depth reaches %d layers.", target, maxDepth),
			Solution: "Under strict latency requirements, consider changing some synchronous calls to asynchronous or using Cache to shorten critical paths.",
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
			Message:  fmt.Sprintf("📖 Read-Heavy Missing Read Replicas: Read ratio is %.0f%%, but there is only %d database node(s).", ctx.Params.ReadWriteRatio*100, len(dbNodes)),
			Solution: "When the read/write ratio exceeds 80% reads, it is recommended to set up Read Replicas to distribute read traffic and improve performance.",
			NodeIDs:  ids,
		}}
	}

	return nil
}
