package models

import "time"

type Quote struct {
	Price            float64   `db:"price"`
	PercentChange1h  float64   `db:"percent_change_1h"`
	PercentChange24h float64   `db:"percent_change_24h"`
	PercentChange7d  float64   `db:"percent_change_7d"`
	Ticker           string    `db:"ticker"`
	Date             time.Time `db:"date"`
}
