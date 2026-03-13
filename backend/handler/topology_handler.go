package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/architectmind/backend/model"
	"github.com/gin-gonic/gin"
)

// Warning represents a structured linting result with associated node IDs.
type Warning struct {
	Rule     string   `json:"rule"`
	Message  string   `json:"message"`
	Solution string   `json:"solution"`
	NodeIDs  []string `json:"nodeIds"`
}

// AnalyzeResponse is returned after parsing and validating the topology.
type AnalyzeResponse struct {
	Success   bool      `json:"success"`
	NodeCount int       `json:"nodeCount"`
	EdgeCount int       `json:"edgeCount"`
	Warnings  []Warning `json:"warnings,omitempty"`
}

// PostTopology receives a SystemTopology JSON payload, validates it,
// and returns a summary. This will later feed into the rule engine.
func PostTopology(c *gin.Context) {
	var topology model.SystemTopology
	if err := c.ShouldBindJSON(&topology); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("invalid payload: %v", err),
		})
		return
	}

	warnings := validate(topology)

	c.JSON(http.StatusOK, AnalyzeResponse{
		Success:   true,
		NodeCount: len(topology.Nodes),
		EdgeCount: len(topology.Edges),
		Warnings:  warnings,
	})
}

// validate performs semantic checks and architecture linting on the topology.
func validate(t model.SystemTopology) []Warning {
	var warnings []Warning

	// Build lookup maps
	nodeByID := make(map[string]model.SystemNode, len(t.Nodes))
	for _, node := range t.Nodes {
		if !model.ValidComponentTypes[node.ComponentType] {
			warnings = append(warnings, Warning{
				Rule:     "schema",
				Message:  fmt.Sprintf("unknown component type %q on node %q", node.ComponentType, node.ID),
				Solution: "請在節點屬性中選擇正確的組件類型 (Component Type)。",
				NodeIDs:  []string{node.ID},
			})
		}
		// Validate all roles in merged nodes
		for _, role := range node.Roles {
			if !model.ValidComponentTypes[role] {
				warnings = append(warnings, Warning{
					Rule:     "schema",
					Message:  fmt.Sprintf("unknown role %q on node %q", role, node.ID),
					Solution: "請確保合併節點的所有角色都是有效的組件類型。",
					NodeIDs:  []string{node.ID},
				})
			}
		}
		nodeByID[node.ID] = node
	}

	// Build adjacency: source -> list of target node IDs
	outgoing := make(map[string][]string)
	for _, edge := range t.Edges {
		if !model.ValidConnectionTypes[edge.ConnectionType] {
			warnings = append(warnings, Warning{
				Rule:     "schema",
				Message:  fmt.Sprintf("unknown connection type %q on edge %q", edge.ConnectionType, edge.ID),
				Solution: "請在連線屬性中選擇正確的連線類型 (Connection Type)。",
				NodeIDs:  []string{},
			})
		}
		if _, ok := nodeByID[edge.Source]; !ok {
			warnings = append(warnings, Warning{
				Rule:     "schema",
				Message:  fmt.Sprintf("edge %q references unknown source node %q", edge.ID, edge.Source),
				Solution: "請確保連線的出發點已連接到有效的節點。",
				NodeIDs:  []string{},
			})
		}
		if _, ok := nodeByID[edge.Target]; !ok {
			warnings = append(warnings, Warning{
				Rule:     "schema",
				Message:  fmt.Sprintf("edge %q references unknown target node %q", edge.ID, edge.Target),
				Solution: "請確保連線的目標點已連接到有效的節點。",
				NodeIDs:  []string{},
			})
		}
		outgoing[edge.Source] = append(outgoing[edge.Source], edge.Target)
	}

	warnings = append(warnings, checkSPOF(nodeByID, outgoing)...)
	warnings = append(warnings, checkDBSelection(t.Nodes, t.Edges)...)
	warnings = append(warnings, checkVerticalPartitioning(nodeByID, outgoing)...)
	warnings = append(warnings, checkCacheConsistency(nodeByID, outgoing)...)
	warnings = append(warnings, checkCAP(t.Nodes)...)
	warnings = append(warnings, checkCDNUsage(nodeByID, outgoing)...)
	warnings = append(warnings, checkAsyncDecoupling(nodeByID, t.Edges)...)
	warnings = append(warnings, checkLBSPOF(t.Nodes)...)
	warnings = append(warnings, checkReadWriteSeparation(t.Nodes, t.Edges)...)
	warnings = append(warnings, checkCacheEviction(t.Nodes)...)
	warnings = append(warnings, checkProtocolMismatch(nodeByID, t.Edges)...)
	warnings = append(warnings, checkConnectionTypeProtocolMismatch(nodeByID, t.Edges)...)
	warnings = append(warnings, checkMQConsumer(nodeByID, outgoing)...)
	warnings = append(warnings, checkMQDLQ(t.Nodes)...)
	warnings = append(warnings, checkAsyncPeakShaving(nodeByID, outgoing)...)
	warnings = append(warnings, checkClientToDB(nodeByID, t.Edges)...)
	warnings = append(warnings, checkClientToCache(nodeByID, t.Edges)...)
	warnings = append(warnings, checkReverseProxySPOF(t.Nodes)...)
	warnings = append(warnings, checkReverseProxySSL(nodeByID, t.Edges)...)
	warnings = append(warnings, checkMissingFirewall(nodeByID, outgoing)...)
	warnings = append(warnings, checkMissingLogger(nodeByID, outgoing)...)
	warnings = append(warnings, checkFirewallMonitorMode(t.Nodes)...)
	warnings = append(warnings, checkFirewallL3Only(nodeByID)...)
	warnings = append(warnings, checkIncompleteObservability(t.Nodes)...)
	warnings = append(warnings, checkAlertingDisabled(t.Nodes)...)
	warnings = append(warnings, checkServerlessReplicas(t.Nodes)...)
	warnings = append(warnings, checkNoAutoScalingSingle(t.Nodes)...)
	warnings = append(warnings, checkNoHealthCheckBehindLB(t.Nodes, t.Edges)...)

	return warnings
}

