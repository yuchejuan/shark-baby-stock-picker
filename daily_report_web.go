package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ========== 網頁展示用結構（V3增強版）==========

type WebStockPick struct {
	Rank           int      `json:"rank"`
	Symbol         string   `json:"symbol"`
	Name           string   `json:"name"`
	Price          float64  `json:"price"`
	Score          float64  `json:"score"`
	MarketState    string   `json:"market_state"`
	RiskLevel      string   `json:"risk_level"`
	RiskScore      int      `json:"risk_score"`
	
	// 詳細推薦理由
	Reasons        []string `json:"reasons"`
	
	// 買賣點建議
	BuyPoint       string   `json:"buy_point"`
	SellPoint      string   `json:"sell_point"`
	StopLoss       float64  `json:"stop_loss"`
	TargetPrice    float64  `json:"target_price"`
	
	// 技術指標詳情
	Indicators     WebIndicators `json:"indicators"`
	
	// 圖表資料
	ChartData      WebChartData  `json:"chart_data"`
	
	// 撿股讚整合 (V3.0 新增)
	DataSources         []string `json:"data_sources"`          // ["Yahoo", "TWSE", "Wespai"]
	InstitutionalSignal string   `json:"institutional_signal"`  // "強力買進", "買進", "中性", "賣出"
	
	// 除權息資訊 (V3.1 新增)
	DividendInfo  *DividendInfo `json:"dividend_info,omitempty"` // 除權息詳細資料
	DividendAlert string        `json:"dividend_alert"`          // 除息提醒訊息
}

type WebIndicators struct {
	RSI        float64 `json:"rsi"`
	RSIStatus  string  `json:"rsi_status"`
	
	MACD_DIF   float64 `json:"macd_dif"`
	MACD_DEA   float64 `json:"macd_dea"`
	MACD_OSC   float64 `json:"macd_osc"`
	MACDStatus string  `json:"macd_status"`
	
	KD_K       float64 `json:"kd_k"`
	KD_D       float64 `json:"kd_d"`
	KDStatus   string  `json:"kd_status"`
	
	MA5        float64 `json:"ma5"`
	MA10       float64 `json:"ma10"`
	MA20       float64 `json:"ma20"`
	MA60       float64 `json:"ma60"`
	MAStatus   string  `json:"ma_status"`
	
	BB_Upper   float64 `json:"bb_upper"`
	BB_Middle  float64 `json:"bb_middle"`
	BB_Lower   float64 `json:"bb_lower"`
	BBPosition string  `json:"bb_position"`
}

type WebChartData struct {
	Labels     []string  `json:"labels"`      // 日期
	Prices     []float64 `json:"prices"`      // 收盤價
	MA5        []float64 `json:"ma5"`         // 5日均線
	MA20       []float64 `json:"ma20"`        // 20日均線
	BBUpper    []float64 `json:"bb_upper"`    // 布林上軌
	BBLower    []float64 `json:"bb_lower"`    // 布林下軌
	Volumes    []int64   `json:"volumes"`     // 成交量
}

type WebDailyReport struct {
	Date           string         `json:"date"`
	GeneratedAt    string         `json:"generated_at"`
	Picks20_30     []WebStockPick `json:"picks_20_30"`
	Picks30_40     []WebStockPick `json:"picks_30_40"`
	Picks40_50     []WebStockPick `json:"picks_40_50"`
	Picks50_60     []WebStockPick `json:"picks_50_60"`
	BestPick       WebStockPick   `json:"best_pick"`
	Summary        WebSummary     `json:"summary"`
	BacktestStats  *BacktestStats `json:"backtest_stats,omitempty"`
}

type WebSummary struct {
	TotalStocks     int     `json:"total_stocks"`
	TopPicks        int     `json:"top_picks"`
	BullishCount    int     `json:"bullish_count"`
	BearishCount    int     `json:"bearish_count"`
	SidewaysCount   int     `json:"sideways_count"`
	LowRiskCount    int     `json:"low_risk_count"`
	HighRiskCount   int     `json:"high_risk_count"`
	AvgScore        float64 `json:"avg_score"`
	UpdateTime      string  `json:"update_time"`
}

