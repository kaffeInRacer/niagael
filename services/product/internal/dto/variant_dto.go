package dto

import "encoding/json"

type CreateVariantDto struct {
	Id         string
	ProductId  string
	Name       string          `json:"name" validate:"required,max=255"`
	Price      int64           `json:"price" validate:"gte=0"`
	Stock      int64           `json:"stock" validate:"gte=0"`
	Attributes json.RawMessage `json:"attributes"`
	IsActive   bool            `json:"is_active"`
}

type UpdateVariantDto struct {
	Id         string
	ProductId  string
	Name       string          `json:"name" validate:"required,max=255"`
	Price      int64           `json:"price" validate:"gte=0"`
	Stock      int64           `json:"stock" validate:"gte=0"`
	Attributes json.RawMessage `json:"attributes"`
	IsActive   bool            `json:"is_active"`
}
