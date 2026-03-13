package handler

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
			Message: fmt.Sprintf("🔌 協議不匹配：%s → %s 使用了 %s，但 %s 節點通常不使用此協議。",
				source.Label, target.Label, protoName, target.Label),
			Solution: fmt.Sprintf("建議將協議改為 %s。", strings.Join(suggestion, " 或 ")),
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
				Message: fmt.Sprintf("⚠️ 屬性不匹配：%s → %s 使用了 %s 連線，但不應搭配 %s 協議。",
					source.Label, target.Label, connName, protoName),
				Solution: fmt.Sprintf("請將連線類型改為異步 (Async) 或更換為與 %s 相容的協議（如 HTTP/gRPC 用於 Sync，AMQP/MQTT 用於 Async）。", connName),
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
			Message:  "🌐 CDN 全球加速建議：拓補中存在 Client 但缺乏 CDN 節點。",
			Solution: "考慮在 Client 與後端入口之間加入 CDN (Content Delivery Network) 以加速靜態資源分發並減少延遲。",
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
				Message:  "🌐 CDN 全球加速建議：拓補中存在 Client 與 CDN 但兩者未連線。",
				Solution: "建議將 Client 連接至 CDN 節點，以發揮其內容快取與全球加速的優勢。",
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
					Message:  fmt.Sprintf("🔒 SSL 終止建議：Reverse Proxy %q 接收 HTTPS 流量 but has SSL termination disabled。", node.Label),
					Solution: "建議在 Reverse Proxy 啟用 SSL Termination，統一管理 TLS 憑證並減輕後端服務的加解密負擔。",
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
					Message:  fmt.Sprintf("🌐 CDN 孤立提醒：%q 尚未設定 Origin 源站連線。", node.Label),
					Solution: "CDN 需要從源站 (Origin，如 Storage 或 API Gateway) 抓取內容。請建立一條連線並將 Connection Type 設為 'CDN Origin'。",
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
					Message:  fmt.Sprintf("🔗 同步鏈過長：偵測到長度為 %d 的同步呼叫鏈 (%s)。", depth, joinLabelsByID(path, ctx)),
					Solution: "同步鏈過長會導致延遲堆積與可用性下降。建議將部分呼叫改為並行 (Parallel Fan-out) 或異步處理 (Async/MQ)。",
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
					Message:  fmt.Sprintf("🔌 內部通訊協定建議：%s → %s 使用 HTTP。", source.Label, target.Label),
					Solution: "微服務內部通訊建議使用 gRPC 以獲得更高的序列化效率與多路復用優點。",
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

