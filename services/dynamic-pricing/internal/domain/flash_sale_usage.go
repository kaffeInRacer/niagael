package domain

import "time"

type FlashSaleUsage struct {
	Id          string     `json:"id"`
	FlashSaleId string     `json:"flash_sale_id"`
	UserId      string     `json:"user_id"`
	Quantity    int        `json:"quantity"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
