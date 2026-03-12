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
	warnings = append(warnings, checkDBSelection(t.Nodes)...)
	warnings = append(warnings, checkVerticalPartitioning(t.Nodes)...)
	warnings = append(warnings, checkCacheConsistency(nodeByID, outgoing)...)
	warnings = append(warnings, checkCAP(t.Nodes)...)
	warnings = append(warnings, checkCDNUsage(t.Nodes)...)
	warnings = append(warnings, checkAsyncDecoupling(nodeByID, outgoing)...)
	warnings = append(warnings, checkLBSPOF(t.Nodes)...)
	warnings = append(warnings, checkReadWriteSeparation(t.Nodes)...)
	warnings = append(warnings, checkCacheEviction(t.Nodes)...)
	warnings = append(warnings, checkProtocolMismatch(nodeByID, t.Edges)...)
	warnings = append(warnings, checkMQConsumer(nodeByID, outgoing)...)
	warnings = append(warnings, checkMQDLQ(t.Nodes)...)
	warnings = append(warnings, checkAsyncPeakShaving(nodeByID, outgoing)...)
	warnings = append(warnings, checkClientToDB(nodeByID, t.Edges)...)
	warnings = append(warnings, checkClientToCache(nodeByID, t.Edges)...)

	return warnings
}

// checkCDNUsage warns if a Client exists but no CDN is found.
func checkCDNUsage(nodes []model.SystemNode) []Warning {
	hasClient := false
	hasCDN := false
	var clientIDs []string
	for _, node := range nodes {
		if node.ComponentType == "client" {
			hasClient = true
			clientIDs = append(clientIDs, node.ID)
		}
		if node.ComponentType == "cdn" {
			hasCDN = true
		}
	}
	if hasClient && !hasCDN {
		return []Warning{{
			Rule:     "cdn_usage",
			Message:  "🌐 CDN 全球加速建議：拓撲中存在 Client 但缺乏 CDN 節點。",
			Solution: "考慮在 Client 與後端入口之間加入 CDN (Content Delivery Network) 以加速靜態資源分發並減少延遲。",
			NodeIDs:  clientIDs,
		}}
	}
	return nil
}

