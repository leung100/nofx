#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
News Monitoring Integration - Automated Patch Script
"""

import os
import re
import shutil

def backup_file(filepath):
    backup_path = f"{filepath}.backup"
    shutil.copy2(filepath, backup_path)
    print(f"[OK] Backed up: {filepath}")

def apply_changes():
    print(">>> Applying News Monitoring Integration...")
    print("")

    # Change 1: config/database.go
    print("[1/7] Updating config/database.go...")
    db_file = "config/database.go"
    backup_file(db_file)

    with open(db_file, 'r', encoding='utf-8') as f:
        content = f.read()

    # Add ALTER TABLE
    content = content.replace(
        "`ALTER TABLE traders ADD COLUMN system_prompt_template TEXT DEFAULT 'default'`, // 系统提示词模板名称",
        "`ALTER TABLE traders ADD COLUMN system_prompt_template TEXT DEFAULT 'default'`, // 系统提示词模板名称\n\t\t`ALTER TABLE traders ADD COLUMN enable_news_monitoring BOOLEAN DEFAULT 0`,      // 是否启用新闻监控"
    )

    # Add field to TraderRecord struct
    content = content.replace(
        'IsCrossMargin        bool      `json:"is_cross_margin"`        // 是否为全仓模式（true=全仓，false=逐仓）',
        'IsCrossMargin        bool      `json:"is_cross_margin"`        // 是否为全仓模式（true=全仓，false=逐仓）\n\tEnableNewsMonitoring bool      `json:"enable_news_monitoring"` // 是否启用新闻监控'
    )

    # Update CreateTrader INSERT
    content = content.replace(
        'INSERT INTO traders (id, user_id, name, ai_model_id, exchange_id, initial_balance, scan_interval_minutes, is_running, btc_eth_leverage, altcoin_leverage, trading_symbols, use_coin_pool, use_oi_top, custom_prompt, override_base_prompt, system_prompt_template, is_cross_margin)',
        'INSERT INTO traders (id, user_id, name, ai_model_id, exchange_id, initial_balance, scan_interval_minutes, is_running, btc_eth_leverage, altcoin_leverage, trading_symbols, use_coin_pool, use_oi_top, custom_prompt, override_base_prompt, system_prompt_template, is_cross_margin, enable_news_monitoring)'
    )

    content = content.replace(
        'VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)',
        'VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)'
    )

    # Update CreateTrader values
    content = re.sub(
        r'(\s+`, trader\.ID, trader\.UserID, trader\.Name, trader\.AIModelID, trader\.ExchangeID, trader\.InitialBalance, trader\.ScanIntervalMinutes, trader\.IsRunning, trader\.BTCETHLeverage, trader\.AltcoinLeverage, trader\.TradingSymbols, trader\.UseCoinPool, trader\.UseOITop, trader\.CustomPrompt, trader\.OverrideBasePrompt, trader\.SystemPromptTemplate, trader\.IsCrossMargin)\n',
        r'\1, trader.EnableNewsMonitoring)\n',
        content
    )

    # Update GetTraders SELECT
    content = content.replace(
        'COALESCE(is_cross_margin, 1) as is_cross_margin, created_at, updated_at',
        'COALESCE(is_cross_margin, 1) as is_cross_margin,\n\t\t       COALESCE(enable_news_monitoring, 0) as enable_news_monitoring,\n\t\t       created_at, updated_at'
    )

    # Update GetTraders Scan
    content = content.replace(
        '&trader.IsCrossMargin,\n\t\t\t&trader.CreatedAt, &trader.UpdatedAt,',
        '&trader.IsCrossMargin, &trader.EnableNewsMonitoring,\n\t\t\t&trader.CreatedAt, &trader.UpdatedAt,'
    )

    # Update UpdateTrader SET
    content = content.replace(
        'system_prompt_template = ?, is_cross_margin = ?, updated_at = CURRENT_TIMESTAMP',
        'system_prompt_template = ?, is_cross_margin = ?, enable_news_monitoring = ?, updated_at = CURRENT_TIMESTAMP'
    )

    # Update UpdateTrader values
    content = re.sub(
        r'(trader\.SystemPromptTemplate, trader\.IsCrossMargin, trader\.ID, trader\.UserID\))',
        r'trader.SystemPromptTemplate, trader.IsCrossMargin, trader.EnableNewsMonitoring, trader.ID, trader.UserID)',
        content
    )

    # Update GetTraderConfig SELECT
    content = content.replace(
        'COALESCE(t.is_cross_margin, 1) as is_cross_margin,\n\t\t\tt.created_at, t.updated_at,',
        'COALESCE(t.is_cross_margin, 1) as is_cross_margin,\n\t\t\tCOALESCE(t.enable_news_monitoring, 0) as enable_news_monitoring,\n\t\t\tt.created_at, t.updated_at,'
    )

    with open(db_file, 'w', encoding='utf-8') as f:
        f.write(content)

    print("[OK] config/database.go updated")
    print("")

    # Change 2: trader/auto_trader.go
    print("[2/7] Updating trader/auto_trader.go...")
    trader_file = "trader/auto_trader.go"
    backup_file(trader_file)

    with open(trader_file, 'r', encoding='utf-8') as f:
        content = f.read()

    content = content.replace(
        '// 系统提示词模板\n\tSystemPromptTemplate string // 系统提示词模板名称（如 "default", "aggressive"）\n}',
        '// 系统提示词模板\n\tSystemPromptTemplate string // 系统提示词模板名称（如 "default", "aggressive"）\n\n\t// 新闻监控\n\tEnableNewsMonitoring bool // 是否启用新闻监控\n}'
    )

    with open(trader_file, 'w', encoding='utf-8') as f:
        f.write(content)

    print("[OK] trader/auto_trader.go updated")
    print("")

    # Change 3: decision/engine.go
    print("[3/7] Updating decision/engine.go...")
    decision_file = "decision/engine.go"
    backup_file(decision_file)

    with open(decision_file, 'r', encoding='utf-8') as f:
        content = f.read()

    if '"nofx/news"' not in content:
        content = content.replace(
            'import (\n',
            'import (\n\t"nofx/news"\n'
        )

    if 'EnableNewsMonitoring bool' not in content:
        content = content.replace(
            'AltcoinLeverage int                     `json:"-"` // 山寨币杠杆倍数（从配置读取）\n}',
            'AltcoinLeverage int                     `json:"-"` // 山寨币杠杆倍数（从配置读取）\n\tEnableNewsMonitoring bool               `json:"-"` // Whether news monitoring is enabled\n}'
        )

    build_prompt_pattern = r'func buildUserPrompt\(ctx \*Context\) string \{\s+var sb strings\.Builder'
    news_context_code = '''func buildUserPrompt(ctx *Context) string {
\tvar sb strings.Builder

\t// NEW: Add news context if enabled
\tif ctx.EnableNewsMonitoring {
\t\tnewsCtx, err := news.BuildNewsContext()
\t\tif err == nil && newsCtx != nil {
\t\t\tsb.WriteString(news.Format(newsCtx))
\t\t} else if err != nil {
\t\t\tlog.Printf("⚠️ [News] Failed to fetch news context: %v", err)
\t\t}
\t}'''

    content = re.sub(build_prompt_pattern, news_context_code, content)

    with open(decision_file, 'w', encoding='utf-8') as f:
        f.write(content)

    print("[OK] decision/engine.go updated")
    print("")

    # Change 4: api/server.go
    print("[4/7] Updating api/server.go...")
    api_file = "api/server.go"
    backup_file(api_file)

    with open(api_file, 'r', encoding='utf-8') as f:
        content = f.read()

    content = re.sub(
        r'(type CreateTraderRequest struct \{[^}]+UseOITop\s+bool\s+`json:"use_oi_top"`)',
        r'\1\n\tEnableNewsMonitoring bool `json:"enable_news_monitoring"`',
        content,
        count=1
    )

    content = re.sub(
        r'(type UpdateTraderRequest struct \{[^}]+UseOITop\s+bool\s+`json:"use_oi_top"`)',
        r'\1\n\tEnableNewsMonitoring bool `json:"enable_news_monitoring"`',
        content,
        count=1
    )

    content = re.sub(
        r'(UseOITop:\s+req\.UseOITop,)\n(\s+\})',
        r'\1\n\t\tEnableNewsMonitoring: req.EnableNewsMonitoring,\n\2',
        content
    )

    if content.count('UseOITop:             req.UseOITop,') > 1:
        parts = content.split('UseOITop:             req.UseOITop,')
        if len(parts) >= 3:
            content = parts[0] + 'UseOITop:             req.UseOITop,' + parts[1] + 'UseOITop:             req.UseOITop,\n\t\tEnableNewsMonitoring: req.EnableNewsMonitoring,' + ''.join(parts[2:])

    with open(api_file, 'w', encoding='utf-8') as f:
        f.write(content)

    print("[OK] api/server.go updated")
    print("")

    # Change 5: web/src/types.ts
    print("[5/7] Updating web/src/types.ts...")
    types_file = "web/src/types.ts"
    backup_file(types_file)

    with open(types_file, 'r', encoding='utf-8') as f:
        content = f.read()

    content = re.sub(
        r'(export interface CreateTraderRequest \{[^}]+use_oi_top\?: boolean)',
        r'\1\n  enable_news_monitoring?: boolean',
        content,
        count=1
    )

    content = re.sub(
        r'(export interface TraderConfigData \{[^}]+use_oi_top: boolean)',
        r'\1\n  enable_news_monitoring: boolean',
        content,
        count=1
    )

    with open(types_file, 'w', encoding='utf-8') as f:
        f.write(content)

    print("[OK] web/src/types.ts updated")
    print("")

    # Change 6: web/src/components/TraderConfigModal.tsx
    print("[6/7] Updating web/src/components/TraderConfigModal.tsx...")
    modal_file = "web/src/components/TraderConfigModal.tsx"
    backup_file(modal_file)

    with open(modal_file, 'r', encoding='utf-8') as f:
        content = f.read()

    content = re.sub(
        r'(interface TraderConfigData \{[^}]+use_oi_top: boolean)',
        r'\1\n  enable_news_monitoring: boolean',
        content,
        count=1
    )

    content = re.sub(
        r'(use_oi_top: false,)',
        r'\1\n    enable_news_monitoring: false,',
        content,
        count=1
    )

    content = re.sub(
        r'(use_oi_top: formData\.use_oi_top,)',
        r'\1\n      enable_news_monitoring: formData.enable_news_monitoring,',
        content,
        count=1
    )

    with open(modal_file, 'w', encoding='utf-8') as f:
        f.write(content)

    print("[OK] web/src/components/TraderConfigModal.tsx updated")
    print("[WARN] You still need to manually add the UI checkbox")
    print("")

    # Change 7: .env.example
    print("[7/7] Updating .env.example...")
    env_file = ".env.example"

    if not os.path.exists(env_file):
        with open(env_file, 'w') as f:
            f.write("")

    with open(env_file, 'r') as f:
        env_content = f.read()

    if 'CRYPTOPANIC_API_KEY' not in env_content:
        with open(env_file, 'a') as f:
            f.write('\n# News Monitoring APIs (Optional)\n')
            f.write('CRYPTOPANIC_API_KEY=\n')
            f.write('NITTER_INSTANCE=nitter.net\n')

    print("[OK] .env.example updated")
    print("")

    print(">>> Automated Changes Complete!")
    print("")
    print("[INFO] Manual Step Required:")
    print("Add UI checkbox in TraderConfigModal.tsx after use_oi_top checkbox")
    print("")
    print("[BACKUP] All files backed up with .backup extension")
    print("[CHECK] Review changes: git diff")
    print("[RUN] Test: ./start.sh restart --build")

if __name__ == "__main__":
    try:
        apply_changes()
    except Exception as e:
        print(f"[ERROR] {e}")
        import traceback
        traceback.print_exc()
