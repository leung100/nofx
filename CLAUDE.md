# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NOFX is an AI-powered agentic trading operating system supporting multiple exchanges (Binance, Hyperliquid, Aster DEX) and AI models (DeepSeek, Qwen). The system uses multi-agent competition, unified risk control, and low-latency execution for automated cryptocurrency trading.

**Key Architecture**: Go backend (Gin framework) + React 18 TypeScript frontend (Vite) + SQLite database.

## Development Commands

### Backend (Go)

```bash
# Install dependencies
go mod download

# Build
go build -o nofx

# Run backend (default port 8080)
./nofx

# Run with custom database
./nofx path/to/config.db

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Frontend (React/TypeScript)

```bash
cd web

# Install dependencies
npm install

# Development server (port 3000)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint
npm run lint
npm run lint:fix

# Format
npm run format
npm run format:check
```

### Docker Deployment

```bash
# Start all services (recommended for development)
./start.sh start --build

# View logs
./start.sh logs

# Stop services
./start.sh stop

# Restart services
./start.sh restart

# Check status
./start.sh status
```

## Architecture Overview

### Backend Structure (Go)

**Core Packages**:
- `main.go` - Entry point, initializes database, syncs config.json, starts API server and trader manager
- `api/` - RESTful API server (Gin), handles trader management, configuration, authentication
- `config/` - Database operations (SQLite), system configuration, trader/AI model/exchange CRUD
- `trader/` - Exchange implementations:
  - `interface.go` - Unified `Trader` interface for all exchanges
  - `auto_trader.go` - Core trading logic, decision loop orchestrator
  - `binance_futures.go` - Binance Futures API integration
  - `hyperliquid_trader.go` - Hyperliquid DEX integration
  - `aster_trader.go` - Aster DEX integration
- `decision/` - AI decision engine:
  - `engine.go` - Calls AI models (OpenAI-compatible APIs), parses JSON decisions
  - `prompt_manager.go` - Generates prompts from market data, account info, historical performance
- `manager/` - `TraderManager` manages multiple trader instances, handles start/stop, API queries
- `market/` - Market data:
  - `data.go` - Fetches klines, technical indicators (uses TA-Lib), calculates EMA/RSI/MACD
  - `websocket_client.go` - WebSocket streams for real-time price updates
  - `monitor.go` - Aggregates market data across traders
- `pool/` - Coin pool management (default coins, AI500 API, OI Top API)
- `news/` - News monitoring and filtering (optional):
  - `twitter_monitor.go` - Twitter monitoring via Nitter RSS (Trump, Elon, SEC, regulators)
  - `crypto_news.go` - CryptoPanic API + RSS feeds aggregation
  - `whale_alerts.go` - Whale Alert API for large transactions (>$50M)
  - `sentiment.go` - Keyword-based sentiment analysis
  - `news_context.go` - Main orchestrator, parallel fetching, formatting
  - `ai_filter.go` - Optional DeepSeek sub-agent for AI-powered news scoring (0-10)
- `logger/` - Decision logging, Telegram integration, structured logging
- `auth/` - JWT authentication, 2FA support, admin mode

**Key Flow**:
1. `main.go` loads config.json → syncs to SQLite (`config.db`)
2. `TraderManager.LoadTradersFromDatabase()` loads all traders into memory
3. Each `AutoTrader` runs an independent decision loop:
   - Fetches account balance, positions via `Trader` interface
   - Calls `market.Data` to get klines + technical indicators
   - **(Optional)** If `enable_news_monitoring` enabled: Fetches news context via `news.BuildNewsContext()`
     - Polls Twitter (Nitter RSS), CryptoPanic, Whale Alert APIs in parallel
     - Applies keyword filtering (macro-aware, time-based, deduplication)
     - **(Optional)** If `ENABLE_AI_NEWS_FILTER=true`: DeepSeek sub-agent scores news 0-10, keeps ≥7
     - Formats as bilingual markdown, injects at top of prompt
   - Generates prompt with `PromptManager` (includes news context if enabled)
   - Sends to AI via `decision.Engine`
   - Parses JSON decision (actions: open_long/short, close_long/short, hold, wait)
   - Executes trades via exchange-specific `Trader` implementation
   - Logs decision to `decision_logs/{trader_id}/`
   - Updates performance database

### Frontend Structure (React/TypeScript)

**Directory Layout**:
- `src/pages/` - Main pages:
  - `LandingPage.tsx` - Entry point (login/registration redirect)
- `src/components/` - Core components:
  - `AITradersPage.tsx` - Main trader management dashboard
  - `CompetitionPage.tsx` - Multi-agent performance comparison
  - `TraderConfigModal.tsx` - Create/edit trader configuration
  - `EquityChart.tsx` - Account equity curve visualization
  - `LoginPage.tsx`, `RegisterPage.tsx` - Authentication
- `src/hooks/` - Custom hooks (SWR for data fetching)
- `src/contexts/` - React contexts (authentication, global state)
- `src/lib/` - Utility libraries (API client, formatting)
- `src/types/` - TypeScript type definitions

**State Management**:
- **Zustand** for global state (user auth, trader selection)
- **SWR** for server state caching with 5-10s polling intervals

### Database Schema (SQLite)

**Key Tables**:
- `system_config` - System-wide settings (leverage, coin pool, API ports)
- `users` - User accounts (username, hashed password, 2FA secrets)
- `ai_models` - AI model configurations (DeepSeek, Qwen API keys)
- `exchanges` - Exchange configurations (API keys, private keys)
- `traders` - Trader instances (links AI model + exchange + initial balance)
- `performance_history` - Trade history (symbol, side, PnL, leverage, timestamps)
- `equity_history` - Account equity snapshots (for charting)
- `beta_codes` - Beta access codes (if beta_mode enabled)

### Configuration System

**Dual Configuration Approach**:
1. **config.json** (optional, legacy) - Synced to database on startup via `syncConfigToDatabase()`
2. **Database** (primary) - All runtime configuration stored in SQLite, editable via web UI

**Priority**: Database always takes precedence. `config.json` only used to seed initial values.

**News Monitoring Environment Variables** (`.env` file):
- `ENABLE_AI_NEWS_FILTER` - Enable AI-powered news filtering (default: `false`)
  - `true`: DeepSeek sub-agent scores news 0-10, keeps only score ≥7 (cost: ~$1.86/month)
  - `false`: Uses keyword filtering only (free, still filters 85-90% noise)
- `DEEPSEEK_API_KEY` - DeepSeek API key for news filtering sub-agent (separate from trading agent)
- `CRYPTOPANIC_API_KEY` - Optional CryptoPanic API key for better rate limits (free tier available)
- `NITTER_INSTANCE` - Nitter instance for Twitter monitoring (default: `nitter.net`)
- `AI_FILTER_BASE_URL` - DeepSeek API base URL (default: `https://api.deepseek.com/v1`)
- `AI_FILTER_MODEL` - Model for news filtering (default: `deepseek-chat`)

