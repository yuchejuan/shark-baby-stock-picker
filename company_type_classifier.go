package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 股票資料結構
type StockFinancial struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	EPS             float64 `json:"eps"`              // 最近一年 EPS
	EPSPrev         float64 `json:"eps_prev"`         // 前一年 EPS
	EPSTrend        string  `json:"eps_trend"`        // 成長/穩定/衰退
	PayoutRatio     float64 `json:"payout_ratio"`     // 配息率 (%)
	CashDividend    float64 `json:"cash_dividend"`    // 現金股利
	CompanyType     string  `json:"company_type"`     // 公司類型
	TypeEmoji       string  `json:"type_emoji"`       // 類型圖示
	TypeDescription string  `json:"type_description"` // 類型說明
}

// 更新後的配息資料結構
type DividendData struct {
	Symbol               string  `json:"symbol"`
	Name                 string  `json:"name"`
	CashDividend         float64 `json:"cash_dividend"`
	StockDividend        float64 `json:"stock_dividend"`
	TotalDividend        float64 `json:"total_dividend"`
	ExDividendDate       string  `json:"ex_dividend_date"`
	CurrentPrice         float64 `json:"current_price"`
	Yield                float64 `json:"yield"`
	Consecutive10Years   bool    `json:"consecutive_10_years"`
	DividendCount10Year  int     `json:"dividend_count_10year"`
	UpdateTime           string  `json:"update_time"`
	EPS                  float64 `json:"eps"`
	EPSPrev              float64 `json:"eps_prev"`
	EPSTrend             string  `json:"eps_trend"`
	PayoutRatio          float64 `json:"payout_ratio"`
	CompanyType          string  `json:"company_type"`
	TypeEmoji            string  `json:"type_emoji"`
	TypeDescription      string  `json:"type_description"`
}

func main() {
	fmt.Println("🦈 鯊魚寶寶公司類型分類器")
	fmt.Println("========================================")
	
	// 讀取現有配息資料
	fmt.Println("📊 讀取配息資料...")
	dividendData, err := loadDividendData()
	if err != nil {
		fmt.Printf("❌ 讀取失敗: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ 已載入 %d 支股票\n", len(dividendData))
	fmt.Println("")
	
	// 分析每支股票
	fmt.Println("🔍 開始分析公司類型...")
	analyzed := 0
	
	for symbol, data := range dividendData {
		if symbol == "" || data.CashDividend == 0 {
			continue
		}
		
		// 抓取 EPS 資料
		eps, epsPrev, err := fetchEPS(symbol)
		if err != nil {
			fmt.Printf("  ⚠️  %s (%s) - EPS 資料不足\n", data.Name, symbol)
			continue
		}
		
		// 計算配息率
		payoutRatio := 0.0
		if eps > 0 {
			payoutRatio = (data.CashDividend / eps) * 100
		}
		
		// 判斷 EPS 趨勢
		epsTrend := judgeEPSTrend(eps, epsPrev)
		
		// 判斷公司類型
		companyType, emoji, description := classifyCompany(eps, epsPrev, payoutRatio, epsTrend)
		
		// 更新資料
		data.EPS = eps
		data.EPSPrev = epsPrev
		data.EPSTrend = epsTrend
		data.PayoutRatio = payoutRatio
		data.CompanyType = companyType
		data.TypeEmoji = emoji
		data.TypeDescription = description
		
		dividendData[symbol] = data
		analyzed++
		
		if analyzed%10 == 0 {
			fmt.Printf("  ⏳ 進度: %d/%d\n", analyzed, len(dividendData))
			time.Sleep(1 * time.Second) // 避免 API 限制
		}
	}
	
	fmt.Printf("\n✅ 分析完成：%d 支股票\n", analyzed)
	fmt.Println("")
	
	// 儲存結果
	fmt.Println("💾 儲存結果...")
	err = saveDividendData(dividendData)
	if err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ 資料已更新至 stock_web/dividend_data.json")
	
	// 顯示統計
	showStatistics(dividendData)
	
	fmt.Println("========================================")
	fmt.Println("🦈 分析完成")
}

func loadDividendData() (map[string]DividendData, error) {
	data, err := os.ReadFile("stock_web/dividend_data.json")
	if err != nil {
		return nil, err
	}
	
	var result map[string]DividendData
	err = json.Unmarshal(data, &result)
	return result, err
}

func saveDividendData(data map[string]DividendData) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile("stock_web/dividend_data.json", jsonData, 0644)
}

