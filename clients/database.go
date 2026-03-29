package clients

import models "models"

type DatabaseClient interface {
	InsertNews(news *models.News)
	InsertQuote(quote *models.Quote)
	Migrate() error
}
