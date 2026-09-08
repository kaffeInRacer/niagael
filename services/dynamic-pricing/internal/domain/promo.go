package domain

import "time"

type Promo struct {
	Id                  string     `json:"id"`
	Name                string     `json:"name"`
	Code                string     `json:"code"`
	Description         string     `json:"description"`
	DiscountType        string     `json:"discount_type"`
	DiscountValue       int64      `json:"discount_value"`
	MinPurchase         int64      `json:"min_purchase"`
	MaxDiscount         int64      `json:"max_discount"`
	Quantity            int        `json:"quantity"`
	UsedCount           int        `json:"used_count"`
	MaxUsagePerUser     int        `json:"max_usage_per_user"`
	CanCombineFlashSale bool       `json:"can_combine_flash_sale"`
	StartDate           time.Time  `json:"start_date"`
	EndDate             time.Time  `json:"end_date"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at"`
}
