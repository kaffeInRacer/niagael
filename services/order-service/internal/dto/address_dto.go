package dto

type CreateAddressDto struct {
	Id         string
	UserId     string `json:"user_id" validate:"required"`
	Street     string `json:"street" validate:"required,max=255"`
	City       string `json:"city" validate:"required,max=100"`
	Province   string `json:"province" validate:"required,max=100"`
	PostalCode string `json:"postal_code" validate:"required,max=10"`
	Country    string `json:"country" validate:"required,max=100"`
	IsDefault  bool   `json:"is_default"`
}

type UpdateAddressDto struct {
	UserId     string
	Street     string `json:"street" validate:"required,max=255"`
	City       string `json:"city" validate:"required,max=100"`
	Province   string `json:"province" validate:"required,max=100"`
	PostalCode string `json:"postal_code" validate:"required,max=10"`
	Country    string `json:"country" validate:"required,max=100"`
	IsDefault  bool   `json:"is_default"`
}

type ListAddressParams struct {
	UserId string `form:"user_id" validate:"required"`
}
