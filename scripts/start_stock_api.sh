#!/bin/bash

# 🦈 鯊魚寶寶股票查詢 API 啟動腳本

cd ~/.openclaw/workspace

echo "🦈 鯊魚寶寶股票查詢 API"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 檢查是否已經在運行
if lsof -i :8765 >/dev/null 2>&1; then
    echo "⚠️  API 服務已在運行中 (端口 8765)"
    echo ""
    echo "如需重啟，請先執行："
    echo "  pkill -f stock_query_api"
    echo ""
    exit 1
fi

echo "🚀 正在啟動 API 服務..."
echo ""

# 啟動服務（背景執行）
nohup go run stock_query_api.go > stock_api.log 2>&1 &
API_PID=$!

sleep 2

# 檢查是否成功啟動
if lsof -i :8765 >/dev/null 2>&1; then
    echo "✅ API 服務啟動成功！"
    echo ""
    echo "📡 服務資訊："
    echo "  - URL: http://localhost:8765"
    echo "  - PID: $API_PID"
    echo "  - 日誌: ~/.openclaw/workspace/stock_api.log"
    echo ""
    echo "📝 測試查詢："
    echo "  curl \"http://localhost:8765/api/query?symbol=2330\""
    echo ""
    echo "🛑 停止服務："
    echo "  pkill -f stock_query_api"
    echo ""
else
    echo "❌ API 服務啟動失敗"
    echo ""
    echo "請檢查日誌："
    echo "  tail -f ~/.openclaw/workspace/stock_api.log"
    echo ""
    exit 1
fi
