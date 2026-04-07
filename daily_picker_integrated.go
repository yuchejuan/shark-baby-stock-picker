// daily_picker_integrated.go - 每日選股 V3.0 整合版
// 81支股票池 + 20-60元區間 + TWSE API + 6大技術指標 + 100分制評分系統
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// ========== 股票池結構 ==========

type StockPoolConfig struct {
	Version     string       `json:"version"`
	LastUpdate  string       `json:"last_update"`
	Total       int          `json:"total"`
	ETF         ETFSection   `json:"etf"`
	Stocks      StockSection `json:"stocks"`
	PriceRanges []PriceRange `json:"price_ranges"`
}

type ETFSection struct {
	Count int               `json:"count"`
	List  map[string]string `json:"list"`
}

type StockSection struct {
	Count      int                            `json:"count"`
	Categories map[string]CategoryInfo        `json:"categories"`
}

type CategoryInfo struct {
	Name  string            `json:"name"`
	Count int               `json:"count"`
	List  map[string]string `json:"list"`
}

type PriceRange struct {
	Min   int    `json:"min"`
	Max   int    `json:"max"`
	Label string `json:"label"`
}

// ========== 資料結構 ==========

type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

type StockHistoryData struct {
	Date   string
	Close  float64
	Volume int64
}

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

type DailyPick struct {
	Rank      int     `json:"rank"`
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
}

type DailyReport struct {
	Date     string      `json:"date"`
	AllPicks []DailyPick `json:"all_picks"` // 全市場選股，無價格限制
	BestPick *DailyPick  `json:"best_pick"`
	Summary  Summary     `json:"summary"`
}

type Summary struct {
	TotalStocks  int    `json:"total_stocks"`
	PoolSize     int    `json:"pool_size"`
	TopPicks     int    `json:"top_picks"`
	BuySignals   int    `json:"buy_signals"`
	RiskWarnings int    `json:"risk_warnings"`
	UpdateTime   string `json:"update_time"`
}

// ========== 股票池載入 ==========

func loadStockPool(filepath string) (*StockPoolConfig, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("無法讀取股票池設定: %v", err)
	}
	
	var pool StockPoolConfig
	if err := json.Unmarshal(data, &pool); err != nil {
		return nil, fmt.Errorf("無法解析股票池設定: %v", err)
	}
	
	return &pool, nil
}

