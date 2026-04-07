// 🦈 分批次持倉載入 - 覆蓋原有的 loadPortfolio 函數

async function loadPortfolio() {
    try {
        // 從交易資料庫讀取分批次持股
        const batchesResponse = await fetch('http://localhost:8888/api/holdings/batches');
        const batches = await batchesResponse.json();
        
        // 計算總覽統計
        let totalCost = 0;
        let totalValue = 0;
        
        batches.forEach(batch => {
            totalCost += batch.shares * batch.buy_price;
            totalValue += batch.current_value;
        });
        
        const totalPnL = totalValue - totalCost;
        const totalReturn = totalCost > 0 ? (totalPnL / totalCost) * 100 : 0;
        
        // 更新總覽
        document.getElementById('total-cost').textContent = 
            '$ ' + Math.round(totalCost).toLocaleString();
        document.getElementById('total-value').textContent = 
            '$ ' + Math.round(totalValue).toLocaleString();
        
        const pnlElement = document.getElementById('total-pnl');
        const returnElement = document.getElementById('total-return');
        
        pnlElement.textContent = (totalPnL >= 0 ? '+' : '') + 
            '$ ' + Math.round(totalPnL).toLocaleString();
        pnlElement.className = 'stat-value ' + 
            (totalPnL >= 0 ? 'positive' : 'negative');
        
        returnElement.textContent = (totalReturn >= 0 ? '+' : '') + 
            totalReturn.toFixed(2) + '%';
        returnElement.className = 'stat-value ' + 
            (totalReturn >= 0 ? 'positive' : 'negative');
        
        // 更新持倉表格（分批次顯示）
        const tbody = document.getElementById('portfolio-body');
        if (batches.length === 0) {
            tbody.innerHTML = '<tr><td colspan="11" style="text-align:center;padding:20px;">🎯 尚無持股<br>請從「🏆 V3.0 TOP 5」或「📊 每日選股」點擊「📈 買入」開始模擬交易！</td></tr>';
        } else {
            tbody.innerHTML = batches.map(batch => {
                const pnlClass = batch.profit_loss >= 0 ? 'positive' : 'negative';
                const buyDate = new Date(batch.buy_date).toLocaleDateString('zh-TW');
                
                return `
                    <tr>
                        <td><strong>${batch.symbol}</strong></td>
                        <td>${batch.name}</td>
                        <td>${batch.shares.toLocaleString()}</td>
                        <td>$ ${batch.buy_price.toFixed(2)}</td>
                        <td>$ ${batch.current_price.toFixed(2)}</td>
                        <td class="${pnlClass}">
                            ${batch.profit_loss >= 0 ? '+' : ''}$ ${Math.round(batch.profit_loss).toLocaleString()}
                        </td>
                        <td class="${pnlClass}">
                            ${batch.return_pct >= 0 ? '+' : ''}${batch.return_pct.toFixed(2)}%
                        </td>
                        <td style="white-space:nowrap;">${buyDate}</td>
                        <td>${batch.days_held} 天</td>
                        <td style="font-size:0.85em;">${batch.reason || '-'}</td>
                        <td>
                            <button onclick="simulateSell('${batch.symbol}', '${batch.name}', ${batch.current_price}, ${batch.shares})" 
                                    style="background:#e74c3c;color:white;border:none;padding:8px 15px;border-radius:5px;cursor:pointer;font-weight:bold;">
                                📉 賣出
                            </button>
                        </td>
                    </tr>
                `;
            }).join('');
        }
        
        // 註：update-time 元素不存在於持倉明細區塊，已移除
            
    } catch (error) {
        console.error('載入分批次持股失敗:', error);
        // 降級到舊版 portfolio.json
        loadPortfolioLegacy();
    }
}

