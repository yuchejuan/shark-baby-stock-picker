package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// 啟動簡單的 HTTP API 服務
func main() {
	http.HandleFunc("/api/query", handleStockQuery)
	http.HandleFunc("/health", handleHealth)
	
	port := "8765"
	fmt.Printf("🦈 股票查詢 API 啟動於 http://localhost:%s\n", port)
	fmt.Println("📡 端點: /api/query?symbol=2330")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// 健康檢查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}

// 股票查詢處理
func handleStockQuery(w http.ResponseWriter, r *http.Request) {
	// 允許跨域
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	// 取得股票代號
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "缺少股票代號參數",
		})
		return
	}
	
	log.Printf("📊 查詢股票: %s", symbol)
	
	// 執行 stock_query_service.go
	cmd := exec.Command("go", "run", "stock_query_service.go", symbol)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		log.Printf("❌ 查詢失敗: %v", err)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("查詢失敗: %v", err),
		})
		return
	}
	
	// 解析 JSON（從輸出中提取）
	outputStr := string(output)
	
	// 找到最後一個完整的 JSON 區塊（從 { 到對應的 }）
	jsonStart := -1
	jsonEnd := -1
	braceCount := 0
	
	for i := len(outputStr) - 1; i >= 0; i-- {
		if outputStr[i] == '}' && jsonEnd == -1 {
			jsonEnd = i + 1
			braceCount = 1
		} else if jsonEnd != -1 {
			if outputStr[i] == '}' {
				braceCount++
			} else if outputStr[i] == '{' {
				braceCount--
				if braceCount == 0 {
					jsonStart = i
					break
				}
			}
		}
	}
	
	if jsonStart == -1 || jsonEnd == -1 {
		log.Printf("❌ 無法找到 JSON: %s", outputStr)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "無法解析查詢結果",
			"output": outputStr,
		})
		return
	}
	
	jsonStr := outputStr[jsonStart:jsonEnd]
	
	// 驗證 JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("❌ JSON 解析失敗: %v | JSON: %s", err, jsonStr)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "JSON 解析失敗",
		})
		return
	}
	
	// 返回結果
	w.Write([]byte(jsonStr))
	log.Printf("✅ 查詢成功: %s (%v)", symbol, result["name"])
}
