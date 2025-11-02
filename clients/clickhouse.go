package clients

import (
	"btc-news-crawler/models"
	env "btc-news-crawler/shared"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
)

type ClickhouseClient struct {
	ConnectionPool *sqlx.DB
}

func NewClickhouseClient() *ClickhouseClient {
	connection_string := os.Getenv(env.CLICKHOUSE_CONNECTION_STRING_VAR)
	db := sqlx.MustConnect("clickhouse", connection_string)
	return &ClickhouseClient{
		ConnectionPool: db,
	}
}

func (c *ClickhouseClient) InsertNews(news *models.News) {
	query := `
		INSERT INTO news (url, title, content, publication_date, created_at, source)
		VALUES (:url, :title, :content, :publication_date, :created_at, :source)
	`
	_, err := c.ConnectionPool.NamedExec(query, news)
	if err != nil {
		log.Printf("🛑 Error inserting news: %v", err)
	}
}

func (c *ClickhouseClient) InsertQuote(quote *models.Quote) {
	query := `
		INSERT INTO quotes (price, percent_change_1h, percent_change_24h, percent_change_7d, ticker, date)
		SELECT :price, :percent_change_1h, :percent_change_24h, :percent_change_7d, :ticker, :date
	`
	_, err := c.ConnectionPool.NamedExec(query, quote)
	if err != nil {
		log.Printf("🛑 Error inserting quote: %v", err)
	}
}

func (c *ClickhouseClient) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS quotes (
			price Float64,
			percent_change_1h Float64,
			percent_change_24h Float64,
			percent_change_7d Float64,
			ticker String,
			date DateTime
		) ENGINE = MergeTree()
		ORDER BY (date);

		CREATE TABLE IF NOT EXISTS news (
			url String,
			title String,
			content String,
			publication_date String,
			created_at DateTime,
			source String
		) ENGINE = ReplacingMergeTree(-toUnixTimestamp(created_at))
		ORDER BY (title, publication_date, url, source);
	`
	_, err := c.ConnectionPool.Exec(query)
	return err
}