// ========== 轉換函數 ==========

// 從基礎資料產生網頁用資料
func ConvertToWebPick(symbol, name string, klines []KLineData, indicators TechnicalIndicators, 
	score float64, marketState MarketState, riskReport RiskReport) WebStockPick {
	
	currentPrice := klines[len(klines)-1].Close
	
	// 產生推薦理由
	reasons := generateReasons(indicators, marketState, riskReport)
	
	// 計算買賣點
	buyPoint, sellPoint := calculateTradingPoints(currentPrice, indicators)
	stopLoss := calculateStopLoss(currentPrice, indicators)
	targetPrice := calculateTargetPrice(currentPrice, indicators)
	
	// 產生圖表資料
	chartData := generateChartData(klines, indicators)
	
	// 產生指標狀態說明
	webIndicators := WebIndicators{
		RSI:        indicators.RSI,
		RSIStatus:  getRSIStatus(indicators.RSI),
		MACD_DIF:   indicators.MACD.DIF,
		MACD_DEA:   indicators.MACD.DEA,
		MACD_OSC:   indicators.MACD.OSC,
		MACDStatus: getMACDStatus(indicators.MACD),
		KD_K:       indicators.KD.K,
		KD_D:       indicators.KD.D,
		KDStatus:   getKDStatus(indicators.KD),
		MA5:        indicators.MA5,
		MA10:       indicators.MA10,
		MA20:       indicators.MA20,
		MA60:       indicators.MA60,
		MAStatus:   getMAStatus(indicators),
		BB_Upper:   indicators.BB.Upper,
		BB_Middle:  indicators.BB.Middle,
		BB_Lower:   indicators.BB.Lower,
		BBPosition: getBBPosition(currentPrice, indicators.BB),
	}
	
	return WebStockPick{
		Symbol:      symbol,
		Name:        name,
		Price:       currentPrice,
		Score:       score,
		MarketState: marketState.String(),
		RiskLevel:   riskReport.RiskLevelName,
		RiskScore:   riskReport.TotalScore,
		Reasons:     reasons,
		BuyPoint:    buyPoint,
		SellPoint:   sellPoint,
		StopLoss:    stopLoss,
		TargetPrice: targetPrice,
		Indicators:  webIndicators,
		ChartData:   chartData,
	}
}

// 產生推薦理由
func generateReasons(indicators TechnicalIndicators, marketState MarketState, riskReport RiskReport) []string {
	reasons := []string{}
	
	// 市場狀態
	reasons = append(reasons, fmt.Sprintf("市場狀態: %s", marketState.String()))
	
	// RSI狀態
	if indicators.RSI < 30 {
		reasons = append(reasons, "RSI超賣，反彈機會高")
	} else if indicators.RSI < 40 {
		reasons = append(reasons, "RSI偏低，進場時機佳")
	}
	
	// MACD狀態
	if indicators.MACD.DIF > indicators.MACD.DEA && indicators.MACD.OSC > 0 {
		reasons = append(reasons, "MACD黃金交叉，趨勢向上")
	}
	
	// KD狀態
	if indicators.KD.K < 20 {
		reasons = append(reasons, "KD超賣，即將反彈")
	} else if indicators.KD.K > indicators.KD.D && indicators.KD.K < 50 {
		reasons = append(reasons, "KD低檔黃金交叉")
	}
	
	// 均線狀態
	if indicators.MA5 > indicators.MA10 && indicators.MA10 > indicators.MA20 {
		reasons = append(reasons, "均線多頭排列，趨勢強勁")
	}
	
	// 布林通道
	bbPos := (indicators.BB.Middle - indicators.BB.Lower) / (indicators.BB.Upper - indicators.BB.Lower)
	if bbPos < 0.3 {
		reasons = append(reasons, "價格接近布林下軌，支撐強")
	}
	
	// 風險評估
	if riskReport.RiskLevel == RiskLow {
		reasons = append(reasons, "風險評估: 低風險，適合進場")
	}
	
	return reasons
}

