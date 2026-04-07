// daily_picker_v3.go - 每日選股 V3.0 (81支股票池 + 20-60元區間)
// TWSE API + 6大技術指標 + 100分制評分系統
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

type DailyReportV3 struct {
	Date      string      `json:"date"`
	Picks2030 []DailyPick `json:"picks_20_30"`
	Picks3040 []DailyPick `json:"picks_30_40"`
	Picks4050 []DailyPick `json:"picks_40_50"`
	Picks5060 []DailyPick `json:"picks_50_60"`
	BestPick  *DailyPick  `json:"best_pick"`
	Summary   SummaryV3   `json:"summary"`
}

type SummaryV3 struct {
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

// ========== TWSE API 函式（與 V2 相同）==========
// 為了節省空間，這裡保留原有的 TWSE API 函式
// GetTWSEHistoricalData, CalculateMA, CalculateRSI, CalculateMACDSignal, CalculateKDSignal, CalculateScore

