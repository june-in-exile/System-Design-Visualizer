package model

// ExpectedProtocols defines which protocols are appropriate for connections
// targeting a given component type. A nil value means any protocol is acceptable.
var ExpectedProtocols = map[string]map[string]bool{
	"database": {
		"database": true,
		"resp":     true,
		"binary":   true,
		"uds":      true,
		"tcp":      true,
	},
	"cache": {
		"database": true,
		"resp":     true,
		"binary":   true,
		"uds":      true,
		"tcp":      true,
		"http":     true, // Couchbase/Elasticsearch
	},
	"message_queue": {
		"amqp": true,
		"mqtt": true,
		"tcp":  true,
	},
	"dns": {
		"dns": true,
		"udp": true,
	},
	"client": {
		"http":      true,
		"https":     true,
		"websocket": true,
	},
	"cdn": {
		"http":  true,
		"https": true,
	},
	"storage": {
		"http":  true,
		"https": true,
	},
}

// ProtocolDisplayName maps protocol values to human-readable names.
var ProtocolDisplayName = map[string]string{
	"http":      "HTTP",
	"https":     "HTTPS",
	"grpc":      "gRPC",
	"websocket": "WebSocket",
	"ssh":       "SSH",
	"tcp":       "TCP",
	"udp":       "UDP",
	"amqp":      "AMQP",
	"mqtt":      "MQTT",
	"database":  "Database Protocol",
	"resp":      "RESP (Redis)",
	"binary":    "Binary Protocol",
	"uds":       "UDS (Unix Domain Socket)",
	"dns":       "DNS",
}

// ValidConnectionProtocolPairs defines which protocols are allowed for each connection type.
var ValidConnectionProtocolPairs = map[string]map[string]bool{
	"sync": {
		"unspecified": true,
		"http":        true,
		"https":       true,
		"grpc":        true,
		"websocket":   true,
		"ssh":         true,
		"tcp":         true,
		"database":    true,
		"resp":        true,
		"binary":      true,
		"uds":         true,
		"dns":         true,
	},
	"async": {
		"unspecified": true,
		"amqp":        true,
		"mqtt":        true,
		"http":        true, // Webhooks
		"https":       true, // Webhooks
		"tcp":         true,
		"udp":         true,
	},
	"replication": {
		"unspecified": true,
		"database":    true,
		"binary":      true,
		"tcp":         true,
		"udp":         true,
		"ssh":         true, // rsync over ssh
	},
	"cdn_origin": {
		"unspecified": true,
		"http":        true,
		"https":       true,
	},
}
