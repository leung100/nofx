#!/bin/bash

# News Monitoring Integration - Automated Patch Script
# This script applies all necessary changes for the news monitoring feature

set -e

echo "🚀 Applying News Monitoring Integration..."
echo ""

# Backup files before modification
echo "📦 Creating backups..."
cp config/database.go config/database.go.backup
cp trader/auto_trader.go trader/auto_trader.go.backup
cp decision/engine.go decision/engine.go.backup
cp api/server.go api/server.go.backup
cp web/src/types.ts web/src/types.ts.backup
cp web/src/components/TraderConfigModal.tsx web/src/components/TraderConfigModal.tsx.backup
echo "✅ Backups created (.backup files)"
echo ""

# 1. Update config/database.go
echo "1️⃣ Updating config/database.go..."

# Add ALTER TABLE statement
sed -i "/ALTER TABLE traders ADD COLUMN system_prompt_template/a\\		\`ALTER TABLE traders ADD COLUMN enable_news_monitoring BOOLEAN DEFAULT 0\`,      // 是否启用新闻监控" config/database.go

# Add field to TraderRecord struct
sed -i "/IsCrossMargin.*bool.*json:\"is_cross_margin\"/a\\	EnableNewsMonitoring bool      \`json:\"enable_news_monitoring\"\` // 是否启用新闻监控" config/database.go

# Update CreateTrader INSERT statement
sed -i "s/is_cross_margin)/is_cross_margin, enable_news_monitoring)/" config/database.go
sed -i "s/trader.IsCrossMargin)/trader.IsCrossMargin, trader.EnableNewsMonitoring)/" config/database.go

# Update GetTraders SELECT statement
sed -i "/COALESCE(is_cross_margin, 1) as is_cross_margin/a\\		       COALESCE(enable_news_monitoring, 0) as enable_news_monitoring," config/database.go

# Update GetTraders Scan
sed -i "s/&trader.IsCrossMargin,$/&trader.IsCrossMargin, \&trader.EnableNewsMonitoring,/" config/database.go

# Update UpdateTrader SET clause
sed -i "s/is_cross_margin = ?, updated_at/is_cross_margin = ?, enable_news_monitoring = ?, updated_at/" config/database.go
sed -i "s/trader.IsCrossMargin, trader.ID/trader.IsCrossMargin, trader.EnableNewsMonitoring, trader.ID/" config/database.go

# Update GetTraderConfig SELECT statement
sed -i "/COALESCE(t.is_cross_margin, 1) as is_cross_margin/a\\			COALESCE(t.enable_news_monitoring, 0) as enable_news_monitoring," config/database.go

echo "✅ config/database.go updated"
echo ""

# 2. Update trader/auto_trader.go
echo "2️⃣ Updating trader/auto_trader.go..."

# Add field to AutoTraderConfig struct
sed -i "/SystemPromptTemplate string.*系统提示词模板名称/a\\\\n\\	// 新闻监控\\n\\	EnableNewsMonitoring bool // 是否启用新闻监控" trader/auto_trader.go

echo "✅ trader/auto_trader.go updated"
echo ""

# 3. Update decision/engine.go
echo "3️⃣ Updating decision/engine.go..."

# Add import
sed -i 's/import (/import (\n\t"nofx\/news"/' decision/engine.go

# Add field to Context struct
sed -i "/BTCETHLeverage.*int.*json:\"-\"/a\\	EnableNewsMonitoring bool                     \`json:\"-\"\` // Whether news monitoring is enabled" decision/engine.go

# Add news context to buildUserPrompt (this is more complex, we'll add it after 'var sb strings.Builder')
sed -i "/func buildUserPrompt(ctx \*Context) string {/,/var sb strings.Builder/ {
    /var sb strings.Builder/a\\\\n\\	// NEW: Add news context if enabled\\n\\	if ctx.EnableNewsMonitoring {\\n\\		newsCtx, err := news.BuildNewsContext()\\n\\		if err == nil && newsCtx != nil {\\n\\			sb.WriteString(news.Format(newsCtx))\\n\\		} else if err != nil {\\n\\			log.Printf(\"⚠️ [News] Failed to fetch news context: %v\", err)\\n\\		}\\n\\	}
}" decision/engine.go

echo "✅ decision/engine.go updated"
echo ""

# 4. Update api/server.go
echo "4️⃣ Updating api/server.go..."

# Add to CreateTraderRequest
sed -i "/type CreateTraderRequest struct/,/}/ {
    /UseOITop.*bool.*json:\"use_oi_top\"/a\\	EnableNewsMonitoring bool  \`json:\"enable_news_monitoring\"\`
}" api/server.go

# Add to UpdateTraderRequest
sed -i "/type UpdateTraderRequest struct/,/}/ {
    /UseOITop.*bool.*json:\"use_oi_top\"/a\\	EnableNewsMonitoring bool \`json:\"enable_news_monitoring\"\`
}" api/server.go

# Update handleCreateTrader
sed -i "/UseOITop:.*req.UseOITop,$/a\\		EnableNewsMonitoring: req.EnableNewsMonitoring," api/server.go

# Update handleUpdateTrader
sed -i "/UseOITop:.*req.UseOITop,$/a\\		EnableNewsMonitoring: req.EnableNewsMonitoring," api/server.go

echo "✅ api/server.go updated"
echo ""

# 5. Update web/src/types.ts
echo "5️⃣ Updating web/src/types.ts..."

# Add to CreateTraderRequest
sed -i "/use_oi_top\\?:.*boolean$/a\\  enable_news_monitoring?: boolean" web/src/types.ts

# Add to TraderConfigData
sed -i "/use_oi_top:.*boolean$/a\\  enable_news_monitoring: boolean" web/src/types.ts

echo "✅ web/src/types.ts updated"
echo ""

# 6. Update web/src/components/TraderConfigModal.tsx
echo "6️⃣ Updating web/src/components/TraderConfigModal.tsx..."

# Add to interface (first occurrence)
sed -i "0,/use_oi_top:.*boolean/{s/use_oi_top:.*boolean$/&\n  enable_news_monitoring: boolean/}" web/src/components/TraderConfigModal.tsx

# Add to useState initialization
sed -i "/use_oi_top:.*false,$/a\\    enable_news_monitoring: false," web/src/components/TraderConfigModal.tsx

# Add UI checkbox (this requires finding the right location - after use_oi_top checkbox)
# This is complex - we'll add a marker comment and instructions

echo "✅ web/src/components/TraderConfigModal.tsx partially updated"
echo "   ⚠️  You need to manually add the UI checkbox after the 'use_oi_top' checkbox"
echo ""

# 7. Update .env.example
echo "7️⃣ Updating .env.example..."

if [ ! -f .env.example ]; then
    touch .env.example
fi

if ! grep -q "CRYPTOPANIC_API_KEY" .env.example; then
    echo "" >> .env.example
    echo "# News Monitoring APIs (Optional - fallback to free sources if not set)" >> .env.example
    echo "CRYPTOPANIC_API_KEY=" >> .env.example
    echo "NITTER_INSTANCE=nitter.net" >> .env.example
fi

echo "✅ .env.example updated"
echo ""

echo "🎉 News Monitoring Integration Complete!"
echo ""
echo "📋 Next Steps:"
echo "1. Manually add UI checkbox in TraderConfigModal.tsx (see NEWS_MONITORING_CHANGES.md)"
echo "2. Review changes with: git diff"
echo "3. Test with: ./start.sh restart --build"
echo ""
echo "💾 Backups saved with .backup extension if you need to rollback"
