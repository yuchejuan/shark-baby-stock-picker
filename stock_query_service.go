package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// ========== 引入 TWSE 爬蟲（部分複製） ==========
// 為了簡化，這裡直接引用必要的函數

// K線資料
type KLine struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Amount int64   `json:"amount"`
	Change float64 `json:"change"`
}

// 股票資料
type StockData struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	CurrentPrice  float64 `json:"current_price"`
	KLines        []KLine `json:"klines"`
	PE            float64 `json:"pe"`
	PB            float64 `json:"pb"`
	DividendYield float64 `json:"dividend_yield"`
	UpdateTime    string  `json:"update_time"`
}

// 技術分析結果（與 daily_stock_picker_all.go 格式一致）
type TechnicalAnalysis struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Score     int     `json:"score"`
	RSI       float64 `json:"rsi"`
	MACD      string  `json:"macd"`
	KD        string  `json:"kd"`
	Signal    string  `json:"signal"`
	Advantage string  `json:"advantage"`
	MA5       float64 `json:"ma5"`
	MA20      float64 `json:"ma20"`
	MA60      float64 `json:"ma60"`
	MATrend   string  `json:"ma_trend"`
	PE        float64 `json:"pe"`
	PB        float64 `json:"pb"`
	DY        float64 `json:"dividend_yield"`
}

// ========== 技術指標計算（真實算法） ==========

// 計算 RSI（14 天）
func calculateRSI(klines []KLine, period int) float64 {
	if len(klines) < period+1 {
		return 50 // 預設值
	}
	
	gains := 0.0
	losses := 0.0
	
	// 計算最近 period 天的漲跌
	for i := len(klines) - period; i < len(klines); i++ {
		change := klines[i].Change
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}
	
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	
	if avgLoss == 0 {
		return 100
	}
	
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	
	return rsi
}

// 計算移動平均線
func calculateMA(klines []KLine, period int) float64 {
	if len(klines) < period {
		return 0
	}
	
	sum := 0.0
	for i := len(klines) - period; i < len(klines); i++ {
		sum += klines[i].Close
	}
	
	return sum / float64(period)
}

// 計算 MACD（12, 26, 9）
func calculateMACD(klines []KLine) string {
	if len(klines) < 26 {
		return "中性"
	}
	
	ema12 := calculateEMA(klines, 12)
	ema26 := calculateEMA(klines, 26)
	
	dif := ema12 - ema26
	
	if dif > 0 {
		return "多頭"
	} else if dif < 0 {
		return "空頭"
	}
	return "中性"
}

// 計算 EMA
func calculateEMA(klines []KLine, period int) float64 {
	if len(klines) < period {
		return 0
	}
	
	// 簡化版：使用 SMA 作為初始值
	k := 2.0 / float64(period+1)
	ema := calculateMA(klines, period)
	
	for i := len(klines) - period; i < len(klines); i++ {
		ema = klines[i].Close*k + ema*(1-k)
	}
	
	return ema
}

// 計算 KD（9, 3, 3）
func calculateKD(klines []KLine) string {
	if len(klines) < 9 {
		return "中性"
	}
	
	// 取最近 9 天
	recent := klines[len(klines)-9:]
	
	// 找最高價和最低價
	high := recent[0].High
	low := recent[0].Low
	for _, k := range recent {
		if k.High > high {
			high = k.High
		}
		if k.Low < low {
			low = k.Low
		}
	}
	
	// 計算 RSV
	currentClose := recent[len(recent)-1].Close
	rsv := 0.0
	if high != low {
		rsv = (currentClose - low) / (high - low) * 100
	}
	
	// 簡化版 K 值（通常需要遞迴計算）
	k := rsv
	
	if k > 80 {
		return "超買"
	} else if k < 20 {
		return "超賣"
	}
	return "中性"
}

// ========== 評分系統（與 daily_stock_picker_all.go 一致） ==========

func calculateScore(rsi float64, macd string, kd string, maTrend string, price float64, ma20 float64, ma60 float64) int {
	score := 30 // 基礎分
	
	// RSI 評分
	if rsi < 30 {
		score += 15
	} else if rsi < 40 {
		score += 10
	} else if rsi <= 60 {
		score += 5
	} else if rsi > 70 {
		score -= 10
	}
	
	// MACD 評分
	if macd == "多頭" {
		score += 15
	} else if macd == "空頭" {
		score -= 10
	}
	
	// KD 評分
	if kd == "超賣" {
		score += 12
	} else if kd == "超買" {
		score -= 12
	} else {
		score += 5
	}
	
	// 均線評分
	if maTrend == "多頭" || maTrend == "多頭排列" {
		score += 15
	} else if maTrend == "空頭" {
		score -= 15
	}
	
	// 價格位置評分
	if price > ma20 && price > ma60 {
		score += 8
	}
	
	return max(0, min(100, score))
}

