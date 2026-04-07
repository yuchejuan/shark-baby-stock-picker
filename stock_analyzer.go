// stock_analyzer.go - 股票分析函式庫
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TWSE API 回應結構
type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

// 股票歷史資料
type StockHistoryData struct {
	Date   string
	Close  float64
	Volume int64
}

// 股票資訊
type StockInfo struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	MA5       float64 `json:"ma5"`
	MA10      float64 `json:"ma10"`
	MA20      float64 `json:"ma20"`
	MA60      float64 `json:"ma60"`
	RSI14     float64 `json:"rsi"`
	MACD      string  `json:"macd"`
	KD        string  `json:"kd"`
	Signal    string  `json:"signal"`
	Score     int     `json:"score"`
	MATrend   string  `json:"ma_trend"`
	Advantage string  `json:"advantage"`
}

// ========== TWSE API 函式 ==========

func GetTWSEHistoricalData(code string, yearMonth string) ([]StockHistoryData, error) {
	if len(yearMonth) >= 6 {
		yearMonth = yearMonth[:6]
	}
	
	url := fmt.Sprintf("https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=%s01&stockNo=%s", yearMonth, code)
	
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var twseResp TWSEResponse
	if err := json.Unmarshal(body, &twseResp); err != nil {
		return nil, err
	}
	
	if twseResp.Stat != "OK" {
		return nil, fmt.Errorf("TWSE API 錯誤: %s", twseResp.Stat)
	}
	
	if len(twseResp.Data) == 0 {
		return nil, fmt.Errorf("查無資料")
	}
	
	var history []StockHistoryData
	
	for _, row := range twseResp.Data {
		if len(row) < 7 {
			continue
		}
		
		closeStr := strings.ReplaceAll(row[6], ",", "")
		volumeStr := strings.ReplaceAll(row[1], ",", "")
		
		var close float64
		var volume int64
		
		fmt.Sscanf(closeStr, "%f", &close)
		fmt.Sscanf(volumeStr, "%d", &volume)
		
		dateStr := row[0]
		dateParts := strings.Split(dateStr, "/")
		if len(dateParts) == 3 {
			year := 0
			fmt.Sscanf(dateParts[0], "%d", &year)
			year += 1911
			dateStr = fmt.Sprintf("%04d-%s-%s", year, dateParts[1], dateParts[2])
		}
		
		history = append(history, StockHistoryData{
			Date:   dateStr,
			Close:  close,
			Volume: volume,
		})
	}
	
	return history, nil
}

func GetRecentClosePrices(code string, days int) ([]float64, error) {
	now := time.Now()
	
	var allHistory []StockHistoryData
	
	// 查詢本月
	thisMonth := now.Format("200601")
	history1, err := GetTWSEHistoricalData(code, thisMonth)
	if err == nil {
		allHistory = append(allHistory, history1...)
	}
	
	// 查詢上個月
	if len(allHistory) < days {
		lastMonth := now.AddDate(0, -1, 0).Format("200601")
		history2, err := GetTWSEHistoricalData(code, lastMonth)
		if err == nil {
			allHistory = append(history2, allHistory...)
		}
		time.Sleep(300 * time.Millisecond)
	}
	
	// 查詢上上個月
	if len(allHistory) < days {
		twoMonthsAgo := now.AddDate(0, -2, 0).Format("200601")
		history3, err := GetTWSEHistoricalData(code, twoMonthsAgo)
		if err == nil {
			allHistory = append(history3, allHistory...)
		}
		time.Sleep(300 * time.Millisecond)
	}
	
	// 查詢三個月前
	if len(allHistory) < days {
		threeMonthsAgo := now.AddDate(0, -3, 0).Format("200601")
		history4, err := GetTWSEHistoricalData(code, threeMonthsAgo)
		if err == nil {
			allHistory = append(history4, allHistory...)
		}
		time.Sleep(300 * time.Millisecond)
	}
	
	if len(allHistory) == 0 {
		return nil, fmt.Errorf("無歷史資料")
	}
	
	start := len(allHistory) - days
	if start < 0 {
		start = 0
	}
	
	var closes []float64
	for i := start; i < len(allHistory); i++ {
		closes = append(closes, allHistory[i].Close)
	}
	
	return closes, nil
}

