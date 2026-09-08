package domain

import "time"

type Cart struct {
	Id        string     `json:"id"`
	UserId    string     `json:"user_id"`
	Items     []CartItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type CartItem struct {
	Id        string     `json:"id"`
	CartId    string     `json:"cart_id"`
	ProductId string     `json:"product_id"`
	VariantId *string    `json:"variant_id"`
	Quantity  int        `json:"quantity"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
