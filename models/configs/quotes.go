package configs

type CoinQuotesCrawlerConfig struct {
	TickerSelector string   `json:"ticker_selector"`
	PriceSelector  string   `json:"price_selector"`
	StartURL       string   `json:"start_url"`
	AllowedDomains []string `json:"allowed_domains"`
	FollowLinks    bool     `json:"follow_links"`
}
