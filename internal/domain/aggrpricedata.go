package domain

import "time"

type AggrPriceData struct {
	ID           int
	PairName     string
	Exchange     string
	Timestamp    time.Time
	AveragePrice float64
	MinPrice     float64
	MaxPrice     float64
}
