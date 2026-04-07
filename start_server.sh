#!/bin/bash

echo "🦈 鯊魚寶寶選股系統 - 啟動中..."
echo ""

# 1. 更新投資組合資料
echo "📊 正在更新投資組合資料..."
cd /home/administrator/.openclaw/workspace
go run web_updater.go

echo ""
echo "✅ 資料更新完成！"
echo ""

# 2. 啟動網頁伺服器
echo "🌐 啟動網頁伺服器..."
cd html

PORT=8080
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🦈 鯊魚寶寶選股系統已啟動！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  📱 開啟瀏覽器訪問:"
echo "     http://localhost:$PORT"
echo ""
echo "  🛑 停止伺服器: 按 Ctrl+C"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

python3 -m http.server $PORT
