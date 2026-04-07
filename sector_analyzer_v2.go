package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
}

// 產業熱度報告
type SectorHeatmap struct {
	Date        string        `json:"date"`
	UpdateTime  string        `json:"update_time"`
	HotSectors  []SectorIndex `json:"hot_sectors"`
	ColdSectors []SectorIndex `json:"cold_sectors"`
	AllSectors  []SectorIndex `json:"all_sectors"`
}

// 股票資料（從選股結果讀取）
type StockData struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Change     float64 `json:"change"`
	ChangeRate float64 `json:"change_rate"`
	Volume     float64 `json:"volume"`
	Category   string  `json:"category"`
}

// 產業分類映射
var sectorMapping = map[string]string{
	"金融": "金融保險",
	"電子": "電子",
	"傳產": "傳統產業",
	"ETF": "ETF",
}

func main() {
	fmt.Println("🦈 鯊魚寶寶產業熱度分析 V2")
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
	
	// 從 Yahoo Finance 取得股價（使用現有的 TWSE API 邏輯）
	fmt.Println("🔍 抓取即時股價...")
	stockDataMap := fetchStockPrices(stockPool.Stocks)
	fmt.Printf("✅ 成功取得 %d 支股票資料\n", len(stockDataMap))
	fmt.Println("")
	
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
		fmt.Printf("%d. %s %+.2f%% (%d支股票)\n", 
			i+1, sector.Name, sector.ChangeRate, sector.StockCount)
		if len(sector.LeadStocks) > 0 {
			fmt.Printf("   領漲: %s\n", sector.LeadStocks[0])
		}
	}
	fmt.Println("")
	
	fmt.Println("❄️  冷門產業（跌幅前5）")
	fmt.Println("----------------------------------------")
	for i, sector := range coldSectors {
		fmt.Printf("%d. %s %+.2f%% (%d支股票)\n", 
			i+1, sector.Name, sector.ChangeRate, sector.StockCount)
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

// StockPool 結構（對應 stock_pool.json 格式）
type StockPoolJSON struct {
	Version    string `json:"version"`
	LastUpdate string `json:"last_update"`
	Total      int    `json:"total"`
	ETF        struct {
		Count int               `json:"count"`
		List  map[string]string `json:"list"`
	} `json:"etf"`
	Stocks struct {
		Count      int `json:"count"`
		Categories struct {
			Finance struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"finance"`
			Leading struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"leading"`
			Electronic struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"electronic"`
			Traditional struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"traditional"`
			MidSmall struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"mid_small"`
			AI struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"ai"`
			Power struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"power"`
			Telecom struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"telecom"`
			Other struct {
				Name  string            `json:"name"`
				Count int               `json:"count"`
				List  map[string]string `json:"list"`
			} `json:"other"`
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

// 讀取股票池
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
	
	// 轉換為 StockPool
	pool := &StockPool{
		Stocks: make([]Stock, 0),
	}
	
	// ETF
	for code, name := range poolJSON.ETF.List {
		pool.Stocks = append(pool.Stocks, Stock{
			Code:     code,
			Name:     name,
			Category: "ETF",
		})
	}
	
	// 個股
	categories := []struct {
		list map[string]string
		name string
	}{
		{poolJSON.Stocks.Categories.Finance.List, "金融"},
		{poolJSON.Stocks.Categories.Leading.List, "權值股"},
		{poolJSON.Stocks.Categories.Electronic.List, "電子"},
		{poolJSON.Stocks.Categories.Traditional.List, "傳產"},
		{poolJSON.Stocks.Categories.MidSmall.List, "中小型"},
		{poolJSON.Stocks.Categories.AI.List, "AI相關"},
		{poolJSON.Stocks.Categories.Power.List, "電力"},
		{poolJSON.Stocks.Categories.Telecom.List, "通訊"},
		{poolJSON.Stocks.Categories.Other.List, "其他"},
	}
	
	for _, cat := range categories {
		for code, name := range cat.list {
			pool.Stocks = append(pool.Stocks, Stock{
				Code:     code,
				Name:     name,
				Category: cat.name,
			})
		}
	}
	
	return pool, nil
}

// 快取資料結構
type CacheData struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	LastUpdate  string `json:"last_update"`
	CurrentPrice float64 `json:"current_price"`
	HistoryData []struct {
		Date   string  `json:"date"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	} `json:"history_data"`
}

// 抓取股價（使用 TWSE API cache）
func fetchStockPrices(stocks []Stock) map[string]*StockData {
	result := make(map[string]*StockData)
	
	// 從快取讀取（使用早晨建立的快取）
	cachePath := ".cache/stock_data/%s.json"
	
	for _, stock := range stocks {
		cacheFile := fmt.Sprintf(cachePath, stock.Code)
		data, err := os.ReadFile(cacheFile)
		if err != nil {
			continue
		}
		
		var cacheData CacheData
		err = json.Unmarshal(data, &cacheData)
		if err != nil {
			continue
		}
		
		// 計算漲跌幅（使用最新兩天資料）
		if len(cacheData.HistoryData) < 2 {
			continue
		}
		
		latestData := cacheData.HistoryData[len(cacheData.HistoryData)-1]
		previousData := cacheData.HistoryData[len(cacheData.HistoryData)-2]
		
		change := latestData.Close - previousData.Close
		changeRate := (change / previousData.Close) * 100
		
		stockData := &StockData{
			Code:       stock.Code,
			Name:       stock.Name,
			Price:      latestData.Close,
			Change:     change,
			ChangeRate: changeRate,
			Volume:     latestData.Volume,
			Category:   stock.Category,
		}
		
		result[stock.Code] = stockData
	}
	
	return result
}

// 計算產業表現
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
				ChangeRate: 0,
				Volume:     0,
				LeadStocks: []string{},
				StockCount: 0,
			}
			sectorMap[category] = sector
		}
		
		// 累計漲跌幅
		sector.ChangeRate += data.ChangeRate
		sector.Volume += data.Volume
		sector.StockCount++
		
		// 記錄領漲股（取漲幅最大的）
		if len(sector.LeadStocks) == 0 {
			sector.LeadStocks = append(sector.LeadStocks, 
				fmt.Sprintf("%s (%s) %+.2f%%", data.Name, data.Code, data.ChangeRate))
		} else {
			// 簡單比較（可優化）
			sector.LeadStocks[0] = fmt.Sprintf("%s (%s) %+.2f%%", 
				data.Name, data.Code, data.ChangeRate)
		}
	}
	
	// 計算平均漲跌幅
	sectors := make([]SectorIndex, 0)
	for _, sector := range sectorMap {
		if sector.StockCount > 0 {
			sector.ChangeRate = sector.ChangeRate / float64(sector.StockCount)
		}
		sectors = append(sectors, *sector)
	}
	
	return sectors
}
