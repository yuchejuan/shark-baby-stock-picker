#!/bin/bash

echo "🦈 股票查詢功能測試"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

cd ~/.openclaw/workspace

# 測試 1: 不存在的股票
echo "📋 測試 1: 查詢不存在的股票 (3105)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
go run stock_query_service.go 3105 2>&1
echo ""
echo ""

# 測試 2: 正常的股票
echo "📋 測試 2: 查詢存在的股票 (2330 台積電)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
timeout 30 go run stock_query_service.go 2330 2>&1 | head -15
echo ""
echo ""

# 測試 3: 另一個正常股票
echo "📋 測試 3: 查詢存在的股票 (2812 台中銀)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
timeout 30 go run stock_query_service.go 2812 2>&1 | grep -E "symbol|name|price|score|signal"
echo ""

echo "✅ 測試完成"
