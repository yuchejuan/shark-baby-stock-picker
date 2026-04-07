#!/bin/bash

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🦈 驗證頁面修復"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. 檢查頁面結構
echo "1️⃣ 檢查頁面結構..."
echo ""

PICKS=$(curl -s http://localhost:8080 | grep -c 'id="picks"')
SECTORS=$(curl -s http://localhost:8080 | grep -c 'id="sectors"')
PORTFOLIO=$(curl -s http://localhost:8080 | grep -c 'id="portfolio"')
HISTORY=$(curl -s http://localhost:8080 | grep -c 'id="history"')

echo "   📊 TOP 選股 (picks): $PICKS $([ $PICKS -eq 1 ] && echo '✅' || echo '❌')"
echo "   🔥 產業熱度 (sectors): $SECTORS $([ $SECTORS -eq 1 ] && echo '✅' || echo '❌')"
echo "   💼 持倉明細 (portfolio): $PORTFOLIO $([ $PORTFOLIO -eq 1 ] && echo '✅' || echo '❌')"
echo "   📜 歷次買賣 (history): $HISTORY $([ $HISTORY -eq 1 ] && echo '✅' || echo '❌')"

echo ""

# 2. 檢查更新按鈕
echo "2️⃣ 檢查更新股價按鈕..."
UPDATE_BTN=$(curl -s http://localhost:8080 | grep -c 'update-price-btn')
echo "   🔄 更新按鈕出現次數: $UPDATE_BTN $([ $UPDATE_BTN -eq 2 ] && echo '✅ (HTML + JS)' || echo '❌')"

echo ""

# 3. 檢查 JS 模組載入
echo "3️⃣ 檢查 JavaScript 模組..."
BATCHES_JS=$(curl -s http://localhost:8080 | grep -c 'portfolio_batches.js')
echo "   📦 portfolio_batches.js: $BATCHES_JS $([ $BATCHES_JS -eq 1 ] && echo '✅' || echo '❌')"

echo ""

# 4. 測試 API
echo "4️⃣ 測試 API 連線..."
if curl -s http://localhost:8888/api > /dev/null 2>&1; then
    echo "   ✅ API 伺服器運行正常 (Port 8888)"
else
    echo "   ❌ API 伺服器未運行"
fi

echo ""

# 5. 檢查持倉資料
echo "5️⃣ 檢查持倉資料..."
BATCHES_RESULT=$(curl -s http://localhost:8888/api/holdings/batches 2>&1)
if echo "$BATCHES_RESULT" | grep -q "symbol"; then
    BATCH_COUNT=$(echo "$BATCHES_RESULT" | grep -o '"symbol"' | wc -l)
    echo "   ✅ 持倉批次資料載入成功 ($BATCH_COUNT 筆)"
else
    echo "   ❌ 持倉資料載入失敗"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🎉 驗證完成！"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📱 開啟瀏覽器測試："
echo "   http://localhost:8080"
echo ""
echo "💡 測試步驟："
echo "   1. 點擊「💼 持倉明細」"
echo "   2. 確認資料顯示正常"
echo "   3. 點擊「🔄 手動更新股價」"
echo "   4. 等待更新完成"
echo ""
