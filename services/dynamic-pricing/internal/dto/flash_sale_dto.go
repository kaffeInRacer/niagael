package dto

import (
	"time"

	"kaffein/dynamic-pricing-service/internal/domain"
)

type CreateFlashSaleDto struct {
	Id              string
	Name            string    `json:"name"`
	ProductId       string    `json:"product_id" validate:"required,uuid"`
	VariantId       *string   `json:"variant_id" validate:"omitempty,uuid"`
	DiscountPercent int       `json:"discount_percent" validate:"gte=1,lte=100"`
	Stock           int64     `json:"stock" validate:"gte=0"`
	MaxPerUser      int       `json:"max_per_user" validate:"gte=0"`
	StartTime       time.Time `json:"start_time" validate:"required"`
	EndTime         time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	IsActive        bool      `json:"is_active"`
}

type CreateFlashSaleItemDto struct {
	Id              string  `json:"-"`
	ProductId       string  `json:"product_id" validate:"required,uuid"`
	VariantId       *string `json:"variant_id" validate:"omitempty,uuid"`
	DiscountPercent int     `json:"discount_percent" validate:"gte=1,lte=100"`
	Stock           int64   `json:"stock" validate:"gte=0"`
	MaxPerUser      int     `json:"max_per_user" validate:"gte=0"`
}

type CreateFlashSaleBulkDto struct {
	Name      string                   `json:"name" validate:"required"`
	StartTime time.Time                `json:"start_time" validate:"required"`
	EndTime   time.Time                `json:"end_time" validate:"required,gtfield=StartTime"`
	IsActive  bool                     `json:"is_active"`
	Items     []CreateFlashSaleItemDto `json:"items" validate:"required,min=1,dive"`
}

type UpdateFlashSaleDto struct {
	Id              string
	Name            string    `json:"name"`
	ProductId       string    `json:"product_id" validate:"required,uuid"`
	VariantId       *string   `json:"variant_id" validate:"omitempty,uuid"`
	DiscountPercent int       `json:"discount_percent" validate:"gte=1,lte=100"`
	Stock           int64     `json:"stock" validate:"gte=0"`
	MaxPerUser      int       `json:"max_per_user" validate:"gte=0"`
	StartTime       time.Time `json:"start_time" validate:"required"`
	EndTime         time.Time `json:"end_time" validate:"required,gtfield=StartTime"`
	IsActive        bool      `json:"is_active"`
}

type ListFlashSaleParams struct {
	Search             string `form:"search" validate:"omitempty,max=100"`
	Name               string `form:"name" validate:"omitempty,max=255"`
	IsActive           *bool  `form:"is_active" validate:"omitempty"`
	ProductId          string `form:"product_id" validate:"omitempty,uuid"`
	MinDiscountPercent string `form:"min_discount_percent" validate:"omitempty,numeric"`
	MaxDiscountPercent string `form:"max_discount_percent" validate:"omitempty,numeric"`
	OrderBy            string `form:"order_by" validate:"def_enum=name start_time end_time created_at"`
	OrderDir           string `form:"order_dir" validate:"def_enum=desc asc"`
	PageSize           int32  `form:"page_size" validate:"clamp=10 100"`
	Page               int32  `form:"page" validate:"clamp=1"`
	PageOffset         int32  `form:"-"`
	CurrentOnly        bool   `form:"-"`
}

type PublicFlashSaleResponse struct {
	Id              string    `json:"id"`
	Name            string    `json:"name"`
	ProductId       string    `json:"product_id"`
	VariantId       *string   `json:"variant_id"`
	DiscountPercent int       `json:"discount_percent"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
}

func ToPublicFlashSaleResponse(fs domain.FlashSale) PublicFlashSaleResponse {
	return PublicFlashSaleResponse{
		Id: fs.Id, Name: fs.Name, ProductId: fs.ProductId, VariantId: fs.VariantId,
		DiscountPercent: fs.DiscountPercent, StartTime: fs.StartTime, EndTime: fs.EndTime,
	}
}

type VariantIdPair struct {
	ProductId string `json:"product_id"`
	VariantId string `json:"variant_id"`
}

type FlashSaleSession struct {
	Name        string    `json:"name"`
	ItemCount   int64     `json:"item_count"`
	TotalStock  int64     `json:"total_stock"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	IsActive    bool      `json:"is_active"`
}
