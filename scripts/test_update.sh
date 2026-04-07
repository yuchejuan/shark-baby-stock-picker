#!/bin/bash

echo "🧪 測試手動更新股價功能"
echo ""
echo "1️⃣ 檢查 API 伺服器狀態..."
if curl -s http://localhost:8888/api > /dev/null 2>&1; then
    echo "✅ API 伺服器運行中 (Port 8888)"
else
    echo "❌ API 伺服器未運行！"
    echo "   請執行: cd ~/.openclaw/workspace && ./trade_manager &"
    exit 1
fi

echo ""
echo "2️⃣ 測試更新股價端點..."
echo ""
curl -X POST http://localhost:8888/api/holdings/update-prices 2>&1 | jq '.'

echo ""
echo "3️⃣ 檢查 portfolio.json 更新時間..."
stat -c "最後修改: %y" ~/.openclaw/workspace/stock_web/portfolio.json

echo ""
echo "✅ 測試完成！"
echo ""
echo "📱 現在可以開啟瀏覽器："
echo "   http://localhost:8080"
echo ""
echo "   點擊「💼 持倉明細」→「🔄 手動更新股價」按鈕"
