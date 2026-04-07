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

// 產業趨勢分析

type StockData struct {
	Symbol           string
	Name             string
	Industry         string  // 產業分類
	CurrentPrice     float64
	MonthReturn      float64 // 近1月報酬率
	ThreeMonthReturn float64 // 近3月報酬率
	SixMonthReturn   float64 // 近6月報酬率
	RSI              float64
	Score            float64 // 綜合評分
}

type IndustryTrend struct {
	Name          string
	AvgReturn     float64
	StockCount    int
	TopStocks     []string
	Strength      string // 強勢/中性/弱勢
}

func main() {
	fmt.Println("🦈 鯊魚寶寶 - 2026 產業趨勢分析 + 潛力股篩選")
	fmt.Println("========================================")
	fmt.Println()

	// 定義產業股票池
	industries := map[string][]map[string]string{
		"AI與半導體": {
			{"symbol": "2330", "name": "台積電"},
			{"symbol": "2454", "name": "聯發科"},
			{"symbol": "2303", "name": "聯電"},
			{"symbol": "2408", "name": "南亞科"},
			{"symbol": "2344", "name": "華邦電"},
			{"symbol": "3443", "name": "創意"},
			{"symbol": "6669", "name": "緯穎"},
			{"symbol": "2308", "name": "台達電"},
			{"symbol": "3711", "name": "日月光投控"},
			{"symbol": "00891", "name": "中信關鍵半導體"},
		},
		"記憶體": {
			{"symbol": "2408", "name": "南亞科"},
			{"symbol": "2344", "name": "華邦電"},
		},
		"被動元件": {
			{"symbol": "2327", "name": "國巨"},
			{"symbol": "2379", "name": "瑞昱"},
		},
		"面板": {
			{"symbol": "3481", "name": "群創"},
			{"symbol": "2409", "name": "友達"},
		},
		"金融": {
			{"symbol": "2882", "name": "國泰金"},
			{"symbol": "2891", "name": "中信金"},
			{"symbol": "2886", "name": "兆豐金"},
			{"symbol": "2887", "name": "台新金"},
			{"symbol": "2890", "name": "永豐金"},
			{"symbol": "2881", "name": "富邦金"},
			{"symbol": "2892", "name": "第一金"},
			{"symbol": "2884", "name": "玉山金"},
		},
		"電動車與綠能": {
			{"symbol": "00896", "name": "中信綠能及電動車"},
			{"symbol": "00895", "name": "富邦未來車"},
			{"symbol": "1101", "name": "台泥"},
		},
		"傳產與塑化": {
			{"symbol": "1301", "name": "台塑"},
			{"symbol": "1303", "name": "南亞"},
			{"symbol": "1326", "name": "台化"},
			{"symbol": "6505", "name": "台塑化"},
			{"symbol": "2002", "name": "中鋼"},
		},
		"高股息ETF": {
			{"symbol": "0056", "name": "元大高股息"},
			{"symbol": "00878", "name": "國泰永續高股息"},
			{"symbol": "00919", "name": "群益台灣精選高息"},
			{"symbol": "00929", "name": "復華台灣科技優息"},
		},
		"市值型ETF": {
			{"symbol": "0050", "name": "元大台灣50"},
			{"symbol": "006208", "name": "富邦台50"},
			{"symbol": "00692", "name": "富邦公司治理"},
		},
	}

	fmt.Println("📊 產業趨勢強度分析")
	fmt.Println("========================================")
	fmt.Println()

	var allTrends []IndustryTrend

	for industry, stocks := range industries {
		fmt.Printf("🔍 分析產業：%s（%d 支股票）\n", industry, len(stocks))
		
		var returns []float64
		var topStocks []string

		for _, stock := range stocks {
			// 獲取 6 個月報酬率
			ret, err := getSixMonthReturn(stock["symbol"])
			if err != nil {
				fmt.Printf("   ⚠️ %s (%s) 資料不足\n", stock["symbol"], stock["name"])
				continue
			}
			
			returns = append(returns, ret)
			topStocks = append(topStocks, fmt.Sprintf("%s(%.1f%%)", stock["name"], ret))
			fmt.Printf("   ✅ %s (%s): %.2f%%\n", stock["symbol"], stock["name"], ret)
		}

		if len(returns) == 0 {
			fmt.Printf("   ❌ 無可用資料\n\n")
			continue
		}

		// 計算平均報酬率
		avgReturn := 0.0
		for _, r := range returns {
			avgReturn += r
		}
		avgReturn /= float64(len(returns))

		// 判斷產業強度
		strength := "中性"
		if avgReturn > 40 {
			strength = "🔥 超強勢"
		} else if avgReturn > 20 {
			strength = "🚀 強勢"
		} else if avgReturn > 10 {
			strength = "📈 偏強"
		} else if avgReturn > 0 {
			strength = "➡️ 中性"
		} else {
			strength = "📉 弱勢"
		}

		trend := IndustryTrend{
			Name:       industry,
			AvgReturn:  avgReturn,
			StockCount: len(returns),
			TopStocks:  topStocks,
			Strength:   strength,
		}

		allTrends = append(allTrends, trend)

		fmt.Printf("   💡 平均報酬率: %.2f%% (%s)\n\n", avgReturn, strength)

		// 避免請求過快
		time.Sleep(500 * time.Millisecond)
	}

	// 排序：報酬率由高到低
	sort.Slice(allTrends, func(i, j int) bool {
		return allTrends[i].AvgReturn > allTrends[j].AvgReturn
	})

	fmt.Println("========================================")
	fmt.Println("🏆 產業趨勢排行榜")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Printf("%-20s %10s %8s %s\n", "產業", "平均報酬率", "股票數", "強度")
	fmt.Println("----------------------------------------------------------------")

	for i, trend := range allTrends {
		fmt.Printf("#%-2d %-20s %9.2f%% %7d   %s\n",
			i+1, trend.Name, trend.AvgReturn, trend.StockCount, trend.Strength)
	}

	// 儲存 JSON
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"analysis_date": time.Now().Format("2006-01-02"),
		"trends":        allTrends,
	}, "", "  ")

	os.WriteFile("trend_analysis_2026.json", jsonData, 0644)

	fmt.Println()
	fmt.Println("💾 結果已儲存至: trend_analysis_2026.json")
	fmt.Println()
	fmt.Println("🦈 分析完成！")
}

// 獲取 6 個月報酬率
func getSixMonthReturn(symbol string) (float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, -6, 0)

	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s.TW?period1=%d&period2=%d&interval=1d",
		symbol,
		startDate.Unix(),
		endDate.Unix(),
	)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	chart, ok := data["chart"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("無chart資料")
	}

	result, ok := chart["result"].([]interface{})
	if !ok || len(result) == 0 {
		return 0, fmt.Errorf("無result資料")
	}

	firstResult := result[0].(map[string]interface{})
	indicators := firstResult["indicators"].(map[string]interface{})
	quote := indicators["quote"].([]interface{})[0].(map[string]interface{})
	closes := quote["close"].([]interface{})

	if len(closes) < 2 {
		return 0, fmt.Errorf("資料不足")
	}

	// 取第一個和最後一個非 nil 的收盤價
	var startPrice, endPrice float64
	
	for _, c := range closes {
		if c != nil {
			startPrice = c.(float64)
			break
		}
	}

	for i := len(closes) - 1; i >= 0; i-- {
		if closes[i] != nil {
			endPrice = closes[i].(float64)
			break
		}
	}

	if startPrice == 0 {
		return 0, fmt.Errorf("無起始價")
	}

	returnPct := ((endPrice - startPrice) / startPrice) * 100
	return returnPct, nil
}
