#!/bin/bash
echo "=========================================="
echo "🔍 資料流分析報告"
echo "=========================================="
echo ""

echo "📊 程式列表與資料來源："
echo "=========================================="

for file in daily_picker_integrated.go web_updater.go realtime_price_updater.go portfolio_simple.go; do
    if [ -f "$file" ]; then
        echo ""
        echo "📁 $file"
        echo "   資料來源："
        grep -o "twse.com.tw/[^\"]*" "$file" | head -3 | sed 's/^/     - /'
        echo "   輸出檔案："
        grep -o '"[^"]*\.json"' "$file" | head -3 | sed 's/^/     - /'
    fi
done

echo ""
echo "=========================================="
echo "📝 重複 API 呼叫分析"
echo "=========================================="

echo ""
echo "TWSE API 使用統計："
grep -r "twse.com.tw" *.go 2>/dev/null | wc -l | xargs -I {} echo "  - 總共 {} 處呼叫"

echo ""
echo "=========================================="
