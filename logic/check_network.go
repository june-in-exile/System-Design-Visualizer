package logic

import (
	"fmt"
	"strings"

	"github.com/architectmind/backend/model"
)

// checkProtocolMismatch warns when an edge uses a protocol that doesn't match
// the target component type (e.g. HTTP to a Database).
func checkProtocolMismatch(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, edge := range ctx.Edges {
		if edge.Protocol == "" || edge.Protocol == "unspecified" {
			continue
		}
		target, ok := ctx.NodeByID[edge.Target]
		if !ok {
			continue
		}

		// For merged nodes, build the union of expected protocols across all roles
		allAllowed := make(map[string]bool)
		hasAnyRule := false
		for _, role := range model.GetEffectiveRoles(target) {
			if allowed, ok := model.ExpectedProtocols[role]; ok {
				hasAnyRule = true
				for p := range allowed {
					allAllowed[p] = true
				}
			}
		}
		if !hasAnyRule {
			continue
		}
		if allAllowed[edge.Protocol] {
			continue
		}

		protoName := model.ProtocolDisplayName[edge.Protocol]
		if protoName == "" {
			protoName = edge.Protocol
		}

		var suggestion []string
		for p := range allAllowed {
			name := model.ProtocolDisplayName[p]
			if name == "" {
				name = p
			}
			suggestion = append(suggestion, name)
		}

		source := ctx.NodeByID[edge.Source]
		warnings = append(warnings, Warning{
			Rule: "protocol_mismatch",
			Message: fmt.Sprintf("🔌 Protocol Mismatch: %s → %s uses %s, but %s nodes typically do not use this protocol.",
				source.Label, target.Label, protoName, target.Label),
			Solution: fmt.Sprintf("Consider changing the protocol to %s.", strings.Join(suggestion, " or ")),
			NodeIDs:  []string{edge.Source, edge.Target},
		})
	}
	return warnings
}

// checkConnectionTypeProtocolMismatch warns when a connection type and protocol are logically inconsistent.
func checkConnectionTypeProtocolMismatch(ctx model.TopologyContext) []Warning {
	var warnings []Warning

	connectionTypeNames := map[string]string{
		"sync":        "Synchronous",
		"async":       "Asynchronous",
		"replication": "Replication",
		"cdn_origin":  "CDN Origin",
	}

	for _, edge := range ctx.Edges {
		proto := edge.Protocol
		if proto == "" {
			proto = "unspecified"
		}

		allowed, hasRule := model.ValidConnectionProtocolPairs[edge.ConnectionType]
		if !hasRule {
			continue
		}

		if !allowed[proto] {
			source := ctx.NodeByID[edge.Source]
			target := ctx.NodeByID[edge.Target]
			connName := connectionTypeNames[edge.ConnectionType]
			protoName := model.ProtocolDisplayName[proto]
			if protoName == "" {
				protoName = proto
			}

			warnings = append(warnings, Warning{
				Rule: "protocol_connection_mismatch",
				Message: fmt.Sprintf("⚠️ Property Mismatch: %s → %s uses a %s connection, which should not be paired with the %s protocol.",
					source.Label, target.Label, connName, protoName),
				Solution: fmt.Sprintf("Change the connection type to Async or switch to a protocol compatible with %s (e.g., HTTP/gRPC for Sync, AMQP/MQTT for Async).", connName),
				NodeIDs:  []string{source.ID, target.ID},
			})
		}
	}
	return warnings
}

// checkCDNUsage warns if a Client exists but no CDN is found, or if both exist but are not connected.
func checkCDNUsage(ctx model.TopologyContext) []Warning {
	var clientIDs []string
	var cdnIDs []string
	for id, node := range ctx.NodeByID {
		if model.NodeHasRole(node, "client") {
			clientIDs = append(clientIDs, id)
		}
		if model.NodeHasRole(node, "cdn") {
			cdnIDs = append(cdnIDs, id)
		}
	}

	if len(clientIDs) > 0 && len(cdnIDs) == 0 {
		return []Warning{{
			Rule:     "cdn_usage",
			Message:  "🌐 CDN Acceleration Suggestion: Clients exist in the topology, but no CDN node is present.",
			Solution: "Consider adding a CDN (Content Delivery Network) between Clients and the backend entry points to accelerate static resource distribution and reduce latency.",
			NodeIDs:  clientIDs,
		}}
	}

	if len(clientIDs) > 0 && len(cdnIDs) > 0 {
		// Check if any client connects to a cdn
		connected := false
		for _, clientID := range clientIDs {
			targets := ctx.Outgoing[clientID]
			for _, targetID := range targets {
				if model.NodeHasRole(ctx.NodeByID[targetID], "cdn") {
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
				Rule:     "cdn_usage",
				Message:  "🌐 CDN Acceleration Suggestion: Both Clients and CDN exist, but they are not connected.",
				Solution: "Connect Clients to the CDN node to leverage its content caching and global acceleration benefits.",
				NodeIDs:  append(clientIDs, cdnIDs...),
			}}
		}
	}

	return nil
}

