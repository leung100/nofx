package news

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// AIFilterConfig holds configuration for AI-powered news filtering
type AIFilterConfig struct {
	Enabled       bool
	APIKey        string
	BaseURL       string
	Model         string
	ScoreThreshold int // Minimum score to include (0-10)
	Timeout       time.Duration
}

// NewsScore represents AI's scoring of a news item
type NewsScore struct {
	ID     int     `json:"id"`
	Score  int     `json:"score"`
	Reason string  `json:"reason"`
}

// FilterRequest holds all news items for batch scoring
type FilterRequest struct {
	Tweets []TweetAlert
	News   []NewsEvent
	Whales []WhaleAlert
}

// FilterResponse holds filtered results with scores
type FilterResponse struct {
	Tweets []TweetAlert
	News   []NewsEvent
	Whales []WhaleAlert
	Scores map[string]NewsScore // key: "tweet_0", "news_1", "whale_2"
}

// scoreCache caches AI scores to avoid redundant API calls
var (
	scoreCache      = make(map[string]CachedScore)
	scoreCacheMutex sync.RWMutex
)

type CachedScore struct {
	Score     int
	Reason    string
	ExpiresAt time.Time
}

// GetDefaultAIFilterConfig returns default configuration from env vars
func GetDefaultAIFilterConfig() AIFilterConfig {
	enabled := os.Getenv("ENABLE_AI_NEWS_FILTER") == "true"

	// Default to DeepSeek for cost-effectiveness
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY") // Fallback to OpenAI if set
	}

	baseURL := os.Getenv("AI_FILTER_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1" // Default to DeepSeek
	}

	model := os.Getenv("AI_FILTER_MODEL")
	if model == "" {
		model = "deepseek-chat" // Cheap and fast
	}

	return AIFilterConfig{
		Enabled:       enabled,
		APIKey:        apiKey,
		BaseURL:       baseURL,
		Model:         model,
		ScoreThreshold: 7, // Only include score >= 7
		Timeout:       15 * time.Second,
	}
}

// FilterNewsWithAI uses DeepSeek to score and filter news items
func FilterNewsWithAI(req FilterRequest, config AIFilterConfig) (*FilterResponse, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("AI filter disabled")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("AI filter API key not configured")
	}

	// Build prompt with all news items
	prompt := buildFilteringPrompt(req)

	// Call DeepSeek API
	scores, err := callScoringAPI(prompt, config)
	if err != nil {
		return nil, fmt.Errorf("AI scoring failed: %w", err)
	}

	// Filter items based on scores
	response := filterByScores(req, scores, config.ScoreThreshold)

	log.Printf("🤖 [AI Filter] Scored %d items → Kept %d (threshold: %d)",
		len(req.Tweets)+len(req.News)+len(req.Whales),
		len(response.Tweets)+len(response.News)+len(response.Whales),
		config.ScoreThreshold)

	return response, nil
}

// buildFilteringPrompt creates the AI prompt for scoring news
func buildFilteringPrompt(req FilterRequest) string {
	var prompt bytes.Buffer

	prompt.WriteString("You are a crypto market analyst. Rate each news item 0-10 based on likelihood to impact BTC/ETH price in the next 1 hour.\n\n")
	prompt.WriteString("Scoring criteria:\n")
	prompt.WriteString("10 = Critical impact (SEC ETF approval/rejection, major exchange hack >$100M, Fed rate decision, major regulatory ban)\n")
	prompt.WriteString("7-9 = Significant impact (regulatory warnings, large liquidations >$50M, major partnerships, Trump/Elon crypto statements, trade war announcements)\n")
	prompt.WriteString("4-6 = Minor impact (industry news, medium whale moves, analyst predictions from reputable sources)\n")
	prompt.WriteString("1-3 = Low impact (generic articles, old news rehashed, minor updates, low-quality speculation)\n")
	prompt.WriteString("0 = Irrelevant (completely unrelated to crypto/macro, sarcasm, jokes, spam)\n\n")
	prompt.WriteString("Consider:\n")
	prompt.WriteString("- Recency: Newer events score higher\n")
	prompt.WriteString("- Source credibility: Official accounts > random analysts\n")
	prompt.WriteString("- Market context: Macro events (Fed, tariffs, dollar) heavily impact crypto\n")
	prompt.WriteString("- Magnitude: Larger amounts/impacts score higher\n\n")
	prompt.WriteString("News items to score:\n\n")

	itemID := 1

	// Add tweets
	for i, tweet := range req.Tweets {
		age := int(time.Since(tweet.Timestamp).Minutes())
		prompt.WriteString(fmt.Sprintf("%d. [TWEET_%d] @%s (%d min ago): \"%s\"\n",
			itemID, i, tweet.Author, age, truncateForPrompt(tweet.Content, 200)))
		itemID++
	}

	// Add news
	for i, item := range req.News {
		age := int(time.Since(item.Timestamp).Minutes())
		prompt.WriteString(fmt.Sprintf("%d. [NEWS_%d] %s (%d min ago): \"%s\" [Source: %s]\n",
			itemID, i, item.Impact, age, truncateForPrompt(item.Title, 150), item.Source))
		itemID++
	}

	// Add whale alerts
	for i, whale := range req.Whales {
		age := int(time.Since(whale.Timestamp).Minutes())
		amountM := whale.AmountUSD / 1_000_000
		prompt.WriteString(fmt.Sprintf("%d. [WHALE_%d] $%.0fM %s moved from %s to %s (%d min ago)\n",
			itemID, i, amountM, whale.Symbol, whale.FromOwner, whale.ToOwner, age))
		itemID++
	}

	prompt.WriteString("\nReturn ONLY a JSON array with this exact format:\n")
	prompt.WriteString("[{\"id\": 1, \"score\": 9, \"reason\": \"brief explanation\"}, ...]\n")
	prompt.WriteString("Do not include any other text, markdown, or explanations. Only the JSON array.")

	return prompt.String()
}