// 從 Goodinfo 抓取 EPS 資料
func fetchEPS(symbol string) (float64, float64, error) {
	url := fmt.Sprintf("https://goodinfo.tw/tw/StockBzPerformance.asp?STOCK_ID=%s", symbol)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}
	
	html := string(body)
	
	// 解析 EPS（簡化版，實際需要更精確的 HTML 解析）
	// 這裡使用正則表達式尋找 EPS 數值
	re := regexp.MustCompile(`EPS.*?(\d+\.\d+)`)
	matches := re.FindAllStringSubmatch(html, -1)
	
	if len(matches) < 2 {
		return 0, 0, fmt.Errorf("EPS 資料不足")
	}
	
	eps, _ := strconv.ParseFloat(matches[0][1], 64)
	epsPrev, _ := strconv.ParseFloat(matches[1][1], 64)
	
	return eps, epsPrev, nil
}

// 判斷 EPS 趨勢
func judgeEPSTrend(eps, epsPrev float64) string {
	if epsPrev == 0 {
		return "資料不足"
	}
	
	change := ((eps - epsPrev) / epsPrev) * 100
	
	if change > 10 {
		return "成長"
	} else if change < -10 {
		return "衰退"
	} else {
		return "穩定"
	}
}

// 分類公司類型
func classifyCompany(eps, epsPrev, payoutRatio float64, epsTrend string) (string, string, string) {
	// 1. EPS 成長，且配息率 < 60% → 成長型好公司 ✅
	if epsTrend == "成長" && payoutRatio < 60 {
		return "成長型好公司",
			"✅",
			fmt.Sprintf("EPS 成長且保留盈餘再投資，配息率 %.1f%% 健康", payoutRatio)
	}
	
	// 2. EPS 穩定，且配息率 50–70% → 穩健型 ✅
	if epsTrend == "穩定" && payoutRatio >= 50 && payoutRatio <= 70 {
		return "穩健型",
			"✅",
			fmt.Sprintf("EPS 穩定，配息率 %.1f%% 適中，穩定配息", payoutRatio)
	}
	
	// 3. EPS 衰退，且配息率 > 80% → 撐配息 ⚠️
	if epsTrend == "衰退" && payoutRatio > 80 {
		return "撐配息",
			"⚠️",
			fmt.Sprintf("EPS 衰退但仍高配息（%.1f%%），需留意持續性", payoutRatio)
	}
	
	// 4. EPS 偏低，且配息率 > 100% → 吃老本 ❌
	if payoutRatio > 100 {
		return "吃老本",
			"❌",
			fmt.Sprintf("配息率 %.1f%% 超過盈餘，動用保留盈餘配息", payoutRatio)
	}
	
	// 其他情況
	if payoutRatio > 70 && payoutRatio <= 100 {
		return "高配息",
			"🟡",
			fmt.Sprintf("配息率 %.1f%% 偏高，盈餘多用於配息", payoutRatio)
	}
	
	return "一般",
		"➡️",
		fmt.Sprintf("EPS %s，配息率 %.1f%%", epsTrend, payoutRatio)
}

func showStatistics(data map[string]DividendData) {
	stats := make(map[string]int)
	
	for _, stock := range data {
		if stock.CompanyType != "" {
			stats[stock.CompanyType]++
		}
	}
	
	fmt.Println("\n📊 分類統計:")
	fmt.Println("----------------------------------------")
	
	types := []string{"成長型好公司", "穩健型", "高配息", "撐配息", "吃老本", "一般"}
	for _, t := range types {
		if count, exists := stats[t]; exists {
			fmt.Printf("  %s: %d 支\n", t, count)
		}
	}
}
