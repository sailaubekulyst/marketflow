package ports

import "marketflow/internal/domain"

type PriceRepository interface {
	AddNewPrice(NewPrice domain.Price) error
}
