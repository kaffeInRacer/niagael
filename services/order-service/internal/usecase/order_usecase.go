package usecase

import (
	"context"
	"errors"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/internal/interfaces/IUseCase"
	grpcclient "kaffein/order-service/pkg/grpc/client"
	"kaffein/order-service/proto/dynamic-pricing"
	"kaffein/order-service/proto/product"
	"kaffein/order-service/utils"
	"kaffein/order-service/utils/constants"
	"time"

	"github.com/google/uuid"
)

type orderUseCase struct {
	repo          IRepository.OrderRepository
	productClient *grpcclient.ProductClient
	pricingClient *grpcclient.DynamicPricingClient
}

func NewOrderUseCase(
	repo IRepository.OrderRepository,
	productClient *grpcclient.ProductClient,
	pricingClient *grpcclient.DynamicPricingClient,
) IUseCase.OrderUseCase {
	return &orderUseCase{
		repo:          repo,
		productClient: productClient,
		pricingClient: pricingClient,
	}
}

func (uc *orderUseCase) Create(ctx context.Context, args dto.CreateOrderDto) (*domain.Order, error) {
	orderId := uuid.New().String()
	orderRef := utils.GenerateOrderRef()

	// Collect all product IDs for batch fetch
	var productIds []string
	for _, item := range args.Items {
		productIds = append(productIds, item.ProductId)
	}

	// Fetch all products in single query (avoid N+1)
	products, err := uc.productClient.GetProducts(ctx, productIds)
	if err != nil {
		return nil, err
	}

	// Create product map for quick lookup
	productMap := make(map[string]*productpb.Product)
	for _, p := range products {
		productMap[p.Id] = p
	}

	// Validate all products exist
	for _, item := range args.Items {
		if _, ok := productMap[item.ProductId]; !ok {
			return nil, errors.New(constants.ErrProductNotFound)
		}
	}

	// Fetch promo if provided (for snapshot and validation)
	var promo *dynamicpricingpb.Promo
	if args.PromoCode != "" {
		promo, err = uc.pricingClient.GetPromoByCode(ctx, args.PromoCode)
		if err != nil {
			return nil, err
		}
		if promo == nil {
			return nil, errors.New(constants.ErrPromoNotFound)
		}
	}

	var totalAmount int64
	var orderItems []domain.OrderItem
	var stockItems []*productpb.StockItem
	hasFlashSale := false

	for _, item := range args.Items {
		product := productMap[item.ProductId]

		// Find variant (per variant logic)
		var variant *productpb.ProductVariant
		if item.VariantId != "" {
			for _, v := range product.Variants {
				if v.Id == item.VariantId {
					variant = v
					break
				}
			}
			if variant == nil {
				return nil, errors.New(constants.ErrVariantNotFound)
			}
		}

		// Determine price and stock (variant > product)
		var unitPrice int64
		var availableStock int64
		if variant != nil {
			unitPrice = variant.Price
			availableStock = variant.Stock - variant.StockReserved
		} else {
			unitPrice = product.Price
			availableStock = product.Stock - product.StockReserved
		}

		// Check available stock (stock minus reserved)
		if availableStock < int64(item.Quantity) {
			return nil, errors.New(constants.ErrInsufficientStock)
		}

		// Prepare stock item for reservation
		stockItem := &productpb.StockItem{
			ProductId: item.ProductId,
			VariantId: item.VariantId,
			Quantity:  int32(item.Quantity),
		}
		stockItems = append(stockItems, stockItem)

		// Auto check flash sale for this variant
		var flashSaleId, flashSaleName *string
		var flashSaleDiscountPrice, flashSaleOriginalPrice *int64
		var flashSaleDiscountPercent *int
		var flashSaleQuantity *int

		var flashSale *dynamicpricingpb.FlashSale
		if variant != nil {
			// Check flash sale by variant
			flashSale, err = uc.pricingClient.GetFlashSaleByVariantId(ctx, item.ProductId, item.VariantId)
		} else {
			// Check flash sale by product
			flashSale, err = uc.pricingClient.GetFlashSaleByProductId(ctx, item.ProductId)
		}

		if err == nil && flashSale != nil && flashSale.IsActive && flashSale.DiscountPercent > 0 {
			hasFlashSale = true

			// Flash sale is active, calculate allocation
			fsId := flashSale.Id
			fsName := flashSale.Name
			fsDiscountPercent := int(flashSale.DiscountPercent)
			fsDiscountPrice := utils.DiscountedPrice(unitPrice, flashSale.DiscountPercent)
			fsOriginalPrice := unitPrice

			// Calculate how many items get flash sale price
			fsQuantity := item.Quantity
			if flashSale.MaxPerUser > 0 && fsQuantity > int(flashSale.MaxPerUser) {
				fsQuantity = int(flashSale.MaxPerUser)
			}
			if int64(fsQuantity) > flashSale.Stock {
				fsQuantity = int(flashSale.Stock)
			}

			flashSaleId = &fsId
			flashSaleName = &fsName
			flashSaleDiscountPercent = &fsDiscountPercent
			flashSaleDiscountPrice = &fsDiscountPrice
			flashSaleOriginalPrice = &fsOriginalPrice
			flashSaleQuantity = &fsQuantity
		}

		// Calculate final price
		var finalPrice int64
		if flashSaleDiscountPrice != nil && flashSaleQuantity != nil {
			// Items at flash sale price
			flashSaleTotal := *flashSaleDiscountPrice * int64(*flashSaleQuantity)
			// Remaining items at normal price
			normalQuantity := item.Quantity - *flashSaleQuantity
			normalTotal := unitPrice * int64(normalQuantity)
			finalPrice = flashSaleTotal + normalTotal
		} else {
			// No flash sale, all items at normal price
			finalPrice = unitPrice * int64(item.Quantity)
		}

		itemId := uuid.New().String()
		var variantId *string
		if item.VariantId != "" {
			id := item.VariantId
			variantId = &id
		}
		orderItem := domain.OrderItem{
			Id:                       itemId,
			OrderId:                  orderId,
			ProductId:                item.ProductId,
			VariantId:                variantId,
			ProductName:              product.Name,
			ProductPrice:             unitPrice,
			Quantity:                 item.Quantity,
			FlashSaleId:              flashSaleId,
			FlashSaleName:            flashSaleName,
			FlashSaleDiscountPercent: flashSaleDiscountPercent,
			FlashSaleDiscountPrice:   flashSaleDiscountPrice,
			FlashSaleOriginalPrice:   flashSaleOriginalPrice,
			FlashSaleQuantity:        flashSaleQuantity,
			FinalPrice:               finalPrice,
		}

		orderItems = append(orderItems, orderItem)
		totalAmount += finalPrice
	}

	// Validate and apply promo
	var promoSnapshot *PromoSnapshot
	if promo != nil {
		// Check if promo can combine with flash sale
		if hasFlashSale && !promo.CanCombineFlashSale {
			return nil, errors.New(constants.ErrPromoCannotCombineFlashSale)
		}

		// Validate promo
		if !promo.IsActive {
			return nil, errors.New(constants.ErrPromoNotActive)
		}

		now := time.Now()
		startDate, _ := time.Parse(time.RFC3339, promo.StartDate)
		endDate, _ := time.Parse(time.RFC3339, promo.EndDate)
		if now.Before(startDate) || now.After(endDate) {
			return nil, errors.New(constants.ErrPromoExpired)
		}

		if promo.Quantity > 0 && promo.UsedCount >= promo.Quantity {
			return nil, errors.New(constants.ErrPromoLimitReached)
		}

		// Check per-user usage limit
		if promo.MaxUsagePerUser > 0 {
			usage, err := uc.pricingClient.GetPromoUsage(ctx, promo.Id, args.UserId)
			if err == nil && usage != nil && usage.Quantity >= int32(promo.MaxUsagePerUser) {
				return nil, errors.New(constants.ErrPromoLimitReached)
			}
		}

		if totalAmount < promo.MinPurchase {
			return nil, errors.New(constants.ErrPromoMinPurchase)
		}

		// Calculate discount
		var discount int64
		switch promo.DiscountType {
		case "percentage":
			discount = totalAmount * promo.DiscountValue / 100
			if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
				discount = promo.MaxDiscount
			}
		case "fixed":
			discount = promo.DiscountValue
			if discount > totalAmount {
				discount = totalAmount
			}
		}

		// Save promo snapshot
		promoSnapshot = &PromoSnapshot{
			Id:             promo.Id,
			Code:           promo.Code,
			Name:           promo.Name,
			DiscountType:   promo.DiscountType,
			DiscountAmount: discount,
		}

		// Apply discount to total
		totalAmount -= discount
	}

	// Add promo snapshot to order items
	for i := range orderItems {
		if promoSnapshot != nil {
			orderItems[i].PromoId = &promoSnapshot.Id
			orderItems[i].PromoCode = &promoSnapshot.Code
			orderItems[i].PromoName = &promoSnapshot.Name
			orderItems[i].PromoDiscountType = &promoSnapshot.DiscountType
			orderItems[i].PromoDiscountAmount = &promoSnapshot.DiscountAmount
		}
	}

	order := domain.Order{
		Id:          orderId,
		OrderRef:    orderRef,
		UserId:      args.UserId,
		AddressId:   args.AddressId,
		TotalAmount: totalAmount,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	if err := uc.repo.CreateWithItems(ctx, order, orderItems); err != nil {
		return nil, err
	}

	// Reserve stock via gRPC
	_, err = uc.productClient.ReserveStock(ctx, orderId, stockItems)
	if err != nil {
		// If stock reservation fails, we should delete the order
		// For now, we'll just return the error
		return nil, err
	}

	// Decrement flash sale stock and record usage
	for _, item := range orderItems {
		if item.FlashSaleId != nil && item.FlashSaleQuantity != nil {
			// Decrement flash sale stock
			_, err = uc.pricingClient.DecrementFlashSaleStock(ctx, *item.FlashSaleId, int32(*item.FlashSaleQuantity))
			if err != nil {
				// Log error but don't fail the order
				// Flash sale stock will be handled separately
			}

			// Record flash sale usage
			_, err = uc.pricingClient.IncrementFlashSaleUsage(ctx, *item.FlashSaleId, args.UserId, int32(*item.FlashSaleQuantity))
			if err != nil {
				// Log error but don't fail the order
			}
		}
	}

	return &order, nil
}

