# Demo 架構審查報告

> 以下問題是現有規則引擎**尚未覆蓋**的架構隱患，基於系統設計最佳實踐分析。

---

## 🐦 Twitter Demo

| # | 問題 | 為什麼是問題 | 修復方向 | 實作步驟 |
|---|------|-------------|---------|---------|
| T1 | **單一 Database 承擔所有服務** — User Service、Tweet Service、Fan-out Service、Search Service 全部寫入同一個 MySQL (`tw-db`)，且 readWriteRatio=0.2（高寫入） | Twitter 級別的寫入量下，單一 MySQL 會成為瓶頸。User 資料和 Tweet 資料的存取模式完全不同（User 讀多寫少，Tweet 寫多讀多），混在一起無法針對性優化 | 按業務拆分資料庫（Database per Service） | 1. 新增 `tw-user-db`（SQL, readWriteRatio=0.8）給 User Service<br>2. 新增 `tw-tweet-db`（NoSQL/Cassandra, readWriteRatio=0.3）給 Tweet Service<br>3. `tw-db` 改為 Search Service 專用的搜尋索引庫<br>4. Fan-out Service 只寫 Cache，不直接碰 DB |
| T2 | **Timeline Service 只讀 Cache，沒有 DB fallback** — `tw-timeline-srv` → `tw-cache` 是唯一的資料路徑 | Cache miss 或 Redis 故障時，Timeline 完全無法服務。沒有 cold start 的回退策略 | 加入 DB 讀取路徑作為 fallback | 1. 新增 `tw-timeline-srv` → `tw-db`（或專用 timeline DB）的 sync 連線<br>2. 實作 Cache-aside 模式：先查 Cache → miss 時查 DB → 回填 Cache |
| T3 | **Fan-out Service 和 Search Service 缺少 Monitor 連線** — 只有 User/Tweet/Timeline Service 連到 Monitor | Fan-out 是 Twitter 最關鍵的非同步處理（推送到所有 follower），無法觀測它的延遲和錯誤率等於盲飛 | 所有 Service 都應連到 Monitor | 1. 新增 `tw-fanout-srv` → `tw-monitor` (async, http)<br>2. 新增 `tw-search-srv` → `tw-monitor` (async, http) |
| T4 | **CDN 沒有連接 Storage** — `tw-cdn` 存在但與 `tw-storage` 無連線 | CDN 的靜態資源（圖片、影片縮圖）需要從 Origin（Storage）拉取，沒有 origin pull 等於 CDN 無法正常工作 | CDN 設定 Origin Pull 到 Storage | 1. 新增 `tw-cdn` → `tw-storage` (cdn_origin, https) |
| T5 | **Media Service 是同步上傳到 Storage** — Client → API Gateway → Media Service → Storage 全程同步 | 大檔案上傳（圖片/影片）在同步路徑上會長時間佔用連線，導致 API Gateway 和 Service 的 worker thread 被卡住 | 改為異步上傳流程 | 1. Media Service 收到上傳請求後，產生 presigned URL 回給 Client<br>2. Client 直接上傳到 Storage（或經由 CDN）<br>3. 上傳完成後透過 MQ 通知 Media Service 處理後續（縮圖、轉碼） |
| T6 | **所有服務間通訊用 HTTP** — API Gateway 到下游 Service 全用 `protocol: http` | 內部服務間用 HTTP/1.1 效率低（文字序列化、無多路複用），在微服務架構下每個請求都有序列化/反序列化開銷 | 內部通訊改用 gRPC | 1. API Gateway → 各 Service 的 protocol 改為 `grpc`<br>2. Service 之間的呼叫也統一用 gRPC<br>3. 只有面對 Client 的入口保持 HTTP/HTTPS |
| T7 | **Search Service 直接查 MySQL** — 搜尋功能直接打 SQL 資料庫 | 全文搜尋在 RDBMS 上效率極差，Twitter 規模下 LIKE 查詢會拖垮資料庫 | 使用專門的搜尋引擎 | 1. 新增 Elasticsearch/Solr 節點<br>2. Search Service 改為查 Elasticsearch<br>3. Tweet Service 寫入時透過 MQ 同步到搜尋索引 |

---

## 🎬 YouTube Demo

