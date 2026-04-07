#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
KD 評分優化腳本
自動更新 stock_analyzer.go 中的 KD 評分邏輯
"""

import re
import sys

def apply_kd_optimization():
    file_path = "/home/administrator/.openclaw/workspace/stock_analyzer.go"
    
    # 讀取原始檔案
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
    except Exception as e:
        print(f"❌ 讀取檔案失敗：{e}")
        return False
    
    # 備份
    backup_path = file_path + ".backup"
    try:
        with open(backup_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"✅ 備份完成：{backup_path}")
    except Exception as e:
        print(f"⚠️  備份失敗：{e}")
    
    # 定義要替換的舊程式碼（使用正則表達式）
    old_pattern = r'''(\t// KD 評分 \(15分\)\n\tif stock\.KD == "超賣" \{\n\t\tscore \+= 15\n\t\tadvantages = append\(advantages, "KD超賣"\)\n\t\} else if stock\.KD == "中性" \{\n\t\tscore \+= 10\n\t\} else \{\n\t\tscore \+= 5\n\t\})'''
    
    # 新的程式碼
    new_code = '''	// KD 評分 (15分) - 優化版
	if stock.KD == "超賣" {
		score += 15
		advantages = append(advantages, "KD超賣")
	} else if stock.KD == "偏低" {
		score += 12
		advantages = append(advantages, "KD偏低")
	} else if stock.KD == "偏多" {
		score += 10
	} else if stock.KD == "中性" {
		score += 8
	} else if stock.KD == "偏空" {
		score += 6
	} else if stock.KD == "偏高" {
		score += 4
	} else if stock.KD == "超買" {
		score += 2
	} else {
		score += 5
	}'''
    
    # 執行替換
    new_content, count = re.subn(old_pattern, new_code, content, flags=re.MULTILINE)
    
    if count == 0:
        print("⚠️  未找到匹配的程式碼，嘗試簡化版本...")
        
        # 簡化版本的匹配
        simple_pattern = r'// KD 評分 \(15分\).*?(?=\n\t// 設定預設訊號)'
        new_content, count = re.subn(simple_pattern, new_code, content, flags=re.DOTALL)
    
    if count > 0:
        # 寫入新檔案
        try:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"✅ KD 評分優化完成！（替換了 {count} 處）")
            return True
        except Exception as e:
            print(f"❌ 寫入檔案失敗：{e}")
            return False
    else:
        print("❌ 未找到要替換的程式碼")
        print("\n請手動檢查 stock_analyzer.go 第 429-437 行")
        return False

if __name__ == "__main__":
    success = apply_kd_optimization()
    sys.exit(0 if success else 1)