// ========== 技術指標計算 ==========

func CalculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	
	return sum / float64(period)
}

func CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0
	}
	
	gains := 0.0
	losses := 0.0
	
	for i := len(prices) - period; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}
	
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	
	if avgLoss == 0 {
		return 100.0
	}
	
	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	
	return rsi
}

func CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += prices[i]
	}
	ema := sum / float64(period)
	
	multiplier := 2.0 / float64(period+1)
	for i := period; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}
	
	return ema
}

// CalculateMACD - 優化版本：加入信號線判斷
func CalculateMACD(prices []float64) string {
	if len(prices) < 35 {
		return "中性"
	}
	
	// 計算 MACD 歷史值
	var macdHistory []float64
	for i := 26; i <= len(prices); i++ {
		ema12 := CalculateEMA(prices[:i], 12)
		ema26 := CalculateEMA(prices[:i], 26)
		macdHistory = append(macdHistory, ema12-ema26)
	}
	
	if len(macdHistory) < 9 {
		return "中性"
	}
	
	// 計算信號線（9日EMA）
	signal := CalculateEMA(macdHistory, 9)
	currentMACD := macdHistory[len(macdHistory)-1]
	
	// 計算前一日的 MACD 和信號線
	var prevMACD float64
	var prevSignal float64
	if len(macdHistory) >= 2 {
		prevMACD = macdHistory[len(macdHistory)-2]
		if len(macdHistory) >= 10 {
			prevSignal = CalculateEMA(macdHistory[:len(macdHistory)-1], 9)
		}
	}
	
	// 判斷黃金交叉/死亡交叉
	if currentMACD > signal && prevMACD <= prevSignal {
		return "黃金交叉"
	} else if currentMACD < signal && prevMACD >= prevSignal {
		return "死亡交叉"
	} else if currentMACD > signal {
		return "多頭"
	} else if currentMACD < signal {
		return "空頭"
	}
	
	return "中性"
}

// CalculateKD - 優化版本：加入平滑處理
func CalculateKD(prices []float64, period int) string {
	if len(prices) < period+2 {
		return "中性"
	}
	
	// 計算最近 3 天的 RSV 用於平滑
	var rsvHistory []float64
	for i := len(prices) - 3; i < len(prices); i++ {
		if i < period-1 {
			continue
		}
		
		recent := prices[i-period+1 : i+1]
		high := recent[0]
		low := recent[0]
		
		for _, p := range recent {
			if p > high {
				high = p
			}
			if p < low {
				low = p
			}
		}
		
		current := prices[i]
		rsv := 50.0
		if high != low {
			rsv = (current - low) / (high - low) * 100
		}
		rsvHistory = append(rsvHistory, rsv)
	}
	
	if len(rsvHistory) == 0 {
		return "中性"
	}
	
	// 平滑計算 K 值（使用最近 3 個 RSV 的加權平均）
	k := 50.0
	if len(rsvHistory) >= 3 {
		// K = 前一日K × 2/3 + 當日RSV × 1/3 的簡化版本
		// 使用 3 個 RSV 的加權平均來近似
		k = (rsvHistory[0]*1 + rsvHistory[1]*2 + rsvHistory[2]*3) / 6
	} else if len(rsvHistory) >= 2 {
		k = (rsvHistory[0] + rsvHistory[1]*2) / 3
	} else {
		k = rsvHistory[0]
	}
	
	// D 值（K 值的 3 日移動平均）
	d := k
	if len(rsvHistory) >= 3 {
		k1 := (rsvHistory[0]*1 + rsvHistory[1]*2 + rsvHistory[2]*3) / 6
		k2 := (rsvHistory[0]*2 + rsvHistory[1]*3) / 5
		d = (k1 + k2) / 2
	}
	
	// 判斷超買超賣
	if k > 80 && d > 80 {
		return "超買"
	} else if k < 20 && d < 20 {
		return "超賣"
	} else if k > 80 || d > 80 {
		return "偏高"
	} else if k < 20 || d < 20 {
		return "偏低"
	} else if k > d {
		return "偏多"
	} else if k < d {
		return "偏空"
	}
	
	return "中性"
}

