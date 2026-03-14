package logic

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
				Message: fmt.Sprintf("⚖️ SQL Scalability Tradeoff: %q is under high write pressure (read ratio %.0f%%).",
					node.Label, dbProps.ReadWriteRatio*100),
				Solution: "Implement Master-Slave read-write separation, or consider switching to a NoSQL database with better write performance.",
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
				Message:  fmt.Sprintf("📖 Read-Write Separation Suggestion: %q has an extremely high read ratio (%.0f%%).", node.Label, dbProps.ReadWriteRatio*100),
				Solution: "Consider implementing a Master-Slave read-write separation architecture, where the Master handles writes and Replicas handle reads to improve performance.",
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
					Message: fmt.Sprintf("🚀 Performance vs. Consistency Tradeoff: %q (%s) is an AP system, but you requested Strong Consistency.",
						node.Label, dbProps.Product),
					Solution: "Implementing strong consistency (e.g., Quorum R/W) on an AP system will significantly increase latency and reduce availability. Please evaluate if this is necessary.",
					NodeIDs:  []string{node.ID},
				})
				continue
			}

			// Default warning if no explicit choice is made.
			warnings = append(warnings, Warning{
				Rule: "cap_theorem",
				Message: fmt.Sprintf("📐 CAP Theorem: %q (%s) is an AP system.",
					node.Label, dbProps.Product),
				Solution: "Note that AP systems only provide eventual consistency. If strong consistency is required, select 'Strong' in the properties (watch for performance impact) or switch to a CP system (like RDBMS).",
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
			Message: fmt.Sprintf("🔍 Vertical Partitioning (Federation) Reminder: Detected %d database nodes (%s).",
				len(dbNodes), joinLabels(labels)),
			Solution: "Ensure the application layer supports cross-database data aggregation, and be mindful of consistency issues in cross-database transactions.",
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
					Message:  fmt.Sprintf("🔍 Search Engine Suggestion: Search service %q connects directly to %q.", node.Label, target.Label),
					Solution: "Full-text search or complex filtering is less efficient on RDBMS/NoSQL. It is recommended to use specialized search engines (e.g., Elasticsearch, Solr) and ensure data consistency through synchronization mechanisms.",
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
				Message:  fmt.Sprintf("❓ Missing Data Source: Service %q currently has no connection to any data storage or downstream service.", node.Label),
				Solution: "Each business service should have a clear data source. Please connect to a Database, Cache, Storage, or another data-providing Service.",
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
				Message:  fmt.Sprintf("🗄️ Shared Database Reminder: Multiple services (%s) share the same database %q.", joinLabels(labels), dbNode.Label),
				Solution: "In a microservices architecture, it is recommended to adopt a Database-per-Service pattern to decouple data dependencies and scaling bottlenecks between services. Consider splitting the database based on business functionality.",
				NodeIDs:  append([]string{dbID}, serviceIDs...),
			})
		}
	}
	return warnings
}
