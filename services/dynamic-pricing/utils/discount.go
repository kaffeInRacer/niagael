package utils

// DiscountedUnitPrice applies a percentage discount to a unit price.
func DiscountedUnitPrice(unitPrice int64, discountPercent int) int64 {
	if discountPercent <= 0 {
		return unitPrice
	}
	if discountPercent > 100 {
		discountPercent = 100
	}
	return unitPrice - (unitPrice * int64(discountPercent) / 100)
}
