package grpc

import (
	"context"
	"kaffein/product-service/proto/dynamic-pricing"

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

func (c *DynamicPricingClient) GetFlashSalesByProductIds(ctx context.Context, productIds []string) (*dynamicpricingpb.GetFlashSalesByProductIdsResponse, error) {
	return c.client.GetFlashSalesByProductIds(ctx, &dynamicpricingpb.GetFlashSalesByProductIdsRequest{
		ProductIds: productIds,
	})
}

func (c *DynamicPricingClient) GetFlashSalesByVariantIds(ctx context.Context, items []*dynamicpricingpb.FlashSaleVariantRequest) (*dynamicpricingpb.GetFlashSalesByVariantIdsResponse, error) {
	return c.client.GetFlashSalesByVariantIds(ctx, &dynamicpricingpb.GetFlashSalesByVariantIdsRequest{
		Items: items,
	})
}
