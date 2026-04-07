#!/bin/bash

echo "🦈 鯊魚寶寶全市場選股系統"
echo "📊 無價格限制版本"
echo ""

cd /home/administrator/.openclaw/workspace

# 編譯
echo "📦 編譯選股程式..."
go build -o daily_stock_picker_all daily_stock_picker_all.go

if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗！"
    exit 1
fi

echo "✅ 編譯成功"
echo ""

# 執行選股
echo "🔍 開始選股分析..."
echo ""
./daily_stock_picker_all

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ 更新完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  📂 報告位置: stock_web/daily_report.json"
echo "  🌐 查看網頁: http://localhost:8080"
echo ""
