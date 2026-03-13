package logic

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkSPOF detects load balancers and reverse proxies with only one downstream service node,
// but suppresses the warning if that service node has Replicas > 1.
func checkSPOF(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "load_balancer") && !model.NodeHasRole(node, "reverse_proxy") {
			continue
		}
		targets := ctx.Outgoing[id]
		var serviceIDs []string
		isRedundant := false

		for _, targetID := range targets {
			target, ok := ctx.NodeByID[targetID]
			if !ok || !model.NodeHasRole(target, "service") {
				continue
			}
			serviceIDs = append(serviceIDs, targetID)

			// Check if this specific service node has multiple replicas
			props, err := model.ParseNodeProperties(target)
			if err == nil {
				if svcProps, ok := props.(*model.ServiceProperties); ok && svcProps.Replicas > 1 {
					isRedundant = true
				}
			}
		}

		// If there's only 1 distinct service node AND its replicas = 1, it's a SPOF
		if len(serviceIDs) == 1 && !isRedundant {
			warnings = append(warnings, Warning{
				Rule: "spof",
				Message: fmt.Sprintf("⚠️ 檢測到單點故障 (SPOF)：Load Balancer %q 後方僅連接 1 個 Service 節點複本。",
					node.Label),
				Solution: "增加 Service 節點數量 or 在屬性面板中提高 Replicas 複本數。",
				NodeIDs:  append([]string{id}, serviceIDs...),
			})
		} else if len(serviceIDs) > 1 {
			// Multiple distinct nodes also solve the SPOF
			isRedundant = true
		}
	}
	return warnings
}

// checkEntryPointSPOF warns if there is only one entry point node (LB/RP) in the system.
func checkEntryPointSPOF(ctx model.TopologyContext, role, rule, emoji, label string) []Warning {
	var matched []model.SystemNode
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, role) {
			matched = append(matched, node)
		}
	}
	if len(matched) != 1 {
		return nil
	}

	node := matched[0]
	props, err := model.ParseNodeProperties(node)
	if err == nil {
		// Check Replicas > 1
		switch p := props.(type) {
		case *model.LoadBalancerProperties:
			if p.Replicas > 1 {
				return nil
			}
		case *model.ReverseProxyProperties:
			if p.Replicas > 1 {
				return nil
			}
		}
	}

	return []Warning{{
		Rule:     rule,
		Message:  fmt.Sprintf("%s 入口單點故障：整體架構中僅存在 1 個 %s。", emoji, label),
		Solution: fmt.Sprintf("生產環境建議部署多個 %s，或在屬性面板中將 Replicas 複本數設為 2 以上。", label),
		NodeIDs:  []string{node.ID},
	}}
}

// checkNoAutoScalingSingle warns if a service has only one replica and auto-scaling is disabled.
func checkNoAutoScalingSingle(ctx model.TopologyContext) []Warning {
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
		if !svcProps.AutoScaling && svcProps.Replicas == 1 {
			warnings = append(warnings, Warning{
				Rule:     "no_autoscaling_single",
				Message:  fmt.Sprintf("⚠️ 單一複本缺乏可用性：服務 %q 僅有 1 個 Replica 且未啟用自動擴縮。", node.Label),
				Solution: "建議啟用 Auto Scaling 或將 Replicas 增加至 2 以上以確保可用性，防止流量突增時成為瓶頸。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkNoHealthCheckBehindLB warns if a service without health check is connected to a Load Balancer.
func checkNoHealthCheckBehindLB(ctx model.TopologyContext) []Warning {
	var warnings []Warning

	// Build a map of nodes that are Load Balancers
	lbNodes := make(map[string]bool)
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "load_balancer") {
			lbNodes[node.ID] = true
		}
	}

	// For each edge from LB to Service, check if Service has healthCheck
	for _, edge := range ctx.Edges {
		if !lbNodes[edge.Source] {
			continue
		}
		target, ok := ctx.NodeByID[edge.Target]
		if !ok || !model.NodeHasRole(target, "service") {
			continue
		}

		props, err := model.ParseNodeProperties(target)
		if err != nil {
			continue
		}
		svcProps, ok := props.(*model.ServiceProperties)
		if !ok {
			continue
		}
		if !svcProps.HealthCheck {
			warnings = append(warnings, Warning{
				Rule:     "no_healthcheck_behind_lb",
				Message:  fmt.Sprintf("🏥 健康檢查缺失：服務 %q 位於 Load Balancer 後方但未啟用健康檢查。", target.Label),
				Solution: "建議在服務中暴露 /health 端點並在 Load Balancer 中設定 health check 間隔，以便自動偵測並移除不健康的實例。",
				NodeIDs:  []string{target.ID},
			})
		}
	}
	return warnings
}

// checkServerlessReplicas warns if a serverless service has manually set replicas.
func checkServerlessReplicas(ctx model.TopologyContext) []Warning {
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
		if svcProps.ComputeType == "serverless" && svcProps.Replicas > 1 {
			warnings = append(warnings, Warning{
				Rule:     "serverless_replicas",
				Message:  fmt.Sprintf("⚠️ Serverless 不需手動設定 Replicas：服務 %q 為 Serverless 運算且設定了多個複本。", node.Label),
				Solution: "Serverless 模式下 Replicas 由雲端平台自動管理，手動設定 replicas 無意義。建議移除 replicas 設定，改由平台自動擴縮。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}
