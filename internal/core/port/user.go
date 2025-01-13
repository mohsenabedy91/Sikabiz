package port

import "github.com/mohsenabedy91/Sikabiz/internal/core/domain"

type UserRepository interface {
	GetByID(id uint64) (*domain.User, error)
}

type UserService interface {
	GetByID(uow UserUnitOfWork, id uint64) (*domain.User, error)
}
