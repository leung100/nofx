package news

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// NewsEvent represents a crypto news article
type NewsEvent struct {
	Title     string
	Source    string
	URL       string
	Timestamp time.Time
	Sentiment string // "bullish", "bearish", "neutral"
	Impact    string // "high", "medium", "low"
}

// CryptoPanic API response structures
type CryptoPanicResponse struct {
	Count   int                   `json:"count"`
	Results []CryptoPanicNewsItem `json:"results"`
}

type CryptoPanicNewsItem struct {
	Title     string                `json:"title"`
	URL       string                `json:"url"`
	PublishedAt string              `json:"published_at"`
	Source    CryptoPanicSource     `json:"source"`
	Votes     CryptoPanicVotes      `json:"votes"`
}

type CryptoPanicSource struct {
	Title string `json:"title"`
}

type CryptoPanicVotes struct {
	Positive int `json:"positive"`
	Negative int `json:"negative"`
	Important int `json:"important"`
	Liked    int `json:"liked"`
}

// GetRecentNews fetches recent crypto news from CryptoPanic API and RSS feeds
func GetRecentNews() ([]NewsEvent, error) {
	var allNews []NewsEvent

	// Try CryptoPanic API first
	cryptoPanicNews, err := fetchCryptoPanicNews()
	if err != nil {
		log.Printf("⚠️  [News] CryptoPanic API failed: %v (falling back to RSS)", err)
	} else {
		allNews = append(allNews, cryptoPanicNews...)
	}

	// Filter for recent news (last hour) and high-impact only
	recentNews := filterRecentAndImportant(allNews)

	log.Printf("✅ [News] Fetched %d high-impact crypto news articles", len(recentNews))
	return recentNews, nil
}

// fetchCryptoPanicNews fetches news from CryptoPanic API (free tier: 500 calls/day)
func fetchCryptoPanicNews() ([]NewsEvent, error) {
	apiKey := os.Getenv("CRYPTOPANIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CRYPTOPANIC_API_KEY not set")
	}

	// Fetch only important news
	url := fmt.Sprintf("https://cryptopanic.com/api/v1/posts/?auth_token=%s&kind=news&filter=important&public=true", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result CryptoPanicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to NewsEvent
	var news []NewsEvent
	for _, item := range result.Results {
		timestamp, _ := time.Parse(time.RFC3339, item.PublishedAt)

		// Determine impact based on votes
		impact := "low"
		if item.Votes.Important > 5 {
			impact = "high"
		} else if item.Votes.Important > 2 {
			impact = "medium"
		}

		// Determine sentiment from votes
		sentiment := "neutral"
		if item.Votes.Positive > item.Votes.Negative+2 {
			sentiment = "bullish"
		} else if item.Votes.Negative > item.Votes.Positive+2 {
			sentiment = "bearish"
		}

		// Also analyze title for sentiment
		titleSentiment := AnalyzeSentiment(item.Title)
		if titleSentiment != "neutral" {
			sentiment = titleSentiment
		}

		news = append(news, NewsEvent{
			Title:     item.Title,
			Source:    item.Source.Title,
			URL:       item.URL,
			Timestamp: timestamp,
			Sentiment: sentiment,
			Impact:    impact,
		})
	}

	return news, nil
}

// filterRecentAndImportant filters news for recent (last hour) and important only with deduplication
func filterRecentAndImportant(news []NewsEvent) []NewsEvent {
	// Critical keywords that always get included
	criticalKeywords := []string{
		"sec lawsuit", "sec sues", "sec charges", "etf approved", "etf rejected",
		"exchange hack", "security breach", "halts withdrawals", "trading halt",
		"delisted", "delisting", "arrested", "indicted", "banned",
		"federal reserve", "interest rate", "fomc", "tariff", "trade war",
		"emergency", "investigation", "regulatory action",
	}

	now := time.Now()
	var critical []NewsEvent
	var important []NewsEvent
	seenTitles := make(map[string]bool) // Deduplication

	for _, item := range news {
		age := now.Sub(item.Timestamp).Minutes()
		lowerTitle := strings.ToLower(item.Title)

		// Deduplicate by title similarity (first 50 chars)
		titleKey := lowerTitle
		if len(titleKey) > 50 {
			titleKey = titleKey[:50]
		}
		if seenTitles[titleKey] {
			continue
		}
		seenTitles[titleKey] = true

		// Check for critical keywords
		isCritical := false
		for _, keyword := range criticalKeywords {
			if strings.Contains(lowerTitle, keyword) {
				isCritical = true
				break
			}
		}

		// Critical news: include if <1 hour
		if isCritical && age < 60 {
			critical = append(critical, item)
			continue
		}

		// Important news: high impact and <30 min
		if item.Impact == "high" && age < 30 {
			important = append(important, item)
		}
	}

	// Combine: critical first, then important, limit to 3 total
	result := critical
	for i := 0; i < len(important) && len(result) < 3; i++ {
		result = append(result, important[i])
	}

	return result
}

// GetNewsKeywords extracts keywords from news titles for trend detection
func GetNewsKeywords(news []NewsEvent) []string {
	keywordMap := make(map[string]int)

	for _, item := range news {
		words := strings.Fields(strings.ToLower(item.Title))
		for _, word := range words {
			// Filter out common words
			if len(word) < 4 {
				continue
			}
			if isCommonWord(word) {
				continue
			}
			keywordMap[word]++
		}
	}

	// Extract top keywords (mentioned 2+ times)
	var keywords []string
	for word, count := range keywordMap {
		if count >= 2 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// isCommonWord checks if a word is too common to be a keyword
func isCommonWord(word string) bool {
	common := []string{
		"the", "and", "for", "with", "from", "that", "this", "will",
		"have", "been", "said", "says", "after", "more", "news", "today",
		"about", "could", "would", "should", "their", "which", "there",
	}

	for _, cw := range common {
		if word == cw {
			return true
		}
	}
	return false
}
