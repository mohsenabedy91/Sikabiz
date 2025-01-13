package domain

type User struct {
	Base
	Modifier

	Name        *string
	Email       string
	PhoneNumber string

	Addresses []*Address
}
