/**
 * 🦈 阿哲投資記錄試算表 - 自動建立腳本
 * 
 * 使用方法：
 * 1. 開啟 Google 試算表：https://sheets.google.com
 * 2. 建立空白試算表
 * 3. 擴充功能 → Apps Script
 * 4. 刪除所有內容，貼上此腳本
 * 5. 點選「執行」→ 選擇 createInvestmentTracker
 * 6. 授權後等待完成（約 10-15 秒）
 * 7. 關閉 Apps Script，返回試算表
 * 8. 完成！✅
 */

function createInvestmentTracker() {
  var ss = SpreadsheetApp.getActiveSpreadsheet();
  
  // 顯示進度訊息
  SpreadsheetApp.getUi().alert('🦈 開始建立投資記錄表格...\n\n預計 10-15 秒完成，請稍候！');
  
  // 刪除預設工作表
  var sheets = ss.getSheets();
  if (sheets.length > 0 && sheets[0].getName() === "工作表1") {
    ss.deleteSheet(sheets[0]);
  }
  
  // 1. 建立交易明細表
  createTransactionSheet(ss);
  
  // 2. 建立配息記錄表
  createDividendSheet(ss);
  
  // 3. 建立持股總覽表
  createHoldingsSheet(ss);
  
  // 4. 建立定期定額計畫表
  createMonthlyPlanSheet(ss);
  
  // 5. 建立總覽儀表板
  createDashboardSheet(ss);
  
  // 設定預設工作表為總覽儀表板
  ss.setActiveSheet(ss.getSheetByName("總覽儀表板"));
  
  // 完成訊息
  SpreadsheetApp.getUi().alert(
    '🎉 完成！\n\n' +
    '已建立以下工作表：\n' +
    '✅ 交易明細表\n' +
    '✅ 配息記錄表\n' +
    '✅ 持股總覽表\n' +
    '✅ 定期定額計畫表\n' +
    '✅ 總覽儀表板\n\n' +
    '現在可以開始記錄交易了！🦈'
  );
}

