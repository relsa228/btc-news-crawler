package services

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"btc-news-crawler/clients"
	"btc-news-crawler/models"
	"btc-news-crawler/models/configs"
	consts "btc-news-crawler/shared/consts"

	"github.com/gocolly/colly/v2"
)

type NewsCrawlerService struct {
	Crawlers       map[string]*colly.Collector
	DatabaseClient *clients.PostgresClient
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

	re := regexp.MustCompile(config.ArticleIdentifier)

	c := colly.NewCollector(
		colly.Async(true),
		colly.AllowedDomains(config.AllowedDomains...),
	)

	// c.OnRequest(func(r *colly.Request) {
	// 	log.Printf("🔗 Visiting: %s", r.URL.String())
	// })
	// c.OnResponse(func(r *colly.Response) {
	// 	log.Printf("✅ Received: %s", r.Request.URL)
	// })
	//c.SetRequestTimeout(15 * time.Second)

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
		log.Printf("📰 Got news %s [%s]\n", news.Title, news.Source)

	})

	c.OnError(func(resp *colly.Response, err error) {
		log.Printf("🛑 Request error %s: %v", resp.Request.URL, err)
	})
	c.OnScraped(func(_ *colly.Response) {
		log.Printf("✅ Resource %s has been scraped \n", config.Name)
	})

	s.Crawlers[config.StartURL] = c
	log.Printf("✅ Crawler for %s has been set up\n", config.Name)
}

func (s *NewsCrawlerService) StartCrawlers() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		for startURL, crawler := range s.Crawlers {
			go func(url string, c *colly.Collector) {
				log.Printf("🚀 Running news crawler for %s", url)
				if err := c.Visit(url); err != nil {
					log.Printf("🛑 News scraping error for %s: %v", url, err)
				}
				c.Wait()
				log.Printf("✅ News crawler for %s has finished", url)
			}(startURL, crawler)
		}
		<-ticker.C
	}
}

func NewNewsCrawlerService(c *clients.PostgresClient) *NewsCrawlerService {
	return &NewsCrawlerService{
		Crawlers:       make(map[string]*colly.Collector),
		DatabaseClient: c,
	}
}
