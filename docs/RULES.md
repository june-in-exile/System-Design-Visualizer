# Rule Index

This document lists all rules supported by the ArchitectMind backend validation engine, grouped by category.

| Rule ID | Implementation File | Description |
|---------|---------------------|-------------|
| **Schema** | | |
| `schema` | `topology_handler.go` | Validates node types, connection types, and the validity of connection endpoints. |
| **Availability** | | |
| `spof` | `check_availability.go` | Detects Single Point of Failure (SPOF) behind Load Balancer/Reverse Proxy. |
| `lb_spof` | `check_availability.go` | Checks if there is only a single instance of Load Balancer in the architecture. |
| `reverse_proxy_spof` | `check_availability.go` | Checks if there is only a single instance of Reverse Proxy in the architecture. |
| `no_autoscaling_single` | `check_availability.go` | Warns if a service has only one replica and no auto-scaling enabled. |
| `no_healthcheck_behind_lb` | `check_availability.go` | Checks if services behind Load Balancer have health checks enabled. |
| `serverless_replicas` | `check_availability.go` | Reminds that Serverless services do not need manual replica configuration. |
| **Database & Performance** | | |
| `db_selection` | `check_database.go` | Provides scalability suggestions for SQL databases under high write pressure. |
| `read_write_separation` | `check_database.go` | Suggests read-write separation when read ratio is extremely high. |
| `cap_theorem` | `check_database.go` | Reminds about consistency tradeoffs for AP systems based on CAP theorem. |
| `federation` | `check_database.go` | Detects vertical partitioning (Federation) patterns for multi-database nodes. |
| `search_engine_recommendation` | `check_database.go` | Suggests using specialized search engines instead of direct DB queries for search requirements. |
| `missing_data_source` | `check_database.go` | Checks if business services are missing connections to data sources or downstream services. |
| `shared_database` | `check_database.go` | Warns that multiple microservices should not share the same database instance. |
| **Cache** | | |
| `cache_consistency` | `check_cache.go` | Checks cache consistency and TTL settings when both Cache and DB are connected. |
| `cache_eviction` | `check_cache.go` | Checks if cache nodes have an eviction policy configured. |
| `cache_no_fallback` | `check_cache.go` | Warns if a service only connects to Cache without a backend database fallback path. |
| **Network & CDN** | | |
| `cdn_usage` | `check_network.go` | Suggests using CDN for global acceleration based on the presence of clients. |
| `cdn_isolated` | `check_network.go` | Detects if CDN nodes are missing connections to origin servers. |
| `protocol_mismatch` | `check_network.go` | Checks if connection protocol matches the target component type (e.g., HTTP for DB). |
| `protocol_connection_mismatch` | `check_network.go` | Checks if connection type (Sync/Async) is logically consistent with the protocol. |
| `reverse_proxy_ssl` | `check_network.go` | Suggests enabling SSL termination on Reverse Proxy receiving HTTPS traffic. |
| `long_sync_chain` | `check_network.go` | Detects overly long synchronous service call chains (depth >= 3). |
| `internal_http` | `check_network.go` | Suggests using gRPC instead of plain HTTP for internal service communication. |
| **Messaging** | | |
| `async_decoupling` | `check_messaging.go` | Suggests using MQ for asynchronous decoupling of time-consuming operations. |
| `mq_consumer_missing` | `check_messaging.go` | Detects if MQ nodes lack downstream consumer connections. |
| `mq_dlq_missing` | `check_messaging.go` | Reminds that MQ nodes should configure Dead Letter Queues (DLQ) and retry mechanisms. |
| `async_peak_shaving` | `check_messaging.go` | Suggests using MQ for peak shaving for databases under high load. |
| `sync_upload_bottleneck` | `check_messaging.go` | Warns about performance bottlenecks when synchronously uploading large files to storage. |
| **Security** | | |
| `invalid_connection` | `check_security.go` | Detects unreasonable component connection directions (e.g., DB→Client, LB→DB, etc.). |
| `missing_firewall` | `check_security.go` | Checks if a Firewall/WAF is configured between Client and entry nodes. |
| `firewall_monitor_mode` | `check_security.go` | Reminds that the Firewall is currently in non-blocking monitor mode. |
| `firewall_l3_only` | `check_security.go` | Warns that L3 firewalls cannot defend against application-layer (L7) attacks. |
| **Observability** | | |
| `missing_observability` | `check_observability.go` | Checks if the architecture is completely missing Logger/Monitor nodes. |
| `incomplete_service_observability` | `check_observability.go` | Detects services not yet connected to the monitoring system. |
| `incomplete_observability` | `check_observability.go` | Warns that telemetry data collected by Logger is incomplete (missing Logs/Metrics/Traces). |
| `alerting_disabled` | `check_observability.go` | Reminds that Logger nodes have not enabled alerting and notification features. |
| **Capacity Planning** | | |
| | | *The following rules are triggered only if System Parameters (DAU, QPS, etc.) are set in the toolbar.* |
| `high_qps_no_cache` | `check_capacity.go` | Suggests adding a cache layer like Redis when Peak QPS > 5,000 and no cache nodes exist. |
| `high_qps_single_lb` | `check_capacity.go` | Suggests increasing LB replicas when Peak QPS > 10,000 and there is only 1 LB replica. |
| `high_qps_no_autoscaling` | `check_capacity.go` | Suggests enabling auto-scaling when Peak QPS > 5,000 and Service auto-scaling is disabled. |
| `high_dau_no_cdn` | `check_capacity.go` | Suggests adding CDN when DAU > 100,000 and no CDN nodes exist in the architecture. |
| `storage_growth_no_partitioning` | `check_capacity.go` | Suggests Sharding when daily data growth > 10 GB and no database horizontal scaling strategy is set. |
| `high_availability_insufficient_replicas` | `check_capacity.go` | Higher redundancy is needed when availability target is 99.99%+ but service replicas < 3. |
| `latency_long_sync_chain` | `check_capacity.go` | Suggests making parts of the chain asynchronous when latency target p99 < 100ms and sync chain depth > 3. |
| `read_heavy_no_read_replica` | `check_capacity.go` | Suggests setting up Read Replicas when read/write ratio > 80% and there is only a single DB node. |
