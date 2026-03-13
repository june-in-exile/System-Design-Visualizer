# Backend Refactor 操作手冊

> 目標：提升可讀性、消除重複代碼、縮小 codebase。不改變任何功能行為。
> 每完成一個步驟，跑 `go test ./...` 確認無 regression。

---

## 步驟 1：建立 `TopologyContext` 統一參數

**位置：** `model/topology.go`

在檔案末尾新增：

```go
// TopologyContext holds pre-computed lookup structures for validation.
type TopologyContext struct {
    Nodes    []SystemNode
    Edges    []SystemEdge
    NodeByID map[string]SystemNode
    Outgoing map[string][]string // source ID -> list of target IDs
}

// NewTopologyContext builds a TopologyContext from a SystemTopology.
func NewTopologyContext(t SystemTopology) TopologyContext {
    nodeByID := make(map[string]SystemNode, len(t.Nodes))
    for _, node := range t.Nodes {
        nodeByID[node.ID] = node
    }
    outgoing := make(map[string][]string)
    for _, edge := range t.Edges {
        outgoing[edge.Source] = append(outgoing[edge.Source], edge.Target)
    }
    return TopologyContext{
        Nodes:    t.Nodes,
        Edges:    t.Edges,
        NodeByID: nodeByID,
        Outgoing: outgoing,
    }
}
```

然後將所有 check 函數的簽名統一改為接收 `model.TopologyContext`：

```go
// 改前（各種不同簽名）
func checkSPOF(nodes map[string]model.SystemNode, outgoing map[string][]string) []Warning
func checkDBSelection(nodes []model.SystemNode, edges []model.SystemEdge) []Warning
func checkProtocolMismatch(nodes map[string]model.SystemNode, edges []model.SystemEdge) []Warning

// 改後（統一）
func checkSPOF(ctx model.TopologyContext) []Warning
func checkDBSelection(ctx model.TopologyContext) []Warning
func checkProtocolMismatch(ctx model.TopologyContext) []Warning
```

`validate()` 改為：

```go
func validate(t model.SystemTopology) []Warning {
    ctx := model.NewTopologyContext(t)
    var warnings []Warning

    // schema validation（保留原有邏輯，改用 ctx.NodeByID）
    for _, node := range ctx.Nodes { ... }
    for _, edge := range ctx.Edges { ... }

    warnings = append(warnings, checkSPOF(ctx)...)
    warnings = append(warnings, checkDBSelection(ctx)...)
    // ... 其餘 check 函數同理
    return warnings
}
```

---

## 步驟 2：拆分 `topology_handler.go` 為多個檔案

**在 `handler/` 目錄下建立以下檔案，將對應函數搬過去：**

### `handler/warning.go`

搬入：

- `type Warning struct`
- `type AnalyzeResponse struct`

### `handler/helpers.go`

搬入：

- `func contains(s, substr string) bool` — 順便改名為 `labelContains`
- `func joinLabels(labels []string) string`
- 新增 `func hasReplicationEdge(nodeID string, edges []model.SystemEdge) bool`（見步驟 3）

### `handler/check_availability.go`

搬入：

- `checkSPOF`
- `checkLBSPOF`
- `checkReverseProxySPOF`
- `checkNoAutoScalingSingle`
- `checkNoHealthCheckBehindLB`
- `checkServerlessReplicas`

### `handler/check_database.go`

搬入：

- `checkDBSelection`
- `checkReadWriteSeparation`
- `checkCAP`
- `checkVerticalPartitioning`

### `handler/check_cache.go`

搬入：

- `checkCacheConsistency`
- `checkCacheEviction`

### `handler/check_messaging.go`

搬入：

- `checkAsyncDecoupling`
- `checkMQConsumer`
- `checkMQDLQ`
- `checkAsyncPeakShaving`

### `handler/check_network.go`

搬入：

- `checkProtocolMismatch`
- `checkConnectionTypeProtocolMismatch`
- `checkCDNUsage`
- `checkReverseProxySSL`
- 搬入相關 map：`expectedProtocols`, `validConnectionProtocolPairs`, `protocolDisplayName`, `connectionTypeNames`

### `handler/check_security.go`

搬入：

- `checkClientToDB`
- `checkClientToCache`
- `checkMissingFirewall`
- `checkFirewallMonitorMode`
- `checkFirewallL3Only`

### `handler/check_observability.go`

搬入：

- `checkMissingLogger`
- `checkIncompleteObservability`
- `checkAlertingDisabled`

### `handler/topology_handler.go`（剩餘）

只保留：

- `func PostTopology(c *gin.Context)`
- `func validate(t model.SystemTopology) []Warning`

---

## 步驟 3：消除重複代碼

### 3a. 抽出 `hasReplicationEdge`

**位置：** `handler/helpers.go`

```go
func hasReplicationEdge(nodeID string, edges []model.SystemEdge) bool {
    for _, edge := range edges {
        if (edge.Source == nodeID || edge.Target == nodeID) && edge.ConnectionType == "replication" {
            return true
        }
    }
    return false
}
```

**替換位置：**

- `checkDBSelection` 中的 replication 迴圈 → `if hasReplicationEdge(node.ID, ctx.Edges) { continue }`
- `checkReadWriteSeparation` 中的 replication 迴圈 → 同上

### 3b. 合併 `checkClientToDB` + `checkClientToCache`

