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

// 簡化版回測 - 只測試核心訊號組合

type KLine struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type SignalResult struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	SignalDate   string  `json:"signal_date"`
	SignalPrice  float64 `json:"signal_price"`
	NextOpen     float64 `json:"next_open"`
	OpenChange   float64 `json:"open_change"`
	CloseChange  float64 `json:"close_change"`
	Advantage    string  `json:"advantage"`
	Win          bool    `json:"win"`
}

// 測試股票清單（縮小範圍加快速度）
var testStocks = []string{
	// 金融股
	"2812", "2816", "2834", "2836", "2838", "2849", "2850", "2851", "2852",
	"2880", "2881", "2882", "2883", "2884", "2886", "2887", "2892", "5876",
	// 傳產
	"1101", "1102", "1103", "1301", "1326", "2002", "2027", "2105",
	// 電子
	"2303", "2308", "2317", "2330", "2353", "2354", "2412", "2454",
	// 航運
	"2603", "2618",
	// 其他
	"6505",
}

func main() {
	fmt.Println("🦈 鯊魚寶寶訊號回測（簡化版）")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 測試訊號模式：")
	fmt.Println("  1️⃣ RSI偏低 + MACD黃金交叉 + KD超賣")
	fmt.Println("  2️⃣ MACD黃金交叉 + KD超賣")
	fmt.Println("  3️⃣ RSI偏低 + KD超賣")
	fmt.Println("")

	results := make(map[string][]SignalResult)
	signalTypes := []string{
		"RSI偏低+MACD+KD",
		"MACD+KD",
		"RSI+KD",
	}

	for _, st := range signalTypes {
		results[st] = []SignalResult{}
	}

	fmt.Printf("🔍 分析 %d 支股票...\n\n", len(testStocks))

	for i, symbol := range testStocks {
		fmt.Printf("\r  處理: %d/%d (%s)    ", i+1, len(testStocks), symbol)
		
		klines, name := fetchKLines(symbol, 100)
		if len(klines) < 40 {
			continue
		}

		// 分析最近 30 天的訊號
		startDay := len(klines) - 30
		if startDay < 20 {
			startDay = 20
		}

		for day := startDay; day < len(klines)-1; day++ {
			data := klines[:day+1]
			
			// 計算指標
			rsi := calculateRSI(data, 14)
			macd := checkMACD(data)
			kd := checkKD(data)

			// 條件判斷（放寬標準）
			hasRSILow := rsi < 45
			hasMACDBullish := macd
			hasKDOversold := kd

			// 組合訊號
			if hasRSILow && hasMACDBullish && hasKDOversold {
				recordSignal(&results, "RSI偏低+MACD+KD", symbol, name, klines, day, rsi, macd, kd)
			}
			if hasMACDBullish && hasKDOversold {
				recordSignal(&results, "MACD+KD", symbol, name, klines, day, rsi, macd, kd)
			}
			if hasRSILow && hasKDOversold {
				recordSignal(&results, "RSI+KD", symbol, name, klines, day, rsi, macd, kd)
			}
		}
		
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println("\n")

	// 統計結果
	fmt.Println("📊 回測結果")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%-20s %8s %8s %12s\n", "訊號類型", "樣本數", "勝率", "開盤平均漲幅")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, signalType := range signalTypes {
		signalResults := results[signalType]
		if len(signalResults) == 0 {
			fmt.Printf("%-20s %8d %8s %12s\n", signalType, 0, "-", "-")
			continue
		}

		winCount := 0
		totalChange := 0.0
		for _, r := range signalResults {
			if r.Win {
				winCount++
			}
			totalChange += r.OpenChange
		}

		winRate := float64(winCount) / float64(len(signalResults)) * 100
		avgChange := totalChange / float64(len(signalResults))

		fmt.Printf("%-20s %8d %7.1f%% %+11.2f%%\n",
			signalType, len(signalResults), winRate, avgChange)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 顯示「RSI偏低+MACD+KD」的詳細案例
	fmt.Println("\n📈 「RSI偏低+MACD+KD」最近案例：")
	fmt.Println("─────────────────────────────────────────────────────────────")
	
	targetResults := results["RSI偏低+MACD+KD"]
	sort.Slice(targetResults, func(i, j int) bool {
		return targetResults[i].SignalDate > targetResults[j].SignalDate
	})

	showCount := 15
	if len(targetResults) < showCount {
		showCount = len(targetResults)
	}

	for i := 0; i < showCount; i++ {
		r := targetResults[i]
		mark := "❌"
		if r.Win {
			mark = "✅"
		}
		fmt.Printf("%s %s %s %.2f → %.2f (%+.2f%%) | %s\n",
			mark, r.SignalDate, r.Symbol+" "+r.Name,
			r.SignalPrice, r.NextOpen, r.OpenChange, r.Advantage)
	}

	// 儲存結果
	output := map[string]interface{}{
		"date":    time.Now().Format("2006-01-02 15:04:05"),
		"results": results,
	}
	jsonData, _ := json.MarshalIndent(output, "", "  ")
	ioutil.WriteFile("signal_backtest_result.json", jsonData, 0644)

	fmt.Println("\n✅ 完整結果已儲存")
}

func recordSignal(results *map[string][]SignalResult, signalType, symbol, name string, klines []KLine, day int, rsi float64, macd, kd bool) {
	signalPrice := klines[day].Close
	nextOpen := klines[day+1].Open
	nextClose := klines[day+1].Close
	openChange := (nextOpen - signalPrice) / signalPrice * 100
	closeChange := (nextClose - signalPrice) / signalPrice * 100

	advantage := fmt.Sprintf("RSI:%.1f MACD:%v KD:%v", rsi, macd, kd)

	(*results)[signalType] = append((*results)[signalType], SignalResult{
		Symbol:      symbol,
		Name:        name,
		SignalDate:  klines[day].Date,
		SignalPrice: signalPrice,
		NextOpen:    nextOpen,
		OpenChange:  openChange,
		CloseChange: closeChange,
		Advantage:   advantage,
		Win:         openChange > 0,
	})
}

// 簡化的 MACD 判斷（只判斷是否多頭）
func checkMACD(klines []KLine) bool {
	if len(klines) < 26 {
		return false
	}

	ema12 := calculateEMA(klines, 12)
	ema26 := calculateEMA(klines, 26)
	
	return ema12 > ema26
}

// 簡化的 KD 判斷（RSV < 30）
func checkKD(klines []KLine) bool {
	if len(klines) < 9 {
		return false
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
		return false
	}

	currentClose := klines[len(klines)-1].Close
	rsv := (currentClose - lowest) / (highest - lowest) * 100

	return rsv < 30
}

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

func fetchKLines(symbol string, days int) ([]KLine, string) {
	client := &http.Client{Timeout: 10 * time.Second}
	
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

	sort.Slice(allKLines, func(i, j int) bool {
		return allKLines[i].Date < allKLines[j].Date
	})

	return allKLines, name
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
