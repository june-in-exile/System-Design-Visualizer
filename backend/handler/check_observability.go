package handler

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkMissingLogger warns if there are 3+ Services but no Logger,
// or if Logger exists but is not connected to any Service.
func checkMissingLogger(ctx model.TopologyContext) []Warning {
	serviceCount := 0
	hasLogger := false
	var serviceIDs []string
	var loggerIDs []string

	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "service") {
			serviceCount++
			serviceIDs = append(serviceIDs, node.ID)
		}
		if model.NodeHasRole(node, "logger") || model.NodeHasRole(node, "monitor") {
			hasLogger = true
			loggerIDs = append(loggerIDs, node.ID)
		}
	}

	if serviceCount >= 3 && !hasLogger {
		if len(serviceIDs) > 3 {
			serviceIDs = serviceIDs[:3]
		}
		return []Warning{{
			Rule:     "missing_observability",
			Message:  "📊 缺少 Logger/Monitor：架構中有 3 個（含）以上的 Service 但缺少 Logger/Monitor。",
			Solution: "建議加入 Logger/Monitor 節點（如 ELK Stack、Prometheus + Grafana、Datadog），並讓各 Service 連線至該節點。",
			NodeIDs:  serviceIDs,
		}}
	}

	if serviceCount >= 3 && hasLogger {
		connected := false
		// Map of logger IDs for quick lookup
		isLogger := make(map[string]bool)
		for _, id := range loggerIDs {
			isLogger[id] = true
		}

		// Check if any service connects to a logger
		for _, serviceID := range serviceIDs {
			targets := ctx.Outgoing[serviceID]
			for _, targetID := range targets {
				if isLogger[targetID] {
					connected = true
					break
				}
			}
			if connected {
				break
			}
		}

		if !connected {
			return []Warning{{
				Rule:     "missing_observability",
				Message:  "📊 Logger/Monitor 未正確連線：架構中有 Logger/Monitor 但未與任何 Service 建立連線。",
				Solution: "建議將各 Service 連接至 Logger/Monitor 節點，以收集日誌與監控數據。",
				NodeIDs:  append(loggerIDs, serviceIDs[:3]...),
			}}
		}
	}

	return nil
}

// checkIncompleteObservability warns if a Logger collects only partial telemetry
// in an architecture with 3+ services.
func checkIncompleteObservability(ctx model.TopologyContext) []Warning {
	serviceCount := 0
	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "service") {
			serviceCount++
		}
	}
	if serviceCount < 3 {
		return nil
	}

	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "logger") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		logProps, ok := props.(*model.LoggerProperties)
		if !ok {
			continue
		}
		if logProps.LogType == "all" || logProps.LogType == "" {
			continue
		}
		if logProps.LogType == "metrics" {
			warnings = append(warnings, Warning{
				Rule:     "incomplete_observability",
				Message:  fmt.Sprintf("📊 觀測性不完整：%q 僅收集 Metrics，缺少 Logs 和 Traces。微服務架構中缺乏 distributed tracing 將難以定位跨服務的效能瓶頸。", node.Label),
				Solution: "建議將 logType 設為 All，或額外加入 Traces 類型的 Logger（如 Jaeger）實現完整觀測性。",
				NodeIDs:  []string{node.ID},
			})
		} else if logProps.LogType == "logs" {
			warnings = append(warnings, Warning{
				Rule:     "incomplete_observability",
				Message:  fmt.Sprintf("📊 觀測性不完整：%q 僅收集 Logs，缺少 Metrics 和 Traces。沒有 Metrics 將無法設定有效的告警規則。", node.Label),
				Solution: "建議將 logType 設為 All，或額外加入 Metrics 類型的 Logger（如 Prometheus）。",
				NodeIDs:  []string{node.ID},
			})
		} else if logProps.LogType == "traces" {
			warnings = append(warnings, Warning{
				Rule:     "incomplete_observability",
				Message:  fmt.Sprintf("📊 觀測性不完整：%q 僅收集 Traces，缺少 Metrics 和 Logs。缺乏 Metrics 和 Logs 將無法進行全面的問題診斷。", node.Label),
				Solution: "建議將 logType 設為 All，或額外加入 Metrics 和 Logs 類型的 Logger。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkAlertingDisabled warns if a Logger node has alerting turned off.
func checkAlertingDisabled(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "logger") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		logProps, ok := props.(*model.LoggerProperties)
		if !ok {
			continue
		}
		if !logProps.Alerting {
			warnings = append(warnings, Warning{
				Rule:     "alerting_disabled",
				Message:  fmt.Sprintf("🔔 告警未啟用：%q 未啟用告警功能。問題發生時將無法及時收到通知，只能依賴人工巡檢。", node.Label),
				Solution: "建議啟用告警並配置通知管道（如 Slack、PagerDuty、Email），設定關鍵指標的閾值觸發條件。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}
