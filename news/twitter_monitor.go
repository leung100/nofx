package news

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// TweetAlert represents a tweet from a monitored account
type TweetAlert struct {
	Author    string
	Content   string
	Timestamp time.Time
	Sentiment string // "bullish", "bearish", "neutral"
}

// RSS feed structures for Nitter
type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Items []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// Monitored Twitter accounts
var monitoredAccounts = []string{
	"realDonaldTrump", // Trump - Presidential crypto statements
	"elonmusk",        // Elon Musk - Tech mogul crypto influence
	"SECGov",          // SEC - Regulatory announcements
	"GaryGensler",     // SEC Chair statements
	"cz_binance",      // Binance CEO - Exchange operations
	"brian_armstrong", // Coinbase CEO - Regulatory, listings
}

// GetRecentTweets fetches recent tweets from monitored accounts via Nitter RSS
func GetRecentTweets() ([]TweetAlert, error) {
	nitterInstance := os.Getenv("NITTER_INSTANCE")
	if nitterInstance == "" {
		nitterInstance = "nitter.net" // Default free instance
	}

	var allTweets []TweetAlert
	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch tweets from each account in parallel
	for _, account := range monitoredAccounts {
		tweets, err := fetchTweetsForAccount(client, nitterInstance, account)
		if err != nil {
			log.Printf("⚠️  [News] Failed to fetch tweets from @%s: %v", account, err)
			continue // Skip failed accounts, don't break entire flow
		}
		allTweets = append(allTweets, tweets...)
	}

	// Filter for crypto-relevant tweets (last hour only)
	relevantTweets := filterCryptoTweets(allTweets)

	log.Printf("✅ [News] Fetched %d crypto-relevant tweets from %d accounts", len(relevantTweets), len(monitoredAccounts))
	return relevantTweets, nil
}

// fetchTweetsForAccount fetches RSS feed for a single account
func fetchTweetsForAccount(client *http.Client, nitterInstance, account string) ([]TweetAlert, error) {
	url := fmt.Sprintf("https://%s/%s/rss", nitterInstance, account)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add user agent to avoid blocks
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse RSS XML
	var rss RSS
	if err := xml.Unmarshal(body, &rss); err != nil {
		return nil, err
	}

	// Convert RSS items to TweetAlerts
	var tweets []TweetAlert
	for _, item := range rss.Channel.Items {
		// Parse timestamp
		timestamp, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			timestamp, _ = time.Parse(time.RFC1123, item.PubDate)
		}

		// Only include tweets from last hour
		if time.Since(timestamp) > 1*time.Hour {
			continue
		}

		tweet := TweetAlert{
			Author:    account,
			Content:   cleanTweetContent(item.Description),
			Timestamp: timestamp,
			Sentiment: AnalyzeSentiment(item.Description), // From sentiment.go
		}
		tweets = append(tweets, tweet)
	}

	return tweets, nil
}

// filterCryptoTweets filters tweets for crypto-relevant keywords with tiered priority
func filterCryptoTweets(tweets []TweetAlert) []TweetAlert {
	// CRITICAL keywords (always include if <60 min)
	criticalKeywords := []string{
		"sec lawsuit", "etf approved", "etf rejected", "etf denied",
		"exchange hack", "hack", "security breach", "stolen funds",
		"halts withdrawals", "halt trading", "suspended", "delisted", "delisting",
		"arrested", "indicted", "charged with", "banned", "ban on",
		"federal reserve", "fed rate", "interest rate", "fomc", "powell",
		"cpi", "inflation data", "employment data", "jobs report",
		"tariff", "trade war", "china tariff", "import tax",
		"regulatory action", "emergency order", "cease and desist",
	}

	// IMPORTANT keywords for crypto/macro (include if <30 min)
	importantKeywords := []string{
		"bitcoin", "btc", "ethereum", "eth", "crypto", "cryptocurrency",
		"binance", "coinbase", "sec", "etf", "blockchain",
		"dollar", "dxy", "usd", "treasury", "bond yield",
		"stock market", "s&p 500", "nasdaq", "dow jones", "vix",
		"regulation", "compliance", "license", "approval",
	}

	now := time.Now()
	var critical []TweetAlert
	var important []TweetAlert

	for _, tweet := range tweets {
		age := now.Sub(tweet.Timestamp).Minutes()
		lowerContent := strings.ToLower(tweet.Content)

		// Check for critical keywords
		isCritical := false
		for _, keyword := range criticalKeywords {
			if strings.Contains(lowerContent, keyword) {
				isCritical = true
				break
			}
		}

		if isCritical && age < 60 {
			critical = append(critical, tweet)
			continue
		}

		// Check for important keywords (Trump/Elon get priority even without keywords)
		if age < 30 {
			isImportant := false

			// Trump and Elon always important if recent
			if tweet.Author == "realDonaldTrump" || tweet.Author == "elonmusk" {
				isImportant = true
			}

			// Or if contains important keywords
			if !isImportant {
				for _, keyword := range importantKeywords {
					if strings.Contains(lowerContent, keyword) {
						isImportant = true
						break
					}
				}
			}

			if isImportant {
				important = append(important, tweet)
			}
		}
	}

	// Combine: critical first, then important, limit total to 3
	result := critical
	for i := 0; i < len(important) && len(result) < 3; i++ {
		result = append(result, important[i])
	}

	return result
}

// cleanTweetContent removes HTML tags and extra whitespace
func cleanTweetContent(html string) string {
	// Remove HTML tags
	cleaned := strings.ReplaceAll(html, "<br>", " ")
	cleaned = strings.ReplaceAll(cleaned, "<br/>", " ")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")

	// Remove any remaining HTML tags
	for strings.Contains(cleaned, "<") && strings.Contains(cleaned, ">") {
		start := strings.Index(cleaned, "<")
		end := strings.Index(cleaned, ">")
		if end > start {
			cleaned = cleaned[:start] + cleaned[end+1:]
		} else {
			break
		}
	}

	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)

	// Truncate if too long (keep first 280 chars)
	if len(cleaned) > 280 {
		cleaned = cleaned[:277] + "..."
	}

	return cleaned
}
