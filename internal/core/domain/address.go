package domain

type Address struct {
	Base
	Modifier

	Street  *string
	City    *string
	State   *string
	ZipCode *string
	Country *string
}
