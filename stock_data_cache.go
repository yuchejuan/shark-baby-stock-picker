// stock_data_cache.go - 股票資料中央快取系統
// 避免重複呼叫 TWSE API，提升效能並節省 Token
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ========== 快取結構 ==========

type StockDataCache struct {
	Symbol       string          `json:"symbol"`
	Name         string          `json:"name"`
	LastUpdate   string          `json:"last_update"`
	UpdateTimeMs int64           `json:"update_time_ms"`
	CurrentPrice float64         `json:"current_price"`
	HistoryData  []CachedKLine   `json:"history_data"`
	DividendInfo *DividendCache  `json:"dividend_info,omitempty"`
}

type CachedKLine struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type DividendCache struct {
	CashDividend  float64 `json:"cash_dividend"`
	StockDividend float64 `json:"stock_dividend"`
	ExDate        string  `json:"ex_date"`
	Yield         float64 `json:"yield"`
	LastUpdate    string  `json:"last_update"`
}

type CacheIndex struct {
	Version     string            `json:"version"`
	LastUpdate  string            `json:"last_update"`
	StockCount  int               `json:"stock_count"`
	CacheFiles  map[string]string `json:"cache_files"` // symbol -> filepath
}

// ========== 設定 ==========

const (
	CacheDir     = ".cache/stock_data"
	CacheVersion = "1.0"
	CacheTTL     = 24 * time.Hour // 快取有效期 24 小時
)

// ========== TWSE API ==========

type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

// ========== 快取管理 ==========

// InitCache 初始化快取目錄
func InitCache() error {
	return os.MkdirAll(CacheDir, 0755)
}

// GetCacheFilePath 取得快取檔案路徑
func GetCacheFilePath(symbol string) string {
	return filepath.Join(CacheDir, fmt.Sprintf("%s.json", symbol))
}

// IsCacheValid 檢查快取是否有效
func IsCacheValid(cacheFile string) bool {
	info, err := os.Stat(cacheFile)
	if err != nil {
		return false
	}
	
	// 檢查檔案修改時間
	age := time.Since(info.ModTime())
	return age < CacheTTL
}

// LoadFromCache 從快取載入資料
func LoadFromCache(symbol string) (*StockDataCache, error) {
	cacheFile := GetCacheFilePath(symbol)
	
	// 檢查快取是否有效
	if !IsCacheValid(cacheFile) {
		return nil, fmt.Errorf("快取過期或不存在")
	}
	
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	
	var cache StockDataCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	
	return &cache, nil
}

// SaveToCache 儲存資料到快取
func SaveToCache(cache *StockDataCache) error {
	if err := InitCache(); err != nil {
		return err
	}
	
	cache.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	cache.UpdateTimeMs = time.Now().UnixMilli()
	
	cacheFile := GetCacheFilePath(cache.Symbol)
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(cacheFile, data, 0644)
}

// FetchFromTWSE 從證交所取得資料
func FetchFromTWSE(symbol string, yearMonth string) (*StockDataCache, error) {
	if len(yearMonth) >= 6 {
		yearMonth = yearMonth[:6]
	}
	
	url := fmt.Sprintf("https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=%s01&stockNo=%s", 
		yearMonth, symbol)
	
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
	
	// 轉換資料
	cache := &StockDataCache{
		Symbol:       symbol,
		HistoryData:  make([]CachedKLine, 0),
	}
	
	for _, row := range twseResp.Data {
		if len(row) < 9 {
			continue
		}
		
		// 解析日期
		dateStr := row[0]
		dateParts := strings.Split(dateStr, "/")
		if len(dateParts) == 3 {
			year := 0
			fmt.Sscanf(dateParts[0], "%d", &year)
			year += 1911
			dateStr = fmt.Sprintf("%04d-%s-%s", year, dateParts[1], dateParts[2])
		}
		
		// 解析價格和成交量
		var open, high, low, close float64
		var volume int64
		
		fmt.Sscanf(strings.ReplaceAll(row[3], ",", ""), "%f", &open)
		fmt.Sscanf(strings.ReplaceAll(row[4], ",", ""), "%f", &high)
		fmt.Sscanf(strings.ReplaceAll(row[5], ",", ""), "%f", &low)
		fmt.Sscanf(strings.ReplaceAll(row[6], ",", ""), "%f", &close)
		fmt.Sscanf(strings.ReplaceAll(row[1], ",", ""), "%d", &volume)
		
		kline := CachedKLine{
			Date:   dateStr,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		}
		
		cache.HistoryData = append(cache.HistoryData, kline)
	}
	
	// 取得最新價格
	if len(cache.HistoryData) > 0 {
		cache.CurrentPrice = cache.HistoryData[len(cache.HistoryData)-1].Close
	}
	
	return cache, nil
}

