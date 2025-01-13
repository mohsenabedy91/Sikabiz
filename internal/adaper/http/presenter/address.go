package presenter

import "github.com/mohsenabedy91/Sikabiz/internal/core/domain"

type Address struct {
	Street  *string `json:"street,omitempty" example:"817 East Lodgeville"`
	City    *string `json:"city,omitempty" example:"New York City"`
	State   *string `json:"state,omitempty" example:"Arkansas"`
	ZipCode *string `json:"zip_code,omitempty" example:"58532"`
	Country *string `json:"country,omitempty" example:"France"`
}

func PrepareAddress(address *domain.Address) *Address {
	if address == nil {
		return nil
	}

	return &Address{
		Street:  address.Street,
		City:    address.City,
		State:   address.State,
		ZipCode: address.ZipCode,
		Country: address.Country,
	}
}

func ToAddressCollection(addresses []*domain.Address) []Address {
	var response []Address
	for _, address := range addresses {
		result := PrepareAddress(address)
		if result != nil {
			response = append(response, *result)
		}
	}

	return response
}
