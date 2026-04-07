# 🦈 鯊魚寶寶選股系統

台灣股市技術分析選股工具，整合 TWSE 即時資料，提供每日選股推薦、持倉管理、配息追蹤等功能。

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)](https://golang.org)

---

## ✨ 特色功能

### 📊 每日選股
- **技術指標組合**：RSI、MACD、KD、均線趨勢
- **訊號評分系統**：0-100 分，自動標示買點/賣點/中性
- **分價格區間推薦**：20-30元、30-40元、40-50元、50-60元
- **Telegram 自動推播**：每日 06:10 自動推送選股結果

### 💰 持倉管理
- **即時股價更新**：整合 TWSE API
- **損益追蹤**：即時計算持倉損益和報酬率
- **交易記錄**：完整的買賣歷史記錄
- **批次管理**：分批買入的成本追蹤

### 📈 配息追蹤
- **高股息追蹤**：追蹤 81 支熱門股票配息
- **殖利率排行**：自動計算並排序
- **配息日曆**：除息日、發放日提醒

### 🔥 產業熱度
- **板塊分析**：追蹤各產業漲跌幅
- **熱點發掘**：找出強勢產業

---

## 🚀 快速開始

### 環境需求

- **Go** 1.18+
- **Python** 3.8+
- **Git**

### 安裝步驟

```bash
# 1. Clone 專案
git clone https://github.com/YOUR_USERNAME/shark-baby-stock-picker.git
cd shark-baby-stock-picker

# 2. 啟動服務
./quick_start.sh

# 3. 開啟瀏覽器
http://localhost:8080/index.html
```

### 手動啟動

```bash
# 網頁伺服器 (Port 8080)
python3 -m http.server 8080 &

# 交易管理 API (Port 8888)
go build -o bin/trade_api trade_manager.go
./bin/trade_api &

# 股票查詢 API (Port 8765) - 可選
go run stock_query_api.go &
```

---

## 📂 專案結構

```
stock_web/
├── index.html              # 主頁（TOP 選股）
├── trade.html              # 交易管理
├── dividend_tracker.html   # 配息追蹤
├── sector_tab.html         # 產業熱度
├── daily_stock_picker_all.go   # 每日選股程式
├── trade_manager.go        # 交易管理 API
├── stock_query_api.go      # 股票查詢 API
├── twse_crawler.go         # TWSE 爬蟲
├── signal_backtest_simple.go   # 訊號回測
├── daily_report.json       # 每日選股結果
├── portfolio.json          # 持倉資料
├── dividend_data.json      # 配息資料
└── ARCHITECTURE.md         # 架構文件
```

---

## 🌐 API 端點

### 網頁伺服器 (Port 8080)
- `GET /index.html` - 主頁
- `GET /trade.html` - 交易管理
- `GET /dividend_tracker.html` - 配息追蹤
- `GET /daily_report.json` - 選股資料
- `GET /portfolio.json` - 持倉資料

### 交易 API (Port 8888)
- `GET /api/trades` - 查詢交易記錄
- `GET /api/holdings` - 查詢持股
- `POST /api/trade/add` - 新增交易
- `DELETE /api/trade/delete` - 刪除交易
- `GET /api/stats` - 統計資料

### 股票查詢 API (Port 8765)
- `GET /api/query?symbol=2330` - 查詢股票技術分析
- `GET /health` - 健康檢查

---

## 📊 技術指標說明

### RSI (相對強弱指標)
- **< 30**：超賣（可能反彈）
- **30-70**：正常區間
- **> 70**：超買（可能回檔）

### MACD (指數平滑異同移動平均線)
- **黃金交叉**：DIF 向上穿越 MACD，買進訊號
- **死亡交叉**：DIF 向下穿越 MACD，賣出訊號

### KD (隨機指標)
- **< 20**：超賣
- **> 80**：超買

### 均線排列
- **多頭排列**：MA5 > MA20 > MA60
- **空頭排列**：MA5 < MA20 < MA60

---

## 🎯 選股訊號評分

系統根據技術指標組合自動評分（0-100）：

| 評分 | 訊號 | 說明 |
|------|------|------|
| 75+ | **買點** ⭐ | RSI偏低 + MACD黃金交叉 + KD超賣 |
| 50-74 | **中性** | 技術面整理 |
| 0-49 | **賣點** ⚠️ | RSI超買 + MACD死亡交叉 + KD超買 |

### 🏆 最佳訊號組合

根據回測驗證（37支股票、30天、48筆樣本）：

**「RSI偏低 + MACD黃金交叉 + KD超賣」**

- ✅ **勝率：64.6%**
- 📈 **開盤平均漲幅：+0.56%**
- ⭐ **最佳案例：台積電 +3.04%**

詳細回測報告：[SIGNAL_BACKTEST_REPORT.md](./SIGNAL_BACKTEST_REPORT.md)

---

## 📅 Cron 排程

### 每日選股推播
```bash
# 每天 06:10 執行選股並推播 Telegram
10 6 * * * cd ~/.openclaw/workspace/stock_web && go run daily_stock_picker_all.go
```

### 配息資料更新
```bash
# 每天 06:00 更新配息資料
0 6 * * * cd ~/.openclaw/workspace/stock_web && ./update_dividend_data.sh
```

---

## 🔧 設定檔

### Telegram 推播（可選）
建立 `.env` 檔案（不會上傳 GitHub）：

```bash
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
```

---

## 📝 使用教學

### 1. 查看每日選股
開啟 `http://localhost:8080/index.html`，會顯示：
- 🏆 今日最佳推薦
- 📊 各價格區間 TOP 1
- 📈 本次選股統計
- 💡 技術面重點提示

### 2. 管理持倉
開啟 `http://localhost:8080/trade.html`，可以：
- 新增交易記錄（買入/賣出）
- 查看持倉明細
- 查看交易歷史
- 更新即時股價

### 3. 追蹤配息
開啟 `http://localhost:8080/dividend_tracker.html`，可以：
- 查看高殖利率股票排行
- 追蹤除息日、發放日
- 計算配息收入

---

## 🧪 訊號回測

驗證選股訊號的有效性：

```bash
# 執行回測
go run signal_backtest_simple.go

# 查看結果
cat SIGNAL_BACKTEST_REPORT.md
```

---

## ⚠️ 注意事項

### 重要限制

1. **資料來源**：使用 TWSE 公開資料，有 5 秒延遲
2. **請勿移動檔案**：HTML/JSON/DB 必須在根目錄（詳見 [ARCHITECTURE.md](./ARCHITECTURE.md)）
3. **投資風險**：本系統僅供參考，投資決策請自行判斷

### 免責聲明

本系統提供的選股資訊僅供參考，不構成投資建議。投資有風險，請謹慎評估。

---

## 📖 文件

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 系統架構說明
- [SIGNAL_BACKTEST_REPORT.md](./SIGNAL_BACKTEST_REPORT.md) - 訊號回測報告
- [QUICKSTART.md](./QUICKSTART.md) - 快速開始指南

---

## 🤝 貢獻

歡迎 Issue 和 Pull Request！

---

## 📄 授權

MIT License

---

## 👨‍💻 作者

**鯊魚寶寶** 🦈  
為 Rick 的 2029 退休計畫而生

---

## 🙏 致謝

- 資料來源：[台灣證券交易所 (TWSE)](https://www.twse.com.tw)
- 靈感來源：Rick 的女兒最愛的歌 🎵

---

**祝投資順利！** 🚀📈
