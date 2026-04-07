# 🦈 鯊魚寶寶選股系統 - 網頁版

## 📋 功能特色

✅ **投資組合追蹤** - 即時顯示持倉、損益、報酬率  
✅ **每日選股報告** - 20-30 / 30-40 元區間精選股票  
✅ **權值股監控** - 台積電、聯發科等 10 支重點股票  
✅ **技術分析** - RSI、均線、超賣機會分析  
✅ **自動更新** - 串接 Yahoo Finance API 即時股價  

---

## 🚀 快速開始

### 1. 更新投資組合資料

```bash
cd /home/administrator/.openclaw/workspace
go run web_updater.go
```

**輸出**: `stock_web/portfolio.json` （網頁會自動讀取）

### 2. 開啟網頁

**方式一：直接用瀏覽器開啟**
```bash
# Linux
xdg-open stock_web/index.html

# macOS
open stock_web/index.html

# Windows (WSL)
explorer.exe stock_web/index.html
```

**方式二：啟動本地伺服器**
```bash
cd stock_web
python3 -m http.server 8080
```

然後開啟瀏覽器訪問: `http://localhost:8080`

---

## 📊 網頁分頁說明

### 💼 投資組合
- 顯示 8 支持倉股票的即時損益
- 總成本、市值、報酬率統計
- 建倉理由與日期記錄

### 📊 每日選股
- **20-30 元區間**: 選擇最多，適合小資族
- **30-40 元區間**: 中價位股票
- 顯示評分、RSI、技術訊號

### 👀 監控清單
- 台積電 (2330)、聯發科 (2454) 等權值股
- 技術面評分與均線排列狀況

### 📈 技術分析
- 市場漲跌統計
- RSI < 35 超賣機會列表
- 即時技術指標分析

---

## ⏰ 自動化排程

### 建議使用 OpenClaw Cron 定時更新

```bash
# 每週一到週五 15:30 更新（收盤後）
/cron add --schedule "30 15 * * 1-5" \
  --task "執行台股投資組合更新：cd /home/administrator/.openclaw/workspace && go run web_updater.go" \
  --name "股票網頁更新"
```

或者手動設定 Linux crontab:
```bash
# 每天 15:30 更新
30 15 * * 1-5 cd /home/administrator/.openclaw/workspace && go run web_updater.go
```

---

## 📁 檔案結構

```
stock_web/
├── index.html          # 主網頁（已完成）
├── portfolio.json      # 投資組合資料（自動生成）
└── README.md           # 本說明文件

web_updater.go          # 資料更新程式
```

---

## 🔧 進階功能

### 自訂持倉股票

編輯 `web_updater.go` 的 `holdings` 變數：

```go
var holdings = []Stock{
    {"2330.TW", "台積電", 1000, 580.00, 0, 0, 0, "定期定額"},
    {"2454.TW", "聯發科", 500, 1200.00, 0, 0, 0, "5G 題材"},
    // 新增你的股票...
}
```

### 添加每日選股資料

建立 `stock_web/daily_picks.json`，網頁會自動讀取並顯示。

---

## ❗ 注意事項

1. **交易時段**: Yahoo Finance API 在休市時可能回傳舊資料
2. **更新頻率**: 建議收盤後（15:30）執行，盤中頻繁更新意義不大
3. **資料延遲**: 免費 API 可能有 15-20 分鐘延遲
4. **網路需求**: 需要連網才能取得最新股價

---

## 🦈 維護者

**鯊魚寶寶 (Baby Shark)**  
協助 Rick 達成 2029 年退休計畫

**建立日期**: 2026-03-15  
**投資組合建倉日**: 2026-03-13

---

## 📞 支援

遇到問題？直接跟鯊魚寶寶說！🦈
```
