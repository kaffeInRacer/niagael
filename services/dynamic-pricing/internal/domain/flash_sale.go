package domain

import "time"

type FlashSale struct {
	Id              string     `json:"id"`
	Name            string     `json:"name"`
	ProductId       string     `json:"product_id"`
	VariantId       *string    `json:"variant_id"`
	DiscountPercent int        `json:"discount_percent"`
	Stock           int64      `json:"stock"`
	MaxPerUser      int        `json:"max_per_user"`
	StartTime       time.Time  `json:"start_time"`
	EndTime         time.Time  `json:"end_time"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}
