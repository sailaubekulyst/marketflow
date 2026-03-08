package domain

type Price struct {
	Symbol    string `json:"symbol"`
	Exchange  string
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}
