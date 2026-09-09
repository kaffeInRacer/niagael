package grpc

import (
	"context"
	"errors"
	"kaffein/order-service/proto/product"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	client productpb.ProductServiceClient
	conn   *grpc.ClientConn
}

func NewProductClient(addr string) (*ProductClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ProductClient{
		client: productpb.NewProductServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *ProductClient) Close() error {
	return c.conn.Close()
}

func (c *ProductClient) GetProduct(ctx context.Context, id string) (*productpb.Product, error) {
	return c.client.GetProduct(ctx, &productpb.GetProductRequest{Id: id})
}

func (c *ProductClient) GetProducts(ctx context.Context, ids []string) ([]*productpb.Product, error) {
	resp, err := c.client.GetProducts(ctx, &productpb.GetProductsRequest{Ids: ids})
	if err != nil {
		return nil, err
	}
	return resp.Products, nil
}

func (c *ProductClient) ReserveStock(ctx context.Context, orderId string, items []*productpb.StockItem) (*productpb.ReserveStockResponse, error) {
	resp, err := c.client.ReserveStock(ctx, &productpb.ReserveStockRequest{
		OrderId: orderId,
		Items:   items,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.Success {
		if resp != nil && resp.Message != "" {
			return resp, errors.New(resp.Message)
		}
		return resp, errors.New("stock reservation failed")
	}
	return resp, nil
}

func (c *ProductClient) ReleaseStock(ctx context.Context, orderId string, items []*productpb.StockItem) (*productpb.ReleaseStockResponse, error) {
	resp, err := c.client.ReleaseStock(ctx, &productpb.ReleaseStockRequest{
		OrderId: orderId,
		Items:   items,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.Success {
		if resp != nil && resp.Message != "" {
			return resp, errors.New(resp.Message)
		}
		return resp, errors.New("stock release failed")
	}
	return resp, nil
}

func (c *ProductClient) ConfirmStock(ctx context.Context, orderId string, items []*productpb.StockItem) (*productpb.ConfirmStockResponse, error) {
	resp, err := c.client.ConfirmStock(ctx, &productpb.ConfirmStockRequest{
		OrderId: orderId,
		Items:   items,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.Success {
		if resp != nil && resp.Message != "" {
			return resp, errors.New(resp.Message)
		}
		return resp, errors.New("stock confirmation failed")
	}
	return resp, nil
}
