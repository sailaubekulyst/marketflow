package sqlite

import (
	"database/sql"
	"fmt"

	"marketflow/internal/domain"
	"marketflow/internal/ports"
)

type PriceRepositorySqlite struct {
	db *sql.DB
}

// Конструктор
func GetPriceRepositorySqlite(db *sql.DB) ports.PriceRepository {
	return &PriceRepositorySqlite{
		db: db,
	}
}

// AddNewPrice добавляет новый Price в таблицу
func (r *PriceRepositorySqlite) AddNewPrice(newPrice domain.Price) error {
	query := `
		INSERT INTO prices (symbol, exchange, price, timestamp)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		newPrice.Symbol,
		newPrice.Exchange,
		newPrice.Price,
		newPrice.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to insert price: %w", err)
	}

	return nil
}
