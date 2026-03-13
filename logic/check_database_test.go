package logic

import (
	"testing"

	"github.com/architectmind/backend/model"
)

// --- checkDBSelection ---

func TestCheckDBSelection_HighWriteSQL(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "MySQL", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.3,
		}},
	}
	edges := []model.SystemEdge{}
	w := checkDBSelection(makeCtx(nodes, edges))
	if len(w) != 1 || w[0].Rule != "db_selection" {
		t.Errorf("expected 1 db_selection warning, got %d", len(w))
	}
}

func TestCheckDBSelection_HighReadSQL_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "MySQL", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.8,
		}},
	}
	w := checkDBSelection(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for high-read SQL, got %d", len(w))
	}
}

func TestCheckDBSelection_SuppressedByReplication(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "MySQL", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.3,
		}},
		"db2": {ID: "db2", ComponentType: "database", Label: "Slave"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "db1", Target: "db2", ConnectionType: "replication"},
	}
	w := checkDBSelection(makeCtx(nodes, edges))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings when replication exists, got %d", len(w))
	}
}

// --- checkReadWriteSeparation ---

func TestCheckReadWriteSeparation_HighRead(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "PG", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.9,
		}},
	}
	w := checkReadWriteSeparation(makeCtx(nodes, nil))
	if len(w) != 1 || w[0].Rule != "read_write_separation" {
		t.Errorf("expected 1 warning, got %d", len(w))
	}
}

func TestCheckReadWriteSeparation_SuppressedByReplication(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "PG", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.95,
		}},
		"db2": {ID: "db2", ComponentType: "database", Label: "Slave"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "db1", Target: "db2", ConnectionType: "replication"},
	}
	w := checkReadWriteSeparation(makeCtx(nodes, edges))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings with replication, got %d", len(w))
	}
}

func TestCheckReadWriteSeparation_LowRead_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "PG", Properties: map[string]interface{}{
			"dbType": "sql", "readWriteRatio": 0.5,
		}},
	}
	w := checkReadWriteSeparation(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings, got %d", len(w))
	}
}

// --- checkCAP ---

func TestCheckCAP_APSystemNoChoice(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "Cassandra", Properties: map[string]interface{}{
			"product": "cassandra",
		}},
	}
	w := checkCAP(makeCtx(nodes, nil))
	if len(w) != 1 || w[0].Rule != "cap_theorem" {
		t.Errorf("expected 1 cap warning, got %d", len(w))
	}
}

func TestCheckCAP_APSystemEventualConsistency_Suppressed(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "Cassandra", Properties: map[string]interface{}{
			"product": "cassandra", "consistencyLevel": "eventual",
		}},
	}
	w := checkCAP(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings when eventual consistency chosen, got %d", len(w))
	}
}

func TestCheckCAP_APSystemStrongConsistency(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "Cassandra", Properties: map[string]interface{}{
			"product": "cassandra", "consistencyLevel": "strong",
		}},
	}
	w := checkCAP(makeCtx(nodes, nil))
	if len(w) != 1 {
		t.Errorf("expected 1 warning for strong consistency on AP, got %d", len(w))
	}
}

func TestCheckCAP_NonAPProduct_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "PostgreSQL", Properties: map[string]interface{}{
			"product": "postgresql",
		}},
	}
	w := checkCAP(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for non-AP product, got %d", len(w))
	}
}

// --- checkVerticalPartitioning ---

func TestCheckVerticalPartitioning_MultiDB_NoAggregator(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "UserDB"},
		"db2": {ID: "db2", ComponentType: "database", Label: "OrderDB"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "UserSvc"},
	}
	edges := []model.SystemEdge{{ID: "e1", Source: "s1", Target: "db1"}}
	w := checkVerticalPartitioning(makeCtx(nodes, edges))
	if len(w) != 1 || w[0].Rule != "federation" {
		t.Errorf("expected 1 federation warning, got %d", len(w))
	}
}

func TestCheckVerticalPartitioning_MultiDB_WithAggregator(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "UserDB"},
		"db2": {ID: "db2", ComponentType: "database", Label: "OrderDB"},
		"s1":  {ID: "s1", ComponentType: "service", Label: "Aggregator"},
	}
	edges := []model.SystemEdge{
		{ID: "e1", Source: "s1", Target: "db1"},
		{ID: "e2", Source: "s1", Target: "db2"},
	}
	w := checkVerticalPartitioning(makeCtx(nodes, edges))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings with aggregator, got %d", len(w))
	}
}

func TestCheckVerticalPartitioning_SingleDB_NoWarning(t *testing.T) {
	nodes := map[string]model.SystemNode{
		"db1": {ID: "db1", ComponentType: "database", Label: "DB"},
	}
	w := checkVerticalPartitioning(makeCtx(nodes, nil))
	if len(w) != 0 {
		t.Errorf("expected 0 warnings for single DB, got %d", len(w))
	}
}

