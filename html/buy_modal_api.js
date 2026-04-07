// 確認買入（使用後端 API）
async function confirmBuy() {
    const symbol = document.getElementById('buySymbol').value;
    const name = document.getElementById('buyName').value;
    const price = parseFloat(document.getElementById('buyPrice').value);
    const shares = parseInt(document.getElementById('buyShares').value);
    const reason = document.getElementById('buyReason').value || '技術面買點';
    
    if (!symbol || !price || !shares) {
        alert('❌ 請填寫完整資料');
        return;
    }
    
    const totalCost = price * shares * 1000;
    
    if (!confirm(`確認買入 ${name} (${symbol})\n` +
                 `數量: ${shares}張 (${shares * 1000}股)\n` +
                 `價格: $${price}\n` +
                 `總金額: $${totalCost.toLocaleString()}\n` +
                 `理由: ${reason}`)) {
        return;
    }
    
    try {
        // 呼叫後端 API
        const response = await fetch('http://localhost:8766/api/portfolio/buy', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                symbol: symbol,
                name: name,
                price: price,
                shares: shares,
                reason: reason
            })
        });
        
        const result = await response.json();
        
        if (result.error) {
            throw new Error(result.error);
        }
        
        if (result.success) {
            alert(`✅ ${result.message}\n\n` +
                  `持股總數: ${result.portfolio.holdings.length}檔\n` +
                  `總成本: $${result.portfolio.total_cost.toLocaleString()}\n` +
                  `總損益: $${result.portfolio.total_pnl.toLocaleString()}`);
            
            closeBuyModal();
            
            // 重新載入投資組合頁面
            if (typeof loadPortfolio === 'function') {
                loadPortfolio();
            }
        }
        
    } catch (error) {
        alert('❌ 買入失敗: ' + error.message + '\n\n' +
              '請確認投資組合 API 服務已啟動：\n' +
              'cd ~/.openclaw/workspace\n' +
              'go run portfolio_manager.go &');
    }
}
