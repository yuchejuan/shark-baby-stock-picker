#!/usr/bin/env python3
"""
台股基本面資料更新工具
資料來源：
1. https://mops.twse.com.tw/mops/ - 公開資訊觀測站（EPS、配息率）
2. https://www.twse.com.tw/zh/listed/selection-criteria.html - 台股漲幅排行
3. Goodinfo - 10年配息記錄
"""

import json
import time
import requests
from bs4 import BeautifulSoup
from datetime import datetime
import sys

class FundamentalDataUpdater:
    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
        })
        self.results = {}
        
    def fetch_eps_data(self, symbol):
        """從公開資訊觀測站抓取 EPS 資料"""
        try:
            # 使用 MOPS API 抓取財報資料
            year = datetime.now().year - 1911  # 民國年
            url = f"https://mops.twse.com.tw/mops/web/ajax_t163sb04"
            
            payload = {
                'encodeURIComponent': '1',
                'step': '1',
                'firstin': '1',
                'off': '1',
                'co_id': symbol,
                'year': str(year)
            }
            
            response = self.session.post(url, data=payload, timeout=10)
            
            if response.status_code == 200:
                # 解析 HTML 取得 EPS
                soup = BeautifulSoup(response.text, 'html.parser')
                tables = soup.find_all('table')
                
                if len(tables) > 0:
                    # 尋找 EPS 欄位
                    for table in tables:
                        rows = table.find_all('tr')
                        for row in rows:
                            cells = row.find_all('td')
                            if len(cells) > 1:
                                # 檢查是否包含 EPS 相關文字
                                text = cells[0].get_text(strip=True)
                                if 'EPS' in text or '每股盈餘' in text:
                                    try:
                                        eps_value = float(cells[1].get_text(strip=True))
                                        return eps_value
                                    except:
                                        continue
            
            return None
            
        except Exception as e:
            print(f"  ❌ 抓取 {symbol} EPS 失敗: {e}")
            return None
    
    def fetch_dividend_history(self, symbol):
        """從 Goodinfo 抓取歷年配息記錄"""
        try:
            url = f"https://goodinfo.tw/tw/StockDividendPolicy.asp?STOCK_ID={symbol}"
            response = self.session.get(url, timeout=10)
            
            if response.status_code == 200:
                response.encoding = 'utf-8'
                soup = BeautifulSoup(response.text, 'html.parser')
                
                # 找到配息表格
                table = soup.find('table', {'class': 'solid_1_padding_4_0_tbl'})
                
                if table:
                    rows = table.find_all('tr')[1:]  # 跳過表頭
                    dividend_history = []
                    
                    for row in rows[:10]:  # 只取最近 10 年
                        cells = row.find_all('td')
                        if len(cells) >= 3:
                            try:
                                year = cells[0].get_text(strip=True)
                                cash_dividend = float(cells[2].get_text(strip=True) or 0)
                                dividend_history.append({
                                    'year': year,
                                    'cash_dividend': cash_dividend
                                })
                            except:
                                continue
                    
                    # 計算連續配息年數
                    consecutive_years = 0
                    for record in dividend_history:
                        if record['cash_dividend'] > 0:
                            consecutive_years += 1
                        else:
                            break
                    
                    return {
                        'history': dividend_history,
                        'consecutive_10_years': consecutive_years >= 10,
                        'dividend_count_10year': consecutive_years
                    }
            
            return None
            
        except Exception as e:
            print(f"  ❌ 抓取 {symbol} 配息歷史失敗: {e}")
            return None
    
    def fetch_top_gainers(self):
        """從台灣證交所抓取漲幅排行 Top 50"""
        try:
            # 使用證交所 API
            url = "https://www.twse.com.tw/rwd/zh/afterTrading/MI_INDEX"
            params = {
                'date': datetime.now().strftime('%Y%m%d'),
                'type': 'ALL',
                'response': 'json'
            }
            
            response = self.session.get(url, params=params, timeout=10)
            
            if response.status_code == 200:
                data = response.json()
                
                if 'data9' in data:
                    # data9 是漲幅排行
                    top_gainers = []
                    for row in data['data9'][:50]:  # 取前 50
                        try:
                            symbol = row[0].strip()
                            name = row[1].strip()
                            change_percent = float(row[9].strip('%'))
                            
                            top_gainers.append({
                                'symbol': symbol,
                                'name': name,
                                'change_percent': change_percent
                            })
                        except:
                            continue
                    
                    return top_gainers
            
            return None
            
        except Exception as e:
            print(f"❌ 抓取漲幅排行失敗: {e}")
            return None
    
    def calculate_company_type(self, eps, payout_ratio, dividend, consecutive_years):
        """
        判斷公司類型
        - 成長型好公司：EPS成長 + 配息率50-70%
        - 穩健型：EPS穩定 + 連續配息10年
        - 高配息：殖利率 > 6%
        - 撐配息：配息率 > 80%
        - 吃老本：配息率 > 100%
        """
        
        if payout_ratio > 100:
            return {
                'company_type': '吃老本',
                'type_emoji': '💀',
                'type_description': 'EPS不足以支撐配息，可能侵蝕資本'
            }
        elif payout_ratio > 80:
            return {
                'company_type': '撐配息',
                'type_emoji': '⚠️',
                'type_description': '配息率偏高，成長空間有限'
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
                'type_description': '連續配息10年以上，穩定可靠'
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
    
    def update_stock_data(self, symbol, name):
        """更新單一股票的基本面資料"""
        print(f"\n📊 更新 {symbol} {name}...")
        
        result = {
            'symbol': symbol,
            'name': name,
            'updated_at': datetime.now().isoformat()
        }
        
        # 1. 抓取 EPS
        print(f"  ⏳ 抓取 EPS...")
        eps = self.fetch_eps_data(symbol)
        if eps:
            result['eps'] = eps
            print(f"  ✅ EPS: {eps}")
        else:
            result['eps'] = 0
            print(f"  ⚠️ EPS 資料無法取得")
        
        time.sleep(1)  # 避免被 ban
        
        # 2. 抓取配息歷史
        print(f"  ⏳ 抓取配息歷史...")
        dividend_history = self.fetch_dividend_history(symbol)
        if dividend_history:
            result.update(dividend_history)
            print(f"  ✅ 連續配息: {dividend_history['dividend_count_10year']} 年")
        else:
            result['consecutive_10_years'] = False
            result['dividend_count_10year'] = 0
            print(f"  ⚠️ 配息歷史無法取得")
        
        time.sleep(2)  # Goodinfo 需要更長的間隔
        
        # 3. 計算配息率
        if 'eps' in result and result['eps'] > 0:
            # 從現有資料取得股息
            result['payout_ratio'] = 0  # 需要從外部資料取得
        
        # 4. 判斷公司類型
        company_type = self.calculate_company_type(
            result.get('eps', 0),
            result.get('payout_ratio', 0),
            0,  # dividend 需要從外部取得
            result.get('dividend_count_10year', 0)
        )
        result.update(company_type)
        
        return result
    
    def update_all_stocks(self, stock_list):
        """批次更新所有股票"""
        print(f"\n🚀 開始更新 {len(stock_list)} 支股票的基本面資料...\n")
        
        for i, stock in enumerate(stock_list, 1):
            symbol = stock['symbol']
            name = stock['name']
            
            print(f"[{i}/{len(stock_list)}] 處理 {symbol} {name}")
            
            try:
                result = self.update_stock_data(symbol, name)
                self.results[symbol] = result
                
            except Exception as e:
                print(f"  ❌ 處理失敗: {e}")
                continue
            
            # 每 10 支儲存一次
            if i % 10 == 0:
                self.save_results()
                print(f"\n💾 已儲存進度 ({i}/{len(stock_list)})\n")
        
        # 最終儲存
        self.save_results()
        print(f"\n✅ 全部完成！共更新 {len(self.results)} 支股票")
    
    def save_results(self):
        """儲存結果到 JSON"""
        output_file = 'fundamental_data.json'
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(self.results, f, ensure_ascii=False, indent=2)
        print(f"  💾 已儲存至 {output_file}")
    
    def fetch_and_save_top_gainers(self):
        """抓取並儲存漲幅排行"""
        print("\n📈 抓取台股漲幅排行 Top 50...\n")
        
        top_gainers = self.fetch_top_gainers()
        
        if top_gainers:
            output_file = 'top_gainers.json'
            with open(output_file, 'w', encoding='utf-8') as f:
                json.dump({
                    'updated_at': datetime.now().isoformat(),
                    'data': top_gainers
                }, f, ensure_ascii=False, indent=2)
            
            print(f"✅ 漲幅排行已儲存至 {output_file}")
            print(f"\n📊 Top 10 漲幅股票：")
            for i, stock in enumerate(top_gainers[:10], 1):
                print(f"  {i}. {stock['symbol']} {stock['name']} (+{stock['change_percent']}%)")
        else:
            print("❌ 無法取得漲幅排行資料")

def main():
    """主程式"""
    print("=" * 60)
    print("🦈 台股基本面資料更新工具")
    print("=" * 60)
    
    # 讀取現有的股票清單
    try:
        with open('dividend_tracker.html', 'r', encoding='utf-8') as f:
            content = f.read()
            
        # 從 HTML 中解析股票清單（簡化版）
        # 實際應該從 stockDatabase 解析
        stock_list = []
        
        # 這裡需要手動定義要更新的股票清單
        # 或從 dividend_data.json 讀取
        
        with open('dividend_data.json', 'r', encoding='utf-8') as f:
            dividend_data = json.load(f)
            stock_list = [{'symbol': k, 'name': v.get('name', '')} for k, v in dividend_data.items()]
        
        print(f"\n✅ 載入 {len(stock_list)} 支股票")
        
    except Exception as e:
        print(f"❌ 載入股票清單失敗: {e}")
        sys.exit(1)
    
    # 建立更新器
    updater = FundamentalDataUpdater()
    
    # 選單
    print("\n請選擇操作：")
    print("1. 更新所有股票基本面資料（需要約 30-60 分鐘）")
    print("2. 只抓取漲幅排行 Top 50")
    print("3. 更新指定股票")
    
    choice = input("\n請輸入選項 (1/2/3): ").strip()
    
    if choice == '1':
        updater.update_all_stocks(stock_list)
    elif choice == '2':
        updater.fetch_and_save_top_gainers()
    elif choice == '3':
        symbol = input("請輸入股票代號: ").strip()
        name = input("請輸入股票名稱: ").strip()
        result = updater.update_stock_data(symbol, name)
        updater.results[symbol] = result
        updater.save_results()
    else:
        print("❌ 無效選項")

if __name__ == '__main__':
    main()