// checkCDNUsage warns if a Client exists but no CDN is found, or if both exist but are not connected.
func checkCDNUsage(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var clientIDs []string
	var cdnIDs []string
	for id, node := range nodes {
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
			Message:  "🌐 CDN 全球加速建議：拓撲中存在 Client 但缺乏 CDN 節點。",
			Solution: "考慮在 Client 與後端入口之間加入 CDN (Content Delivery Network) 以加速靜態資源分發並減少延遲。",
			NodeIDs:  clientIDs,
		}}
	}

	if len(clientIDs) > 0 && len(cdnIDs) > 0 {
		// Check if any client connects to a cdn
		connected := false
		for _, clientID := range clientIDs {
			targets := outgoing[clientID]
			for _, targetID := range targets {
				if model.NodeHasRole(nodes[targetID], "cdn") {
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
				Message:  "🌐 CDN 全球加速建議：拓撲中存在 Client 與 CDN 但兩者未連線。",
				Solution: "建議將 Client 連接至 CDN 節點，以發揮其內容快取與全球加速的優勢。",
				NodeIDs:  append(clientIDs, cdnIDs...),
			}}
		}
	}

	return nil
}

// checkAsyncDecoupling suggests Message Queues for time-consuming operations.
func checkAsyncDecoupling(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	keywords := []string{
		"mail", "img", "image", "photo", "pic",
		"vid", "video", "media", "stream", "transcode",
		"report", "pdf", "export", "csv", "excel",
		"task", "worker", "job", "batch", "process",
		"notify", "sms", "push", "upload", "download",
		"ai", "ml", "inference", "prediction", "training",
		"payment", "billing", "checkout", "payout",
		"index", "search", "crawl", "scrape",
		"sync", "migration", "archive", "cleanup",
		"audit", "analytics", "stats",
		"webhook", "slack", "discord", "telegram",
	}

	for id, node := range nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		isTimeConsuming := false
		for _, kw := range keywords {
			if contains(node.Label, kw) {
				isTimeConsuming = true
				break
			}
		}

		if isTimeConsuming {
			// Find who is calling this service
			var synchronousCallers []string
			isDecoupled := false

			for _, edge := range edges {
				if edge.Target == id {
					sourceNode := nodes[edge.Source]
					// If any caller is a Message Queue, or the connection itself is async, we consider it decoupled!
					if model.NodeHasRole(sourceNode, "message_queue") || edge.ConnectionType == "async" {
						isDecoupled = true
						break
					}
					synchronousCallers = append(synchronousCallers, edge.Source)
				}
			}

			// If called synchronously and not through an MQ, issue warning
			if isTimeConsuming && !isDecoupled && len(synchronousCallers) > 0 {
				warnings = append(warnings, Warning{
					Rule:     "async_decoupling",
					Message:  fmt.Sprintf("📬 異步解耦提醒：服務 %q 似乎涉及耗時操作且目前接收同步呼叫。", node.Label),
					Solution: "建議在呼叫方與此服務之間加入 Message Queue (MQ)，將操作改為異步處理，以提高系統吞吐量。",
					NodeIDs:  append([]string{id}, synchronousCallers...),
				})
			}
		}
	}
	return warnings
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// checkLBSPOF warns if there is only one Load Balancer in the entire system,
// but suppresses the warning if that Load Balancer has Replicas > 1.
func checkLBSPOF(nodes []model.SystemNode) []Warning {
	var lbNodes []model.SystemNode
	for _, node := range nodes {
		if model.NodeHasRole(node, "load_balancer") {
			lbNodes = append(lbNodes, node)
		}
	}

	if len(lbNodes) == 1 {
		node := lbNodes[0]
		props, err := model.ParseNodeProperties(node)
		if err == nil {
			if lbProps, ok := props.(*model.LoadBalancerProperties); ok && lbProps.Replicas > 1 {
				// Multiple replicas of the same LB node solve the SPOF
				return nil
			}
		}

		return []Warning{{
			Rule:     "lb_spof",
			Message:  "⚖️ 入口單點故障：整體架構中僅存在 1 個 Load Balancer。",
			Solution: "生產環境建議部署多個 LB，或在屬性面板中將 Replicas 複本數設為 2 以上。",
			NodeIDs:  []string{node.ID},
		}}
	}
	return nil
}

