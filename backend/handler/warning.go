package handler

// Warning represents a structured linting result with associated node IDs.
type Warning struct {
	Rule     string   `json:"rule"`
	Message  string   `json:"message"`
	Solution string   `json:"solution"`
	NodeIDs  []string `json:"nodeIds"`
}

// AllRuleNames lists every distinct rule identifier the engine can emit.
var AllRuleNames = []string{
	"schema",
	"spof",
	"db_selection",
	"federation",
	"cache_consistency",
	"cap_theorem",
	"cdn_usage",
	"async_decoupling",
	"lb_spof",
	"read_write_separation",
	"cache_eviction",
	"protocol_mismatch",
	"protocol_connection_mismatch",
	"mq_consumer_missing",
	"mq_dlq_missing",
	"async_peak_shaving",
	"sync_upload_bottleneck",
	"invalid_connection",
	"reverse_proxy_spof",
	"reverse_proxy_ssl",
	"cdn_isolated",
	"long_sync_chain",
	"internal_http",
	"search_engine_recommendation",
	"missing_data_source",
	"shared_database",
	"cache_no_fallback",
	"missing_firewall",
	"missing_observability",
	"incomplete_service_observability",
	"firewall_monitor_mode",
	"firewall_l3_only",
	"incomplete_observability",
	"alerting_disabled",
	"serverless_replicas",
	"no_autoscaling_single",
	"no_healthcheck_behind_lb",
}

// AnalyzeResponse is returned after parsing and validating the topology.
type AnalyzeResponse struct {
	Success    bool      `json:"success"`
	NodeCount  int       `json:"nodeCount"`
	EdgeCount  int       `json:"edgeCount"`
	TotalRules int       `json:"totalRules"`
	RulesPassed int      `json:"rulesPassed"`
	Warnings   []Warning `json:"warnings,omitempty"`
}

// ComputeRulesPassed returns (totalRules, rulesPassed) from a set of warnings.
func ComputeRulesPassed(warnings []Warning) (int, int) {
	total := len(AllRuleNames)
	failed := make(map[string]bool)
	for _, w := range warnings {
		failed[w.Rule] = true
	}
	return total, total - len(failed)
}
