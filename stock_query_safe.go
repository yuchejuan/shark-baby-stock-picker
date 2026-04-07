package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// 安全版股票代號查詢
func queryStockCodeSafe(symbol string) (string, error) {
	url := fmt.Sprintf("https://www.twse.com.tw/zh/api/codeQuery?query=%s", symbol)
	
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("網路請求失敗: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取回應失敗: %v", err)
	}
	
	// 解析 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("JSON 解析失敗: %v", err)
	}
	
	// 檢查 suggestions
	suggestions, ok := result["suggestions"].([]interface{})
	if !ok {
		return "", fmt.Errorf("API 回應格式異常")
	}
	
	if len(suggestions) == 0 {
		return "", fmt.Errorf("找不到股票代號 %s（可能不存在或已下市）", symbol)
	}
	
	// 解析第一個建議
	firstSuggestion, ok := suggestions[0].(string)
	if !ok {
		return "", fmt.Errorf("建議格式異常")
	}
	
	// 檢查是否為「無符合之代碼或名稱」
	if strings.Contains(firstSuggestion, "無符合") || strings.Contains(firstSuggestion, "查無") {
		return "", fmt.Errorf("股票代號 %s 不存在或已下市", symbol)
	}
	
	// 嘗試分割
	parts := strings.Split(firstSuggestion, "\t")
	
	if len(parts) >= 2 {
		// 標準格式: "2330\t台積電"
		return parts[1], nil
	} else if len(parts) == 1 {
		// 只有代號: "3105"
		return parts[0], nil
	} else {
		return "", fmt.Errorf("無法解析股票名稱，原始資料: %s", firstSuggestion)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run stock_query_safe.go <股票代號>")
		os.Exit(1)
	}
	
	symbol := os.Args[1]
	
	fmt.Printf("🔍 查詢股票代號: %s\n", symbol)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	name, err := queryStockCodeSafe(symbol)
	if err != nil {
		log.Fatalf("❌ 查詢失敗: %v", err)
	}
	
	fmt.Printf("✅ 股票名稱: %s\n", name)
	fmt.Printf("📊 代號: %s\n", symbol)
}
