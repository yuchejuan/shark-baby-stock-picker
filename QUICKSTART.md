# 🦈 快速開始 - 5 分鐘上手

## 1️⃣ 一鍵啟動（最簡單）

```bash
cd /home/administrator/.openclaw/workspace/stock_web
./start_server.sh
```

然後開啟瀏覽器訪問: **http://localhost:8080**

---

## 2️⃣ 手動更新資料

### 更新投資組合（取得最新股價）

```bash
cd /home/administrator/.openclaw/workspace
go run web_updater.go
```

### 產生每日選股報告

```bash
cd /home/administrator/.openclaw/workspace
go run daily_report_web.go
```

---

## 3️⃣ 查看網頁

### 方式一：直接開啟（不需伺服器）

```bash
# Windows (WSL)
explorer.exe stock_web/index.html

# Linux
xdg-open stock_web/index.html

# macOS
open stock_web/index.html
```

### 方式二：本地伺服器（推薦）

```bash
cd stock_web
python3 -m http.server 8080
```

訪問: http://localhost:8080

---

## 4️⃣ 自動化排程

### 使用 OpenClaw Cron（推薦）

找鯊魚寶寶說：

```
幫我設定每週一到週五下午 3:30 自動更新股票網頁
```

或手動設定：

```bash
/cron add \
  --schedule "30 15 * * 1-5" \
  --task "cd /home/administrator/.openclaw/workspace && go run web_updater.go && go run daily_report_web.go" \
  --name "股票網頁自動更新"
```

---

## 📱 網頁功能

### 💼 投資組合
- ✅ 即時股價更新
- ✅ 損益計算
- ✅ 報酬率統計
- ✅ 建倉理由記錄

### 📊 每日選股
- ✅ 20-30 元區間精選
- ✅ 30-40 元區間精選
- ✅ 技術評分排名
- ✅ RSI 超賣機會

### 👀 監控清單
- ✅ 台積電、聯發科等權值股
- ✅ 技術面評分
- ✅ 均線排列狀況

### 📈 技術分析
- ✅ 市場漲跌統計
- ✅ 超賣買點提示
- ✅ 風險警告

---

## ⚡ 常見問題

### Q: 為什麼股價都是 0 或沒更新？

A: 可能是以下原因：
1. **休市時段** - Yahoo Finance 在週末/假日不提供資料
2. **網路問題** - 檢查網路連線
3. **API 限制** - 稍後再試

解決方式：**等到週一開盤後再執行**

### Q: 網頁打不開？

A: 確認：
1. 檔案路徑正確 (`stock_web/index.html`)
2. Python 伺服器有在跑 (`python3 -m http.server 8080`)
3. 瀏覽器訪問 `http://localhost:8080`（不是檔案路徑）

### Q: 想新增/修改持倉股票？

A: 編輯 `web_updater.go` 的 `holdings` 變數：

```go
var holdings = []Stock{
    {"2330.TW", "台積電", 100, 580.00, 0, 0, 0, "長期投資"},
    // 新增你的股票...
}
```

然後重新執行 `go run web_updater.go`

---

## 🎯 每日流程建議

### 📅 週一至週五（交易日）

**下午 3:30 收盤後:**

```bash
# 1. 更新資料
cd /home/administrator/.openclaw/workspace
go run web_updater.go
go run daily_report_web.go

# 2. 查看網頁
cd stock_web
./start_server.sh
```

**或者設定自動化，完全不用動手！**

### 📅 週末

- 檢視週報表現
- 規劃下週操作
- 學習技術分析

---

## 🦈 需要幫助？

直接問鯊魚寶寶：

- "幫我更新股票網頁"
- "今天選股報告如何？"
- "我的投資組合現在損益多少？"
- "設定自動更新股票網頁"

---

## 🎉 恭喜！

你已經擁有一個完整的台股分析網頁系統了！

**投資愉快，穩健獲利！** 🚀