func determineSignal(score int) string {
	if score >= 70 {
		return "買點"
	}
	if score <= 40 {
		return "賣點"
	}
	return "中性"
}

func generateAdvantage(rsi float64, macd string, kd string, maTrend string) string {
	reasons := []string{}
	
	if macd == "多頭" {
		reasons = append(reasons, "MACD黃金交叉")
	}
	
	if kd == "超賣" {
		reasons = append(reasons, "KD超賣")
	}
	
	if rsi < 30 {
		reasons = append(reasons, "RSI超賣反彈機會")
	}
	
	if rsi < 40 && macd == "多頭" {
		reasons = append(reasons, "RSI偏低 + MACD黃金交叉")
	}
	
	if maTrend == "多頭排列" {
		reasons = append(reasons, "多頭排列")
	}
	
	if len(reasons) == 0 {
		return "技術面中性"
	}
	
	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += " + " + reasons[i]
	}
	
	return result
}

// ========== TWSE 爬蟲函數 ==========

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	if s == "-" || s == "" || s == "N/A" {
		return 0
	}
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

func parseInt(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	if s == "-" || s == "" {
		return 0
	}
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func fetchStockData(symbol string) (*StockData, error) {
	// 1. 查詢股票名稱
	name, err := queryStockCode(symbol)
	if err != nil {
		return nil, err
	}
	
	// 2. 取得 K 線
	klines, err := fetchRecentKLines(symbol, 60)
	if err != nil || len(klines) == 0 {
		return nil, fmt.Errorf("無法取得 K 線數據")
	}
	
	// 3. 取得本益比
	pe, pb, dy, _ := fetchPERatio(symbol)
	
	return &StockData{
		Symbol:        symbol,
		Name:          name,
		CurrentPrice:  klines[len(klines)-1].Close,
		KLines:        klines,
		PE:            pe,
		PB:            pb,
		DividendYield: dy,
		UpdateTime:    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func queryStockCode(symbol string) (string, error) {
	url := fmt.Sprintf("https://www.twse.com.tw/zh/api/codeQuery?query=%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	suggestions := result["suggestions"].([]interface{})
	if len(suggestions) == 0 {
		return "", fmt.Errorf("股票代號 %s 不存在", symbol)
	}
	
	// 檢查是否為「無符合之代碼或名稱」
	firstSuggestion := suggestions[0].(string)
	if strings.Contains(firstSuggestion, "無符合") || strings.Contains(firstSuggestion, "查無") {
		return "", fmt.Errorf("股票代號 %s 不存在或已下市", symbol)
	}
	
	parts := strings.Split(firstSuggestion, "\t")
	if len(parts) < 2 {
		return "", fmt.Errorf("無法解析股票名稱，原始資料: %s", firstSuggestion)
	}
	return parts[1], nil
}

func fetchRecentKLines(symbol string, days int) ([]KLine, error) {
	allKLines := []KLine{}
	now := time.Now()
	
	for i := 0; i < 3; i++ {
		targetDate := now.AddDate(0, -i, 0)
		klines, _ := fetchMonthlyKLines(symbol, targetDate.Year(), int(targetDate.Month()))
		allKLines = append(allKLines, klines...)
		time.Sleep(500 * time.Millisecond)
	}
	
	if len(allKLines) > days {
		allKLines = allKLines[len(allKLines)-days:]
	}
	
	return allKLines, nil
}

func fetchMonthlyKLines(symbol string, year int, month int) ([]KLine, error) {
	dateStr := fmt.Sprintf("%d%02d01", year, month)
	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/STOCK_DAY?date=%s&stockNo=%s&response=json", dateStr, symbol)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	
	if result["stat"] != "OK" {
		return nil, fmt.Errorf("API 失敗")
	}
	
	data := result["data"].([]interface{})
	klines := []KLine{}
	
	for _, row := range data {
		r := row.([]interface{})
		if len(r) < 9 {
			continue
		}
		
		kline := KLine{
			Date:   r[0].(string),
			Open:   parseFloat(r[3].(string)),
			High:   parseFloat(r[4].(string)),
			Low:    parseFloat(r[5].(string)),
			Close:  parseFloat(r[6].(string)),
			Change: parseFloat(r[7].(string)),
			Volume: parseInt(r[1].(string)),
			Amount: parseInt(r[2].(string)),
		}
		klines = append(klines, kline)
	}
	
	return klines, nil
}

func fetchPERatio(symbol string) (float64, float64, float64, error) {
	yesterday := time.Now().AddDate(0, 0, -1)
	dateStr := yesterday.Format("20060102")
	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/BWIBBU_d?date=%s&selectType=ALL&response=json", dateStr)
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, 0, fmt.Errorf("JSON 解析失敗")
	}
	
	// 檢查 data 是否存在
	if result["data"] == nil {
		return 0, 0, 0, nil // 沒有本益比資料，返回 0 但不報錯
	}
	
	data, ok := result["data"].([]interface{})
	if !ok {
		return 0, 0, 0, nil // 格式不對，返回 0
	}
	
	for _, row := range data {
		r, ok := row.([]interface{})
		if !ok {
			continue
		}
		if len(r) < 7 {
			continue
		}
		if fmt.Sprintf("%v", r[0]) == symbol {
			pe := parseFloat(fmt.Sprintf("%v", r[5]))
			pb := parseFloat(fmt.Sprintf("%v", r[6]))
			dy := parseFloat(fmt.Sprintf("%v", r[3]))
			return pe, pb, dy, nil
		}
	}
	
	return 0, 0, 0, nil // 找不到也不報錯，返回 0 值
}

// ========== 主查詢函數 ==========

func analyzeStockTechnical(symbol string) (*TechnicalAnalysis, error) {
	fmt.Printf("🔍 查詢股票 %s...\n", symbol)
	
	// 1. 取得股票資料
	stock, err := fetchStockData(symbol)
	if err != nil {
		return nil, err
	}
	
	fmt.Printf("  ✅ %s (%s) - 現價: %.2f\n", stock.Name, stock.Symbol, stock.CurrentPrice)
	
	// 2. 計算技術指標
	rsi := calculateRSI(stock.KLines, 14)
	macd := calculateMACD(stock.KLines)
	kd := calculateKD(stock.KLines)
	
	ma5 := calculateMA(stock.KLines, 5)
	ma20 := calculateMA(stock.KLines, 20)
	ma60 := calculateMA(stock.KLines, 60)
	
	// 判斷均線趨勢
	maTrend := "中性"
	if ma5 > ma20 && ma20 > ma60 {
		maTrend = "多頭排列"
	} else if ma60 > ma20 && ma20 > ma5 {
		maTrend = "空頭"
	} else if ma5 > ma20 {
		maTrend = "多頭"
	}
	
	// 3. 計算評分
	score := calculateScore(rsi, macd, kd, maTrend, stock.CurrentPrice, ma20, ma60)
	signal := determineSignal(score)
	advantage := generateAdvantage(rsi, macd, kd, maTrend)
	
	fmt.Printf("  📊 評分: %d | 訊號: %s\n", score, signal)
	
	return &TechnicalAnalysis{
		Symbol:    stock.Symbol,
		Name:      stock.Name,
		Price:     stock.CurrentPrice,
		Score:     score,
		RSI:       rsi,
		MACD:      macd,
		KD:        kd,
		Signal:    signal,
		Advantage: advantage,
		MA5:       ma5,
		MA20:      ma20,
		MA60:      ma60,
		MATrend:   maTrend,
		PE:        stock.PE,
		PB:        stock.PB,
		DY:        stock.DividendYield,
	}, nil
}

// ========== 主程式 ==========

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❌ 用法: go run stock_query_service.go <股票代號>")
		fmt.Println("範例: go run stock_query_service.go 2330")
		os.Exit(1)
	}
	
	symbol := os.Args[1]
	
	fmt.Println("🦈 鯊魚寶寶股票查詢服務")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	analysis, err := analyzeStockTechnical(symbol)
	if err != nil {
		log.Fatalf("❌ 查詢失敗: %v", err)
	}
	
	// 輸出 JSON（供網頁介面使用）
	output, _ := json.MarshalIndent(analysis, "", "  ")
	fmt.Println("\n" + string(output))
	
	fmt.Println("\n✅ 查詢完成")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
