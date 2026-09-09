package domain

import "time"

type FlashSaleProjection struct {
	ProductID       string    `json:"product_id"`
	VariantID       string    `json:"variant_id"`
	Name            string    `json:"name"`
	DiscountPercent int       `json:"discount_percent"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	IsActive        bool      `json:"is_active"`
}
