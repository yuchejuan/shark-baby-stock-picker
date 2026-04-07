#!/bin/bash
# 快取系統測試腳本

echo "🦈 測試中央快取系統"
echo "=========================================="

WORKSPACE="/home/administrator/.openclaw/workspace"
cd "$WORKSPACE" || exit 1

# 測試 1：取得單支股票資料（建立快取）
echo ""
echo "📊 測試 1：取得 2330 台積電資料（建立快取）"
echo "----------------------------------------"
time go run stock_data_cache.go refresh 2330
echo ""

# 測試 2：再次取得相同股票（應使用快取）
echo "📊 測試 2：再次取得 2330 台積電資料（應使用快取）"
echo "----------------------------------------"
time go run stock_data_cache.go get 2330
echo ""

# 測試 3：查看快取統計
echo "📊 測試 3：查看快取統計"
echo "----------------------------------------"
go run stock_data_cache.go stats
echo ""

# 測試 4：查看快取檔案內容
echo "📊 測試 4：查看快取檔案內容"
echo "----------------------------------------"
if [ -f ".cache/stock_data/2330.json" ]; then
    echo "✅ 快取檔案存在"
    echo "檔案大小："
    ls -lh .cache/stock_data/2330.json
    echo ""
    echo "檔案內容（前 20 行）："
    head -20 .cache/stock_data/2330.json
else
    echo "❌ 快取檔案不存在"
fi

echo ""
echo "=========================================="
echo "✅ 測試完成"
echo ""
echo "📝 觀察重點："
echo "  1. 測試1（建立快取）應該需要 2-3 秒"
echo "  2. 測試2（使用快取）應該 < 0.1 秒"
echo "  3. 快取檔案應該包含完整的歷史 K 線資料"
echo "=========================================="
