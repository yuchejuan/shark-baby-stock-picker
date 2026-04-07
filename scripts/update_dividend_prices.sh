#!/bin/bash

echo "🔄 更新 dividend_tracker.html 的所有參考價為真實股價..."

# 備份
cp stock_web/dividend_tracker.html stock_web/dividend_tracker.html.backup

# 使用 sed 批次更新股價（根據之前檢查的真實股價）
sed -i "s/price: 198,/price: 76.20,/g" stock_web/dividend_tracker.html  # 0050
sed -i "s/price: 129,/price: 176.55,/g" stock_web/dividend_tracker.html # 006208
sed -i "s/price: 8,/price: 14.10,/g" stock_web/dividend_tracker.html    # 00632R
sed -i "s/price: 45,/price: 67.05,/g" stock_web/dividend_tracker.html   # 00692 (第一個45)
sed -i "s/price: 35,/price: 28.92,/g" stock_web/dividend_tracker.html   # 00701
sed -i "s/price: 28,/price: 37.38,/g" stock_web/dividend_tracker.html   # 00881 (第一個28)
sed -i "s/price: 45,/price: 24.50,/g" stock_web/dividend_tracker.html   # 00891 (第二個45)
sed -i "s/price: 32,/price: 39.95,/g" stock_web/dividend_tracker.html   # 00895 (第一個32)
sed -i "s/price: 28,/price: 21.19,/g" stock_web/dividend_tracker.html   # 00896 (第二個28)

echo "✅ 更新完成！"
echo "📄 備份檔案：stock_web/dividend_tracker.html.backup"
