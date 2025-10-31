package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"btc-news-crawler/clients"
	"btc-news-crawler/models"

	"github.com/gocolly/colly/v2"
)

type NewsCrawlerService struct {
	Crawlers         map[string]colly.Collector
	ClickhouseClient *clients.ClickhouseClient
}

func (s *NewsCrawlerService) AddCrawlerFromConfig(config_path string) {
	data, err := os.ReadFile(config_path)
	if err != nil {
		log.Fatal(err)
	}

	var config models.CrawlerConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("🛑 JSON parsing error:", err)
	}

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(config.AllowedDomains...),
	)
	c.SetRequestTimeout(15 * time.Second)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4, RandomDelay: 2 * time.Second})

	c.OnHTML(config.TitleSelector, func(e *colly.HTMLElement) {
		title := strings.TrimSpace(e.Text)
		content := strings.TrimSpace(e.DOM.ParentsFiltered("body").Find(config.ContentSelector).Text())
		if title != "" && content != "" {
			news := new(models.News)
			news.Title = title
			news.Content = content
			news.Url = e.Request.URL.String()
			news.Source = config.Name
			news.PublicationDate = strings.TrimSpace(e.DOM.ParentsFiltered("body").Find(config.DateSelector).Text())
			news.CreatedAt = time.Now()

			s.ClickhouseClient.InsertNews(news)
			fmt.Printf("📰 Got news %s [%s]\n", news.Title, news.Source)
		}
	})
	c.OnError(func(resp *colly.Response, err error) {
		log.Printf("🛑 Request error %s: %v", resp.Request.URL, err)
	})
	c.OnScraped(func(_ *colly.Response) {
		log.Printf("✅ Resource %s has been scraped \n", config.Name)
	})

	s.Crawlers[config.StartURL] = *c.Clone()
}

func (s *NewsCrawlerService) StartCrawlers() {
	for start_URL, crawler := range s.Crawlers {
		crawler.Visit(start_URL)
	}
}

func NewNewsCrawlerService() *NewsCrawlerService {
	connection_string := os.Getenv("CLICKHOUSE_CONNECTION_STRING")
	return &NewsCrawlerService{
		Crawlers:         make(map[string]colly.Collector),
		ClickhouseClient: clients.NewClickhouseClient(connection_string),
	}
}
