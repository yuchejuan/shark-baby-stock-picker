# 🦈 Rick 交易管理系統使用指南

## 🎯 系統概述

這是一個**完整的股票交易記錄與損益追蹤系統**，專為 Rick 的退休投資計畫設計。

### 核心功能

1. **買賣操作** - 記錄每筆買入/賣出交易
2. **永久保存** - 所有交易記錄存入 SQLite 資料庫（`stock_trades.db`）
3. **自動計算** - 持股成本、損益、報酬率
4. **歷史查詢** - 隨時查看過去所有交易
5. **統計分析** - 勝率、總損益、已/未實現損益

---

## 🚀 快速啟動

### 方法一：完整啟動（推薦）

```bash
cd /home/administrator/.openclaw/workspace/stock_web
./start_all.sh
```

這會同時啟動：
- 🌐 網頁介面 (Port 8080)
- 🔌 API 服務 (Port 8888)

### 方法二：分別啟動

```bash
# 終端機 1: 啟動 API
cd /home/administrator/.openclaw/workspace
./trade_manager

# 終端機 2: 啟動網頁
cd /home/administrator/.openclaw/workspace/stock_web
python3 -m http.server 8080
```

---

## 📖 使用方式

### 1. 訊問交易管理介面

瀏覽器開啟：**http://localhost:8080/trade.html**

### 2. 新增買入交易

1. 點選「💰 買賣操作」分頁
2. 選擇「買入」
3. 填寫資料：
   - **股票代號**: 例如 2330
   - **股票名稱**: 例如 台積電
   - **股數**: 例如 1000
   - **價格**: 例如 500.50
   - **日期**: 選擇交易日期
   - **備註**: 可選，記錄買入理由

4. 點選「💰 確認交易」

### 3. 新增賣出交易

1. 同樣在「💰 買賣操作」分頁
2. 選擇「賣出」
3. 填寫相同資料（代號、名稱、股數、價格、日期）
4. 系統會自動計算損益

### 4. 查看當前持股

點選「📁 當前持股」分頁，會顯示：
- 每支股票的總股數
- 平均買入成本
- 總成本
- 未實現損益（需手動更新現價）

### 5. 查看交易歷史

點選「📜 交易歷史」分頁，可以：
- 查看所有交易記錄
- 按股票代號篩選
- 刪除錯誤的記錄

---

## 📊 實際範例

### 範例 1: 買入台積電

```
交易類型: 買入
股票代號: 2330
股票名稱: 台積電
股數: 2000
價格: 500.00
日期: 2026-03-25
備註: RSI 超賣，技術面轉強
```

**結果**: 成本 1,000,000 元

### 範例 2: 部分賣出

```
交易類型: 賣出
股票代號: 2330
股票名稱: 台積電
股數: 1000
價格: 520.00
日期: 2026-04-15
備註: 獲利 4% 出場
```

**結果**: 
- 獲利: 20,000 元 (1000股 × 20元)
- 剩餘: 1000股，平均成本仍為 500 元

---

## 🔧 進階功能

### API 端點

系統提供 RESTful API，可以用程式自動化操作：

```bash
# 查詢所有交易
curl http://localhost:8888/api/trades

# 查詢特定股票
curl http://localhost:8888/api/trades?symbol=2330

# 查詢當前持股
curl http://localhost:8888/api/holdings

# 新增交易（POST JSON）
curl -X POST http://localhost:8888/api/trade/add \
  -H "Content-Type: application/json" \
  -d '{
    "symbol": "2330",
    "name": "台積電",
    "type": "buy",
    "shares": 1000,
    "price": 500.00,
    "date": "2026-03-25T00:00:00Z",
    "note": "買入理由"
  }'

# 刪除交易
curl -X DELETE http://localhost:8888/api/trade/delete?id=1
```

---

## 📁 資料庫結構

### 檔案位置
`/home/administrator/.openclaw/workspace/stock_trades.db`

### 資料表: `trades`

| 欄位 | 類型 | 說明 |
|------|------|------|
| id | INTEGER | 交易編號（自動遞增）|
| symbol | TEXT | 股票代號 |
| name | TEXT | 股票名稱 |
| type | TEXT | 交易類型（buy/sell）|
| shares | INTEGER | 股數 |
| price | REAL | 價格 |
| amount | REAL | 金額（自動計算）|
| date | DATETIME | 交易日期 |
| note | TEXT | 備註 |
| created_at | DATETIME | 記錄建立時間 |

### 查詢範例（SQLite）

```bash
# 進入資料庫
sqlite3 stock_trades.db

# 查看所有交易
SELECT * FROM trades ORDER BY date DESC;

# 查看特定股票
SELECT * FROM trades WHERE symbol = '2330';

# 計算總損益（簡化）
SELECT 
  SUM(CASE WHEN type='buy' THEN -amount ELSE amount END) as total_pnl 
FROM trades;

# 離開
.quit
```

---

## 🛡️ 安全性與備份

### 自動備份

建議定期備份資料庫：

```bash
# 手動備份
cp stock_trades.db stock_trades_backup_$(date +%Y%m%d).db

# 自動備份（加入 crontab）
0 0 * * * cp /home/administrator/.openclaw/workspace/stock_trades.db /home/administrator/backups/stock_trades_$(date +\%Y\%m\%d).db
```

### 匯出 CSV

```bash
sqlite3 -header -csv stock_trades.db "SELECT * FROM trades;" > trades_export.csv
```

---

## 🎯 整合現有系統

### 與選股系統整合

系統已經整合到主網頁（`index.html`），點選「💰 交易管理」按鈕即可進入。

### 與每日報告整合

可以在每日選股後，直接透過交易管理系統記錄實際操作，形成完整的閉環：

1. **每天 06:00** - 系統產生選股報告
2. **開盤後** - Rick 決定是否買入
3. **成交後** - 記錄到交易系統
4. **持續追蹤** - 查看損益表現

---

## 🐛 常見問題

### Q: API 無法連線？

```bash
# 檢查 API 是否運行
curl http://localhost:8888/api/trades

# 重新啟動
cd /home/administrator/.openclaw/workspace
killall trade_manager
./trade_manager &
```

### Q: 網頁顯示「載入中」不動？

1. 確認 API 正在運行（Port 8888）
2. 檢查瀏覽器 Console (F12) 是否有錯誤
3. 確認防火牆沒有封鎖

### Q: 如何修改錯誤的交易？

目前系統不支援編輯，只能：
1. 在「📜 交易歷史」刪除錯誤記錄
2. 重新新增正確的交易

### Q: 持股損益如何更新？

目前需要手動更新，未來版本會整合：
- 自動抓取即時股價（TWSE API）
- 自動計算未實現損益

---

## 🚀 未來改進計畫

- [ ] 自動更新現價（整合 `web_updater.go`）
- [ ] 圖表視覺化（損益曲線、持股分布）
- [ ] 匯出 Excel 報表
- [ ] 手機 App（PWA）
- [ ] 多帳戶支援
- [ ] 股息記錄
- [ ] 配對演算法優化（FIFO/LIFO）

---

## 📞 技術支援

有任何問題，直接問鯊魚寶寶！🦈

**系統位置**: `/home/administrator/.openclaw/workspace/`
- `trade_manager.go` - API 伺服器原始碼
- `stock_web/trade.html` - 交易介面
- `stock_trades.db` - 資料庫

---

**版本**: v1.0  
**建立日期**: 2026-03-25  
**作者**: 鯊魚寶寶 🦈  
**為**: Rick 的 2029 退休計畫
