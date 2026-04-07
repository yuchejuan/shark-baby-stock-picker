# 🔄 手動更新股價功能 - 使用指南

## ✨ 功能說明

在「💼 持倉明細」頁面新增了「🔄 手動更新股價」按鈕，可即時從證交所 API 抓取最新股價並更新持倉資料。

---

## 📋 前置需求

### 1️⃣ 確保 API 伺服器運行

**方式 A：使用完整啟動腳本（推薦）**
```bash
cd ~/.openclaw/workspace
./stock_web/start_all.sh
```

**方式 B：手動啟動 API**
```bash
cd ~/.openclaw/workspace
go build -o trade_manager trade_manager.go
./trade_manager &
```

### 2️⃣ 檢查服務狀態
```bash
# 檢查 API（Port 8888）
curl http://localhost:8888/api

# 檢查網頁伺服器（Port 8080）
curl http://localhost:8080
```

---

## 🎯 使用步驟

### 1. 開啟網頁
```
http://localhost:8080
```

### 2. 切換到「💼 持倉明細」頁面

### 3. 點擊右上角「🔄 手動更新股價」按鈕

### 4. 等待更新完成
- 按鈕會顯示「⏳ 更新中...」
- 約 5-10 秒後顯示「✅ 更新完成」
- 持倉資料自動刷新

---

## 🔍 技術細節

### 執行流程
```
前端按鈕點擊
    ↓
POST /api/holdings/update-prices
    ↓
執行 web_updater.go
    ↓
從證交所 API 抓取即時股價
    ↓
更新 portfolio.json
    ↓
前端重新載入持倉資料
    ↓
顯示最新股價與損益
```

### API 端點
- **URL:** `POST http://localhost:8888/api/holdings/update-prices`
- **回應格式:**
  ```json
  {
    "success": true,
    "message": "股價已更新",
    "output": "更新日誌..."
  }
  ```

### 更新內容
- ✅ 所有持股的即時股價
- ✅ 持倉損益與報酬率
- ✅ 總市值與總損益統計

---

## 🧪 測試功能

執行測試腳本：
```bash
cd ~/.openclaw/workspace/stock_web
./test_update.sh
```

---

## ❗ 常見問題

### Q1: 點擊後沒有反應？
**A:** 檢查 API 伺服器是否運行：
```bash
curl http://localhost:8888/api
```
如果沒有回應，執行：
```bash
cd ~/.openclaw/workspace
./trade_manager &
```

### Q2: 更新失敗顯示錯誤？
**A:** 檢查 API log：
```bash
tail -f ~/.openclaw/workspace/trade_manager.log
```

### Q3: 股價沒有變動？
**A:** 可能原因：
- 非交易時間（週末或收盤後）
- 證交所 API 回傳快取資料
- 該股票停牌或資料錯誤

---

## 🛠️ 故障排除

### 重新編譯 API 伺服器
```bash
cd ~/.openclaw/workspace
pkill -f trade_manager
go build -o trade_manager trade_manager.go
./trade_manager &
```

### 手動更新股價
```bash
cd ~/.openclaw/workspace
go run web_updater.go
```

### 檢查 portfolio.json 更新時間
```bash
ls -lh ~/.openclaw/workspace/stock_web/portfolio.json
```

---

## 📝 程式碼位置

- **API 端點:** `trade_manager.go` → `updatePricesHandler()`
- **前端函數:** `index.html` → `updatePortfolioPrices()`
- **股價更新:** `web_updater.go`
- **分批次載入:** `portfolio_batches.js` → `loadPortfolio()`

---

## 🦈 更新日誌

- **2026-03-31:** 新增手動更新股價功能
  - 新增 API 端點 `/api/holdings/update-prices`
  - 新增前端更新按鈕與動畫效果
  - 整合 `web_updater.go` 即時抓取股價

---

🦈 **鯊魚寶寶選股系統**
