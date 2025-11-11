# NOFX Development Setup Guide

## ✅ Completed Setup Steps

Your development environment is now fully configured:

- ✅ **Go 1.25.0** installed at `C:\Go-local\go\`
- ✅ **Backend dependencies** installed (go mod download)
- ✅ **Frontend dependencies** installed (npm install)
- ✅ **Backend build** successful (nofx.exe created - 36MB)
- ✅ **Frontend build** successful (production build in web/dist/)
- ✅ **config.json** created from template

---

## 🚀 Quick Start - Running the System

### Step 1: Get AI API Keys

You need **both DeepSeek and Qwen** API keys for competition mode.

#### DeepSeek API Key (Recommended First)

1. **Visit**: https://platform.deepseek.com
2. **Register** with email
3. **Verify** your email
4. **Top-up** your account ($5-20 USD recommended for testing)
5. **Create API Key**:
   - Go to "API Keys" section
   - Click "Create New Key"
   - Copy the key (starts with `sk-`)
   - ⚠️ **Save it immediately** - you won't see it again!

**Cost**: ~$0.14 per 1M tokens (very cheap for testing)

#### Qwen API Key (Alibaba Cloud)

1. **Visit**: https://dashscope.console.aliyun.com
2. **Register** with Alibaba Cloud account
3. **Enable DashScope** service
4. **Create API Key**:
   - Go to "API Key Management"
   - Create new key
   - Copy and save (starts with `sk-`)

**Note**: May require Chinese phone number for registration

---

### Step 2: Start the Backend

Open a terminal in the project root and run:

```bash
# Add Go to PATH (required for each new terminal session)
export PATH="/c/Go-local/go/bin:$PATH"

# Start the backend
./nofx.exe
```

**What you should see**:
```
╔════════════════════════════════════════════════════════════╗
║    🤖 AI多模型交易系统 - 支持 DeepSeek & Qwen            ║
╚════════════════════════════════════════════════════════════╝

📋 初始化配置数据库: config.db
✓ 配置数据库初始化成功

🤖 数据库中的AI交易员配置:
  • 暂无配置的交易员，请通过Web界面创建

🌐 API服务器启动在 http://localhost:8080
```

✅ **Backend is running!** Keep this terminal open.

---

### Step 3: Start the Frontend

Open a **NEW terminal window** (keep the backend running):

```bash
cd web
npm run dev
```

**What you should see**:
```
VITE v6.x.x  ready in xxx ms

