package main

import (
	"btc-news-crawler/clients"
	"btc-news-crawler/services"
	shared "btc-news-crawler/shared"
	logger "btc-news-crawler/shared/log"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	defer logger.Log.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Log.Error("🛑 .env file not found")
		return
	}

	// Collecting configs paths
	var files []string
	err := filepath.WalkDir(os.Getenv(shared.CONFIG_DIR_VAR), func(path string, d os.DirEntry, err error) error {
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
		logger.Log.Error("🛑 Config directory read error: ", zap.Error(err))
	}

	client := clients.NewDatabaseClient()
	if os.Getenv(shared.NEED_MIGRATION_VAR) == "true" {
		if err := client.Migrate(); err != nil {
			logger.Log.Error("🛑 Error creating tables: ", zap.Error(err))
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
