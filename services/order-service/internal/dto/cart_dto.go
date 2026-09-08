package dto

import "time"

type AddCartItemDto struct {
	UserId    string `json:"user_id" validate:"required"`
	ProductId string `json:"product_id" validate:"required,uuid"`
	VariantId string `json:"variant_id" validate:"omitempty,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0,lte=100"`
}

type UpdateCartItemDto struct {
	UserId   string `json:"user_id" validate:"required"`
	Quantity int    `json:"quantity" validate:"required,gt=0,lte=100"`
}

type CartUserParams struct {
	UserId string `form:"user_id" validate:"required"`
}

type CartResponse struct {
	Id         string             `json:"id"`
	UserId     string             `json:"user_id"`
	Items      []CartItemResponse `json:"items"`
	Subtotal   int64              `json:"subtotal"`
	TotalItems int                `json:"total_items"`
	CreatedAt  *time.Time         `json:"created_at,omitempty"`
	UpdatedAt  *time.Time         `json:"updated_at,omitempty"`
}

type CartItemResponse struct {
	Id                       string            `json:"id"`
	ProductId                string            `json:"product_id"`
	VariantId                *string           `json:"variant_id,omitempty"`
	ProductName              string            `json:"product_name"`
	VariantName              string            `json:"variant_name,omitempty"`
	VariantAttributes        map[string]string `json:"variant_attributes,omitempty"`
	Quantity                 int               `json:"quantity"`
	AvailableStock           int64             `json:"available_stock"`
	OriginalUnitPrice        int64             `json:"original_unit_price"`
	FlashSaleUnitPrice       int64             `json:"flash_sale_unit_price"`
	FlashSaleQuantity        int               `json:"flash_sale_quantity"`
	NormalQuantity           int               `json:"normal_quantity"`
	LineTotal                int64             `json:"line_total"`
	FlashSaleId              *string           `json:"flash_sale_id,omitempty"`
	FlashSaleName            *string           `json:"flash_sale_name,omitempty"`
	FlashSaleDiscountPercent int32             `json:"flash_sale_discount_percent"`
}
