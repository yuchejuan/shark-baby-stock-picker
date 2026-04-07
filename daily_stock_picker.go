// daily_stock_picker.go - 每日選股程式
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// 報告結構
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
	Date      string      `json:"date"`
	Picks2030 []DailyPick `json:"picks_20_30"`
	Picks3040 []DailyPick `json:"picks_30_40"`
	Picks4050 []DailyPick `json:"picks_40_50"`
	BestPick  *DailyPick  `json:"best_pick"`
	Summary   Summary     `json:"summary"`
}

type Summary struct {
	TotalStocks int    `json:"total_stocks"`
	TopPicks    int    `json:"top_picks"`
	UpCount     int    `json:"up_count"`
	DownCount   int    `json:"down_count"`
	UpdateTime  string `json:"update_time"`
}

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

func analyzeStocksInRange(stocks map[string]string, minPrice, maxPrice float64) []DailyPick {
	var results []*StockInfo
	processed := 0
	total := len(stocks)
	
	fmt.Printf("處理 %.0f-%.0f 元區間...\n", minPrice, maxPrice)
	
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
			fmt.Printf("  進度: %d/%d\n", processed, total)
		}
		
		// 避免請求過快
		time.Sleep(500 * time.Millisecond)
	}
	
	// 依評分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	// 轉換為 DailyPick
	var picks []DailyPick
	limit := 5
	if len(results) < limit {
		limit = len(results)
	}
	
	for i := 0; i < limit; i++ {
		picks = append(picks, stockInfoToDailyPick(results[i], i+1))
	}
	
	fmt.Printf("  ✅ 找到 %d 支，取前 %d 名\n", len(results), len(picks))
	
	return picks
}

func main() {
	fmt.Println("📊 每日選股報告產生器 V3.0")
	fmt.Println("使用證交所官方API")
	fmt.Println("=====================================")
	
	// 候選股票池
	candidates := map[string]string{
		// 金融股
		"2834": "臺企銀", "2845": "遠東銀", "2849": "安泰銀",
		"5876": "上海商銀", "2836": "高雄銀", "2838": "聯邦銀",
		"2812": "台中銀", "2887": "台新金", "2888": "新光金",
		"2880": "華南金", "2884": "玉山金", "2890": "永豐金",
		"2892": "第一金", "2891": "中信金", "2882": "國泰金",
		"2886": "兆豐金", "2851": "中再保", "2816": "旺旺保",
		
		// 傳產
		"1101": "台泥", "1102": "亞泥", "1216": "統一",
		"1301": "台塑", "1303": "南亞", "1326": "台化",
		"2201": "裕隆", "2618": "長榮航", "2603": "長榮",
		
		// 電子
		"2330": "台積電", "2317": "鴻海", "2454": "聯發科",
		"2308": "台達電", "2353": "宏碁", "2357": "華碩",
		"2382": "廣達", "2324": "仁寶", "3008": "大立光",
	}
	
	// 分析各價格區間
	fmt.Println("\n🔍 開始分析...")
	picks2030 := analyzeStocksInRange(candidates, 20, 30)
	picks3040 := analyzeStocksInRange(candidates, 30, 40)
	picks4050 := analyzeStocksInRange(candidates, 40, 50)
	
	// 找出最佳推薦
	allPicks := append(append(picks2030, picks3040...), picks4050...)
	var bestPick *DailyPick
	if len(allPicks) > 0 {
		sort.Slice(allPicks, func(i, j int) bool {
			return allPicks[i].Score > allPicks[j].Score
		})
		bestPick = &allPicks[0]
	}
	
	// 產生報告
	report := DailyReport{
		Date:      time.Now().Format("2006-01-02"),
		Picks2030: picks2030,
		Picks3040: picks3040,
		Picks4050: picks4050,
		BestPick:  bestPick,
		Summary: Summary{
			TotalStocks: len(allPicks),
			TopPicks:    len(allPicks),
			UpCount:     len(allPicks),
			DownCount:   0,
			UpdateTime:  time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	
	// 儲存 JSON
	outputFile := "stock_web/daily_report.json"
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON 產生失敗: %v\n", err)
		os.Exit(1)
	}
	
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		os.Exit(1)
	}
	
	// 顯示結果
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 選股結果摘要")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Printf("\n💰 20-30元區間 (TOP %d):\n", len(picks2030))
	for _, p := range picks2030 {
		fmt.Printf("  #%d %s (%s) - $%.2f ⭐%d分\n", p.Rank, p.Name, p.Symbol, p.Price, p.Score)
	}
	
	fmt.Printf("\n💰 30-40元區間 (TOP %d):\n", len(picks3040))
	for _, p := range picks3040 {
		fmt.Printf("  #%d %s (%s) - $%.2f ⭐%d分\n", p.Rank, p.Name, p.Symbol, p.Price, p.Score)
	}
	
	fmt.Printf("\n💰 40-50元區間 (TOP %d):\n", len(picks4050))
	for _, p := range picks4050 {
		fmt.Printf("  #%d %s (%s) - $%.2f ⭐%d分\n", p.Rank, p.Name, p.Symbol, p.Price, p.Score)
	}
	
	if bestPick != nil {
		fmt.Printf("\n🏆 今日最佳推薦: %s (%s) - 評分 %d\n", bestPick.Name, bestPick.Symbol, bestPick.Score)
		fmt.Printf("    現價: $%.2f｜RSI: %.1f｜MACD: %s｜訊號: %s\n", 
			bestPick.Price, bestPick.RSI, bestPick.MACD, bestPick.Signal)
		fmt.Printf("    優勢: %s\n", bestPick.Advantage)
	}
	
	fmt.Printf("\n✅ 報告已儲存至: %s\n", outputFile)
	fmt.Println("🦈 選股完成！")
}
