# ArchitectMind: 系統設計視覺化工具 (System Design Visualizer)

ArchitectMind 是一個網頁版的系統架構繪圖與分析工具。它提供了一個直觀的互動式畫布（Canvas），讓使用者可以拖放各種基礎設施組件（如 Load Balancer, Database, Service 等），並透過後端 API 進行架構邏輯的語法驗證。

## 🚀 功能特點

- **互動式畫布**：基於 React Flow，支援節點拖放、連線與自定義屬性。
- **組件側邊欄**：提供標準的系統設計組件庫（Client, CDN, API Gateway 等）。
- **後端語法驗證**：自動檢查架構圖中的組件類型與連線邏輯是否正確。
- **響應式設計**：簡潔現代的使用者介面，適合快速原型設計。

## 🛠 技術棧

### 前端 (Frontend)

- **React 19** + **TypeScript**
- **Vite** (開發與構建工具)
- **React Flow (@xyflow/react)** (畫布引擎)
- **Vanilla CSS** (樣式設計)

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
2. **建立連線**：點擊並拖曳組件邊緣的點（Handle）來連接不同節點。
3. **語法分析**：前端會將目前的拓撲結構（Topology）發送至後端 `/api/topology` 進行驗證。
4. **檢查警告**：若連線不合規範（例如非法的連接類型），畫面上會顯示後端回傳的警告訊息。

## 📂 專案結構

- `/frontend`: 包含 React 原始碼、React Flow 組件與 API 調用邏輯。
- `/backend`: 包含 Go API 處理程序（Handlers）、數據模型（Models）與驗證邏輯。
- `GEMINI.md`: 專案的進階技術上下文與開發規範。

## 🛠 開發規範

- **新增組件**：若需新增系統組件類型，請同時更新 `backend/model/topology.go` 中的 `ValidComponentTypes` 以及前端的 `nodeConfig.ts`。
- **API 通訊**：前端使用 `fetch` 調用後端，請確保跨域設定（CORS）在 `backend/main.go` 中包含你的開發環境地址。
