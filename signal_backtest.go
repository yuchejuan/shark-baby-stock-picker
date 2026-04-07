package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

// K線資料
type KLine struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

// 訊號測試結果
type SignalResult struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	SignalDate   string  `json:"signal_date"`
	SignalPrice  float64 `json:"signal_price"`
	NextOpen     float64 `json:"next_open"`
	NextClose    float64 `json:"next_close"`
	NextHigh     float64 `json:"next_high"`
	OpenChange   float64 `json:"open_change"`   // 隔日開盤漲跌幅
	CloseChange  float64 `json:"close_change"`  // 隔日收盤漲跌幅
	HighChange   float64 `json:"high_change"`   // 隔日最高漲幅
	SignalType   string  `json:"signal_type"`
	Win          bool    `json:"win"`           // 隔日開盤上漲
}

// 回測統計
type BacktestStats struct {
	SignalType     string  `json:"signal_type"`
	TotalSignals   int     `json:"total_signals"`
	WinCount       int     `json:"win_count"`
	WinRate        float64 `json:"win_rate"`
	AvgOpenChange  float64 `json:"avg_open_change"`
	AvgCloseChange float64 `json:"avg_close_change"`
	AvgHighChange  float64 `json:"avg_high_change"`
	MaxGain        float64 `json:"max_gain"`
	MaxLoss        float64 `json:"max_loss"`
}

// 股票清單（使用與 daily picker 相同的）
var testStocks = []string{
	// 金融股
	"2812", "2816", "2834", "2836", "2838", "2845", "2849", "2850", "2851", "2852",
	"2867", "2880", "2881", "2882", "2883", "2884", "2885", "2886", "2887", "2888",
	"2889", "2890", "2891", "2892", "5876", "5880",
	// 傳產股
	"1101", "1102", "1103", "1216", "1301", "1303", "1326", "1402", "1476",
	"2002", "2006", "2027", "2101", "2105", "2201", "2207", "2301", "2303",
	// 電子股
	"2308", "2317", "2324", "2327", "2330", "2345", "2353", "2354", "2356",
	"2357", "2376", "2377", "2379", "2382", "2395", "2408", "2412", "2454",
	// 航運股
	"2603", "2609", "2610", "2615", "2618",
	// 鋼鐵股
	"2006", "2014", "2027",
	// 其他
	"6505", "9910", "9914", "9917",
}

