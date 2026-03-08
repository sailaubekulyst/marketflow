package sqlite

import (
	"database/sql"
	"fmt"

	"marketflow/internal/domain"
	"marketflow/internal/ports"
)

type AggrPriceDataRepositorySqlite struct {
	db *sql.DB
}

// Конструктор
func GetAggrPriceDataRepositorySqlite(db *sql.DB) ports.AggrPriceDataRepository {
	return &AggrPriceDataRepositorySqlite{
		db: db,
	}
}

// GelAllAggregatedPriceDataBySymbol возвращает все записи для конкретной пары
func (r *AggrPriceDataRepositorySqlite) GelAllAggregatedPriceDataBySymbol(
	symbolName string,
) ([]domain.AggrPriceData, error) {

	query := `
        SELECT
            id,
            pair_name,
            exchange,
            timestamp,
            average_price,
            min_price,
            max_price
        FROM aggr_price_data
        WHERE pair_name = ?
        ORDER BY timestamp ASC
    `

	rows, err := r.db.Query(query, symbolName)
	if err != nil {
		return nil, fmt.Errorf("failed to query aggregated price data: %w", err)
	}
	defer rows.Close()

	results := make([]domain.AggrPriceData, 0)

	for rows.Next() {
		var d domain.AggrPriceData

		if err := rows.Scan(
			&d.ID,
			&d.PairName,
			&d.Exchange,
			&d.Timestamp, // ← сразу time.Time
			&d.AveragePrice,
			&d.MinPrice,
			&d.MaxPrice,
		); err != nil {
			return nil, fmt.Errorf("failed to scan aggregated price data: %w", err)
		}

		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// AddNewAggregatedPriceData добавляет новую запись агрегированных цен
func (r *AggrPriceDataRepositorySqlite) AddNewAggregatedPriceData(
	newData domain.AggrPriceData,
) error {

	query := `
        INSERT INTO aggr_price_data
            (pair_name, exchange, timestamp, average_price, min_price, max_price)
        VALUES (?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.Exec(
		query,
		newData.PairName,
		newData.Exchange,
		newData.Timestamp, // ← передаём time.Time
		newData.AveragePrice,
		newData.MinPrice,
		newData.MaxPrice,
	)
	if err != nil {
		return fmt.Errorf("failed to insert aggregated price data: %w", err)
	}

	return nil
}

// GelAllAggregatedPriceDataBySymbolAndPort возвращает все записи для пары и конкретного Exchange/Port
func (r *AggrPriceDataRepositorySqlite) GelAllAggregatedPriceDataBySymbolAndPort(
	symbolName, port string,
) ([]domain.AggrPriceData, error) {

	query := `
        SELECT
            id,
            pair_name,
            exchange,
            timestamp,
            average_price,
            min_price,
            max_price
        FROM aggr_price_data
        WHERE pair_name = ? AND exchange = ?
        ORDER BY timestamp ASC
    `

	rows, err := r.db.Query(query, symbolName, port)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to query aggregated price data by symbol and port: %w",
			err,
		)
	}
	defer rows.Close()

	results := make([]domain.AggrPriceData, 0)

	for rows.Next() {
		var d domain.AggrPriceData

		if err := rows.Scan(
			&d.ID,
			&d.PairName,
			&d.Exchange,
			&d.Timestamp, // ← сразу time.Time
			&d.AveragePrice,
			&d.MinPrice,
			&d.MaxPrice,
		); err != nil {
			return nil, fmt.Errorf("failed to scan aggregated price data: %w", err)
		}

		results = append(results, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}