**Note**: News monitoring toggle (`enable_news_monitoring`) is per-trader and configured in UI. Environment variables control AI filtering behavior system-wide.

## Development Guidelines

### Adding a New Exchange

1. **Create trader implementation** in `trader/{exchange}_trader.go`:
   - Implement the `Trader` interface (defined in `trader/interface.go`)
   - Must implement: `GetBalance()`, `GetPositions()`, `OpenLong/Short()`, `CloseLong/Short()`, `SetLeverage()`, `FormatQuantity()`
   - Handle exchange-specific precision, order types, margin modes

2. **Register in database**:
   - Add exchange entry via API: `POST /api/exchanges`
   - Exchange ID must match switch case in `auto_trader.go:NewAutoTrader()`

3. **Update AutoTrader factory** (`auto_trader.go`):
   - Add case in `NewAutoTrader()` switch statement to instantiate your trader

4. **Add frontend icon** in `web/src/components/ExchangeIcons.tsx`

### Adding a New AI Model

1. **Ensure OpenAI-compatible API**:
   - The `decision.Engine` uses OpenAI's chat completion format
   - Custom models must expose compatible endpoints

2. **Register in database**:
   - Add AI model via API: `POST /api/models`
   - Store API key, base URL, model name

3. **Update frontend icon** in `web/src/components/ModelIcons.tsx`

### Modifying AI Decision Logic

**Prompt Engineering**:
- Prompts generated in `decision/prompt_manager.go`
- Key sections: historical performance feedback, account status, existing positions, candidate coins with market data
- AI receives raw kline sequences (not just latest values) for maximum flexibility

**Decision Parsing**:
- AI must return valid JSON array: `[{action, symbol, leverage, position_size_usd, stop_loss, take_profit, confidence, reason}]`
- Parsed in `decision/engine.go:ParseDecisions()`
- Actions: `open_long`, `open_short`, `close_long`, `close_short`, `update_stop_loss`, `update_take_profit`, `partial_close`, `hold`, `wait`

