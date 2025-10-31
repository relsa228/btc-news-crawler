package clients

import (
	"btc-news-crawler/models"
	"log"

	"github.com/jmoiron/sqlx"
)

type ClickhouseClient struct {
	ConnectionPool *sqlx.DB
}

func NewClickhouseClient(connectionUrl string) *ClickhouseClient {
	db := sqlx.MustConnect("clickhouse", connectionUrl)
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
		log.Fatal(err)
	}
}