// GetStockData 取得股票資料（優先使用快取）
func GetStockData(symbol string, forceRefresh bool) (*StockDataCache, error) {
	// 如果不強制刷新，先嘗試從快取載入
	if !forceRefresh {
		cache, err := LoadFromCache(symbol)
		if err == nil {
			fmt.Printf("✅ 從快取載入: %s\n", symbol)
			return cache, nil
		}
	}
	
	// 快取無效或強制刷新，從 TWSE 取得資料
	fmt.Printf("🌐 從 TWSE 取得: %s\n", symbol)
	
	now := time.Now()
	cache, err := FetchFromTWSE(symbol, now.Format("200601"))
	if err != nil {
		return nil, err
	}
	
	// 如果本月資料不足，補充上個月
	if len(cache.HistoryData) < 20 {
		lastMonth := now.AddDate(0, -1, 0)
		lastMonthCache, err := FetchFromTWSE(symbol, lastMonth.Format("200601"))
		if err == nil {
			cache.HistoryData = append(lastMonthCache.HistoryData, cache.HistoryData...)
		}
		time.Sleep(300 * time.Millisecond)
	}
	
	// 儲存到快取
	if err := SaveToCache(cache); err != nil {
		fmt.Printf("⚠️  快取儲存失敗: %v\n", err)
	}
	
	return cache, nil
}

// BatchGetStockData 批次取得股票資料
func BatchGetStockData(symbols []string, forceRefresh bool) (map[string]*StockDataCache, error) {
	result := make(map[string]*StockDataCache)
	
	for i, symbol := range symbols {
		cache, err := GetStockData(symbol, forceRefresh)
		if err != nil {
			fmt.Printf("⚠️  %s 載入失敗: %v\n", symbol, err)
			continue
		}
		
		result[symbol] = cache
		
		// 進度顯示
		if (i+1)%10 == 0 {
			fmt.Printf("  ⏳ 進度: %d/%d\n", i+1, len(symbols))
		}
		
		// 避免請求過快（僅在從 TWSE 取得時延遲）
		if forceRefresh || !IsCacheValid(GetCacheFilePath(symbol)) {
			time.Sleep(500 * time.Millisecond)
		}
	}
	
	return result, nil
}

// ClearExpiredCache 清理過期快取
func ClearExpiredCache() error {
	if err := InitCache(); err != nil {
		return err
	}
	
	files, err := ioutil.ReadDir(CacheDir)
	if err != nil {
		return err
	}
	
	count := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		filePath := filepath.Join(CacheDir, file.Name())
		if !IsCacheValid(filePath) {
			os.Remove(filePath)
			count++
		}
	}
	
	fmt.Printf("🧹 清理 %d 個過期快取\n", count)
	return nil
}

// ========== 命令列工具 ==========

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法：")
		fmt.Println("  go run stock_data_cache.go get <股票代號>")
		fmt.Println("  go run stock_data_cache.go refresh <股票代號>")
		fmt.Println("  go run stock_data_cache.go clear")
		fmt.Println("  go run stock_data_cache.go stats")
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("請提供股票代號")
			return
		}
		symbol := os.Args[2]
		cache, err := GetStockData(symbol, false)
		if err != nil {
			fmt.Printf("錯誤: %v\n", err)
			return
		}
		fmt.Printf("✅ %s (%s) - 最新價: %.2f (共 %d 筆資料)\n", 
			cache.Symbol, cache.Name, cache.CurrentPrice, len(cache.HistoryData))
		
	case "refresh":
		if len(os.Args) < 3 {
			fmt.Println("請提供股票代號")
			return
		}
		symbol := os.Args[2]
		cache, err := GetStockData(symbol, true)
		if err != nil {
			fmt.Printf("錯誤: %v\n", err)
			return
		}
		fmt.Printf("✅ %s 已更新 - 最新價: %.2f\n", symbol, cache.CurrentPrice)
		
	case "clear":
		if err := ClearExpiredCache(); err != nil {
			fmt.Printf("錯誤: %v\n", err)
		}
		
	case "stats":
		InitCache()
		files, _ := ioutil.ReadDir(CacheDir)
		valid := 0
		expired := 0
		for _, file := range files {
			filePath := filepath.Join(CacheDir, file.Name())
			if IsCacheValid(filePath) {
				valid++
			} else {
				expired++
			}
		}
		fmt.Printf("📊 快取統計:\n")
		fmt.Printf("  有效: %d 個\n", valid)
		fmt.Printf("  過期: %d 個\n", expired)
		fmt.Printf("  總計: %d 個\n", valid+expired)
		
	default:
		fmt.Printf("未知命令: %s\n", command)
	}
}