兩個函數邏輯完全相同，只差 target role 和 warning 內容。合併為：

**位置：** `handler/check_security.go`

```go
func checkClientDirectAccess(ctx model.TopologyContext, targetRole, rule, message, solution string) []Warning {
    var warnings []Warning
    for _, edge := range ctx.Edges {
        source, okS := ctx.NodeByID[edge.Source]
        target, okT := ctx.NodeByID[edge.Target]
        if !okS || !okT { continue }
        if model.NodeHasRole(source, "client") && model.NodeHasRole(target, targetRole) {
            warnings = append(warnings, Warning{
                Rule:     rule,
                Message:  fmt.Sprintf(message, source.Label, target.Label),
                Solution: solution,
                NodeIDs:  []string{source.ID, target.ID},
            })
        }
    }
    return warnings
}
```

在 `validate()` 中呼叫：

```go
warnings = append(warnings, checkClientDirectAccess(ctx,
    "database",
    "client_direct_db",
    "🚫 安全風險：禁止從 %q 直接連線至 %q。",
    "Client 不應直接操作資料庫。請在兩者之間加入 API Gateway 或 Service 層進行身份驗證與數據抽象。",
)...)
warnings = append(warnings, checkClientDirectAccess(ctx,
    "cache",
    "client_direct_cache",
    "🧊 暴露風險：不建議從 %q 直接連線至 %q。",
    "不建議 Client 直接操作快取。這可能導致快取穿透風險或數據洩漏。應透過後端 Service 進行快取邏輯封裝。",
)...)
```

**刪除：** 原本的 `checkClientToDB` 和 `checkClientToCache`。

### 3c. 合併 `checkLBSPOF` + `checkReverseProxySPOF`

**位置：** `handler/check_availability.go`

```go
func checkEntryPointSPOF(ctx model.TopologyContext, role, rule, emoji, label string) []Warning {
    var matched []model.SystemNode
    for _, node := range ctx.Nodes {
        if model.NodeHasRole(node, role) {
            matched = append(matched, node)
        }
    }
    if len(matched) != 1 {
        return nil
    }

    node := matched[0]
    props, err := model.ParseNodeProperties(node)
    if err == nil {
        // 檢查是否有 Replicas 欄位 > 1
        switch p := props.(type) {
        case *model.LoadBalancerProperties:
            if p.Replicas > 1 { return nil }
        case *model.ReverseProxyProperties:
            if p.Replicas > 1 { return nil }
        }
    }

    return []Warning{{
        Rule:     rule,
        Message:  fmt.Sprintf("%s 入口單點故障：整體架構中僅存在 1 個 %s。", emoji, label),
        Solution: fmt.Sprintf("生產環境建議部署多個 %s，或在屬性面板中將 Replicas 複本數設為 2 以上。", label),
        NodeIDs:  []string{node.ID},
    }}
}
```

在 `validate()` 中：

```go
warnings = append(warnings, checkEntryPointSPOF(ctx, "load_balancer", "lb_spof", "⚖️", "Load Balancer")...)
warnings = append(warnings, checkEntryPointSPOF(ctx, "reverse_proxy", "reverse_proxy_spof", "🔀", "Reverse Proxy")...)
```

**刪除：** 原本的 `checkLBSPOF` 和 `checkReverseProxySPOF`。

---

## 步驟 4：搬移 protocol maps 到 model 層

將以下 map 從 `handler/check_network.go` 搬到 `model/` 下新檔案 `model/protocols.go`：

- `expectedProtocols`
- `validConnectionProtocolPairs`
- `protocolDisplayName`

這些是 domain knowledge，不屬於 handler。大寫首字母匯出即可：

```go
// model/protocols.go
package model

var ExpectedProtocols = map[string]map[string]bool{ ... }
var ValidConnectionProtocolPairs = map[string]map[string]bool{ ... }
var ProtocolDisplayName = map[string]string{ ... }
```

handler 中改為 `model.ExpectedProtocols` 等。

---

## 步驟 5：更新測試檔

`topology_handler_test.go` 中部分測試直接呼叫 check 函數。改為傳入 `model.TopologyContext`：

```go
// 改前
warnings := checkProtocolMismatch(nodes, edges)

// 改後
ctx := model.TopologyContext{
    Nodes:    slices.Collect(maps.Values(nodes)), // 或手動建構
    Edges:    edges,
    NodeByID: nodes,
    Outgoing: buildOutgoing(edges),
}
warnings := checkProtocolMismatch(ctx)
```

也可以加一個測試用 helper：

```go
func makeCtx(nodes map[string]model.SystemNode, edges []model.SystemEdge) model.TopologyContext { ... }
```

---

## Checklist

- [ ] 步驟 1：新增 `TopologyContext`，統一所有 check 函數簽名
- [ ] 步驟 2：拆分 `topology_handler.go` 為 10 個檔案
- [ ] 步驟 3a：抽出 `hasReplicationEdge` helper
- [ ] 步驟 3b：合併 `checkClientToDB` + `checkClientToCache`
- [ ] 步驟 3c：合併 `checkLBSPOF` + `checkReverseProxySPOF`
- [ ] 步驟 4：搬移 protocol maps 到 `model/protocols.go`
- [ ] 步驟 5：更新測試檔
- [ ] 每步完成後跑 `go test ./...` 確認通過