type PromoSnapshot struct {
	Id             string
	Code           string
	Name           string
	DiscountType   string
	DiscountAmount int64
}

func (uc *orderUseCase) List(ctx context.Context, params dto.ListOrderParams) ([]domain.Order, int64, error) {
	orders, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	count, err := uc.repo.ListCount(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	return orders, count, nil
}

func (uc *orderUseCase) ReadById(ctx context.Context, id string) (*domain.Order, error) {
	return uc.repo.ReadById(ctx, id)
}

func (uc *orderUseCase) ReadByIdWithItems(ctx context.Context, id string) (*domain.Order, []domain.OrderItem, error) {
	order, err := uc.repo.ReadById(ctx, id)
	if err != nil || order == nil {
		return order, nil, err
	}

	items, err := uc.repo.ReadItemsByOrderId(ctx, id)
	if err != nil {
		return order, nil, err
	}

	uc.enrichVariantInfo(ctx, items)
	return order, items, nil
}

func (uc *orderUseCase) enrichVariantInfo(ctx context.Context, items []domain.OrderItem) {
	productIds := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if _, ok := seen[item.ProductId]; !ok {
			seen[item.ProductId] = struct{}{}
			productIds = append(productIds, item.ProductId)
		}
	}
	if len(productIds) == 0 {
		return
	}

	products, err := uc.productClient.GetProducts(ctx, productIds)
	if err != nil {
		return
	}
	variantInfo := make(map[string]*productpb.ProductVariant)
	for _, p := range products {
		for _, v := range p.Variants {
			variantInfo[v.Id] = v
		}
	}

	for i := range items {
		if items[i].VariantId == nil {
			continue
		}
		if v, ok := variantInfo[*items[i].VariantId]; ok {
			items[i].VariantName = v.Name
			items[i].VariantAttributes = v.Attributes
		}
	}
}

func (uc *orderUseCase) ReadByUserId(ctx context.Context, userId string, before string, pageSize int32) ([]domain.Order, error) {
	return uc.repo.ReadByUserId(ctx, userId, before, pageSize)
}

func (uc *orderUseCase) UpdateStatus(ctx context.Context, id string, status string) error {
	allowedFrom := map[string][]string{
		"processing": {"pending", "paid"},
		"cancelled":  {"pending", "processing"},
		"refunded":   {"paid", "processing", "shipped"},
		"shipped":    {"processing"},
		"delivered":  {"shipped"},
	}
	fromStatuses, ok := allowedFrom[status]
	if !ok {
		return errors.New(constants.ErrInvalidOrderStatusTransition)
	}
	return uc.repo.UpdateStatus(ctx, id, fromStatuses, status)
}
