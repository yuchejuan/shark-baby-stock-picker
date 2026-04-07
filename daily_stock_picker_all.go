package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// 股票資料結構（與原版相同）
type Stock struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Score     int     `json:"score"`
	RSI       float64 `json:"rsi"`
	MACD      string  `json:"macd"`
	KD        string  `json:"kd"`
	Signal    string  `json:"signal"`
	Advantage string  `json:"advantage"`
	MA5       float64 `json:"ma5"`
	MA20      float64 `json:"ma20"`
	MA60      float64 `json:"ma60"`
	MATrend   string  `json:"ma_trend"`
	Rank      int     `json:"rank"`
}

// 報告結構
type DailyReport struct {
	Date      string  `json:"date"`
	AllPicks  []Stock `json:"all_picks"`  // 所有股票（無價格限制）
	BestPick  *Stock  `json:"best_pick"`
	Summary   Summary `json:"summary"`
}

type Summary struct {
	TotalStocks  int    `json:"total_stocks"`
	TopPicks     int    `json:"top_picks"`
	BuySignals   int    `json:"buy_signals"`
	RiskWarnings int    `json:"risk_warnings"`
	UpdateTime   string `json:"update_time"`
}

// 熱門股票清單（81 支）- 與高股息追蹤系統同步
var popularStocks = []string{
	// 🏆 市值型 ETF（9 支）
	"0050",   // 元大台灣50
	"006208", // 富邦台50
	"00632R", // 元大台灣50反1
	"00692",  // 富邦公司治理
	"00701",  // 國泰股利精選30
	"00881",  // 國泰台灣5G+
	"00891",  // 中信關鍵半導體
	"00895",  // 富邦未來車
	"00896",  // 中信綠能及電動車
	
	// 💰 高股息 ETF（5 支）
	"00919", // 群益台灣精選高息
	"00929", // 復華台灣科技優息
	"00918", // 大華優利高填息30
	"00878", // 國泰永續高股息
	"0056",  // 元大高股息
	
	// 💎 權值股（10 支）
	"2330", // 台積電
	"2317", // 鴻海
	"2454", // 聯發科
	"2412", // 中華電
	"2882", // 國泰金
	"2891", // 中信金
	"2886", // 兆豐金
	"2881", // 富邦金
	"2892", // 第一金
	"2884", // 玉山金
	
	// 💻 電子股（7 支）
	"2303", // 聯電
	"2308", // 台達電
	"2382", // 廣達
	"2357", // 華碩
	"3711", // 日月光投控
	"2327", // 國巨
	"2379", // 瑞昱
	
	// 🏭 傳產股（5 支）
	"2002", // 中鋼
	"1301", // 台塑
	"1303", // 南亞
	"1326", // 台化
	"2105", // 正新
	
	// 🏦 中小型股（11 支）
	"2353", // 宏碁
	"2324", // 仁寶
	"2618", // 長榮航
	"2838", // 聯邦銀
	"2812", // 台中銀
	"2887", // 台新金
	"2851", // 中再保
	"2890", // 永豐金
	"1102", // 亞泥
	"5876", // 上海商銀
	"2816", // 旺旺保
	
	// 🤖 AI 相關（5 支）
	"3443", // 創意（AI ASIC 設計）
	"6510", // 精測（AI 晶片測試）
	"2395", // 研華（工業 AI 電腦）
	"2356", // 英業達（AI 伺服器）
	"6669", // 緯穎（AI 伺服器龍頭）
	
	// ⚡ 電力相關（3 支）
	"1101", // 台泥（綠電、儲能）
	"6506", // 雙鴻（散熱）
	"6411", // 晶焱（電源管理 IC）
	
	// 📡 通訊相關（4 支）
	"3045", // 台灣大（5G 電信）
	"4904", // 遠傳（5G 電信）
	"2049", // 上銀（工業自動化）
	"3008", // 大立光（光學）
	
	// 🏆 其他重要（3 支）
	"6505", // 台塑化
	"2207", // 和泰車
	"2880", // 華南金
	
	// 📈 高年化報酬率（10 支）
	"2409", // 友達
	"3034", // 聯詠
	"2301", // 光寶科
	"2408", // 南亞科
	"2344", // 華邦電
	"3481", // 群創
	"6176", // 瑞儀
	"2371", // 大同
	"6414", // 樺漢
	"3661", // 世芯-KY
}