// checkReadWriteSeparation suggests master-slave if read ratio is extremely high.
func checkReadWriteSeparation(nodes []model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, node := range nodes {
		if !model.NodeHasRole(node, "database") {
			continue
		}

		// Suppress warning if replication is already set up
		hasReplication := false
		for _, edge := range edges {
			if (edge.Source == node.ID || edge.Target == node.ID) && edge.ConnectionType == "replication" {
				hasReplication = true
				break
			}
		}
		if hasReplication {
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
		if dbProps.ReadWriteRatio > 0.8 {
			warnings = append(warnings, Warning{
				Rule:     "read_write_separation",
				Message:  fmt.Sprintf("📖 讀寫分離建議：%q 讀取比例極高 (%.0f%%)。", node.Label, dbProps.ReadWriteRatio*100),
				Solution: "建議導入 Master-Slave 讀寫分離架構，主庫負責寫入，複本庫負責讀取以提升效能。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkCacheEviction warns if a Cache node lacks an eviction policy.
func checkCacheEviction(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
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

// checkSPOF detects load balancers and reverse proxies with only one downstream service node,
// but suppresses the warning if that service node has Replicas > 1.
func checkSPOF(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	for id, node := range nodes {
		if !model.NodeHasRole(node, "load_balancer") && !model.NodeHasRole(node, "reverse_proxy") {
			continue
		}
		targets := outgoing[id]
		var serviceIDs []string
		isRedundant := false

		for _, targetID := range targets {
			target, ok := nodes[targetID]
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
				Solution: "增加 Service 節點數量或在屬性面板中提高 Replicas 複本數。",
				NodeIDs:  append([]string{id}, serviceIDs...),
			})
		} else if len(serviceIDs) > 1 {
			// Multiple distinct nodes also solve the SPOF
			isRedundant = true
		}
	}
	return warnings
}

// checkDBSelection flags SQL databases under high write pressure.
func checkDBSelection(nodes []model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, node := range nodes {
		if !model.NodeHasRole(node, "database") {
			continue
		}

		// Suppress warning if replication is already set up
		hasReplication := false
		for _, edge := range edges {
			if (edge.Source == node.ID || edge.Target == node.ID) && edge.ConnectionType == "replication" {
				hasReplication = true
				break
			}
		}
		if hasReplication {
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
		if dbProps.DBType == "sql" && dbProps.ReadWriteRatio < 0.5 {
			warnings = append(warnings, Warning{
				Rule: "db_selection",
				Message: fmt.Sprintf("⚖️ SQL 擴展性取捨：%q 為高寫入壓力 (讀寫比 %.0f%%)。",
					node.Label, dbProps.ReadWriteRatio*100),
				Solution: "實施 Master-Slave 讀寫分離，或考慮切換為寫入性能更好的 NoSQL 資料庫。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkVerticalPartitioning warns when multiple database nodes exist (federation pattern),
// but suppresses the warning if a "Service" node is found connecting to two or more different database nodes,
// which indicates that the application layer is handling the data aggregation.
func checkVerticalPartitioning(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var dbNodes []model.SystemNode
	dbSet := make(map[string]bool)
	for _, node := range nodes {
		if model.NodeHasRole(node, "database") {
			dbNodes = append(dbNodes, node)
			dbSet[node.ID] = true
		}
	}

	// Only proceed if there are at least 2 database nodes
	if len(dbNodes) < 2 {
		return nil
	}

	// Search for an "Aggregator" service:
	// A service node that has outgoing edges to at least two different database nodes.
	isAggregated := false
	for id, node := range nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		connectedDBCount := 0
		targets := outgoing[id]
		seenDBs := make(map[string]bool)
		for _, targetID := range targets {
			if dbSet[targetID] && !seenDBs[targetID] {
				connectedDBCount++
				seenDBs[targetID] = true
			}
		}

		if connectedDBCount >= 2 {
			isAggregated = true
			break
		}
	}

	// If no aggregator is found, show the warning
	if !isAggregated {
		labels := make([]string, len(dbNodes))
		ids := make([]string, len(dbNodes))
		for i, n := range dbNodes {
			labels[i] = n.Label
			ids[i] = n.ID
		}
		return []Warning{{
			Rule: "federation",
			Message: fmt.Sprintf("🔍 垂直拆分 (Federation) 提醒：偵測到 %d 個資料庫節點 (%s)。",
				len(dbNodes), joinLabels(labels)),
			Solution: "確保應用層支援跨庫資料聚合，並注意跨庫事務的一致性問題。",
			NodeIDs:  ids,
		}}
	}

	return nil
}

// checkCacheConsistency detects services connected to both cache and database.
// The warning is suppressed if at least one connected cache node has a non-zero TTLSeconds configured.
func checkCacheConsistency(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	for id, node := range nodes {
		if !model.NodeHasRole(node, "service") {
			continue
		}
		targets := outgoing[id]
		var involvedIDs []string
		hasCache := false
		hasDB := false
		anyCacheHasTTL := false

		for _, targetID := range targets {
			if target, ok := nodes[targetID]; ok {
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

// checkCAP flags databases that are known AP systems under CAP theorem.
func checkCAP(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
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

		if model.APProducts[dbProps.Product] {
			// If user explicitly chose Eventual Consistency, they acknowledge the AP nature.
			if dbProps.ConsistencyLevel == "eventual" {
				continue
			}

			// If user chose Strong Consistency on an AP system, warn about performance.
			if dbProps.ConsistencyLevel == "strong" {
				warnings = append(warnings, Warning{
					Rule: "cap_theorem",
					Message: fmt.Sprintf("🚀 效能與一致性權衡：%q (%s) 為 AP 系統，但您要求強一致性。",
						node.Label, dbProps.Product),
					Solution: "在 AP 系統上實施強一致性 (如 Quorum R/W) 將顯著增加延遲並降低可用性。請評估是否必要。",
					NodeIDs:  []string{node.ID},
				})
				continue
			}

			// Default warning if no explicit choice is made.
			warnings = append(warnings, Warning{
				Rule: "cap_theorem",
				Message: fmt.Sprintf("📐 CAP 定理：%q (%s) 為 AP 系統。",
					node.Label, dbProps.Product),
				Solution: "注意 AP 系統僅提供最終一致性。若需強一致性，請在屬性中選擇 'Strong' (注意效能) 或更換為 CP 系統 (如 RDBMS)。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// expectedProtocols defines which protocols are appropriate for connections
// targeting a given component type. A nil value means any protocol is acceptable.
var expectedProtocols = map[string]map[string]bool{
	"database": {
		"database": true,
		"resp":     true,
		"binary":   true,
		"uds":      true,
		"tcp":      true,
	},
	"cache": {
		"database": true,
		"resp":     true,
		"binary":   true,
		"uds":      true,
		"tcp":      true,
		"http":     true, // Couchbase/Elasticsearch
	},
	"message_queue": {
		"amqp": true,
		"mqtt": true,
		"tcp":  true,
	},
	"dns": {
		"dns": true,
		"udp": true,
	},
	"client": {
		"http":      true,
		"https":     true,
		"websocket": true,
	},
	"cdn": {
		"http":  true,
		"https": true,
	},
	"storage": {
		"http":  true,
		"https": true,
	},
}

// protocolDisplayName maps protocol values to human-readable names.
var protocolDisplayName = map[string]string{
	"http":      "HTTP",
	"https":     "HTTPS",
	"grpc":      "gRPC",
	"websocket": "WebSocket",
	"ssh":       "SSH",
	"tcp":       "TCP",
	"udp":       "UDP",
	"amqp":      "AMQP",
	"mqtt":      "MQTT",
	"database":  "Database Protocol",
	"resp":      "RESP (Redis)",
	"binary":    "Binary Protocol",
	"uds":       "UDS (Unix Domain Socket)",
	"dns":       "DNS",
}

// validConnectionProtocolPairs defines which protocols are allowed for each connection type.
var validConnectionProtocolPairs = map[string]map[string]bool{
	"sync": {
		"unspecified": true,
		"http":        true,
		"https":       true,
		"grpc":        true,
		"websocket":   true,
		"ssh":         true,
		"tcp":         true,
		"database":    true,
		"resp":        true,
		"binary":      true,
		"uds":         true,
		"dns":         true,
	},
	"async": {
		"unspecified": true,
		"amqp":        true,
		"mqtt":        true,
		"http":        true, // Webhooks
		"https":       true, // Webhooks
		"tcp":         true,
		"udp":         true,
	},
	"replication": {
		"unspecified": true,
		"database":    true,
		"binary":      true,
		"tcp":         true,
		"udp":         true,
		"ssh":         true, // rsync over ssh
	},
	"cdn_origin": {
		"unspecified": true,
		"http":        true,
		"https":       true,
	},
}

// checkConnectionTypeProtocolMismatch warns when a connection type and protocol are logically inconsistent.
func checkConnectionTypeProtocolMismatch(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning

	connectionTypeNames := map[string]string{
		"sync":        "Synchronous",
		"async":       "Asynchronous",
		"replication": "Replication",
		"cdn_origin":  "CDN Origin",
	}

	for _, edge := range edges {
		proto := edge.Protocol
		if proto == "" {
			proto = "unspecified"
		}

		allowed, hasRule := validConnectionProtocolPairs[edge.ConnectionType]
		if !hasRule {
			continue
		}

		if !allowed[proto] {
			source := nodes[edge.Source]
			target := nodes[edge.Target]
			connName := connectionTypeNames[edge.ConnectionType]
			protoName := protocolDisplayName[proto]
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

// checkProtocolMismatch warns when an edge uses a protocol that doesn't match
// the target component type (e.g. HTTP to a Database).
func checkProtocolMismatch(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, edge := range edges {
		if edge.Protocol == "" || edge.Protocol == "unspecified" {
			continue
		}
		target, ok := nodes[edge.Target]
		if !ok {
			continue
		}

		// For merged nodes, build the union of expected protocols across all roles
		allAllowed := make(map[string]bool)
		hasAnyRule := false
		for _, role := range model.GetEffectiveRoles(target) {
			if allowed, ok := expectedProtocols[role]; ok {
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

		protoName := protocolDisplayName[edge.Protocol]
		if protoName == "" {
			protoName = edge.Protocol
		}

		var suggestion []string
		for p := range allAllowed {
			name := protocolDisplayName[p]
			if name == "" {
				name = p
			}
			suggestion = append(suggestion, name)
		}

		source := nodes[edge.Source]
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

func joinLabels(labels []string) string {
	result := ""
	for i, l := range labels {
		if i > 0 {
			result += ", "
		}
		result += l
	}
	return result
}

// checkMQConsumer warns if a Message Queue node has no outgoing connections.
func checkMQConsumer(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	for id, node := range nodes {
		if !model.NodeHasRole(node, "message_queue") {
			continue
		}
		if len(outgoing[id]) == 0 {
			warnings = append(warnings, Warning{
				Rule:     "mq_consumer_missing",
				Message:  fmt.Sprintf("📥 MQ 消費者缺失檢查：%q 目前沒有任何消費者 (Consumer)。", node.Label),
				Solution: "Message Queue 節點缺乏輸出連線，訊息將在隊列中堆積。請將此節點連接至處理訊息的 Service。",
				NodeIDs:  []string{id},
			})
		}
	}
	return warnings
}

// checkMQDLQ warns if a Message Queue node does not have a dead letter queue configured.
func checkMQDLQ(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
		if !model.NodeHasRole(node, "message_queue") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		mqProps, ok := props.(*model.MessageQueueProperties)
		if !ok {
			continue
		}
		if !mqProps.HasDLQ {
			warnings = append(warnings, Warning{
				Rule:     "mq_dlq_missing",
				Message:  fmt.Sprintf("💀 死信隊列 (DLQ) 提醒：%q 未配置死信隊列或重試機制。", node.Label),
				Solution: "使用 Message Queue 但未配置死信隊列 (Dead Letter Queue)，可能導致處理失敗的訊息直接丟失。建議在屬性中啟用 DLQ。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkAsyncPeakShaving suggests using MQ for high-load direct database writes from entry points.
func checkAsyncPeakShaving(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	for id, node := range nodes {
		// Entry points: LB, API Gateway, or Reverse Proxy
		if !model.NodeHasRole(node, "load_balancer") && !model.NodeHasRole(node, "api_gateway") && !model.NodeHasRole(node, "reverse_proxy") {
			continue
		}

		targets := outgoing[id]
		for _, targetID := range targets {
			target, ok := nodes[targetID]
			if !ok || !model.NodeHasRole(target, "database") {
				continue
			}

			// Check for high write ratio
			props, err := model.ParseNodeProperties(target)
			if err != nil {
				continue
			}
			dbProps, ok := props.(*model.DatabaseProperties)
			if !ok {
				continue
			}

			if dbProps.ReadWriteRatio < 0.5 {
				warnings = append(warnings, Warning{
					Rule:     "async_peak_shaving",
					Message:  fmt.Sprintf("🌊 異步削峰實踐建議：流量入口 %q 直接連線至高負載寫入資料庫 %q。", node.Label, target.Label),
					Solution: "高頻寫入建議先將請求發送至 Message Queue (MQ) 進行削峰填谷 (Load Leveling)，再由後端異步寫入資料庫，以減輕資料庫壓力並提升系統穩定性。",
					NodeIDs:  []string{id, targetID},
				})
			}
		}
	}
	return warnings
}

// checkClientToDB prohibits direct connection from Client to Database.
func checkClientToDB(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, edge := range edges {
		source, okSource := nodes[edge.Source]
		target, okTarget := nodes[edge.Target]

		if !okSource || !okTarget {
			continue
		}

		if model.NodeHasRole(source, "client") && model.NodeHasRole(target, "database") {
			warnings = append(warnings, Warning{
				Rule:     "client_direct_db",
				Message:  fmt.Sprintf("🚫 安全風險：禁止從 %q 直接連線至 %q。", source.Label, target.Label),
				Solution: "Client 不應直接操作資料庫。請在兩者之間加入 API Gateway 或 Service 層進行身份驗證與數據抽象。",
				NodeIDs:  []string{source.ID, target.ID},
			})
		}
	}
	return warnings
}

// checkReverseProxySPOF warns if there is only one Reverse Proxy in the entire system
// and its replicas are <= 1.
func checkReverseProxySPOF(nodes []model.SystemNode) []Warning {
	var rpNodes []model.SystemNode
	for _, node := range nodes {
		if model.NodeHasRole(node, "reverse_proxy") {
			rpNodes = append(rpNodes, node)
		}
	}

	if len(rpNodes) == 1 {
		node := rpNodes[0]
		props, err := model.ParseNodeProperties(node)
		if err == nil {
			if rpProps, ok := props.(*model.ReverseProxyProperties); ok && rpProps.Replicas > 1 {
				return nil
			}
		}

		return []Warning{{
			Rule:     "reverse_proxy_spof",
			Message:  "🔀 入口單點故障：整體架構中僅存在 1 個 Reverse Proxy。",
			Solution: "生產環境建議部署多個 Reverse Proxy，或在屬性面板中將 Replicas 複本數設為 2 以上。",
			NodeIDs:  []string{node.ID},
		}}
	}
	return nil
}

// checkReverseProxySSL warns if a reverse proxy receives HTTPS traffic but has SSL termination disabled.
func checkReverseProxySSL(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, node := range nodes {
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
		for _, edge := range edges {
			if edge.Target == node.ID && edge.Protocol == "https" {
				warnings = append(warnings, Warning{
					Rule:     "reverse_proxy_ssl",
					Message:  fmt.Sprintf("🔒 SSL 終止建議：Reverse Proxy %q 接收 HTTPS 流量但未啟用 SSL Termination。", node.Label),
					Solution: "建議在 Reverse Proxy 啟用 SSL Termination，統一管理 TLS 憑證並減輕後端服務的加解密負擔。",
					NodeIDs:  []string{node.ID},
				})
				break
			}
		}
	}
	return warnings
}

// checkClientToCache prohibits direct connection from Client to Cache.
func checkClientToCache(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, edge := range edges {
		source, okSource := nodes[edge.Source]
		target, okTarget := nodes[edge.Target]

		if !okSource || !okTarget {
			continue
		}

		if model.NodeHasRole(source, "client") && model.NodeHasRole(target, "cache") {
			warnings = append(warnings, Warning{
				Rule:     "client_direct_cache",
				Message:  fmt.Sprintf("🧊 暴露風險：不建議從 %q 直接連線至 %q。", source.Label, target.Label),
				Solution: "不建議 Client 直接操作快取。這可能導致快取穿透風險或數據洩漏。應透過後端 Service 進行快取邏輯封裝。",
				NodeIDs:  []string{source.ID, target.ID},
			})
		}
	}
	return warnings
}

// checkMissingFirewall warns if there's a Client and LB/API Gateway but no Firewall,
// or if a Firewall exists but is not connected to the entry points.
func checkMissingFirewall(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	hasClient := false
	hasEntryPoint := false
	hasFirewall := false
	var clientIDs []string
	var entryPointIDs []string
	var firewallIDs []string

	for _, node := range nodes {
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
			targets := outgoing[firewallID]
			for _, targetID := range targets {
				if model.NodeHasRole(nodes[targetID], "load_balancer") || model.NodeHasRole(nodes[targetID], "api_gateway") {
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

// checkMissingLogger warns if there are 3+ Services but no Logger,
// or if Logger exists but is not connected to any Service.
func checkMissingLogger(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	serviceCount := 0
	hasLogger := false
	var serviceIDs []string
	var loggerIDs []string

	for _, node := range nodes {
		if model.NodeHasRole(node, "service") {
			serviceCount++
			serviceIDs = append(serviceIDs, node.ID)
		}
		if model.NodeHasRole(node, "logger") || model.NodeHasRole(node, "monitor") {
			hasLogger = true
			loggerIDs = append(loggerIDs, node.ID)
		}
	}

	if serviceCount >= 3 && !hasLogger {
		if len(serviceIDs) > 3 {
			serviceIDs = serviceIDs[:3]
		}
		return []Warning{{
			Rule:     "missing_observability",
			Message:  "📊 缺少 Logger/Monitor：架構中有 3 個（含）以上的 Service 但缺少 Logger/Monitor。",
			Solution: "建議加入 Logger/Monitor 節點（如 ELK Stack、Prometheus + Grafana、Datadog），並讓各 Service 連線至該節點。",
			NodeIDs:  serviceIDs,
		}}
	}

	if serviceCount >= 3 && hasLogger {
		connected := false
		for _, loggerID := range loggerIDs {
			targets := outgoing[loggerID]
			for _, targetID := range targets {
				if model.NodeHasRole(nodes[targetID], "service") {
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
				Rule:     "missing_observability",
				Message:  "📊 Logger/Monitor 未正確連線：架構中有 Logger/Monitor 但未連線至任何 Service。",
				Solution: "建議將各 Service 連接至 Logger/Monitor 節點，以收集日誌與監控數據。",
				NodeIDs:  append(loggerIDs, serviceIDs[:3]...),
			}}
		}
	}

	return nil
}

// checkFirewallMonitorMode warns if a Firewall is in monitor mode (not blocking).
func checkFirewallMonitorMode(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
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
func checkFirewallL3Only(nodes map[string]model.SystemNode) []Warning {
	hasAPIGateway := false
	for _, node := range nodes {
		if model.NodeHasRole(node, "api_gateway") {
			hasAPIGateway = true
			break
		}
	}
	if !hasAPIGateway {
		return nil
	}

	var warnings []Warning
	for _, node := range nodes {
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

// checkIncompleteObservability warns if a Logger collects only partial telemetry
// in an architecture with 3+ services.
func checkIncompleteObservability(nodes []model.SystemNode) []Warning {
	serviceCount := 0
	for _, node := range nodes {
		if model.NodeHasRole(node, "service") {
			serviceCount++
		}
	}
	if serviceCount < 3 {
		return nil
	}

	var warnings []Warning
	for _, node := range nodes {
		if !model.NodeHasRole(node, "logger") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		logProps, ok := props.(*model.LoggerProperties)
		if !ok {
			continue
		}
		if logProps.LogType == "all" || logProps.LogType == "" {
			continue
		}
		if logProps.LogType == "metrics" {
			warnings = append(warnings, Warning{
				Rule:     "incomplete_observability",
				Message:  fmt.Sprintf("📊 觀測性不完整：%q 僅收集 Metrics，缺少 Logs 和 Traces。微服務架構中缺乏 distributed tracing 將難以定位跨服務的效能瓶頸。", node.Label),
				Solution: "建議將 logType 設為 All，或額外加入 Traces 類型的 Logger（如 Jaeger）實現完整觀測性。",
				NodeIDs:  []string{node.ID},
			})
		} else if logProps.LogType == "logs" {
			warnings = append(warnings, Warning{
				Rule:     "incomplete_observability",
				Message:  fmt.Sprintf("📊 觀測性不完整：%q 僅收集 Logs，缺少 Metrics 和 Traces。沒有 Metrics 將無法設定有效的告警規則。", node.Label),
				Solution: "建議將 logType 設為 All，或額外加入 Metrics 類型的 Logger（如 Prometheus）。",
				NodeIDs:  []string{node.ID},
			})
		} else if logProps.LogType == "traces" {
			warnings = append(warnings, Warning{
				Rule:     "incomplete_observability",
				Message:  fmt.Sprintf("📊 觀測性不完整：%q 僅收集 Traces，缺少 Metrics 和 Logs。缺乏 Metrics 和 Logs 將無法進行全面的問題診斷。", node.Label),
				Solution: "建議將 logType 設為 All，或額外加入 Metrics 和 Logs 類型的 Logger。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkAlertingDisabled warns if a Logger node has alerting turned off.
func checkAlertingDisabled(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
		if !model.NodeHasRole(node, "logger") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		logProps, ok := props.(*model.LoggerProperties)
		if !ok {
			continue
		}
		if !logProps.Alerting {
			warnings = append(warnings, Warning{
				Rule:     "alerting_disabled",
				Message:  fmt.Sprintf("🔔 告警未啟用：%q 未啟用告警功能。問題發生時將無法及時收到通知，只能依賴人工巡檢。", node.Label),
				Solution: "建議啟用告警並配置通知管道（如 Slack、PagerDuty、Email），設定關鍵指標的閾值觸發條件。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkServerlessReplicas warns if a serverless service has manually set replicas.
func checkServerlessReplicas(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
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

// checkNoAutoScalingSingle warns if a service has only one replica and auto-scaling is disabled.
func checkNoAutoScalingSingle(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
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
func checkNoHealthCheckBehindLB(nodes []model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning

	// Build a map of nodes by ID
	nodeByID := make(map[string]model.SystemNode)
	for _, node := range nodes {
		nodeByID[node.ID] = node
	}

	// Build a map of nodes that are Load Balancers
	lbNodes := make(map[string]bool)
	for _, node := range nodes {
		if model.NodeHasRole(node, "load_balancer") {
			lbNodes[node.ID] = true
		}
	}

	// For each edge from LB to Service, check if Service has healthCheck
	for _, edge := range edges {
		if !lbNodes[edge.Source] {
			continue
		}
		target, ok := nodeByID[edge.Target]
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
