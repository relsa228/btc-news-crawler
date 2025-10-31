package main

import (
	"btc-news-crawler/services"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found")
		return
	}

	service := services.NewNewsCrawlerService()
	// add crawlers
	service.StartCrawlers()
}
