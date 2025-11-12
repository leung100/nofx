package news

import (
	"strings"
)

// AnalyzeSentiment performs simple keyword-based sentiment analysis
// Returns: "bullish", "bearish", or "neutral"
func AnalyzeSentiment(text string) string {
	text = strings.ToLower(text)

	// Bullish keywords
	bullishKeywords := []string{
		"approve", "approved", "approval", "bullish", "pump", "surge", "rally",
		"break", "breakout", "moon", "ath", "all-time high", "adoption",
		"buy", "buying", "accumulate", "institutional", "etf approved",
		"partnership", "integration", "upgrade", "positive", "optimistic",
		"growth", "expand", "innovation", "support", "strong", "gain",
		"rise", "rising", "increase", "up", "higher", "recovery",
	}

	// Bearish keywords
	bearishKeywords := []string{
		"crash", "dump", "bear", "bearish", "decline", "fall", "drop",
		"reject", "rejected", "rejection", "ban", "lawsuit", "fraud",
		"hack", "hacked", "scam", "ponzi", "investigation", "sec charges",
		"enforcement", "warning", "risk", "fear", "panic", "sell",
		"selling", "liquidation", "collapse", "failure", "concern",
		"down", "lower", "decrease", "weak", "negative",
	}

	bullishScore := 0
	bearishScore := 0

	// Count bullish keywords
	for _, keyword := range bullishKeywords {
		if strings.Contains(text, keyword) {
			bullishScore++
		}
	}

	// Count bearish keywords
	for _, keyword := range bearishKeywords {
		if strings.Contains(text, keyword) {
			bearishScore++
		}
	}

	// Determine sentiment
	if bullishScore > bearishScore {
		return "bullish"
	} else if bearishScore > bullishScore {
		return "bearish"
	}

	return "neutral"
}

// GetSentimentEmoji returns an emoji representation of sentiment
func GetSentimentEmoji(sentiment string) string {
	switch sentiment {
	case "bullish":
		return "🟢"
	case "bearish":
		return "🔴"
	default:
		return "⚪"
	}
}

// AggregateSentiment calculates overall market sentiment from multiple sources
func AggregateSentiment(tweets []TweetAlert, news []NewsEvent, whales []WhaleAlert) (float64, string) {
	totalScore := 0.0
	count := 0

	// Twitter sentiment (weight: 1.0)
	for _, tweet := range tweets {
		switch tweet.Sentiment {
		case "bullish":
			totalScore += 1.0
		case "bearish":
			totalScore -= 1.0
		}
		count++
	}

	// News sentiment (weight: 1.5 for high impact, 1.0 for medium)
	for _, item := range news {
		weight := 1.0
		if item.Impact == "high" {
			weight = 1.5
		}

		switch item.Sentiment {
		case "bullish":
			totalScore += weight
		case "bearish":
			totalScore -= weight
		}
		count++
	}

	// Whale movements (weight: 1.0)
	for _, whale := range whales {
		switch whale.Interpretation {
		case "accumulation":
			totalScore += 1.0
		case "distribution":
			totalScore -= 1.0
		}
		count++
	}

	if count == 0 {
		return 0.0, "neutral"
	}

	// Calculate average sentiment score (-1.0 to +1.0)
	avgScore := totalScore / float64(count)

	// Determine overall sentiment label
	var label string
	switch {
	case avgScore > 0.3:
		label = "very bullish"
	case avgScore > 0.1:
		label = "bullish"
	case avgScore < -0.3:
		label = "very bearish"
	case avgScore < -0.1:
		label = "bearish"
	default:
		label = "neutral"
	}

	return avgScore, label
}
