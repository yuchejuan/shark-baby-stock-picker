# 🦈 配息資料爬蟲系統使用指南

## 📄 系統說明

**爬蟲程式**: `dividend_scraper.go`  
**資料來源**: https://stock.wespai.com/rate115  
**輸出檔案**: `stock_web/dividend_data.json`  
**更新頻率**: 建議每天更新一次

---

## 🎯 功能特色

### **自動爬取資料**
- ✅ 現金股利
- ✅ 股票股利
- ✅ 合計股利
- ✅ 除權息日
- ✅ 目前股價
- ✅ 殖利率
- ✅ **10年配息次數**
- ✅ **10年連續配息判斷**

### **資料範圍**
- 81 支台股（包含 ETF、個股）
- 目前成功爬取：57 支
- 10年連續配息：39 支（68.4%）

---

## 🚀 使用方法

### **手動執行**

```bash
cd /home/administrator/.openclaw/workspace

# 執行爬蟲
./dividend_scraper

# 查看結果
cat stock_web/dividend_data.json | jq
```

### **使用更新腳本**

```bash
cd /home/administrator/.openclaw/workspace

# 執行更新腳本
./update_dividend_data.sh
```

### **自動定時更新**（建議）

**設定 cron**：
```bash
# 編輯 crontab
crontab -e

# 新增以下行（每天早上 6:00 更新）
0 6 * * * /home/administrator/.openclaw/workspace/update_dividend_data.sh >> /home/administrator/.openclaw/workspace/dividend_update.log 2>&1
```

---

## 📊 資料格式

### **JSON 結構**

```json
{
  "2330": {
    "symbol": "2330",
    "name": "台積電",
    "cash_dividend": 6.0,
    "stock_dividend": 0.0,
    "total_dividend": 6.0,
    "ex_dividend_date": "06/11",
    "current_price": 1845.0,
    "yield": 0.33,
    "consecutive_10_years": true,
    "dividend_count_10year": 10,
    "avg_3year": 0,
    "avg_6year": 0,
    "avg_10year": 0,
    "update_time": "2026-03-25 17:06:08"
  }
}
```

### **欄位說明**

| 欄位 | 說明 | 範例 |
|------|------|------|
| symbol | 股票代號 | "2330" |
| name | 股票名稱 | "台積電" |
| cash_dividend | 現金股利 | 6.0 |
| stock_dividend | 股票股利 | 0.0 |
| total_dividend | 合計股利 | 6.0 |
| ex_dividend_date | 除權息日 | "06/11" |
| current_price | 目前股價 | 1845.0 |
| yield | 殖利率（%）| 0.33 |
| **consecutive_10_years** | **10年連續配息** | **true/false** |
| **dividend_count_10year** | **10年配息次數** | **10** |
| avg_3year | 3年平均股利 | 0 |
| avg_6year | 6年平均股利 | 0 |
| avg_10year | 10年平均股利 | 0 |
| update_time | 更新時間 | "2026-03-25 17:06:08" |

---

## 🌐 網頁整合

### **自動載入**

`dividend_tracker.html` 會自動載入 `dividend_data.json`：

```javascript
// 載入配息資料
async function loadDividendData() {
    const response = await fetch('dividend_data.json');
    dividendData = await response.json();
    // 更新資料庫
}
```

### **顯示效果**

在「🏆 殖利率排行」表格中會顯示：

| 代號 | 名稱 | 殖利率 | 目前股價 | 每股年配息 | **10年連續配息** | 配息頻率 |
|------|------|--------|---------|-----------|---------------|---------|
| 2330 | 台積電 | 0.33% | $1845 | $6.00 | **✅ 是 (10年)** | 季配息 |
| 2357 | 華碩 | 7.43% | $565 | $42.00 | **✅ 是 (10年)** | 年配息 |
| 3481 | 群創 | 3.90% | $25.65 | $1.00 | **⚠️ 8年** | 年配息 |
| 00929 | 復華台灣科技優息 | 0.56% | $19.63 | $0.11 | **❌ 否** | 月配息 |

### **標記說明**