// 計算買賣點
func calculateTradingPoints(currentPrice float64, indicators TechnicalIndicators) (string, string) {
	buyPoint := ""
	sellPoint := ""
	
	// 買點建議
	if currentPrice < indicators.MA20 {
		buyPoint = fmt.Sprintf("建議分批進場：%.2f (當前) - %.2f (MA20附近)", 
			currentPrice*0.98, indicators.MA20*0.99)
	} else {
		buyPoint = fmt.Sprintf("回測支撐再進場：%.2f (MA5) - %.2f (MA10)", 
			indicators.MA5, indicators.MA10)
	}
	
	// 賣點建議
	if indicators.BB.Upper > 0 {
		sellPoint = fmt.Sprintf("壓力位：%.2f (布林上軌) / %.2f (+5%%)", 
			indicators.BB.Upper, currentPrice*1.05)
	} else {
		sellPoint = fmt.Sprintf("壓力位：%.2f (+5%%)", currentPrice*1.05)
	}
	
	return buyPoint, sellPoint
}

// 計算停損點
func calculateStopLoss(currentPrice float64, indicators TechnicalIndicators) float64 {
	// 以MA20或近期低點-3%為停損
	stopLoss := indicators.MA20 * 0.97
	if currentPrice*0.93 > stopLoss {
		stopLoss = currentPrice * 0.93 // 最多損失7%
	}
	return stopLoss
}

// 計算目標價
func calculateTargetPrice(currentPrice float64, indicators TechnicalIndicators) float64 {
	// 目標價設在布林上軌或+10%
	target := indicators.BB.Upper
	if target < currentPrice*1.1 {
		target = currentPrice * 1.1
	}
	return target
}

// 產生圖表資料（最近30天）
func generateChartData(klines []KLineData, indicators TechnicalIndicators) WebChartData {
	chartData := WebChartData{
		Labels:  []string{},
		Prices:  []float64{},
		MA5:     []float64{},
		MA20:    []float64{},
		BBUpper: []float64{},
		BBLower: []float64{},
		Volumes: []int64{},
	}
	
	// 取最近30天資料
	startIdx := len(klines) - 30
	if startIdx < 0 {
		startIdx = 0
	}
	
	for i := startIdx; i < len(klines); i++ {
		chartData.Labels = append(chartData.Labels, klines[i].Date[5:]) // MM-DD
		chartData.Prices = append(chartData.Prices, klines[i].Close)
		chartData.Volumes = append(chartData.Volumes, klines[i].Volume)
		
		// 計算當日均線（簡化版）
		closes := ExtractClosePrices(klines[:i+1])
		if len(closes) >= 5 {
			chartData.MA5 = append(chartData.MA5, CalculateMA(closes, 5))
		}
		if len(closes) >= 20 {
			chartData.MA20 = append(chartData.MA20, CalculateMA(closes, 20))
			bb := CalculateBollingerBands(closes, 20, 2.0)
			chartData.BBUpper = append(chartData.BBUpper, bb.Upper)
			chartData.BBLower = append(chartData.BBLower, bb.Lower)
		}
	}
	
	return chartData
}

// ========== 指標狀態描述 ==========

func getRSIStatus(rsi float64) string {
	if rsi < 20 {
		return "嚴重超賣"
	} else if rsi < 30 {
		return "超賣"
	} else if rsi < 40 {
		return "偏低"
	} else if rsi < 60 {
		return "中性"
	} else if rsi < 70 {
		return "偏高"
	} else if rsi < 80 {
		return "超買"
	}
	return "嚴重超買"
}

func getMACDStatus(macd MACDData) string {
	if macd.DIF > macd.DEA {
		if macd.OSC > 0 {
			return "黃金交叉"
		}
		return "轉強中"
	} else {
		if macd.OSC < 0 {
			return "死亡交叉"
		}
		return "轉弱中"
	}
}

