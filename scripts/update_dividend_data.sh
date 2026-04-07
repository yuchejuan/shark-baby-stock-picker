#!/bin/bash
# 🦈 高股息追蹤系統 - 自動更新腳本
# 每週日早上 06:00 自動執行

WORKSPACE="/home/administrator/.openclaw/workspace"
LOG_FILE="$WORKSPACE/logs/dividend_update.log"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')

# 建立 logs 目錄
mkdir -p "$WORKSPACE/logs"

echo "========================================" >> "$LOG_FILE"
echo "🦈 開始更新配息資料" >> "$LOG_FILE"
echo "時間：$TIMESTAMP" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

# 切換到工作目錄
cd "$WORKSPACE" || exit 1

# 執行爬蟲
echo "📊 執行 dividend_scraper.go..." >> "$LOG_FILE"
timeout 120 go run dividend_scraper.go >> "$LOG_FILE" 2>&1

if [ $? -eq 0 ]; then
    echo "✅ 配息資料更新成功" >> "$LOG_FILE"
    
    # 檢查 JSON 檔案是否存在
    if [ -f "stock_web/dividend_data.json" ]; then
        FILE_SIZE=$(du -h stock_web/dividend_data.json | cut -f1)
        STOCK_COUNT=$(jq 'length' stock_web/dividend_data.json 2>/dev/null || echo "N/A")
        echo "📁 檔案大小：$FILE_SIZE" >> "$LOG_FILE"
        echo "📊 股票數量：$STOCK_COUNT 支" >> "$LOG_FILE"
        
        # 備份舊資料
        BACKUP_DIR="$WORKSPACE/backups/dividend_data"
        mkdir -p "$BACKUP_DIR"
        cp stock_web/dividend_data.json "$BACKUP_DIR/dividend_data_$(date '+%Y%m%d_%H%M%S').json"
        echo "💾 已備份至：$BACKUP_DIR" >> "$LOG_FILE"
    else
        echo "⚠️ JSON 檔案不存在" >> "$LOG_FILE"
    fi
else
    echo "❌ 配息資料更新失敗（退出碼：$?）" >> "$LOG_FILE"
fi

echo "========================================" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

# 保留最近 30 天的日誌
find "$WORKSPACE/logs" -name "dividend_update.log" -mtime +30 -delete 2>/dev/null

# 保留最近 90 天的備份
find "$WORKSPACE/backups/dividend_data" -name "dividend_data_*.json" -mtime +90 -delete 2>/dev/null

exit 0
