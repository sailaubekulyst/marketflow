package ports

import "marketflow/internal/domain"

type AggrPriceDataRepository interface {
	AddNewAggregatedPriceData(NewAggrPriceData domain.AggrPriceData) error
	GelAllAggregatedPriceDataBySymbol(symbolName string) ([]domain.AggrPriceData, error)
	GelAllAggregatedPriceDataBySymbolAndPort(symbolName, port string) ([]domain.AggrPriceData, error)
}
