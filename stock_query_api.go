// stock_query_api.go - 股票查詢 HTTP API（Port 8765）
// 呼叫預先編譯好的 stock_query_cli 執行查詢
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var cliPath string // stock_query_cli 的絕對路徑

func main() {
	// 找到執行目錄
	exeDir, err := os.Getwd()
	if err != nil {
		log.Fatal("無法取得工作目錄:", err)
	}

	// stock_query_cli 在同一目錄
	cliPath = filepath.Join(exeDir, "stock_query_cli")

	// 確認 cli 存在
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		log.Fatalf("找不到 stock_query_cli（%s）\n請先執行：go build -o stock_query_cli stock_query_service.go", cliPath)
	}

	http.HandleFunc("/api/query", handleStockQuery)
	http.HandleFunc("/health", handleHealth)

	port := "8765"
	fmt.Printf("🦈 股票查詢 API 啟動於 http://localhost:%s\n", port)
	fmt.Printf("📡 CLI 路徑：%s\n", cliPath)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format("2006-01-02 15:04:05"),
	})
}

func handleStockQuery(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "缺少股票代號"})
		return
	}

	log.Printf("📊 查詢股票: %s", symbol)

	// 呼叫預編譯的 stock_query_cli
	cmd := exec.Command(cliPath, symbol)
	cmd.Dir = filepath.Dir(cliPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("❌ CLI 失敗: %v\n輸出: %s", err, string(output))
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("查詢失敗: %v", err),
		})
		return
	}

	// 從輸出中取最後一個 JSON 區塊
	outputStr := string(output)
	jsonStr := extractJSON(outputStr)
	if jsonStr == "" {
		log.Printf("❌ 無法解析 JSON，輸出：%s", outputStr)
		json.NewEncoder(w).Encode(map[string]string{"error": "無法解析查詢結果"})
		return
	}

	// 驗證並轉發 JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON 解析失敗"})
		return
	}

	w.Write([]byte(jsonStr))
	log.Printf("✅ 查詢成功: %s", symbol)
}

// 從輸出中提取最後一個完整 JSON 物件
func extractJSON(output string) string {
	// 找最後一個 { 到對應 }
	end := strings.LastIndex(output, "}")
	if end == -1 {
		return ""
	}
	// 從 end 往前找對應的 {
	depth := 0
	for i := end; i >= 0; i-- {
		if output[i] == '}' {
			depth++
		} else if output[i] == '{' {
			depth--
			if depth == 0 {
				return strings.TrimSpace(output[i : end+1])
			}
		}
	}
	return ""
}
