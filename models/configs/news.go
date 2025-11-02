package configs

type NewsCrawlerConfig struct {
	Name            string   `json:"name"`
	StartURL        string   `json:"start_url"`
	AllowedDomains  []string `json:"allowed_domains"`
	ArticleSelector string   `json:"article_selector"`
	TitleSelector   string   `json:"title_selector"`
	ContentSelector string   `json:"content_selector"`
	DateSelector    string   `json:"date_selector"`
	FollowLinks     bool     `json:"follow_links"`
}
