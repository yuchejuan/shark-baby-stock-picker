#!/bin/bash

# 🦈 模擬買入功能自動整合腳本

echo "🦈 鯊魚寶寶 - 模擬買入功能整合"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd ~/.openclaw/workspace

# 檢查檔案是否存在
if [ ! -f "add_buy_modal.html" ]; then
    echo "❌ 找不到 add_buy_modal.html"
    exit 1
fi

if [ ! -f "stock_web/index.html" ]; then
    echo "❌ 找不到 stock_web/index.html"
    exit 1
fi

# 備份原檔案
echo "📦 備份原始檔案..."
cp stock_web/index.html stock_web/index.html.backup.$(date +%Y%m%d_%H%M%S)
echo "✅ 備份完成"
echo ""

# 檢查是否已整合
if grep -q "id=\"buyModal\"" stock_web/index.html; then
    echo "⚠️  買入對話框已存在，跳過整合"
else
    echo "🔧 整合買入對話框到網頁..."
    
    # 在 </body> 之前插入
    sed -i '/<\/body>/i\<!-- 模擬買入功能 -->' stock_web/index.html
    sed -i '/<\/body>/r add_buy_modal.html' stock_web/index.html
    
    echo "✅ 整合完成"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 整合完成！"
echo ""
echo "📋 下一步："
echo ""
echo "1. 啟動投資組合 API："
echo "   go run portfolio_manager.go &"
echo ""
echo "2. 啟動網頁服務："
echo "   cd stock_web && python3 -m http.server 8080"
echo ""
echo "3. 瀏覽器開啟："
echo "   http://localhost:8080"
echo ""
echo "4. 查詢股票 → 點擊「📈 模擬買入」"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
