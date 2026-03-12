package handler

import (
	"fmt"
	"net/http"

	"github.com/architectmind/backend/model"
	"github.com/gin-gonic/gin"
)

// Warning represents a structured linting result with associated node IDs.
type Warning struct {
	Rule    string   `json:"rule"`
	Message string   `json:"message"`
	NodeIDs []string `json:"nodeIds"`
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
				Rule:    "schema",
				Message: fmt.Sprintf("unknown component type %q on node %q", node.ComponentType, node.ID),
				NodeIDs: []string{node.ID},
			})
		}
		nodeByID[node.ID] = node
	}

	// Build adjacency: source -> list of target node IDs
	outgoing := make(map[string][]string)
	for _, edge := range t.Edges {
		if !model.ValidConnectionTypes[edge.ConnectionType] {
			warnings = append(warnings, Warning{
				Rule:    "schema",
				Message: fmt.Sprintf("unknown connection type %q on edge %q", edge.ConnectionType, edge.ID),
				NodeIDs: []string{},
			})
		}
		if _, ok := nodeByID[edge.Source]; !ok {
			warnings = append(warnings, Warning{
				Rule:    "schema",
				Message: fmt.Sprintf("edge %q references unknown source node %q", edge.ID, edge.Source),
				NodeIDs: []string{},
			})
		}
		if _, ok := nodeByID[edge.Target]; !ok {
			warnings = append(warnings, Warning{
				Rule:    "schema",
				Message: fmt.Sprintf("edge %q references unknown target node %q", edge.ID, edge.Target),
				NodeIDs: []string{},
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
				Message: fmt.Sprintf("⚠️ 檢測到單點故障 (SPOF)：Load Balancer %q 後方僅連接 1 個 Service 節點。建議增加冗餘實例以提升可用性 (Availability)。",
					node.Label),
				NodeIDs: append([]string{id}, serviceIDs...),
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
				Message: fmt.Sprintf("⚖️ SQL 擴展性取捨：%q 為高寫入壓力 (讀寫比 %.0f%%)，建議考慮 Master-Slave 讀寫分離，或評估是否切換至 NoSQL 以獲得更好的寫入擴展性。",
					node.Label, dbProps.ReadWriteRatio*100),
				NodeIDs: []string{node.ID},
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
			Message: fmt.Sprintf("🔍 垂直拆分 (Federation) 提醒：偵測到 %d 個資料庫節點 (%s)。此架構無法執行跨庫 Join，應用層需處理資料聚合，並注意跨庫事務的一致性 (Distributed Transactions)。",
				len(dbNodes), joinLabels(labels)),
			NodeIDs: ids,
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
				Message: fmt.Sprintf("⚡ 快取權衡：Service %q 同時連接 Cache 與 Database。請選擇更新策略（Write-through, Write-back 或 Cache-aside）。注意 Write-back 效能最高但有遺失資料風險。",
					node.Label),
				NodeIDs: append([]string{id}, involvedIDs...),
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
				Message: fmt.Sprintf("📐 CAP 定理：%q (%s) 為 AP 系統。優點是高可用性，但需接受「最終一致性 (Eventual Consistency)」，不建議用於金流結算。",
					node.Label, dbProps.Product),
				NodeIDs: []string{node.ID},
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
