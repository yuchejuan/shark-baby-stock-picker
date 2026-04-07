package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// 持股資料結構
type Holding struct {
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Shares       int     `json:"shares"`
	BuyPrice     float64 `json:"buy_price"`
	CurrentPrice float64 `json:"current_price"`
	ProfitLoss   float64 `json:"profit_loss"`
	ReturnPct    float64 `json:"return_pct"`
	Reason       string  `json:"reason"`
}

type PortfolioData struct {
	Holdings     []Holding `json:"holdings"`
	TotalCost    float64   `json:"total_cost"`
	CurrentValue float64   `json:"current_value"`
	TotalPnL     float64   `json:"total_pnl"`
	TotalReturn  float64   `json:"total_return"`
	LastUpdate   string    `json:"last_update"`
}

// TWSE 即時股價 API 回應
type TWSeQuote struct {
	MsgArray []struct {
		C string `json:"c"` // 股票代號
		N string `json:"n"` // 股票名稱
		Z string `json:"z"` // 成交價
		Y string `json:"y"` // 昨收價
	} `json:"msgArray"`
}

func main() {
	fmt.Println("🦈 即時股價更新系統啟動")
	fmt.Println("⏰ 更新間隔：20 分鐘")
	fmt.Println("📅 運作時間：週一至週五 09:00-14:00")
	fmt.Println("")

	// 無限循環
	for {
		now := time.Now()
		
		// 檢查是否為週末
		if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			fmt.Printf("[%s] 週末不交易，休息中...\n", now.Format("15:04:05"))
			time.Sleep(1 * time.Hour)
			continue
		}
		
		// 檢查是否在盤中時間（09:00-14:00）
		hour := now.Hour()
		minute := now.Minute()
		
		if hour < 9 || (hour >= 14 && minute > 0) {
			nextUpdate := getNextUpdateTime(now)
			waitDuration := time.Until(nextUpdate)
			fmt.Printf("[%s] 非盤中時間，等待至 %s\n", now.Format("15:04:05"), nextUpdate.Format("15:04:05"))
			time.Sleep(waitDuration)
			continue
		}
		
		// 執行更新
		fmt.Printf("[%s] 🔄 開始更新股價...\n", now.Format("15:04:05"))
		err := updatePrices()
		if err != nil {
			log.Printf("❌ 更新失敗: %v\n", err)
		} else {
			fmt.Printf("[%s] ✅ 股價更新完成\n", now.Format("15:04:05"))
		}
		
		// 等待 20 分鐘
		fmt.Printf("[%s] ⏳ 下次更新時間: %s\n\n", now.Format("15:04:05"), now.Add(20*time.Minute).Format("15:04:05"))
		time.Sleep(20 * time.Minute)
	}
}

// 計算下次更新時間
func getNextUpdateTime(now time.Time) time.Time {
	// 如果在 14:00 之後，返回明天 09:00
	if now.Hour() >= 14 {
		tomorrow := now.AddDate(0, 0, 1)
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, tomorrow.Location())
	}
	
	// 如果在 09:00 之前，返回今天 09:00
	if now.Hour() < 9 {
		return time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
	}
	
	return now
}

// 更新股價
func updatePrices() error {
	// 1. 讀取 portfolio.json
	portfolioPath := "stock_web/portfolio.json"
	data, err := ioutil.ReadFile(portfolioPath)
	if err != nil {
		return fmt.Errorf("無法讀取 portfolio.json: %v", err)
	}
	
	var portfolio PortfolioData
	if err := json.Unmarshal(data, &portfolio); err != nil {
		return fmt.Errorf("無法解析 portfolio.json: %v", err)
	}
	
	// 2. 取得所有股票代號
	symbols := []string{}
	for _, h := range portfolio.Holdings {
		symbols = append(symbols, h.Symbol)
	}
	
	if len(symbols) == 0 {
		fmt.Println("⚠️  目前無持股，跳過更新")
		return nil
	}
	
	// 3. 從 TWSE API 取得即時股價
	prices, err := fetchTWSEPrices(symbols)
	if err != nil {
		return fmt.Errorf("無法取得股價: %v", err)
	}
	
	// 4. 更新持股資料
	totalCost := 0.0
	currentValue := 0.0
	
	for i := range portfolio.Holdings {
		h := &portfolio.Holdings[i]
		
		// 更新現價
		if price, ok := prices[h.Symbol]; ok {
			h.CurrentPrice = price
			fmt.Printf("  📊 %s (%s): $%.2f\n", h.Symbol, h.Name, price)
		} else {
			fmt.Printf("  ⚠️  %s (%s): 無法取得價格，使用舊價格 $%.2f\n", h.Symbol, h.Name, h.CurrentPrice)
		}
		
		// 重新計算損益
		cost := float64(h.Shares) * h.BuyPrice
		value := float64(h.Shares) * h.CurrentPrice
		h.ProfitLoss = value - cost
		h.ReturnPct = (h.ProfitLoss / cost) * 100
		
		totalCost += cost
		currentValue += value
	}
	
	// 5. 更新總覽
	portfolio.TotalCost = totalCost
	portfolio.CurrentValue = currentValue
	portfolio.TotalPnL = currentValue - totalCost
	portfolio.TotalReturn = (portfolio.TotalPnL / totalCost) * 100
	portfolio.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	
	// 6. 寫回 portfolio.json
	updatedData, err := json.MarshalIndent(portfolio, "", "  ")
	if err != nil {
		return fmt.Errorf("無法序列化資料: %v", err)
	}
	
	if err := ioutil.WriteFile(portfolioPath, updatedData, 0644); err != nil {
		return fmt.Errorf("無法寫入檔案: %v", err)
	}
	
	fmt.Printf("  💰 總成本: $%.0f | 市值: $%.0f | 損益: %+.0f (%.2f%%)\n",
		totalCost, currentValue, portfolio.TotalPnL, portfolio.TotalReturn)
	
	return nil
}

// 從 TWSE API 取得股價
func fetchTWSEPrices(symbols []string) (map[string]float64, error) {
	prices := make(map[string]float64)
	
	// TWSE 即時股價 API（非官方，但穩定）
	// 官方 API: https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_XXXX.tw
	
	for _, symbol := range symbols {
		url := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_%s.tw", symbol)
		
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("⚠️  無法取得 %s 的股價: %v", symbol, err)
			continue
		}
		defer resp.Body.Close()
		
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("⚠️  無法讀取 %s 的回應: %v", symbol, err)
			continue
		}
		
		var quote TWSeQuote
		if err := json.Unmarshal(body, &quote); err != nil {
			log.Printf("⚠️  無法解析 %s 的資料: %v", symbol, err)
			continue
		}
		
		if len(quote.MsgArray) > 0 && quote.MsgArray[0].Z != "-" {
			// 解析價格
			priceStr := strings.TrimSpace(quote.MsgArray[0].Z)
			var price float64
			fmt.Sscanf(priceStr, "%f", &price)
			prices[symbol] = price
		}
		
		// 避免過於頻繁請求
		time.Sleep(200 * time.Millisecond)
	}
	
	return prices, nil
}
