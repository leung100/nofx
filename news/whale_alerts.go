package news

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// WhaleAlert represents a large cryptocurrency transaction
type WhaleAlert struct {
	Amount       float64
	AmountUSD    float64
	Symbol       string
	From         string
	FromOwner    string // Exchange name if applicable
	To           string
	ToOwner      string // Exchange name if applicable
	Timestamp    time.Time
	TxHash       string
	Blockchain   string
	Interpretation string // "accumulation", "distribution", "exchange_flow", etc.
}

// Whale Alert API response structures
type WhaleAlertResponse struct {
	Result      string              `json:"result"`
	Count       int                 `json:"count"`
	Transactions []WhaleTransaction `json:"transactions"`
}

type WhaleTransaction struct {
	Blockchain   string  `json:"blockchain"`
	Symbol       string  `json:"symbol"`
	TransactionType string `json:"transaction_type"`
	Hash         string  `json:"hash"`
	From         WhaleAddress `json:"from"`
	To           WhaleAddress `json:"to"`
	Timestamp    int64   `json:"timestamp"`
	Amount       float64 `json:"amount"`
	AmountUSD    float64 `json:"amount_usd"`
}

type WhaleAddress struct {
	Address string `json:"address"`
	Owner   string `json:"owner"`
	OwnerType string `json:"owner_type"` // "exchange", "wallet", etc.
}

// GetRecentWhaleAlerts fetches recent large crypto transactions
func GetRecentWhaleAlerts() ([]WhaleAlert, error) {
	// Whale Alert free tier: 10 calls/minute, 1000 calls/day
	// We'll call this conservatively (every 3 minutes = 480 calls/day)

	url := "https://api.whale-alert.io/v1/transactions"

	// Build query parameters
	// Get transactions from last hour
	minValue := 50000000 // Minimum $50M USD transactions (reduce noise)
	startTime := time.Now().Add(-1 * time.Hour).Unix()

	url = fmt.Sprintf("%s?api_key=demo&start=%d&min_value=%d", url, startTime, minValue)

	// Note: "demo" key is limited. Users should get free key from whale-alert.io
	// If WHALE_ALERT_API_KEY env var is set, use it instead
	// apiKey := os.Getenv("WHALE_ALERT_API_KEY")
	// if apiKey != "" {
	//     url = fmt.Sprintf("%s?api_key=%s&start=%d&min_value=%d", baseURL, apiKey, startTime, minValue)
	// }

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

	var result WhaleAlertResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Convert to WhaleAlert format
	var alerts []WhaleAlert
	for _, tx := range result.Transactions {
		alert := WhaleAlert{
			Amount:       tx.Amount,
			AmountUSD:    tx.AmountUSD,
			Symbol:       tx.Symbol,
			From:         tx.From.Address,
			FromOwner:    tx.From.Owner,
			To:           tx.To.Address,
			ToOwner:      tx.To.Owner,
			Timestamp:    time.Unix(tx.Timestamp, 0),
			TxHash:       tx.Hash,
			Blockchain:   tx.Blockchain,
			Interpretation: interpretWhaleMovement(tx),
		}
		alerts = append(alerts, alert)
	}

	log.Printf("✅ [News] Fetched %d whale alerts (>$1M transactions)", len(alerts))
	return alerts, nil
}

// interpretWhaleMovement provides interpretation of whale transaction
func interpretWhaleMovement(tx WhaleTransaction) string {
	fromExchange := tx.From.OwnerType == "exchange"
	toExchange := tx.To.OwnerType == "exchange"

	switch {
	case fromExchange && !toExchange:
		// Moving from exchange to wallet = potential accumulation (bullish)
		return "accumulation" // Holders withdrawing from exchanges
	case !fromExchange && toExchange:
		// Moving from wallet to exchange = potential distribution (bearish)
		return "distribution" // Holders depositing to sell
	case fromExchange && toExchange:
		// Exchange to exchange transfer = arbitrage or rebalancing (neutral)
		return "exchange_flow"
	case !fromExchange && !toExchange:
		// Wallet to wallet = OTC deal or personal transfer (neutral)
		return "wallet_transfer"
	default:
		return "unknown"
	}
}

// FilterRelevantWhaleAlerts filters whale alerts for major coins and recent transactions
func FilterRelevantWhaleAlerts(alerts []WhaleAlert) []WhaleAlert {
	relevantCoins := map[string]bool{
		"btc":  true,
		"eth":  true,
		"usdt": true,
		"usdc": true,
		"bnb":  true,
		"sol":  true,
		"xrp":  true,
		"ada":  true,
	}

	now := time.Now()
	var critical []WhaleAlert // >$100M or <30 min
	var important []WhaleAlert // $50M-100M and <1 hour

	for _, alert := range alerts {
		symbol := alert.Symbol
		if !relevantCoins[symbol] {
			continue
		}

		age := now.Sub(alert.Timestamp).Minutes()
		amountM := alert.AmountUSD / 1_000_000

		// Critical: >$100M or very recent large moves
		if amountM >= 100 || (amountM >= 50 && age < 30) {
			critical = append(critical, alert)
		} else if amountM >= 50 && age < 60 {
			important = append(important, alert)
		}
	}

	// Combine: critical first, then important, limit to 3 total
	result := critical
	for i := 0; i < len(important) && len(result) < 3; i++ {
		result = append(result, important[i])
	}

	return result
}
