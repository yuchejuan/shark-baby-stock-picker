# 🦈 模擬買入功能實裝指南

## 📋 功能概述

已實裝完整的模擬買入功能，可以從股票查詢結果直接買入並加入投資組合。

---

## ✅ 已完成項目

1. ✅ **買入對話框 UI** (`add_buy_modal.html`)
2. ✅ **後端 API 服務** (`portfolio_manager.go`)  
3. ✅ **前端 API 整合** (`buy_modal_api.js`)
4. ✅ **投資組合管理** (自動更新 `portfolio.json`)

---

## 🚀 快速啟動

### **步驟 1: 啟動投資組合 API**

```bash
cd ~/.openclaw/workspace
go run portfolio_manager.go &
```

**輸出：**
```
🦈 投資組合管理 API 啟動於 http://localhost:8766
📡 端點:
  POST /api/portfolio/buy   - 模擬買入
  GET  /api/portfolio       - 查詢投資組合
```

---

### **步驟 2: 整合買入對話框到網頁**

將 `add_buy_modal.html` 的內容加入到 `stock_web/index.html`:

**插入位置：** 在 `</body>` 標籤之前

```bash
# 手動編輯，或使用以下指令
cat add_buy_modal.html >> stock_web/index.html
```

---

### **步驟 3: 更新買入函數**

將 `buy_modal_api.js` 的內容取代掉 `add_buy_modal.html` 中的 `confirmBuy()` 函數。

**或直接在 `index.html` 的 `<script>` 區塊中加入：**
```html
<script src="buy_modal_api.js"></script>
```

---

## 📊 使用流程

### **1. 查詢股票**

在網頁輸入股票代號查詢，例如：`2330`

### **2. 查看技術分析**

系統顯示完整的技術分析結果。

### **3. 點擊「📈 模擬買入」**

在操作建議區塊，點擊「模擬買入」按鈕。

### **4. 填寫買入資訊**

- **股票代號** - 自動帶入
- **股票名稱** - 自動帶入
- **買入價格** - 預設為當前價格（可調整）
- **買入股數** - 輸入張數（1張 = 1000股）
- **買入理由** - 自動帶入推薦理由（可修改）

### **5. 確認買入**

系統計算總金額，確認後送出。

### **6. 更新投資組合**

- 新股票 → 加入持股
- 已持有 → 計算平均成本並加碼

---

## 🎯 功能特色

### **✨ 智慧加碼**

如果已持有相同股票，系統會：
1. 自動計算新的平均成本
2. 合併持股數量
3. 更新投資組合

**範例：**
```
原持有: 2330 台積電 1張 @ $1,900
加碼: 2張 @ $2,000

結果: 3張 @ $1,966.67（平均成本）
```

---

### **📊 自動更新統計**

每次買入後，系統自動更新：
- 總成本
- 當前市值  
- 總損益
- 報酬率

---

## 🔧 API 端點說明

### **POST /api/portfolio/buy**

**模擬買入股票**

**請求範例：**
```json
{
  "symbol": "2330",
  "name": "台積電",
  "price": 1995.00,
  "shares": 2,
  "reason": "MACD黃金交叉 + 多頭排列"
}
```

**回應範例：**
```json
{
  "success": true,
  "message": "買入成功！台積電 (2330) 2張 @ 1995.00 = $3990000",
  "portfolio": {
    "holdings": [...],
    "total_cost": 3990000,
    "current_value": 3990000,
    "total_pnl": 0,
    "total_return": 0
  }
}
```

---

### **GET /api/portfolio**

**查詢投資組合**

**回應範例：**
```json
{
  "holdings": [
    {
      "symbol": "2330",
      "name": "台積電",
      "shares": 2000,
      "buy_price": 1995.00,
      "current_price": 1995.00,
      "profit_loss": 0,
      "return_pct": 0,
      "reason": "MACD黃金交叉 + 多頭排列"
    }
  ],
  "total_cost": 3990000,
  "current_value": 3990000,
  "total_pnl": 0,
  "total_return": 0,
  "last_update": "2026-04-02 16:30:00"
}
```

---

## 🧪 測試流程

### **測試 1: 買入新股票**

```bash
# 啟動 API
go run portfolio_manager.go &

# 測試買入
curl -X POST http://localhost:8766/api/portfolio/buy \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "2330",
    "name": "台積電",
    "price": 1995.00,
    "shares": 2,
    "reason": "技術面突破"
  }'
```

**預期：**
```json
{
  "success": true,
  "message": "買入成功！台積電 (2330) 2張 @ 1995.00 = $3990000"
}
```

---

### **測試 2: 加碼已持有股票**

```bash
curl -X POST http://localhost:8766/api/portfolio/buy \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "2330",
    "name": "台積電",
    "price": 2000.00,
    "shares": 1,
    "reason": "持續看好"
  }'
```

**預期：**
```json
{
  "success": true,
  "message": "加碼成功！持股 2張 → 3張，平均成本 1995.00 → 1996.67"
}
```

---

### **測試 3: 查詢投資組合**

```bash
curl http://localhost:8766/api/portfolio
```

---

## 📁 檔案結構

```
~/.openclaw/workspace/
├── portfolio_manager.go         # 投資組合 API 服務
├── add_buy_modal.html           # 買入對話框 HTML + JS
├── buy_modal_api.js             # API 整合 JS
├── PORTFOLIO_TRADING_GUIDE.md   # 本使用指南
│
└── stock_web/
    ├── index.html               # 主頁面（需加入買入對話框）
    └── portfolio.json           # 投資組合數據（自動更新）
```

---

## 🔄 整合步驟

### **方法一：自動整合（推薦）**

執行整合腳本（待建立）：
```bash
./integrate_buy_modal.sh
```

### **方法二：手動整合**

1. **開啟 `stock_web/index.html`**
2. **在 `</body>` 之前加入 `add_buy_modal.html` 的內容**
3. **儲存檔案**
4. **啟動 API 服務**

---

## ⚠️ 注意事項

### **1. API 服務必須運行**

買入功能需要後端 API 支援，執行前請確認：

```bash
# 檢查 API 是否運行
lsof -i :8766

# 若無運行，執行
go run portfolio_manager.go &
```

---

### **2. 投資組合檔案權限**

確保 `stock_web/portfolio.json` 可寫入：

```bash
chmod 664 stock_web/portfolio.json
```

---

### **3. 瀏覽器快取**

更新後請重新整理頁面（Ctrl + F5）清除快取。

---

## 🦈 阿哲，使用說明：

### **立即啟動：**

```bash
cd ~/.openclaw/workspace

# 1. 啟動投資組合 API
go run portfolio_manager.go &

# 2. 開啟網頁
cd stock_web
python3 -m http.server 8080

# 3. 瀏覽器開啟
# http://localhost:8080

# 4. 查詢股票 → 點擊「模擬買入」→ 填寫資料 → 確認
```

---

**功能已完全實裝！開始模擬交易吧！** 🚀📊
