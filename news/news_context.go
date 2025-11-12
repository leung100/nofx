package news

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// NewsContext aggregates all news sources for AI decision-making
type NewsContext struct {
	TwitterAlerts   []TweetAlert
	NewsEvents      []NewsEvent
	WhaleAlerts     []WhaleAlert
	SentimentScore  float64 // -1.0 (very bearish) to +1.0 (very bullish)
	SentimentLabel  string  // "very bullish", "bullish", "neutral", "bearish", "very bearish"
	RiskLevel       string  // "low", "medium", "high", "extreme"
	FetchedAt       time.Time
	FetchDurationMs int64
}

// BuildNewsContext fetches and aggregates news from all sources (parallel)
func BuildNewsContext() (*NewsContext, error) {
	startTime := time.Now()
	ctx := &NewsContext{
		FetchedAt: startTime,
	}

	// Fetch from all sources in parallel for speed
	var wg sync.WaitGroup
	var twitterErr, newsErr, whaleErr error

	wg.Add(3)

	// Fetch Twitter alerts
	go func() {
		defer wg.Done()
		tweets, err := GetRecentTweets()
		if err != nil {
			twitterErr = err
			log.Printf("⚠️  [News] Twitter fetch failed: %v", err)
		} else {
			ctx.TwitterAlerts = tweets
		}
	}()

	// Fetch crypto news
	go func() {
		defer wg.Done()
		news, err := GetRecentNews()
		if err != nil {
			newsErr = err
			log.Printf("⚠️  [News] Crypto news fetch failed: %v", err)
		} else {
			ctx.NewsEvents = news
		}
	}()

	// Fetch whale alerts
	go func() {
		defer wg.Done()
		whales, err := GetRecentWhaleAlerts()
		if err != nil {
			whaleErr = err
			log.Printf("⚠️  [News] Whale alerts fetch failed: %v", err)
		} else {
			// Filter for relevant coins only
			ctx.WhaleAlerts = FilterRelevantWhaleAlerts(whales)
		}
	}()

	wg.Wait()

	// Optional: Apply AI-powered filtering (if enabled)
	aiConfig := GetDefaultAIFilterConfig()
	if aiConfig.Enabled {
		log.Printf("🤖 [News] AI filter enabled - scoring %d items...",
			len(ctx.TwitterAlerts)+len(ctx.NewsEvents)+len(ctx.WhaleAlerts))

		filterReq := FilterRequest{
			Tweets: ctx.TwitterAlerts,
			News:   ctx.NewsEvents,
			Whales: ctx.WhaleAlerts,
		}

		filtered, err := FilterNewsWithAI(filterReq, aiConfig)
		if err != nil {
			log.Printf("⚠️  [News] AI filter failed, using keyword filtering: %v", err)
			// Fallback to existing keyword filtering (already applied above)
		} else {
			// Replace with AI-filtered results
			ctx.TwitterAlerts = filtered.Tweets
			ctx.NewsEvents = filtered.News
			ctx.WhaleAlerts = filtered.Whales
			log.Printf("✅ [News] AI filtering complete")
		}
	} else {
		log.Printf("📋 [News] AI filter disabled, using keyword filtering")
	}

	// Calculate aggregate sentiment
	ctx.SentimentScore, ctx.SentimentLabel = AggregateSentiment(
		ctx.TwitterAlerts,
		ctx.NewsEvents,
		ctx.WhaleAlerts,
	)

	// Assess risk level
	ctx.RiskLevel = assessRiskLevel(ctx)

	ctx.FetchDurationMs = time.Since(startTime).Milliseconds()

	log.Printf("✅ [News] Context built in %dms | Tweets: %d | News: %d | Whales: %d | Sentiment: %s",
		ctx.FetchDurationMs,
		len(ctx.TwitterAlerts),
		len(ctx.NewsEvents),
		len(ctx.WhaleAlerts),
		ctx.SentimentLabel,
	)

	// Return context even if some sources failed (graceful degradation)
	if twitterErr != nil && newsErr != nil && whaleErr != nil {
		return ctx, fmt.Errorf("all news sources failed")
	}

	return ctx, nil
}