// 降級方案：使用舊版 portfolio.json
async function loadPortfolioLegacy() {
    console.log('使用降級方案: portfolio.json');
    try {
        const response = await fetch('portfolio.json');
        const data = await response.json();
        
        document.getElementById('total-cost').textContent = 
            '$ ' + data.total_cost.toLocaleString();
        document.getElementById('total-value').textContent = 
            '$ ' + data.current_value.toLocaleString();
        
        const pnlElement = document.getElementById('total-pnl');
        const returnElement = document.getElementById('total-return');
        
        pnlElement.textContent = (data.total_pnl >= 0 ? '+' : '') + 
            '$ ' + data.total_pnl.toLocaleString();
        pnlElement.className = 'stat-value ' + 
            (data.total_pnl >= 0 ? 'positive' : 'negative');
        
        returnElement.textContent = (data.total_return >= 0 ? '+' : '') + 
            data.total_return.toFixed(2) + '%';
        returnElement.className = 'stat-value ' + 
            (data.total_return >= 0 ? 'positive' : 'negative');
        
        const tbody = document.getElementById('portfolio-body');
        tbody.innerHTML = data.holdings.map(stock => {
            const pnlClass = stock.profit_loss >= 0 ? 'positive' : 'negative';
            return `
                <tr>
                    <td><strong>${stock.symbol}</strong></td>
                    <td>${stock.name}</td>
                    <td>${stock.shares.toLocaleString()}</td>
                    <td>$ ${stock.buy_price.toFixed(2)}</td>
                    <td>$ ${stock.current_price.toFixed(2)}</td>
                    <td class="${pnlClass}">
                        ${stock.profit_loss >= 0 ? '+' : ''}$ ${stock.profit_loss.toLocaleString()}
                    </td>
                    <td class="${pnlClass}">
                        ${stock.return_pct >= 0 ? '+' : ''}${stock.return_pct.toFixed(2)}%
                    </td>
                    <td colspan="2">2026-03-13</td>
                    <td style="font-size:0.85em;">${stock.reason}</td>
                    <td>
                        <button onclick="simulateSell('${stock.symbol}', '${stock.name}', ${stock.current_price}, ${stock.shares})" 
                                style="background:#e74c3c;color:white;border:none;padding:8px 15px;border-radius:5px;cursor:pointer;font-weight:bold;">
                            📉 賣出
                        </button>
                    </td>
                </tr>
            `;
        }).join('');
        
        // 註：update-time 元素不存在於持倉明細區塊，已移除
    } catch (error) {
        console.error('載入舊版投資組合也失敗:', error);
        document.getElementById('portfolio-body').innerHTML = 
            '<tr><td colspan="11" style="text-align:center;color:red;">載入失敗，請確認 API 服務正在運行</td></tr>';
    }
}

// 賣出功能（呼叫主頁面的 sellStock 函數或直接實作）
async function simulateSell(symbol, name, price, maxShares) {
    // 檢查是否有主頁面的 sellStock 函數
    if (typeof sellStock === 'function') {
        return sellStock(symbol, name, price, maxShares);
    }
    
    // 獨立實作賣出邏輯
    const shares = prompt(`請輸入要賣出的股數（目前持有 ${maxShares} 股）:`, maxShares);
    
    if (!shares || shares <= 0) return;
    if (parseInt(shares) > maxShares) {
        alert('❌ 賣出股數不能超過持有股數');
        return;
    }
    
    if (!confirm(`確定要賣出 ${name} (${symbol})？\n\n股數: ${shares}\n價格: $${price}\n總金額: $${(shares * price).toLocaleString()}`)) {
        return;
    }
    
    const trade = {
        type: 'sell',
        symbol: symbol,
        name: name,
        shares: parseInt(shares),
        price: price,
        date: new Date().toISOString(),
        note: '模擬賣出'
    };
    
    try {
        const response = await fetch('http://localhost:8888/api/trade/add', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(trade)
        });
        
        const result = await response.json();
        
        if (result.success) {
            alert(`✅ 賣出成功！\n\n${name} (${symbol})\n股數: ${shares}\n總金額: $${(shares * price).toLocaleString()}`);
            // 重新載入持倉資料
            loadPortfolio();
        } else {
            alert('❌ 賣出失敗: ' + result.message);
        }
    } catch (error) {
        console.error('賣出交易失敗:', error);
        alert('❌ 連線錯誤: ' + error.message);
    }
}

console.log('🦈 分批次持倉模組已載入');
