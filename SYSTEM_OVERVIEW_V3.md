# 🦈 鯊魚寶寶台股系統 - 完整總覽 V3.0

**更新時間**: 2026-03-27  
**版本**: v3.0  
**為**: Rick 的台股投資管理系統

---

## 📊 **系統架構總覽**

```
┌─────────────────────────────────────────────────────────────┐
│                    🦈 鯊魚寶寶台股系統 V3.0                   │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 選股系統     │  │ 交易管理系統 │  │ 高股息追蹤  │      │
│  │ 81支股票池   │  │ 分批次持股   │  │ 配息資料    │      │
│  │ 20-60元區間  │  │ FIFO邏輯     │  │ 10年連配    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         ↓                  ↓                  ↓              │
│  ┌──────────────────────────────────────────────────┐      │
│  │            前端網頁 (http://localhost:8080)       │      │
│  └──────────────────────────────────────────────────┘      │
│         ↓                  ↓                  ↓              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 選股 Go      │  │ 交易 Go      │  │ 配息 Go      │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         ↓                  ↓                  ↓              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │daily_report  │  │stock_trades  │  │dividend_data │      │
│  │.json         │  │.db           │  │.json         │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                               │
│  ┌─────────────────────────────────────────────────┐        │
│  │         自動排程 (Linux Cron + OpenClaw Cron)   │        │
│  │  Linux Cron:                                     │        │
│  │  • 06:00 - 每日選股 (81支，20-60元)             │        │
│  │  • 06:00 - 配息資料更新                          │        │
│  │  • 15:00 - 股票快照記錄 (週一至週五)            │        │
│  │  • 15:30 - 網頁資料更新 (週一至週五)            │        │
│  │                                                   │        │
│  │  OpenClaw Cron (需 Telegram 推播):               │        │
│  │  • 15:00 - AI台股綜合評分                        │        │
│  │  • 16:00 - 台股收盤摘要                          │        │
│  │  • 19:00 - 學習提醒                              │        │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎯 **核心功能清單**

### **1. 選股系統（V3.0）** 📊

#### **功能**
- ✅ **81 支股票池**（ETF 14 + 個股 67）
- ✅ **4 個價格區間**：20-30元 / 30-40元 / 40-50元 / 50-60元
- ✅ **6 大技術指標**：RSI、MACD、KD、MA、布林通道、OBV
- ✅ **100分制評分系統**
- ✅ **各區間 TOP 3 推薦**
- ✅ **最佳標的推薦**
- ✅ **股票池獨立設定檔**（stock_pool.json）

#### **股票池組成**
```
ETF（14支）：
- 市值型：0050、006208、00632R、00692
- 高股息：00919、00929、00918、00878、0056
- 主題型：00881、00891、00895、00896、00701

