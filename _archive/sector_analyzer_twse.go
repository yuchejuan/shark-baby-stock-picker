package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
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
	DataSource  string        `json:"data_source"`
}

// TWSE 批次股票資料結構
type TWSEStockData struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

// TWSE 個股日資料結構
type TWSEDayData struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
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

func main() {
	fmt.Println("🦈 鯊魚寶寶產業熱度分析 - TWSE 測試版")
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
	
	// 方法 1: 嘗試批次取得收盤價
	fmt.Println("🔍 方法 1: 批次取得收盤價（TWSE BWIBBU_d API）...")
	stockDataMap1, err := fetchStockPricesBatch(now)
	if err != nil {
		fmt.Printf("⚠️  批次 API 失敗: %v\n", err)
		stockDataMap1 = make(map[string]*StockData)
	} else {
		fmt.Printf("✅ 批次成功取得 %d 支股票資料\n", len(stockDataMap1))
	}
	fmt.Println("")
	
	// 方法 2: 對缺少的股票個別取得（限制最多 10 支測試）
	missingStocks := []Stock{}
	for _, stock := range stockPool.Stocks {
		if _, exists := stockDataMap1[stock.Code]; !exists {
			missingStocks = append(missingStocks, stock)
		}
	}
	
	if len(missingStocks) > 0 {
		testCount := 10
		if len(missingStocks) < 10 {
			testCount = len(missingStocks)
		}
		
		fmt.Printf("🔍 方法 2: 個別取得缺少的股票（測試前 %d 支）...\n", testCount)
		for i := 0; i < testCount; i++ {
			stock := missingStocks[i]
			fmt.Printf("  ⏳ [%d/%d] 取得 %s (%s)...\n", i+1, testCount, stock.Name, stock.Code)
			
			data, err := fetchSingleStock(stock, now)
			if err != nil {
				fmt.Printf("     ❌ 失敗: %v\n", err)
				continue
			}
			
			stockDataMap1[stock.Code] = data
			fmt.Printf("     ✅ 成功: 價格 %.2f, 漲跌 %+.2f%%\n", data.Price, data.ChangeRate)
			
			time.Sleep(2 * time.Second) // 避免 rate limit
		}
		fmt.Println("")
	}
	
	// 整合股票池分類資訊
	for _, stock := range stockPool.Stocks {
		if data, exists := stockDataMap1[stock.Code]; exists {
			data.Category = stock.Category
		}
	}
	
	fmt.Printf("✅ 最終成功取得 %d 支股票資料\n", len(stockDataMap1))
	fmt.Println("")
	
	if len(stockDataMap1) == 0 {
		fmt.Println("❌ 無法取得股票資料，無法進行產業分析")
		os.Exit(1)
	}
	
	// 依產業分類統計
	fmt.Println("📈 計算產業漲跌幅...")
	sectors := calculateSectorPerformance(stockPool.Stocks, stockDataMap1)
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
		DataSource:  "TWSE Official API (Test)",
	}
	
	// 儲存 JSON
	outputPath := "stock_web/sector_heatmap_twse_test.json"
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
	
	fmt.Printf("✅ TWSE 測試版報告已儲存: %s\n", outputPath)
	fmt.Println("")
	
	// 統計資訊
	fmt.Println("📊 測試統計")
	fmt.Println("----------------------------------------")
	fmt.Printf("成功取得: %d/%d 支股票 (%.1f%%)\n", 
		len(stockDataMap1), len(stockPool.Stocks), 
		float64(len(stockDataMap1))/float64(len(stockPool.Stocks))*100)
	fmt.Printf("產業數量: %d 個\n", len(sectors))
	fmt.Println("========================================")
	fmt.Println("🦈 TWSE 測試完成")
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

