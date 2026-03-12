package model

import "encoding/json"

// DatabaseProperties holds attributes for database nodes.
type DatabaseProperties struct {
	DBType           string  `json:"dbType"`
	Product          string  `json:"product,omitempty"`
	ACIDRequired     bool    `json:"acidRequired"`
	ReadWriteRatio   float64 `json:"readWriteRatio"`
	ScalingStrategy  string  `json:"scalingStrategy"`
	Replicas         int     `json:"replicas,omitempty"`
	ConsistencyLevel string  `json:"consistencyLevel,omitempty"`
}

// APProducts lists database products that follow AP (Availability + Partition tolerance) in CAP.
var APProducts = map[string]bool{
	"cassandra": true,
	"dynamodb":  true,
	"couchdb":   true,
	"riak":      true,
}

// LoadBalancerProperties holds attributes for load balancer nodes.
type LoadBalancerProperties struct {
	Algorithm   string `json:"algorithm"`
	HealthCheck bool   `json:"healthCheck"`
	Layer       string `json:"layer"`
	Replicas    int    `json:"replicas,omitempty"`
}

// CacheProperties holds attributes for cache nodes.
type CacheProperties struct {
	CacheType      string `json:"cacheType"`
	EvictionPolicy string `json:"evictionPolicy"`
	TTLSeconds     int    `json:"ttlSeconds,omitempty"`
	Replicas       int    `json:"replicas,omitempty"`
}

// MessageQueueProperties holds attributes for message queue nodes.
type MessageQueueProperties struct {
	Category          string `json:"category"`
	Product           string `json:"product,omitempty"`
	QueueType         string `json:"queueType"`
	DeliveryGuarantee string `json:"deliveryGuarantee"`
	Ordered           bool   `json:"ordered"`
	HasDLQ            bool   `json:"hasDLQ"`
}

// StorageProperties holds attributes for object storage nodes.
type StorageProperties struct {
	AccessLevel  string `json:"accessLevel"`
	StorageClass string `json:"storageClass"`
	Versioning   bool   `json:"versioning"`
}

// ServiceProperties holds attributes for service nodes.
type ServiceProperties struct {
	Replicas  int  `json:"replicas"`
	Stateless bool `json:"stateless"`
}

// ReverseProxyProperties holds attributes for reverse proxy nodes.
type ReverseProxyProperties struct {
	Product        string `json:"product,omitempty"`
	SSLTermination bool   `json:"sslTermination"`
	Caching        bool   `json:"caching"`
	Compression    bool   `json:"compression"`
	RateLimiting   bool   `json:"rateLimiting"`
	Replicas       int    `json:"replicas,omitempty"`
}

// ParseNodeProperties converts the generic Properties map into a typed struct
// based on the node's ComponentType. Returns the original map for unknown types.
func ParseNodeProperties(node SystemNode) (interface{}, error) {
	raw, err := json.Marshal(node.Properties)
	if err != nil {
		return nil, err
	}

	switch node.ComponentType {
	case "database":
		var props DatabaseProperties
		return &props, json.Unmarshal(raw, &props)
	case "load_balancer":
		var props LoadBalancerProperties
		return &props, json.Unmarshal(raw, &props)
	case "cache":
		var props CacheProperties
		return &props, json.Unmarshal(raw, &props)
	case "message_queue":
		var props MessageQueueProperties
		return &props, json.Unmarshal(raw, &props)
	case "storage":
		var props StorageProperties
		return &props, json.Unmarshal(raw, &props)
	case "service":
		var props ServiceProperties
		return &props, json.Unmarshal(raw, &props)
	case "reverse_proxy":
		var props ReverseProxyProperties
		return &props, json.Unmarshal(raw, &props)
	default:
		return node.Properties, nil
	}
}

// GetEffectiveRoles returns the roles a node fulfills. If Roles is set, it is
// returned directly; otherwise a single-element slice of ComponentType is returned.
func GetEffectiveRoles(node SystemNode) []string {
	if len(node.Roles) > 0 {
		return node.Roles
	}
	return []string{node.ComponentType}
}

// NodeHasRole reports whether node fulfills the given role.
func NodeHasRole(node SystemNode, role string) bool {
	for _, r := range GetEffectiveRoles(node) {
		if r == role {
			return true
		}
	}
	return false
}
