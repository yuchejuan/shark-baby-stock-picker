# 🦈 鯊魚寶寶台股系統 - 快速上手指南

**版本**: V3.1 (優化版)  
**更新日期**: 2026-03-27

---

## 🚀 **3 步驟快速啟動**

### **步驟 1：遷移到優化版**
```bash
bash /home/administrator/.openclaw/workspace/cron/migrate_to_optimized.sh
```

### **步驟 2：測試快取系統**
```bash
bash /home/administrator/.openclaw/workspace/test_cache_system.sh
```

### **步驟 3：查看系統狀態**
```bash
# 查看排程
crontab -l

# 查看快取
go run stock_data_cache.go stats

# 查看日誌
ls -lh logs/
```

---

## 📊 **每日執行流程**

```
06:00 🌅 早晨資料同步
      ├─ 清理過期快取
      ├─ 配息資料更新
      ├─ 每日選股（建立 81 個快取檔案）
      └─ 快取統計
        ↓
15:00 📸 股票快照記錄
        ↓
15:30 🌐 網頁更新（使用快取，5秒完成）
        ↓
16:00 📰 收盤摘要（使用 daily_report.json）
```

---

## 🔧 **常用指令**

### **快取管理**
```bash
# 查看快取統計
go run stock_data_cache.go stats

# 取得單支股票（優先使用快取）
go run stock_data_cache.go get 2330

# 強制刷新
go run stock_data_cache.go refresh 2330

# 清理過期快取
go run stock_data_cache.go clear

# 查看快取檔案
ls -lh .cache/stock_data/
```

### **手動執行任務**
```bash
# 早晨資料同步（完整流程）
bash cron/morning_data_sync.sh

# 網頁更新
bash cron/web_updater_optimized.sh

# 股票快照
bash cron/stock_snapshot.sh
```

### **查看日誌**
```bash
# 即時監控早晨同步
tail -f logs/morning_sync_$(date +%Y%m%d).log

# 查看網頁更新日誌
tail -f logs/web_update_$(date +%Y%m%d).log

# 查看所有日誌
ls -lh logs/
```

---

## 📁 **重要檔案位置**

### **資料檔案**
```
stock_web/daily_report.json      - 每日選股結果
stock_web/portfolio.json         - 投資組合（即時股價）
stock_web/dividend_data.json     - 配息資料
.cache/stock_data/*.json         - TWSE 快取（81個檔案）
```

### **設定檔案**
```
stock_pool.json                  - 股票池設定（81支）
cron/crontab_optimized.txt       - 優化版排程設定
```

### **日誌檔案**
```
logs/morning_sync_YYYYMMDD.log   - 早晨同步日誌
logs/web_update_YYYYMMDD.log     - 網頁更新日誌
logs/snapshot_YYYYMMDD.log       - 快照記錄日誌
```

---

## 🎯 **核心優化成果**

| 項目 | 優化前 | 優化後 | 提升 |
|------|--------|--------|------|
| **API 呼叫次數** | 3-4 次/天 | 1 次/天 | ⬇️ 75% |
| **網頁更新速度** | 30 秒 | 5 秒 | ⬆️ 6 倍 |
| **資料一致性** | ⚠️ 不保證 | ✅ 保證 | ⬆️ 100% |

---

## 📖 **詳細文件**

- **優化說明**：`cat DATA_FLOW_OPTIMIZATION.md`
- **系統總覽**：`cat stock_web/SYSTEM_OVERVIEW_V3.md`
- **排程設定**：`cat cron/crontab_optimized.txt`

---

## 🆘 **常見問題**

### **Q1：快取失效怎麼辦？**
```bash
# 強制重新取得所有資料
bash cron/morning_data_sync.sh
```

### **Q2：如何查看快取內容？**
```bash
# 查看台積電快取
cat .cache/stock_data/2330.json | jq .
```

### **Q3：如何恢復舊版排程？**
```bash
# 查看備份
ls -lh cron/crontab_backup_*

# 恢復備份
crontab cron/crontab_backup_YYYYMMDD_HHMMSS.txt
```

### **Q4：網頁更新慢怎麼辦？**
```bash
# 檢查快取狀態
go run stock_data_cache.go stats

# 如果快取過期，會自動重新取得（需 30 秒）
# 建議：確保早晨同步正常執行（06:00）
```

---

## 🦈 **技術支援**

- **文件位置**：`/home/administrator/.openclaw/workspace/`
- **日誌位置**：`logs/`
- **快取位置**：`.cache/stock_data/`

**記住**：系統是為了讓投資更簡單，不是更複雜！🦈
