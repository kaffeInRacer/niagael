package dto

import (
	"kaffein/product-service/utils"

	"encoding/json"
	"testing"

	"kaffein/product-service/internal/domain"
)

func TestDiscountedPrice(t *testing.T) {
	if got := utils.DiscountedPrice(100_000, 20); got != 80_000 {
		t.Fatalf("utils.DiscountedPrice() = %d, want 80000", got)
	}
}

func TestToProductDetailResponseAppliesProductDiscountPercent(t *testing.T) {
	product := &domain.Product{Id: "product-1", Price: 100_000}
	flashSale := &FlashSaleInfo{ProductId: product.Id, DiscountPercent: 20}

	response := ToProductDetailResponse(product, "", flashSale, nil)

	if response.FinalPrice != 80_000 || response.DiscountPercent != 20 || !response.IsFlashSale {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestToProductDetailResponseAppliesVariantDiscountPercent(t *testing.T) {
	product := &domain.Product{
		Id: "product-1",
		Variants: []domain.ProductVariant{
			{Id: "variant-1", Price: 100_000, Attributes: json.RawMessage(`{"size":"S"}`)},
			{Id: "variant-2", Price: 120_000, Attributes: json.RawMessage(`{"size":"M"}`)},
		},
	}

	response := ToProductDetailResponse(product, "", nil, []VariantFlashSale{
		{VariantId: "variant-2", DiscountPercent: 25},
	})

	if response.Variants == nil || len(response.Variants.Items) != 2 {
		t.Fatal("expected two variant items")
	}
	if response.Variants.Items[0].FinalPrice != 100_000 || response.Variants.Items[0].DiscountPercent != 0 {
		t.Fatalf("unexpected regular variant: %+v", response.Variants.Items[0])
	}
	if response.Variants.Items[1].FinalPrice != 90_000 || response.Variants.Items[1].DiscountPercent != 25 {
		t.Fatalf("unexpected discounted variant: %+v", response.Variants.Items[1])
	}
}

func TestToProductListResponseSelectsLowestFinalVariant(t *testing.T) {
	product := &domain.Product{
		Id: "product-1",
		Variants: []domain.ProductVariant{
			{Id: "variant-1", Price: 80_000},
			{Id: "variant-2", Price: 100_000},
		},
	}

	response := ToProductListResponse(product, "", nil, []VariantFlashSale{
		{VariantId: "variant-2", DiscountPercent: 30},
	})

	if response.Price != 100_000 || response.FinalPrice != 70_000 {
		t.Fatalf("prices = %d/%d, want 100000/70000", response.Price, response.FinalPrice)
	}
	if response.DiscountPercent != 30 || response.Discount != 30_000 || !response.IsFlashSale {
		t.Fatalf("unexpected flash sale metadata: %+v", response)
	}
}
