#!/bin/bash

cd "$(dirname "$0")"

PORT=8080
while lsof -i :$PORT > /dev/null 2>&1; do
    echo "⚠️  Port $PORT 已被佔用，嘗試 $((PORT+1))..."
    PORT=$((PORT+1))
done

echo "🌐 啟動網頁伺服器 (Port $PORT)..."
echo "  開啟瀏覽器訪問：http://localhost:$PORT"
echo "  停止：Ctrl+C"
echo ""
cd html
python3 -m http.server $PORT