➜  Local:   http://localhost:3000/
➜  Network: use --host to expose
```

✅ **Frontend is running!** Keep this terminal open too.

---

### Step 4: Access the Web Interface

Open your browser and visit: **http://localhost:3000**

You'll see the NOFX landing page with login/registration options.

**First Time Setup**:
1. The system runs in **admin mode** by default (no login required)
2. You'll be automatically logged in as admin user
3. You can start configuring AI models and exchanges immediately

---

## 📋 Configure AI Models & Exchanges

### Configure AI Models (Required First)

1. In the web interface, find and click **"AI模型配置"** (AI Model Configuration)
2. **Enable DeepSeek**:
   - Toggle "Enable DeepSeek" to ON
   - Enter your DeepSeek API key
   - Click "Save"
3. **Enable Qwen**:
   - Toggle "Enable Qwen" to ON
   - Enter your Qwen API key
   - Click "Save"

✅ Both AI models are now configured!

---

### Configure Exchanges (Easy Toggle Setup)

NOFX supports **easy exchange switching**. You can configure all three exchanges and toggle between them per trader.

#### Option 1: Binance Futures (Most Popular)

**Prerequisites**:
- Binance account with KYC completed
- Futures trading enabled
- API key with "Futures" permission

**Setup**:
1. Visit Binance → Account → API Management
2. Create new API key with **"Enable Futures"** permission
3. Whitelist your IP address for security
4. Copy API Key and Secret Key

**Configure in NOFX**:
1. Click **"交易所配置"** (Exchange Configuration)
2. Enable **Binance**
3. Enter:
   - API Key: `your_binance_api_key`
   - Secret Key: `your_binance_secret_key`
4. Click "Save"

**Register Binance**: https://www.binance.com/join?ref=TINKLEVIP (fee discount link)

---

#### Option 2: Hyperliquid (Decentralized)

**Prerequisites**:
- Ethereum wallet (MetaMask, etc.)
- Funds deposited on Hyperliquid

**Setup**:
1. Open MetaMask (or your wallet)
2. Export your private key
3. **Remove the `0x` prefix** from the key
4. Fund your wallet at https://hyperliquid.xyz

**Configure in NOFX**:
1. Click **"交易所配置"** (Exchange Configuration)
2. Enable **Hyperliquid**
3. Enter:
   - Private Key: `your_private_key_without_0x`
   - Wallet Address: `0xYourEthereumAddress`
   - Testnet: `false` (or `true` for testing)
4. Click "Save"

⚠️ **Security**: Use a dedicated wallet for trading, not your main wallet!

---

#### Option 3: Aster DEX (Binance-Compatible)

**Prerequisites**:
- Aster DEX account
- API wallet created

**Setup**:
1. Register: https://www.asterdex.com/en/referral/fdfc0e (referral link for fee discount)
2. Visit: https://www.asterdex.com/en/api-wallet
3. Connect your main wallet
4. Click "Create API Wallet"
5. **Save immediately**:
   - User address (your main wallet)
   - Signer address (API wallet)
   - Private key (shown only once!)

**Configure in NOFX**:
1. Click **"交易所配置"** (Exchange Configuration)
2. Enable **Aster**
3. Enter:
   - User Address: `0xYourMainWallet`
   - Signer Address: `0xAPIWalletAddress`
   - Private Key: `api_wallet_private_key_without_0x`
4. Click "Save"

---

## 🤖 Create Your First Trader

Now that AI models and exchanges are configured, create a trader:

1. Click **"创建交易员"** (Create Trader)
2. Fill in:
   - **Name**: e.g., "DeepSeek + Binance Trader"
   - **AI Model**: Select "DeepSeek" (or "Qwen")
   - **Exchange**: Select "Binance" (or "Hyperliquid" or "Aster")
   - **Initial Balance**: e.g., `1000.0` (for tracking purposes)
   - **Scan Interval**: `3` minutes (default, recommended)
3. Click **"Create"**

✅ Your trader is created!

---

## 🔄 Easy Exchange Switching

**The key feature you wanted**: Each trader can use a different exchange!

### Create Multiple Traders with Different Exchanges:

**Example Setup**:
```
Trader 1: DeepSeek + Binance
Trader 2: DeepSeek + Hyperliquid
Trader 3: Qwen + Binance
Trader 4: Qwen + Aster
```

**To switch exchanges for a trader**:
1. Stop the trader if it's running
2. Click on the trader card
3. Edit the exchange selection
4. Save changes
5. Restart the trader

**No code changes needed!** All configuration is done through the web UI.

---

## ▶️ Start Trading

1. Find your trader card in the dashboard
2. Click the **"Start"** button
3. Monitor the trader's decisions in real-time

**What happens next**:
- Every 3 minutes (or your configured interval), the AI analyzes the market
- AI decisions appear in the "Decision Logs" section
- You can expand each decision to see the full Chain of Thought reasoning
- Trades are executed automatically based on AI decisions
- Account balance and positions update in real-time

---

## 🛡️ Risk Management (Important!)

**Default Settings** (configured in config.json):
- **Max Daily Loss**: 10% of account equity
- **Max Drawdown**: 20% from peak equity
- **Leverage**:
  - BTC/ETH: 5x maximum (safe for subaccounts)
  - Altcoins: 5x maximum (safe for subaccounts)

**For Binance Subaccounts**:
⚠️ Subaccounts are **limited to 5x leverage** by Binance. The default config (5x) is safe.

**For Main Accounts**:
You can increase leverage in config.json:
```json
"leverage": {
  "btc_eth_leverage": 20,
  "altcoin_leverage": 15
}
```

**Recommended for Testing**:
- Start with **100-500 USDT**
- Use **default leverage (5x)**
- Monitor closely for the first few hours
- Review AI decision logs to understand strategy

---

## 🎯 Competition Mode (DeepSeek vs Qwen)

You configured both AI models, so you can run competition mode:

1. Create **Trader 1**: DeepSeek + Binance (or any exchange)
2. Create **Trader 2**: Qwen + Binance (or same/different exchange)
3. Start both traders
4. Visit the **"Competition"** page in the web UI
5. Watch them compete in real-time!

**Competition Features**:
- Live ROI comparison charts
- Head-to-head performance metrics
- Win rate statistics
- Best/worst trades comparison

---

## 📁 Important File Locations

```
C:\Users\leung100\Desktop\AI-TRADERS\
├── config.json              # System configuration (synced to database)
├── config.db                # SQLite database (primary config storage)
├── nofx.exe                 # Backend executable
├── decision_logs/           # AI decision logs (per trader)
│   └── {trader_id}/
│       └── decision_*.json
├── web/
│   ├── dist/                # Frontend production build
│   └── src/                 # Frontend source code
└── SETUP_GUIDE.md          # This file!
```

---

## 🔧 Development Commands (Quick Reference)

### Backend
```bash
# Add Go to PATH
export PATH="/c/Go-local/go/bin:$PATH"

# Build backend
go build -o nofx.exe

# Run backend
./nofx.exe

# Format code
go fmt ./...
```

### Frontend
```bash
cd web

# Development server
npm run dev

# Production build
npm run build

# Lint
npm run lint
```

---

## 🐛 Troubleshooting

### Backend won't start
- **Check**: Is port 8080 already in use?
- **Solution**: Change `api_server_port` in config.json to 8081 or another port

### Frontend can't connect to backend
- **Check**: Is backend running? Visit http://localhost:8080/api/health
- **Solution**: Ensure backend is running before starting frontend

### Go command not found
- **Solution**: Run `export PATH="/c/Go-local/go/bin:$PATH"` in each new terminal

### AI API timeout
- **Check**: Is your API key valid? Do you have sufficient balance?
- **Solution**: Verify on DeepSeek/Qwen platform, check network connection

### Exchange API errors
- **Check**: Are API keys correct? Is IP whitelisted?
- **Solution**: Verify exchange configuration, check API key permissions

---

## 📚 Next Steps

1. **Get AI API keys** (DeepSeek + Qwen)
2. **Choose an exchange** and create account/API keys
3. **Configure AI models** in web UI
4. **Configure exchange** in web UI
5. **Create a trader**
6. **Start small** (100-500 USDT for testing)
7. **Monitor decisions** closely for first few hours
8. **Review CLAUDE.md** for deeper technical understanding

---

## 🆘 Need Help?

- **GitHub Issues**: https://github.com/tinkle-community/nofx/issues
- **Telegram Developer Community**: https://t.me/nofx_dev_community
- **Documentation**: Check `docs/` folder for detailed guides
- **CLAUDE.md**: Technical architecture and development guidelines

---

**🎉 Your development environment is ready! Start by getting your AI API keys, then configure everything through the web interface.**