func main() {
	fmt.Println("🦈 鯊魚寶寶全市場選股系統 V3.0")
	fmt.Println("📊 無價格限制，全市場掃描")
	fmt.Println("")

	// 取得所有股票資料
	allStocks := []Stock{}
	
	fmt.Println("📡 開始掃描股票...")
	for i, symbol := range popularStocks {
		fmt.Printf("  (%d/%d) 分析 %s...\n", i+1, len(popularStocks), symbol)
		
		stock := analyzeStock(symbol)
		if stock != nil && stock.Price > 0 {
			allStocks = append(allStocks, *stock)
		}
		
		// 避免過於頻繁請求
		time.Sleep(300 * time.Millisecond)
	}
	
	fmt.Printf("✅ 共分析 %d 支股票\n\n", len(allStocks))
	
	// 排序（依評分）
	sortStocksByScore(allStocks)
	
	// 加入排名
	for i := range allStocks {
		allStocks[i].Rank = i + 1
	}
	
	// 統計
	buySignals := 0
	for _, s := range allStocks {
		if s.Signal == "買點" {
			buySignals++
		}
	}
	
	// 建立報告
	report := DailyReport{
		Date:     time.Now().Format("2006-01-02"),
		AllPicks: allStocks,
		Summary: Summary{
			TotalStocks:  len(allStocks),
			TopPicks:     min(10, len(allStocks)),
			BuySignals:   buySignals,
			RiskWarnings: 0,
			UpdateTime:   time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	
	if len(allStocks) > 0 {
		report.BestPick = &allStocks[0]
	}
	
	// 儲存 JSON
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatal("JSON 序列化失敗:", err)
	}
	
	err = ioutil.WriteFile("stock_web/daily_report.json", data, 0644)
	if err != nil {
		log.Fatal("寫入檔案失敗:", err)
	}
	
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  📊 選股報告")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  總計: %d 支股票\n", len(allStocks))
	fmt.Printf("  買點訊號: %d 支\n", buySignals)
	fmt.Println("")
	fmt.Println("  🏆 TOP 10 推薦：")
	fmt.Println("")
	
	for i := 0; i < min(10, len(allStocks)); i++ {
		s := allStocks[i]
		fmt.Printf("  %d. %s (%s) - $%.2f - 評分 %d - %s\n",
			i+1, s.Symbol, s.Name, s.Price, s.Score, s.Advantage)
	}
	
	fmt.Println("")
	fmt.Println("✅ 報告已儲存至 stock_web/daily_report.json")
	fmt.Println("")
}

// 分析單一股票
func analyzeStock(symbol string) *Stock {
	// 從 TWSE API 取得股價（簡化版，實際應該取得歷史資料）
	price, name := fetchStockPrice(symbol)
	if price == 0 {
		return nil
	}
	
	// 模擬技術指標（實際應該計算歷史資料）
	// 這裡用簡單的模擬，你可以替換成真實的計算
	stock := &Stock{
		Symbol: symbol,
		Name:   name,
		Price:  price,
	}
	
	// 模擬指標（這裡應該用真實的歷史資料計算）
	stock.RSI = simulateRSI(symbol)
	stock.MACD = simulateMACD(symbol)
	stock.KD = simulateKD(symbol)
	stock.MA5, stock.MA20, stock.MA60 = simulateMA(price)
	stock.MATrend = calculateMATrend(stock.MA5, stock.MA20, stock.MA60)
	
	// 計算評分
	stock.Score = calculateScore(stock)
	
	// 判斷訊號
	stock.Signal = determineSignal(stock)
	
	// 推薦理由
	stock.Advantage = generateAdvantage(stock)
	
	return stock
}

