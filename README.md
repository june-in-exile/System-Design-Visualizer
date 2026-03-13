# ArchitectMind: 系統設計視覺化工具 (System Design Visualizer)

![ArchitectMind Demo](./demo.png)

ArchitectMind 是一個網頁版的系統架構繪圖與分析工具。它提供了一個直觀的互動式畫布 (Canvas)，讓使用者可以拖放各種基礎設施組件（如 Load Balancer, Database, Service 等），並透過後端 API 進行架構邏輯的語法驗證與最佳實踐建議。

## 🚀 功能特點

- **互動式畫布**：基於 React Flow v12，支援節點拖放、連線與自定義屬性。
- **組件側邊欄**：提供標準的系統設計組件庫（Client, CDN, API Gateway, Message Queue 等）。
- **後端語法驗證 (Analyze)**：自動檢查架構圖中的組件類型與連線邏輯是否符合系統設計最佳實踐。
- **一鍵生成預設架構 (Presets)**：提供經典系統設計模板（Standard Demo, Twitter, YouTube, Google），協助快速上手。
- **快捷操作**：支援 **Duplicate** (複製) 與 **Undo** (還原) 功能，提升繪圖效率。
- **響應式設計**：支援深色模式 (Dark Mode)，簡潔現代的使用者介面。

## 🛠 技術棧

### 前端 (Frontend)

- **React 19** + **TypeScript**
- **Vite** (開發與構建工具)
- **React Flow (@xyflow/react v12)** (畫布引擎)
- **Vanilla CSS** (樣式設計，支援 CSS Variables)

### 後端 (Backend)

- **Go 1.25+**
- **Gin Web Framework** (API 路由)
- **MongoDB** (預留資料持久化介面)

## 📦 安裝與運行步驟

### 1. 複製專案

```bash
git clone <your-repo-url>
cd "System Design Visualizer"
```

### 2. 快速啟動 (Quick Start)

專案根目錄提供了一個便利的腳本，可同時啟動後端與前端開發伺服器：

```bash
# 賦予執行權限 (若需要)
chmod +x start-dev.sh

# 執行啟動腳本
./start-dev.sh
```

啟動後：

- 前端：`http://localhost:5173`
- 後端：`http://localhost:8080`

### 3. 分別啟動 (Manual Start)

如果你需要分別查看日誌或進行除錯，可以手動啟動：

#### A. 啟動後端 (Go)

後端預設執行於 `http://localhost:8080`。

```bash
cd backend
go mod download
go run main.go
```

#### B. 啟動前端 (React)

前端預設執行於 `http://localhost:5173`。

```bash
cd frontend
npm install
npm run dev
```

## ✅ 測試 (Backend Tests)

目前測試集中於後端。請在 `backend` 目錄下執行：

```bash
cd backend
go test ./... -v -count=1
```

