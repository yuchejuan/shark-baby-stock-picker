# 🦈 即時股價更新系統使用指南

## 🎯 功能概述

**自動化股價更新服務**，讓你的持倉明細保持最新！

---

## ✨ 核心功能

1. ✅ **自動更新** - 每 20 分鐘自動爬取最新股價
2. ✅ **盤中運作** - 只在週一至週五 09:00-14:00 運行
3. ✅ **即時損益** - 自動計算最新的損益和報酬率
4. ✅ **背景執行** - 不影響其他操作
5. ✅ **網頁同步** - 網頁自動刷新（盤中時間）

---

## 🚀 啟動服務

### 方法一：一鍵啟動（推薦）

```bash
cd /home/administrator/.openclaw/workspace
./start_realtime_updater.sh
```

### 方法二：手動啟動

```bash
cd /home/administrator/.openclaw/workspace
go build -o realtime_price_updater realtime_price_updater.go
nohup ./realtime_price_updater > realtime_updater.log 2>&1 &
```

---

## 📊 運作邏輯

### 時間控制

```
週一至週五：
09:00-14:00 → 每 20 分鐘更新一次
14:00-23:59 → 休息（等待明天 09:00）
00:00-08:59 → 休息（等待今天 09:00）

週六、週日：
全天休息（等待下週一 09:00）
```

### 更新流程

```
1. 讀取 portfolio.json（持股資料）
   ↓
2. 提取所有股票代號（2353, 2838, ...）
   ↓
3. 向 TWSE API 請求即時股價
   ↓
4. 更新每支股票的現價
   ↓
5. 重新計算損益與報酬率
   ↓
6. 寫回 portfolio.json
   ↓
7. 網頁自動重新載入（如果開著）
```

---

## 🌐 網頁自動刷新

網頁端也會自動重新載入！

### 刷新邏輯

```javascript
盤中時間（週一至週五 09:00-14:00）：
- 整點（09:00, 10:00, 11:00, 12:00, 13:00）
- 20 分（09:20, 10:20, 11:20, 12:20, 13:20）
- 40 分（09:40, 10:40, 11:40, 12:40, 13:40）

→ 自動重新載入持倉明細
```

**你不需要手動刷新網頁！**

---

## 🔧 管理指令

### 檢查服務狀態

```bash
ps aux | grep realtime_price_updater
```

**輸出範例**：
```
adminis+  200704  0.0  0.1 1234 5678 ?  S  11:43  0:00 ./realtime_price_updater
```

如果有輸出 → 服務運行中 ✅  
如果沒有輸出 → 服務未運行 ❌

---

### 查看即時日誌

```bash
tail -f /home/administrator/.openclaw/workspace/realtime_updater.log
```

**按 Ctrl+C 停止查看**

**日誌範例**：
```
[11:43:29] 🔄 開始更新股價...
  📊 2353 (宏碁): $27.50
  📊 2838 (聯邦銀): $20.15
  💰 總成本: $230100 | 市值: $231500 | 損益: +1400 (0.61%)
[11:43:29] ✅ 股價更新完成
[11:43:29] ⏳ 下次更新時間: 12:03:29
```

---

### 停止服務

```bash
pkill -f realtime_price_updater
```

**確認已停止**：
```bash
ps aux | grep realtime_price_updater
```

應該沒有任何輸出（或只有 grep 本身）

---

### 重新啟動服務

```bash
# 先停止
pkill -f realtime_price_updater

# 再啟動
cd /home/administrator/.openclaw/workspace
./start_realtime_updater.sh
```

---

## 📱 實際使用範例

### 場景一：早上開盤

```
08:55 - 你打開網頁，看到昨天收盤價
09:00 - 開盤！即時更新服務自動啟動
09:00 - 第一次更新股價
09:20 - 第二次更新
09:40 - 第三次更新
...
```

### 場景二：盤中監控

```
10:30 - 你正在看持倉明細
10:40 - 網頁自動刷新，顯示最新損益
11:00 - 又自動刷新一次
11:20 - 又自動刷新一次
...
```

**你不需要做任何事，一切自動！**

### 場景三：收盤後

```
13:40 - 最後一次更新
14:00 - 收盤，服務進入休眠
14:10 - 即使網頁開著，也不會再更新
...
明天 09:00 - 服務自動恢復運作
```

---

## 📊 資料來源

### TWSE 即時股價 API

**API 端點**：
```
https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_XXXX.tw
```

**範例**（查詢台積電 2330）：
```bash
curl "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_2330.tw"
```

