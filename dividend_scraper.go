package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"os"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// 股票配息資料
type DividendData struct {
	Symbol              string  `json:"symbol"`
	Name                string  `json:"name"`
	CashDividend        float64 `json:"cash_dividend"`         // 現金股利
	StockDividend       float64 `json:"stock_dividend"`        // 股票股利
	TotalDividend       float64 `json:"total_dividend"`        // 合計股利
	ExDividendDate      string  `json:"ex_dividend_date"`      // 除權息日
	CurrentPrice        float64 `json:"current_price"`         // 目前股價
	Yield               float64 `json:"yield"`                 // 殖利率 (%)
	Consecutive10Years  bool    `json:"consecutive_10_years"`  // 10年連續配息
	DividendCount10Year int     `json:"dividend_count_10year"` // 10年配息次數
	Avg3Year            float64 `json:"avg_3year"`             // 3年平均股利
	Avg6Year            float64 `json:"avg_6year"`             // 6年平均股利
	Avg10Year           float64 `json:"avg_10year"`            // 10年平均股利
	// 新增：財務分析欄位
	EPS                 float64 `json:"eps"`                   // 最近一年 EPS
	EPSPrev             float64 `json:"eps_prev"`              // 前一年 EPS
	EPSTrend            string  `json:"eps_trend"`             // 成長/穩定/衰退
	PayoutRatio         float64 `json:"payout_ratio"`          // 配息率 (%)
	CompanyType         string  `json:"company_type"`          // 公司類型
	TypeEmoji           string  `json:"type_emoji"`            // 類型圖示
	TypeDescription     string  `json:"type_description"`      // 類型說明
	UpdateTime          string  `json:"update_time"`
}

// 爬取網站資料
func scrapeDividendData(symbols []string) (map[string]DividendData, error) {
	url := "https://stock.wespai.com/rate115"
	
	log.Printf("🔍 開始爬取 %s\n", url)
	
	// 發送 HTTP 請求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP 請求失敗: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP 狀態碼: %d", resp.StatusCode)
	}
	
	// 解析 HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML 解析失敗: %v", err)
	}
	
	// 建立股票代號 map 方便查詢
	symbolMap := make(map[string]bool)
	for _, s := range symbols {
		symbolMap[s] = true
	}
	
	// 儲存結果
	results := make(map[string]DividendData)
	
	// 遍歷 HTML，找到表格資料
	var parseTable func(*html.Node)
	parseTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			// 解析表格行
			data := extractRowData(n)
			if data != nil && symbolMap[data.Symbol] {
				results[data.Symbol] = *data
				log.Printf("✅ 找到 %s: 股利 %.2f, 殖利率 %.2f%%, 10年配息 %d 次\n", 
					data.Symbol, data.TotalDividend, data.Yield, data.DividendCount10Year)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parseTable(c)
		}
	}
	
	parseTable(doc)
	
	log.Printf("✅ 爬取完成，找到 %d 支股票資料\n", len(results))
	
	// 補充 Yahoo Finance 財務資料（EPS、配息率、公司類型）
	log.Printf("\n🔍 開始補充 Yahoo Finance 財務資料...\n")
	enrichWithYahooFinancial(results)
	
	return results, nil
}

