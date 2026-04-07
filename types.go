package main

import (
	"database/sql"
	"time"
)

// ========== K線與技術指標 ==========

// K線資料
type KLineData struct {
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// MACD 資料
type MACDData struct {
	DIF float64
	DEA float64
	OSC float64
}

// KD 資料
type KDData struct {
	K float64
	D float64
}

// 布林通道
type BollingerBands struct {
	Upper  float64
	Middle float64
	Lower  float64
	Width  float64
}

// 技術指標
type TechnicalIndicators struct {
	RSI    float64
	MACD   MACDData
	KD     KDData
	BB     BollingerBands
	MA5    float64
	MA10   float64
	MA20   float64
	MA60   float64
	Volume int64
}

// ========== 除權息資料 ==========

// 除權息資訊
type DividendInfo struct {
	Symbol        string  `json:"symbol"`         // 股票代號
	CompanyName   string  `json:"company_name"`   // 公司名稱
	ExDivDate     string  `json:"ex_div_date"`    // 除權息日期 (YYYY-MM-DD)
	CashDividend  float64 `json:"cash_dividend"`  // 現金股利
	StockDividend float64 `json:"stock_dividend"` // 股票股利
	DaysUntil     int     `json:"days_until"`     // 距離除息天數（負數=已除息）
	FillStatus    string  `json:"fill_status"`    // 填息狀態（已填息/填息中/未填息/未除息）
	FillRate      float64 `json:"fill_rate"`      // 填息率 %
	
	// 計算用欄位
	ExDivPrice    float64 `json:"ex_div_price"`   // 除息前收盤價
	CurrentPrice  float64 `json:"current_price"`  // 當前價格（用於計算填息率）
}

// ========== 市場狀態 ==========

// 市場狀態
type MarketState int

const (
	MarketBullish    MarketState = iota // 多頭
	MarketBearish                       // 空頭
	MarketSideways                      // 盤整
)

func (ms MarketState) String() string {
	switch ms {
	case MarketBullish:
		return "多頭市場"
	case MarketBearish:
		return "空頭市場"
	case MarketSideways:
		return "盤整市場"
	default:
		return "未知"
	}
}

// 市場狀態報告
type MarketStateReport struct {
	State           MarketState `json:"state"`
	StateName       string      `json:"state_name"`
	Confidence      float64     `json:"confidence"`
	Description     string      `json:"description"`
	RecommendAction string      `json:"recommend_action"`
}

// ========== 風險評估 ==========

// 風險類型
type RiskType int

const (
	RiskBearishAlignment RiskType = iota // 空頭排列
	RiskBelowMA60                        // 跌破季線
	RiskVolumeShrink                     // 成交量萎縮
	RiskMACDDeath                        // MACD死亡交叉
	RiskRSIOverbought                    // RSI超買
	RiskKDOverbought                     // KD超買
	RiskPriceBreakdown                   // 價格破底
)

func (rt RiskType) String() string {
	switch rt {
	case RiskBearishAlignment:
		return "空頭排列"
	case RiskBelowMA60:
		return "跌破季線"
	case RiskVolumeShrink:
		return "成交量萎縮"
	case RiskMACDDeath:
		return "MACD死亡交叉"
	case RiskRSIOverbought:
		return "RSI超買"
	case RiskKDOverbought:
		return "KD超買"
	case RiskPriceBreakdown:
		return "價格破底"
	default:
		return "未知風險"
	}
}

// 風險等級
type RiskLevel int

const (
	RiskLow    RiskLevel = iota // 低風險
	RiskMedium                  // 中風險
	RiskHigh                    // 高風險
	RiskExtreme                 // 極高風險
)

func (rl RiskLevel) String() string {
	switch rl {
	case RiskLow:
		return "🟢 低風險"
	case RiskMedium:
		return "🟡 中風險"
	case RiskHigh:
		return "🟠 高風險"
	case RiskExtreme:
		return "🔴 極高風險"
	default:
		return "未知"
	}
}

// 風險警示
type RiskAlert struct {
	Type        RiskType  `json:"type"`
	TypeName    string    `json:"type_name"`
	Severity    int       `json:"severity"`    // 嚴重程度 1-10
	Description string    `json:"description"` // 詳細說明
	Triggered   bool      `json:"triggered"`
}

// 風險報告
type RiskReport struct {
	Symbol         string      `json:"symbol"`
	Name           string      `json:"name"`
	Price          float64     `json:"price"`
	RiskLevel      RiskLevel   `json:"risk_level"`
	RiskLevelName  string      `json:"risk_level_name"`
	TotalScore     int         `json:"total_score"`      // 風險總分 (0-100)
	Alerts         []RiskAlert `json:"alerts"`           // 觸發的警示
	Recommendation string      `json:"recommendation"`   // 操作建議
}

// ========== 回測系統 ==========

// 回測記錄
type BacktestRecord struct {
	ID             int       `json:"id"`
	Date           string    `json:"date"`
	Symbol         string    `json:"symbol"`
	Name           string    `json:"name"`
	EntryPrice     float64   `json:"entry_price"`
	Score          int       `json:"score"`
	MarketState    string    `json:"market_state"`
	RiskLevel      string    `json:"risk_level"`
	ExitPrice      float64   `json:"exit_price"`
	Return7Days    float64   `json:"return_7days"`
	ProfitLoss     float64   `json:"profit_loss"`
	IsWin          bool      `json:"is_win"`
	Checked        bool      `json:"checked"`
	CheckedAt      string    `json:"checked_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// 回測統計
type BacktestStats struct {
	TotalTrades    int     `json:"total_trades"`
	WinCount       int     `json:"win_count"`
	LossCount      int     `json:"loss_count"`
	WinRate        float64 `json:"win_rate"`
	AvgReturn      float64 `json:"avg_return"`
	MaxReturn      float64 `json:"max_return"`
	MaxLoss        float64 `json:"max_loss"`
	TotalReturn    float64 `json:"total_return"`
	BestStock      string  `json:"best_stock"`
	WorstStock     string  `json:"worst_stock"`
}

// ========== 指標權重配置 ==========

// 指標權重
type IndicatorWeights struct {
	RSI    float64 `json:"rsi"`
	MACD   float64 `json:"macd"`
	KD     float64 `json:"kd"`
	MA     float64 `json:"ma"`
	BB     float64 `json:"bb"`
	Volume float64 `json:"volume"`
}

// 基礎權重（盤整市場）
var BaseWeights = IndicatorWeights{
	RSI:    0.15,
	MACD:   0.15,
	KD:     0.15,
	MA:     0.15,
	BB:     0.25, // 盤整時布林通道最重要
	Volume: 0.15,
}

// ========== 輔助函數 ==========

// 計算簡單移動平均
func CalculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}

	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}

	return sum / float64(period)
}

// ========== 資料庫相關（共用變數）==========

var DB *sql.DB // 共用資料庫連線