func getAllStocksFromPool(pool *StockPoolConfig) map[string]string {
	result := make(map[string]string)
	
	// 加入 ETF
	for code, name := range pool.ETF.List {
		result[code] = name
	}
	
	// 加入個股
	for _, category := range pool.Stocks.Categories {
		for code, name := range category.List {
			if _, exists := result[code]; !exists {
				result[code] = name
			}
		}
	}
	
	return result
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

func CalculateMACD(prices []float64) string {
	if len(prices) < 26 {
		return "中性"
	}
	
	ema12 := CalculateEMA(prices, 12)
	ema26 := CalculateEMA(prices, 26)
	macd := ema12 - ema26
	
	if macd > 0 {
		return "多頭"
	} else if macd < 0 {
		return "空頭"
	}
	
	return "中性"
}

func CalculateKD(prices []float64, period int) string {
	if len(prices) < period {
		return "中性"
	}
	
	recent := prices[len(prices)-period:]
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
	
	current := prices[len(prices)-1]
	rsv := 50.0
	if high != low {
		rsv = (current - low) / (high - low) * 100
	}
	
	k := rsv
	
	if k > 80 {
		return "超買"
	} else if k < 20 {
		return "超賣"
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
		stock.Signal = "風險"
	} else if stock.RSI14 > 60 && stock.RSI14 <= 70 {
		score += 15
	} else {
		score += 18
	}
	
	// MACD 評分 (25分)
	if stock.MACD == "多頭" {
		score += 25
		advantages = append(advantages, "MACD黃金交叉")
	} else if stock.MACD == "空頭" {
		score += 5
	} else {
		score += 15
	}
	
	// 均線趨勢評分 (30分)
	if stock.Price > stock.MA5 && stock.MA5 > stock.MA20 && stock.MA20 > stock.MA60 {
		score += 30
		stock.MATrend = "多頭排列"
		advantages = append(advantages, "多頭排列")
		if stock.Signal == "" {
			stock.Signal = "買點"
		}
	} else if stock.MA5 > stock.MA20 && stock.MA20 > stock.MA60 {
		score += 25
		stock.MATrend = "多頭"
	} else if stock.Price < stock.MA5 && stock.MA5 < stock.MA20 && stock.MA20 < stock.MA60 {
		score += 5
		stock.MATrend = "空頭排列"
		stock.Signal = "風險"
	} else {
		score += 15
		stock.MATrend = "中性"
	}
	
	// KD 評分 (15分)
	if stock.KD == "超賣" {
		score += 15
		advantages = append(advantages, "KD超賣")
	} else if stock.KD == "中性" {
		score += 10
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
			stock.Signal = "中性"
		}
	}
	
	// 組合優勢說明
	if len(advantages) > 0 {
		stock.Advantage = strings.Join(advantages, " + ")
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

// ========== 選股邏輯 ==========

func stockInfoToDailyPick(s *StockInfo, rank int) DailyPick {
	return DailyPick{
		Rank:      rank,
		Symbol:    s.Code,
		Name:      s.Name,
		Price:     s.Price,
		Score:     s.Score,
		RSI:       s.RSI14,
		MACD:      s.MACD,
		KD:        s.KD,
		Signal:    s.Signal,
		Advantage: s.Advantage,
		MA5:       s.MA5,
		MA20:      s.MA20,
		MA60:      s.MA60,
		MATrend:   s.MATrend,
	}
}

func analyzeAllStocks(stocks map[string]string) []DailyPick {
	var results []*StockInfo
	total := len(stocks)
	processed := 0

	fmt.Printf("🔍 全市場掃描 %d 支股票（無價格限制）...\n", total)

	for code, name := range stocks {
		stock, err := AnalyzeStock(code, name)
		if err != nil {
			continue
		}
		results = append(results, stock)

		processed++
		if processed%10 == 0 {
			fmt.Printf("  ⏳ 進度: %d/%d（%.0f%%）\n",
				processed, total, float64(processed)/float64(total)*100)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// 依評分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 全部轉為 DailyPick，加上排名
	var picks []DailyPick
	for i, s := range results {
		pick := stockInfoToDailyPick(s, i+1)
		picks = append(picks, pick)
	}

	fmt.Printf("✅ 分析完成：%d 支成功 / %d 支總計\n\n", len(picks), total)
	return picks
}

func analyzeStocksInRange(stocks map[string]string, minPrice, maxPrice float64) []DailyPick {
	var results []*StockInfo
	processed := 0
	total := len(stocks)
	
	fmt.Printf("🔍 處理 %.0f-%.0f 元區間...\n", minPrice, maxPrice)
	
	for code, name := range stocks {
		stock, err := AnalyzeStock(code, name)
		if err != nil {
			continue
		}
		
		// 價格篩選
		if stock.Price >= minPrice && stock.Price <= maxPrice {
			results = append(results, stock)
		}
		
		processed++
		if processed%10 == 0 {
			fmt.Printf("  ⏳ 進度: %d/%d\n", processed, total)
		}
		
		// 避免請求過快
		time.Sleep(500 * time.Millisecond)
	}
	
	// 依評分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// 轉換為 DailyPick，取前3名
	var picks []DailyPick
	limit := 3
	if len(results) < limit {
		limit = len(results)
	}
	
	for i := 0; i < limit; i++ {
		picks = append(picks, stockInfoToDailyPick(results[i], i+1))
	}
	
	fmt.Printf("  ✅ 找到 %d 支，取前 %d 名\n\n", len(results), len(picks))
	
	return picks
}

// ========== 主程式 ==========

func main() {
	fmt.Println("🦈 每日選股報告 V3.0 整合版")
	fmt.Println("📊 135支股票池 + 全價格區間 + 6大技術指標 + 100分制評分")
	fmt.Println(strings.Repeat("=", 60))

	// 載入股票池
	pool, err := loadStockPool("stock_pool.json")
	if err != nil {
		fmt.Printf("❌ 載入股票池失敗: %v\n", err)
		os.Exit(1)
	}

	candidates := getAllStocksFromPool(pool)
	fmt.Printf("\n✅ 載入股票池：%d 支（ETF %d + 個股 %d）\n",
		pool.Total, pool.ETF.Count, pool.Stocks.Count)

	// 分析全市場（無價格限制）
	fmt.Println("\n🚀 開始分析全市場股票...\n")
	allPicks := analyzeAllStocks(candidates)
	var bestPick *DailyPick
	if len(allPicks) > 0 {
		sort.Slice(allPicks, func(i, j int) bool {
			return allPicks[i].Score > allPicks[j].Score
		})
		bestPick = &allPicks[0]
	}
	
	// 統計資料
	buySignals := 0
	riskWarnings := 0
	topPicks := 0
	for _, p := range allPicks {
		if p.Signal == "買點" {
			buySignals++
		}
		if p.Signal == "風險" || p.Score < 40 {
			riskWarnings++
		}
		if p.Score >= 70 {
			topPicks++
		}
	}

	// 產生報告
	report := DailyReport{
		Date:     time.Now().Format("2006-01-02"),
		AllPicks: allPicks,
		BestPick: bestPick,
		Summary: Summary{
			TotalStocks:  len(allPicks),
			PoolSize:     pool.Total,
			TopPicks:     topPicks,
			BuySignals:   buySignals,
			RiskWarnings: riskWarnings,
			UpdateTime:   time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// 儲存 JSON
	outputFile := "stock_web/daily_report.json"
	os.MkdirAll("stock_web", 0755)
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON 產生失敗: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		os.Exit(1)
	}

	// 顯示 TOP 10 結果
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🏆 全市場 TOP 10 選股")
	fmt.Println(strings.Repeat("=", 60))

	limit := 10
	if len(allPicks) < limit {
		limit = len(allPicks)
	}
	for _, p := range allPicks[:limit] {
		fmt.Printf("\n  #%d %s (%s) - $%.2f  ⭐%d分  [%s]\n",
			p.Rank, p.Name, p.Symbol, p.Price, p.Score, p.Signal)
		fmt.Printf("      RSI:%.1f | MACD:%s | KD:%s | 均線:%s\n",
			p.RSI, p.MACD, p.KD, p.MATrend)
		fmt.Printf("      優勢: %s\n", p.Advantage)
	}

	if bestPick != nil {
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("🏆 今日最佳推薦: %s (%s)\n", bestPick.Name, bestPick.Symbol)
		fmt.Printf("    評分: ⭐%d分 | 現價: $%.2f | 訊號: [%s]\n",
			bestPick.Score, bestPick.Price, bestPick.Signal)
		fmt.Printf("    RSI: %.1f | MACD: %s | KD: %s\n",
			bestPick.RSI, bestPick.MACD, bestPick.KD)
		fmt.Printf("    均線趨勢: %s\n", bestPick.MATrend)
		fmt.Printf("    優勢: %s\n", bestPick.Advantage)
		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Printf("\n📊 本次選股統計:\n")
	fmt.Printf("  股票池: %d 支（全價格區間）\n", report.Summary.PoolSize)
	fmt.Printf("  成功分析: %d 支\n", report.Summary.TotalStocks)
	fmt.Printf("  評分70+: %d 支\n", report.Summary.TopPicks)
	fmt.Printf("  買點訊號: %d 支\n", report.Summary.BuySignals)
	fmt.Printf("  風險警示: %d 支\n", report.Summary.RiskWarnings)

	if riskWarnings > 0 {
		fmt.Println("\n⚠️  風險提示:")
		for _, p := range allPicks {
			if p.Signal == "風險" || p.Score < 40 {
				reason := "技術面弱勢"
				if p.RSI > 70 {
					reason = "RSI超買"
				} else if p.MATrend == "空頭" {
					reason = "空頭排列"
				}
				fmt.Printf("  ⚠️  %s (%s) - 評分:%d [%s]\n", p.Name, p.Symbol, p.Score, reason)
			}
		}
	}

	fmt.Printf("\n✅ 報告已儲存至: %s\n", outputFile)
	fmt.Println("🦈 選股完成！")
}

