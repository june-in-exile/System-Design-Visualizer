package handler

import (
	"fmt"
	"github.com/architectmind/backend/model"
)

// checkAsyncDecoupling suggests Message Queues for time-consuming operations.
func checkAsyncDecoupling(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	keywords := []string{
		"mail", "img", "image", "photo", "pic",
		"vid", "video", "media", "stream", "transcode",
		"report", "pdf", "export", "csv", "excel",
		"task", "worker", "job", "batch", "process",
		"notify", "sms", "push", "upload", "download",
		"ai", "ml", "inference", "prediction", "training",
		"payment", "billing", "checkout", "payout",
		"index", "search", "crawl", "scrape",
		"sync", "migration", "archive", "cleanup",
		"audit", "analytics", "stats",
		"webhook", "slack", "discord", "telegram",
	}

	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "service") {
			continue
		}

		isTimeConsuming := false
		for _, kw := range keywords {
			if labelContains(node.Label, kw) {
				isTimeConsuming = true
				break
			}
		}

		if isTimeConsuming {
			// Find who is calling this service
			var synchronousCallers []string
			isDecoupled := false

			for _, edge := range ctx.Edges {
				if edge.Target == id {
					sourceNode := ctx.NodeByID[edge.Source]
					// If any caller is a Message Queue, or the connection itself is async, we consider it decoupled!
					if model.NodeHasRole(sourceNode, "message_queue") || edge.ConnectionType == "async" {
						isDecoupled = true
						break
					}
					synchronousCallers = append(synchronousCallers, edge.Source)
				}
			}

			// If called synchronously and not through an MQ, issue warning
			if isTimeConsuming && !isDecoupled && len(synchronousCallers) > 0 {
				warnings = append(warnings, Warning{
					Rule:     "async_decoupling",
					Message:  fmt.Sprintf("📬 異步解耦提醒：服務 %q 似乎涉及耗時操作且目前接收同步呼叫。", node.Label),
					Solution: "建議在呼叫方與此服務之間加入 Message Queue (MQ)，將操作改為異步處理，以提高系統吞吐量。",
					NodeIDs:  append([]string{id}, synchronousCallers...),
				})
			}
		}
	}
	return warnings
}

// checkMQConsumer warns if a Message Queue node has no outgoing connections.
func checkMQConsumer(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		if !model.NodeHasRole(node, "message_queue") {
			continue
		}
		if len(ctx.Outgoing[id]) == 0 {
			warnings = append(warnings, Warning{
				Rule:     "mq_consumer_missing",
				Message:  fmt.Sprintf("📥 MQ 消費者缺失檢查：%q 目前沒有任何消費者 (Consumer)。", node.Label),
				Solution: "Message Queue 節點缺乏輸出連線，訊息將在隊列中堆積。請將此節點連接至處理訊息的 Service。",
				NodeIDs:  []string{id},
			})
		}
	}
	return warnings
}

// checkMQDLQ warns if a Message Queue node does not have a dead letter queue configured.
func checkMQDLQ(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, node := range ctx.Nodes {
		if !model.NodeHasRole(node, "message_queue") {
			continue
		}
		props, err := model.ParseNodeProperties(node)
		if err != nil {
			continue
		}
		mqProps, ok := props.(*model.MessageQueueProperties)
		if !ok {
			continue
		}
		if !mqProps.HasDLQ {
			warnings = append(warnings, Warning{
				Rule:     "mq_dlq_missing",
				Message:  fmt.Sprintf("💀 死信隊列 (DLQ) 提醒：%q 未配置死信隊列或重試機制。", node.Label),
				Solution: "使用 Message Queue home未配置死信隊列 (Dead Letter Queue)，可能導致處理失敗的訊息直接丟失。建議在屬性中啟用 DLQ。",
				NodeIDs:  []string{node.ID},
			})
		}
	}
	return warnings
}

// checkAsyncPeakShaving suggests using MQ for high-load direct database writes from entry points.
func checkAsyncPeakShaving(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for id, node := range ctx.NodeByID {
		// Entry points: LB, API Gateway, or Reverse Proxy
		if !model.NodeHasRole(node, "load_balancer") && !model.NodeHasRole(node, "api_gateway") && !model.NodeHasRole(node, "reverse_proxy") {
			continue
		}

		targets := ctx.Outgoing[id]
		for _, targetID := range targets {
			target, ok := ctx.NodeByID[targetID]
			if !ok || !model.NodeHasRole(target, "database") {
				continue
			}

			// Check for high write ratio
			props, err := model.ParseNodeProperties(target)
			if err != nil {
				continue
			}
			dbProps, ok := props.(*model.DatabaseProperties)
			if !ok {
				continue
			}

			if dbProps.ReadWriteRatio < 0.5 {
				warnings = append(warnings, Warning{
					Rule:     "async_peak_shaving",
					Message:  fmt.Sprintf("🌊 異步削峰實踐建議：流量入口 %q 直接連線至高負載寫入資料庫 %q。", node.Label, target.Label),
					Solution: "高頻寫入建議先將請求發送至 Message Queue (MQ) 進行削峰填谷 (Load Leveling)，再由後端異步寫入資料庫，以減輕資料庫壓力並提升系統穩定性。",
					NodeIDs:  []string{id, targetID},
				})
			}
		}
	}
	return warnings
}

// checkSyncUpload warns if a service labeled "upload" or "media" connects synchronously to storage.
// This handles the "sync upload" issue (T5).
func checkSyncUpload(ctx model.TopologyContext) []Warning {
	var warnings []Warning
	for _, edge := range ctx.Edges {
		if edge.ConnectionType != "sync" {
			continue
		}
		source, okS := ctx.NodeByID[edge.Source]
		target, okT := ctx.NodeByID[edge.Target]
		if okS && okT && model.NodeHasRole(target, "storage") {
			if labelContains(source.Label, "upload") || labelContains(source.Label, "media") || labelContains(source.Label, "transcode") {
				warnings = append(warnings, Warning{
					Rule:     "sync_upload_bottleneck",
					Message:  fmt.Sprintf("📤 同步上傳瓶頸：服務 %q 同步上傳至 Storage。", source.Label),
					Solution: "大檔案上傳建議改為異步流程：Service 產生 Presigned URL 給 Client 直傳 Storage，完成後再透過 MQ 通知 Service 處理後續。",
					NodeIDs:  []string{edge.Source, edge.Target},
				})
			}
		}
	}
	return warnings
}
