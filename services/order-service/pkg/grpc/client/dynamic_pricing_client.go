package grpc

import (
	"context"
	"kaffein/order-service/proto/dynamic-pricing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DynamicPricingClient struct {
	client dynamicpricingpb.DynamicPricingServiceClient
	conn   *grpc.ClientConn
}

func NewDynamicPricingClient(addr string) (*DynamicPricingClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &DynamicPricingClient{
		client: dynamicpricingpb.NewDynamicPricingServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *DynamicPricingClient) Close() error {
	return c.conn.Close()
}

func (c *DynamicPricingClient) GetFlashSaleByProductId(ctx context.Context, productId string) (*dynamicpricingpb.FlashSale, error) {
	return c.client.GetFlashSaleByProductId(ctx, &dynamicpricingpb.GetFlashSaleByProductIdRequest{ProductId: productId})
}

func (c *DynamicPricingClient) GetFlashSaleByVariantId(ctx context.Context, productId string, variantId string) (*dynamicpricingpb.FlashSale, error) {
	return c.client.GetFlashSaleByVariantId(ctx, &dynamicpricingpb.GetFlashSaleByVariantIdRequest{
		ProductId: productId,
		VariantId: variantId,
	})
}

func (c *DynamicPricingClient) GetPromoByCode(ctx context.Context, code string) (*dynamicpricingpb.Promo, error) {
	return c.client.GetPromoByCode(ctx, &dynamicpricingpb.GetPromoByCodeRequest{Code: code})
}

func (c *DynamicPricingClient) ApplyPromo(ctx context.Context, code string, userId string, totalAmount int64) (*dynamicpricingpb.ApplyPromoResponse, error) {
	return c.client.ApplyPromo(ctx, &dynamicpricingpb.ApplyPromoRequest{
		Code:        code,
		UserId:      userId,
		TotalAmount: totalAmount,
	})
}

func (c *DynamicPricingClient) DecrementFlashSaleStock(ctx context.Context, flashSaleId string, quantity int32) (*dynamicpricingpb.DecrementFlashSaleStockResponse, error) {
	return c.client.DecrementFlashSaleStock(ctx, &dynamicpricingpb.DecrementFlashSaleStockRequest{
		FlashSaleId: flashSaleId,
		Quantity:    quantity,
	})
}

func (c *DynamicPricingClient) IncrementFlashSaleUsage(ctx context.Context, flashSaleId string, userId string, quantity int32) (*dynamicpricingpb.IncrementFlashSaleUsageResponse, error) {
	return c.client.IncrementFlashSaleUsage(ctx, &dynamicpricingpb.IncrementFlashSaleUsageRequest{
		FlashSaleId: flashSaleId,
		UserId:      userId,
		Quantity:    quantity,
	})
}

func (c *DynamicPricingClient) GetPromoUsage(ctx context.Context, promoId string, userId string) (*dynamicpricingpb.PromoUsageResponse, error) {
	return c.client.GetPromoUsage(ctx, &dynamicpricingpb.GetPromoUsageRequest{
		PromoId: promoId,
		UserId:  userId,
	})
}
