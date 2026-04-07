#!/bin/bash
# 股票報告更新腳本（證交所API版本）

echo "🦈 開始更新股票報告..."
echo "時間: $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

cd /home/administrator/.openclaw/workspace

# 1. 更新投資組合資料
echo "📊 步驟1: 更新投資組合資料..."
go run web_updater.go
if [ $? -eq 0 ]; then
    echo "✅ 投資組合更新完成"
else
    echo "❌ 投資組合更新失敗"
    exit 1
fi

echo ""

# 2. 產生每日選股報告
echo "📊 步驟2: 產生每日選股報告..."
go run stock_analyzer.go daily_stock_picker.go
if [ $? -eq 0 ]; then
    echo "✅ 選股報告產生完成"
else
    echo "❌ 選股報告產生失敗"
    exit 1
fi

echo ""
echo "🎯 所有報告更新完成！"
echo "📁 檔案位置:"
echo "   - 投資組合: stock_web/portfolio.json"
echo "   - 選股報告: stock_web/daily_report.json"
echo "🌐 網頁查看: http://localhost:8081"
echo ""
echo "🦈 完成！"