個股（67支）：
- 金融股：18支（臺企銀、遠東銀、中信金、國泰金等）
- 權值股：10支（台積電、鴻海、聯發科、台達電等）
- 電子股：7支（宏碁、華碩、廣達、仁寶等）
- 傳產股：8支（台泥、亞泥、統一、長榮航等）
- 中小型股：11支（台塑化、彰銀、京城銀等）
- AI相關：5支（日月光、聯電、聯詠、創意、藥華藥）
- 電力相關：3支（宏捷科、緯穎、同欣電）
- 通訊相關：4支（微星、英業達、緯創、和碩）
- 其他：1支（世紀鋼）
```

#### **資料檔案**
- **程式**：`daily_picker_integrated.go`（V3.0）
- **設定**：`stock_pool.json`（獨立管理）
- **輸出**：`stock_web/daily_report.json`

#### **更新頻率**
- **自動**：每天 06:00（Linux Cron）
- **手動**：`bash cron/daily_stock_picker.sh`

#### **執行時間**
- **81 支股票**：約 8-10 分鐘

---

### **2. 交易管理系統** 💼

#### **功能**
- ✅ 模擬買入/賣出
- ✅ 分批次持股管理
- ✅ FIFO 賣出邏輯
- ✅ 即時損益計算
- ✅ 歷史交易記錄
- ✅ CSV 匯出

#### **資料檔案**
- **程式**：`trade_manager.go`（API 伺服器，Port 8888）
- **資料庫**：`stock_trades.db`（SQLite）
- **快照**：`portfolio.json`（根目錄 & stock_web/）

#### **更新頻率**
- 即時（手動交易觸發）

---

### **3. 高股息追蹤系統** 💰

#### **功能**
- ✅ 81 支股票殖利率排行
- ✅ 持股管理
- ✅ 自動計算年股息
- ✅ 健保補充保費提醒（20,000 元門檻）
- ✅ 10 年連續配息顯示
- ✅ 即時股價更新（24 小時快取）

#### **資料檔案**
- **前端**：`stock_web/dividend_tracker.html`
- **資料**：`stock_web/dividend_data.json`
- **持股**：LocalStorage（瀏覽器）

#### **更新頻率**
- **配息資料**：每天 06:00（Linux Cron）
- **即時股價**：24 小時快取

---

### **4. 配息資料爬蟲** 🕷️

#### **功能**
- ✅ 自動爬取 wespai.com
- ✅ 81 支股票配息資料
- ✅ 10 年配息次數
- ✅ 10 年連續配息判斷
- ✅ JSON 格式輸出

#### **資料檔案**
- **程式**：`dividend_scraper.go`
- **輸出**：`stock_web/dividend_data.json`

#### **更新頻率**
- **自動**：每天 06:00（Linux Cron）
- **手動**：`bash update_dividend_data.sh`

#### **執行時間**
- **81 支股票**：約 10-15 秒

---

### **5. 網頁資料更新** 🌐

#### **功能**
- ✅ 更新 `portfolio.json` 即時股價
- ✅ 自動計算即時損益
- ✅ 供網頁介面顯示

#### **資料檔案**
- **程式**：`web_updater.go`
- **輸出**：`stock_web/portfolio.json`

#### **更新頻率**
- **自動**：週一至週五 15:30（Linux Cron）
- **手動**：`bash cron/web_updater.sh`

#### **執行時間**
- 約 5-10 秒

---

## 📅 **排程系統設計（雙軌制）**

### **🔧 為何使用 Linux Cron + OpenClaw Cron？**

| 任務類型 | 使用排程 | 原因 |
|---------|---------|------|
| **純資料更新** | Linux Cron | 不依賴 OpenClaw 服務，穩定性高 |
| **需要 Telegram 推播** | OpenClaw Cron | 需要 delivery.announce 功能 |

---

### **📊 Linux Cron 排程（獨立運行）**

```bash
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 每天 06:00 - 每日選股（81支股票池）
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
0 6 * * * /home/administrator/.openclaw/workspace/cron/daily_stock_picker.sh

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 每天 06:00 - 配息資料更新
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
0 6 * * * /home/administrator/.openclaw/workspace/cron/update_dividend_data.sh

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 週一至週五 15:00 - 股票快照記錄
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
0 15 * * 1-5 /home/administrator/.openclaw/workspace/cron/stock_snapshot.sh

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 週一至週五 15:30 - 網頁資料更新
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
30 15 * * 1-5 /home/administrator/.openclaw/workspace/cron/web_updater.sh
```

#### **安裝方式**
```bash
# 方式 1：使用安裝腳本（推薦）
bash /home/administrator/.openclaw/workspace/cron/install_cron.sh

# 方式 2：手動安裝
crontab /home/administrator/.openclaw/workspace/cron/crontab_stock_system.txt

# 查看目前設定
crontab -l