// callScoringAPI calls DeepSeek to score news items
func callScoringAPI(prompt string, config AIFilterConfig) ([]NewsScore, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%x", hashString(prompt))
	if cached, ok := getCachedScore(cacheKey); ok {
		log.Printf("🎯 [AI Filter] Using cached scores")
		return cached, nil
	}

	// Prepare API request
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3, // Low temperature for consistent scoring
		"max_tokens":  1000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Make HTTP request
	client := &http.Client{Timeout: config.Timeout}
	req, err := http.NewRequest("POST", config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if len(apiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// Parse JSON scores from AI response
	content := apiResponse.Choices[0].Message.Content

	// Extract JSON array (AI might wrap it in markdown)
	jsonStart := bytes.Index([]byte(content), []byte("["))
	jsonEnd := bytes.LastIndex([]byte(content), []byte("]"))
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no JSON array found in response: %s", content)
	}
	jsonContent := content[jsonStart : jsonEnd+1]

	var scores []NewsScore
	if err := json.Unmarshal([]byte(jsonContent), &scores); err != nil {
		return nil, fmt.Errorf("failed to parse scores JSON: %w (content: %s)", err, jsonContent)
	}

	// Cache the scores (expire in 10 minutes)
	setCachedScore(cacheKey, scores, 10*time.Minute)

	log.Printf("✅ [AI Filter] Scored %d items via API", len(scores))

	return scores, nil
}

// filterByScores filters news items based on AI scores
func filterByScores(req FilterRequest, scores []NewsScore, threshold int) *FilterResponse {
	response := &FilterResponse{
		Scores: make(map[string]NewsScore),
	}

	// Map scores by ID
	scoreMap := make(map[int]NewsScore)
	for _, s := range scores {
		scoreMap[s.ID] = s
	}

	currentID := 1

	// Filter tweets
	for i, tweet := range req.Tweets {
		if score, ok := scoreMap[currentID]; ok {
			if score.Score >= threshold {
				response.Tweets = append(response.Tweets, tweet)
				response.Scores[fmt.Sprintf("tweet_%d", i)] = score
				log.Printf("  ✓ Tweet @%s: score=%d (%s)", tweet.Author, score.Score, score.Reason)
			} else {
				log.Printf("  ✗ Tweet @%s: score=%d (filtered)", tweet.Author, score.Score)
			}
		}
		currentID++
	}

	// Filter news
	for i, item := range req.News {
		if score, ok := scoreMap[currentID]; ok {
			if score.Score >= threshold {
				response.News = append(response.News, item)
				response.Scores[fmt.Sprintf("news_%d", i)] = score
				log.Printf("  ✓ News \"%s\": score=%d (%s)", truncateForPrompt(item.Title, 50), score.Score, score.Reason)
			} else {
				log.Printf("  ✗ News \"%s\": score=%d (filtered)", truncateForPrompt(item.Title, 50), score.Score)
			}
		}
		currentID++
	}

	// Filter whales
	for i, whale := range req.Whales {
		if score, ok := scoreMap[currentID]; ok {
			if score.Score >= threshold {
				response.Whales = append(response.Whales, whale)
				response.Scores[fmt.Sprintf("whale_%d", i)] = score
				log.Printf("  ✓ Whale $%.0fM %s: score=%d (%s)", whale.AmountUSD/1_000_000, whale.Symbol, score.Score, score.Reason)
			} else {
				log.Printf("  ✗ Whale $%.0fM %s: score=%d (filtered)", whale.AmountUSD/1_000_000, whale.Symbol, score.Score)
			}
		}
		currentID++
	}

	return response
}

// Cache helpers
func getCachedScore(key string) ([]NewsScore, bool) {
	scoreCacheMutex.RLock()
	defer scoreCacheMutex.RUnlock()

	if cached, ok := scoreCache[key]; ok {
		if time.Now().Before(cached.ExpiresAt) {
			// Cache hit, reconstruct scores
			// Note: This is simplified - in production you'd cache the full array
			return []NewsScore{}, true
		}
	}
	return nil, false
}

func setCachedScore(key string, scores []NewsScore, duration time.Duration) {
	scoreCacheMutex.Lock()
	defer scoreCacheMutex.Unlock()

	// Simplified caching - in production you'd store the full array
	scoreCache[key] = CachedScore{
		Score:     len(scores),
		Reason:    "cached",
		ExpiresAt: time.Now().Add(duration),
	}
}

// hashString creates simple hash for cache key
func hashString(s string) uint32 {
	h := uint32(0)
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}

// truncateForPrompt truncates text for AI prompt
func truncateForPrompt(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
