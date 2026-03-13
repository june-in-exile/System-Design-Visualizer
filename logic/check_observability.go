package logic

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkMissingLogger warns if any Service is NOT connected to a Logger/Monitor.
func checkMissingLogger(ctx model.TopologyContext) []Warning {
	var serviceNodes []model.SystemNode
	var loggerIDs []string

	for _, node := range ctx.Nodes {
		if model.NodeHasRole(node, "service") {
			serviceNodes = append(serviceNodes, node)
		}
		if model.NodeHasRole(node, "logger") || model.NodeHasRole(node, "monitor") {
			loggerIDs = append(loggerIDs, node.ID)
		}
	}

	if len(serviceNodes) == 0 {
		return nil
	}

	// Map of logger IDs for quick lookup
	isLogger := make(map[string]bool)
	for _, id := range loggerIDs {
		isLogger[id] = true
	}

	var disconnectedServices []model.SystemNode
	for _, svc := range serviceNodes {
		connected := false
		targets := ctx.Outgoing[svc.ID]
		for _, targetID := range targets {
			if isLogger[targetID] {
				connected = true
				break
			}
		}
		if !connected {
			disconnectedServices = append(disconnectedServices, svc)
		}
	}

	if len(disconnectedServices) > 0 {
		if len(loggerIDs) == 0 {
			// No logger exists at all
			var nodeIDs []string
			for i, svc := range disconnectedServices {
				if i >= 3 {
					break
				}
				nodeIDs = append(nodeIDs, svc.ID)
			}
			return []Warning{{
				Rule:     "missing_observability",
				Message:  "📊 缺少 Logger/Monitor：架構中存在 Service 但缺少 Logger/Monitor 節點。",
				Solution: "建議加入 Logger/Monitor 節點（如 ELK Stack、Prometheus），並讓各 Service 連線至該節點以收集觀測數據。",
				NodeIDs:  nodeIDs,
			}}
		} else {
			// Logger exists but some services are not connected
			var nodeIDs []string
			var labels []string
			for i, svc := range disconnectedServices {
				nodeIDs = append(nodeIDs, svc.ID)
				if i < 3 {
					labels = append(labels, svc.Label)
				}
			}
			msg := fmt.Sprintf("📊 觀測性缺失：Service %s 等尚未連線至 Logger/Monitor。", joinLabels(labels))
			if len(disconnectedServices) == 1 {
				msg = fmt.Sprintf("📊 觀測性缺失：Service %q 尚未連線至 Logger/Monitor。", disconnectedServices[0].Label)
			}

			return []Warning{{
				Rule:     "incomplete_service_observability",
				Message:  msg,
				Solution: "請將這些 Service 連接至 Logger/Monitor 節點，以確保系統具備完整的監控與除錯能力。",
				NodeIDs:  nodeIDs,
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