// 從表格行中提取資料
func extractRowData(tr *html.Node) *DividendData {
	var cells []string
	
	// 提取所有 td 的文字內容
	var extractCells func(*html.Node)
	extractCells = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "td" {
			text := getTextContent(n)
			cells = append(cells, strings.TrimSpace(text))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractCells(c)
		}
	}
	
	extractCells(tr)
	
	// 檢查是否有足夠的欄位（網站格式：代號、名稱、現金股利、除息日、股票股利、除權日、目前股價、殖利率...）
	if len(cells) < 10 {
		return nil
	}
	
	// 解析資料
	symbol := cells[0]
	if symbol == "" || symbol == "代號" {
		return nil
	}
	
	data := &DividendData{
		Symbol:     symbol,
		Name:       cells[1],
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	
	// 解析現金股利
	if val, err := parseFloat(cells[2]); err == nil {
		data.CashDividend = val
	}
	
	// 解析股票股利
	if len(cells) > 4 {
		if val, err := parseFloat(cells[4]); err == nil {
			data.StockDividend = val
		}
	}
	
	// 合計股利
	data.TotalDividend = data.CashDividend + data.StockDividend
	
	// 除權息日
	if len(cells) > 3 {
		data.ExDividendDate = cells[3]
	}
	
	// 目前股價
	if len(cells) > 6 {
		if val, err := parseFloat(cells[6]); err == nil {
			data.CurrentPrice = val
		}
	}
	
	// 殖利率
	if len(cells) > 7 {
		if val, err := parseFloat(strings.TrimSuffix(cells[7], "%")); err == nil {
			data.Yield = val
		}
	}
	
	// 10年配息次數（在第15個欄位，索引14）
	if len(cells) > 14 {
		cellText := strings.TrimSpace(cells[14])
		if val, err := strconv.Atoi(cellText); err == nil && val >= 0 && val <= 10 {
			data.DividendCount10Year = val
			data.Consecutive10Years = (val == 10)
		}
	}
	
	// 如果第15欄失敗，嘗試其他可能的位置
	if data.DividendCount10Year == 0 {
		for i := 10; i < len(cells) && i < 20; i++ {
			cellText := strings.TrimSpace(cells[i])
			if val, err := strconv.Atoi(cellText); err == nil && val >= 0 && val <= 10 {
				data.DividendCount10Year = val
				data.Consecutive10Years = (val == 10)
				break
			}
		}
	}
	
	return data
}

// 獲取節點的文字內容
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getTextContent(c)
	}
	return text
}

// 解析浮點數（處理各種格式）
func parseFloat(s string) (float64, error) {
	// 移除逗號和空格
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.TrimSpace(s)
	
	if s == "" || s == "-" || s == " " {
		return 0, nil
	}
	
	return strconv.ParseFloat(s, 64)
}

// 主程式
func main() {
	// 讀取要查詢的股票清單
	symbols := []string{
		// 市值型 ETF
		"0050", "006208", "00632R", "00692", "00701", "00881", "00891", "00895", "00896",
		// 高股息 ETF
		"00919", "00929", "00918", "00878", "0056",
		// 權值股
		"2330", "2317", "2454", "2412", "2882", "2891", "2886", "2881", "2892", "2884",
		// 電子股
		"2303", "2308", "2382", "2357", "3711", "2327", "2379",
		// 傳產股
		"2002", "1301", "1303", "1326", "2105",
		// 中小型股
		"2353", "2324", "2618", "2838", "2812", "2887", "2851", "2890", "1102", "5876", "2816",
		// AI 相關
		"3443", "6510", "2395", "2356", "6669",
		// 電力相關
		"1101", "6506", "6411",
		// 通訊相關
		"3045", "4904", "2049", "3008",
		// 其他重要
		"6505", "2207", "2880",
		// 高年化報酬率
		"2409", "3034", "2301", "2408", "2344", "3481", "6176", "2371", "6414", "3661",
	}
	
	fmt.Println("🦈 鯊魚寶寶配息資料爬蟲")
	fmt.Println("📊 準備爬取 81 支股票的配息資料...")
	fmt.Println()
	
	// 爬取資料
	results, err := scrapeDividendData(symbols)
	if err != nil {
		log.Fatalf("❌ 爬取失敗: %v\n", err)
	}
	
	// 輸出 JSON
	wd, _ := os.Getwd()
	os.MkdirAll(filepath.Join(wd, "html"), 0755)
	outputFile := filepath.Join(wd, "html", "dividend_data.json")
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("❌ JSON 編碼失敗: %v\n", err)
	}
	
	err = ioutil.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		log.Fatalf("❌ 寫入檔案失敗: %v\n", err)
	}
	
	fmt.Printf("\n✅ 配息資料已儲存至: %s\n", outputFile)
	fmt.Printf("📊 共爬取 %d 支股票\n", len(results))
	
	// 統計
	var consecutive10 int
	var avgYield float64
	for _, data := range results {
		if data.Consecutive10Years {
			consecutive10++
		}
		avgYield += data.Yield
	}
	
	if len(results) > 0 {
		avgYield /= float64(len(results))
	}
	
	fmt.Printf("\n📈 統計資訊：\n")
	fmt.Printf("  10年連續配息: %d 支 (%.1f%%)\n", consecutive10, float64(consecutive10)/float64(len(results))*100)
	fmt.Printf("  平均殖利率: %.2f%%\n", avgYield)
	
	// 統計公司類型
	typeCount := make(map[string]int)
	for _, data := range results {
		if data.CompanyType != "" {
			typeCount[data.CompanyType]++
		}
	}
	
	if len(typeCount) > 0 {
		fmt.Printf("\n🏢 公司類型分布：\n")
		for typeStr, count := range typeCount {
			fmt.Printf("  %s: %d 支\n", typeStr, count)
		}
	}
}