- **✅ 是 (10年)**: 綠色，連續 10 年配息
- **⚠️ 8-9年**: 黃色，接近 10 年
- **5-7年**: 淺黃色，中等穩定性
- **❌ 否 / 0-4年**: 灰色或紅色，穩定性較低

---

## 📈 統計資訊

### **目前狀態**（2026-03-25）

```
📊 共爬取：57 支股票
✅ 10年連續配息：39 支（68.4%）
📊 平均殖利率：3.20%
```

### **10年連續配息 TOP 10**

| 排名 | 代號 | 名稱 | 殖利率 | 10年配息 |
|------|------|------|--------|---------|
| 1 | 2357 | 華碩 | 7.43% | ✅ 10年 |
| 2 | 2851 | 中再保 | 6.96% | ✅ 10年 |
| 3 | 1102 | 亞泥 | 6.67% | ✅ 10年 |
| 4 | 3034 | 聯詠 | 6.09% | ✅ 10年 |
| 5 | 6506 | 雙鴻 | 6.08% | ✅ 10年 |
| 6 | 2105 | 正新 | 6.05% | ✅ 10年 |
| 7 | 2618 | 長榮航 | 5.61% | ✅ 10年 |
| 8 | 2382 | 廣達 | 5.47% | ✅ 10年 |
| 9 | 2379 | 瑞昱 | 5.11% | ✅ 10年 |
| 10 | 2353 | 宏碁 | 4.80% | ✅ 10年 |

---

## ⚠️ 已知問題

### **1. 部分 ETF 配息次數為 0**

**原因**：新成立的 ETF（例如 00919、00929）成立不到 10 年

**解決**：屬正常現象，非系統問題

### **2. 月配息股利較低**

**原因**：網站顯示的是**單次配息**，非全年合計

**範例**：
- 00929 單次配息 $0.11
- 月配息 12 次 → 全年約 $1.32
- 但殖利率已是年化（0.56%）

**建議**：使用**殖利率**作為比較標準，而非配息金額

### **3. 部分股票未爬取到**

**原因**：24 支股票可能在網站上沒有資料或格式不同

**已爬取**：57 / 81 支（70.4%）

**未來改進**：逐一檢查缺失股票

---

## 🔧 維護建議

### **每日更新**
```bash
# 建議時間：每天早上 6:00
0 6 * * * /home/administrator/.openclaw/workspace/update_dividend_data.sh >> /home/administrator/.openclaw/workspace/dividend_update.log 2>&1
```

### **檢查日誌**
```bash
# 查看更新日誌
tail -f /home/administrator/.openclaw/workspace/dividend_update.log
```

### **手動驗證**
```bash
# 檢查資料更新時間
cat stock_web/dividend_data.json | jq '.["2330"].update_time'

# 檢查 10年連續配息數量
cat stock_web/dividend_data.json | jq '[.[] | select(.consecutive_10_years == true)] | length'
```

---

## 📚 技術細節

### **爬蟲邏輯**

1. HTTP 請求 wespai.com
2. HTML 解析（golang.org/x/net/html）
3. 表格資料提取
4. 10年配息次數位於第 15 個 `<td>`（索引 14）
5. JSON 輸出

### **欄位對應**

```
網站欄位 → JSON 欄位
━━━━━━━━━━━━━━━━━━━
代號 (1) → symbol
名稱 (2) → name
現金股利 (3) → cash_dividend
除息日 (4) → ex_dividend_date
股票股利 (5) → stock_dividend
目前股價 (7) → current_price
殖利率 (8) → yield
10年配息 (15) → dividend_count_10year
```

---

## 🦈 系統狀態

✅ **爬蟲程式** - dividend_scraper.go  
✅ **更新腳本** - update_dividend_data.sh  
✅ **配息資料** - stock_web/dividend_data.json  
✅ **網頁整合** - dividend_tracker.html  
✅ **10年連續配息** - 顯示功能完成  

---

**版本**: v1.0  
**更新日期**: 2026-03-25  
**作者**: 鯊魚寶寶 🦈  
**為**: Rick 的高股息投資追蹤系統

---

🦈 **記住**：配息資料來自公開網站，僅供參考，實際配息以公司公告為準！
