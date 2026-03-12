package model

// SystemTopology is the top-level graph sent from the frontend.
type SystemTopology struct {
	ID      string       `json:"id" binding:"required"`
	Name    string       `json:"name" binding:"required"`
	Version int          `json:"version"`
	Nodes   []SystemNode `json:"nodes" binding:"required"`
	Edges   []SystemEdge `json:"edges"`
}

// Position represents the x/y coordinates of a node on the canvas.
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SystemNode represents a single architecture component.
type SystemNode struct {
	ID            string                 `json:"id" binding:"required"`
	ComponentType string                 `json:"componentType" binding:"required"`
	Label         string                 `json:"label" binding:"required"`
	Position      Position               `json:"position"`
	Properties    map[string]interface{} `json:"properties"`
}

// SystemEdge represents a connection between two nodes.
type SystemEdge struct {
	ID             string `json:"id" binding:"required"`
	Source         string `json:"source" binding:"required"`
	Target         string `json:"target" binding:"required"`
	ConnectionType string `json:"connectionType" binding:"required"`
	Protocol       string `json:"protocol,omitempty"`
	Label          string `json:"label,omitempty"`
}

// ValidComponentTypes enumerates all supported component types.
var ValidComponentTypes = map[string]bool{
	"client":        true,
	"load_balancer": true,
	"api_gateway":   true,
	"service":       true,
	"database":      true,
	"cache":         true,
	"message_queue": true,
	"cdn":           true,
	"dns":           true,
	"storage":       true,
}

// ValidConnectionTypes enumerates all supported connection types.
var ValidConnectionTypes = map[string]bool{
	"sync":        true,
	"async":       true,
	"replication": true,
	"cdn_origin":  true,
	"unspecified": true,
}
