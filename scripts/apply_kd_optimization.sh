#!/bin/bash
# KD 評分優化腳本

FILE="/home/administrator/.openclaw/workspace/stock_analyzer.go"

# 備份原始檔案
cp "$FILE" "${FILE}.backup"

# 使用 sed 進行替換
sed -i '429,437s/.*/\
\t\/\/ KD 評分 (15分) - 優化版\
\tif stock.KD == "超賣" {\
\t\tscore += 15\
\t\tadvantages = append(advantages, "KD超賣")\
\t} else if stock.KD == "偏低" {\
\t\tscore += 12\
\t\tadvantages = append(advantages, "KD偏低")\
\t} else if stock.KD == "偏多" {\
\t\tscore += 10\
\t} else if stock.KD == "中性" {\
\t\tscore += 8\
\t} else if stock.KD == "偏空" {\
\t\tscore += 6\
\t} else if stock.KD == "偏高" {\
\t\tscore += 4\
\t} else if stock.KD == "超買" {\
\t\tscore += 2\
\t} else {\
\t\tscore += 5\
\t}/' "$FILE"

echo "✅ KD 評分優化完成！"
echo "📦 備份檔案：${FILE}.backup"
