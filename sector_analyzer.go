package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// 產業指數結構
type SectorIndex struct {
	Code        string  `json:"code"`         // 產業代碼
	Name        string  `json:"name"`         // 產業名稱
	Price       float64 `json:"price"`        // 收盤價
	Change      float64 `json:"change"`       // 漲跌點數
	ChangeRate  float64 `json:"change_rate"`  // 漲跌幅(%)
	Volume      float64 `json:"volume"`       // 成交量(千股)
	LeadStocks  []string `json:"lead_stocks"` // 領漲/跌股票
}

// 產業熱度報告
type SectorHeatmap struct {
	Date        string         `json:"date"`         // 日期
	UpdateTime  string         `json:"update_time"`  // 更新時間
	HotSectors  []SectorIndex  `json:"hot_sectors"`  // 熱門產業(前5)
	ColdSectors []SectorIndex  `json:"cold_sectors"` // 冷門產業(後5)
	AllSectors  []SectorIndex  `json:"all_sectors"`  // 所有產業
}

// TWSE API 回應結構
type TWSEResponse struct {
	Stat   string     `json:"stat"`
	Date   string     `json:"date"`
	Fields []string   `json:"fields"`
	Data   [][]string `json:"data"`
}

func main() {
	fmt.Println("🦈 鯊魚寶寶產業熱度分析")
	fmt.Println("========================================")
	
	// 取得今日日期
	now := time.Now()
	dateStr := now.Format("20060102")
	
	fmt.Printf("📅 分析日期: %s\n", now.Format("2006-01-02"))
	fmt.Println("")
	
	// 抓取證交所產業分類指數
	fmt.Println("🔍 開始抓取產業分類指數...")
	sectors, err := fetchSectorIndices(dateStr)
	if err != nil {
		fmt.Printf("❌ 抓取失敗: %v\n", err)
		// 嘗試前一個交易日
		yesterday := now.AddDate(0, 0, -1)
		dateStr = yesterday.Format("20060102")
		fmt.Printf("🔄 嘗試前一日 %s...\n", yesterday.Format("2006-01-02"))
		sectors, err = fetchSectorIndices(dateStr)
		if err != nil {
			fmt.Printf("❌ 仍然失敗: %v\n", err)
			os.Exit(1)
		}
	}
	
	fmt.Printf("✅ 成功抓取 %d 個產業指數\n", len(sectors))
	fmt.Println("")
	
	// 排序（按漲跌幅）
	sort.Slice(sectors, func(i, j int) bool {
		return sectors[i].ChangeRate > sectors[j].ChangeRate
	})
	
	// 取前5和後5
	hotCount := 5
	if len(sectors) < 5 {
		hotCount = len(sectors)
	}
	
	hotSectors := sectors[:hotCount]
	coldSectors := make([]SectorIndex, 0)
	if len(sectors) >= 5 {
		coldSectors = sectors[len(sectors)-5:]
		// 反轉順序（從最差開始）
		for i, j := 0, len(coldSectors)-1; i < j; i, j = i+1, j-1 {
			coldSectors[i], coldSectors[j] = coldSectors[j], coldSectors[i]
		}
	}
	
	// 顯示結果
	fmt.Println("🔥 熱門產業（漲幅前5）")
	fmt.Println("----------------------------------------")
	for i, sector := range hotSectors {
		fmt.Printf("%d. %s %+.2f%%\n", i+1, sector.Name, sector.ChangeRate)
	}
	fmt.Println("")
	
	fmt.Println("❄️  冷門產業（跌幅前5）")
	fmt.Println("----------------------------------------")
	for i, sector := range coldSectors {
		fmt.Printf("%d. %s %+.2f%%\n", i+1, sector.Name, sector.ChangeRate)
	}
	fmt.Println("")
	
	// 建立報告
	report := SectorHeatmap{
		Date:        now.Format("2006-01-02"),
		UpdateTime:  now.Format("2006-01-02 15:04:05"),
		HotSectors:  hotSectors,
		ColdSectors: coldSectors,
		AllSectors:  sectors,
	}
	
	// 儲存 JSON
	outputPath := "stock_web/sector_heatmap.json"
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("❌ JSON 編碼失敗: %v\n", err)
		os.Exit(1)
	}
	
	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		fmt.Printf("❌ 儲存失敗: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✅ 產業熱度報告已儲存: %s\n", outputPath)
	fmt.Println("========================================")
	fmt.Println("🦈 分析完成")
}

// 抓取證交所產業分類指數
func fetchSectorIndices(date string) ([]SectorIndex, error) {
	// 證交所 API
	url := fmt.Sprintf("https://www.twse.com.tw/rwd/zh/indices/MKT9U?date=%s&response=json", date)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	// 設定 User-Agent（模擬瀏覽器）
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP 狀態碼: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	// 解析 JSON
	var twseResp TWSEResponse
	err = json.Unmarshal(body, &twseResp)
	if err != nil {
		return nil, err
	}
	
	if twseResp.Stat != "OK" {
		return nil, fmt.Errorf("API 回應狀態: %s", twseResp.Stat)
	}
	
	// 轉換資料
	sectors := make([]SectorIndex, 0)
	
	for _, row := range twseResp.Data {
		if len(row) < 4 {
			continue
		}
		
		// 解析欄位
		name := strings.TrimSpace(row[0])
		price := parseFloat(row[1])
		change := parseFloat(row[2])
		changeRate := parseFloat(row[3])
		
		// 過濾無效資料
		if name == "" || price == 0 {
			continue
		}
		
		sector := SectorIndex{
			Code:       generateCode(name),
			Name:       name,
			Price:      price,
			Change:     change,
			ChangeRate: changeRate,
			LeadStocks: []string{}, // 暫時空白（未來可從個股資料推算）
		}
		
		sectors = append(sectors, sector)
	}
	
	return sectors, nil
}

// 解析浮點數
func parseFloat(s string) float64 {
	// 移除逗號
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	
	var val float64
	fmt.Sscanf(s, "%f", &val)
	return val
}

// 產生產業代碼
func generateCode(name string) string {
	// 簡單映射（未來可擴充）
	codeMap := map[string]string{
		"水泥工業":   "M1100",
		"食品工業":   "M1200",
		"塑膠工業":   "M1300",
		"紡織纖維":   "M1400",
		"電機機械":   "M1500",
		"電器電纜":   "M1600",
		"化學生技醫療": "M1700",
		"玻璃陶瓷":   "M1800",
		"造紙工業":   "M1900",
		"鋼鐵工業":   "M2000",
		"橡膠工業":   "M2100",
		"汽車工業":   "M2200",
		"半導體業":   "M2300",
		"電腦及週邊設備業": "M2400",
		"光電業":    "M2500",
		"通信網路業":  "M2600",
		"電子零組件業": "M2700",
		"電子通路業":  "M2800",
		"資訊服務業":  "M2900",
		"其他電子業":  "M3000",
		"建材營造業":  "M3100",
		"航運業":    "M3200",
		"觀光事業":   "M3300",
		"金融保險業":  "M3400",
		"貿易百貨業":  "M3500",
		"油電燃氣業":  "M3600",
		"其他":     "M9900",
	}
	
	if code, exists := codeMap[name]; exists {
		return code
	}
	
	return "M0000"
}