// ========================================
// 1. 交易明細表
// ========================================
function createTransactionSheet(ss) {
  var sheet = ss.insertSheet("交易明細表");
  
  // 設定標題列
  var headers = [
    "交易日期", "交易類型", "股票代號", "股票名稱", "買入價格", "賣出價格",
    "股數", "交易金額", "手續費", "交易稅", "實際成本",
    "累積股數", "平均成本", "目前市價", "市值", "未實現損益", "報酬率(%)", "備註"
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // 標題列格式
  sheet.getRange(1, 1, 1, headers.length)
    .setFontWeight("bold")
    .setBackground("#4A86E8")
    .setFontColor("#FFFFFF")
    .setHorizontalAlignment("center");
  
  // 凍結標題列
  sheet.setFrozenRows(1);
  
  // 設定欄寬
  var widths = [100, 80, 80, 150, 80, 80, 60, 100, 80, 80, 100, 80, 80, 80, 100, 100, 80, 120];
  for (var i = 0; i < widths.length; i++) {
    sheet.setColumnWidth(i + 1, widths[i]);
  }
  
  // 輸入範例資料（第2列）
  var sampleData = [
    [
      new Date(2026, 3, 1), "買入", "0050", "元大台灣50", 81.00, "",
      100, 8100, 23, 0, 8123,
      "=SUMIF($C$2:$C, C2, $H$2:$H)",
      "=IF(L2=0, 0, SUMIF($C$2:$C, C2, $K$2:$K) / L2)",
      81.00,
      "=L2 * N2",
      "=O2 - (L2 * M2)",
      "=IF(L2=0, 0, P2 / (L2 * M2))",
      "第1批建倉"
    ]
  ];
  
  sheet.getRange(2, 1, 1, headers.length).setValues(sampleData);
  
  // 設定格式
  sheet.getRange("A2:A100").setNumberFormat("yyyy-mm-dd");
  sheet.getRange("E2:F100").setNumberFormat("0.00");
  sheet.getRange("H2:K100").setNumberFormat("$#,##0");
  sheet.getRange("M2:O100").setNumberFormat("0.00");
  sheet.getRange("Q2:Q100").setNumberFormat("0.00%");
  
  // 複製公式到下方 50 列
  sheet.getRange("L2:Q2").copyTo(sheet.getRange("L3:Q51"), SpreadsheetApp.CopyPasteType.PASTE_FORMULA);
  
  // 條件式格式（未實現損益）
  var lossRule = SpreadsheetApp.newConditionalFormatRule()
    .whenNumberLessThan(0)
    .setBackground("#F4CCCC")
    .setRanges([sheet.getRange("P2:P100")])
    .build();
  
  var profitRule = SpreadsheetApp.newConditionalFormatRule()
    .whenNumberGreaterThan(0)
    .setBackground("#D9EAD3")
    .setRanges([sheet.getRange("P2:P100")])
    .build();
  
  sheet.setConditionalFormatRules([lossRule, profitRule]);
}

// ========================================
// 2. 配息記錄表
// ========================================
function createDividendSheet(ss) {
  var sheet = ss.insertSheet("配息記錄表");
  
  // 設定標題列
  var headers = [
    "除息日", "配息類型", "股票代號", "股票名稱", "持有股數",
    "每股配息", "配息總額", "扣繳稅額", "實收金額", "累計配息", "備註"
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // 標題列格式
  sheet.getRange(1, 1, 1, headers.length)
    .setFontWeight("bold")
    .setBackground("#4A86E8")
    .setFontColor("#FFFFFF")
    .setHorizontalAlignment("center");
  
  // 凍結標題列
  sheet.setFrozenRows(1);
  
  // 設定欄寬
  sheet.setColumnWidth(1, 100);
  sheet.setColumnWidth(2, 100);
  sheet.setColumnWidth(3, 80);
  sheet.setColumnWidth(4, 150);
  sheet.setColumnWidth(5, 80);
  sheet.setColumnWidth(6, 100);
  sheet.setColumnWidth(7, 100);
  sheet.setColumnWidth(8, 100);
  sheet.setColumnWidth(9, 100);
  sheet.setColumnWidth(10, 100);
  sheet.setColumnWidth(11, 150);
  
  // 輸入公式（第2列）
  var formulas = [
    [
      new Date(2026, 7, 15), "現金股利", "00878", "國泰永續高股息", 500,
      0.60, "=E2*F2", 0, "=G2-H2", "=SUM($I$2:I2)", "第一次領息"
    ]
  ];
  
  sheet.getRange(2, 1, 1, headers.length).setValues(formulas);
  
  // 設定格式
  sheet.getRange("A2:A100").setNumberFormat("yyyy-mm-dd");
  sheet.getRange("F2:J100").setNumberFormat("$#,##0");
  
  // 複製公式到下方
  sheet.getRange("G2:J2").copyTo(sheet.getRange("G3:J51"), SpreadsheetApp.CopyPasteType.PASTE_FORMULA);
}

// ========================================
// 3. 持股總覽表
// ========================================
function createHoldingsSheet(ss) {
  var sheet = ss.insertSheet("持股總覽表");
  
  // 設定標題列
  var headers = [
    "股票代號", "股票名稱", "持有股數", "平均成本", "總成本",
    "目前市價", "目前市值", "未實現損益", "報酬率(%)", "佔比(%)",
    "目標比例(%)", "調整建議"
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // 標題列格式
  sheet.getRange(1, 1, 1, headers.length)
    .setFontWeight("bold")
    .setBackground("#4A86E8")
    .setFontColor("#FFFFFF")
    .setHorizontalAlignment("center");
  
  // 凍結標題列
  sheet.setFrozenRows(1);
  
  // 設定欄寬
  var widths = [80, 150, 80, 100, 100, 100, 100, 100, 80, 80, 100, 120];
  for (var i = 0; i < widths.length; i++) {
    sheet.setColumnWidth(i + 1, widths[i]);
  }
  
  // 輸入範例資料
  var sampleData = [
    ["0050", "元大台灣50", 300, 81.16, "=C2*D2", 81.00, "=C2*F2", "=G2-E2", "=IF(E2=0, 0, H2/E2)", "", 30.0, "持有"],
    ["00878", "國泰永續高股息", 500, 23.84, "=C3*D3", 23.70, "=C3*F3", "=G3-E3", "=IF(E3=0, 0, H3/E3)", "", 25.0, "可加碼"],
    ["00919", "群益台灣精選高息", 250, 24.66, "=C4*D4", 24.60, "=C4*F4", "=G4-E4", "=IF(E4=0, 0, H4/E4)", "", 20.0, "可加碼"],
    ["2330", "台積電", 6, 1825.33, "=C5*D5", 1820.00, "=C5*F5", "=G5-E5", "=IF(E5=0, 0, H5/E5)", "", 15.0, "可加碼"],
    ["現金", "", 0, 0, 0, 0, 29600, 0, 0, "", 0, "保留備用"]
  ];
  
  sheet.getRange(2, 1, sampleData.length, headers.length).setValues(sampleData);
  
  // 設定格式
  sheet.getRange("D2:H100").setNumberFormat("$#,##0");
  sheet.getRange("I2:J100").setNumberFormat("0.00%");
  sheet.getRange("K2:K100").setNumberFormat("0.0%");
  
  // 條件式格式
  var lossRule = SpreadsheetApp.newConditionalFormatRule()
    .whenNumberLessThan(0)
    .setBackground("#F4CCCC")
    .setRanges([sheet.getRange("H2:H100")])
    .build();
  
  var profitRule = SpreadsheetApp.newConditionalFormatRule()
    .whenNumberGreaterThan(0)
    .setBackground("#D9EAD3")
    .setRanges([sheet.getRange("H2:H100")])
    .build();
  
  sheet.setConditionalFormatRules([lossRule, profitRule]);
}

// ========================================
// 4. 定期定額計畫表
// ========================================
function createMonthlyPlanSheet(ss) {
  var sheet = ss.insertSheet("定期定額計畫表");
  
  // 設定標題列
  var headers = [
    "年月", "總投入", "0050", "00878", "00919", "2330", "2454", "其他",
    "當月累積", "累計總額", "備註"
  ];
  
  sheet.getRange(1, 1, 1, headers.length).setValues([headers]);
  
  // 標題列格式
  sheet.getRange(1, 1, 1, headers.length)
    .setFontWeight("bold")
    .setBackground("#4A86E8")
    .setFontColor("#FFFFFF")
    .setHorizontalAlignment("center");
  
  // 凍結標題列
  sheet.setFrozenRows(1);
  
  // 設定欄寬
  for (var i = 1; i <= headers.length; i++) {
    sheet.setColumnWidth(i, 100);
  }
  
  // 輸入計畫資料
  var planData = [
    ["2026-04", 80000, 24000, 20000, 16000, 12000, 0, 8000, "=SUM(B2:H2)", "=SUM($I$2:I2)", "初始建倉"],
    ["2026-05", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B3:H3)", "=SUM($I$2:I3)", "定期定額"],
    ["2026-06", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B4:H4)", "=SUM($I$2:I4)", "定期定額"],
    ["2026-07", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B5:H5)", "=SUM($I$2:I5)", "定期定額"],
    ["2026-08", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B6:H6)", "=SUM($I$2:I6)", "定期定額"],
    ["2026-09", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B7:H7)", "=SUM($I$2:I7)", "定期定額"],
    ["2026-10", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B8:H8)", "=SUM($I$2:I8)", "定期定額"],
    ["2026-11", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B9:H9)", "=SUM($I$2:I9)", "定期定額"],
    ["2026-12", 5000, 2000, 1500, 1000, 500, 0, 0, "=SUM(B10:H10)", "=SUM($I$2:I10)", "定期定額"]
  ];
  
  sheet.getRange(2, 1, planData.length, headers.length).setValues(planData);
  
  // 設定格式
  sheet.getRange("B2:J100").setNumberFormat("$#,##0");
}

// ========================================
// 5. 總覽儀表板
// ========================================
function createDashboardSheet(ss) {
  var sheet = ss.insertSheet("總覽儀表板");
  
  // 設定背景色
  sheet.getRange("A1:Z100").setBackground("#F3F3F3");
  
  // 標題
  sheet.getRange("A1:F1")
    .merge()
    .setValue("🦈 阿哲投資記錄總覽")
    .setFontSize(20)
    .setFontWeight("bold")
    .setHorizontalAlignment("center")
    .setBackground("#4A86E8")
    .setFontColor("#FFFFFF");
  
  // 卡片 1：總投入金額
  createCard(sheet, "A3", "總投入金額", "=SUM(持股總覽表!E:E)", "↑ 本月 +$5,000");
  
  // 卡片 2：目前市值
  createCard(sheet, "D3", "目前市值", "=SUM(持股總覽表!G:G)", "");
  
  // 卡片 3：累計配息
  createCard(sheet, "G3", "累計配息", "=IF(ISBLANK(MAX(配息記錄表!J:J)), 0, MAX(配息記錄表!J:J))", "");
  
  // 卡片 4：持股檔數
  createCard(sheet, "J3", "持股檔數", "=COUNTA(持股總覽表!A2:A10)-1", "85% ETF | 15% 個股");
  
  // 說明文字
  sheet.getRange("A7").setValue("📊 快速提示：");
  sheet.getRange("A8").setValue("• 記得每次交易後更新「交易明細表」");
  sheet.getRange("A9").setValue("• 每月/雙月定期定額，持之以恆");
  sheet.getRange("A10").setValue("• 每季檢視一次持股狀況");
  sheet.getRange("A11").setValue("• 退休倒數：2029 年（3 年後）");
  
  // 設定行高
  sheet.setRowHeight(1, 40);
  sheet.setRowHeight(3, 80);
}

// 建立卡片輔助函數
function createCard(sheet, startCell, title, formula, note) {
  var row = sheet.getRange(startCell).getRow();
  var col = sheet.getRange(startCell).getColumn();
  
  // 合併儲存格（3x3）
  sheet.getRange(row, col, 3, 2).merge();
  
  // 設定卡片樣式
  sheet.getRange(row, col)
    .setBackground("#FFFFFF")
    .setBorder(true, true, true, true, false, false, "#CCCCCC", SpreadsheetApp.BorderStyle.SOLID)
    .setVerticalAlignment("middle")
    .setHorizontalAlignment("center")
    .setFontSize(12);
  
  // 設定內容
  var content = title + "\n" + formula + "\n" + note;
  sheet.getRange(row, col).setValue(content);
}
