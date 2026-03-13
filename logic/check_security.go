package logic

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkMissingFirewall warns if there's a Client and LB/API Gateway but no Firewall,
// or if a Firewall exists but is not connected to the entry points.
func checkMissingFirewall(ctx model.TopologyContext) []Warning {
	hasClient := false
	hasEntryPoint := false
	hasFirewall := false
	var clientIDs []string
	var entryPointIDs []string
	var firewallIDs []string

	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "client") {
			hasClient = true
			clientIDs = append(clientIDs, node.ID)
		}
		if model.NodeHasRole(node, "load_balancer") || model.NodeHasRole(node, "api_gateway") {
			hasEntryPoint = true
			entryPointIDs = append(entryPointIDs, node.ID)
		}
		if model.NodeHasRole(node, "firewall") {
			hasFirewall = true
			firewallIDs = append(firewallIDs, node.ID)
		}
	}

	if hasClient && hasEntryPoint && !hasFirewall {
		return []Warning{{
			Rule:     "missing_firewall",
			Message:  "🛡️ 缺少 Firewall/WAF：架構中有 Client 與入口節點 (LB/API Gateway)，但缺少 Firewall。",
			Solution: "建議在 Client 與 Load Balancer/API Gateway 之間加入 Firewall 或 WAF，進行流量過濾與惡意請求攔截。",
			NodeIDs:  clientIDs,
		}}
	}

	if hasClient && hasEntryPoint && hasFirewall {
		connected := false
		for _, firewallID := range firewallIDs {
			targets := ctx.Outgoing[firewallID]
			for _, targetID := range targets {
				if model.NodeHasRole(ctx.NodeByID[targetID], "load_balancer") || model.NodeHasRole(ctx.NodeByID[targetID], "api_gateway") {
					connected = true
					break
				}
			}
			if connected {
				break
			}
		}

		if !connected {
			return []Warning{{
				Rule:     "missing_firewall",
				Message:  "🛡️ Firewall 未正確連線：架構中有 Firewall 但未連線至入口節點。",
				Solution: "建議將 Firewall 連接至 Load Balancer 或 API Gateway，以發揮流量過濾與安全防護的效果。",
				NodeIDs:  append(firewallIDs, entryPointIDs...),
			}}
		}
	}

	return nil
}

// checkFirewallMonitorMode warns if a Firewall is in monitor mode (not blocking).
func checkFirewallMonitorMode(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "firewall") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		fwProps, ok := props.(*model.FirewallProperties)
		if !ok {
			continue
		}
		if fwProps.Mode == "monitor" {
			warnings = append(warnings, Warning{
				Rule:     "firewall_monitor_mode",
				Message:  fmt.Sprintf("🛡️ Firewall 監控模式提醒：%q 目前為監控模式 (Monitor)，惡意流量不會被攔截。", node.Label),
				Solution: "若已完成測試，建議將 Firewall 切換為 Inline 模式以啟用實際攔截。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkFirewallL3Only warns if a Firewall operates at L3/L4 only while an API Gateway exists.
func checkFirewallL3Only(ctx model.TopologyContext) []Warning {
	hasAPIGateway := false
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "api_gateway") {
			hasAPIGateway = true
			break
		}
	}
	if !hasAPIGateway {
		return nil
	}

	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "firewall") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		fwProps, ok := props.(*model.FirewallProperties)
		if !ok {
			continue
		}
		if fwProps.Layer == "l3" {
			warnings = append(warnings, Warning{
				Rule:     "firewall_l3_only",
				Message:  fmt.Sprintf("🛡️ 防護層級不足：%q 僅在 L3/L4 層運作，無法防禦應用層攻擊（如 SQL Injection、XSS）。", node.Label),
				Solution: "建議升級為 L7 WAF 或額外加入應用層 Firewall 以保護 API 端點。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkInvalidConnection detects architecturally invalid connections
// using the ForbiddenConnections lookup table.
func checkInvalidConnection(ctx model.TopologyContext) []Warning {
	var warnings []Warning

	for _, edge := range ctx.Edges {
		source, okS := ctx.NodeByID[edge.Source]
		target, okT := ctx.NodeByID[edge.Target]
		if !okS || !okT {
			continue
		}

		// Check all source roles against all target roles
		for _, srcRole := range model.GetEffectiveRoles(source) {
			forbidden, hasForbidden := model.ForbiddenConnections[srcRole]
			if !hasForbidden {
				continue
			}
			for _, tgtRole := range model.GetEffectiveRoles(target) {
				for _, f := range forbidden {
					if f == tgtRole {
						reason := getConnectionReason(srcRole, tgtRole)
						warnings = append(warnings, Warning{
							Rule:    "invalid_connection",
							Message: fmt.Sprintf("🚫 不合理的連線：%s (%s) → %s (%s)。", source.Label, srcRole, target.Label, tgtRole),
							Solution: reason + " 請檢查連線方向或在兩者之間加入適當的中間層。",
							NodeIDs: []string{source.ID, target.ID},
						})
					}
				}
			}
		}
	}
	return warnings
}

// getConnectionReason returns a human-readable reason for a forbidden connection.
func getConnectionReason(srcRole, tgtRole string) string {
	if reasons, ok := model.ForbiddenConnectionReasons[srcRole]; ok {
		if reason, ok := reasons[tgtRole]; ok {
			return reason
		}
		if reason, ok := reasons["*"]; ok {
			return reason
		}
	}
	return "此連線方向不符合系統設計最佳實踐。"
}
