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

	"btc-news-crawler/models/configs"

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

	var config configs.NewsCrawlerConfig
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
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		for startURL, crawler := range s.Crawlers {
			go func(url string, c *colly.Collector) {
				log.Printf("🚀 Running news crawler for %s", url)
				if err := c.Visit(url); err != nil {
					log.Printf("🛑 News scraping error for %s: %v", url, err)
				}
			}(startURL, crawler.Clone())
		}
	}
}

func NewNewsCrawlerService(c *clients.ClickhouseClient) *NewsCrawlerService {
	return &NewsCrawlerService{
		Crawlers:         make(map[string]colly.Collector),
		ClickhouseClient: c,
	}
}
