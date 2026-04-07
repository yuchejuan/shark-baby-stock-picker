#!/bin/bash

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🦈 鯊魚寶寶選股系統 - 快速啟動"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 切換到工作目錄
cd ~/.openclaw/workspace

# 1. 檢查並啟動 API 伺服器
echo "🔍 檢查 API 伺服器..."
if curl -s http://localhost:8888/api > /dev/null 2>&1; then
    echo "✅ API 已運行"
else
    echo "🚀 啟動 API 伺服器..."
    
    # 編譯（如果需要）
    if [ ! -f "./trade_manager" ] || [ "trade_manager.go" -nt "./trade_manager" ]; then
        echo "📦 編譯 API..."
        go build -o trade_manager trade_manager.go
        if [ $? -ne 0 ]; then
            echo "❌ 編譯失敗！"
            exit 1
        fi
    fi
    
    # 啟動 API
    ./trade_manager > trade_manager.log 2>&1 &
    sleep 2
    
    # 確認啟動成功
    if curl -s http://localhost:8888/api > /dev/null 2>&1; then
        echo "✅ API 啟動成功"
    else
        echo "❌ API 啟動失敗！請檢查 trade_manager.log"
        exit 1
    fi
fi

echo ""

# 2. 更新股價資料
echo "📊 更新股價資料..."
go run web_updater.go
echo ""

# 3. 啟動網頁伺服器
cd stock_web

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  ✅ 系統啟動完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  📱 主頁面:"
echo "     http://localhost:8080"
echo ""
echo "  💼 持倉明細 → 點擊「🔄 手動更新股價」即時刷新"
echo ""
echo "  🔌 API 端點:"
echo "     http://localhost:8888/api"
echo ""
echo "  🛑 停止服務: 按 Ctrl+C"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 啟動網頁伺服器
python3 -m http.server 8080