**Risk Controls** (enforced in `auto_trader.go`):
- Max position size: 1.5x equity (altcoins), 10x equity (BTC/ETH)
- Margin usage cap: 90% of account equity
- Leverage limits: Configurable per trader (default 5x altcoins, 5x BTC/ETH)
- Stop-loss/take-profit ratio: AI-controlled, recommended ≥1:2

### Working with Market Data

**Data Sources**:
- **Klines**: Fetched from exchange APIs (3min for short-term, 4h for trend)
- **Technical Indicators**: Calculated in `market/data.go` using TA-Lib bindings
- **WebSocket**: Real-time price streams in `market/websocket_client.go` (optional, for UI updates)

**Adding New Indicators**:
- Install TA-Lib: `brew install ta-lib` (macOS) or `apt-get install libta-lib0-dev` (Ubuntu)
- Import: `github.com/markcheno/go-talib`
- Add calculation in `market/data.go:FetchMarketData()`
- Update `market.Data` struct with new fields

### Working with News Monitoring

**Architecture**:
- **Two-tier filtering**: Keyword filtering (always) → Optional AI filtering (DeepSeek sub-agent)
- **Modular design**: Can disable via `ENABLE_AI_NEWS_FILTER=false` or per-trader via UI toggle
- **Graceful fallback**: If AI filter fails, falls back to keyword filtering automatically

**News Sources** (`news/` package):
- `twitter_monitor.go`: Polls Nitter RSS for tweets from Trump, Elon, SEC, regulators, exchange CEOs
- `crypto_news.go`: Polls CryptoPanic API + RSS feeds (CoinDesk, CoinTelegraph)
- `whale_alerts.go`: Polls Whale Alert API for transactions >$50M
- All sources fetched in parallel every trading cycle

**Filtering Pipeline**:
1. **Keyword filtering** (always applied):
   - Critical keywords: SEC, ETF, hack, tariffs, Fed, trade war, regulatory action
   - Time-based windows: Critical <60 min, Important <30 min
   - Deduplication by title similarity
   - Result: 3-9 items per cycle (from 20-30 raw items)

2. **AI filtering** (optional, if `ENABLE_AI_NEWS_FILTER=true`):
   - Sends filtered items to DeepSeek sub-agent for scoring (0-10)
   - Keeps only items with score ≥7
   - Understands context, sarcasm, macro signals without explicit keywords
   - Result: 3-5 highest quality items
   - Cost: ~$0.00013 per cycle (~$1.86/month)

**Integration Points**:
- `decision/engine.go:buildUserPrompt()` - Checks `ctx.EnableNewsMonitoring` flag
- If enabled: Calls `news.BuildNewsContext()` and injects formatted markdown at top of prompt
- Main trading agent receives news as formatted text (never sees scores or filtering logic)

**Adding New News Sources**:
1. Create new file in `news/` package (e.g., `reddit_monitor.go`)
2. Implement fetching function that returns slice of items with `Timestamp`, `Content`, `Sentiment`
3. Add parallel fetch in `news_context.go:BuildNewsContext()`
4. Add filtering logic (keyword or AI will auto-apply)
5. Update `Format()` to include new source section

**Modifying Filtering**:
- **Keyword thresholds**: Edit `news/twitter_monitor.go`, `crypto_news.go`, `whale_alerts.go`
- **AI score threshold**: Edit `news/ai_filter.go:55` (default: 7)
- **Time windows**: Edit filtering functions (criticalKeywords <60min, important <30min)
- **Max items**: Edit `news/news_context.go:161,177,194` (currently 3 per category)

**Testing News Monitoring**:
1. Create trader with `enable_news_monitoring` checked in UI
2. Set `ENABLE_AI_NEWS_FILTER=true` in `.env` (optional)
3. Start trader and wait for first cycle (3 min default)
4. Check `decision_logs/{trader_id}/cycle_*.txt` for news section
5. Verify news appears at top of prompt as bilingual markdown

**Troubleshooting**:
- No news appearing: Check trader has `enable_news_monitoring=true` in database
- AI filter not working: Check `.env` has `DEEPSEEK_API_KEY` set
- Rate limits: Get free API keys from CryptoPanic and Whale Alert
- Too much noise: Increase AI score threshold or tighten keyword filters

**Documentation**: See `NEWS_MONITORING_GUIDE.md` for complete user guide.

### API Development

**Adding New Endpoints** (`api/server.go`):
- Use Gin router groups: `api := r.Group("/api")`
- Authentication: Apply `AuthMiddleware()` for protected routes
- Admin-only: Check `adminMode` flag
- CORS: Pre-configured for localhost:3000

