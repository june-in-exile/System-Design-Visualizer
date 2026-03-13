package handler

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkDBSelection flags SQL databases under high write pressure.
func checkDBSelection(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "database") {
			continue
		}

		// Suppress warning if replication is already set up
		if hasReplicationEdge(node.ID, ctx.Edges) {
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

// checkReadWriteSeparation suggests master-slave if read ratio is extremely high.
func checkReadWriteSeparation(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "database") {
			continue
		}

		// Suppress warning if replication is already set up
		if hasReplicationEdge(node.ID, ctx.Edges) {
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

// checkCAP flags databases that are known AP systems under CAP theorem.
func checkCAP(ctx model.TopologyContext) []Warning {
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

// checkVerticalPartitioning warns when multiple database nodes exist (federation pattern),
// but suppresses the warning if a "Service" node is found connecting to two or more different database nodes,
// which indicates that the application layer is handling the data aggregation.
func checkVerticalPartitioning(ctx model.TopologyContext) []Warning {
	var dbNodes []model.SystemNode
	dbSet := make(map[string]bool)
	for _, node := range ctx.Nodes {
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
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		connectedDBCount := 0
		targets := ctx.Outgoing[id]
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

// checkSearchDatabase warns if a service labeled "search" connects directly to SQL/NoSQL.
// This handles the "search engine selection" issue (C4/T7/Y6).
func checkSearchDatabase(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") || !labelContains(node.Label, "search") {
			continue
		}

		targets := ctx.Outgoing[id]
		for _, targetID := range targets {
			target, ok := ctx.NodeByID[targetID]
			if !ok || !model.NodeHasRole(target, "database") {
				continue
			}

			props, err := model.ParseNodeProperties(target)
			if err != nil {
				continue
			}
			dbProps, ok := props.(*model.DatabaseProperties)
			if !ok {
				continue
			}

			if dbProps.Product != "elasticsearch" && dbProps.Product != "solr" {
				warnings = append(warnings, Warning{
					Rule:     "search_engine_recommendation",
					Message:  fmt.Sprintf("🔍 搜尋引擎建議：搜尋服務 %q 直接連線至 %q。", node.Label, target.Label),
					Solution: "全文搜尋或複雜篩選在 RDBMS/NoSQL 上效率較低。建議使用專門的搜尋引擎（如 Elasticsearch, Solr），並透過同步機制確保資料一致性。",
					NodeIDs:  []string{id, targetID},
				})
			}
		}
	}
	return warnings
}

// checkServiceDataSource warns if a service has no outgoing connections to data sources.
// This handles the "missing data source" issue (C6/G4/G5).
func checkServiceDataSource(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		// Skip known proxy/gateway roles
		if model.NodeHasRole(node, "api_gateway") || model.NodeHasRole(node, "load_balancer") || model.NodeHasRole(node, "reverse_proxy") {
			continue
		}
		// Skip if label implies proxy
		if labelContains(node.Label, "gateway") || labelContains(node.Label, "proxy") || labelContains(node.Label, "lb") {
			continue
		}

		targets := ctx.Outgoing[id]
		hasDataSource := false
		for _, targetID := range targets {
			target, ok := ctx.NodeByID[targetID]
			if !ok {
				continue
			}
			// If it connects to DB, Cache, Storage, or ANOTHER service (which might be the source)
			if model.NodeHasRole(target, "database") || model.NodeHasRole(target, "cache") || model.NodeHasRole(target, "storage") || model.NodeHasRole(target, "service") {
				hasDataSource = true
				break
			}
		}

		if !hasDataSource {
			warnings = append(warnings, Warning{
				Rule:     "missing_data_source",
				Message:  fmt.Sprintf("❓ 缺少資料來源：服務 %q 目前沒有連線至任何資料儲存或下游服務。", node.Label),
				Solution: "每個業務服務應有明確的資料來源。請連接至 Database、Cache、Storage 或其他提供資料的 Service。",
				NodeIDs:  []string{id},
			})
		}
	}
	return warnings
}

// checkDatabasePerService warns if multiple distinct services share the same database node.
// This handles the "shared database" issue (T1/Y1).
func checkDatabasePerService(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	dbToServices := make(map[string][]string)

	for _, edge := range ctx.Edges {
		target, okT := ctx.NodeByID[edge.Target]
		source, okS := ctx.NodeByID[edge.Source]
		if okT && okS && model.NodeHasRole(target, "database") && model.NodeHasRole(source, "service") {
			// Avoid duplicates
			found := false
			for _, s := range dbToServices[edge.Target] {
				if s == edge.Source {
					found = true
					break
				}
			}
			if !found {
				dbToServices[edge.Target] = append(dbToServices[edge.Target], edge.Source)
			}
		}
	}

	for dbID, serviceIDs := range dbToServices {
		if len(serviceIDs) >= 2 {
			dbNode := ctx.NodeByID[dbID]
			var labels []string
			for _, sID := range serviceIDs {
				labels = append(labels, ctx.NodeByID[sID].Label)
			}

			warnings = append(warnings, Warning{
				Rule:     "shared_database",
				Message:  fmt.Sprintf("🗄️ 資料庫共享提醒：多個服務 (%s) 共享同一個資料庫 %q。", joinLabels(labels), dbNode.Label),
				Solution: "在微服務架構中，建議採用 Database-per-Service 模式以解耦服務間的資料依賴與擴展瓶頸。考慮按業務功能拆分資料庫。",
				NodeIDs:  append([]string{dbID}, serviceIDs...),
			})
		}
	}
	return warnings
}
