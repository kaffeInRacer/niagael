package dto

import (
	"time"

	"kaffein/dynamic-pricing-service/internal/domain"
)

type CreatePromoDto struct {
	Id                  string
	Name                string    `json:"name" validate:"required,max=255"`
	Code                string    `json:"code" validate:"required,max=50"`
	Description         string    `json:"description"`
	DiscountType        string    `json:"discount_type" validate:"required,def_enum=percentage fixed"`
	DiscountValue       int64     `json:"discount_value" validate:"gt=0"`
	MinPurchase         int64     `json:"min_purchase" validate:"gte=0"`
	MaxDiscount         int64     `json:"max_discount" validate:"gte=0"`
	Quantity            int       `json:"quantity" validate:"gte=0"`
	MaxUsagePerUser     int       `json:"max_usage_per_user" validate:"gte=0"`
	CanCombineFlashSale bool      `json:"can_combine_flash_sale"`
	StartDate           time.Time `json:"start_date" validate:"required"`
	EndDate             time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	IsActive            bool      `json:"is_active"`
}

type UpdatePromoDto struct {
	Id                  string
	Name                string    `json:"name" validate:"required,max=255"`
	Code                string    `json:"code" validate:"required,max=50"`
	Description         string    `json:"description"`
	DiscountType        string    `json:"discount_type" validate:"required,def_enum=percentage fixed"`
	DiscountValue       int64     `json:"discount_value" validate:"gt=0"`
	MinPurchase         int64     `json:"min_purchase" validate:"gte=0"`
	MaxDiscount         int64     `json:"max_discount" validate:"gte=0"`
	Quantity            int       `json:"quantity" validate:"gte=0"`
	MaxUsagePerUser     int       `json:"max_usage_per_user" validate:"gte=0"`
	CanCombineFlashSale bool      `json:"can_combine_flash_sale"`
	StartDate           time.Time `json:"start_date" validate:"required"`
	EndDate             time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
	IsActive            bool      `json:"is_active"`
}

type ListPromoParams struct {
	Search      string `form:"search" validate:"omitempty,max=100"`
	IsActive    *bool  `form:"is_active" validate:"omitempty"`
	OrderBy     string `form:"order_by" validate:"def_enum=name code start_date created_at"`
	OrderDir    string `form:"order_dir" validate:"def_enum=desc asc"`
	PageSize    int32  `form:"page_size" validate:"clamp=10 100"`
	PageOffset  int32  `form:"page" validate:"clamp=0"`
	CurrentOnly bool   `form:"-"`
}

type PublicPromoResponse struct {
	Id                  string    `json:"id"`
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	DiscountType        string    `json:"discount_type"`
	DiscountValue       int64     `json:"discount_value"`
	MinPurchase         int64     `json:"min_purchase"`
	MaxDiscount         int64     `json:"max_discount"`
	CanCombineFlashSale bool      `json:"can_combine_flash_sale"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
}

func ToPublicPromoResponse(p domain.Promo) PublicPromoResponse {
	return PublicPromoResponse{
		Id: p.Id, Name: p.Name, Description: p.Description,
		DiscountType: p.DiscountType, DiscountValue: p.DiscountValue,
		MinPurchase: p.MinPurchase, MaxDiscount: p.MaxDiscount,
		CanCombineFlashSale: p.CanCombineFlashSale,
		StartDate:           p.StartDate, EndDate: p.EndDate,
	}
}