**Endpoint Conventions**:
- `GET /api/traders` - List resources
- `POST /api/traders` - Create resource
- `GET /api/traders/:id` - Get specific resource
- `PUT /api/traders/:id` - Update resource
- `DELETE /api/traders/:id` - Delete resource
- `POST /api/traders/:id/{action}` - Perform action (e.g., start, stop)

### Testing

**Manual Testing**:
- Use Postman/curl to test API endpoints directly
- Frontend: `npm run dev` + browser DevTools
- Backend: Add log statements, check `decision_logs/` for AI outputs

**Important**: The system does not currently have automated tests. When adding features:
- Test thoroughly with small capital first
- Monitor `decision_logs/{trader_id}/` for AI reasoning
- Check database integrity after config changes: `sqlite3 config.db ".schema"`

### Database Migrations

**Schema Changes**:
- Database schema defined in `config/database.go:initSchema()`
- For new tables: Add `CREATE TABLE IF NOT EXISTS` in `initSchema()`
- For new columns: Add `ALTER TABLE` logic with error handling (SQLite limitations apply)

**Migration Strategy**:
- Backup `config.db` before making changes
- Test migration on copy first
- Consider data migration scripts if changing existing table structures

## Common Pitfalls

### Exchange Precision Errors
**Problem**: "Precision is over the maximum" errors
**Solution**: Always use `trader.FormatQuantity(symbol, quantity)` before placing orders. Exchange info is cached in `auto_trader.go:exchangeInfoCache`.

### TA-Lib Installation
**Problem**: Compilation errors about missing TA-Lib
**Solution**: Install native library first:
- macOS: `brew install ta-lib`
- Ubuntu: `sudo apt-get install libta-lib0-dev`
- Cannot be installed via `go get` alone

### AI API Timeouts
**Problem**: Decision loops stalling
**Solution**:
- Default timeout: 120s (configurable via `AI_MAX_TOKENS` env var)
- Check API key validity and account balance
- Monitor AI provider status pages

### WebSocket Connection Issues
**Problem**: Market data WebSocket disconnects
**Solution**: Implemented automatic reconnection in `market/websocket_client.go`. Check exchange API limits and IP whitelisting.

### Database Locks
**Problem**: SQLite "database is locked" errors
**Solution**: SQLite uses file-level locking. Ensure:
- Not running multiple backend instances on same `config.db`
- Use `defer database.Close()` properly
- WAL mode enabled (set in `initSchema()`)

### Leverage Restrictions
**Problem**: Binance subaccounts fail with "leverage greater than 5x"
**Solution**: Check account type. Subaccounts limited to 5x leverage. Configure `btc_eth_leverage` and `altcoin_leverage` accordingly in database.

## Security Considerations

- **API Keys**: Stored in database (`config.db`). Consider encrypting in production.
- **JWT Secret**: Set via `jwt_secret` in config.json or database. Default is insecure.
- **Admin Mode**: When `admin_mode: true`, no authentication required (dev only).
- **Beta Mode**: When `beta_mode: true`, requires valid beta code for registration.
- **2FA**: Optional TOTP support. Secrets stored in `users.totp_secret` (Base32 encoded).

## Useful Debugging Commands

```bash
# Check database schema
sqlite3 config.db ".schema"

# View all traders
sqlite3 config.db "SELECT * FROM traders;"

# Check system config
sqlite3 config.db "SELECT * FROM system_config;"

# View recent performance
sqlite3 config.db "SELECT * FROM performance_history ORDER BY close_time DESC LIMIT 10;"

# Check API server health
curl http://localhost:8080/api/health

# View decision logs (latest)
ls -lt decision_logs/{trader_id}/ | head

# Monitor backend logs (Docker)
docker logs -f nofx-trading

# Check frontend API calls (browser)
# Open DevTools → Network tab → filter by /api/
```

## Project Conventions

- **Go**: Standard Go conventions, use `go fmt` before committing
- **TypeScript**: ESLint + Prettier configured, enforced via Husky pre-commit hooks
- **Commit Messages**: Conventional Commits format preferred but not enforced
- **Branch Strategy**: Feature branches merged to `dev`, releases to `main`
- **Logging**: Use `log.Printf()` for backend, structured logs for critical events
- **Error Handling**: Always check errors, log before returning/failing

## Resources

- **README.md**: Comprehensive user guide, deployment instructions, feature overview
- **docs/**: Detailed documentation (architecture, troubleshooting, roadmap)
- **CHANGELOG.md**: Version history and release notes
- **CONTRIBUTING.md**: Contribution guidelines and development workflow
- **SECURITY.md**: Security policy and vulnerability reporting