// 方法 1: 批次取得收盤價
func fetchStockPricesBatch(date time.Time) (map[string]*StockData, error) {
	result := make(map[string]*StockData)
	
	dateStr := date.Format("20060102")
	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/BWIBBU_d?date=%s&selectType=ALL&response=json", dateStr)
	
	client := &http.Client{Timeout: 30 * time.Second}
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
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var twseData TWSEStockData
	err = json.Unmarshal(body, &twseData)
	if err != nil {
		return nil, err
	}
	
	if twseData.Stat != "OK" {
		return nil, fmt.Errorf("TWSE API 回應: %s", twseData.Stat)
	}
	
	// 解析欄位索引
	codeIdx := -1
	nameIdx := -1
	priceIdx := -1
	
	for i, field := range twseData.Fields {
		switch field {
		case "證券代號":
			codeIdx = i
		case "證券名稱":
			nameIdx = i
		case "收盤價":
			priceIdx = i
		}
	}
	
	if codeIdx == -1 || nameIdx == -1 || priceIdx == -1 {
		return nil, fmt.Errorf("無法找到必要欄位")
	}
	
	// 解析資料（此 API 僅有收盤價，無漲跌幅）
	for _, row := range twseData.Data {
		if len(row) <= priceIdx {
			continue
		}
		
		code := strings.TrimSpace(row[codeIdx])
		name := strings.TrimSpace(row[nameIdx])
		priceStr := strings.ReplaceAll(strings.TrimSpace(row[priceIdx]), ",", "")
		
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil || price == 0 {
			continue
		}
		
		result[code] = &StockData{
			Code:       code,
			Name:       name,
			Price:      price,
			ChangeRate: 0, // 批次 API 無漲跌幅，需從個股 API 取得
			Volume:     0,
			Category:   "",
		}
	}
	
	return result, nil
}

// 方法 2: 個別取得股票日資料
func fetchSingleStock(stock Stock, date time.Time) (*StockData, error) {
	dateStr := date.Format("20060102")
	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/STOCK_DAY?date=%s&stockNo=%s&response=json", dateStr, stock.Code)
	
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
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var dayData TWSEDayData
	err = json.Unmarshal(body, &dayData)
	if err != nil {
		return nil, err
	}
	
	if dayData.Stat != "OK" || len(dayData.Data) == 0 {
		return nil, fmt.Errorf("無資料或 API 失敗")
	}
	
	// 取最後一筆（最新日期）
	latestRow := dayData.Data[len(dayData.Data)-1]
	
	// 解析欄位
	priceIdx := -1
	changeIdx := -1
	volumeIdx := -1
	
	for i, field := range dayData.Fields {
		switch field {
		case "收盤價":
			priceIdx = i
		case "漲跌價差":
			changeIdx = i
		case "成交股數":
			volumeIdx = i
		}
	}
	
	if priceIdx == -1 || changeIdx == -1 || len(latestRow) <= priceIdx {
		return nil, fmt.Errorf("欄位解析失敗")
	}
	
	priceStr := strings.ReplaceAll(strings.TrimSpace(latestRow[priceIdx]), ",", "")
	changeStr := strings.ReplaceAll(strings.TrimSpace(latestRow[changeIdx]), ",", "")
	
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price == 0 {
		return nil, fmt.Errorf("價格解析失敗: %v", err)
	}
	
	// 漲跌價差可能有 +/- 符號
	changeStr = strings.TrimPrefix(changeStr, "+")
	changeStr = strings.TrimPrefix(changeStr, "<span style='color:red'>+</span>")
	changeStr = strings.TrimPrefix(changeStr, "<span style='color:green'>-</span>")
	changeStr = strings.TrimSpace(changeStr)
	
	change, err := strconv.ParseFloat(changeStr, 64)
	if err != nil {
		change = 0
	}
	
	// 計算漲跌幅
	var changeRate float64
	if price > 0 && change != 0 {
		prevPrice := price - change
		if prevPrice > 0 {
			changeRate = (change / prevPrice) * 100
		}
	}
	
	// 成交量
	var volume float64
	if volumeIdx != -1 && len(latestRow) > volumeIdx {
		volumeStr := strings.ReplaceAll(strings.TrimSpace(latestRow[volumeIdx]), ",", "")
		volume, _ = strconv.ParseFloat(volumeStr, 64)
	}
	
	return &StockData{
		Code:       stock.Code,
		Name:       stock.Name,
		Price:      price,
		ChangeRate: changeRate,
		Volume:     volume,
		Category:   stock.Category,
	}, nil
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
