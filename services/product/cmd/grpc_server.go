package main

import (
	"context"
	"encoding/json"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/interfaces/IUseCase"
	"kaffein/product-service/proto/product"
	"net"
	"time"

	"google.golang.org/grpc"
)

type grpcServer struct {
	productpb.UnimplementedProductServiceServer
	productUseCase IUseCase.ProductUseCase
}

func NewGrpcServer(productUseCase IUseCase.ProductUseCase) *grpcServer {
	return &grpcServer{
		productUseCase: productUseCase,
	}
}

func (app *application) startGrpcServer(ctx context.Context, productUseCase IUseCase.ProductUseCase) {
	lis, err := net.Listen("tcp", app.config.GRPC.Server)
	if err != nil {
		app.logger.Fatal().Err(err).Msg("failed to listen for gRPC")
	}

	s := grpc.NewServer()
	productpb.RegisterProductServiceServer(s, NewGrpcServer(productUseCase))

	go func() {
		app.logger.Info().Str("addr", app.config.GRPC.Server).Msg("starting gRPC server")

		if err := s.Serve(lis); err != nil {
			app.logger.Fatal().Err(err).Msg("failed to serve gRPC")
		}
	}()

	go func() {
		<-ctx.Done()

		app.logger.Info().Msg("shutting down gRPC server")
		s.GracefulStop()
		app.logger.Info().Msg("gRPC server stopped")
	}()
}

func (s *grpcServer) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {
	products, err := s.productUseCase.BatchById(ctx, []string{req.Id})
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, nil
	}

	product := products[0]
	return productToProto(product), nil
}

func productToProto(product domain.Product) *productpb.Product {
	variants := make([]*productpb.ProductVariant, 0, len(product.Variants))
	for _, variant := range product.Variants {
		if !variant.IsActive {
			continue
		}
		attributes := make(map[string]string)
		_ = json.Unmarshal(variant.Attributes, &attributes)
		variants = append(variants, &productpb.ProductVariant{
			Id:            variant.Id,
			ProductId:     variant.ProductId,
			Name:          variant.Name,
			Price:         variant.Price,
			Stock:         variant.Stock,
			StockReserved: variant.StockReserved,
			IsActive:      variant.IsActive,
			Attributes:    attributes,
		})
	}

	return &productpb.Product{
		Id:              product.Id,
		CategoryId:      product.CategoryId,
		CategoryName:    product.CategoryName,
		Name:            product.Name,
		Slug:            product.Slug,
		Description:     product.Description,
		Price:           product.Price,
		Stock:           product.Stock,
		StockReserved:   product.StockReserved,
		IsActive:        product.IsActive,
		IsPromoExcluded: product.IsPromo,
		CreatedAt:       product.CreatedAt.Format(time.RFC3339),
		Variants:        variants,
	}
}

func (s *grpcServer) GetProducts(ctx context.Context, req *productpb.GetProductsRequest) (*productpb.ProductsResponse, error) {
	products, err := s.productUseCase.BatchById(ctx, req.Ids)
	if err != nil {
		return nil, err
	}

	var pbProducts []*productpb.Product
	for _, p := range products {
		pbProducts = append(pbProducts, productToProto(p))
	}

	return &productpb.ProductsResponse{
		Products: pbProducts,
	}, nil
}

func (s *grpcServer) GetProductsByCategory(ctx context.Context, req *productpb.GetProductsByCategoryRequest) (*productpb.ProductsResponse, error) {
	// TODO: Implement GetProductsByCategory
	return &productpb.ProductsResponse{}, nil
}

func (s *grpcServer) ReserveStock(ctx context.Context, req *productpb.ReserveStockRequest) (*productpb.ReserveStockResponse, error) {
	for _, item := range req.Items {
		err := s.productUseCase.ReserveStock(ctx, item.ProductId, item.VariantId, item.Quantity)
		if err != nil {
			return &productpb.ReserveStockResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
	}

	return &productpb.ReserveStockResponse{
		Success: true,
		Message: "stock reserved successfully",
	}, nil
}

func (s *grpcServer) ReleaseStock(ctx context.Context, req *productpb.ReleaseStockRequest) (*productpb.ReleaseStockResponse, error) {
	for _, item := range req.Items {
		err := s.productUseCase.ReleaseStock(ctx, item.ProductId, item.VariantId, item.Quantity)
		if err != nil {
			return &productpb.ReleaseStockResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
	}

	return &productpb.ReleaseStockResponse{
		Success: true,
		Message: "stock released successfully",
	}, nil
}

func (s *grpcServer) ConfirmStock(ctx context.Context, req *productpb.ConfirmStockRequest) (*productpb.ConfirmStockResponse, error) {
	for _, item := range req.Items {
		err := s.productUseCase.ConfirmStock(ctx, item.ProductId, item.VariantId, item.Quantity)
		if err != nil {
			return &productpb.ConfirmStockResponse{
				Success: false,
				Message: err.Error(),
			}, nil
		}
	}

	return &productpb.ConfirmStockResponse{
		Success: true,
		Message: "stock confirmed successfully",
	}, nil
}
