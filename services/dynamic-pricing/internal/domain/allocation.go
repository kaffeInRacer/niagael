package domain

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"time"
)

type PromoAllocationInput struct {
	Active              bool
	StartDate           time.Time
	EndDate             time.Time
	Quantity            int
	UsedCount           int
	MaxPerUser          int
	UserUsed            int
	CanCombineFlashSale bool
	MinPurchase         int64
	HasFlashSale        bool
	Subtotal            int64
	DiscountType        string
	DiscountValue       int64
	MaxDiscount         int64
}

func AllocatableQuantity(requested int, stock int64, maxPerUser int, used int) int {
	quantity := requested
	if int64(quantity) > stock {
		quantity = int(stock)
	}
	if maxPerUser > 0 && quantity > maxPerUser-used {
		quantity = maxPerUser - used
	}
	if quantity < 0 {
		return 0
	}
	return quantity
}

func ValidatePromoAllocation(input PromoAllocationInput, now time.Time) (int64, error) {
	switch {
	case !input.Active:
		return 0, fmt.Errorf("promo is not active")
	case now.Before(input.StartDate) || now.After(input.EndDate):
		return 0, fmt.Errorf("promo has expired")
	case input.Quantity > 0 && input.UsedCount >= input.Quantity:
		return 0, fmt.Errorf("promo usage limit reached")
	case input.MaxPerUser > 0 && input.UserUsed >= input.MaxPerUser:
		return 0, fmt.Errorf("promo max usage per user reached")
	case input.HasFlashSale && !input.CanCombineFlashSale:
		return 0, fmt.Errorf("promo cannot be combined with flash sale")
	case input.Subtotal < input.MinPurchase:
		return 0, fmt.Errorf("minimum purchase not reached")
	}

	var discount int64
	switch input.DiscountType {
	case "percentage":
		discount = input.Subtotal * input.DiscountValue / 100
		if input.MaxDiscount > 0 && discount > input.MaxDiscount {
			discount = input.MaxDiscount
		}
	case "fixed":
		discount = input.DiscountValue
		if discount > input.Subtotal {
			discount = input.Subtotal
		}
	default:
		return 0, fmt.Errorf("unsupported promo discount type %q", input.DiscountType)
	}
	return discount, nil
}

type AllocationFingerprintItem struct {
	ItemID      string
	FlashSaleID string
	Quantity    int
	UnitPrice   int64
}

func AllocationFingerprint(userID, promoCode string, subtotal int64, items []AllocationFingerprintItem) string {
	sorted := append([]AllocationFingerprintItem(nil), items...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ItemID < sorted[j].ItemID })
	hash := sha256.New()
	fmt.Fprintf(hash, "%s\x00%s\x00%d\x00", userID, promoCode, subtotal)
	for _, item := range sorted {
		fmt.Fprintf(hash, "%s\x00%s\x00%d\x00%d\x00", item.ItemID, item.FlashSaleID, item.Quantity, item.UnitPrice)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
