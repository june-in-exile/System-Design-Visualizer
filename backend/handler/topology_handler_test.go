package handler

import (
	"testing"
	"github.com/architectmind/backend/model"
)

func TestCheckProtocolMismatch_Unspecified(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"1": {ID: "1", ComponentType: "database", Label: "Master"},
		"2": {ID: "2", ComponentType: "database", Label: "Slave"},
	}
	edges := []model.SystemEdge{
		{
			ID:             "e1",
			Source:         "1",
			Target:         "2",
			ConnectionType: "replication",
			Protocol:       "",
		},
	}

	warnings := checkProtocolMismatch(nodes, edges)
	if len(warnings) > 0 {
		t.Errorf("Expected no warnings for unspecified protocol, got %v", warnings)
	}
}

func TestCheckProtocolMismatch_Mismatch(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"1": {ID: "1", ComponentType: "service", Label: "Service"},
		"2": {ID: "2", ComponentType: "database", Label: "Database"},
	}
	edges := []model.SystemEdge{
		{
			ID:             "e1",
			Source:         "1",
			Target:         "2",
			ConnectionType: "sync",
			Protocol:       "http",
		},
	}

	warnings := checkProtocolMismatch(nodes, edges)
	if len(warnings) == 0 {
		t.Error("Expected warning for HTTP protocol on database, got none")
	}
}

func TestCheckConnectionTypeProtocolMismatch_Unspecified(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"1": {ID: "1", ComponentType: "database", Label: "Master"},
		"2": {ID: "2", ComponentType: "database", Label: "Slave"},
	}
	edges := []model.SystemEdge{
		{
			ID:             "e1",
			Source:         "1",
			Target:         "2",
			ConnectionType: "replication",
			Protocol:       "",
		},
	}

	warnings := checkConnectionTypeProtocolMismatch(nodes, edges)
	if len(warnings) > 0 {
		t.Errorf("Expected no warnings for replication with unspecified protocol, got %v", warnings)
	}
}
