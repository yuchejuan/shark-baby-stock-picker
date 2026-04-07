# 🦈 快速開始

## 📁 專案結構

```
shark-baby-stock-picker/
├── html/                    ← 🌐 Web 根目錄（瀏覽器開這裡）
│   ├── index.html           ← 主頁面（選股 / 持倉 / 產業）
│   ├── dividend_tracker.html← 高股息追蹤
│   ├── dividend_strategy.html
│   ├── trade.html           ← 交易記錄
│   ├── portfolio_batches.js
│   ├── simulation.js
│   └── test_runner.html     ← 自動化測試頁
├── data/                    ← 📊 資料備份
├── docs/                    ← 📚 說明文件
├── scripts/                 ← 🔧 工具腳本
├── _archive/                ← 📦 舊版備份（不需理會）
│
├── daily_picker_integrated.go ← 每日選股主程式
├── trade_manager.go           ← 交易 API（Port 8888）
├── web_updater.go             ← 更新持倉股價
├── stock_query_api.go         ← 個股查詢（Port 8765）
├── dividend_scraper.go        ← 高股息爬蟲
├── sector_analyzer_twse_full.go← 產業分析
│
├── stock_pool.json            ← 股票池設定（135支）
├── start_all.sh               ← 一鍵啟動
└── test_static.sh             ← 靜態測試
```

---

## 1️⃣ 一鍵啟動

```bash
bash start_all.sh
```

開啟瀏覽器：**http://localhost:8080**

---

## 2️⃣ 每日選股報告

```bash
go run daily_picker_integrated.go
```

輸出至 `html/daily_report.json`，重新整理網頁即可看到結果。

---

## 3️⃣ 更新持倉股價

```bash
go run web_updater.go
```

---

## 4️⃣ 啟動交易 API

```bash
go build -o trade_manager trade_manager.go
./trade_manager
```

API 在 Port 8888，網頁的買入/賣出功能需要這個。

---

## 5️⃣ 個股即時查詢

```bash
go run stock_query_api.go
```

啟動後在網頁查詢任意股票代號（Port 8765）。

---

## 📅 每日建議流程

```bash
# 收盤後（15:30 以後）
go run daily_picker_integrated.go   # 產生選股報告
go run web_updater.go               # 更新持倉股價
```

---

## 🧪 驗證系統正常

```bash
bash test_static.sh   # 靜態分析（不需伺服器）
bash test_api.sh      # API 測試（需先啟動 trade_manager）
```

瀏覽器測試：開啟 `html/test_runner.html`（需 HTTP 伺服器）
