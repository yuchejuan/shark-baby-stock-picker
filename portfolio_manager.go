package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// 持股資料
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

// 投資組合
type Portfolio struct {
	Holdings     []Holding `json:"holdings"`
	TotalCost    float64   `json:"total_cost"`
	CurrentValue float64   `json:"current_value"`
	TotalPnL     float64   `json:"total_pnl"`
	TotalReturn  float64   `json:"total_return"`
	LastUpdate   string    `json:"last_update"`
}

// 買入請求
type BuyRequest struct {
	Symbol string  `json:"symbol"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
	Shares int     `json:"shares"` // 張數
	Reason string  `json:"reason"`
}

var portfolioFile string

func init() {
	wd, _ := os.Getwd()
	portfolioFile = wd + "/html/portfolio.json"
	os.MkdirAll(wd+"/html", 0755)
}

// 讀取投資組合
func loadPortfolio() (*Portfolio, error) {
	data, err := ioutil.ReadFile(portfolioFile)
	if err != nil {
		return nil, err
	}
	
	var portfolio Portfolio
	if err := json.Unmarshal(data, &portfolio); err != nil {
		return nil, err
	}
	
	return &portfolio, nil
}

// 儲存投資組合
func savePortfolio(portfolio *Portfolio) error {
	portfolio.LastUpdate = time.Now().Format("2006-01-02 15:04:05")
	
	data, err := json.MarshalIndent(portfolio, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(portfolioFile, data, 0644)
}

// 處理買入請求
func handleBuy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// 解析請求
	var req BuyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request",
		})
		return
	}
	
	log.Printf("📈 買入請求: %s %s, 數量: %d張, 價格: %.2f",
		req.Symbol, req.Name, req.Shares, req.Price)
	
	// 讀取投資組合
	portfolio, err := loadPortfolio()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to load portfolio: %v", err),
		})
		return
	}
	
	// 計算總股數
	totalShares := req.Shares * 1000
	
	// 檢查是否已持有
	existingIndex := -1
	for i, h := range portfolio.Holdings {
		if h.Symbol == req.Symbol {
			existingIndex = i
			break
		}
	}
	
	var message string
	
	if existingIndex >= 0 {
		// 已持有，加碼
		existing := portfolio.Holdings[existingIndex]
		newTotalShares := existing.Shares + totalShares
		newTotalCost := (existing.BuyPrice * float64(existing.Shares)) + (req.Price * float64(totalShares))
		newAvgPrice := newTotalCost / float64(newTotalShares)
		
		portfolio.Holdings[existingIndex] = Holding{
			Symbol:       req.Symbol,
			Name:         req.Name,
			Shares:       newTotalShares,
			BuyPrice:     newAvgPrice,
			CurrentPrice: req.Price,
			ProfitLoss:   0,
			ReturnPct:    0,
			Reason:       req.Reason,
		}
		
		message = fmt.Sprintf("加碼成功！持股 %d張 → %d張，平均成本 %.2f → %.2f",
			existing.Shares/1000, newTotalShares/1000, existing.BuyPrice, newAvgPrice)
		
		log.Printf("✅ %s", message)
	} else {
		// 新買入
		portfolio.Holdings = append(portfolio.Holdings, Holding{
			Symbol:       req.Symbol,
			Name:         req.Name,
			Shares:       totalShares,
			BuyPrice:     req.Price,
			CurrentPrice: req.Price,
			ProfitLoss:   0,
			ReturnPct:    0,
			Reason:       req.Reason,
		})
		
		message = fmt.Sprintf("買入成功！%s (%s) %d張 @ %.2f = $%.0f",
			req.Name, req.Symbol, req.Shares, req.Price, req.Price*float64(totalShares))
		
		log.Printf("✅ %s", message)
	}
	
	// 更新總計
	portfolio.TotalCost = 0
	portfolio.CurrentValue = 0
	for _, h := range portfolio.Holdings {
		portfolio.TotalCost += h.BuyPrice * float64(h.Shares)
		portfolio.CurrentValue += h.CurrentPrice * float64(h.Shares)
	}
	portfolio.TotalPnL = portfolio.CurrentValue - portfolio.TotalCost
	portfolio.TotalReturn = (portfolio.TotalPnL / portfolio.TotalCost) * 100
	
	// 儲存
	if err := savePortfolio(portfolio); err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to save portfolio: %v", err),
		})
		return
	}
	
	// 返回結果
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": message,
		"portfolio": portfolio,
	})
}

// 處理投資組合查詢
func handleGetPortfolio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	portfolio, err := loadPortfolio()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to load portfolio: %v", err),
		})
		return
	}
	
	json.NewEncoder(w).Encode(portfolio)
}

func main() {
	http.HandleFunc("/api/portfolio/buy", handleBuy)
	http.HandleFunc("/api/portfolio", handleGetPortfolio)
	
	port := "8766"
	fmt.Printf("🦈 投資組合管理 API 啟動於 http://localhost:%s\n", port)
	fmt.Println("📡 端點:")
	fmt.Println("  POST /api/portfolio/buy   - 模擬買入")
	fmt.Println("  GET  /api/portfolio       - 查詢投資組合")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
