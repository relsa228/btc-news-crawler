package models

import "time"

type News struct {
	Url             string    `db:"url"`
	Title           string    `db:"title"`
	Content         string    `db:"content"`
	Source          string    `db:"source"`
	PublicationDate string    `db:"publication_date"`
	CreatedAt       time.Time `db:"created_at"`
}
