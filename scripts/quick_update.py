#!/usr/bin/env python3
"""
快速更新工具 - 針對 dividend_tracker.html 的 82 支股票
"""

import json
import time
import requests
from bs4 import BeautifulSoup
from datetime import datetime

# 從 dividend_tracker.html 中定義的 82 支股票
STOCK_LIST = [
    # 🏆 市值型 ETF TOP 9
    ('0050', '元大台灣50'), ('006208', '富邦台50'), ('00632R', '元大台灣50反1'),
    ('00692', '富邦公司治理'), ('00701', '國泰股利精選30'), ('00881', '國泰台灣5G+'),
    ('00891', '中信關鍵半導體'), ('00895', '富邦未來車'), ('00896', '中信綠能及電動車'),
    
    # 💰 高股息 ETF (5 檔)
    ('00919', '群益台灣精選高息'), ('00929', '復華台灣科技優息'), ('00918', '大華優利高填息30'),
    ('00878', '國泰永續高股息'), ('0056', '元大高股息'),
    
    # 權值股 (10 檔)
    ('2330', '台積電'), ('2317', '鴻海'), ('2454', '聯發科'), ('2412', '中華電'),
    ('2882', '國泰金'), ('2891', '中信金'), ('2886', '兆豐金'), ('2881', '富邦金'),
    ('2892', '第一金'), ('2884', '玉山金'),
    
    # 電子股 (7 檔)
    ('2303', '聯電'), ('2308', '台達電'), ('2382', '廣達'), ('2357', '華碩'),
    ('3711', '日月光投控'), ('2327', '國巨'), ('2379', '瑞昱'),
    
    # 傳產股 (5 檔)
    ('2002', '中鋼'), ('1301', '台塑'), ('1303', '南亞'), ('1326', '台化'), ('2105', '正新'),
    
    # 中小型股 (11 檔)
    ('2353', '宏碁'), ('2324', '仁寶'), ('2618', '長榮航'), ('2838', '聯邦銀'),
    ('2812', '台中銀'), ('2887', '台新金'), ('2851', '中再保'), ('2890', '永豐金'),
    ('1102', '亞泥'), ('5876', '上海商銀'), ('2816', '旺旺保'),
    
    # 🤖 AI 相關 (5 檔)
    ('3443', '創意'), ('6510', '精測'), ('2395', '研華'), ('2356', '英業達'), ('6669', '緯穎'),
    
    # ⚡ 電力相關 (3 檔)
    ('1101', '台泥'), ('6506', '雙鴻'), ('6411', '晶焱'),
    
    # 📡 通訊相關 (4 檔)
    ('3045', '台灣大'), ('4904', '遠傳'), ('2049', '上銀'), ('3008', '大立光'),
    
    # 🏆 其他重要 (3 檔)
    ('6505', '台塑化'), ('2207', '和泰車'), ('2880', '華南金'),
    
    # 📈 高年化報酬率 TOP 10
    ('2409', '友達'), ('3034', '聯詠'), ('2301', '光寶科'), ('2408', '南亞科'),
    ('2344', '華邦電'), ('3481', '群創'), ('6176', '瑞儀'), ('2371', '大同'),
    ('6414', '樺漢'), ('3661', '世芯-KY'),
]

def fetch_goodinfo_data(symbol):
    """從 Goodinfo 抓取完整資料"""
    try:
        url = f"https://goodinfo.tw/tw/StockDividendPolicy.asp?STOCK_ID={symbol}"
        
        headers = {
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        }
        
        response = requests.get(url, headers=headers, timeout=15)
        response.encoding = 'cp950'  # Goodinfo 使用 Big5
        
        soup = BeautifulSoup(response.text, 'html.parser')
        
        result = {
            'symbol': symbol,
            'eps': 0,
            'payout_ratio': 0,
            'consecutive_10_years': False,
            'dividend_count_10year': 0,
            'dividend_history': []
        }
        
        # 找到配息表格
        tables = soup.find_all('table', {'class': 'solid_1_padding_4_0_tbl'})
        
        for table in tables:
            rows = table.find_all('tr')
            
            if len(rows) > 1:
                headers = [th.get_text(strip=True) for th in rows[0].find_all('th')]
                
                # 檢查是否為配息表格
                if '年度' in headers or '發放年度' in headers:
                    dividend_years = 0
                    
                    for row in rows[1:11]:  # 只取最近 10 年
                        cells = row.find_all('td')
                        if len(cells) >= 7:
                            try:
                                year = cells[0].get_text(strip=True)
                                cash_div = cells[2].get_text(strip=True).replace(',', '')
                                eps_text = cells[5].get_text(strip=True).replace(',', '')
                                
                                cash_dividend = float(cash_div) if cash_div and cash_div != '-' else 0
                                eps = float(eps_text) if eps_text and eps_text != '-' else 0
                                
                                # 記錄配息歷史
                                result['dividend_history'].append({
                                    'year': year,
                                    'cash_dividend': cash_dividend,
                                    'eps': eps
                                })
                                
                                # 計算連續配息
                                if cash_dividend > 0:
                                    dividend_years += 1
                                else:
                                    break  # 中斷連續配息
                                
                            except Exception as e:
                                continue
                    
                    # 更新結果
                    result['dividend_count_10year'] = dividend_years
                    result['consecutive_10_years'] = dividend_years >= 10
                    
                    # 取最近一年的 EPS
                    if result['dividend_history']:
                        latest = result['dividend_history'][0]
                        result['eps'] = latest['eps']
                        
                        # 計算配息率
                        if latest['eps'] > 0 and latest['cash_dividend'] > 0:
                            result['payout_ratio'] = (latest['cash_dividend'] / latest['eps']) * 100
        
        return result
        
    except Exception as e:
        print(f"  ❌ 抓取失敗: {e}")
        return None