| # | 問題 | 為什麼是問題 | 修復方向 | 實作步驟 |
|---|------|-------------|---------|---------|
| Y1 | **單一 SQL Database 承擔多服務** — User/Search/Streaming/Metadata 全打同一個 `yt-db`（SQL, readWriteRatio=0.8） | YouTube 的影片 metadata 數量龐大且讀取頻率極高，混合 User 資料和 Metadata 在同一 DB 會導致 lock contention | 分庫 | 1. 新增 `yt-user-db`（SQL）給 User Service<br>2. `yt-db` 改為 Metadata 專用<br>3. Search Service 使用 Elasticsearch 而非 SQL |
| Y2 | **Streaming Service 直接查 DB** — `yt-streaming-srv` → `yt-db` (sync, database) | 影片串流是 YouTube 最高頻的操作，每次播放都查 DB 拿 metadata 不現實，應該 Cache-first | Cache-first 策略 | 1. Streaming Service 應先查 Cache 拿影片 URL/metadata<br>2. 已有 `yt-streaming-srv` → `yt-cache`，但也連 DB，需確保 Cache-aside 邏輯正確<br>3. 熱門影片應在 CDN 層直接服務 |
| Y3 | **CDN 沒有連接到 Client 的服務路徑** — Client → CDN 存在，但 CDN 沒連到 Streaming 流程 | 影片實際播放時，Client 應該從 CDN 拉取影片，而非穿透到 Streaming Service。目前的連線暗示 CDN 只用於靜態資源 | CDN 作為影片分發主力 | 1. Streaming Service 回傳 CDN URL 給 Client<br>2. Client 直接從 CDN 拉取影片串流<br>3. CDN → Storage 的 origin pull 已存在（`e-yt-cdn-storage`），架構正確但流程不完整 |
| Y4 | **User/Search/Recommendation Service 缺少 Monitor 連線** — 只有 Streaming 和 Upload 連到 Monitor | 6 個 Service 中只有 2 個有監控，Search 和 Recommendation 的效能和錯誤無法追蹤 | 全部 Service 都應連到 Monitor | 1. 新增 `yt-user-srv` → `yt-monitor` (async, http)<br>2. 新增 `yt-search-srv` → `yt-monitor` (async, http)<br>3. 新增 `yt-reco-srv` → `yt-monitor` (async, http)<br>4. 新增 `yt-metadata-srv` → `yt-monitor` (async, http) |
| Y5 | **Recommendation Service 只讀 Cache** — `yt-reco-srv` → `yt-cache`，沒有 DB 或 ML pipeline 連線 | 推薦系統需要大量使用者行為數據和 ML 模型，只從 Cache 讀取暗示預計算結果已被快取，但沒有任何機制生成或更新這些結果 | 加入 ML pipeline 或 DB 連線 | 1. 新增 `yt-reco-srv` → `yt-db` 讀取使用者行為<br>2. 或新增 ML inference service + feature store 節點<br>3. 定期透過 batch job 更新 Cache 中的推薦結果 |
| Y6 | **Search Service 直接查 SQL** — 同 Twitter T7 問題 | 影片搜尋需要全文搜尋 + faceted search（按時長、上傳日期、類別篩選），SQL 無法高效處理 | 使用 Elasticsearch | 1. 新增 Elasticsearch 節點<br>2. Search Service 改查 Elasticsearch<br>3. Metadata Service 更新時透過 MQ 同步索引 |
| Y7 | **Transcoding Service 缺少 Monitor 連線** — 轉碼是最耗資源的操作，5 個 replicas，但沒有監控 | 轉碼失敗、卡住、或資源耗盡都無法被發現 | 加入 Monitor | 1. 新增 `yt-transcoding-srv` → `yt-monitor` (async, http) |

---

## 🔍 Google Search Demo

