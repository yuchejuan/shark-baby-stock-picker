// KD 評分優化版本 - 貼到 stock_analyzer.go 的 CalculateScore 函數中

// 替換這段：
/*
	// KD 評分 (15分)
	if stock.KD == "超賣" {
		score += 15
		advantages = append(advantages, "KD超賣")
	} else if stock.KD == "中性" {
		score += 10
	} else {
		score += 5
	}
*/

// 改為：
/*
	// KD 評分 (15分) - 優化版
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
	}
*/