**回應範例**：
```json
{
  "msgArray": [{
    "c": "2330",
    "n": "台積電",
    "z": "500.00",  ← 成交價
    "y": "495.00"   ← 昨收價
  }]
}
```

---

## ⚠️ 注意事項

### 1. API 限制

- TWSE API 有請求頻率限制
- 系統已加入延遲（每支股票間隔 200ms）
- 如果持股太多（> 20 支），可能需要調整

### 2. 網路問題

如果網路不穩定，可能出現：
```
⚠️  2353 (宏碁): 無法取得價格，使用舊價格 $27.15
```

**不影響系統運作**，下次更新會繼續嘗試。

### 3. 非交易時間

- 週末、國定假日不會更新
- 14:00 後不會更新
- 09:00 前不會更新

### 4. 資料延遲

- TWSE API 可能有 1-2 分鐘延遲
- 非即時逐筆，而是「接近即時」

---

## 🔍 故障排除

### 問題一：服務沒有啟動

**檢查**：
```bash
ps aux | grep realtime_price_updater
```

**解決**：
```bash
cd /home/administrator/.openclaw/workspace
./start_realtime_updater.sh
```

---

### 問題二：股價沒有更新

**可能原因**：
1. 非盤中時間（週末、收盤後）
2. 網路問題
3. TWSE API 異常

**檢查日誌**：
```bash
tail -50 realtime_updater.log
```

---

### 問題三：網頁沒有自動刷新

**可能原因**：
1. 非盤中時間
2. 瀏覽器快取

**解決**：
1. 確認現在是 09:00-14:00（週一至週五）
2. 清除瀏覽器快取（Ctrl+Shift+R）
3. 手動點「持倉明細」分頁重新載入

---

### 問題四：CPU 使用率過高

**不太可能發生**，因為：
- 只在盤中時間運作
- 20 分鐘才更新一次
- 大部分時間在休眠

**如果真的發生**：
```bash
# 停止服務
pkill -f realtime_price_updater

# 檢查日誌
tail -100 realtime_updater.log

# 回報問題給鯊魚寶寶
```

---

## 📈 系統效益

### Before（之前）

```
❌ 需要手動運行 web_updater.go
❌ 每天只更新一次（15:30）
❌ 盤中時間看不到即時損益
❌ 需要記得執行更新
```

### After（現在）

```
✅ 全自動，無需手動操作
✅ 盤中每 20 分鐘更新一次
✅ 即時看到損益變化
✅ 網頁自動刷新
✅ 背景執行，不影響其他操作
```

---

## 🎯 進階設定

### 修改更新間隔

編輯 `realtime_price_updater.go`：

```go
// 原本：20 分鐘
time.Sleep(20 * time.Minute)

// 改為：10 分鐘
time.Sleep(10 * time.Minute)
```

重新編譯並重啟：
```bash
pkill -f realtime_price_updater
cd /home/administrator/.openclaw/workspace
./start_realtime_updater.sh
```

---

### 修改運作時間

編輯 `realtime_price_updater.go`：

```go
// 原本：09:00-14:00
if hour < 9 || (hour >= 14 && minute > 0) {

// 改為：08:30-14:30
if hour < 8 || (hour == 8 && minute < 30) || (hour >= 14 && minute > 30) {
```

---

## 📚 檔案清單

| 檔案 | 說明 | 位置 |
|------|------|------|
| `realtime_price_updater.go` | 主程式 | `/home/administrator/.openclaw/workspace/` |
| `start_realtime_updater.sh` | 啟動腳本 | `/home/administrator/.openclaw/workspace/` |
| `realtime_updater.log` | 運行日誌 | `/home/administrator/.openclaw/workspace/` |
| `portfolio.json` | 持股資料（會被更新）| `/home/administrator/.openclaw/workspace/stock_web/` |

---

## 🦈 快速指令速查

```bash
# 啟動服務
./start_realtime_updater.sh

# 檢查狀態
ps aux | grep realtime_price_updater

# 查看日誌
tail -f realtime_updater.log

# 停止服務
pkill -f realtime_price_updater

# 重新啟動
pkill -f realtime_price_updater && ./start_realtime_updater.sh
```

---

**版本**: v1.0  
**更新日期**: 2026-03-25  
**作者**: 鯊魚寶寶 🦈  
**為**: Rick 的即時股價追蹤系統

---

🦈 **記住**：服務會自動在背景運行，你什麼都不用做！只要打開網頁，就能看到最新的持倉損益！