# 編輯設定
crontab -e
```

---

### **📱 OpenClaw Cron 排程（需 Telegram）**

以下任務**必須保留在 OpenClaw Cron**，因為需要推播到 Telegram：

| 任務名稱 | 時間 | 說明 |
|---------|------|------|
| **AI 台股綜合評分** | 15:00 (週一至週五) | 執行 `ai_stock_scorer_v3.go`，推播分析結果 |
| **台股收盤摘要** | 16:00 (週一至週五) | 大盤指數、三大法人、類股表現 |
| **FSI 學習提醒** | 19:00 (每天) | 英語/越南語學習提醒 |
| **優惠追蹤** | 08:00 (每天) | 商品價格追蹤 |
| **投資組合月報** | 4/13 09:00 | 30天績效報告 |

**保留原因**：這些任務都有 `delivery.announce`，需要 Telegram 推播功能。

---

## 🗂️ **檔案結構總覽**

```
/home/administrator/.openclaw/workspace/
├── 📊 選股系統（V3.0）
│   ├── daily_picker_integrated.go - 主程式（V3.0，81支，20-60元）
│   ├── stock_pool.json - 股票池設定（獨立管理）
│   ├── stock_pool_loader.go - 股票池載入工具
│   └── stock_web/daily_report.json - 選股結果
│
├── 💼 交易系統
│   ├── trade_manager.go - API 伺服器 (Port 8888)
│   ├── stock_trades.db - 交易資料庫（SQLite）
│   ├── portfolio.json - 持股快照（根目錄）
│   └── stock_web/
│       ├── index.html - 主介面（已支援 50-60元區間）
│       ├── simulation.js - 交易邏輯
│       ├── portfolio_batches.js - 分批次持股
│       └── portfolio.json - 持股快照（stock_web）
│
├── 💰 高股息追蹤
│   ├── dividend_scraper.go - 爬蟲程式
│   ├── update_dividend_data.sh - 更新腳本
│   └── stock_web/
│       ├── dividend_tracker.html - 追蹤介面
│       ├── dividend_strategy.html - 策略分析
│       └── dividend_data.json - 配息資料
│
├── 🌐 網頁更新
│   └── web_updater.go - 即時股價更新
│
├── 🕐 排程系統（Linux Cron）
│   └── cron/
│       ├── daily_stock_picker.sh - 每日選股腳本
│       ├── update_dividend_data.sh - 配息更新腳本
│       ├── web_updater.sh - 網頁更新腳本
│       ├── stock_snapshot.sh - 快照記錄腳本
│       ├── crontab_stock_system.txt - Crontab 設定檔
│       └── install_cron.sh - 安裝腳本
│
├── 📚 文件
│   └── stock_web/
│       ├── SYSTEM_OVERVIEW_V3.md - 本文件
│       ├── TRADING_GUIDE.md - 交易管理指南
│       ├── DIVIDEND_SCRAPER_GUIDE.md - 爬蟲使用指南
│       └── REALTIME_UPDATER_GUIDE.md - 即時更新指南
│
└── 📊 日誌
    └── logs/
        ├── daily_picker_YYYYMMDD.log - 選股日誌
        ├── dividend_update_YYYYMMDD.log - 配息更新日誌
        ├── web_update_YYYYMMDD.log - 網頁更新日誌
        └── snapshot_YYYYMMDD.log - 快照日誌
```

---

## 🚀 **快速開始指南**

### **1. 安裝 Linux Cron 排程**

```bash
# 安裝排程（互動式）
bash /home/administrator/.openclaw/workspace/cron/install_cron.sh

# 或手動安裝
crontab /home/administrator/.openclaw/workspace/cron/crontab_stock_system.txt
```

### **2. 測試執行**

```bash
# 測試選股系統
bash /home/administrator/.openclaw/workspace/cron/daily_stock_picker.sh

# 測試配息更新
bash /home/administrator/.openclaw/workspace/cron/update_dividend_data.sh

# 測試網頁更新
bash /home/administrator/.openclaw/workspace/cron/web_updater.sh
```

### **3. 查看日誌**

```bash
# 即時監控選股執行
tail -f ~/workspace/logs/daily_picker_$(date +%Y%m%d).log

# 查看配息更新日誌
tail -f ~/workspace/logs/dividend_update_$(date +%Y%m%d).log

# 查看所有日誌
ls -lh ~/workspace/logs/
```

### **4. 修改股票池**

```bash
# 編輯股票池（無需重新編譯程式）
nano /home/administrator/.openclaw/workspace/stock_pool.json

