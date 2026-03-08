package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"marketflow/internal/domain"
	"marketflow/internal/service"
)

type Handler struct {
	service *service.Service
}

func GetHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

type HighestPriceResponse struct {
	Symbol    string    `json:"symbol"`
	Exchange  string    `json:"exchange"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"` // <- теперь время, а не int64
}

func (h *Handler) GetHighestPriceHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	var exchange, symbol string

	if len(parts) == 4 {
		symbol = parts[3]
	} else if len(parts) >= 5 {
		exchange = parts[3]
		symbol = parts[4]
	} else {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	period := r.URL.Query().Get("period")
	var priceData domain.Price
	var err error
	var ExchangeName string
	ExchangeName = exchange
	if exchange != "" {
		for exchangename, port := range map[string]string{
			"exchange1": "40101",
			"exchange2": "40102",
			"exchange3": "40303",
		} {
			if exchangename == exchange {
				exchange = port
			}
		}
		priceData, err = h.service.GetHighestPriceBySymbolAndExchange(symbol, exchange, period)
	} else {
		priceData, err = h.service.GetHighestPriceBySymbol(symbol, period)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	// Преобразуем в структуру для вывода
	resp := HighestPriceResponse{
		Symbol:    priceData.Symbol,
		Exchange:  ExchangeName,
		Price:     priceData.Price,                   // или MaxPrice, если нужно
		Timestamp: time.Unix(priceData.Timestamp, 0), // <- конвертация,    // <- time.Time
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
