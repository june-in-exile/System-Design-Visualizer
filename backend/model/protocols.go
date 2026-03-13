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

// ForbiddenConnections defines source→target role pairs that are architecturally invalid.
// Key = source role, Value = list of forbidden target roles.
var ForbiddenConnections = map[string][]string{
	// 入口節點 → CDN（流量已進入基礎設施，不應繞回 CDN）
	"firewall": {"cdn"},

	// 入口節點 → 資料層（跳過 Service 層，違反分層架構）
	"load_balancer": {"cdn", "database", "cache", "message_queue"},
	"reverse_proxy": {"cdn", "database", "cache", "message_queue"},
	"api_gateway":   {"cdn", "database", "cache", "message_queue"},

	// 資料層 → 入口層/Client（資料層是被動的，不主動連線）
	"database": {"client", "load_balancer", "reverse_proxy", "api_gateway", "firewall", "cdn"},
	"cache":    {"client", "load_balancer", "reverse_proxy", "api_gateway", "firewall", "cdn"},
	"storage":  {"client"},

	// MQ → Client（Client 不應直接消費 MQ）
	"message_queue": {"client"},

	// CDN → 資料層（CDN 只應 origin pull 到 Storage 或入口節點）
	"cdn": {"database", "cache", "message_queue"},

	// DNS 只對 Client 有意義
	"dns": {"service", "database", "cache", "message_queue", "storage",
		"load_balancer", "reverse_proxy", "api_gateway", "firewall", "cdn", "logger"},

	// Logger/Monitor 是被動接收端，不應有 outgoing 連線
	"logger": {"service", "database", "cache", "message_queue", "storage",
		"client", "cdn", "load_balancer", "reverse_proxy", "api_gateway", "firewall", "dns"},

	// Client 不應直連資料層 (原 client_direct_db/cache)
	"client": {"database", "cache"},
}

// ForbiddenConnectionReasons provides human-readable explanations for each category.
var ForbiddenConnectionReasons = map[string]map[string]string{
	"firewall":      {"cdn": "流量已進入基礎設施後不應繞回 CDN，CDN 應位於 Firewall 前方"},
	"load_balancer": {"cdn": "同上", "database": "入口節點不應跳過 Service 層直接存取資料庫", "cache": "入口節點不應跳過 Service 層直接存取快取", "message_queue": "入口節點不應跳過 Service 層直接存取 MQ"},
	"reverse_proxy": {"cdn": "同上", "database": "入口節點不應跳過 Service 層直接存取資料庫", "cache": "入口節點不應跳過 Service 層直接存取快取", "message_queue": "入口節點不應跳過 Service 層直接存取 MQ"},
	"api_gateway":   {"cdn": "同上", "database": "入口節點不應跳過 Service 層直接存取資料庫", "cache": "入口節點不應跳過 Service 層直接存取快取", "message_queue": "入口節點不應跳過 Service 層直接存取 MQ"},
	"database":      {"*": "資料庫是被動元件，不應主動連線至入口層或 Client"},
	"cache":         {"*": "快取是被動元件，不應主動連線至入口層或 Client"},
	"storage":       {"client": "Object Storage 不應直連 Client，應透過 CDN 或 Service 分發"},
	"message_queue": {"client": "Client 不應直接消費 MQ，應由 Service 處理後推送"},
	"cdn":           {"*": "CDN 只應 origin pull 至 Storage 或入口節點，不應存取資料層"},
	"dns":           {"*": "DNS 只做域名解析，不路由實際流量至後端元件"},
	"logger":        {"*": "Logger/Monitor 是被動接收端（sink），不應主動連線至其他元件"},
	"client": {
		"database": "Client 不應直接操作資料庫。請在兩者之間加入 API Gateway 或 Service 層進行身份驗證與數據抽象。",
		"cache":    "不建議 Client 直接操作快取。這可能導致快取穿透風險或數據洩漏。應透過後端 Service 進行快取邏輯封裝。",
	},
}
