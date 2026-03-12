package handler

import (
	"fmt"
	"net/http"

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
