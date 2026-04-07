// 🦈 模擬交易系統 JavaScript

// [FIX] 避免與 index.html CONFIG 衝突，改用相容寫法
const _SIM_API = (typeof CONFIG !== 'undefined') ? CONFIG.TRADE_API : 'http://localhost:8888/api';

async function simulateBuy(symbol, name, price, shares = 1000) {
    if (!confirm(`確定要模擬買入 ${name} (${symbol})？\n\n股數: ${shares}\n價格: $${price}\n總金額: $${(shares * price).toLocaleString()}`)) return;

    const trade = {
        type: 'buy', symbol, name,
        shares: parseInt(shares),
        price: parseFloat(price),
        date: new Date().toISOString(),
        note: '模擬買入（來自每日選股）'
    };

    try {
        const response = await fetch(`${_SIM_API}/trade/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(trade)
        });
        const result = await response.json();
        if (result.success) {
            alert(`✅ 模擬買入成功！\n\n${name} (${symbol})\n股數: ${shares}\n價格: $${price}\n總金額: $${(shares * price).toLocaleString()}`);
            if (typeof loadPortfolio === 'function') loadPortfolio();
        } else {
            alert('❌ 買入失敗: ' + (result.message || '未知錯誤'));
        }
    } catch (error) {
        alert('❌ 連線錯誤: ' + error.message);
    }
}

// [FIX] 賣出：優先使用新版 showSellModal，其次 sellStock，最後才 prompt()
async function simulateSell(symbol, name, price, shares) {
    if (typeof showSellModal === 'function') return showSellModal(symbol, name, price, parseInt(shares));
    if (typeof sellStock === 'function')    return sellStock(symbol, name, price, parseInt(shares));

    // fallback
    const currentShares = parseInt(shares);
    const sellShares = prompt(`請輸入要賣出的股數（目前持有 ${currentShares} 股）:`, currentShares);
    if (!sellShares || sellShares <= 0) return;
    if (parseInt(sellShares) > currentShares) { alert('❌ 賣出股數不能超過持有股數！'); return; }
    if (!confirm(`確定要模擬賣出 ${name} (${symbol})？\n\n股數: ${sellShares}\n價格: $${price}\n總金額: $${(sellShares * price).toLocaleString()}`)) return;

    try {
        const response = await fetch(`${_SIM_API}/trade/add`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ type: 'sell', symbol, name, shares: parseInt(sellShares), price: parseFloat(price), date: new Date().toISOString(), note: '模擬賣出' })
        });
        const result = await response.json();
        if (result.success) {
            const profit = (price * sellShares) - (result.avg_cost || price) * sellShares;
            const profitText = profit >= 0 ? `獲利 $${profit.toLocaleString()}` : `虧損 $${Math.abs(profit).toLocaleString()}`;
            alert(`✅ 模擬賣出成功！\n\n${name} (${symbol})\n股數: ${sellShares}\n價格: $${price}\n${profitText}`);
            if (typeof loadPortfolio === 'function') loadPortfolio();
        } else {
            alert('❌ 賣出失敗: ' + (result.message || '未知錯誤'));
        }
    } catch (error) {
        alert('❌ 連線錯誤: ' + error.message);
    }
}

function quickBuy(symbol, name, price) {
    const shares = prompt(`快速買入 ${name} (${symbol})\n當前價格: $${price}\n\n請輸入股數:`, '1000');
    if (!shares || shares <= 0) return;
    simulateBuy(symbol, name, price, parseInt(shares));
}

function viewTradeHistory(symbol) {
    window.open(`trade.html?tab=history&symbol=${symbol}`, '_blank');
}

async function loadPerformanceStats() {
    try {
        const response = await fetch(`${_SIM_API}/stats?period=1year`);
        return await response.json();
    } catch (error) {
        console.error('載入績效統計失敗:', error);
        return null;
    }
}

async function exportTrades() {
    try {
        const response = await fetch(`${_SIM_API}/trades`);
        const trades = await response.json();
        if (!trades || trades.length === 0) { alert('❌ 沒有交易記錄可匯出'); return; }

        let csv = 'ID,日期,類型,代號,名稱,股數,價格,金額,備註\n';
        trades.forEach(t => {
            const date = new Date(t.date).toLocaleDateString('zh-TW');
            const type = t.type === 'buy' ? '買入' : '賣出';
            csv += `${t.id},${date},${type},${t.symbol},${t.name},${t.shares},${t.price},${t.amount},"${t.note || ''}"\n`;
        });

        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
        const link = document.createElement('a');
        link.href     = URL.createObjectURL(blob);
        link.download = `trades_${new Date().toISOString().split('T')[0]}.csv`;
        link.click();
        alert('✅ 交易記錄已匯出！');
    } catch (error) {
        alert('❌ 匯出失敗: ' + error.message);
    }
}

async function showPerformanceReport() {
    const stats = await loadPerformanceStats();
    if (!stats) { alert('❌ 無法載入績效統計'); return; }

    alert(`📊 模擬交易績效報告（1年）\n\n總交易次數: ${stats.total_trades || 0}\n獲利交易: ${stats.win_trades || 0}\n虧損交易: ${stats.loss_trades || 0}\n勝率: ${(stats.win_rate || 0).toFixed(1)}%\n\n已實現損益: $${(stats.total_realized || 0).toLocaleString()}\n未實現損益: $${(stats.total_unrealized || 0).toLocaleString()}\n總損益: $${(stats.total_profit || 0).toLocaleString()}\n\n最佳交易: ${stats.best_trade ? stats.best_trade.name : 'N/A'}\n最差交易: ${stats.worst_trade ? stats.worst_trade.name : 'N/A'}`);
}

function initSimulationSystem() {
    console.log('🦈 模擬交易系統已載入');
    fetch(`${_SIM_API}/trades`)
        .then(r => console.log(r.ok ? '✅ API 連線正常' : '⚠️ API 連線異常'))
        .catch(e => console.error('❌ API 無法連線:', e));
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initSimulationSystem);
} else {
    initSimulationSystem();
}