// ========================================
// Yahoo Finance 財務資料整合
// ========================================

// 補充 Yahoo Finance 財務資料
func enrichWithYahooFinancial(results map[string]DividendData) {
	successCount := 0
	
	for symbol, data := range results {
		log.Printf("   抓取 %s (%s) 財務資料...", symbol, data.Name)
		
		financial, err := fetchYahooFinancial(symbol)
		if err != nil {
			log.Printf(" ❌ 失敗: %v\n", err)
			continue
		}
		
		// 更新 EPS 資料
		data.EPS = financial.EPS
		data.EPSPrev = financial.EPSPrev
		data.EPSTrend = financial.EPSTrend
		
		// 計算配息率（使用現金股利 / EPS）
		if data.EPS > 0 && data.CashDividend > 0 {
			data.PayoutRatio = (data.CashDividend / data.EPS) * 100
		} else {
			data.PayoutRatio = 0
		}
		
		// 分類公司類型
		data.CompanyType, data.TypeEmoji, data.TypeDescription = 
			classifyCompany(data.EPS, data.EPSPrev, data.PayoutRatio, data.EPSTrend)
		
		results[symbol] = data
		successCount++
		
		log.Printf(" ✅ %s (配息率 %.1f%%)\n", financial.CompanyType, financial.PayoutRatio)
		
		// 避免請求過快
		time.Sleep(500 * time.Millisecond)
	}
	
	log.Printf("\n✅ 財務資料補充完成：%d/%d 成功\n", successCount, len(results))
}

// 從 Yahoo Finance 抓取財務資料
func fetchYahooFinancial(symbol string) (*DividendData, error) {
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
	
	financial := &DividendData{
		Symbol: symbol,
	}
	
	// 解析年度 EPS
	epsRe := regexp.MustCompile(`"incomesY":\[([^\]]+)\]`)
	if matches := epsRe.FindStringSubmatch(html); len(matches) > 1 {
		epsData := matches[1]
		epsItems := regexp.MustCompile(`"eps":"([0-9.]+)"`).FindAllStringSubmatch(epsData, -1)
		
		if len(epsItems) >= 2 {
			financial.EPS, _ = strconv.ParseFloat(epsItems[0][1], 64)
			financial.EPSPrev, _ = strconv.ParseFloat(epsItems[1][1], 64)
		}
	}
	
	// 判斷 EPS 趨勢
	financial.EPSTrend = judgeEPSTrend(financial.EPS, financial.EPSPrev)
	
	// 使用已知的現金股利計算配息率（從撿股讚網站已取得）
	// 這裡不需要重新抓取股利，直接計算即可
	// 配息率會在外部計算
	
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
	
	// 4. 配息率 > 100% → 吃老本 ❌
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
