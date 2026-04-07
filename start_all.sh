#!/bin/bash

echo "🦈 鯊魚寶寶完整系統啟動中..."
echo ""

# 1. 編譯 API 伺服器
echo "📦 編譯交易管理 API..."
cd /home/administrator/.openclaw/workspace
go build -o trade_manager trade_manager.go
if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗！"
    exit 1
fi

# 2. 啟動 API 伺服器（背景執行）
echo "🚀 啟動交易管理 API (Port 8888)..."
./trade_manager > trade_manager.log 2>&1 &
API_PID=$!
echo "   API PID: $API_PID"

# 等待 API 啟動
sleep 2

# 檢查 API 是否成功啟動
if ! curl -s http://localhost:8888/api/trades > /dev/null 2>&1; then
    echo "❌ API 啟動失敗！"
    kill $API_PID 2>/dev/null
    exit 1
fi

echo "✅ API 已成功啟動"
echo ""

# 3. 更新投資組合資料
echo "📊 正在更新投資組合資料..."
go run web_updater.go

echo ""
echo "✅ 資料更新完成！"
echo ""

# 4. 啟動網頁伺服器
echo "🌐 啟動網頁伺服器 (Port 8080)..."
cd html

PORT=8080
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🦈 鯊魚寶寶完整系統已啟動！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  📱 主頁面:"
echo "     http://localhost:$PORT"
echo ""
echo "  💰 交易管理:"
echo "     http://localhost:$PORT/trade.html"
echo ""
echo "  🔌 API 端點:"
echo "     http://localhost:8888/api"
echo ""
echo "  🛑 停止伺服器: 按 Ctrl+C"
echo "     （會自動清理所有服務）"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 清理函數
cleanup() {
    echo ""
    echo "🛑 正在停止服務..."
    kill $API_PID 2>/dev/null
    echo "✅ 已清理所有背景服務"
    exit 0
}

# 捕捉 Ctrl+C
trap cleanup INT TERM

# 啟動網頁伺服器
python3 -m http.server $PORT
