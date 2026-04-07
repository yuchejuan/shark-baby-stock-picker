# 🦈 鯊魚寶寶選股系統 - 快速上手

## 🚀 啟動系統

```bash
cd ~/.openclaw/workspace/stock_web
./quick_start.sh
```

開啟瀏覽器：`http://localhost:8080`

---

## 💼 手動更新股價（NEW！）

### 步驟：
1. 點擊「💼 持倉明細」頁籤
2. 點擊右上角「🔄 手動更新股價」按鈕
3. 等待 5-10 秒
4. 看到「✅ 更新完成」→ 股價已更新！

### 注意：
- ✅ 即時從證交所抓取股價
- ✅ 自動計算損益與報酬率
- ⏰ 交易時間外會抓取收盤價
- 🔄 可隨時手動刷新

---

## 📊 其他功能

### 📈 TOP 選股
- 每日精選推薦股票
- 技術面分析（MACD、KD、RSI）
- 一鍵買入模擬交易

### 📜 歷次買賣
- 查看所有交易記錄
- 匯出 CSV 報表
- 計算已實現損益

### 🏭 產業熱度圖
- 視覺化產業表現
- 即時產業輪動分析

---

## 🛑 停止服務

按 `Ctrl+C` 停止網頁伺服器

如需停止 API：
```bash
pkill -f trade_manager
```

---

## ❓ 遇到問題？

### 無法更新股價？
```bash
# 檢查 API 狀態
curl http://localhost:8888/api

# 重啟 API
cd ~/.openclaw/workspace
./trade_manager &
```

### 更多說明
- 完整指南：`UPDATE_PRICE_GUIDE.md`
- 系統架構：`SYSTEM_OVERVIEW_V3.md`

---

🦈 **Have Fun Trading!**
