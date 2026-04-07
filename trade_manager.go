package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 交易記錄結構
type Trade struct {
	ID          int       `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "buy" or "sell"
	Shares      int       `json:"shares"`
	Price       float64   `json:"price"`
	Amount      float64   `json:"amount"`
	Date        time.Time `json:"date"`
	Note        string    `json:"note"`
	CreatedAt   time.Time `json:"created_at"`
}

// 持股記錄
type Holding struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	TotalShares int       `json:"total_shares"`
	AvgBuyPrice float64   `json:"avg_buy_price"`
	TotalCost   float64   `json:"total_cost"`
	CurrentPrice float64  `json:"current_price"`
	CurrentValue float64  `json:"current_value"`
	ProfitLoss  float64   `json:"profit_loss"`
	ReturnPct   float64   `json:"return_pct"`
	LastUpdate  time.Time `json:"last_update"`
}

// 持股批次記錄（分批次顯示）
type HoldingBatch struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Shares      int       `json:"shares"`
	BuyPrice    float64   `json:"buy_price"`
	BuyDate     time.Time `json:"buy_date"`
	CurrentPrice float64  `json:"current_price"`
	CurrentValue float64  `json:"current_value"`
	ProfitLoss  float64   `json:"profit_loss"`
	ReturnPct   float64   `json:"return_pct"`
	Reason      string    `json:"reason"`
	DaysHeld    int       `json:"days_held"`
}

// 損益統計
type ProfitStats struct {
	TotalRealized   float64 `json:"total_realized"`    // 已實現損益
	TotalUnrealized float64 `json:"total_unrealized"`  // 未實現損益
	TotalProfit     float64 `json:"total_profit"`      // 總損益
	WinRate         float64 `json:"win_rate"`          // 勝率
	TotalTrades     int     `json:"total_trades"`      // 總交易數
	WinTrades       int     `json:"win_trades"`        // 獲利交易數
}

var db *sql.DB

func main() {
	var err error
	
	// 初始化資料庫
	db, err = sql.Open("sqlite3", "./stock_trades.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// 建立資料表
	createTables()
	
	// 設定 HTTP 路由
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/api", apiRootHandler)
	http.HandleFunc("/api/", apiRootHandler)
	http.HandleFunc("/api/trade/add", corsMiddleware(addTradeHandler))
	http.HandleFunc("/api/trades", corsMiddleware(getTradesHandler))
	http.HandleFunc("/api/holdings", corsMiddleware(getHoldingsHandler))
	http.HandleFunc("/api/holdings/batches", corsMiddleware(getHoldingsBatchesHandler))
	http.HandleFunc("/api/holdings/update-prices", corsMiddleware(updatePricesHandler))
	http.HandleFunc("/api/stats", corsMiddleware(getStatsHandler))
	http.HandleFunc("/api/trade/delete", corsMiddleware(deleteTradeHandler))
	
	fmt.Println("🦈 交易管理 API 伺服器啟動於 http://localhost:8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func createTables() {
	// 交易記錄表
	tradeTable := `
	CREATE TABLE IF NOT EXISTS trades (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		shares INTEGER NOT NULL,
		price REAL NOT NULL,
		amount REAL NOT NULL,
		date DATETIME NOT NULL,
		note TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err := db.Exec(tradeTable)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("✅ 資料庫初始化完成")
}

// CORS 中介層
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next(w, r)
	}
}