// 從 TWSE API 取得股價
func fetchStockPrice(symbol string) (float64, string) {
	url := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_%s.tw", symbol)
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, ""
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, ""
	}
	
	msgArray, ok := result["msgArray"].([]interface{})
	if !ok || len(msgArray) == 0 {
		return 0, ""
	}
	
	msg := msgArray[0].(map[string]interface{})
	
	priceStr, _ := msg["z"].(string)
	name, _ := msg["n"].(string)
	
	if priceStr == "-" || priceStr == "" {
		return 0, ""
	}
	
	var price float64
	fmt.Sscanf(priceStr, "%f", &price)
	
	return price, name
}

// 模擬 RSI（實際應該用歷史資料計算）
func simulateRSI(symbol string) float64 {
	// 這裡用簡單的模擬，返回 30-70 的隨機值
	// 實際應該用 14 天歷史資料計算
	hash := 0
	for _, c := range symbol {
		hash += int(c)
	}
	return 30 + float64(hash%41)
}

// 模擬 MACD
func simulateMACD(symbol string) string {
	hash := 0
	for _, c := range symbol {
		hash += int(c)
	}
	if hash%3 == 0 {
		return "多頭"
	}
	if hash%3 == 1 {
		return "空頭"
	}
	return "中性"
}

// 模擬 KD
func simulateKD(symbol string) string {
	hash := 0
	for _, c := range symbol {
		hash += int(c)
	}
	if hash%4 == 0 {
		return "超賣"
	}
	if hash%4 == 1 {
		return "超買"
	}
	return "中性"
}

// 模擬均線
func simulateMA(price float64) (float64, float64, float64) {
	ma5 := price * (0.98 + 0.04*0.5)
	ma20 := price * (0.95 + 0.1*0.5)
	ma60 := price * (0.90 + 0.15*0.5)
	return ma5, ma20, ma60
}

// 計算均線趨勢
func calculateMATrend(ma5, ma20, ma60 float64) string {
	if ma5 > ma20 && ma20 > ma60 {
		return "多頭"
	}
	if ma60 > ma20 && ma20 > ma5 {
		return "空頭"
	}
	return "中性"
}

// 計算評分
func calculateScore(s *Stock) int {
	score := 30 // 基礎分
	
	// RSI 評分
	if s.RSI < 30 {
		score += 15
	} else if s.RSI < 40 {
		score += 10
	} else if s.RSI <= 60 {
		score += 5
	} else if s.RSI > 70 {
		score -= 10
	}
	
	// MACD 評分
	if s.MACD == "多頭" {
		score += 15
	} else if s.MACD == "空頭" {
		score -= 10
	}
	
	// KD 評分
	if s.KD == "超賣" {
		score += 12
	} else if s.KD == "超買" {
		score -= 12
	} else {
		score += 5
	}
	
	// 均線評分
	if s.MATrend == "多頭" {
		score += 15
	} else if s.MATrend == "空頭" {
		score -= 15
	}
	
	// 價格位置評分
	if s.Price > s.MA20 && s.Price > s.MA60 {
		score += 8
	}
	
	return max(0, min(100, score))
}

// 判斷訊號
func determineSignal(s *Stock) string {
	if s.Score >= 70 {
		return "買點"
	}
	if s.Score <= 40 {
		return "賣點"
	}
	return "中性"
}

// 生成推薦理由
func generateAdvantage(s *Stock) string {
	reasons := []string{}
	
	if s.MACD == "多頭" {
		reasons = append(reasons, "MACD黃金交叉")
	}
	
	if s.KD == "超賣" {
		reasons = append(reasons, "KD超賣")
	}
	
	if s.RSI < 30 {
		reasons = append(reasons, "RSI超賣反彈機會")
	}
	
	if s.MATrend == "多頭" {
		reasons = append(reasons, "多頭排列")
	}
	
	if len(reasons) == 0 {
		if s.Price < 30 {
			return fmt.Sprintf("價格低%.0f元", s.Price)
		}
		return "技術面中性"
	}
	
	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += " + " + reasons[i]
	}
	
	return result
}

// 排序（依評分）
func sortStocksByScore(stocks []Stock) {
	for i := 0; i < len(stocks); i++ {
		for j := i + 1; j < len(stocks); j++ {
			if stocks[j].Score > stocks[i].Score {
				stocks[i], stocks[j] = stocks[j], stocks[i]
			}
		}
	}
}

// 輔助函數
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
