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
			Message:  "🛡️ Firewall/WAF Missing: Architecture contains Clients and entry nodes (LB/API Gateway), but lacks a Firewall.",
			Solution: "It is recommended to add a Firewall or WAF between Clients and Load Balancers/API Gateways for traffic filtering and malicious request blocking.",
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
				Message:  "🛡️ Firewall Not Correctly Connected: Firewall exists but is not connected to any entry nodes.",
				Solution: "Consider connecting the Firewall to the Load Balancer or API Gateway to enable traffic filtering and security protection.",
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
				Message:  fmt.Sprintf("🛡️ Firewall Monitor Mode Reminder: %q is currently in Monitor mode; malicious traffic will not be blocked.", node.Label),
				Solution: "Once testing is complete, it is recommended to switch the Firewall to Inline mode to enable actual blocking.",
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
				Message:  fmt.Sprintf("🛡️ Insufficient Protection Level: %q operates only at L3/L4 and cannot defend against application-layer attacks (e.g., SQL Injection, XSS).", node.Label),
				Solution: "Consider upgrading to an L7 WAF or adding an application-layer Firewall to protect API endpoints.",
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
							Message: fmt.Sprintf("🚫 Invalid Connection: %s (%s) → %s (%s).", source.Label, srcRole, target.Label, tgtRole),
							Solution: reason + " Please check the connection direction or add an appropriate intermediate layer.",
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
	return "This connection direction does not align with system design best practices."
}
