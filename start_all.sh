#!/bin/bash

cd "$(dirname "$0")"
PROJ_DIR="$(pwd)"

echo "🦈 鯊魚寶寶完整系統啟動中..."
echo ""

# ── 1. 編譯交易管理 API ──────────────────────────
echo "📦 編譯交易管理 API..."
go build -o trade_manager trade_manager.go
if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗！請確認已安裝 Go 1.23+ 與 gcc"
    exit 1
fi

# ── 2. 啟動交易管理 API (Port 8888) ──────────────
echo "🚀 啟動交易管理 API (Port 8888)..."
./trade_manager > trade_manager.log 2>&1 &
API_PID=$!
sleep 2

if ! curl -s http://localhost:8888/api/trades > /dev/null 2>&1; then
    echo "❌ API 啟動失敗，請查看 trade_manager.log"
    kill $API_PID 2>/dev/null
    exit 1
fi
echo "✅ 交易 API 已啟動"
echo ""

# ── 3. 編譯並啟動股票查詢 API (Port 8765) ────────
echo "📦 編譯股票查詢服務..."
go build -o stock_query_cli stock_query_service.go && \
go build -o stock_query_server stock_query_api.go
if [ $? -eq 0 ]; then
    "$PROJ_DIR/stock_query_server" > stock_query.log 2>&1 &
    QUERY_PID=$!
    sleep 1
    echo "✅ 查詢 API 已啟動 (Port 8765)"
else
    echo "⚠️  查詢 API 編譯失敗（不影響其他功能）"
fi
echo ""

# ── 4. 更新持倉股價 ───────────────────────────────
echo "📊 更新持倉股價..."
go run web_updater.go
echo ""

# ── 5. 確認 html/ 資料（若不存在就從根目錄複製備份）──
echo "📂 確認 html/ 資料狀態..."
for f in daily_report.json sector_heatmap.json portfolio.json dividend_data.json; do
    if [ ! -f "html/$f" ] && [ -f "$f" ]; then
        cp "$f" "html/$f"
        echo "  📋 補充初始資料：html/$f"
    elif [ -f "html/$f" ]; then
        echo "  ✅ html/$f"
    else
        echo "  ⚠️  html/$f 不存在（請執行 bash update_data.sh）"
    fi
done
echo ""

# ── 6. 啟動網頁伺服器 ────────────────────────────
PORT=8080
while lsof -i :$PORT > /dev/null 2>&1; do
    echo "⚠️  Port $PORT 已被佔用，嘗試 $((PORT+1))..."
    PORT=$((PORT+1))
done

echo "🌐 啟動網頁伺服器 (Port $PORT)..."
cd html

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🦈 鯊魚寶寶系統已啟動！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📱 主頁面：  http://localhost:$PORT"
echo "  💰 交易管理：http://localhost:$PORT/trade.html"
echo "  🔌 交易 API：http://localhost:8888/api"
echo "  🔍 查詢 API：http://localhost:8765/api/query?symbol=2330"
echo ""
echo "  📊 更新資料：另開終端執行 bash update_data.sh"
echo "  🛑 停止：    Ctrl+C"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cleanup() {
    echo ""
    echo "🛑 停止服務..."
    kill $API_PID 2>/dev/null
    kill $QUERY_PID 2>/dev/null
    echo "✅ 已清理"
    exit 0
}
trap cleanup INT TERM

python3 -m http.server $PORT