| # | 問題 | 為什麼是問題 | 修復方向 | 實作步驟 |
|---|------|-------------|---------|---------|
| G1 | **Query → Ranking → Ads → Snippet 是同步串行鏈** — 4 個 Service 串行呼叫 (grpc) | 搜尋延遲 = 所有服務延遲之和。任一服務變慢，整個搜尋就慢。且 Ads 和 Snippet 不依賴彼此，不需要串行 | 改為並行 fan-out | 1. Query Service 同時呼叫 Ranking、Ads、Snippet（parallel fan-out）<br>2. 等結果回來後 merge<br>3. 或引入 Orchestrator Service 負責 scatter-gather |
| G2 | **Crawler Service 缺少 Monitor 連線** — Crawler 和 Indexing Service 都沒連 Monitor | 爬蟲是 Google 搜尋的資料來源，如果爬蟲失敗或被 block，搜尋品質會逐漸衰退但沒人知道 | 全部 Service 加 Monitor | 1. 新增 `gg-crawler-srv` → `gg-monitor` (async, http)<br>2. 新增 `gg-indexing-srv` → `gg-monitor` (async, http)<br>3. 新增 `gg-auto-srv` → `gg-monitor` (async, http)<br>4. 新增 `gg-ads-srv` → `gg-monitor` (async, http)<br>5. 新增 `gg-snippet-srv` → `gg-monitor` (async, http) |
| G3 | **Autocomplete Service 只讀 Cache** — `gg-auto-srv` → `gg-cache`，沒有 DB fallback | Cache miss 或 Redis 故障時自動完成功能完全失效，對使用者體驗影響極大 | 加入 DB/Trie 的 fallback | 1. 新增 `gg-auto-srv` → `gg-db-index` (sync, database) 作為 fallback<br>2. 或新增專門的 Trie database 節點 |
| G4 | **Ranking 和 Ads Service 沒有資料來源** — 這兩個 Service 沒有連到任何 DB 或 Cache | Ranking 需要 PageRank 分數和排名信號，Ads 需要廣告庫存和出價數據，但架構中完全沒有資料儲存連線 | 加入資料來源 | 1. 新增 `gg-ranking-srv` → `gg-db-index` (sync, database) 讀取 PageRank<br>2. 新增 `gg-ads-db`（Ads Database）<br>3. 新增 `gg-ads-srv` → `gg-ads-db` (sync, database)<br>4. 或兩者都從 Cache 讀取預計算結果 |
| G5 | **Snippet Service 沒有資料來源** — 需要產生搜尋結果摘要但沒有連到 Storage 或 DB | Snippet 需要讀取網頁內容（或預先產生的摘要）才能呈現搜尋結果的預覽文字 | 連接 Storage 或 Cache | 1. 新增 `gg-snippet-srv` → `gg-storage` (sync, https) 讀取 HTML<br>2. 或新增 `gg-snippet-srv` → `gg-cache` (sync, resp) 讀取預計算摘要 |
| G6 | **CDN 沒有 Origin 連線** — `gg-cdn` 存在但沒有連到任何後端資源 | CDN 需要 origin server 來拉取和快取內容，孤立的 CDN 無法提供服務 | 設定 Origin | 1. 新增 `gg-cdn` → `gg-storage` (cdn_origin, https)<br>2. 或 `gg-cdn` → `gg-rp` (cdn_origin, https) |
| G7 | **Indexing Service 只寫 DB，沒有讀取 Storage** — Crawler 爬取的 HTML 存到 Storage，但 Indexing Service 只寫 `gg-db-index` | Indexing 需要讀取 Crawler 存入 Storage 的原始 HTML 才能建立倒排索引，目前的資料流斷裂了 | 連接 Storage | 1. 新增 `gg-indexing-srv` → `gg-storage` (sync, https) 讀取爬取的 HTML<br>2. 完整流程：Crawler → Storage + MQ → Indexing → 讀取 Storage → 寫入 DB |

---

## 📊 三個 Demo 的共通問題

| # | 問題 | 影響範圍 | 修復方向 | 新規則建議 |
|---|------|---------|---------|-----------|
| C1 | **部分 Service 未連接 Monitor** | 全部三個 Demo | 所有 Service 應連到 Logger/Monitor | 現有 `checkMissingLogger` 只檢查「是否存在 Logger」和「是否有任一 Service 連線」，應改為「未連線到 Logger 的 Service 列表」 |
| C2 | **Service 只讀 Cache 無 DB fallback** | Twitter (Timeline)、YouTube (Reco)、Google (Autocomplete) | Cache-aside 模式應有 DB 回退路徑 | 新增規則：偵測只連 Cache 不連 DB 的 Service（排除 Fan-out 等純寫 Cache 場景） |
| C3 | **CDN 存在但未正確連線到 Origin** | Twitter、Google | CDN 必須有 cdn_origin 連線 | 新增規則：偵測孤立的 CDN 節點（無 cdn_origin 類型的 outgoing edge） |
| C4 | **Search Service 直接查 SQL/NoSQL** | Twitter、YouTube | 搜尋應使用專門的搜尋引擎 | 新增規則：Label 含 "search" 的 Service 直接連 Database 時提醒考慮 Elasticsearch |
| C5 | **同步串行呼叫鏈過長** | Google (Query→Ranking→Ads→Snippet) | 可並行的呼叫不應串行 | 新增規則：偵測 ≥3 個 Service 的同步串行鏈（A→B→C→D） |
| C6 | **Service 無任何資料來源** | Google (Ranking, Ads, Snippet) | 每個 Service 應有明確的資料讀取來源 | 新增規則：偵測沒有 outgoing edge 到 DB/Cache/Storage 的 Service（排除純 gateway/proxy 角色） |
