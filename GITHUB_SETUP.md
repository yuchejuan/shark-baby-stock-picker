# 📤 GitHub 上傳指南

## ✅ 本地準備完成

- ✅ Git repository 已初始化
- ✅ 檔案已 commit（162 個檔案，23,564 行程式碼）
- ✅ README.md 已建立
- ✅ .gitignore 已設定
- ✅ ARCHITECTURE.md 架構文件已完成

---

## 🚀 上傳到 GitHub 步驟

### 方案一：使用 GitHub 網頁介面

1. **登入 GitHub**  
   前往 https://github.com

2. **建立新 Repository**  
   - 點擊右上角 `+` → `New repository`
   - Repository name: `shark-baby-stock-picker`（或你喜歡的名字）
   - Description: `🦈 台灣股市技術分析選股工具`
   - **不要勾選** "Initialize this repository with a README"
   - 點擊 `Create repository`

3. **取得 Repository URL**  
   建立完成後會看到類似：
   ```
   https://github.com/YOUR_USERNAME/shark-baby-stock-picker.git
   ```

4. **在本機執行**  
   ```bash
   cd ~/.openclaw/workspace/stock_web
   
   # 添加 remote
   git remote add origin https://github.com/YOUR_USERNAME/shark-baby-stock-picker.git
   
   # 推送到 GitHub
   git branch -M main
   git push -u origin main
   ```

5. **輸入 GitHub 憑證**  
   如果要求輸入帳號密碼，建議使用 Personal Access Token：
   - 前往 GitHub → Settings → Developer settings → Personal access tokens
   - Generate new token (classic)
   - 勾選 `repo` 權限
   - 複製 token 並在推送時貼上

---

### 方案二：使用 SSH（建議）

1. **生成 SSH Key**  
   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com"
   ```

2. **添加到 GitHub**  
   ```bash
   # 複製公鑰
   cat ~/.ssh/id_ed25519.pub
   
   # 前往 GitHub → Settings → SSH and GPG keys → New SSH key
   # 貼上公鑰
   ```

3. **建立 Repository** （同方案一步驟 2）

4. **推送**  
   ```bash
   cd ~/.openclaw/workspace/stock_web
   git remote add origin git@github.com:YOUR_USERNAME/shark-baby-stock-picker.git
   git branch -M main
   git push -u origin main
   ```

---

## 📋 Repository 設定建議

### Description
```
🦈 台灣股市技術分析選股工具 | Taiwan Stock Picker with RSI/MACD/KD Indicators
```

### Topics (Tags)
```
taiwan-stock
stock-analysis
technical-indicators
rsi
macd
kd
go
python
stock-picker
```

### README Badges
已在 README.md 中包含：
- MIT License
- Go Version

---

## 🔒 安全提醒

### 已在 .gitignore 中排除：
- ✅ 資料庫檔案 (*.db)
- ✅ 日誌檔案 (*.log)
- ✅ 敏感配置 (.env, config.json)
- ✅ 執行檔

### 請確認不要上傳：
- ❌ Telegram Bot Token
- ❌ API Keys
- ❌ 個人持倉資料
- ❌ 交易記錄

---

## 📝 後續更新

每次修改後：

```bash
cd ~/.openclaw/workspace/stock_web

# 查看變更
git status

# 加入變更
git add .

# 提交
git commit -m "描述你的變更"

# 推送到 GitHub
git push
```

---

## 🎯 完成後

GitHub Repository 連結：
```
https://github.com/YOUR_USERNAME/shark-baby-stock-picker
```

可以分享給其他人，或在 README.md 中更新實際的 clone 連結。

---

**需要我協助執行任何步驟嗎？** 🦈
