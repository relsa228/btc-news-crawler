package services

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"btc-news-crawler/clients"
	"btc-news-crawler/models"
	"btc-news-crawler/models/responses"

	env "btc-news-crawler/shared"
)

const (
	REQUEST_METHOD = "GET"
	AUTH_HEADER    = "X-CMC_PRO_API_KEY"
)

type QuotesCollectorService struct {
	ApiKey           string
	Endpoint         string
	ClickhouseClient *clients.ClickhouseClient
}

func (s *QuotesCollectorService) StartQuotesCollecting() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		req, _ := http.NewRequest(REQUEST_METHOD, s.Endpoint, nil)
		req.Header.Add(AUTH_HEADER, s.ApiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("🛑 Quote response error: %s", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		var apiResp responses.CoinmarketcapResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			log.Printf("🛑 JSON parsing error: %s", err)
		}

		if apiResp.Status.ErrorCode != 0 {
			log.Printf("🛑 CoinMarketCap error: %s", apiResp.Status.ErrorMessage)
		}

		btc := apiResp.Data[0]

		var q = models.Quote{
			Price:            btc.Quote.USD.Price,
			PercentChange1h:  btc.Quote.USD.PercentChange1h,
			PercentChange24h: btc.Quote.USD.PercentChange24h,
			PercentChange7d:  btc.Quote.USD.PercentChange7d,
			Ticker:           btc.Symbol,
			Date:             time.Now().UTC(),
		}

		s.ClickhouseClient.InsertQuote(&q)
	}
}

func NewQuotesCollectorService(c *clients.ClickhouseClient) *QuotesCollectorService {
	api_key := os.Getenv(env.API_KEY_ENV_VAR)
	url := os.Getenv(env.API_ENDPOINT_ENV_VAR)
	return &QuotesCollectorService{
		ApiKey:           api_key,
		Endpoint:         url,
		ClickhouseClient: c,
	}
}