def classify_company_type(data, dividend):
    """分類公司類型"""
    eps = data.get('eps', 0)
    payout_ratio = data.get('payout_ratio', 0)
    consecutive_years = data.get('dividend_count_10year', 0)
    
    if payout_ratio > 100:
        return {
            'company_type': '吃老本',
            'type_emoji': '💀',
            'type_description': 'EPS不足以支撐配息'
        }
    elif payout_ratio > 80:
        return {
            'company_type': '撐配息',
            'type_emoji': '⚠️',
            'type_description': '配息率偏高'
        }
    elif 50 <= payout_ratio <= 70 and eps > 0:
        return {
            'company_type': '成長型好公司',
            'type_emoji': '🚀',
            'type_description': 'EPS成長且配息穩健'
        }
    elif consecutive_years >= 10:
        return {
            'company_type': '穩健型',
            'type_emoji': '🛡️',
            'type_description': '連續配息10年以上'
        }
    elif dividend > 0:
        return {
            'company_type': '高配息',
            'type_emoji': '💰',
            'type_description': '股息收益為主'
        }
    else:
        return {
            'company_type': '其他',
            'type_emoji': '❓',
            'type_description': '資料不足'
        }

def main():
    """主程式"""
    print("=" * 60)
    print("🦈 快速更新 82 支股票基本面資料")
    print("=" * 60)
    
    # 讀取現有的 dividend_data.json
    try:
        with open('dividend_data.json', 'r', encoding='utf-8') as f:
            existing_data = json.load(f)
        print(f"✅ 載入現有資料: {len(existing_data)} 支股票")
    except:
        existing_data = {}
        print("⚠️ 無現有資料，將建立新檔案")
    
    print(f"\n🚀 開始更新 {len(STOCK_LIST)} 支股票...\n")
    
    updated_count = 0
    failed_list = []
    
    for i, (symbol, name) in enumerate(STOCK_LIST, 1):
        print(f"[{i}/{len(STOCK_LIST)}] 📊 {symbol} {name}")
        
        try:
            # 抓取資料
            data = fetch_goodinfo_data(symbol)
            
            if data:
                # 合併到現有資料
                if symbol in existing_data:
                    existing_data[symbol].update(data)
                else:
                    existing_data[symbol] = data
                
                # 取得股息（從現有資料）
                dividend = existing_data[symbol].get('total_dividend', 0)
                
                # 分類公司類型
                company_type = classify_company_type(data, dividend)
                existing_data[symbol].update(company_type)
                existing_data[symbol]['name'] = name
                
                # 顯示結果
                print(f"  ✅ EPS: {data['eps']:.2f} | 配息率: {data['payout_ratio']:.1f}% | 連續: {data['dividend_count_10year']}年")
                print(f"  🏷️ 類型: {company_type['type_emoji']} {company_type['company_type']}")
                
                updated_count += 1
            else:
                failed_list.append(f"{symbol} {name}")
                print(f"  ⚠️ 無法取得資料")
            
        except Exception as e:
            failed_list.append(f"{symbol} {name}")
            print(f"  ❌ 處理失敗: {e}")
        
        # 每 10 支儲存一次
        if i % 10 == 0:
            with open('dividend_data.json', 'w', encoding='utf-8') as f:
                json.dump(existing_data, f, ensure_ascii=False, indent=2)
            print(f"\n💾 已儲存進度 ({i}/{len(STOCK_LIST)})\n")
        
        # 延遲避免被 ban
        time.sleep(3)
    
    # 最終儲存
    with open('dividend_data.json', 'w', encoding='utf-8') as f:
        json.dump(existing_data, f, ensure_ascii=False, indent=2)
    
    # 備份
    backup_file = f"dividend_data_backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
    with open(backup_file, 'w', encoding='utf-8') as f:
        json.dump(existing_data, f, ensure_ascii=False, indent=2)
    
    # 總結
    print("\n" + "=" * 60)
    print("📊 更新完成！")
    print("=" * 60)
    print(f"✅ 成功: {updated_count}/{len(STOCK_LIST)}")
    print(f"❌ 失敗: {len(failed_list)}")
    
    if failed_list:
        print(f"\n⚠️ 失敗清單:")
        for item in failed_list:
            print(f"  - {item}")
    
    print(f"\n💾 資料已儲存至: dividend_data.json")
    print(f"📦 備份檔案: {backup_file}")

if __name__ == '__main__':
    main()
