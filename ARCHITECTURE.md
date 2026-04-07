# 🦈 鯊魚寶寶選股系統 - 架構文件

**建立時間**：2026-04-07  
**維護者**：鯊魚寶寶 🦈

---

## ⚠️ 重要警告

### **絕對不要移動以下檔案！**

網頁 JavaScript 使用**相對路徑**讀取資料，這些檔案必須保持在 `stock_web/` 根目錄：

```
stock_web/
├── index.html          ← 主頁（必須在根目錄）
├── trade.html          ← 交易管理頁
├── dividend_tracker.html
├── dividend_strategy.html
├── sector_tab.html
├── daily_report.json   ← 選股資料（JS 讀取 /daily_report.json）
├── portfolio.json      ← 持倉資料（JS 讀取 /portfolio.json）
├── dividend_data.json  ← 配息資料
├── sector_heatmap.json ← 產業熱度資料
├── stock_trades.db     ← 交易記錄資料庫
├── stock_data.db       ← 股票資料庫
└── *.go                ← Go 原始碼
```

**原因**：HTML 中的 JavaScript 使用 `fetch('/daily_report.json')` 等路徑，如果移動檔案會造成 404 錯誤。

---

## 📂 目錄結構

```
~/.openclaw/workspace/
├── 核心配置檔（不要動）
│   ├── SOUL.md
│   ├── USER.md
│   ├── AGENTS.md
│   ├── TOOLS.md
│   ├── IDENTITY.md
│   └── HEARTBEAT.md
│
├── stock_web/              ← 股票系統主目錄
│   ├── *.html              ← 網頁檔案（8個）
│   ├── *.json              ← 資料檔案（7個）
│   ├── *.db                ← 資料庫（3個）
│   ├── *.go                ← Go 原始碼（31個）
│   ├── go.mod / go.sum     ← Go 模組配置
│   ├── bin/                ← 編譯好的執行檔
│   │   └── trade_api       ← 交易 API 執行檔
│   └── *.sh                ← 啟動腳本
│
├── memory/                 ← 記憶資料夾
├── docs/                   ← 文件庫
├── backups/                ← 備份
└── portfolio.json          ← 根目錄也有一份（web_updater 會更新這裡）
```

---

## 🌐 服務架構

### **Port 配置**

| Port | 服務 | 執行方式 | 說明 |
|------|------|----------|------|
| 8080 | 網頁伺服器 | `python3 -m http.server 8080` | 靜態檔案服務 |
| 8888 | 交易管理 API | `./bin/trade_api` | 持倉、交易記錄 |
| 8765 | 股票查詢 API | `go run stock_query_api.go` | 即時技術分析 |

### **啟動指令**

```bash
cd ~/.openclaw/workspace/stock_web

# 1. 網頁伺服器
nohup python3 -m http.server 8080 > server.log 2>&1 &

# 2. 交易 API
nohup ./bin/trade_api > trade_api.log 2>&1 &

# 3. 股票查詢 API（可選）
nohup go run stock_query_api.go > stock_query.log 2>&1 &
```

### **停止服務**

```bash
pkill -f "python3 -m http.server"
pkill -f "trade_api"
pkill -f "stock_query_api"
```

---

## 📊 資料流

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  daily picker   │────▶│ daily_report.json│────▶│   index.html    │
│  (每日選股)      │     │  (選股結果)       │     │   (TOP 選股)    │
└─────────────────┘     └──────────────────┘     └─────────────────┘

┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  web_updater    │────▶│  portfolio.json  │────▶│   index.html    │
│  (股價更新)      │     │  (持倉資料)       │     │   (持倉明細)    │
└─────────────────┘     └──────────────────┘     └─────────────────┘

┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   trade_api     │◀───▶│ stock_trades.db  │────▶│   trade.html    │
│   (Port 8888)   │     │  (交易記錄)       │     │   (歷次買賣)    │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

---

## ⚡ 網頁訪問路徑

| 功能 | URL |
|------|-----|
| 主頁（TOP 選股） | http://localhost:8080/index.html |
| 交易管理 | http://localhost:8080/trade.html |
| 配息追蹤 | http://localhost:8080/dividend_tracker.html |
| 配息策略 | http://localhost:8080/dividend_strategy.html |
| 產業熱度 | http://localhost:8080/sector_tab.html |
| 交易 API | http://localhost:8888/api |
| 股票查詢 API | http://localhost:8765/api/query?symbol=2330 |

---

## 🔄 Cron 排程

| 排程 | 時間 | 說明 |
|------|------|------|
| 每日選股推播 | 06:10 | 執行 daily picker，推播 Telegram |
| 配息資料更新 | 06:00 | 更新 dividend_data.json |

---

## 💡 維護注意事項

### **✅ 可以做的事**
- 修改 `.go` 原始碼後重新編譯
- 更新 `daily_report.json`、`portfolio.json` 資料
- 新增 `.sh` 腳本
- 備份整個 `stock_web/` 目錄

### **❌ 不要做的事**
- 移動 `.html` 檔案到子目錄
- 移動 `.json` 資料檔到子目錄
- 移動 `.db` 資料庫到子目錄
- 更改網頁伺服器的工作目錄

### **如果要整理目錄**
1. 先備份：`tar -czf backup.tar.gz stock_web/`
2. 只整理「非網頁依賴」的檔案（如舊文件、日誌）
3. 測試網頁是否正常載入
4. 如有問題，立即復原備份

---

## 📝 更新記錄

| 日期 | 事件 |
|------|------|
| 2026-04-07 | 記錄架構文件（整理失敗後的教訓） |

---

**記住：網頁正常運作最重要，目錄整齊是其次！** 🦈
