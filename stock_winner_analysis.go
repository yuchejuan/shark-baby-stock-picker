package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

// Yahoo Finance 歷史資料
type YahooResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol string `json:"symbol"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close []float64 `json:"close"`
					High  []float64 `json:"high"`
					Low   []float64 `json:"low"`
					Open  []float64 `json:"open"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// 股票表現
type StockPerformance struct {
	Symbol        string
	Name          string
	StartPrice    float64
	EndPrice      float64
	HighPrice     float64
	LowPrice      float64
	Return        float64 // 報酬率 (%)
	Volatility    float64 // 波動率
	MaxDrawdown   float64 // 最大回撤
}

func main() {
	fmt.Println("🦈 鯊魚寶寶 - 半年飆股分析系統")
	fmt.Println("分析期間：2025-09-01 到 2026-03-01（6個月）")
	fmt.Println("========================================")
	fmt.Println()

	// 股票池（81支）
	stocks := map[string]string{
		// 市值型 ETF
		"0050": "元大台灣50", "006208": "富邦台50", "00632R": "元大台灣50反1",
		"00692": "富邦公司治理", "00701": "國泰股利精選30", "00881": "國泰台灣5G+",
		"00891": "中信關鍵半導體", "00895": "富邦未來車", "00896": "中信綠能及電動車",
		// 高股息 ETF
		"00919": "群益台灣精選高息", "00929": "復華台灣科技優息", "00918": "大華優利高填息30",
		"00878": "國泰永續高股息", "0056": "元大高股息",
		// 權值股
		"2330": "台積電", "2317": "鴻海", "2454": "聯發科", "2412": "中華電",
		"2882": "國泰金", "2891": "中信金", "2886": "兆豐金", "2881": "富邦金",
		"2892": "第一金", "2884": "玉山金",
		// 電子股
		"2303": "聯電", "2308": "台達電", "2382": "廣達", "2357": "華碩",
		"3711": "日月光投控", "2327": "國巨", "2379": "瑞昱",
		// 傳產股
		"2002": "中鋼", "1301": "台塑", "1303": "南亞", "1326": "台化", "2105": "正新",
		// 中小型股
		"2353": "宏碁", "2324": "仁寶", "2618": "長榮航", "2838": "聯邦銀",
		"2812": "台中銀", "2887": "台新金", "2851": "中再保", "2890": "永豐金",
		"1102": "亞泥", "5876": "上海商銀", "2816": "旺旺保",
		// AI 相關
		"3443": "創意", "6510": "精測", "2395": "研華", "2356": "英業達", "6669": "緯穎",
		// 電力相關
		"1101": "台泥", "6506": "雙鴻", "6411": "晶焱",
		// 通訊相關
		"3045": "台灣大", "4904": "遠傳", "2049": "上銀", "3008": "大立光",
		// 其他重要
		"6505": "台塑化", "2207": "和泰車", "2880": "華南金",
		// 高報酬潛力
		"2409": "友達", "3034": "聯詠", "2301": "光寶科", "2408": "南亞科",
		"2344": "華邦電", "3481": "群創", "6176": "瑞儀", "2371": "大同",
		"6414": "樺漢", "3661": "世芯-KY",
	}

	// 計算起始與結束日期（Unix timestamp）
	endDate := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	startDate := endDate.AddDate(0, -6, 0) // 往前推 6 個月

	var performances []StockPerformance

	fmt.Println("🔍 開始抓取歷史資料...")
	fmt.Println()

	count := 0
	for symbol, name := range stocks {
		count++
		fmt.Printf("[%d/%d] 分析 %s (%s)...", count, len(stocks), symbol, name)

		perf, err := analyzeStock(symbol, name, startDate, endDate)
		if err != nil {
			fmt.Printf(" ❌ 失敗: %v\n", err)
			continue
		}

		performances = append(performances, *perf)
		fmt.Printf(" ✅ 報酬率: %.2f%%\n", perf.Return)

		// 避免請求過快
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("📊 分析完成！")
	fmt.Println("========================================")
	fmt.Println()

	// 排序：報酬率由高到低
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].Return > performances[j].Return
	})

	// 篩選報酬率 > 20% 的股票
	winners := []StockPerformance{}
	for _, p := range performances {
		if p.Return >= 20.0 {
			winners = append(winners, p)
		}
	}

	// 輸出結果
	fmt.Printf("🏆 半年報酬率 > 20%% 的股票（共 %d 支）\n", len(winners))
	fmt.Println("========================================")
	fmt.Println()

	if len(winners) == 0 {
		fmt.Println("⚠️ 沒有股票達到 20% 報酬率門檻")
	} else {
		fmt.Printf("%-8s %-15s %10s %10s %10s %10s %10s\n",
			"代號", "名稱", "起始價", "結束價", "最高價", "報酬率", "最大回撤")
		fmt.Println("--------------------------------------------------------------------------------")

		for _, p := range winners {
			fmt.Printf("%-8s %-15s %10.2f %10.2f %10.2f %9.2f%% %9.2f%%\n",
				p.Symbol, p.Name, p.StartPrice, p.EndPrice, p.HighPrice,
				p.Return, p.MaxDrawdown)
		}
	}

	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("📈 全部排行榜（TOP 20）")
	fmt.Println("========================================")
	fmt.Println()

	topN := 20
	if len(performances) < topN {
		topN = len(performances)
	}

	fmt.Printf("%-8s %-15s %10s %10s %10s\n",
		"排名", "代號", "名稱", "報酬率", "最大回撤")
	fmt.Println("--------------------------------------------------------")

	for i := 0; i < topN; i++ {
		p := performances[i]
		emoji := ""
		if p.Return >= 30 {
			emoji = "🔥"
		} else if p.Return >= 20 {
			emoji = "🚀"
		} else if p.Return >= 10 {
			emoji = "📈"
		} else if p.Return >= 0 {
			emoji = "➡️"
		} else {
			emoji = "📉"
		}

		fmt.Printf("%-8s %-8s %-15s %8.2f%% %9.2f%% %s\n",
			fmt.Sprintf("#%d", i+1), p.Symbol, p.Name,
			p.Return, p.MaxDrawdown, emoji)
	}

	// 儲存 JSON 結果
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"analysis_period": map[string]string{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
		},
		"winners":      winners,
		"all_rankings": performances,
	}, "", "  ")

	err := os.WriteFile("stock_winner_analysis.json", jsonData, 0644)
	if err == nil {
		fmt.Println()
		fmt.Println("💾 結果已儲存至: stock_winner_analysis.json")
	}

	fmt.Println()
	fmt.Println("🦈 分析完成！")
}

// 分析單支股票
func analyzeStock(symbol, name string, startDate, endDate time.Time) (*StockPerformance, error) {
	// 構建 Yahoo Finance API URL
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s.TW?period1=%d&period2=%d&interval=1d",
		symbol,
		startDate.Unix(),
		endDate.Unix(),
	)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data YahooResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("無資料")
	}

	result := data.Chart.Result[0]
	closes := result.Indicators.Quote[0].Close

	if len(closes) < 2 {
		return nil, fmt.Errorf("資料不足")
	}

	// 計算指標
	startPrice := closes[0]
	endPrice := closes[len(closes)-1]
	returnPct := ((endPrice - startPrice) / startPrice) * 100

	// 計算最高價、最大回撤
	highPrice := startPrice
	maxDrawdown := 0.0
	peak := startPrice

	for _, price := range closes {
		if price > highPrice {
			highPrice = price
		}
		if price > peak {
			peak = price
		}
		drawdown := ((peak - price) / peak) * 100
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return &StockPerformance{
		Symbol:      symbol,
		Name:        name,
		StartPrice:  startPrice,
		EndPrice:    endPrice,
		HighPrice:   highPrice,
		Return:      returnPct,
		MaxDrawdown: maxDrawdown,
	}, nil
}
