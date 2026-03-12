package handler

import (
	"fmt"
	"net/http"

	"github.com/architectmind/backend/model"
	"github.com/gin-gonic/gin"
)

// AnalyzeResponse is returned after parsing and validating the topology.
type AnalyzeResponse struct {
	Success  bool     `json:"success"`
	NodeCount int     `json:"nodeCount"`
	EdgeCount int     `json:"edgeCount"`
	Warnings []string `json:"warnings,omitempty"`
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

// validate performs basic semantic checks on the topology.
func validate(t model.SystemTopology) []string {
	var warnings []string

	nodeIDs := make(map[string]bool, len(t.Nodes))
	for _, node := range t.Nodes {
		if !model.ValidComponentTypes[node.ComponentType] {
			warnings = append(warnings, fmt.Sprintf("unknown component type %q on node %q", node.ComponentType, node.ID))
		}
		nodeIDs[node.ID] = true
	}

	for _, edge := range t.Edges {
		if !model.ValidConnectionTypes[edge.ConnectionType] {
			warnings = append(warnings, fmt.Sprintf("unknown connection type %q on edge %q", edge.ConnectionType, edge.ID))
		}
		if !nodeIDs[edge.Source] {
			warnings = append(warnings, fmt.Sprintf("edge %q references unknown source node %q", edge.ID, edge.Source))
		}
		if !nodeIDs[edge.Target] {
			warnings = append(warnings, fmt.Sprintf("edge %q references unknown target node %q", edge.ID, edge.Target))
		}
	}

	return warnings
}
