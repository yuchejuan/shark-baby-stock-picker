// stock_pool_loader.go - 股票池載入工具
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// StockPool 股票池結構
type StockPool struct {
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

// LoadStockPool 載入股票池
func LoadStockPool(filepath string) (*StockPool, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("無法讀取股票池設定: %v", err)
	}
	
	var pool StockPool
	if err := json.Unmarshal(data, &pool); err != nil {
		return nil, fmt.Errorf("無法解析股票池設定: %v", err)
	}
	
	return &pool, nil
}

// GetAllStocks 取得所有股票（ETF + 個股）
func (p *StockPool) GetAllStocks() map[string]string {
	result := make(map[string]string)
	
	// 加入 ETF
	for code, name := range p.ETF.List {
		result[code] = name
	}
	
	// 加入個股
	for _, category := range p.Stocks.Categories {
		for code, name := range category.List {
			// 避免重複（如果有的話）
			if _, exists := result[code]; !exists {
				result[code] = name
			}
		}
	}
	
	return result
}

// GetStocksByCategory 依類別取得股票
func (p *StockPool) GetStocksByCategory(categoryName string) (map[string]string, error) {
	if categoryName == "etf" {
		return p.ETF.List, nil
	}
	
	category, exists := p.Stocks.Categories[categoryName]
	if !exists {
		return nil, fmt.Errorf("類別不存在: %s", categoryName)
	}
	
	return category.List, nil
}

// PrintSummary 顯示股票池摘要
func (p *StockPool) PrintSummary() {
	fmt.Println("📊 股票池總覽")
	fmt.Println(fmt.Sprintf("版本: %s | 更新日期: %s | 總計: %d 支", p.Version, p.LastUpdate, p.Total))
	fmt.Println()
	
	fmt.Printf("📦 ETF: %d 支\n", p.ETF.Count)
	for code, name := range p.ETF.List {
		fmt.Printf("  - %s (%s)\n", name, code)
	}
	fmt.Println()
	
	fmt.Printf("📦 個股: %d 支\n", p.Stocks.Count)
	for catKey, category := range p.Stocks.Categories {
		fmt.Printf("  【%s】%d 支\n", category.Name, category.Count)
		for code, name := range category.List {
			fmt.Printf("    - %s (%s)\n", name, code)
		}
	}
	fmt.Println()
	
	fmt.Println("📏 價格區間:")
	for _, r := range p.PriceRanges {
		fmt.Printf("  - %s (%d-%d元)\n", r.Label, r.Min, r.Max)
	}
}