// checkReverseProxySSL warns if a reverse proxy receives HTTPS traffic but has SSL termination disabled.
func checkReverseProxySSL(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "reverse_proxy") {
			continue
		}

		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		rpProps, ok := props.(*model.ReverseProxyProperties)
		if !ok {
			continue
		}
		if rpProps.SSLTermination {
			continue
		}

		// Check if any incoming edge uses HTTPS
		for _, edge := range ctx.Edges {
			if edge.Target == node.ID && edge.Protocol == "https" {
				warnings = append(warnings, Warning{
					Rule:     "reverse_proxy_ssl",
					Message:  fmt.Sprintf("🔒 SSL Termination Suggestion: Reverse Proxy %q receives HTTPS traffic but has SSL termination disabled.", node.Label),
					Solution: "Enable SSL Termination on the Reverse Proxy to centrally manage TLS certificates and reduce the decryption load on backend services.",
					NodeIDs:  []string{node.ID},
				})
				break
			}
		}
	}
	return warnings
}

// checkCDNOrigin warns if a CDN node lacks a cdn_origin outgoing edge.
// This handles the "isolated CDN" issue (C3).
func checkCDNOrigin(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "cdn") {
			continue
		}

		hasOrigin := false
		for _, edge := range ctx.Edges {
			if edge.Source == id && edge.ConnectionType == "cdn_origin" {
				hasOrigin = true
				break
			}
		}

		if !hasOrigin {
			warnings = append(warnings, Warning{
				Rule:     "cdn_isolated",
				Message:  fmt.Sprintf("🌐 CDN Isolation Reminder: %q has no Origin server connection configured.", node.Label),
				Solution: "A CDN needs to pull content from an Origin (e.g., Storage or API Gateway). Please establish a connection and set the Connection Type to 'CDN Origin'.",
				NodeIDs:  []string{id},
			})
		}
	}
	return warnings
}

// checkSyncChain detects synchronous call chains of length >= 3 between services.
// This handles the "long sync chain" issue (C5/G1).
func checkSyncChain(ctx model.TopologyContext) []Warning {
	var warnings []Warning

	// Helper to find max sync depth from a service
	var findMaxDepth func(currentID string, visited map[string]bool) (int, []string)
	findMaxDepth = func(currentID string, visited map[string]bool) (int, []string) {
		visited[currentID] = true
		maxD := 0
		var longestPath []string

		for _, edge := range ctx.Edges {
			if edge.Source == currentID && edge.ConnectionType == "sync" {
				target, ok := ctx.NodeByID[edge.Target]
				if ok && model.NodeHasRole(target, "service") && !visited[edge.Target] {
					d, p := findMaxDepth(edge.Target, copyMap(visited))
					if d > maxD {
						maxD = d
						longestPath = p
					}
				}
			}
		}

		return maxD + 1, append([]string{currentID}, longestPath...)
	}

	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		// Only check from entry points or nodes with no incoming sync edges from services
		isEntry := true
		for _, edge := range ctx.Edges {
			if edge.Target == id && edge.ConnectionType == "sync" {
				source, ok := ctx.NodeByID[edge.Source]
				if ok && model.NodeHasRole(source, "service") {
					isEntry = false
					break
				}
			}
		}

		if isEntry {
			depth, path := findMaxDepth(id, make(map[string]bool))
			if depth >= 3 {
				warnings = append(warnings, Warning{
					Rule:     "long_sync_chain",
					Message:  fmt.Sprintf("🔗 Long Synchronous Chain: Detected a synchronous call chain of length %d (%s).", depth, joinLabelsByID(path, ctx)),
					Solution: "Overly long synchronous chains lead to latency accumulation and decreased availability. Consider changing some calls to parallel (Fan-out) or asynchronous processing (Async/MQ).",
					NodeIDs:  path,
				})
			}
		}
	}
	return warnings
}

// checkInternalHTTP warns if Service-to-Service communication uses plain HTTP.
// This handles the "internal http" issue (T6).
func checkInternalHTTP(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, edge := range ctx.Edges {
		if edge.Protocol != "http" {
			continue
		}
		source, okS := ctx.NodeByID[edge.Source]
		target, okT := ctx.NodeByID[edge.Target]
		if okS && okT && model.NodeHasRole(source, "service") && model.NodeHasRole(target, "service") {
			warnings = append(warnings, Warning{
				Rule:     "internal_http",
				Message:  fmt.Sprintf("🔌 Internal Communication Protocol Suggestion: %s → %s uses HTTP.", source.Label, target.Label),
				Solution: "For internal microservices communication, gRPC is recommended for better serialization efficiency and multiplexing benefits.",
				NodeIDs:  []string{edge.Source, edge.Target},
			})
		}
	}
	return warnings
}

func copyMap(m map[string]bool) map[string]bool {
	nm := make(map[string]bool)
	for k, v := range m {
		nm[k] = v
	}
	return nm
}

func joinLabelsByID(ids []string, ctx model.TopologyContext) string {
	labels := make([]string, len(ids))
	for i, id := range ids {
		labels[i] = ctx.NodeByID[id].Label
	}
	return joinLabels(labels)
}