// 新增交易記錄
func addTradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只接受 POST 請求", http.StatusMethodNotAllowed)
		return
	}
	
	var trade Trade
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "無效的 JSON 格式", http.StatusBadRequest)
		return
	}
	
	// 計算交易金額
	trade.Amount = float64(trade.Shares) * trade.Price
	
	// 如果沒有指定日期，使用現在時間
	if trade.Date.IsZero() {
		trade.Date = time.Now()
	}
	
	// 插入資料庫
	result, err := db.Exec(`
		INSERT INTO trades (symbol, name, type, shares, price, amount, date, note)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, trade.Symbol, trade.Name, trade.Type, trade.Shares, trade.Price, trade.Amount, trade.Date, trade.Note)
	
	if err != nil {
		http.Error(w, fmt.Sprintf("資料庫錯誤: %v", err), http.StatusInternalServerError)
		return
	}
	
	id, _ := result.LastInsertId()
	trade.ID = int(id)
	
	// 更新 portfolio.json
	updatePortfolio()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "交易記錄新增成功",
		"trade":   trade,
	})
}

// 取得所有交易記錄
func getTradesHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	limit := r.URL.Query().Get("limit")
	
	query := "SELECT id, symbol, name, type, shares, price, amount, date, note, created_at FROM trades"
	args := []interface{}{}
	
	if symbol != "" {
		query += " WHERE symbol = ?"
		args = append(args, symbol)
	}
	
	query += " ORDER BY date DESC"
	
	if limit != "" {
		query += " LIMIT ?"
		limitInt, _ := strconv.Atoi(limit)
		args = append(args, limitInt)
	}
	
	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("資料庫錯誤: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var trades []Trade
	for rows.Next() {
		var t Trade
		err := rows.Scan(&t.ID, &t.Symbol, &t.Name, &t.Type, &t.Shares, &t.Price, &t.Amount, &t.Date, &t.Note, &t.CreatedAt)
		if err != nil {
			continue
		}
		trades = append(trades, t)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trades)
}

// 取得當前持股
func getHoldingsHandler(w http.ResponseWriter, r *http.Request) {
	// 計算每支股票的持股
	rows, err := db.Query(`
		SELECT symbol, name,
			SUM(CASE WHEN type = 'buy' THEN shares ELSE -shares END) as total_shares,
			SUM(CASE WHEN type = 'buy' THEN amount ELSE -amount END) as total_cost
		FROM trades
		GROUP BY symbol, name
		HAVING total_shares > 0
	`)
	
	if err != nil {
		http.Error(w, fmt.Sprintf("資料庫錯誤: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	var holdings []Holding
	for rows.Next() {
		var h Holding
		err := rows.Scan(&h.Symbol, &h.Name, &h.TotalShares, &h.TotalCost)
		if err != nil {
			continue
		}
		
		// 計算平均買入價
		h.AvgBuyPrice = h.TotalCost / float64(h.TotalShares)
		h.LastUpdate = time.Now()
		
		holdings = append(holdings, h)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(holdings)
}

// 從 portfolio.json 讀取即時股價
// 股價快取（避免頻繁查詢證交所 API）
var priceCache = make(map[string]struct {
	price     float64
	timestamp time.Time
})
var priceCacheDuration = 3 * time.Minute // 快取 3 分鐘

// 從證交所即時報價 API 取得股價
func fetchRealtimePrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://mis.twse.com.tw/stock/api/getStockInfo.jsp?ex_ch=tse_%s.tw", symbol)
	
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	
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
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	
	msgArray, ok := result["msgArray"].([]interface{})
	if !ok || len(msgArray) == 0 {
		return 0, fmt.Errorf("查無股票資料: %s", symbol)
	}
	
	msg := msgArray[0].(map[string]interface{})
	priceStr, _ := msg["z"].(string)
	
	if priceStr == "-" || priceStr == "" {
		// 如果沒有即時價（可能盤後），嘗試取得昨收價
		priceStr, _ = msg["y"].(string)
	}
	
	if priceStr == "-" || priceStr == "" {
		return 0, fmt.Errorf("無有效價格: %s", symbol)
	}
	
	var price float64
	_, err = fmt.Sscanf(priceStr, "%f", &price)
	if err != nil {
		return 0, fmt.Errorf("解析價格失敗: %s", priceStr)
	}
	
	return price, nil
}

func getCurrentPrices() map[string]float64 {
	prices := make(map[string]float64)
	
	// 查詢資料庫中所有持倉的股票代號
	rows, err := db.Query(`
		SELECT DISTINCT symbol FROM trades WHERE type = 'buy'
		AND symbol NOT IN (
			SELECT DISTINCT symbol FROM trades WHERE type = 'sell'
			GROUP BY symbol HAVING SUM(shares) >= (
				SELECT SUM(shares) FROM trades WHERE type = 'buy' AND trades.symbol = symbol
			)
		)
	`)
	
	if err != nil {
		fmt.Println("查詢持倉股票失敗:", err)
		return prices
	}
	defer rows.Close()
	
	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err == nil {
			symbols = append(symbols, symbol)
		}
	}
	
	// 對每支股票查詢即時價格（使用快取）
	for _, symbol := range symbols {
		now := time.Now()
		
		// 檢查快取
		if cached, ok := priceCache[symbol]; ok {
			if now.Sub(cached.timestamp) < priceCacheDuration {
				prices[symbol] = cached.price
				fmt.Printf("💾 使用快取: %s = %.2f (%.0f秒前)\n", symbol, cached.price, now.Sub(cached.timestamp).Seconds())
				continue
			}
		}
		
		// 查詢即時價格
		price, err := fetchRealtimePrice(symbol)
		if err != nil {
			fmt.Printf("⚠️  %s 查詢失敗: %v\n", symbol, err)
			// 降級：嘗試從 portfolio.json 取得
			continue
		}
		
		prices[symbol] = price
		priceCache[symbol] = struct {
			price     float64
			timestamp time.Time
		}{price, now}
		
		fmt.Printf("✅ %s = %.2f\n", symbol, price)
		
		// 避免過度查詢證交所 API
		time.Sleep(200 * time.Millisecond)
	}
	
	// 降級方案：從 portfolio.json 補充缺失的股價
	if len(prices) < len(symbols) {
		wd, _ := os.Getwd()
		data, err := ioutil.ReadFile(filepath.Join(wd, "html", "portfolio.json"))
		if err == nil {
			var portfolioData struct {
				Holdings []struct {
					Symbol       string  `json:"symbol"`
					CurrentPrice float64 `json:"current_price"`
				} `json:"holdings"`
			}
			
			if err := json.Unmarshal(data, &portfolioData); err == nil {
				for _, h := range portfolioData.Holdings {
					if _, exists := prices[h.Symbol]; !exists && h.CurrentPrice > 0 {
						prices[h.Symbol] = h.CurrentPrice
						fmt.Printf("📁 降級使用 portfolio.json: %s = %.2f\n", h.Symbol, h.CurrentPrice)
					}
				}
			}
		}
	}
	
	return prices
}

// 取得分批次持股（每一筆買入分開顯示）
func getHoldingsBatchesHandler(w http.ResponseWriter, r *http.Request) {
	// 讀取即時股價
	currentPrices := getCurrentPrices()
	
	// 查詢所有買入記錄，減去賣出記錄
	// 使用子查詢計算每筆買入的剩餘股數
	rows, err := db.Query(`
		WITH buy_trades AS (
			SELECT id, symbol, name, shares, price, date, note
			FROM trades
			WHERE type = 'buy'
		),
		sell_totals AS (
			SELECT symbol, SUM(shares) as sold_shares
			FROM trades
			WHERE type = 'sell'
			GROUP BY symbol
		)
		SELECT 
			b.id, b.symbol, b.name, b.shares, b.price, b.date, b.note,
			COALESCE(s.sold_shares, 0) as sold
		FROM buy_trades b
		LEFT JOIN sell_totals s ON b.symbol = s.symbol
		ORDER BY b.symbol, b.date
	`)
	
	if err != nil {
		http.Error(w, fmt.Sprintf("資料庫錯誤: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	// 用來追蹤每支股票已分配的賣出股數
	soldAllocated := make(map[string]int)
	
	var batches []HoldingBatch
	for rows.Next() {
		var batch HoldingBatch
		var id int
		var buyDate time.Time
		var note string
		var totalSold int
		
		err := rows.Scan(&id, &batch.Symbol, &batch.Name, &batch.Shares, &batch.BuyPrice, &buyDate, &note, &totalSold)
		if err != nil {
			continue
		}
		
		// 計算這批次的剩餘股數（FIFO 原則）
		alreadySold := soldAllocated[batch.Symbol]
		remainingShares := batch.Shares
		
		if alreadySold < totalSold {
			toSell := totalSold - alreadySold
			if toSell >= remainingShares {
				// 這批已經全部賣出
				soldAllocated[batch.Symbol] += remainingShares
				continue
			} else {
				// 部分賣出
				remainingShares -= toSell
				soldAllocated[batch.Symbol] += toSell
			}
		}
		
		// 如果沒有剩餘股數，跳過
		if remainingShares <= 0 {
			continue
		}
		
		batch.Shares = remainingShares
		batch.BuyDate = buyDate
		batch.Reason = note
		
		// 計算持有天數
		batch.DaysHeld = int(time.Since(buyDate).Hours() / 24)
		
		// 取得即時股價
		if currentPrice, ok := currentPrices[batch.Symbol]; ok {
			batch.CurrentPrice = currentPrice
		} else {
			batch.CurrentPrice = batch.BuyPrice // 如果沒有即時價，用買入價
		}
		
		// 計算損益
		batch.CurrentValue = float64(batch.Shares) * batch.CurrentPrice
		costValue := float64(batch.Shares) * batch.BuyPrice
		batch.ProfitLoss = batch.CurrentValue - costValue
		
		if costValue > 0 {
			batch.ReturnPct = (batch.ProfitLoss / costValue) * 100
		}
		
		batches = append(batches, batch)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(batches)
}

// 取得損益統計
func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	var stats ProfitStats
	
	// 取得時間範圍參數（預設 1 年）
	period := r.URL.Query().Get("period")
	var dateFilter string
	if period == "1year" || period == "" {
		dateFilter = "AND date >= datetime('now', '-1 year')"
	}
	
	// 計算總交易數
	db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM trades WHERE 1=1 %s
	`, dateFilter)).Scan(&stats.TotalTrades)
	
	// 計算已實現損益（配對買賣計算）
	// 簡化版本：賣出金額 - 買入成本
	var totalBuyAmount, totalSellAmount float64
	db.QueryRow(fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(CASE WHEN type='buy' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type='sell' THEN amount ELSE 0 END), 0)
		FROM trades 
		WHERE 1=1 %s
	`, dateFilter)).Scan(&totalBuyAmount, &totalSellAmount)
	
	stats.TotalRealized = totalSellAmount - totalBuyAmount
	
	// 計算勝率（簡化：假設每筆賣出是一次交易）
	var winTrades int
	db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*) FROM trades 
		WHERE type='sell' AND amount > 0 %s
	`, dateFilter)).Scan(&winTrades)
	
	stats.WinTrades = winTrades
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(winTrades) / float64(stats.TotalTrades) * 100
	}
	
	stats.TotalProfit = stats.TotalRealized + stats.TotalUnrealized
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 刪除交易記錄
func deleteTradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "只接受 DELETE 請求", http.StatusMethodNotAllowed)
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "缺少交易 ID", http.StatusBadRequest)
		return
	}
	
	_, err := db.Exec("DELETE FROM trades WHERE id = ?", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("刪除失敗: %v", err), http.StatusInternalServerError)
		return
	}
	
	// 更新 portfolio.json
	updatePortfolio()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "交易記錄已刪除",
	})
}

