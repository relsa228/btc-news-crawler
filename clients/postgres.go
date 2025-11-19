package clients

import (
	"btc-news-crawler/models"
	env "btc-news-crawler/shared"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresClient struct {
	ConnectionPool *sqlx.DB
}

func NewDatabaseClient() *PostgresClient {
	connection_string := os.Getenv(env.DATABASE_CONNECTION_STRING_VAR)
	db := sqlx.MustConnect("postgres", connection_string)
	return &PostgresClient{
		ConnectionPool: db,
	}
}

func (c *PostgresClient) InsertNews(news *models.News) {
	query := `
		INSERT INTO news (url, title, content, publication_date, source)
		VALUES (:url, :title, :content, :publication_date, :source)
		ON CONFLICT (url, source) DO UPDATE SET
			url = EXCLUDED.url,
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			publication_date = EXCLUDED.publication_date,
			source = EXCLUDED.source
	`
	_, err := c.ConnectionPool.NamedExec(query, news)
	if err != nil {
		log.Printf("🛑 Error inserting news: %v", err)
	}
}

func (c *PostgresClient) InsertQuote(quote *models.Quote) {
	query := `
		INSERT INTO quotes (price, percent_change_1h, percent_change_24h, percent_change_7d, ticker)
		VALUES (:price, :percent_change_1h, :percent_change_24h, :percent_change_7d, :ticker)
	`
	_, err := c.ConnectionPool.NamedExec(query, quote)
	if err != nil {
		log.Printf("🛑 Error inserting quote: %v", err)
	}
}

func (c *PostgresClient) Migrate() error {
	query := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

		CREATE TABLE IF NOT EXISTS quotes (
			id uuid NOT NULL DEFAULT uuid_generate_v4 () PRIMARY KEY,
		    price DOUBLE PRECISION,
		    percent_change_1h DOUBLE PRECISION,
		    percent_change_24h DOUBLE PRECISION,
		    percent_change_7d DOUBLE PRECISION,
		    ticker TEXT,
		    date timestamptz NOT NULL DEFAULT now ()
		);

		CREATE TABLE IF NOT EXISTS news (
		    url TEXT,
		    title TEXT,
		    content TEXT,
		    publication_date TEXT,
		    created_at timestamptz NOT NULL DEFAULT now (),
		    source TEXT,
		    PRIMARY KEY (url, source)
		);
	`
	_, err := c.ConnectionPool.Exec(query)
	return err
}
