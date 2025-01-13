package domain

type User struct {
	Base
	Modifier

	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Email       string
	PhoneNumber string `json:"phone_number"`

	Addresses []*Address
}
