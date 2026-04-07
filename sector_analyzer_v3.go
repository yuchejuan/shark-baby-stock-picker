package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// 產業指數結構
type SectorIndex struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	ChangeRate  float64  `json:"change_rate"`
	Volume      float64  `json:"volume"`
	LeadStocks  []string `json:"lead_stocks"`
	StockCount  int      `json:"stock_count"`
	UpCount     int      `json:"up_count"`
	DownCount   int      `json:"down_count"`
}

// 產業熱度報告
type SectorHeatmap struct {
	Date        string        `json:"date"`
	UpdateTime  string        `json:"update_time"`
	HotSectors  []SectorIndex `json:"hot_sectors"`
	ColdSectors []SectorIndex `json:"cold_sectors"`
	AllSectors  []SectorIndex `json:"all_sectors"`
}

// Yahoo Finance 回應
type YahooQuote struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol             string  `json:"symbol"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
			Indicators struct {
				Quote []struct {
					Close  []float64 `json:"close"`
					Volume []float64 `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

func main() {
	fmt.Println("🦈 鯊魚寶寶產業熱度分析 V3")
	fmt.Println("========================================")
	
	now := time.Now()
	fmt.Printf("📅 分析日期: %s\n", now.Format("2006-01-02"))
	fmt.Println("")
	
	// 從 stock_pool.json 讀取股票池
	fmt.Println("📊 讀取股票池...")
	stockPool, err := loadStockPool()
	if err != nil {
		fmt.Printf("❌ 讀取失敗: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ 已載入 %d 支股票\n", len(stockPool.Stocks))
	fmt.Println("")
	
	// 從 Yahoo Finance 取得股價
	fmt.Println("🔍 抓取即時股價（Yahoo Finance）...")
	stockDataMap := fetchStockPricesFromYahoo(stockPool.Stocks)
	fmt.Printf("✅ 成功取得 %d 支股票資料\n", len(stockDataMap))
	fmt.Println("")
	
	if len(stockDataMap) == 0 {
		fmt.Println("❌ 無法取得股票資料，無法進行產業分析")
		os.Exit(1)
	}
	
	// 依產業分類統計
	fmt.Println("📈 計算產業漲跌幅...")
	sectors := calculateSectorPerformance(stockPool.Stocks, stockDataMap)
	fmt.Printf("✅ 共 %d 個產業\n", len(sectors))
	fmt.Println("")
	
	// 排序
	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].ChangeRate > sectors[j].ChangeRate
	})
	
	// 取前5和後5
	hotCount := 5
	if len(sectors) < 5 {
		hotCount = len(sectors)
	}
	
	hotSectors := sectors[:hotCount]
	coldSectors := make([]SectorIndex, 0)
	if len(sectors) >= 5 {
		coldSectors = sectors[len(sectors)-5:]
		for i, j := 0, len(coldSectors)-1; i < j; i, j = i+1, j-1 {
			coldSectors[i], coldSectors[j] = coldSectors[j], coldSectors[i]
		}
	}
	
	// 顯示結果
	fmt.Println("🔥 熱門產業（漲幅前5）")
	fmt.Println("----------------------------------------")
	for i, sector := range hotSectors {
		fmt.Printf("%d. %s %+.2f%% (%d支股票, %d漲%d跌)\n", 
			i+1, sector.Name, sector.ChangeRate, sector.StockCount,
			sector.UpCount, sector.DownCount)
		if len(sector.LeadStocks) > 0 {
			fmt.Printf("   領漲: %s\n", sector.LeadStocks[0])
		}
	}
	fmt.Println("")
	
	fmt.Println("❄️  冷門產業（跌幅前5）")
	fmt.Println("----------------------------------------")
	for i, sector := range coldSectors {
		fmt.Printf("%d. %s %+.2f%% (%d支股票, %d漲%d跌)\n", 
			i+1, sector.Name, sector.ChangeRate, sector.StockCount,
			sector.UpCount, sector.DownCount)
		if len(sector.LeadStocks) > 0 {
			fmt.Printf("   領跌: %s\n", sector.LeadStocks[0])
		}
	}
	fmt.Println("")
	
	// 建立報告
	report := SectorHeatmap{
		Date:        now.Format("2006-01-02"),
		UpdateTime:  now.Format("2006-01-02 15:04:05"),
		HotSectors:  hotSectors,
		ColdSectors: coldSectors,
		AllSectors:  sectors,
	}
	
	// 儲存 JSON
	outputPath := "stock_web/sector_heatmap.json"
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON 編碼失敗: %v\n", err)
		os.Exit(1)
	}
	
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ 產業熱度報告已儲存: %s\n", outputPath)
	fmt.Println("========================================")
	fmt.Println("🦈 分析完成")
}

// StockPool 結構
type StockPoolJSON struct {
	ETF struct {
		List map[string]string `json:"list"`
	} `json:"etf"`
	Stocks struct {
		Categories struct {
			Finance       struct{ List map[string]string `json:"list"` } `json:"finance"`
			BlueChip      struct{ List map[string]string `json:"list"` } `json:"blue_chip"`
			Electronics   struct{ List map[string]string `json:"list"` } `json:"electronics"`
			Traditional   struct{ List map[string]string `json:"list"` } `json:"traditional"`
			MidSmallCap   struct{ List map[string]string `json:"list"` } `json:"mid_small_cap"`
			AITech        struct{ List map[string]string `json:"list"` } `json:"ai_tech"`
			PowerUtility  struct{ List map[string]string `json:"list"` } `json:"power_utility"`
			Telecom       struct{ List map[string]string `json:"list"` } `json:"telecom"`
			Others        struct{ List map[string]string `json:"list"` } `json:"others"`
		} `json:"categories"`
	} `json:"stocks"`
}

