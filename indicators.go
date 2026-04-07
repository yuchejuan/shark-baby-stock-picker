package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

// ========== 資料來源 API ==========

// TWSE API 結構
type TWSeQuote struct {
	MsgArray []struct {
		Symbol string `json:"c"`
		Name   string `json:"n"`
		Price  string `json:"z"`
		Open   string `json:"o"`
		High   string `json:"h"`
		Low    string `json:"l"`
		Change string `json:"y"`
		Volume string `json:"v"`
	} `json:"msgArray"`
}

// Yahoo Finance 歷史資料結構
type YahooHistoricalData struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// ========== 資料取得函數 ==========

// 取得台股即時報價
func GetTWSEPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_%s.tw", symbol)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	
	var result TWSeQuote
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	
	if len(result.MsgArray) == 0 {
		return 0, fmt.Errorf("無資料")
	}
	
	var price float64
	fmt.Sscanf(result.MsgArray[0].Price, "%f", &price)
	
	return price, nil
}

// 取得歷史 K 線資料（60天）
func FetchHistoricalData(symbol string) ([]KLineData, error) {
	// Yahoo Finance API
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s.TW?interval=1d&range=3mo",
		symbol,
	)

	client := &http.Client{Timeout: 15 * time.Second}
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data YahooHistoricalData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if len(data.Chart.Result) == 0 {
		return nil, fmt.Errorf("無歷史資料")
	}

	result := data.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("無報價資料")
	}

	quote := result.Indicators.Quote[0]
	timestamps := result.Timestamp

	var klines []KLineData
	for i := 0; i < len(timestamps); i++ {
		// 過濾掉無效資料
		if quote.Close[i] == 0 {
			continue
		}

		kline := KLineData{
			Date:   time.Unix(timestamps[i], 0).Format("2006-01-02"),
			Open:   quote.Open[i],
			High:   quote.High[i],
			Low:    quote.Low[i],
			Close:  quote.Close[i],
			Volume: quote.Volume[i],
		}
		klines = append(klines, kline)
	}

	return klines, nil
}

// ========== 資料提取函數 ==========

// 從 K 線提取收盤價
func ExtractClosePrices(klines []KLineData) []float64 {
	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.Close
	}
	return prices
}

// 從 K 線提取高價
func ExtractHighPrices(klines []KLineData) []float64 {
	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.High
	}
	return prices
}

// 從 K 線提取低價
func ExtractLowPrices(klines []KLineData) []float64 {
	prices := make([]float64, len(klines))
	for i, k := range klines {
		prices[i] = k.Low
	}
	return prices
}

// 從 K 線提取成交量
func ExtractVolumes(klines []KLineData) []int64 {
	volumes := make([]int64, len(klines))
	for i, k := range klines {
		volumes[i] = k.Volume
	}
	return volumes
}

// ========== 技術指標計算 ==========

// RSI 計算
func CalculateRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50.0
	}

	var gains, losses float64
	for i := len(prices) - period; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

// EMA 計算
func CalculateEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return prices[len(prices)-1]
	}

	multiplier := 2.0 / float64(period+1)
	ema := prices[len(prices)-period]

	for i := len(prices) - period + 1; i < len(prices); i++ {
		ema = (prices[i] * multiplier) + (ema * (1 - multiplier))
	}

	return ema
}

// MACD 計算
func CalculateMACD(prices []float64) MACDData {
	if len(prices) < 26 {
		return MACDData{0, 0, 0}
	}

	ema12 := CalculateEMA(prices, 12)
	ema26 := CalculateEMA(prices, 26)
	dif := ema12 - ema26

	// 簡化版 DEA（實際應該用 DIF 的 EMA）
	dea := dif * 0.9

	osc := dif - dea

	return MACDData{
		DIF: dif,
		DEA: dea,
		OSC: osc,
	}
}

// KD 計算
func CalculateKD(highs, lows, closes []float64, period int) KDData {
	if len(closes) < period {
		return KDData{50, 50}
	}

	// 找出最高價和最低價
	high := highs[len(highs)-1]
	low := lows[len(lows)-1]
	for i := len(highs) - period; i < len(highs); i++ {
		if highs[i] > high {
			high = highs[i]
		}
		if lows[i] < low {
			low = lows[i]
		}
	}

	close := closes[len(closes)-1]
	rsv := 50.0
	if high != low {
		rsv = ((close - low) / (high - low)) * 100
	}

	k := rsv * 0.33 + 50*0.67 // 簡化版
	d := k * 0.33 + 50*0.67

	return KDData{K: k, D: d}
}

