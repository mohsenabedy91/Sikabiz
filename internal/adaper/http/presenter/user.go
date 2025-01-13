package presenter

import (
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
)

type User struct {
	ID          uint64    `json:"id" example:"1"`
	Name        *string   `json:"firstName,omitempty" example:"john doe"`
	Email       string    `json:"email,omitempty" example:"john.doe@gmail.com"`
	PhoneNumber string    `json:"phone_number,omitempty" example:"09121111111"`
	Addresses   []Address `json:"addresses,omitempty"`
}

func PrepareUser(user *domain.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:          user.Base.ID,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Addresses:   ToAddressCollection(user.Addresses),
	}
}

func ToUserResource(user *domain.User) *User {
	return PrepareUser(user)
}