type StockPool struct {
	Stocks []Stock
}

type Stock struct {
	Code     string
	Name     string
	Category string
}

type StockData struct {
	Code       string
	Name       string
	Price      float64
	ChangeRate float64
	Volume     float64
	Category   string
}

func loadStockPool() (*StockPool, error) {
	data, err := os.ReadFile("stock_pool.json")
	if err != nil {
		return nil, err
	}
	
	var poolJSON StockPoolJSON
	err = json.Unmarshal(data, &poolJSON)
	if err != nil {
		return nil, err
	}
	
	pool := &StockPool{Stocks: make([]Stock, 0)}
	
	// ETF
	for code, name := range poolJSON.ETF.List {
		pool.Stocks = append(pool.Stocks, Stock{code, name, "ETF"})
	}
	
	// 個股
	categories := []struct {
		list map[string]string
		name string
	}{
		{poolJSON.Stocks.Categories.Finance.List, "金融"},
		{poolJSON.Stocks.Categories.BlueChip.List, "權值股"},
		{poolJSON.Stocks.Categories.Electronics.List, "電子"},
		{poolJSON.Stocks.Categories.Traditional.List, "傳產"},
		{poolJSON.Stocks.Categories.MidSmallCap.List, "中小型"},
		{poolJSON.Stocks.Categories.AITech.List, "AI相關"},
		{poolJSON.Stocks.Categories.PowerUtility.List, "電力"},
		{poolJSON.Stocks.Categories.Telecom.List, "通訊"},
		{poolJSON.Stocks.Categories.Others.List, "其他"},
	}
	
	for _, cat := range categories {
		for code, name := range cat.list {
			pool.Stocks = append(pool.Stocks, Stock{code, name, cat.name})
		}
	}
	
	return pool, nil
}

func fetchStockPricesFromYahoo(stocks []Stock) map[string]*StockData {
	result := make(map[string]*StockData)
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	for i, stock := range stocks {
		if i > 0 && i%10 == 0 {
			fmt.Printf("  ⏳ 進度: %d/%d\n", i, len(stocks))
			time.Sleep(1 * time.Second) // 避免 API 限制
		}
		
		// 台股代碼需加 .TW
		symbol := stock.Code + ".TW"
		url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=5d&interval=1d", symbol)
		
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		
		var quote YahooQuote
		err = json.Unmarshal(body, &quote)
		if err != nil || len(quote.Chart.Result) == 0 {
			continue
		}
		
		chartResult := quote.Chart.Result[0]
		quotes := chartResult.Indicators.Quote
		if len(quotes) == 0 || len(quotes[0].Close) < 2 {
			continue
		}
		
		closes := quotes[0].Close
		volumes := quotes[0].Volume
		
		// 取最新兩天計算漲跌幅
		var latestClose, prevClose, totalVolume float64
		count := 0
		for i := len(closes) - 1; i >= 0 && count < 2; i-- {
			if closes[i] > 0 {
				if count == 0 {
					latestClose = closes[i]
					if i < len(volumes) {
						totalVolume = volumes[i]
					}
				} else {
					prevClose = closes[i]
				}
				count++
			}
		}
		
		if prevClose == 0 {
			continue
		}
		
		changeRate := ((latestClose - prevClose) / prevClose) * 100
		
		result[stock.Code] = &StockData{
			Code:       stock.Code,
			Name:       stock.Name,
			Price:      latestClose,
			ChangeRate: changeRate,
			Volume:     totalVolume,
			Category:   stock.Category,
		}
	}
	
	return result
}

func calculateSectorPerformance(stocks []Stock, dataMap map[string]*StockData) []SectorIndex {
	sectorMap := make(map[string]*SectorIndex)
	
	for _, stock := range stocks {
		data, exists := dataMap[stock.Code]
		if !exists {
			continue
		}
		
		category := stock.Category
		if category == "" {
			category = "其他"
		}
		
		sector, exists := sectorMap[category]
		if !exists {
			sector = &SectorIndex{
				Code:       category,
				Name:       category,
				LeadStocks: []string{},
			}
			sectorMap[category] = sector
		}
		
		sector.ChangeRate += data.ChangeRate
		sector.Volume += data.Volume
		sector.StockCount++
		
		if data.ChangeRate > 0 {
			sector.UpCount++
		} else if data.ChangeRate < 0 {
			sector.DownCount++
		}
		
		// 記錄領漲/跌股
		stockInfo := fmt.Sprintf("%s (%s) %+.2f%%", data.Name, data.Code, data.ChangeRate)
		if len(sector.LeadStocks) == 0 {
			sector.LeadStocks = append(sector.LeadStocks, stockInfo)
		} else {
			// 更新為漲跌幅最大的
			var prevRate float64
			fmt.Sscanf(strings.Split(sector.LeadStocks[0], ")")[1], "%f", &prevRate)
			if (sector.ChangeRate > 0 && data.ChangeRate > prevRate) ||
			   (sector.ChangeRate < 0 && data.ChangeRate < prevRate) {
				sector.LeadStocks[0] = stockInfo
			}
		}
	}
	
	sectors := make([]SectorIndex, 0)
	for _, sector := range sectorMap {
		if sector.StockCount > 0 {
			sector.ChangeRate = sector.ChangeRate / float64(sector.StockCount)
		}
		sectors = append(sectors, *sector)
	}
	
	return sectors
}