# 立即執行選股驗證
go run daily_picker_integrated.go
```

---

## 📊 **系統效能統計**

| 項目 | 時間 |
|------|------|
| **選股更新**（81支） | 8-10 分鐘 |
| **配息更新**（81支） | 10-15 秒 |
| **網頁更新** | 5-10 秒 |
| **快照記錄** | 5 秒 |

---

## 💡 **系統優化亮點（V3.0）**

### **1. 股票池獨立管理**
✅ 從程式碼分離到 `stock_pool.json`  
✅ 修改股票不需重新編譯  
✅ 支援分類管理（金融、電子、傳產等）

### **2. 價格區間擴充**
✅ 20-30元 → 30-40元 → 40-50元 → 50-60元  
✅ 涵蓋更廣的投資選擇

### **3. 雙軌排程系統**
✅ Linux Cron：資料更新（不依賴 OpenClaw）  
✅ OpenClaw Cron：Telegram 推播（保留通知功能）  
✅ 提升系統穩定性

### **4. 完整日誌系統**
✅ 所有任務都有獨立日誌  
✅ 自動清理 7 天前日誌  
✅ 集中管理於 `logs/` 目錄

---

## 🔐 **資料備份建議**

### **重要資料**
1. `stock_trades.db` - 交易記錄（最重要！）
2. `stock_pool.json` - 股票池設定
3. `dividend_data.json` - 配息資料
4. `daily_report.json` - 選股結果

### **備份方案**
```bash
# 建立備份目錄
mkdir -p ~/backups/stock_system

# 每週備份腳本
cd /home/administrator/.openclaw/workspace
cp stock_trades.db ~/backups/stock_system/stock_trades_$(date +%Y%m%d).db
cp stock_pool.json ~/backups/stock_system/stock_pool_$(date +%Y%m%d).json
cp stock_web/dividend_data.json ~/backups/stock_system/dividend_data_$(date +%Y%m%d).json

# 保留最近 30 天
find ~/backups/stock_system -name "*.db" -mtime +30 -delete
find ~/backups/stock_system -name "*.json" -mtime +30 -delete
```

---

## 🦈 **系統成熟度評估**

### **完整度**
```
功能完整度：★★★★★ (95%)
資料完整度：★★★★☆ (85%, 81支股票池)
自動化程度：★★★★★ (95%, Linux Cron + OpenClaw Cron)
使用者體驗：★★★★☆ (90%)
文件完整度：★★★★★ (100%)
穩定性：    ★★★★★ (95%, 雙軌排程)
```

---

## 📝 **版本記錄**

- **v1.0** (2026-03-25 12:00) - 初版系統
- **v2.0** (2026-03-25 16:00) - 加入高股息追蹤
- **v2.5** (2026-03-25 17:10) - 加入配息爬蟲、10年連續配息
- **v3.0** (2026-03-27 11:30) - 🎉 重大升級
  - 升級股票池：36支 → 81支（ETF 14 + 個股 67）
  - 擴充價格區間：20-50元（3個）→ 20-60元（4個）
  - 股票池獨立管理：`stock_pool.json`
  - 雙軌排程系統：Linux Cron + OpenClaw Cron
  - 完整日誌系統
  - 網頁支援 50-60元區間

---

## 🚀 **未來規劃**

### **近期（1-2 週）**
- [ ] 補齊剩餘配息資料（81支 → 100%覆蓋）
- [ ] 圖表視覺化（損益曲線、持股比例圓餅圖）
- [ ] 手機 PWA 版本

### **中期（1-2 個月）**
- [ ] 多策略組合比較（保守型 vs 積極型）
- [ ] 股息再投入計算器
- [ ] 股價預警通知

### **長期（3-6 個月）**
- [ ] 財報資料整合（EPS、本益比、ROE）
- [ ] AI 選股模型優化
- [ ] 回測系統

---

**記住**：系統是為了幫助你更好地投資，不是讓投資變複雜！保持簡單實用！🦈
