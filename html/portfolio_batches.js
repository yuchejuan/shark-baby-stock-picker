// 🦈 分批次持倉載入 - 覆蓋原有的 loadPortfolio 函數

// [FIX] 相容 index.html 的 CONFIG，若不存在則用預設值
const _API_BASE_BATCHES = (typeof CONFIG !== 'undefined') ? CONFIG.TRADE_API : 'http://localhost:8888/api';

async function loadPortfolio() {
    const tbody = document.getElementById('portfolio-body');
    try {
        const batchesResponse = await fetch(`${_API_BASE_BATCHES}/holdings/batches`);
        const batches = await batchesResponse.json();

        let totalCost = 0, totalValue = 0;
        batches.forEach(batch => {
            totalCost  += batch.shares * batch.buy_price;
            totalValue += batch.current_value;
        });

        const totalPnL    = totalValue - totalCost;
        const totalReturn = totalCost > 0 ? (totalPnL / totalCost) * 100 : 0;

        document.getElementById('total-cost').textContent  = '$ ' + Math.round(totalCost).toLocaleString();
        document.getElementById('total-value').textContent = '$ ' + Math.round(totalValue).toLocaleString();

        const pnlElement    = document.getElementById('total-pnl');
        const returnElement = document.getElementById('total-return');

        pnlElement.textContent  = (totalPnL >= 0 ? '+' : '') + '$ ' + Math.round(totalPnL).toLocaleString();
        pnlElement.className    = (totalPnL >= 0 ? 'positive' : 'negative');

        returnElement.textContent = (totalReturn >= 0 ? '+' : '') + totalReturn.toFixed(2) + '%';
        returnElement.className   = (totalReturn >= 0 ? 'positive' : 'negative');

        if (batches.length === 0) {
            tbody.innerHTML = '<tr><td colspan="11" style="text-align:center;padding:20px;">🎯 尚無持股<br>請從「🏆 TOP 推薦」點擊「📈 買入」開始模擬交易！</td></tr>';
        } else {
            tbody.innerHTML = batches.map(batch => {
                const pnlClass = batch.profit_loss >= 0 ? 'positive' : 'negative';
                const buyDate  = new Date(batch.buy_date).toLocaleDateString('zh-TW');
                return `
                    <tr>
                        <td><strong>${batch.symbol}</strong></td>
                        <td>${batch.name}</td>
                        <td>${batch.shares.toLocaleString()}</td>
                        <td>$ ${batch.buy_price.toFixed(2)}</td>
                        <td>$ ${batch.current_price.toFixed(2)}</td>
                        <td class="${pnlClass}">${batch.profit_loss >= 0 ? '+' : ''}$ ${Math.round(batch.profit_loss).toLocaleString()}</td>
                        <td class="${pnlClass}">${batch.return_pct >= 0 ? '+' : ''}${batch.return_pct.toFixed(2)}%</td>
                        <td style="white-space:nowrap;">${buyDate}</td>
                        <td>${batch.days_held} 天</td>
                        <td style="font-size:0.85em;">${batch.reason || '-'}</td>
                        <td>
                            <button onclick="simulateSell('${batch.symbol}', '${batch.name}', ${batch.current_price}, ${batch.shares})"
                                    class="btn btn-sell">📉 賣出</button>
                        </td>
                    </tr>
                `;
            }).join('');
        }
    } catch (error) {
        console.error('載入分批次持股失敗:', error);
        if (tbody) tbody.innerHTML = `<tr><td colspan="11" style="text-align:center;padding:20px;color:#e74c3c;">
            <strong>❌ 載入失敗</strong><br>${error.message}<br>
            <span style="color:#888;font-size:0.9em;">嘗試切換到備用方案...</span>
        </td></tr>`;
        loadPortfolioLegacy();
    }
}

async function loadPortfolioLegacy() {
    const tbody = document.getElementById('portfolio-body');
    console.log('使用降級方案: portfolio.json');
    try {
        const response = await fetch('portfolio.json');
        const data = await response.json();

        document.getElementById('total-cost').textContent  = '$ ' + data.total_cost.toLocaleString();
        document.getElementById('total-value').textContent = '$ ' + data.current_value.toLocaleString();

        const pnlElement    = document.getElementById('total-pnl');
        const returnElement = document.getElementById('total-return');

        pnlElement.textContent  = (data.total_pnl >= 0 ? '+' : '') + '$ ' + data.total_pnl.toLocaleString();
        pnlElement.className    = (data.total_pnl >= 0 ? 'positive' : 'negative');
        returnElement.textContent = (data.total_return >= 0 ? '+' : '') + data.total_return.toFixed(2) + '%';
        returnElement.className   = (data.total_return >= 0 ? 'positive' : 'negative');

        tbody.innerHTML = data.holdings.map(stock => {
            const pnlClass = stock.profit_loss >= 0 ? 'positive' : 'negative';
            return `
                <tr>
                    <td><strong>${stock.symbol}</strong></td><td>${stock.name}</td>
                    <td>${stock.shares.toLocaleString()}</td>
                    <td>$ ${stock.buy_price.toFixed(2)}</td><td>$ ${stock.current_price.toFixed(2)}</td>
                    <td class="${pnlClass}">${stock.profit_loss >= 0 ? '+' : ''}$ ${stock.profit_loss.toLocaleString()}</td>
                    <td class="${pnlClass}">${stock.return_pct >= 0 ? '+' : ''}${stock.return_pct.toFixed(2)}%</td>
                    <td colspan="2">-</td>
                    <td style="font-size:0.85em;">${stock.reason}</td>
                    <td>
                        <button onclick="simulateSell('${stock.symbol}', '${stock.name}', ${stock.current_price}, ${stock.shares})"
                                class="btn btn-sell">📉 賣出</button>
                    </td>
                </tr>
            `;
        }).join('');
    } catch (error) {
        console.error('載入舊版投資組合也失敗:', error);
        if (tbody) tbody.innerHTML = '<tr><td colspan="11" style="text-align:center;color:#e74c3c;padding:20px;">❌ 載入失敗，請確認 API 服務正在運行</td></tr>';
    }
}

// [FIX] 賣出：優先使用新版 showSellModal，其次 sellStock，最後才 prompt()
async function simulateSell(symbol, name, price, maxShares) {
    if (typeof showSellModal === 'function') return showSellModal(symbol, name, price, maxShares);
    if (typeof sellStock === 'function')    return sellStock(symbol, name, price, maxShares);

    // fallback
    const shares = prompt(`請輸入要賣出的股數（目前持有 ${maxShares} 股）:`, maxShares);
    if (!shares || shares <= 0) return;
    if (parseInt(shares) > maxShares) { alert('❌ 賣出股數不能超過持有股數'); return; }
    if (!confirm(`確定要賣出 ${name} (${symbol})？\n股數: ${shares}\n價格: $${price}\n總金額: $${(shares * price).toLocaleString()}`)) return;

    try {
        const response = await fetch(`${_API_BASE_BATCHES}/trade/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type: 'sell', symbol, name, shares: parseInt(shares), price, date: new Date().toISOString(), note: '模擬賣出' })
        });
        const result = await response.json();
        if (result.success) {
            alert(`✅ 賣出成功！\n${name} (${symbol})\n股數: ${shares}\n總金額: $${(shares * price).toLocaleString()}`);
            loadPortfolio();
        } else {
            alert('❌ 賣出失敗: ' + result.message);
        }
    } catch (error) {
        alert('❌ 連線錯誤: ' + error.message);
    }
}

console.log('🦈 分批次持倉模組已載入');
