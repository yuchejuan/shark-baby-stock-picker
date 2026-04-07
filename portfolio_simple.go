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

// Position 持倉
type Position struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	BuyDate    string  `json:"buyDate"`
	BuyPrice   float64 `json:"buyPrice"`
	Shares     int     `json:"shares"`
	Reason     string  `json:"reason"`
}

// Portfolio 投資組合
type Portfolio struct {
	Positions  []Position `json:"positions"`
	UpdateTime string     `json:"updateTime"`
}

// TWSE API 回應結構
type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

const portfolioFile = "stock_web/portfolio.json"

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
	
	body, err := io.ReadAll(resp.Body)
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
	closePrice = strings.ReplaceAll(closePrice, ",", "") // 移除千分位逗號
	
	var price float64
	_, err = fmt.Sscanf(closePrice, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("無法解析收盤價: %s", closePrice)
	}
	
	return price, nil
}

// 帶重試機制的價格抓取
func getCurrentPrice(code string) (float64, error) {
	var lastErr error
	maxRetries := 3
	
	for i := 0; i < maxRetries; i++ {
		price, err := getTWSEPrice(code)
		if err == nil {
			return price, nil
		}
		
		lastErr = err
		
		// 重試前等待（避免被證交所擋）
		if i < maxRetries-1 {
			time.Sleep(time.Duration(1+i) * time.Second)
		}
	}
	
	return 0, fmt.Errorf("重試 %d 次後失敗: %v", maxRetries, lastErr)
}

func loadPortfolio() *Portfolio {
	if _, err := os.Stat(portfolioFile); os.IsNotExist(err) {
		return &Portfolio{Positions: []Position{}}
	}
	
	data, _ := ioutil.ReadFile(portfolioFile)
	var portfolio Portfolio
	json.Unmarshal(data, &portfolio)
	return &portfolio
}

func savePortfolio(portfolio *Portfolio) {
	portfolio.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	data, _ := json.MarshalIndent(portfolio, "", "  ")
	ioutil.WriteFile(portfolioFile, data, 0644)
}

func viewPortfolio() {
	portfolio := loadPortfolio()
	
	if len(portfolio.Positions) == 0 {
		fmt.Println("\n⚠️  目前沒有持倉")
		fmt.Println("\n💡 使用方式:")
		fmt.Println("   go run portfolio_simple.go add 2330 台積電 1000 100 \"技術面突破\"")
		return
	}
	
	fmt.Println("\n📊 模擬投資組合報告")
	fmt.Println(strings.Repeat("=", 130))
	fmt.Printf("⏰ 更新時間: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	
	type Status struct {
		Position Position
		Current  float64
		Profit   float64
		Percent  float64
		Days     int
	}
	
	var statuses []Status
	totalCost := 0.0
	totalValue := 0.0
	
	for _, pos := range portfolio.Positions {
		current, err := getCurrentPrice(pos.Code)
		if err != nil {
			fmt.Printf("⚠️  %s (%s): 無法取得股價\n", pos.Name, pos.Code)
			continue
		}
		
		cost := pos.BuyPrice * float64(pos.Shares)
		value := current * float64(pos.Shares)
		profit := value - cost
		percent := (profit / cost) * 100
		
		buyDate, _ := time.Parse("2006-01-02", pos.BuyDate)
		days := int(time.Since(buyDate).Hours() / 24)
		
		statuses = append(statuses, Status{
			Position: pos,
			Current:  current,
			Profit:   profit,
			Percent:  percent,
			Days:     days,
		})
		
		totalCost += cost
		totalValue += value
		
		time.Sleep(200 * time.Millisecond)
	}
	
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Percent > statuses[j].Percent
	})
	
	fmt.Println("持倉明細")
	fmt.Println(strings.Repeat("=", 130))
	fmt.Printf("%-8s %-14s %-12s %10s %10s %8s %12s %12s %12s %10s %8s\n",
		"代號", "名稱", "買入日期", "買入價", "現價", "股數", "成本", "市值", "損益", "報酬率", "持有天")
	fmt.Println(strings.Repeat("=", 130))
	
	for _, s := range statuses {
		icon := " "
		if s.Percent > 0 {
			icon = "📈"
		} else if s.Percent < 0 {
			icon = "📉"
		}
		
		fmt.Printf("%-8s %-14s %-12s %10.2f %10.2f %8d %12.0f %12.0f %12.0f %9.2f%% %8d %s\n",
			s.Position.Code, s.Position.Name, s.Position.BuyDate,
			s.Position.BuyPrice, s.Current, s.Position.Shares,
			s.Position.BuyPrice*float64(s.Position.Shares),
			s.Current*float64(s.Position.Shares),
			s.Profit, s.Percent, s.Days, icon)
		
		if s.Position.Reason != "" {
			fmt.Printf("         理由: %s\n", s.Position.Reason)
		}
	}
	
	fmt.Println(strings.Repeat("=", 130))
	
	totalProfit := totalValue - totalCost
	totalPercent := (totalProfit / totalCost) * 100
	
	fmt.Println("\n📊 組合總結")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("💰 總成本:   %12.0f 元\n", totalCost)
	fmt.Printf("💎 總市值:   %12.0f 元\n", totalValue)
	fmt.Printf("📊 總損益:   %12.0f 元 (%+.2f%%)\n", totalProfit, totalPercent)
	
	profitable := 0
	for _, s := range statuses {
		if s.Percent > 0 {
			profitable++
		}
	}
	
	fmt.Printf("\n✅ 獲利: %d/%d 支 (%.1f%%)\n",
		profitable, len(statuses), float64(profitable)/float64(len(statuses))*100)
	
	if len(statuses) >= 3 {
		fmt.Println("\n🏆 表現最佳 TOP 3")
		fmt.Println(strings.Repeat("-", 70))
		for i := 0; i < 3; i++ {
			s := statuses[i]
			fmt.Printf("#%d  %s (%s)  %+.2f%%  獲利 %.0f 元\n",
				i+1, s.Position.Name, s.Position.Code, s.Percent, s.Profit)
		}
	}
	
	fmt.Println("\n💡 提示: 使用 'go run portfolio_simple.go' 查看組合")
	fmt.Println("        使用 'go run portfolio_simple.go add ...' 新增持倉")
	fmt.Println("        使用 'go run portfolio_simple.go list' 列出代號")
}

