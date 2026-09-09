package dto

type CreateOrderDto struct {
	UserId    string               `json:"user_id" validate:"required"`
	AddressId string               `json:"address_id" validate:"required"`
	Items     []CreateOrderItemDto `json:"items" validate:"required,min=1"`
	PromoCode string               `json:"promo_code"`
}

type CreateOrderItemDto struct {
	ProductId string `json:"product_id" validate:"required"`
	VariantId string `json:"variant_id"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

type UpdateOrderStatusDto struct {
	Status string `json:"status" validate:"required,def_enum=processing cancelled refunded shipped delivered"`
}

type OrderItemSnapshot struct {
	ProductName              string
	ProductPrice             int64
	FlashSaleName            *string
	FlashSaleDiscountPercent *int
	FlashSaleDiscountPrice   *int64
	FlashSaleOriginalPrice   *int64
	FlashSaleQuantity        *int
	PromoCode                *string
	PromoName                *string
	PromoDiscountType        *string
	PromoDiscountAmount      *int64
}

type ListOrderParams struct {
	Search     string `form:"search" validate:"omitempty,max=100"`
	UserId     string `form:"user_id" validate:"omitempty"`
	Status     string `form:"status" validate:"omitempty,def_enum=pending paid processing shipped delivered cancelled refunded"`
	OrderBy    string `form:"order_by" validate:"def_enum=created_at total_amount"`
	OrderDir   string `form:"order_dir" validate:"def_enum=desc asc"`
	PageSize   int32  `form:"page_size" validate:"clamp=10 100"`
	Page       int32  `form:"page" validate:"clamp=1"`
	PageOffset int32  `form:"-"`
}
