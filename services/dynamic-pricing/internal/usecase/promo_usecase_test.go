package usecase

import (
	"context"
	"testing"
	"time"

	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
)

type promoRepoStub struct {
	promo              *domain.Promo
	incrementCallCount int
}

func (r *promoRepoStub) List(context.Context, dto.ListPromoParams) ([]domain.Promo, error) {
	return nil, nil
}
func (r *promoRepoStub) ListCount(context.Context, dto.ListPromoParams) (int64, error) {
	return 0, nil
}
func (r *promoRepoStub) Create(context.Context, dto.CreatePromoDto) error { return nil }
func (r *promoRepoStub) Update(context.Context, dto.UpdatePromoDto) error { return nil }
func (r *promoRepoStub) Delete(context.Context, string) error             { return nil }
func (r *promoRepoStub) ReadById(context.Context, string) (*domain.Promo, error) {
	return r.promo, nil
}
func (r *promoRepoStub) ReadByCode(context.Context, string) (*domain.Promo, error) {
	return r.promo, nil
}
func (r *promoRepoStub) IncrementUsage(context.Context, string) error {
	r.incrementCallCount++
	return nil
}

type promoUsageRepoStub struct {
	incrementCallCount int
}

func (r *promoUsageRepoStub) GetUsage(context.Context, string, string) (*domain.PromoUsage, error) {
	return nil, nil
}
func (r *promoUsageRepoStub) IncrementUsage(context.Context, string, string, int) error {
	r.incrementCallCount++
	return nil
}

func TestApplyPromoQuotesWithoutConsumingUsage(t *testing.T) {
	promoRepo := &promoRepoStub{promo: &domain.Promo{
		Id:            "promo-1",
		DiscountType:  "percentage",
		DiscountValue: 10,
		StartDate:     time.Now().Add(-time.Hour),
		EndDate:       time.Now().Add(time.Hour),
		IsActive:      true,
	}}
	usageRepo := &promoUsageRepoStub{}
	uc := NewPromoUseCase(promoRepo, usageRepo)

	discount, err := uc.ApplyPromo(context.Background(), "SAVE10", "user-1", 100_000)
	if err != nil {
		t.Fatalf("ApplyPromo() error = %v", err)
	}
	if discount != 10_000 {
		t.Fatalf("discount = %d, want 10000", discount)
	}
	if promoRepo.incrementCallCount != 0 || usageRepo.incrementCallCount != 0 {
		t.Fatal("quoting a promo must not consume usage")
	}
}
