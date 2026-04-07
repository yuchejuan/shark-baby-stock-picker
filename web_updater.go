package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TWSE API 回應結構
type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

// 從 Trade API 取得的持倉格式
type HoldingBatch struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Shares       int     `json:"shares"`
	BuyPrice     float64 `json:"buy_price"`
	BuyDate      string  `json:"buy_date"`
	CurrentPrice float64 `json:"current_price"`
	CurrentValue float64 `json:"current_value"`
	ProfitLoss   float64 `json:"profit_loss"`
	ReturnPct    float64 `json:"return_pct"`
	Reason       string  `json:"reason"`
	DaysHeld     int     `json:"days_held"`
}

// portfolio.json 輸出格式
type Stock struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Shares       int     `json:"shares"`
	BuyPrice     float64 `json:"buy_price"`
	CurrentPrice float64 `json:"current_price"`
	ProfitLoss   float64 `json:"profit_loss"`
	ReturnPct    float64 `json:"return_pct"`
	Reason       string  `json:"reason"`
}

type Portfolio struct {
	Holdings     []Stock `json:"holdings"`
	TotalCost    float64 `json:"total_cost"`
	CurrentValue float64 `json:"current_value"`
	TotalPnL     float64 `json:"total_pnl"`
	TotalReturn  float64 `json:"total_return"`
	LastUpdate   string  `json:"last_update"`
}

// 從 Trade API 取得持倉
func getHoldingsFromAPI() ([]HoldingBatch, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:8888/api/holdings/batches")
	if err != nil {
		return nil, fmt.Errorf("Trade API 無回應: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var batches []HoldingBatch
	if err := json.Unmarshal(body, &batches); err != nil {
		return nil, fmt.Errorf("解析持倉資料失敗: %v", err)
	}
	return batches, nil
}

// 從 TWSE 取得收盤價
func getTWSEPrice(code string) (float64, error) {
	dateStr := time.Now().Format("20060102")
	url := fmt.Sprintf(
		"https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=%s&stockNo=%s",
		dateStr, code,
	)

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var twseResp TWSEResponse
	if err := json.Unmarshal(body, &twseResp); err != nil {
		return 0, err
	}
	if twseResp.Stat != "OK" || len(twseResp.Data) == 0 {
		return 0, fmt.Errorf("查無資料（休市或代號錯誤）")
	}

	lastRow := twseResp.Data[len(twseResp.Data)-1]
	if len(lastRow) < 7 {
		return 0, fmt.Errorf("資料格式異常")
	}

	priceStr := strings.ReplaceAll(lastRow[6], ",", "")
	var price float64
	fmt.Sscanf(priceStr, "%f", &price)
	return price, nil
}

func main() {
	fmt.Println("🦈 更新投資組合資料（從 Trade API 讀取持倉）...")

	// Step 1：從 Trade API 取得目前持倉
	batches, err := getHoldingsFromAPI()
	if err != nil {
		fmt.Printf("⚠️  無法連線 Trade API: %v\n", err)
		fmt.Println("   請先啟動 trade_manager：go build -o trade_manager trade_manager.go && ./trade_manager")
		fmt.Println("   或使用 bash start_all.sh 一鍵啟動")
		os.Exit(1)
	}

	if len(batches) == 0 {
		fmt.Println("📭 目前無持倉記錄，請先到網頁買入股票。")
		os.Exit(0)
	}

	fmt.Printf("📋 讀取到 %d 筆持倉\n\n", len(batches))

	// Step 2：查詢每支股票最新股價
	var holdings []Stock
	totalCost := 0.0
	currentValue := 0.0

	for _, b := range batches {
		price, err := getTWSEPrice(b.Symbol)
		if err != nil {
			fmt.Printf("⚠️  %s (%s) 取得股價失敗: %v，使用買入價\n", b.Symbol, b.Name, err)
			price = b.BuyPrice
		}

		cost := b.BuyPrice * float64(b.Shares)
		val := price * float64(b.Shares)
		pnl := val - cost
		ret := 0.0
		if cost > 0 {
			ret = pnl / cost * 100
		}

		holdings = append(holdings, Stock{
			Symbol:       b.Symbol,
			Name:         b.Name,
			Shares:       b.Shares,
			BuyPrice:     b.BuyPrice,
			CurrentPrice: price,
			ProfitLoss:   pnl,
			ReturnPct:    ret,
			Reason:       b.Reason,
		})

		totalCost += cost
		currentValue += val

		sign := "+"
		if pnl < 0 {
			sign = ""
		}
		fmt.Printf("✅ %s (%s) %d股 | 買入 %.2f → 現價 %.2f | 損益 %s%.0f (%.2f%%)\n",
			b.Symbol, b.Name, b.Shares, b.BuyPrice, price, sign, pnl, ret)

		time.Sleep(500 * time.Millisecond)
	}

	// Step 3：儲存到 html/portfolio.json
	totalPnL := currentValue - totalCost
	totalReturn := 0.0
	if totalCost > 0 {
		totalReturn = totalPnL / totalCost * 100
	}

	portfolio := Portfolio{
		Holdings:     holdings,
		TotalCost:    totalCost,
		CurrentValue: currentValue,
		TotalPnL:     totalPnL,
		TotalReturn:  totalReturn,
		LastUpdate:   time.Now().Format("2006-01-02 15:04:05"),
	}

	wd, _ := os.Getwd()
	outputPath := filepath.Join(wd, "html", "portfolio.json")
	os.MkdirAll(filepath.Join(wd, "html"), 0755)

	jsonData, _ := json.MarshalIndent(portfolio, "", "  ")
	if err := ioutil.WriteFile(outputPath, jsonData, 0644); err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		os.Exit(1)
	}

	sign := "+"
	if totalPnL < 0 {
		sign = ""
	}
	fmt.Printf("\n📊 總成本: %.0f | 市值: %.0f | 損益: %s%.0f (%.2f%%)\n",
		totalCost, currentValue, sign, totalPnL, totalReturn)
	fmt.Printf("✅ 已更新 %d 支股票，儲存至 html/portfolio.json\n", len(holdings))
}