// ========== 評分系統 ==========

func CalculateScore(stock *StockInfo) int {
	score := 0
	advantages := []string{}
	
	// RSI 評分 (30分)
	if stock.RSI14 < 30 {
		score += 25
		advantages = append(advantages, "RSI超賣")
		stock.Signal = "買點"
	} else if stock.RSI14 < 40 {
		score += 20
		advantages = append(advantages, "RSI偏低")
	} else if stock.RSI14 > 70 {
		score += 5
		stock.Signal = "賣點"
	} else if stock.RSI14 > 60 && stock.RSI14 <= 70 {
		score += 15
	} else {
		score += 18
	}
	
	// MACD 評分 (30分) - 優化版
	if stock.MACD == "黃金交叉" {
		score += 30  // 真正的黃金交叉，加分更多
		advantages = append(advantages, "MACD黃金交叉")
		if stock.Signal == "" {
			stock.Signal = "買點"
		}
	} else if stock.MACD == "多頭" {
		score += 20  // 持續多頭
		advantages = append(advantages, "MACD多頭")
	} else if stock.MACD == "死亡交叉" {
		score += 0   // 死亡交叉，風險
		stock.Signal = "風險"
	} else if stock.MACD == "空頭" {
		score += 5
	} else {
		score += 15
	}
	
	// 均線趨勢評分 (30分)
	if stock.Price > stock.MA5 && stock.MA5 > stock.MA20 && stock.MA20 > stock.MA60 {
		score += 30
		stock.MATrend = "多頭"
		advantages = append(advantages, "多頭排列")
		if stock.Signal == "" {
			stock.Signal = "買點"
		}
	} else if stock.MA5 > stock.MA20 && stock.MA20 > stock.MA60 {
		score += 25
		stock.MATrend = "多頭"
	} else if stock.Price < stock.MA5 && stock.MA5 < stock.MA20 && stock.MA20 < stock.MA60 {
		score += 5
		stock.MATrend = "空頭"
	} else {
		score += 15
		stock.MATrend = "中性"
	}
	
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
	
	// 設定預設訊號
	if stock.Signal == "" {
		if score >= 70 {
			stock.Signal = "買點"
		} else if score < 40 {
			stock.Signal = "觀察"
		} else {
			stock.Signal = "觀察"
		}
	}
	
	// 組合優勢說明
	if len(advantages) > 0 {
		stock.Advantage = strings.Join(advantages, ", ")
	} else {
		stock.Advantage = "技術面中性"
	}
	
	return score
}

// ========== 主分析函式 ==========

func AnalyzeStock(code, name string) (*StockInfo, error) {
	// 取得歷史資料
	prices, err := GetRecentClosePrices(code, 70)
	if err != nil {
		return nil, err
	}
	
	if len(prices) < 20 {
		return nil, fmt.Errorf("資料不足")
	}
	
	// 計算技術指標
	currentPrice := prices[len(prices)-1]
	ma5 := CalculateMA(prices, 5)
	ma10 := CalculateMA(prices, 10)
	ma20 := CalculateMA(prices, 20)
	ma60 := CalculateMA(prices, 60)
	rsi := CalculateRSI(prices, 14)
	macd := CalculateMACD(prices)
	kd := CalculateKD(prices, 9)
	
	stock := &StockInfo{
		Code:    code,
		Name:    name,
		Price:   currentPrice,
		MA5:     ma5,
		MA10:    ma10,
		MA20:    ma20,
		MA60:    ma60,
		RSI14:   rsi,
		MACD:    macd,
		KD:      kd,
	}
	
	stock.Score = CalculateScore(stock)
	
	return stock, nil
}
