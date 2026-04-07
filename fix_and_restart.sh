#!/bin/bash

echo "🦈 鯊魚寶寶系統修復與重啟..."
echo ""

# 1. 停止所有服務
echo "🛑 停止現有服務..."
lsof -ti:8080 | xargs kill -9 2>/dev/null
lsof -ti:8888 | xargs kill -9 2>/dev/null
sleep 2

# 2. 重新編譯 API
echo "📦 重新編譯 API..."
cd /home/administrator/.openclaw/workspace
go build -o trade_manager trade_manager.go

if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗！"
    exit 1
fi

# 3. 啟動 API
echo "🚀 啟動 API (Port 8888)..."
./trade_manager > trade_manager.log 2>&1 &
API_PID=$!
sleep 2

# 檢查 API
if curl -s http://localhost:8888/api > /dev/null 2>&1; then
    echo "✅ API 啟動成功 (PID: $API_PID)"
else
    echo "❌ API 啟動失敗"
    exit 1
fi

# 4. 更新資料
echo "📊 更新投資組合資料..."
go run web_updater.go

# 5. 啟動網頁
echo "🌐 啟動網頁伺服器 (Port 8080)..."
cd stock_web
python3 -m http.server 8080 > /dev/null 2>&1 &
WEB_PID=$!
sleep 1

# 檢查網頁
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo "✅ 網頁啟動成功 (PID: $WEB_PID)"
else
    echo "❌ 網頁啟動失敗"
    kill $API_PID 2>/dev/null
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🦈 系統修復完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  📱 主頁面: http://localhost:8080"
echo "  🧪 測試頁面: http://localhost:8080/test.html"
echo "  🔌 API: http://localhost:8888/api"
echo ""
echo "  API PID: $API_PID"
echo "  Web PID: $WEB_PID"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "🔧 如果網頁還是顯示錯誤，請："
echo "   1. 清除瀏覽器快取 (Ctrl+Shift+R 或 Cmd+Shift+R)"
echo "   2. 開啟測試頁面檢查資料: http://localhost:8080/test.html"
echo ""