func getKDStatus(kd KDData) string {
	if kd.K < 20 && kd.D < 20 {
		return "超賣區"
	} else if kd.K > 80 && kd.D > 80 {
		return "超買區"
	} else if kd.K > kd.D {
		return "K>D 多方"
	} else {
		return "K<D 空方"
	}
}

func getMAStatus(indicators TechnicalIndicators) string {
	if indicators.MA5 > indicators.MA10 && indicators.MA10 > indicators.MA20 && indicators.MA20 > indicators.MA60 {
		return "完美多頭排列"
	} else if indicators.MA5 > indicators.MA20 {
		return "多頭排列"
	} else if indicators.MA5 < indicators.MA10 && indicators.MA10 < indicators.MA20 && indicators.MA20 < indicators.MA60 {
		return "完美空頭排列"
	} else if indicators.MA5 < indicators.MA20 {
		return "空頭排列"
	}
	return "盤整"
}

func getBBPosition(price float64, bb BollingerBands) string {
	if bb.Upper == bb.Lower {
		return "中性"
	}
	position := (price - bb.Lower) / (bb.Upper - bb.Lower) * 100
	
	if position < 10 {
		return "觸及下軌"
	} else if position < 30 {
		return "接近下軌"
	} else if position < 70 {
		return "中間區域"
	} else if position < 90 {
		return "接近上軌"
	}
	return "觸及上軌"
}

// ========== 產生網頁報告 ==========

func GenerateWebReport(picks []WebStockPick, backtestStats *BacktestStats) (*WebDailyReport, error) {
	now := time.Now()
	
	// 分類股票
	picks20_30 := []WebStockPick{}
	picks30_40 := []WebStockPick{}
	picks40_50 := []WebStockPick{}
	var bestPick WebStockPick
	maxScore := 0.0
	
	bullishCount := 0
	bearishCount := 0
	sidewaysCount := 0
	lowRiskCount := 0
	highRiskCount := 0
	totalScore := 0.0
	
	for i, pick := range picks {
		pick.Rank = i + 1
		
		// 統計
		switch pick.MarketState {
		case "多頭市場":
			bullishCount++
		case "空頭市場":
			bearishCount++
		case "盤整市場":
			sidewaysCount++
		}
		
		if pick.RiskScore < 10 {
			lowRiskCount++
		} else if pick.RiskScore >= 20 {
			highRiskCount++
		}
		
		totalScore += pick.Score
		
		// 找最佳推薦
		if pick.Score > maxScore {
			maxScore = pick.Score
			bestPick = pick
		}
		
		// 分類
		if pick.Price >= 20 && pick.Price < 30 {
			picks20_30 = append(picks20_30, pick)
		} else if pick.Price >= 30 && pick.Price < 40 {
			picks30_40 = append(picks30_40, pick)
		} else if pick.Price >= 40 && pick.Price < 50 {
			picks40_50 = append(picks40_50, pick)
		}
	}
	
	avgScore := 0.0
	if len(picks) > 0 {
		avgScore = totalScore / float64(len(picks))
	}
	
	summary := WebSummary{
		TotalStocks:   len(picks),
		TopPicks:      len(picks),
		BullishCount:  bullishCount,
		BearishCount:  bearishCount,
		SidewaysCount: sidewaysCount,
		LowRiskCount:  lowRiskCount,
		HighRiskCount: highRiskCount,
		AvgScore:      avgScore,
		UpdateTime:    now.Format("2006-01-02 15:04:05"),
	}
	
	report := &WebDailyReport{
		Date:          now.Format("2006-01-02"),
		GeneratedAt:   now.Format("2006-01-02 15:04:05"),
		Picks20_30:    picks20_30,
		Picks30_40:    picks30_40,
		Picks40_50:    picks40_50,
		BestPick:      bestPick,
		Summary:       summary,
		BacktestStats: backtestStats,
	}
	
	return report, nil
}

