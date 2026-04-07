package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ========== TWSE API 資料結構 ==========

// K線日資料回應
type TWSeStockDayResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
	Notes  []string   `json:"notes"`
	Total  int        `json:"total"`
}

// 本益比、殖利率回應(使用 interface{} 因為資料格式不統一)
type TWSePERatioResponse struct {
	Stat   string        `json:"stat"`
	Date   string        `json:"date"`
	Title  string        `json:"title"`
	Fields []string      `json:"fields"`
	Data   [][]interface{} `json:"data"`
}

// 三大法人回應
type TWSeInstitutionalResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Title  string     `json:"title"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

// 股票代號查詢回應
type TWSeCodeQueryResponse struct {
	Query       string   `json:"query"`
	Suggestions []string `json:"suggestions"`
}

// ========== K線數據 ==========

// K線資料(統一格式)
type KLine struct {
	Date   string  `json:"date"`    // 民國日期 115/04/01
	Open   float64 `json:"open"`    // 開盤價
	High   float64 `json:"high"`    // 最高價
	Low    float64 `json:"low"`     // 最低價
	Close  float64 `json:"close"`   // 收盤價
	Volume int64   `json:"volume"`  // 成交股數
	Amount int64   `json:"amount"`  // 成交金額
	Change float64 `json:"change"`  // 漲跌價差
}

// 股票完整資料
type StockData struct {
	Symbol      string   `json:"symbol"`       // 股票代號
	Name        string   `json:"name"`         // 股票名稱
	CurrentPrice float64 `json:"current_price"` // 最新價格
	KLines      []KLine  `json:"klines"`       // K線數據(最近60天)
	PE          float64  `json:"pe"`           // 本益比
	PB          float64  `json:"pb"`           // 股價淨值比
	DividendYield float64 `json:"dividend_yield"` // 殖利率
	UpdateTime  string   `json:"update_time"`  // 更新時間
}

// ========== TWSE API 爬蟲 ==========

// 1. 查詢股票代號是否存在
func queryStockCode(symbol string) (string, error) {
	url := fmt.Sprintf("https://www.twse.com.tw/zh/api/codeQuery?query=%s", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("查詢失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result TWSeCodeQueryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Suggestions) == 0 {
		return "", fmt.Errorf("股票代號 %s 不存在", symbol)
	}
	
	// 檢查是否為「無符合之代碼或名稱」
	firstSuggestion := result.Suggestions[0]
	if strings.Contains(firstSuggestion, "無符合") || strings.Contains(firstSuggestion, "查無") {
		return "", fmt.Errorf("股票代號 %s 不存在或已下市", symbol)
	}
	
	// 解析 "2330\t台積電"
	parts := strings.Split(firstSuggestion, "\t")
	if len(parts) < 2 {
		// 嘗試直接使用第一個建議，可能只有代號
		if len(parts) == 1 {
			return parts[0], nil
		}
		return "", fmt.Errorf("無法解析股票名稱，原始資料: %s", firstSuggestion)
	}
	
	return parts[1], nil
}

// 2. 取得月 K 線資料(指定年月)
func fetchMonthlyKLines(symbol string, year int, month int) ([]KLine, error) {
	// TWSE 格式:20260401(西元年月日)
	dateStr := fmt.Sprintf("%d%02d01", year, month)

	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/STOCK_DAY?date=%s&stockNo=%s&response=json", dateStr, symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TWSeStockDayResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Stat != "OK" {
		return nil, fmt.Errorf("API 回應異常: %s", result.Stat)
	}

	// 解析 K 線數據
	klines := []KLine{}
	for _, row := range result.Data {
		if len(row) < 9 {
			continue
		}

		kline := KLine{
			Date:   row[0], // 民國日期 115/04/01
			Open:   parseFloat(row[3]),
			High:   parseFloat(row[4]),
			Low:    parseFloat(row[5]),
			Close:  parseFloat(row[6]),
			Change: parseFloat(row[7]),
			Volume: parseInt(row[1]),
			Amount: parseInt(row[2]),
		}

		klines = append(klines, kline)
	}

	return klines, nil
}

// 3. 取得最近 60 天 K 線(跨月)
func fetchRecentKLines(symbol string, days int) ([]KLine, error) {
	allKLines := []KLine{}

	now := time.Now()

	// 往前推 3 個月(確保拿到 60 天數據)
	for i := 0; i < 3; i++ {
		targetDate := now.AddDate(0, -i, 0)
		year := targetDate.Year()
		month := int(targetDate.Month())

		klines, err := fetchMonthlyKLines(symbol, year, month)
		if err != nil {
			log.Printf("⚠️ 取得 %d/%02d K線失敗: %v", year, month, err)
			continue
		}

		allKLines = append(allKLines, klines...)

		// 避免請求過快
		time.Sleep(500 * time.Millisecond)
	}

	// 只保留最近 N 天
	if len(allKLines) > days {
		allKLines = allKLines[len(allKLines)-days:]
	}

	return allKLines, nil
}

// 4. 取得本益比、殖利率(當日)
func fetchPERatio(symbol string) (float64, float64, float64, error) {
	// 使用昨天的日期(TWSE 資料會延遲一天)
	yesterday := time.Now().AddDate(0, 0, -1)
	dateStr := yesterday.Format("20060102")

	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/afterTrading/BWIBBU_d?date=%s&selectType=ALL&response=json", dateStr)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, 0, err
	}

	var result TWSePERatioResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, 0, err
	}

	// 查找指定股票
	for _, row := range result.Data {
		if len(row) < 7 {
			continue
		}

		// 轉成字串比對
		stockCode := fmt.Sprintf("%v", row[0])
		if stockCode == symbol {
			pe := parseFloat(fmt.Sprintf("%v", row[5]))      // 本益比
			pb := parseFloat(fmt.Sprintf("%v", row[6]))      // 股價淨值比
			dy := parseFloat(fmt.Sprintf("%v", row[3]))      // 殖利率
			return pe, pb, dy, nil
		}
	}

	return 0, 0, 0, fmt.Errorf("找不到股票 %s 的本益比資料", symbol)
}

// 5. 完整查詢單一股票
func fetchStockData(symbol string) (*StockData, error) {
	fmt.Printf("🔍 查詢股票: %s\n", symbol)

	// 1. 驗證股票代號
	name, err := queryStockCode(symbol)
	if err != nil {
		return nil, err
	}
	fmt.Printf("  ✅ 股票名稱: %s\n", name)

	// 2. 取得 K 線數據
	fmt.Printf("  📊 抓取 K 線數據...\n")
	klines, err := fetchRecentKLines(symbol, 60)
	if err != nil || len(klines) == 0 {
		return nil, fmt.Errorf("無法取得 K 線數據: %v", err)
	}
	fmt.Printf("  ✅ 成功取得 %d 天 K 線\n", len(klines))

	// 3. 取得本益比、殖利率
	fmt.Printf("  💰 抓取本益比、殖利率...\n")
	pe, pb, dy, err := fetchPERatio(symbol)
	if err != nil {
		log.Printf("  ⚠️ 本益比資料失敗: %v\n", err)
		// 不中斷,繼續處理
	} else {
		fmt.Printf("  ✅ 本益比: %.2f | 殖利率: %.2f%%\n", pe, dy)
	}

	// 4. 組合資料
	stock := &StockData{
		Symbol:        symbol,
		Name:          name,
		CurrentPrice:  klines[len(klines)-1].Close,
		KLines:        klines,
		PE:            pe,
		PB:            pb,
		DividendYield: dy,
		UpdateTime:    time.Now().Format("2006-01-02 15:04:05"),
	}

	return stock, nil
}

// ========== 輔助函數 ==========

// 解析浮點數(處理逗號)
func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	// 處理 "-" 或空值
	if s == "-" || s == "" || s == "N/A" {
		return 0
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// 解析整數(處理逗號)
func parseInt(s string) int64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)

	if s == "-" || s == "" {
		return 0
	}

	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// ========== 測試程式 ==========

func main() {
	fmt.Println("🦈 TWSE 爬蟲測試")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 測試查詢
	testSymbols := []string{"2330", "2812", "0050"}

	for _, symbol := range testSymbols {
		stock, err := fetchStockData(symbol)
		if err != nil {
			log.Printf("❌ 查詢失敗: %v\n", err)
			continue
		}

		fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📊 %s (%s)\n", stock.Name, stock.Symbol)
		fmt.Printf("💰 現價: %.2f\n", stock.CurrentPrice)
		fmt.Printf("📈 本益比: %.2f\n", stock.PE)
		fmt.Printf("📉 股價淨值比: %.2f\n", stock.PB)
		fmt.Printf("💎 殖利率: %.2f%%\n", stock.DividendYield)
		fmt.Printf("📊 K線數據: %d 天\n", len(stock.KLines))

		// 顯示最近 5 天
		fmt.Println("\n最近 5 天:")
		start := len(stock.KLines) - 5
		if start < 0 {
			start = 0
		}
		for _, k := range stock.KLines[start:] {
			fmt.Printf("  %s | 開:%.2f 高:%.2f 低:%.2f 收:%.2f | 量:%d\n",
				k.Date, k.Open, k.High, k.Low, k.Close, k.Volume)
		}

		fmt.Println("")
		time.Sleep(1 * time.Second)
	}

	fmt.Println("✅ 測試完成")
}
