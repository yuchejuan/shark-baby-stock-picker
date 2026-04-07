package main
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type StockPoolConfig struct {
	Version     string       `json:"version"`
	Total       int          `json:"total"`
	ETF         struct {
		Count int               `json:"count"`
		List  map[string]string `json:"list"`
	} `json:"etf"`
}

func main() {
	data, err := ioutil.ReadFile("stock_pool.json")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	var pool StockPoolConfig
	if err := json.Unmarshal(data, &pool); err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	
	fmt.Printf("Version: %s, Total: %d, ETF Count: %d\n", pool.Version, pool.Total, pool.ETF.Count)
}
