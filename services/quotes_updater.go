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
	consts "btc-news-crawler/shared/consts"

	env "btc-news-crawler/shared"
)

type QuotesCollectorService struct {
	ApiKey         string
	Endpoint       string
	DatabaseClient *clients.PostgresClient
}

func (s *QuotesCollectorService) StartQuotesCollecting() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		req, _ := http.NewRequest(consts.QUOTE_REQUEST_METHOD, s.Endpoint, nil)
		req.Header.Add(consts.QUOTE_AUTH_HEADER, s.ApiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("🛑 Quote response error: %s", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("🛑 Error reading response body: %s", err)
			return
		}

		var apiResp responses.CoinmarketcapResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			log.Printf("🛑 JSON parsing error: %s", err)
		}

		if apiResp.Status.ErrorCode != 0 {
			log.Printf("🛑 CoinMarketCap error: %s", apiResp.Status.ErrorMessage)
		}

		btc := apiResp.Data["1"]

		var q = models.Quote{
			Price:            btc.Quote.USD.Price,
			PercentChange1h:  btc.Quote.USD.PercentChange1h,
			PercentChange24h: btc.Quote.USD.PercentChange24h,
			PercentChange7d:  btc.Quote.USD.PercentChange7d,
			Ticker:           btc.Symbol,
			Date:             time.Now().UTC(),
		}

		s.DatabaseClient.InsertQuote(&q)

		<-ticker.C
	}
}

func NewQuotesCollectorService(c *clients.PostgresClient) *QuotesCollectorService {
	api_key := os.Getenv(env.API_KEY_VAR)
	url := os.Getenv(env.API_ENDPOINT_VAR)
	return &QuotesCollectorService{
		ApiKey:         api_key,
		Endpoint:       url,
		DatabaseClient: c,
	}
}
