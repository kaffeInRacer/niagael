package utils

import (
	"crypto/rand"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func DiscountedPrice(price int64, discountPercent int32) int64 {
	if discountPercent <= 0 {
		return price
	}
	if discountPercent > 100 {
		discountPercent = 100
	}
	return price - (price * int64(discountPercent) / 100)
}

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

func ParseMidtransGrossAmount(value string) (int64, error) {
	integer, fraction, found := strings.Cut(value, ".")
	if integer == "" || strings.HasPrefix(integer, "+") || strings.HasPrefix(integer, "-") {
		return 0, errors.New("invalid payment amount")
	}
	if found && (fraction == "" || strings.Trim(fraction, "0") != "") {
		return 0, errors.New("invalid payment amount")
	}
	amount, err := strconv.ParseInt(integer, 10, 64)
	if err != nil {
		return 0, errors.New("invalid payment amount")
	}
	return amount, nil
}
