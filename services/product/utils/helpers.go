package utils

// DiscountedPrice returns price after applying a percentage discount.
// discountPercent is clamped between 0 and 100.
func DiscountedPrice(price int64, discountPercent int32) int64 {
	if discountPercent <= 0 {
		return price
	}
	if discountPercent > 100 {
		discountPercent = 100
	}
	return price - (price * int64(discountPercent) / 100)
}
