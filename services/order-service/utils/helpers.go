package utils

import (
	"crypto/rand"

	"github.com/google/uuid"
)

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

// GenerateOrderRef returns a human friendly order reference in the
// form ORD-ABCDE-12345. Letters exclude ambiguous characters (I, O, Q).
func GenerateOrderRef() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	const nums = "0123456789"

	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		b = []byte(uuid.New().String())
	}

	middle := make([]byte, 5)
	for i := range middle {
		middle[i] = alphabet[int(b[i])%len(alphabet)]
	}

	tail := make([]byte, 5)
	for i := range tail {
		tail[i] = nums[int(b[5+i])%len(nums)]
	}

	return "ORD-" + string(middle) + "-" + string(tail)
}
