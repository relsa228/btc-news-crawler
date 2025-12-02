package services

import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"btc-news-crawler/clients"
	"btc-news-crawler/models"
	"btc-news-crawler/models/configs"
	shared "btc-news-crawler/shared"
	consts "btc-news-crawler/shared/consts"
	logger "btc-news-crawler/shared/log"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

type NewsCrawlerService struct {
	Crawlers       map[string]*colly.Collector
	DatabaseClient *clients.PostgresClient
}

func (s *NewsCrawlerService) AddCrawlerFromConfig(config_path string) {
	data, err := os.ReadFile(config_path)
	if err != nil {
		logger.Log.Error("🛑 JSON reading error:", zap.Error(err))
		return
	}

	var config configs.NewsCrawlerConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		logger.Log.Error("🛑 JSON parsing error:", zap.Error(err))
	}

	re := regexp.MustCompile(config.ArticleIdentifier)

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(config.AllowedDomains...),
	)

	c.OnRequest(func(r *colly.Request) {
		logger.Log.Debug("🔗 Visiting", zap.String("url", r.URL.String()))
	})
	c.OnResponse(func(r *colly.Response) {
		logger.Log.Debug("✅ Received:", zap.String("url", r.Request.URL.String()))
	})

	c.IgnoreRobotsTxt = true

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4, RandomDelay: 2 * time.Second})

	c.OnHTML(consts.NEWS_LINK_SELECTOR, func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr(consts.NEWS_REFERENCE_SELECTOR))
		if link != "" {
			e.Request.Visit(link)
		}
	})

	c.OnHTML(config.ArticleSelector, func(e *colly.HTMLElement) {
		title := strings.TrimSpace(e.DOM.ParentsFiltered(consts.NEWS_BODY_SELECTOR).Find(config.TitleSelector).Text())
		content := strings.TrimSpace(e.DOM.ParentsFiltered(consts.NEWS_BODY_SELECTOR).Find(config.ContentSelector).Text())
		isMatch := re.MatchString(e.Request.URL.String())

		if title == "" || content == "" || !isMatch {
			return
		}
		news := new(models.News)
		news.Title = title
		news.Content = content
		news.Url = e.Request.URL.String()
		news.Source = config.Name
		news.PublicationDate = strings.TrimSpace(e.DOM.ParentsFiltered(consts.NEWS_BODY_SELECTOR).Find(config.DateSelector).Text())
		news.CreatedAt = time.Now()

		s.DatabaseClient.InsertNews(news)
		logger.Log.Info("📰 Got news", zap.String("title", news.Title), zap.String("source", news.Source))
	})

	c.OnError(func(resp *colly.Response, err error) {
		logger.Log.Error("🛑 Request error", zap.Error(err), zap.String("url", resp.Request.URL.String()))
	})
	c.OnScraped(func(_ *colly.Response) {
		logger.Log.Info("✅ Resource has been scraped", zap.String("name", config.Name))
	})

	s.Crawlers[config.StartURL] = c
	logger.Log.Info("✅ Crawler has been set up", zap.String("name", config.Name))
}

func (s *NewsCrawlerService) StartCrawlers() {
	batchSize, err := strconv.Atoi(os.Getenv(shared.CRAWLER_BATCH_SIZE_VAR))
	if err != nil {
		logger.Log.Error("🛑 Error parsing batch size", zap.Error(err))
		return
	}
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		crawlers := s.Crawlers
		var wg sync.WaitGroup
		count := 0
		for startURL, crawler := range crawlers {
			wg.Add(1)
			count++
			go func(url string, c *colly.Collector) {
				defer wg.Done()
				logger.Log.Info("🚀 Running crawler", zap.String("url", url))
				if err := c.Visit(url); err != nil {
					logger.Log.Error("🛑 Error crawling", zap.String("url", url), zap.Error(err))
				}
				c.Wait()
				logger.Log.Info("✅ Finished crawler", zap.String("url", url))
			}(startURL, crawler)

			if count%batchSize == 0 {
				wg.Wait()
				logger.Log.Info("✅ Batch completed, next batch starting...")
			}
		}
		wg.Wait()
		logger.Log.Info("🫡 All crawlers finished")
		<-ticker.C
	}
}

func NewNewsCrawlerService(c *clients.PostgresClient) *NewsCrawlerService {
	return &NewsCrawlerService{
		Crawlers:       make(map[string]*colly.Collector),
		DatabaseClient: c,
	}
}
