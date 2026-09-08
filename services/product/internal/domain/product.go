package domain

import "time"

type Product struct {
	Id            string           `json:"id"`
	CategoryId    string           `json:"category_id"`
	CategoryName  string           `json:"category_name"`
	Name          string           `json:"name"`
	Slug          string           `json:"slug"`
	Description   string           `json:"description"`
	Price         int64            `json:"price"`
	Stock         int64            `json:"stock"`
	StockReserved int64            `json:"stock_reserved"`
	IsActive      bool             `json:"is_active"`
	IsPromo       bool             `json:"is_promo_excluded"`
	Images        []ProductImage   `json:"images,omitempty"`
	Variants      []ProductVariant `json:"variants,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     *time.Time       `json:"updated_at"`
	DeletedAt     *time.Time       `json:"deleted_at"`
}
