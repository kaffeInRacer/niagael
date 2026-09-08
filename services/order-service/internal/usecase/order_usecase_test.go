package usecase

import (
	"testing"

	"kaffein/order-service/utils"
)

func TestDiscountedPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    int64
		percent  int32
		expected int64
	}{
		{name: "twenty percent", price: 100_000, percent: 20, expected: 80_000},
		{name: "no discount", price: 100_000, percent: 0, expected: 100_000},
		{name: "full discount", price: 100_000, percent: 100, expected: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := utils.DiscountedPrice(test.price, test.percent); got != test.expected {
				t.Fatalf("utils.DiscountedPrice() = %d, want %d", got, test.expected)
			}
		})
	}
}
