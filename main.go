package main

import (
	"btc-news-crawler/clients"
	"btc-news-crawler/services"
	env "btc-news-crawler/shared"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found")
		return
	}

	// Collecting configs paths
	var files []string
	err := filepath.WalkDir(os.Getenv(env.CONFIG_DIR_VAR), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("🛑 Config directory read error: %v", err)
	}

	client := clients.NewDatabaseClient()
	if os.Getenv(env.NEED_MIGRATION_VAR) == "true" {
		if err := client.Migrate(); err != nil {
			log.Fatalf("🛑 Error creating tables: %v", err)
		}
	}

	news_service := services.NewNewsCrawlerService(client)
	quotes_service := services.NewQuotesCollectorService(client)

	for _, file := range files {
		news_service.AddCrawlerFromConfig(file)
	}

	var wg sync.WaitGroup

	wg.Go(news_service.StartCrawlers)
	wg.Go(quotes_service.StartQuotesCollecting)
	wg.Wait()
}
