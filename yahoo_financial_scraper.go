package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// 財務資料結構
type FinancialData struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	EPS             float64 `json:"eps"`              // 最近一年 EPS
	EPSPrev         float64 `json:"eps_prev"`         // 前一年 EPS
	EPSTrend        string  `json:"eps_trend"`        // 成長/穩定/衰退
	CashDividend    float64 `json:"cash_dividend"`    // 現金股利
	PayoutRatio     float64 `json:"payout_ratio"`     // 配息率 (%)
	CompanyType     string  `json:"company_type"`     // 公司類型
	TypeEmoji       string  `json:"type_emoji"`       // 類型圖示
	TypeDescription string  `json:"type_description"` // 類型說明
}

func main() {
	fmt.Println("🦈 鯊魚寶寶 Yahoo 財務資料爬蟲")
	fmt.Println("========================================")
	
	// 測試股票
	testSymbols := []string{"2330", "2353", "2881"}
	
	for _, symbol := range testSymbols {
		fmt.Printf("\n📊 抓取 %s 資料...\n", symbol)
		
		financial, err := fetchFinancialData(symbol)
		if err != nil {
			fmt.Printf("❌ 失敗: %v\n", err)
			continue
		}
		
		// 顯示結果
		fmt.Printf("✅ %s (%s)\n", financial.Name, financial.Symbol)
		fmt.Printf("   EPS (2025): %.2f\n", financial.EPS)
		fmt.Printf("   EPS (2024): %.2f\n", financial.EPSPrev)
		fmt.Printf("   趨勢: %s\n", financial.EPSTrend)
		fmt.Printf("   現金股利: %.2f\n", financial.CashDividend)
		fmt.Printf("   配息率: %.1f%%\n", financial.PayoutRatio)
		fmt.Printf("   %s %s\n", financial.TypeEmoji, financial.CompanyType)
		fmt.Printf("   說明: %s\n", financial.TypeDescription)
	}
	
	fmt.Println("\n========================================")
	fmt.Println("🦈 測試完成")
}

// 從 Yahoo 抓取財務資料
func fetchFinancialData(symbol string) (*FinancialData, error) {
	url := fmt.Sprintf("https://tw.stock.yahoo.com/quote/%s.TW/dividend", symbol)
	
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	
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
	
	html := string(body)
	
	// 解析 JSON 資料（從 HTML 中提取）
	financial := &FinancialData{
		Symbol: symbol,
	}
	
	// 解析股票名稱
	nameRe := regexp.MustCompile(`"symbolName":"([^"]+)"`)
	if matches := nameRe.FindStringSubmatch(html); len(matches) > 1 {
		financial.Name = matches[1]
	}
	
	// 解析年度 EPS
	epsRe := regexp.MustCompile(`"incomesY":\[([^\]]+)\]`)
	if matches := epsRe.FindStringSubmatch(html); len(matches) > 1 {
		// 解析最近兩年的 EPS
		epsData := matches[1]
		epsItems := regexp.MustCompile(`"eps":"([0-9.]+)"`).FindAllStringSubmatch(epsData, -1)
		
		if len(epsItems) >= 2 {
			financial.EPS, _ = strconv.ParseFloat(epsItems[0][1], 64)
			financial.EPSPrev, _ = strconv.ParseFloat(epsItems[1][1], 64)
		}
	}
	
	// 解析現金股利（取最新一年）
	divRe := regexp.MustCompile(`"cash":"([0-9.]+)"`)
	if matches := divRe.FindStringSubmatch(html); len(matches) > 1 {
		financial.CashDividend, _ = strconv.ParseFloat(matches[1], 64)
	}
	
	// 如果找不到，嘗試從 dividends 陣列中取
	if financial.CashDividend == 0 {
		divArrayRe := regexp.MustCompile(`"dividends":\[([^\]]+)\]`)
		if matches := divArrayRe.FindStringSubmatch(html); len(matches) > 1 {
			// 找第一個現金股利
			cashRe := regexp.MustCompile(`"cash":"([0-9.]+)"`)
			if cashMatches := cashRe.FindStringSubmatch(matches[1]); len(cashMatches) > 1 {
				financial.CashDividend, _ = strconv.ParseFloat(cashMatches[1], 64)
			}
		}
	}
	
	// 計算配息率
	if financial.EPS > 0 {
		financial.PayoutRatio = (financial.CashDividend / financial.EPS) * 100
	}
	
	// 判斷 EPS 趨勢
	financial.EPSTrend = judgeEPSTrend(financial.EPS, financial.EPSPrev)
	
	// 判斷公司類型
	financial.CompanyType, financial.TypeEmoji, financial.TypeDescription = 
		classifyCompany(financial.EPS, financial.EPSPrev, financial.PayoutRatio, financial.EPSTrend)
	
	return financial, nil
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
