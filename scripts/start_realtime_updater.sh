#!/bin/bash

echo "🦈 鯊魚寶寶即時股價更新系統"
echo ""

# 檢查是否已在運行
if pgrep -f "realtime_price_updater" > /dev/null; then
    echo "⚠️  系統已在運行中"
    echo ""
    echo "如要重啟，請先執行："
    echo "  pkill -f realtime_price_updater"
    echo "  然後再執行此腳本"
    exit 1
fi

# 編譯
echo "📦 編譯程式..."
cd /home/administrator/.openclaw/workspace
go build -o realtime_price_updater realtime_price_updater.go

if [ $? -ne 0 ]; then
    echo "❌ 編譯失敗！"
    exit 1
fi

echo "✅ 編譯成功"
echo ""

# 啟動（背景執行）
echo "🚀 啟動即時更新服務..."
nohup ./realtime_price_updater > realtime_updater.log 2>&1 &
PID=$!

echo "✅ 服務已啟動 (PID: $PID)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  📊 即時股價更新系統"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  ⏰ 更新間隔: 20 分鐘"
echo "  📅 運作時間: 週一至週五 09:00-14:00"
echo "  📂 日誌檔案: realtime_updater.log"
echo ""
echo "  🛑 停止服務:"
echo "     pkill -f realtime_price_updater"
echo ""
echo "  📖 查看日誌:"
echo "     tail -f realtime_updater.log"
echo ""
echo "  🔍 檢查狀態:"
echo "     ps aux | grep realtime_price_updater"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 顯示初始日誌
sleep 2
echo "📋 初始日誌："
tail -20 realtime_updater.log