func main() {
	fmt.Println("🦈 鯊魚寶寶訊號回測系統")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 驗證訊號：RSI偏低 + MACD黃金交叉 + KD超賣")
	fmt.Println("📅 回測期間：過去 90 個交易日")
	fmt.Println()

	// 定義要測試的訊號類型
	signalTypes := []string{
		"RSI偏低 + MACD黃金交叉 + KD超賣",
		"MACD黃金交叉 + KD超賣",
		"MACD黃金交叉 + 多頭排列",
		"RSI超賣",
		"KD超賣",
		"MACD黃金交叉",
	}

	results := make(map[string][]SignalResult)
	for _, signalType := range signalTypes {
		results[signalType] = []SignalResult{}
	}

	// 去重
	stockMap := make(map[string]bool)
	var uniqueStocks []string
	for _, s := range testStocks {
		if !stockMap[s] {
			stockMap[s] = true
			uniqueStocks = append(uniqueStocks, s)
		}
	}

	fmt.Printf("🔍 分析 %d 支股票...\n\n", len(uniqueStocks))

	for i, symbol := range uniqueStocks {
		fmt.Printf("\r  處理中: %d/%d (%s)    ", i+1, len(uniqueStocks), symbol)
		
		klines, name := fetchKLines(symbol, 120)
		if len(klines) < 60 {
			continue
		}

		// 分析每一天的訊號
		for day := 60; day < len(klines)-1; day++ {
			// 計算技術指標
			rsi := calculateRSI(klines[:day+1], 14)
			macdSignal := calculateMACD(klines[:day+1])
			kdSignal := calculateKD(klines[:day+1])
			maTrend := calculateMATrend(klines[:day+1])

			// 組合訊號判斷
			advantage := buildAdvantage(rsi, macdSignal, kdSignal, maTrend)

			// 隔日數據
			signalPrice := klines[day].Close
			nextOpen := klines[day+1].Open
			nextClose := klines[day+1].Close
			nextHigh := klines[day+1].High

			openChange := (nextOpen - signalPrice) / signalPrice * 100
			closeChange := (nextClose - signalPrice) / signalPrice * 100
			highChange := (nextHigh - signalPrice) / signalPrice * 100

			// 檢查各種訊號類型
			for _, signalType := range signalTypes {
				if matchesSignal(advantage, signalType) {
					result := SignalResult{
						Symbol:      symbol,
						Name:        name,
						SignalDate:  klines[day].Date,
						SignalPrice: signalPrice,
						NextOpen:    nextOpen,
						NextClose:   nextClose,
						NextHigh:    nextHigh,
						OpenChange:  openChange,
						CloseChange: closeChange,
						HighChange:  highChange,
						SignalType:  signalType,
						Win:         openChange > 0,
					}
					results[signalType] = append(results[signalType], result)
				}
			}
		}
		
		time.Sleep(300 * time.Millisecond) // 避免 API 限制
	}

	fmt.Println("\n")

	// 計算統計
	fmt.Println("📊 回測結果統計")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%-40s %6s %6s %8s %8s %8s\n", "訊號類型", "樣本數", "勝率", "開盤均漲", "收盤均漲", "最高均漲")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var allStats []BacktestStats
	for _, signalType := range signalTypes {
		signalResults := results[signalType]
		if len(signalResults) == 0 {
			continue
		}

		stats := calculateStats(signalType, signalResults)
		allStats = append(allStats, stats)

		fmt.Printf("%-40s %6d %5.1f%% %+7.2f%% %+7.2f%% %+7.2f%%\n",
			signalType,
			stats.TotalSignals,
			stats.WinRate,
			stats.AvgOpenChange,
			stats.AvgCloseChange,
			stats.AvgHighChange,
		)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 輸出詳細結果
	fmt.Println("\n📈 「RSI偏低 + MACD黃金交叉 + KD超賣」詳細記錄（最近 20 筆）：")
	fmt.Println("─────────────────────────────────────────────────────────────────")
	
	targetResults := results["RSI偏低 + MACD黃金交叉 + KD超賣"]
	sort.Slice(targetResults, func(i, j int) bool {
		return targetResults[i].SignalDate > targetResults[j].SignalDate
	})

	showCount := 20
	if len(targetResults) < showCount {
		showCount = len(targetResults)
	}

	for i := 0; i < showCount; i++ {
		r := targetResults[i]
		winMark := "❌"
		if r.Win {
			winMark = "✅"
		}
		fmt.Printf("%s %s %s | 訊號價:%.2f → 開盤:%.2f (%+.2f%%) | 收盤:%+.2f%% | 最高:%+.2f%%\n",
			winMark, r.SignalDate, r.Symbol+" "+r.Name,
			r.SignalPrice, r.NextOpen, r.OpenChange,
			r.CloseChange, r.HighChange,
		)
	}

	// 儲存結果
	output := map[string]interface{}{
		"backtest_date": time.Now().Format("2006-01-02 15:04:05"),
		"period":        "過去90個交易日",
		"stocks_count":  len(uniqueStocks),
		"statistics":    allStats,
		"target_signal": "RSI偏低 + MACD黃金交叉 + KD超賣",
		"target_results": targetResults,
	}

	jsonData, _ := json.MarshalIndent(output, "", "  ")
	ioutil.WriteFile("signal_backtest_result.json", jsonData, 0644)

	fmt.Println("\n✅ 詳細結果已儲存至 signal_backtest_result.json")
}

// 判斷是否符合訊號
func matchesSignal(advantage, signalType string) bool {
	switch signalType {
	case "RSI偏低 + MACD黃金交叉 + KD超賣":
		return strings.Contains(advantage, "RSI偏低") &&
			strings.Contains(advantage, "MACD黃金交叉") &&
			strings.Contains(advantage, "KD超賣")
	case "MACD黃金交叉 + KD超賣":
		return strings.Contains(advantage, "MACD黃金交叉") &&
			strings.Contains(advantage, "KD超賣")
	case "MACD黃金交叉 + 多頭排列":
		return strings.Contains(advantage, "MACD黃金交叉") &&
			strings.Contains(advantage, "多頭排列")
	case "RSI超賣":
		return strings.Contains(advantage, "RSI超賣")
	case "KD超賣":
		return strings.Contains(advantage, "KD超賣")
	case "MACD黃金交叉":
		return strings.Contains(advantage, "MACD黃金交叉")
	}
	return false
}

// 組合優勢說明
func buildAdvantage(rsi float64, macd, kd, maTrend string) string {
	var parts []string

	if rsi < 30 {
		parts = append(parts, fmt.Sprintf("RSI超賣%.1f", rsi))
	} else if rsi < 45 {
		parts = append(parts, "RSI偏低")
	}

	if macd == "多頭" || macd == "多頭趨勢" {
		parts = append(parts, "MACD黃金交叉")
	}

	if kd == "超賣" {
		parts = append(parts, "KD超賣")
	}

	if maTrend == "多頭排列" {
		parts = append(parts, "多頭排列")
	}

	return strings.Join(parts, " + ")
}

// 計算統計
func calculateStats(signalType string, results []SignalResult) BacktestStats {
	if len(results) == 0 {
		return BacktestStats{SignalType: signalType}
	}

	var winCount int
	var sumOpen, sumClose, sumHigh float64
	maxGain := -999.0
	maxLoss := 999.0

	for _, r := range results {
		if r.Win {
			winCount++
		}
		sumOpen += r.OpenChange
		sumClose += r.CloseChange
		sumHigh += r.HighChange

		if r.OpenChange > maxGain {
			maxGain = r.OpenChange
		}
		if r.OpenChange < maxLoss {
			maxLoss = r.OpenChange
		}
	}

	n := float64(len(results))
	return BacktestStats{
		SignalType:     signalType,
		TotalSignals:   len(results),
		WinCount:       winCount,
		WinRate:        float64(winCount) / n * 100,
		AvgOpenChange:  sumOpen / n,
		AvgCloseChange: sumClose / n,
		AvgHighChange:  sumHigh / n,
		MaxGain:        maxGain,
		MaxLoss:        maxLoss,
	}
}

// 取得 K 線資料
func fetchKLines(symbol string, days int) ([]KLine, string) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 計算需要的月份數
	now := time.Now()
	var allKLines []KLine
	name := ""

	for i := 0; i < 4; i++ {
		targetDate := now.AddDate(0, -i, 0)
		url := fmt.Sprintf("https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=%s&stockNo=%s",
			targetDate.Format("20060102"), symbol)

		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}

		if result["data"] == nil {
			continue
		}

		if name == "" {
			if title, ok := result["title"].(string); ok {
				parts := strings.Split(title, " ")
				if len(parts) >= 3 {
					name = parts[2]
				}
			}
		}

		data := result["data"].([]interface{})
		for _, row := range data {
			r := row.([]interface{})
			if len(r) < 7 {
				continue
			}

			kline := KLine{
				Date:   fmt.Sprintf("%v", r[0]),
				Open:   parseFloat(fmt.Sprintf("%v", r[3])),
				High:   parseFloat(fmt.Sprintf("%v", r[4])),
				Low:    parseFloat(fmt.Sprintf("%v", r[5])),
				Close:  parseFloat(fmt.Sprintf("%v", r[6])),
				Volume: parseInt(fmt.Sprintf("%v", r[1])),
			}
			
			if kline.Close > 0 {
				allKLines = append(allKLines, kline)
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	// 按日期排序
	sort.Slice(allKLines, func(i, j int) bool {
		return allKLines[i].Date < allKLines[j].Date
	})

	return allKLines, name
}

// 計算 RSI
func calculateRSI(klines []KLine, period int) float64 {
	if len(klines) < period+1 {
		return 50
	}

	var gains, losses float64
	for i := len(klines) - period; i < len(klines); i++ {
		change := klines[i].Close - klines[i-1].Close
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	if losses == 0 {
		return 100
	}

	rs := gains / losses
	return 100 - (100 / (1 + rs))
}

// 計算 MACD 狀態
func calculateMACD(klines []KLine) string {
	if len(klines) < 26 {
		return "中性"
	}

	ema12 := calculateEMA(klines, 12)
	ema26 := calculateEMA(klines, 26)
	dif := ema12 - ema26

	// 計算前一天的 DIF
	prevEma12 := calculateEMA(klines[:len(klines)-1], 12)
	prevEma26 := calculateEMA(klines[:len(klines)-1], 26)
	prevDif := prevEma12 - prevEma26

	if dif > 0 && prevDif <= 0 {
		return "多頭" // 黃金交叉
	} else if dif < 0 && prevDif >= 0 {
		return "空頭" // 死亡交叉
	} else if dif > 0 {
		return "多頭趨勢"
	}
	return "空頭趨勢"
}

// 計算 KD 狀態
func calculateKD(klines []KLine) string {
	if len(klines) < 9 {
		return "中性"
	}

	period := 9
	recent := klines[len(klines)-period:]
	
	var highest, lowest float64 = recent[0].High, recent[0].Low
	for _, k := range recent {
		if k.High > highest {
			highest = k.High
		}
		if k.Low < lowest {
			lowest = k.Low
		}
	}

	if highest == lowest {
		return "中性"
	}

	currentClose := klines[len(klines)-1].Close
	rsv := (currentClose - lowest) / (highest - lowest) * 100

	if rsv < 30 {
		return "超賣"
	} else if rsv > 70 {
		return "超買"
	}
	return "中性"
}

// 計算均線趨勢
func calculateMATrend(klines []KLine) string {
	if len(klines) < 20 {
		return "中性"
	}

	ma5 := calculateMA(klines, 5)
	ma20 := calculateMA(klines, 20)

	if ma5 > ma20 {
		return "多頭排列"
	} else if ma5 < ma20 {
		return "空頭排列"
	}
	return "中性"
}

// 計算 EMA
func calculateEMA(klines []KLine, period int) float64 {
	if len(klines) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := klines[0].Close

	for i := 1; i < len(klines); i++ {
		ema = (klines[i].Close-ema)*multiplier + ema
	}

	return ema
}

// 計算 MA
func calculateMA(klines []KLine, period int) float64 {
	if len(klines) < period {
		return 0
	}

	var sum float64
	for i := len(klines) - period; i < len(klines); i++ {
		sum += klines[i].Close
	}
	return sum / float64(period)
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

func parseInt(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}