// checkAsyncDecoupling suggests Message Queues for time-consuming operations.
func checkAsyncDecoupling(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	// Keywords that suggest a service might be performing time-consuming or background tasks.
	// Using substrings like "mail" instead of "Email" to match "Gmail", "Mailer", etc.
	keywords := []string{
		"mail", "img", "image", "photo", "pic",
		"vid", "video", "media", "stream", "transcode",
		"report", "pdf", "export", "csv", "excel",
		"task", "worker", "job", "batch", "process",
		"notify", "sms", "push", "upload", "download",
	}
	for id, node := range nodes {
		if node.ComponentType != "service" {
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
			// Check if it's called synchronously
			for sourceID, targets := range outgoing {
				for _, targetID := range targets {
					if targetID == id {
						warnings = append(warnings, Warning{
							Rule:     "async_decoupling",
							Message:  fmt.Sprintf("📬 異步解耦提醒：服務 %q 似乎涉及耗時操作且為同步呼叫。", node.Label),
							Solution: "建議使用 Message Queue (MQ) 將此類操作改為異步處理，以提高系統吞吐量與穩定性。",
							NodeIDs:  []string{sourceID, id},
						})
					}
				}
			}
		}
	}
	return warnings
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// checkLBSPOF warns if there is only one Load Balancer in the entire system.
func checkLBSPOF(nodes []model.SystemNode) []Warning {
	var lbNodes []string
	for _, node := range nodes {
		if node.ComponentType == "load_balancer" {
			lbNodes = append(lbNodes, node.ID)
		}
	}
	if len(lbNodes) == 1 {
		return []Warning{{
			Rule:     "lb_spof",
			Message:  "⚖️ 入口單點故障：整體架構中僅存在 1 個 Load Balancer。",
			Solution: "生產環境建議部署多個 LB 並結合 DNS 負載均衡 (如 Round Robin) 或使用 Active-Passive 備援機制。",
			NodeIDs:  lbNodes,
		}}
	}
	return nil
}

// checkReadWriteSeparation suggests master-slave if read ratio is extremely high.
func checkReadWriteSeparation(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
		if node.ComponentType != "database" {
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
		if node.ComponentType != "cache" {
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

// checkSPOF detects load balancers with only one downstream service node.
func checkSPOF(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	for id, node := range nodes {
		if node.ComponentType != "load_balancer" {
			continue
		}
		targets := outgoing[id]
		var serviceIDs []string
		for _, targetID := range targets {
			if target, ok := nodes[targetID]; ok && target.ComponentType == "service" {
				serviceIDs = append(serviceIDs, targetID)
			}
		}
		if len(serviceIDs) == 1 {
			warnings = append(warnings, Warning{
				Rule: "spof",
				Message: fmt.Sprintf("⚠️ 檢測到單點故障 (SPOF)：Load Balancer %q 後方僅連接 1 個 Service 節點。",
					node.Label),
				Solution: "增加 Service 節點數量或在屬性面板中提高 Replicas 複本數，並確保連線正確。",
				NodeIDs:  append([]string{id}, serviceIDs...),
			})
		}
	}
	return warnings
}

// checkDBSelection flags SQL databases under high write pressure.
func checkDBSelection(nodes []model.SystemNode) []Warning {
	var warnings []Warning
	for _, node := range nodes {
		if node.ComponentType != "database" {
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

// checkVerticalPartitioning warns when multiple database nodes exist (federation pattern).
func checkVerticalPartitioning(nodes []model.SystemNode) []Warning {
	var dbNodes []model.SystemNode
	for _, node := range nodes {
		if node.ComponentType == "database" {
			dbNodes = append(dbNodes, node)
		}
	}
	if len(dbNodes) >= 2 {
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
func checkCacheConsistency(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning {
	var warnings []Warning
	for id, node := range nodes {
		if node.ComponentType != "service" {
			continue
		}
		targets := outgoing[id]
		var involvedIDs []string
		hasCache := false
		hasDB := false
		for _, targetID := range targets {
			if target, ok := nodes[targetID]; ok {
				switch target.ComponentType {
				case "cache":
					hasCache = true
					involvedIDs = append(involvedIDs, targetID)
				case "database":
					hasDB = true
					involvedIDs = append(involvedIDs, targetID)
				}
			}
		}
		if hasCache && hasDB {
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
		if node.ComponentType != "database" {
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
			warnings = append(warnings, Warning{
				Rule: "cap_theorem",
				Message: fmt.Sprintf("📐 CAP 定理：%q (%s) 為 AP 系統。",
					node.Label, dbProps.Product),
				Solution: "注意 AP 系統僅提供最終一致性，若需強一致性（如金流）請更換為 CP 系統（如 RDBMS）。",
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
	},
	"cache": {
		"database": true,
	},
	"message_queue": {
		"amqp": true,
		"mqtt": true,
	},
	"dns": {
		"dns": true,
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
	"dns":       "DNS",
}

// checkProtocolMismatch warns when an edge uses a protocol that doesn't match
// the target component type (e.g. HTTP to a Database).
func checkProtocolMismatch(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, edge := range edges {
		if edge.Protocol == "" {
			continue
		}
		target, ok := nodes[edge.Target]
		if !ok {
			continue
		}
		allowed, hasRule := expectedProtocols[target.ComponentType]
		if !hasRule {
			continue
		}
		if allowed[edge.Protocol] {
			continue
		}

		protoName := protocolDisplayName[edge.Protocol]
		if protoName == "" {
			protoName = edge.Protocol
		}

		var suggestion []string
		for p := range allowed {
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
		if node.ComponentType != "message_queue" {
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
		if node.ComponentType != "message_queue" {
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
		// Entry points: LB or API Gateway
		if node.ComponentType != "load_balancer" && node.ComponentType != "api_gateway" {
			continue
		}

		targets := outgoing[id]
		for _, targetID := range targets {
			target, ok := nodes[targetID]
			if !ok || target.ComponentType != "database" {
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

		if source.ComponentType == "client" && target.ComponentType == "database" {
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

// checkClientToCache prohibits direct connection from Client to Cache.
func checkClientToCache(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning {
	var warnings []Warning
	for _, edge := range edges {
		source, okSource := nodes[edge.Source]
		target, okTarget := nodes[edge.Target]

		if !okSource || !okTarget {
			continue
		}

		if source.ComponentType == "client" && target.ComponentType == "cache" {
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
