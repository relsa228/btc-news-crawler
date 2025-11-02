package main

import (
	"btc-news-crawler/clients"
	"btc-news-crawler/services"
	env "btc-news-crawler/shared"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found")
		return
	}

	client := clients.NewClickhouseClient()
	if os.Getenv(env.CLICKHOUSE_CONNECTION_STRING_VAR) == "true" {
		if err := client.Migrate(); err != nil {
			log.Fatalf("🛑 Error creating tables: %v", err)
		}
	}

	news_service := services.NewNewsCrawlerService(client)
	quotes_service := services.NewQuotesCollectorService(client)

	var wg sync.WaitGroup
	wg.Go(news_service.StartCrawlers)
	wg.Go(quotes_service.StartQuotesCollecting)
	wg.Wait()
}
