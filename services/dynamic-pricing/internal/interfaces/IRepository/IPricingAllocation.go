package IRepository

import (
	"context"
)

type PricingAllocationItem struct {
	ItemID            string
	FlashSaleID       string
	Quantity          int
	UnitPrice         int64
	AllocatedQuantity int
	Name              string
	DiscountPercent   int
}

type PricingAllocation struct {
	OrderID             string
	UserID              string
	PromoCode           string
	Subtotal            int64
	PromoID             string
	PromoName           string
	PromoDiscountType   string
	PromoDiscountAmount int64
	RequestHash         string
	Items               []PricingAllocationItem
}

type PricingAllocationRepository interface {
	AllocateWithTx(context.Context, PricingAllocation) (*PricingAllocation, error)
	ReleaseWithTx(context.Context, string, string) error
}
