package port

import "github.com/mohsenabedy91/Sikabiz/internal/core/domain"

type AddressRepository interface {
	Save(userID uint64, address []*domain.Address) error
}