若要查看覆蓋率：

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go tool cover -html=coverage.out
```

## 🖥 如何使用

1. **一鍵生成預設架構 (Presets)**：
   - 點擊工具列上的 **Demo** 按鈕，畫布將自動載入一個包含 Client、LB、Service、Database 與 Cache 的經典架構，方便進行測試。
   - 點擊 **Twitter** 按鈕，系統將生成一個具備 Async Decoupling (Kafka) 與 Fan-out 機制的社群媒體架構。
   - 點擊 **YouTube** 按鈕，展示包含 Video Upload Pipeline、Transcoding 與 CDN 分發的大規模影音系統。
   - 點擊 **Google** 按鈕，視覺化搜尋引擎的 Crawler、Indexing Pipeline、Ranking 與 Autocomplete 服務架構。

2. **繪製與編輯 (Duplicate & Edit)**：
   - 從左側側邊欄拖曳組件至中央畫布。
   - 選取節點後點擊 **Duplicate** (或按 `Cmd/Ctrl + D`) 快速複製相同類型的組件。
   - 點擊畫布上的節點，右側會出現 **Property Panel**。你可以在此修改 Label、讀寫比例 (Read/Write Ratio) 或指定產品 (如 Cassandra)。

3. **操作還原 (Undo)**：
   - 在繪圖過程中若發生錯誤，可點擊 **Undo** 按鈕 (或按 `Cmd/Ctrl + Z`) 回退至先前的狀態。

4. **架構分析 (Analyze)**：
   - 點擊上方 **Analyze** 按鈕（或使用快捷鍵 `Cmd/Ctrl + A`），系統會將拓撲結構發送至後端進行深度驗證。
   - 若架構存在風險（如單點故障），畫面上會出現警告訊息。

5. **定位警告**：
   - 若出現警告，點擊警告面板中的項目可自動將畫布縮放並定位至有問題的節點。

## 🔍 架構驗證規則 (Validation Rules)

後端目前實作了以下十項進階架構驗證規則，協助開發者識別潛在的系統設計風險：

| 編號 | 規則名稱 | 觸發條件 / 說明 | 觸發方式 (操作建議) |
| :--- | :--- | :--- | :--- |
| 1 | **單點故障檢查 (SPOF)** | 負載均衡器 (LB) 下游僅連接 1 個服務節點時觸發。 | 放置一個 **Load Balancer** 並僅連線至 **1 個 Service**。 |
| 2 | **資料庫選型建議** | SQL 資料庫且寫入比例超過 50% (`readWriteRatio < 0.5`)。系統會自動檢測是否存在 `replication` 連線並智慧過濾已優化項目。 | 選取 **Database**，設 `dbType: "sql"` 且 `readWriteRatio < 0.5`。 |
| 3 | **垂直切分提醒** | 拓撲中存在 2 個（含）以上的資料庫節點。 | 在畫布上放置 **2 個或更多** 的 **Database** 節點。 |
| 4 | **快取一致性檢查**  rule 6| 服務同時連接至快取 (Cache) 與資料庫 (Database)。 | 將同一個 **Service** 節點同時連線至 **Cache** 與 **Database**。 |
| 5 | **CAP 定理提醒** | 使用特定的 NoSQL 產品（如 Cassandra, DynamoDB 等）。 | 選取 **Database**，將屬性中的 `product` 設為 `cassandra` 或 `dynamodb` 等。 |
| 6 | **CDN 全球加速建議** | 拓撲中存在 Client 但缺乏 CDN 節點，可能導致靜態資源加載緩慢。 | 放置一個 **Client** 節點，但未在架構中加入 **CDN**。 |
| 7 | **異步解耦提醒** | 服務間存在同步呼叫且涉及耗時操作（如標籤包含 mail, img, report, task 等關鍵字）。 | 建立 **Service** 連線，並將目標服務命名為 **Gmail Service** 或 **Image Processor**。 |
| 8 | **入口單點故障 (LB SPOF)** | 整體架構中僅存在 1 個 Load Balancer，入口處存在單點故障風險。 | 在畫布上僅放置 **1 個 Load Balancer** 節點。 |
| 9 | **讀寫分離實踐建議** | 資料庫讀取比例極高 (> 80%)，適合導入 Master-Slave 架構。系統會自動檢測是否存在 `replication` 連線並智慧過濾已優化項目。 | 選取 **Database**，將屬性中的 `readWriteRatio` 設為大於 `0.8`。 |
| 10 | **快取失效策略提醒** | 使用 Cache 節點但未配置適當的失效演算法 (Eviction Policy)。 | 放置一個 **Cache** 節點，並在屬性中將 `evictionPolicy` 留空。 |
| 11 | **MQ 消費者缺失檢查** | Message Queue 節點缺乏輸出連線（無消費者），訊息將在隊列中堆積。 | 放置一個 **Message Queue** 節點但不連接任何輸出至 Service。 |
| 12 | **死信隊列 (DLQ) 提醒** | 使用 Message Queue 但未配置死信隊列或重試機制，可能導致訊息丟失。 | 放置 **Message Queue** 節點，但在屬性中將 `hasDLQ` 設為 `false`。 |
| 13 | **異步削峰實踐建議** | 流量入口直接連線至高負載寫入節點（如 Database）而未經過 MQ 緩衝。 | 從 **API Gateway** 或 **LB** 直接連線至 **Database**，且寫入比例高。 |
| 14 | **禁止 Client 直接連線資料庫** | Client 節點直接連線至 Database，跳過了 API Gateway 或 Service 抽象層。 | 建立從 **Client** 直連 **Database** 的連線區域。 |
| 15 | **禁止 Client 直接連線快取** | Client 節點直接連線至 Cache，可能導致安全風險或快取穿透。 | 建立從 **Client** 直連 **Cache** 的連線。 |
| 16 | **協定不匹配 (Protocol Mismatch)** | 連線使用了與目標組件類型不匹配的協議（如對 SQL Database 使用 HTTP 而非 Database Protocol）。 | 建立連線，並在連線屬性中設定與目標不符的 **Protocol**（例如選取 RESP 連接到 SQL 資料庫）。 |
| 17 | **屬性不匹配 (Conn/Protocol Mismatch)** | 連線類型與通訊協定邏輯不一致（如同步連線搭配異步協定 AMQP）。 | 建立連線，將 **Connection Type** 設為 **Synchronous**，但 **Protocol** 選擇 **AMQP**。 |
| 18 | **缺少 Firewall/WAF** | 架構中有 Client 與入口節點 (LB/API Gateway) 但缺乏 Firewall，或 Firewall 存在但未連線至入口節點。 | 放置 **Client** 與 **Load Balancer**，但不放置 **Firewall**；或放置 **Firewall** 但未連線至 **Load Balancer**。 |
| 19 | **缺少 Logger/Monitor** | 架構中有 3 個（含）以上 Service 但缺乏 Logger/Monitor，或 Logger/Monitor 存在但未連線至任何 Service。 | 放置 **3 個或更多 Service**，但不放置 **Monitor**；或放置 **Monitor** 但未連線至任何 **Service**。 |

## 📂 專案結構

- `/frontend`: 包含 React 原始碼、React Flow 組件、自定義 Hook 與 API 調用邏輯。
- `/backend`: 包含 Go API 處理程序 (Handlers)、數據模型 (Models) 與驗證引擎。

## 🗺️ 路線圖 (Roadmap / TODO)

- [ ] **Terraform 整合**：研究與 [Terraform](https://developer.hashicorp.com/terraform) 串接，實現從架構圖自動生成 IaC (Infrastructure as Code) 設定檔。
- [ ] **數據儀表板 (Dashboard)**：加入即時數據模擬面版，根據用戶設定的節點參數（如 Replicas, Read/Write Ratio），動態計算並顯示預估的 DAU, QPS, 延遲等系統指標。
