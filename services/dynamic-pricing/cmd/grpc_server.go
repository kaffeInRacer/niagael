package main

import (
	"context"
	"math"
	"net"
	"time"

	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IRepository"
	"kaffein/dynamic-pricing-service/internal/interfaces/IUseCase"
	"kaffein/dynamic-pricing-service/internal/repository"
	"kaffein/dynamic-pricing-service/internal/usecase"
	"kaffein/dynamic-pricing-service/proto/dynamic-pricing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	dynamicpricingpb.UnimplementedDynamicPricingServiceServer
	flashSaleUseCase IUseCase.FlashSaleUseCase
	promoUseCase     IUseCase.PromoUseCase
	allocationRepo   IRepository.PricingAllocationRepository
}

func NewGrpcServer(flashSaleUseCase IUseCase.FlashSaleUseCase, promoUseCase IUseCase.PromoUseCase, allocationRepo IRepository.PricingAllocationRepository) *grpcServer {
	return &grpcServer{
		flashSaleUseCase: flashSaleUseCase,
		promoUseCase:     promoUseCase,
		allocationRepo:   allocationRepo,
	}
}

func (app *application) startGrpcServer(ctx context.Context, flashSaleRepo IRepository.FlashSaleRepository, flashSaleUsageRepo IRepository.FlashSaleUsageRepository, promoRepo IRepository.PromoRepository, promoUsageRepo IRepository.PromoUsageRepository, allocationRepo IRepository.PricingAllocationRepository) {
	lis, err := net.Listen("tcp", app.config.GRPC.Server)
	if err != nil {
		app.logger.Fatal().Err(err).Msg("failed to listen for gRPC")
	}

	flashSaleUseCase := usecase.NewFlashSaleUseCase(flashSaleRepo, flashSaleUsageRepo)
	promoUseCase := usecase.NewPromoUseCase(promoRepo, promoUsageRepo)

	s := grpc.NewServer()
	dynamicpricingpb.RegisterDynamicPricingServiceServer(s, NewGrpcServer(flashSaleUseCase, promoUseCase, allocationRepo))

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

func (s *grpcServer) GetFlashSaleByProductId(ctx context.Context, req *dynamicpricingpb.GetFlashSaleByProductIdRequest) (*dynamicpricingpb.FlashSale, error) {
	fs, err := s.flashSaleUseCase.ReadByProductId(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	if fs == nil {
		return &dynamicpricingpb.FlashSale{}, nil
	}

	return &dynamicpricingpb.FlashSale{
		Id:              fs.Id,
		Name:            fs.Name,
		ProductId:       fs.ProductId,
		DiscountPercent: int32(fs.DiscountPercent),
		Stock:           fs.Stock,
		MaxPerUser:      int32(fs.MaxPerUser),
		StartTime:       fs.StartTime.Format(time.RFC3339),
		EndTime:         fs.EndTime.Format(time.RFC3339),
		IsActive:        fs.IsActive,
	}, nil
}

func (s *grpcServer) GetFlashSaleByVariantId(ctx context.Context, req *dynamicpricingpb.GetFlashSaleByVariantIdRequest) (*dynamicpricingpb.FlashSale, error) {
	fs, err := s.flashSaleUseCase.ReadByVariantId(ctx, req.ProductId, req.VariantId)
	if err != nil {
		return nil, err
	}

	if fs == nil {
		return &dynamicpricingpb.FlashSale{}, nil
	}

	var variantId string
	if fs.VariantId != nil {
		variantId = *fs.VariantId
	}

	return &dynamicpricingpb.FlashSale{
		Id:              fs.Id,
		Name:            fs.Name,
		ProductId:       fs.ProductId,
		VariantId:       variantId,
		DiscountPercent: int32(fs.DiscountPercent),
		Stock:           fs.Stock,
		MaxPerUser:      int32(fs.MaxPerUser),
		StartTime:       fs.StartTime.Format(time.RFC3339),
		EndTime:         fs.EndTime.Format(time.RFC3339),
		IsActive:        fs.IsActive,
	}, nil
}

func (s *grpcServer) GetFlashSalesByProductIds(ctx context.Context, req *dynamicpricingpb.GetFlashSalesByProductIdsRequest) (*dynamicpricingpb.GetFlashSalesByProductIdsResponse, error) {
	flashSales, err := s.flashSaleUseCase.ReadByProductIds(ctx, req.ProductIds)
	if err != nil {
		return nil, err
	}

	var pbFlashSales []*dynamicpricingpb.FlashSale
	for _, fs := range flashSales {
		pbFlashSales = append(pbFlashSales, &dynamicpricingpb.FlashSale{
			Id:              fs.Id,
			Name:            fs.Name,
			ProductId:       fs.ProductId,
			DiscountPercent: int32(fs.DiscountPercent),
			Stock:           fs.Stock,
			MaxPerUser:      int32(fs.MaxPerUser),
			StartTime:       fs.StartTime.Format(time.RFC3339),
			EndTime:         fs.EndTime.Format(time.RFC3339),
			IsActive:        fs.IsActive,
		})
	}

	return &dynamicpricingpb.GetFlashSalesByProductIdsResponse{
		FlashSales: pbFlashSales,
	}, nil
}

func (s *grpcServer) GetFlashSalesByVariantIds(ctx context.Context, req *dynamicpricingpb.GetFlashSalesByVariantIdsRequest) (*dynamicpricingpb.GetFlashSalesByVariantIdsResponse, error) {
	var items []dto.VariantIdPair
	for _, item := range req.Items {
		items = append(items, dto.VariantIdPair{
			ProductId: item.ProductId,
			VariantId: item.VariantId,
		})
	}

	flashSales, err := s.flashSaleUseCase.ReadByVariantIds(ctx, items)
	if err != nil {
		return nil, err
	}

	var pbFlashSales []*dynamicpricingpb.FlashSale
	for _, fs := range flashSales {
		var variantId string
		if fs.VariantId != nil {
			variantId = *fs.VariantId
		}
		pbFlashSales = append(pbFlashSales, &dynamicpricingpb.FlashSale{
			Id:              fs.Id,
			Name:            fs.Name,
			ProductId:       fs.ProductId,
			VariantId:       variantId,
			DiscountPercent: int32(fs.DiscountPercent),
			Stock:           fs.Stock,
			MaxPerUser:      int32(fs.MaxPerUser),
			StartTime:       fs.StartTime.Format(time.RFC3339),
			EndTime:         fs.EndTime.Format(time.RFC3339),
			IsActive:        fs.IsActive,
		})
	}

	return &dynamicpricingpb.GetFlashSalesByVariantIdsResponse{
		FlashSales: pbFlashSales,
	}, nil
}

func (s *grpcServer) GetPromoByCode(ctx context.Context, req *dynamicpricingpb.GetPromoByCodeRequest) (*dynamicpricingpb.Promo, error) {
	promo, err := s.promoUseCase.ReadByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return &dynamicpricingpb.Promo{}, nil
	}

	return &dynamicpricingpb.Promo{
		Id:                  promo.Id,
		Name:                promo.Name,
		Code:                promo.Code,
		Description:         promo.Description,
		DiscountType:        promo.DiscountType,
		DiscountValue:       promo.DiscountValue,
		MinPurchase:         promo.MinPurchase,
		MaxDiscount:         promo.MaxDiscount,
		Quantity:            int32(promo.Quantity),
		UsedCount:           int32(promo.UsedCount),
		MaxUsagePerUser:     int32(promo.MaxUsagePerUser),
		CanCombineFlashSale: promo.CanCombineFlashSale,
		StartDate:           promo.StartDate.Format(time.RFC3339),
		EndDate:             promo.EndDate.Format(time.RFC3339),
		IsActive:            promo.IsActive,
	}, nil
}

func (s *grpcServer) ApplyPromo(ctx context.Context, req *dynamicpricingpb.ApplyPromoRequest) (*dynamicpricingpb.ApplyPromoResponse, error) {
	discount, err := s.promoUseCase.ApplyPromo(ctx, req.Code, req.UserId, req.TotalAmount)
	if err != nil {
		return &dynamicpricingpb.ApplyPromoResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &dynamicpricingpb.ApplyPromoResponse{
		Discount: discount,
		Success:  true,
		Message:  "success",
	}, nil
}

func (s *grpcServer) GetPromoUsage(ctx context.Context, req *dynamicpricingpb.GetPromoUsageRequest) (*dynamicpricingpb.PromoUsageResponse, error) {
	usage, err := s.promoUseCase.GetUsage(ctx, req.PromoId, req.UserId)
	if err != nil {
		return nil, err
	}

	if usage == nil {
		return &dynamicpricingpb.PromoUsageResponse{}, nil
	}

	return &dynamicpricingpb.PromoUsageResponse{
		Id:       usage.Id,
		PromoId:  usage.PromoId,
		UserId:   usage.UserId,
		Quantity: int32(usage.Quantity),
	}, nil
}

func (s *grpcServer) AllocatePricing(ctx context.Context, req *dynamicpricingpb.AllocatePricingRequest) (*dynamicpricingpb.AllocatePricingResponse, error) {
	if req.OrderId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id and user_id are required")
	}

	allocation := repository.PricingAllocation{
		OrderID:   req.OrderId,
		UserID:    req.UserId,
		PromoCode: req.PromoCode,
		Subtotal:  req.Subtotal,
		Items:     make([]repository.PricingAllocationItem, len(req.Items)),
	}

	var calculatedSubtotal int64
	for i, item := range req.Items {

		if item == nil || item.ItemId == "" || item.Quantity <= 0 || item.UnitPrice < 0 {
			return nil, status.Error(codes.InvalidArgument, "valid item_id, quantity, and unit_price are required")
		}

		if item.UnitPrice > math.MaxInt64/int64(item.Quantity) {
			return nil, status.Error(codes.InvalidArgument, "pricing item total exceeds int64")
		}

		itemSubtotal := item.UnitPrice * int64(item.Quantity)
		if calculatedSubtotal > math.MaxInt64-itemSubtotal {
			return nil, status.Error(codes.InvalidArgument, "subtotal exceeds int64")
		}

		allocation.Items[i] = repository.PricingAllocationItem{
			ItemID:      item.ItemId,
			FlashSaleID: item.FlashSaleId,
			Quantity:    int(item.Quantity),
			UnitPrice:   item.UnitPrice,
		}
		calculatedSubtotal += itemSubtotal
	}

	if calculatedSubtotal != req.Subtotal {
		return nil, status.Error(codes.InvalidArgument, "subtotal does not match pricing items")
	}

	allocated, err := s.allocationRepo.AllocateWithTx(ctx, allocation)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	response := &dynamicpricingpb.AllocatePricingResponse{
		PromoId:             allocated.PromoID,
		PromoCode:           allocated.PromoCode,
		PromoName:           allocated.PromoName,
		PromoDiscountType:   allocated.PromoDiscountType,
		PromoDiscountAmount: allocated.PromoDiscountAmount,
	}
	for _, item := range allocated.Items {
		if item.AllocatedQuantity == 0 {
			continue
		}
		response.Items = append(response.Items, &dynamicpricingpb.PricingAllocationItem{
			ItemId:          item.ItemID,
			FlashSaleId:     item.FlashSaleID,
			Quantity:        int32(item.AllocatedQuantity),
			Name:            item.Name,
			DiscountPercent: int32(item.DiscountPercent),
		})
	}
	return response, nil
}

func (s *grpcServer) ReleasePricing(ctx context.Context, req *dynamicpricingpb.ReleasePricingRequest) (*dynamicpricingpb.ReleasePricingResponse, error) {
	if req.OrderId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id and user_id are required")
	}

	if err := s.allocationRepo.ReleaseWithTx(ctx, req.OrderId, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dynamicpricingpb.ReleasePricingResponse{Success: true}, nil
}
