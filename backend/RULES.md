# Rule Index (架構驗證規則索引)

此文件列出了 ArchitectMind 後端驗證引擎支援的所有規則，按類別分組。

| Rule ID | 實作檔案 | 說明 |
|---------|---------|------|
| **Schema** | | |
| `schema` | `topology_handler.go` | 驗證節點組件類型、連線類型及連線端點的有效性。 |
| **Availability (可用性)** | | |
| `spof` | `check_availability.go` | 偵測 Load Balancer/Reverse Proxy 後方的單點故障 (SPOF)。 |
| `lb_spof` | `check_availability.go` | 檢查架構中是否僅存在單個 Load Balancer 實例。 |
| `reverse_proxy_spof` | `check_availability.go` | 檢查架構中是否僅存在單個 Reverse Proxy 實例。 |
| `no_autoscaling_single` | `check_availability.go` | 警告服務僅有單個複本且未啟用自動擴縮。 |
| `no_healthcheck_behind_lb` | `check_availability.go` | 檢查 Load Balancer 後方的服務是否啟用了健康檢查。 |
| `serverless_replicas` | `check_availability.go` | 提醒 Serverless 服務不需要手動設定複本數。 |
| **Database & Performance (資料庫與效能)** | | |
| `db_selection` | `check_database.go` | 針對高寫入壓力的 SQL 資料庫提供擴展性建議。 |
| `read_write_separation` | `check_database.go` | 當讀取比例極高時，建議導入主從讀寫分離。 |
| `cap_theorem` | `check_database.go` | 根據 CAP 定理提醒 AP 系統的一致性權衡。 |
| `federation` | `check_database.go` | 偵測多資料庫節點的垂直拆分 (Federation) 模式。 |
| `search_engine_recommendation` | `check_database.go` | 建議針對搜尋需求使用專門的搜尋引擎而非直接查詢 DB。 |
| `missing_data_source` | `check_database.go` | 檢查業務服務是否漏連了資料來源或下游服務。 |
| `shared_database` | `check_database.go` | 提醒多個微服務不應共享同一個資料庫實例。 |
| **Cache (快取)** | | |
| `cache_consistency` | `check_cache.go` | 檢查同時連線 Cache 與 DB 時的快取一致性與 TTL 設定。 |
| `cache_eviction` | `check_cache.go` | 檢查快取節點是否配置了失效策略 (Eviction Policy)。 |
| `cache_no_fallback` | `check_cache.go` | 警告 Service 僅連線至 Cache 而無後端資料庫回退路徑。 |
| **Network & CDN (網路與 CDN)** | | |
| `cdn_usage` | `check_network.go` | 根據 Client 存在與否建議使用 CDN 全球加速。 |
| `cdn_isolated` | `check_network.go` | 偵測 CDN 節點是否遺漏了與源站 (Origin) 的連線。 |
| `protocol_mismatch` | `check_network.go` | 檢查連線協議與目標組件類型是否匹配 (如 HTTP 連資料庫)。 |
| `protocol_connection_mismatch` | `check_network.go` | 檢查連線類型 (Sync/Async) 與協議是否邏輯一致。 |
| `reverse_proxy_ssl` | `check_network.go` | 建議在接收 HTTPS 流量的 Reverse Proxy 上啟用 SSL 終止。 |
| `long_sync_chain` | `check_network.go` | 偵測過長的服務同步呼叫鏈 (深度 >= 3)。 |
| `internal_http` | `check_network.go` | 建議服務內部通訊使用 gRPC 取代純 HTTP。 |
| **Messaging (訊息隊列)** | | |
| `async_decoupling` | `check_messaging.go` | 針對耗時操作建議使用 MQ 進行異步解耦。 |
| `mq_consumer_missing` | `check_messaging.go` | 偵測 MQ 節點是否缺乏下游消費者連線。 |
| `mq_dlq_missing` | `check_messaging.go` | 提醒 MQ 節點應配置死信隊列 (DLQ) 與重試機制。 |
| `async_peak_shaving` | `check_messaging.go` | 建議針對高負載寫入資料庫使用 MQ 進行削峰填谷。 |
| `sync_upload_bottleneck` | `check_messaging.go` | 警告大檔案同步上傳至 Storage 可能造成的效能瓶頸。 |
| **Security (安全)** | | |
| `invalid_connection` | `check_security.go` | 偵測不合理的元件連線方向 (如 DB→Client, LB→DB 等)。 |
| `missing_firewall` | `check_security.go` | 檢查 Client 與入口節點之間是否配置了 Firewall/WAF。 |
| `firewall_monitor_mode` | `check_security.go` | 提醒 Firewall 目前處於不攔截的監控模式。 |
| `firewall_l3_only` | `check_security.go` | 警告 L3 防火牆無法防禦應用層 (L7) 攻擊。 |
| **Observability (觀測性)** | | |
| `missing_observability` | `check_observability.go` | 檢查架構中是否完全缺失 Logger/Monitor 節點。 |
| `incomplete_service_observability` | `check_observability.go` | 偵測部分 Service 尚未連線至監控系統。 |
| `incomplete_observability` | `check_observability.go` | 警告 Logger 收集的遙測數據類型不完整 (缺 Logs/Metrics/Traces)。 |
| `alerting_disabled` | `check_observability.go` | 提醒 Logger 節點未啟用告警與通知功能。 |