func addPosition(code, name string, price float64, shares int, reason string) {
	portfolio := loadPortfolio()
	
	pos := Position{
		Code:     code,
		Name:     name,
		BuyDate:  time.Now().Format("2006-01-02"),
		BuyPrice: price,
		Shares:   shares,
		Reason:   reason,
	}
	
	portfolio.Positions = append(portfolio.Positions, pos)
	savePortfolio(portfolio)
	
	fmt.Printf("\n✅ 新增成功！\n")
	fmt.Printf("📊 %s (%s) %d股 @ %.2f = %.0f 元\n",
		name, code, shares, price, price*float64(shares))
	fmt.Printf("📝 買入理由: %s\n", reason)
}

func listPositions() {
	portfolio := loadPortfolio()
	
	if len(portfolio.Positions) == 0 {
		fmt.Println("\n⚠️  目前沒有持倉")
		return
	}
	
	fmt.Println("\n📋 持倉列表")
	fmt.Println(strings.Repeat("-", 60))
	for i, pos := range portfolio.Positions {
		fmt.Printf("%d. %s (%s) - %d股 @ %.2f (%s)\n",
			i+1, pos.Name, pos.Code, pos.Shares, pos.BuyPrice, pos.BuyDate)
	}
}

func main() {
	if len(os.Args) == 1 {
		viewPortfolio()
		return
	}
	
	cmd := os.Args[1]
	
	switch cmd {
	case "add":
		if len(os.Args) < 6 {
			fmt.Println("用法: go run portfolio_simple.go add <代號> <名稱> <買入價> <股數> [理由]")
			fmt.Println("範例: go run portfolio_simple.go add 2330 台積電 1000 100 \"技術面突破\"")
			return
		}
		
		code := os.Args[2]
		name := os.Args[3]
		
		var price float64
		var shares int
		fmt.Sscanf(os.Args[4], "%f", &price)
		fmt.Sscanf(os.Args[5], "%d", &shares)
		
		reason := ""
		if len(os.Args) > 6 {
			reason = strings.Join(os.Args[6:], " ")
		}
		
		addPosition(code, name, price, shares, reason)
		
	case "list":
		listPositions()
		
	case "view":
		viewPortfolio()
		
	default:
		fmt.Println("未知指令:", cmd)
		fmt.Println("\n可用指令:")
		fmt.Println("  (無參數)  - 查看投資組合")
		fmt.Println("  add       - 新增持倉")
		fmt.Println("  list      - 列出持倉")
		fmt.Println("  view      - 查看報告")
	}
}
