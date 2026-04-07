package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TWSE API 回應結構（每日收盤價）
type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

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

// 從證交所取得當日收盤價
func getTWSEPrice(code string) (float64, error) {
	// 取得今天日期 (yyyyMMdd)
	today := time.Now()
	dateStr := today.Format("20060102")
	
	// 證交所API網址
	url := fmt.Sprintf("https://www.twse.com.tw/exchangeReport/STOCK_DAY?response=json&date=%s&stockNo=%s", dateStr, code)
	
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
	// 設定 User-Agent 模擬瀏覽器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	var twseResp TWSEResponse
	if err := json.Unmarshal(body, &twseResp); err != nil {
		return 0, err
	}
	
	// 檢查回應狀態
	if twseResp.Stat != "OK" {
		return 0, fmt.Errorf("TWSE API 回應錯誤: %s", twseResp.Stat)
	}
	
	// 檢查是否有資料
	if len(twseResp.Data) == 0 {
		return 0, fmt.Errorf("查無股票代號 %s 的資料", code)
	}
	
	// 取得最新一天的資料（陣列最後一筆）
	lastData := twseResp.Data[len(twseResp.Data)-1]
	
	// 欄位順序：日期、成交股數、成交金額、開盤價、最高價、最低價、收盤價、漲跌價差、成交筆數
	// 收盤價在 index 6
	if len(lastData) < 7 {
		return 0, fmt.Errorf("資料格式錯誤")
	}
	
	closePrice := lastData[6]
	closePrice = fmt.Sprintf("%s", closePrice) // 確保是字串
	// 移除千分位逗號
	for i := 0; i < len(closePrice); i++ {
		if closePrice[i] == ',' {
			closePrice = closePrice[:i] + closePrice[i+1:]
			i--
		}
	}
	
	var price float64
	_, err = fmt.Sscanf(closePrice, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("無法解析收盤價: %s", closePrice)
	}
	
	return price, nil
}

func main() {
	fmt.Println("🦈 更新投資組合資料（使用 TWSE API）...")
	
	// 定義投資組合
	holdings := []Stock{
		{"2618", "長榮航", 1000, 33.70, 0, 0, 0, "RSI超賣20.6反彈機會"},
		{"5876", "上海商銀", 1000, 39.15, 0, 0, 0, "RSI偏低35.8買點"},
		{"1102", "亞泥", 1000, 34.75, 0, 0, 0, "穩健中性"},
		{"2353", "宏碁", 1000, 27.90, 0, 0, 0, "評分68最高,價格突破均線"},
		{"2851", "中再保", 1000, 28.85, 0, 0, 0, "評分62,上升趨勢"},
		{"2838", "聯邦銀", 1000, 20.15, 0, 0, 0, "價格低20元,RSI中性"},
		{"2812", "台中銀", 1000, 20.95, 0, 0, 0, "金融股穩健"},
		{"2887", "台新金", 1000, 24.65, 0, 0, 0, "今日上漲1.23%"},
	}
	
	totalCost := 0.0
	currentValue := 0.0
	
	// 查詢每支股票
	for i, stock := range holdings {
		price, err := getTWSEPrice(stock.Symbol)
		if err != nil {
			fmt.Printf("⚠️  %s.TW 取得股價失敗: %v\n", stock.Symbol, err)
			price = stock.BuyPrice // 使用成本價
		}
		
		holdings[i].CurrentPrice = price
		
		// 計算損益
		cost := stock.BuyPrice * float64(stock.Shares)
		value := price * float64(stock.Shares)
		pnl := value - cost
		returnPct := (pnl / cost) * 100
		
		holdings[i].ProfitLoss = pnl
		holdings[i].ReturnPct = returnPct
		
		totalCost += cost
		currentValue += value
		
		fmt.Printf("✅ %s (%s) - 現價: %.2f, 損益: %+.0f (%.2f%%)\n",
			stock.Symbol, stock.Name, price, pnl, returnPct)
		
		time.Sleep(500 * time.Millisecond) // 避免被限流
	}
	
	totalPnL := currentValue - totalCost
	totalReturn := (totalPnL / totalCost) * 100
	
	fmt.Printf("\n📊 總成本: %.0f | 市值: %.0f | 損益: %+.0f (%.2f%%)\n\n",
		totalCost, currentValue, totalPnL, totalReturn)
	
	// 產生 JSON
	portfolio := Portfolio{
		Holdings:     holdings,
		TotalCost:    totalCost,
		CurrentValue: currentValue,
		TotalPnL:     totalPnL,
		TotalReturn:  totalReturn,
		LastUpdate:   time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 儲存檔案
	homeDir, _ := os.UserHomeDir()
	outputPath := filepath.Join(homeDir, ".openclaw", "workspace", "stock_web", "portfolio.json")
	
	jsonData, _ := json.MarshalIndent(portfolio, "", "  ")
	if err := ioutil.WriteFile(outputPath, jsonData, 0644); err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 資料已儲存至 stock_web/portfolio.json\n")
	fmt.Printf("🌐 開啟 http://localhost:8081 即可查看網頁！\n")
}
