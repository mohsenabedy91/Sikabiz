package presenter

import (
	"github.com/mohsenabedy91/Sikabiz/internal/core/domain"
)

type User struct {
	ID          string    `json:"id" example:"8f4a1582-6a67-4d85-950b-2d17049c7385"`
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
		ID:          user.Base.UUID.String(),
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Addresses:   ToAddressCollection(user.Addresses),
	}
}

func ToUserResource(user *domain.User) *User {
	return PrepareUser(user)
}
