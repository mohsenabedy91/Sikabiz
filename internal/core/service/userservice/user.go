package userservice

import (
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
	"github.com/mohsenabedy91/Sikabiz/internal/core/port"
	"github.com/mohsenabedy91/Sikabiz/pkg/logger"
)

type UserService struct {
	log logger.Logger
}

func New(log logger.Logger) *UserService {
	return &UserService{
		log: log,
	}
}

func (r *UserService) GetByID(uow port.UserUnitOfWork, id uint64) (user *domain.User, err error) {
	user, err = uow.UserRepository().GetByID(id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