// 布林通道計算
func CalculateBollingerBands(prices []float64, period int, stdDev float64) BollingerBands {
	if len(prices) < period {
		current := prices[len(prices)-1]
		return BollingerBands{current, current, current, 0}
	}

	// 計算中軌 (MA20)
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	middle := sum / float64(period)

	// 計算標準差
	variance := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		variance += math.Pow(prices[i]-middle, 2)
	}
	std := math.Sqrt(variance / float64(period))

	upper := middle + (stdDev * std)
	lower := middle - (stdDev * std)
	width := ((upper - lower) / middle) * 100

	return BollingerBands{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
		Width:  width,
	}
}

// ========== 成交量分析 ==========

// 計算成交量平均值
func CalculateAvgVolume(volumes []int64, period int) int64 {
	if len(volumes) < period {
		return 0
	}

	sum := int64(0)
	for i := len(volumes) - period; i < len(volumes); i++ {
		sum += volumes[i]
	}

	return sum / int64(period)
}

// 檢查成交量是否萎縮（當日量 < 20日均量 50%）
func IsVolumeShrinking(volumes []int64) bool {
	if len(volumes) < 21 {
		return false
	}

	avgVol := CalculateAvgVolume(volumes[:len(volumes)-1], 20)
	currentVol := volumes[len(volumes)-1]

	return currentVol < avgVol/2
}

// ========== 市場分析與評分（簡化版）==========

// 偵測市場狀態（簡化版）
func DetectMarketState(klines []KLineData, currentPrice float64) MarketState {
	if len(klines) < 30 {
		return MarketSideways
	}

	closes := ExtractClosePrices(klines)
	ma5 := CalculateMA(closes, 5)
	ma20 := CalculateMA(closes, 20)

	// 簡單判斷：5日線 > 20日線 = 多頭
	if ma5 > ma20*1.02 {
		return MarketBullish
	} else if ma5 < ma20*0.98 {
		return MarketBearish
	}

	return MarketSideways
}

// 評估風險（簡化版）
func AssessRisk(symbol, name string, klines []KLineData, indicators TechnicalIndicators, price float64) RiskReport {
	var alerts []RiskAlert
	score := 0

	// RSI 風險檢查
	if indicators.RSI > 70 {
		alerts = append(alerts, RiskAlert{
			Type:        RiskRSIOverbought,
			TypeName:    "RSI超買",
			Severity:    6,
			Description: fmt.Sprintf("RSI=%.1f 超買警示", indicators.RSI),
			Triggered:   true,
		})
		score += 15
	}

	// 均線排列風險
	if indicators.MA5 < indicators.MA10 && indicators.MA10 < indicators.MA20 {
		alerts = append(alerts, RiskAlert{
			Type:        RiskBearishAlignment,
			TypeName:    "空頭排列",
			Severity:    8,
			Description: "均線空頭排列",
			Triggered:   true,
		})
		score += 25
	}

	// 判斷風險等級
	riskLevel := RiskLow
	if score > 50 {
		riskLevel = RiskExtreme
	} else if score > 30 {
		riskLevel = RiskHigh
	} else if score > 15 {
		riskLevel = RiskMedium
	}

	return RiskReport{
		Symbol:         symbol,
		Name:           name,
		Price:          price,
		RiskLevel:      riskLevel,
		RiskLevelName:  riskLevel.String(),
		TotalScore:     score,
		Alerts:         alerts,
		Recommendation: getRecommendation(riskLevel),
	}
}

func getRecommendation(level RiskLevel) string {
	switch level {
	case RiskLow:
		return "可建立部位"
	case RiskMedium:
		return "謹慎觀察"
	case RiskHigh:
		return "不宜進場"
	case RiskExtreme:
		return "建議出場"
	default:
		return "觀望"
	}
}

// 動態權重（簡化版）
func GetDynamicWeights(marketState MarketState) IndicatorWeights {
	return BaseWeights // 目前直接回傳基礎權重
}

// 動態評分（簡化版）
func CalculateDynamicScore(indicators TechnicalIndicators, weights IndicatorWeights, marketState MarketState) float64 {
	score := 0

	// RSI 評分
	if indicators.RSI >= 30 && indicators.RSI <= 70 {
		score += 20
	}

	// MACD 評分
	if indicators.MACD.DIF > indicators.MACD.DEA {
		score += 15
	}

	// 均線多頭排列
	if indicators.MA5 > indicators.MA10 && indicators.MA10 > indicators.MA20 {
		score += 25
	}

	// 布林通道評分
	if indicators.BB.Width < 10 {
		score += 10
	}

	return float64(score)
}
