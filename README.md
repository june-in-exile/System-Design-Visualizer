# ArchitectMind: 系統設計視覺化工具 (System Design Visualizer)

ArchitectMind 是一個網頁版的系統架構繪圖與分析工具。它提供了一個直觀的互動式畫布 (Canvas)，讓使用者可以拖放各種基礎設施組件（如 Load Balancer, Database, Service 等），並透過後端 API 進行架構邏輯的語法驗證與最佳實踐建議。

## 🚀 功能特點

- **互動式畫布**：基於 React Flow v12，支援節點拖放、連線與自定義屬性。
- **組件側邊欄**：提供標準的系統設計組件庫（Client, CDN, API Gateway, Message Queue 等）。
- **後端語法驗證**：自動檢查架構圖中的組件類型與連線邏輯是否正確。
- **快捷操作**：支援 `Cmd/Ctrl + A` (分析)、`Cmd/Ctrl + D` (複製) 與 `Cmd/Ctrl + Z` (還原)。
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

### 2. 啟動後端 (Go)

後端預設執行於 `http://localhost:8080`。

```bash
cd backend
go mod download
go run main.go
```

### 3. 啟動前端 (React)

前端預設執行於 `http://localhost:5173`。

```bash
cd frontend
npm install
npm run dev
```

## 🖥 如何使用

1. **繪製架構**：從左側側邊欄拖曳組件至中央畫布。
2. **編輯屬性**：點擊畫布上的節點，右側會出現 **Property Panel**。你可以在此修改 Label、資料庫類型 (SQL/NoSQL)、讀寫比例 (Read/Write Ratio) 或指定產品 (如 Cassandra)。
3. **建立連線**：點擊並拖曳組件邊緣的點 (Handle) 來連接不同節點。
4. **語法分析**：點擊上方 **Analyze** 按鈕（或使用快捷鍵 `Cmd/Ctrl + A`），系統會將拓撲結構發送至後端進行驗證。
5. **檢查警告**：若架構存在風險，畫面上會出現警告訊息，點擊警告可自動定位至相關節點。

## 🔍 架構驗證規則 (Validation Rules)

後端目前實作了以下五項進階架構驗證規則，協助開發者識別潛在的系統設計風險：

| 編號 | 規則名稱 | 觸發條件 / 說明 | 觸發方式 (操作建議) |
| :--- | :--- | :--- | :--- |
| 1 | **單點故障檢查 (SPOF)** | 負載均衡器 (LB) 下游僅連接 1 個服務節點時觸發。 | 放置一個 **Load Balancer** 並僅連線至 **1 個 Service**。 |
| 2 | **資料庫選型建議** | SQL 資料庫且寫入比例超過 50% (`readWriteRatio < 0.5`)。 | 選取 **Database**，設 `dbType: "sql"` 且 `readWriteRatio < 0.5`。 |
| 3 | **垂直切分提醒** | 拓撲中存在 2 個（含）以上的資料庫節點。 | 在畫布上放置 **2 個或更多** 的 **Database** 節點。 |
| 4 | **快取一致性檢查** | 服務同時連接至快取 (Cache) 與資料庫 (Database)。 | 將同一個 **Service** 節點同時連線至 **Cache** 與 **Database**。 |
| 5 | **CAP 定理提醒** | 使用特定的 NoSQL 產品（如 Cassandra, DynamoDB 等）。 | 選取 **Database**，將屬性中的 `product` 設為 `cassandra` 或 `dynamodb` 等。 |

## 📂 專案結構

- `/frontend`: 包含 React 原始碼、React Flow 組件、自定義 Hook 與 API 調用邏輯。
- `/backend`: 包含 Go API 處理程序 (Handlers)、數據模型 (Models) 與驗證引擎。
- `GEMINI.md`: 專案的進階技術上下文與開發規範。

## 🛠 開發規範

- **新增組件**：若需新增系統組件類型，請同時更新 `backend/model/topology.go` 中的 `ValidComponentTypes` 以及前端的 `nodeConfig.ts`。
- **屬性擴充**：在 `backend/model/properties.go` 定義新的屬性結構，並確保前端 `PropertyPanel.tsx` 能夠正確渲染對應欄位。
- **快捷鍵實作**：全域快捷鍵定義在 `frontend/src/components/Canvas.tsx` 的 `useEffect` 中。