// assessRiskLevel determines market risk level from news context
func assessRiskLevel(ctx *NewsContext) string {
	highImpactCount := 0
	for _, news := range ctx.NewsEvents {
		if news.Impact == "high" {
			highImpactCount++
		}
	}

	// Risk factors
	hasRegulatory := false
	hasExchangeIssue := false
	hasMajorWhaleMovement := len(ctx.WhaleAlerts) > 5

	for _, tweet := range ctx.TwitterAlerts {
		lower := strings.ToLower(tweet.Content)
		if strings.Contains(lower, "sec") || strings.Contains(lower, "regulation") ||
			strings.Contains(lower, "lawsuit") || strings.Contains(lower, "ban") {
			hasRegulatory = true
		}
		if strings.Contains(lower, "binance") || strings.Contains(lower, "coinbase") ||
			strings.Contains(lower, "exchange") && (strings.Contains(lower, "halt") || strings.Contains(lower, "withdraw")) {
			hasExchangeIssue = true
		}
	}

	// Determine risk level
	switch {
	case hasRegulatory && hasExchangeIssue:
		return "extreme"
	case highImpactCount >= 3 || hasExchangeIssue:
		return "high"
	case highImpactCount >= 1 || hasMajorWhaleMovement:
		return "medium"
	default:
		return "low"
	}
}

// Format formats NewsContext into AI-readable prompt section
func Format(ctx *NewsContext) string {
	if ctx == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("## 🌐 市场动态 / Market Context (Last 1 Hour)\n\n")

	// Overall sentiment
	sentimentEmoji := GetSentimentEmoji(ctx.SentimentLabel)
	sb.WriteString(fmt.Sprintf("**整体情绪 / Overall Sentiment**: %s %s (Score: %.2f)\n", sentimentEmoji, ctx.SentimentLabel, ctx.SentimentScore))
	sb.WriteString(fmt.Sprintf("**风险等级 / Risk Level**: %s\n\n", strings.ToUpper(ctx.RiskLevel)))

	// Twitter alerts
	if len(ctx.TwitterAlerts) > 0 {
		sb.WriteString("### 📱 Twitter 动态 / Twitter Updates\n\n")
		for i, tweet := range ctx.TwitterAlerts {
			if i >= 3 { // Limit to 3 most critical
				break
			}
			emoji := GetSentimentEmoji(tweet.Sentiment)
			minutesAgo := int(time.Since(tweet.Timestamp).Minutes())
			sb.WriteString(fmt.Sprintf("- %s **@%s** (%d min ago): \"%s\"\n",
				emoji, tweet.Author, minutesAgo, truncateText(tweet.Content, 150)))
			sb.WriteString(fmt.Sprintf("  → Sentiment: %s\n\n", strings.ToUpper(tweet.Sentiment)))
		}
	}

	// News events
	if len(ctx.NewsEvents) > 0 {
		sb.WriteString("### 📰 重要新闻 / Important News\n\n")
		for i, news := range ctx.NewsEvents {
			if i >= 3 { // Limit to 3 most critical
				break
			}
			emoji := GetSentimentEmoji(news.Sentiment)
			minutesAgo := int(time.Since(news.Timestamp).Minutes())
			impactLabel := strings.ToUpper(news.Impact)
			sb.WriteString(fmt.Sprintf("- %s [%s IMPACT] **%s** (%d min ago)\n",
				emoji, impactLabel, news.Title, minutesAgo))
			sb.WriteString(fmt.Sprintf("  → Source: %s | Sentiment: %s\n\n",
				news.Source, strings.ToUpper(news.Sentiment)))
		}
	}

	// Whale alerts
	if len(ctx.WhaleAlerts) > 0 {
		sb.WriteString("### 🐋 巨鲸动态 / Whale Movements (>$50M)\n\n")
		for i, whale := range ctx.WhaleAlerts {
			if i >= 3 { // Limit to 3 most critical
				break
			}
			minutesAgo := int(time.Since(whale.Timestamp).Minutes())

			// Format movement
			movement := fmt.Sprintf("%.2f %s ($%.2fM)", whale.Amount, whale.Symbol, whale.AmountUSD/1_000_000)
			fromLabel := whale.FromOwner
			if fromLabel == "" {
				fromLabel = "Unknown Wallet"
			}
			toLabel := whale.ToOwner
			if toLabel == "" {
				toLabel = "Unknown Wallet"
			}

			sb.WriteString(fmt.Sprintf("- %s moved from **%s** to **%s** (%d min ago)\n",
				movement, fromLabel, toLabel, minutesAgo))
			sb.WriteString(fmt.Sprintf("  → Interpretation: %s\n\n",
				strings.ToUpper(whale.Interpretation)))
		}
	}

	// If no data available
	if len(ctx.TwitterAlerts) == 0 && len(ctx.NewsEvents) == 0 && len(ctx.WhaleAlerts) == 0 {
		sb.WriteString("*No significant market events in the last hour.*\n\n")
	}

	sb.WriteString("---\n\n")

	return sb.String()
}

// truncateText truncates text to maxLen with ellipsis
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
