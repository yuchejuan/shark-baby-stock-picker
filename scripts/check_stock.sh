#!/bin/bash

# 🦈 股票代號檢查工具

if [ -z "$1" ]; then
    echo "❌ 請提供股票代號"
    echo "用法: ./check_stock.sh <股票代號>"
    echo "範例: ./check_stock.sh 3105"
    exit 1
fi

SYMBOL=$1

echo "🔍 檢查股票代號: $SYMBOL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. 查詢 TWSE API
echo "📡 Step 1: 查詢 TWSE 代號搜尋 API..."
RESULT=$(curl -s "https://www.twse.com.tw/zh/api/codeQuery?query=$SYMBOL")

echo "原始回應:"
echo "$RESULT" | python3 -m json.tool 2>/dev/null || echo "$RESULT"
echo ""

# 2. 解析建議
SUGGESTIONS=$(echo "$RESULT" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    suggestions = data.get('suggestions', [])
    if suggestions:
        for i, s in enumerate(suggestions):
            parts = s.split('\t')
            print(f'{i+1}. {s}')
            print(f'   代號: {parts[0] if len(parts) > 0 else \"?\"}')
            print(f'   名稱: {parts[1] if len(parts) > 1 else \"(無名稱)\"}')
    else:
        print('無建議結果')
except Exception as e:
    print(f'解析錯誤: {e}')
" 2>&1)

echo "解析結果:"
echo "$SUGGESTIONS"
echo ""

# 3. 檢查是否存在
if echo "$RESULT" | grep -q '"suggestions":\[\]' || echo "$SUGGESTIONS" | grep -q "無建議結果"; then
    echo "❌ 股票代號 $SYMBOL 不存在或已下市"
    echo ""
    echo "💡 可能原因："
    echo "   1. 股票代號輸入錯誤"
    echo "   2. 該股票已下市或下櫃"
    echo "   3. 興櫃股票（TWSE API 可能不支援）"
    echo ""
    echo "🔍 建議："
    echo "   - 確認代號是否正確（台股通常是 4 碼數字）"
    echo "   - 至 https://www.twse.com.tw 查詢"
    exit 1
fi

echo "✅ 股票代號存在！"
echo ""

# 4. 嘗試查詢完整資料
echo "📊 Step 2: 嘗試查詢股價資料..."
cd ~/.openclaw/workspace
go run stock_query_safe.go "$SYMBOL" 2>&1
