package domain

import "time"

type OrderItem struct {
	Id        string  `json:"id"`
	OrderId   string  `json:"order_id"`
	ProductId string  `json:"product_id"`
	VariantId *string `json:"variant_id"`

	// Product Snapshot
	ProductName  string `json:"product_name"`
	ProductPrice int64  `json:"product_price"`
	Quantity     int    `json:"quantity"`

	// Flash Sale Snapshot
	FlashSaleId              *string `json:"flash_sale_id"`
	FlashSaleName            *string `json:"flash_sale_name"`
	FlashSaleDiscountPercent *int    `json:"flash_sale_discount_percent"`
	FlashSaleDiscountPrice   *int64  `json:"flash_sale_discount_price"`
	FlashSaleOriginalPrice   *int64  `json:"flash_sale_original_price"`
	FlashSaleQuantity        *int    `json:"flash_sale_quantity"`

	// Promo Snapshot
	PromoId             *string `json:"promo_id"`
	PromoCode           *string `json:"promo_code"`
	PromoName           *string `json:"promo_name"`
	PromoDiscountType   *string `json:"promo_discount_type"`
	PromoDiscountAmount *int64  `json:"promo_discount_amount"`

	// Final
	FinalPrice int64      `json:"final_price"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at"`

	// Enriched at read time (not persisted)
	VariantName       string            `json:"variant_name,omitempty"`
	VariantAttributes map[string]string `json:"variant_attributes,omitempty"`
}