// 儲存網頁報告
func SaveWebReport(report *WebDailyReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// ========== 測試函數 ==========

func TestWebReport() {
	fmt.Println("🦈 網頁展示優化測試")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// 模擬一筆資料
	testSymbol := "2330"
	fmt.Printf("\n📊 測試股票: %s\n", testSymbol)
	
	klines, err := FetchHistoricalData(testSymbol)
	if err != nil {
		fmt.Printf("❌ 錯誤: %v\n", err)
		return
	}
	
	currentPrice := klines[len(klines)-1].Close
	closes := ExtractClosePrices(klines)
	highs := ExtractHighPrices(klines)
	lows := ExtractLowPrices(klines)
	volumes := ExtractVolumes(klines)
	
	indicators := TechnicalIndicators{
		RSI:    CalculateRSI(closes, 14),
		MACD:   CalculateMACD(closes),
		KD:     CalculateKD(highs, lows, closes, 9),
		BB:     CalculateBollingerBands(closes, 20, 2.0),
		MA5:    CalculateMA(closes, 5),
		MA10:   CalculateMA(closes, 10),
		MA20:   CalculateMA(closes, 20),
		MA60:   CalculateMA(closes, 60),
		Volume: volumes[len(volumes)-1],
	}
	
	marketState := DetectMarketState(klines, currentPrice)
	riskReport := AssessRisk(testSymbol, "台積電", klines, indicators, currentPrice)
	weights := GetDynamicWeights(marketState)
	score := CalculateDynamicScore(indicators, weights, marketState)
	
	webPick := ConvertToWebPick(testSymbol, "台積電", klines, indicators, score, marketState, riskReport)
	
	fmt.Printf("\n📈 網頁展示資料:\n")
	fmt.Printf("   股票: %s %s\n", webPick.Symbol, webPick.Name)
	fmt.Printf("   價格: %.2f\n", webPick.Price)
	fmt.Printf("   評分: %.2f\n", webPick.Score)
	fmt.Printf("   市場狀態: %s\n", webPick.MarketState)
	fmt.Printf("   風險等級: %s (分數: %d)\n", webPick.RiskLevel, webPick.RiskScore)
	
	fmt.Printf("\n💡 推薦理由:\n")
	for i, reason := range webPick.Reasons {
		fmt.Printf("   %d. %s\n", i+1, reason)
	}
	
	fmt.Printf("\n🎯 買賣建議:\n")
	fmt.Printf("   買點: %s\n", webPick.BuyPoint)
	fmt.Printf("   賣點: %s\n", webPick.SellPoint)
	fmt.Printf("   停損: %.2f\n", webPick.StopLoss)
	fmt.Printf("   目標價: %.2f\n", webPick.TargetPrice)
	
	fmt.Printf("\n📊 技術指標:\n")
	fmt.Printf("   RSI: %.2f (%s)\n", webPick.Indicators.RSI, webPick.Indicators.RSIStatus)
	fmt.Printf("   MACD: %s\n", webPick.Indicators.MACDStatus)
	fmt.Printf("   KD: %s\n", webPick.Indicators.KDStatus)
	fmt.Printf("   均線: %s\n", webPick.Indicators.MAStatus)
	fmt.Printf("   布林: %s\n", webPick.Indicators.BBPosition)
	
	// 儲存報告
	report, _ := GenerateWebReport([]WebStockPick{webPick}, nil)
	wd, _ := os.Getwd()
	filename := filepath.Join(wd, "html", "daily_report.json")
	os.MkdirAll(filepath.Join(wd, "html"), 0755)
	err = SaveWebReport(report, filename)
	if err != nil {
		fmt.Printf("\n❌ 儲存失敗: %v\n", err)
		return
	}
	
	fmt.Printf("\n✅ 網頁報告已儲存: %s\n", filename)
}

// ========== Main 函數 ==========

func main() {
	// 執行測試
	TestWebReport()
}