// 根路徑處理
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head><title>🦈 交易管理 API</title></head>
		<body style="font-family: sans-serif; padding: 40px; background: #f5f5f5;">
			<h1>🦈 鯊魚寶寶交易管理 API</h1>
			<p>API 伺服器正在運行中</p>
			<h2>可用端點：</h2>
			<ul>
				<li><a href="/api/trades">/api/trades</a> - 查詢所有交易</li>
				<li><a href="/api/holdings">/api/holdings</a> - 查詢當前持股</li>
				<li><a href="/api/stats">/api/stats</a> - 績效統計</li>
				<li>POST /api/trade/add - 新增交易</li>
				<li>DELETE /api/trade/delete?id=X - 刪除交易</li>
			</ul>
		</body>
		</html>
	`)
}

// API 根路徑處理
func apiRootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"message": "🦈 鯊魚寶寶交易管理 API",
		"version": "2.0",
		"endpoints": []string{
			"/api/trades",
			"/api/holdings",
			"/api/stats",
			"/api/trade/add",
			"/api/trade/delete",
		},
	})
}

// 更新 portfolio.json（整合到現有系統）
func updatePortfolio() {
	// 這裡會讀取資料庫，重新生成 portfolio.json
	// 與現有的 web_updater.go 整合
	fmt.Println("📊 更新投資組合資料...")
	
	// 執行現有的更新程式
	// （可以選擇性地呼叫 web_updater.go 或整合進來）
}

// 手動更新股價 API
func updatePricesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "只接受 POST 請求", http.StatusMethodNotAllowed)
		return
	}
	
	fmt.Println("🔄 收到手動更新股價請求...")
	
	// 執行 web_updater.go 來重新抓取股價
	cmd := exec.Command("go", "run", "web_updater.go")
	cmd.Dir = "/home/administrator/.openclaw/workspace"
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("更新失敗: %v\n輸出: %s", err, string(output))
		fmt.Println("❌", errMsg)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": errMsg,
		})
		return
	}
	
	fmt.Println("✅ 股價更新成功")
	fmt.Println(string(output))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "股價已更新",
		"output": string(output),
	})
}
