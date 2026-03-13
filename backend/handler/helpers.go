package handler

import (
	"strings"

	"github.com/architectmind/backend/model"
)

func labelContains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
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

func hasReplicationEdge(nodeID string, edges []model.SystemEdge) bool {
	for _, edge := range edges {
		if (edge.Source == nodeID || edge.Target == nodeID) && edge.ConnectionType == "replication" {
			return true
		}
	}
	return false
}
