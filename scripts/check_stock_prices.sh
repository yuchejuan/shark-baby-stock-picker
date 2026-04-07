#!/bin/bash

echo "🔍 開始檢查所有股票資料..."
echo "================================"

# 所有股票代號
stocks=(
  "0050" "006208" "00631L" "00632R" "00692" "00701" "00881" "00891" "00895" "00896"
  "00919" "00929" "00918" "00878" "0056"
  "2330" "2317" "2454" "2412" "2882" "2891" "2886" "2881" "2892" "2884"
  "2303" "2308" "2382" "2357" "3711" "2327" "2379"
  "2002" "1301" "1303" "1326" "2105"
  "2353" "2324" "2618" "2838" "2812" "2887" "2851" "2890" "1102" "5876" "2816"
  "3443" "6510" "2395" "2356" "6669"
  "1101" "6506" "6411"
  "3045" "4904" "2049" "3008"
  "6505" "2207" "2880"
  "2409" "3034" "2301" "2408" "2344" "3481" "6176" "2371" "6414" "3661"
)

error_count=0
success_count=0

for symbol in "${stocks[@]}"; do
  # 嘗試上市
  result=$(curl -s "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_${symbol}.tw&json=1&delay=0")
  price=$(echo "$result" | jq -r '.msgArray[0].z // .msgArray[0].y // "-"')
  
  if [ "$price" = "-" ] || [ "$price" = "null" ]; then
    # 嘗試上櫃
    result=$(curl -s "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=otc_${symbol}.tw&json=1&delay=0")
    price=$(echo "$result" | jq -r '.msgArray[0].z // .msgArray[0].y // "-"')
  fi
  
  if [ "$price" = "-" ] || [ "$price" = "null" ] || [ -z "$price" ]; then
    echo "❌ $symbol - 無法取得股價"
    ((error_count++))
  else
    echo "✅ $symbol - \$$price"
    ((success_count++))
  fi
  
  sleep 0.3
done

echo "================================"
echo "✅ 成功: $success_count 支"
echo "❌ 失敗: $error_count 支"
echo "📊 總計: ${#stocks[@]} 支"
