package port

import (
	"github.com/google/uuid"
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
)

type UserRepository interface {
	GetByID(id uuid.UUID) (*domain.User, error)
	Save(user *domain.User) (uint64, error)
}

type UserService interface {
	GetByID(uow UserUnitOfWork, id string) (*domain.User, error)
}
