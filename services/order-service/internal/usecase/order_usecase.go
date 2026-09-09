package usecase

import (
	"context"
	"errors"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/internal/interfaces/IUseCase"
	grpcclient "kaffein/order-service/pkg/grpc/client"
	"kaffein/order-service/pkg/kafka"
	"kaffein/order-service/proto/dynamic-pricing"
	"kaffein/order-service/proto/product"
	"kaffein/order-service/utils"
	"kaffein/order-service/utils/constants"
	"math"
	"time"

	"github.com/google/uuid"
)

type orderUseCase struct {
	repo          IRepository.OrderRepository
	productClient *grpcclient.ProductClient
	pricingClient *grpcclient.DynamicPricingClient
	outbox        OutboxRepo
}

type OutboxRepo interface {
	Insert(ctx context.Context, topic, key string, payload any) error
}

func NewOrderUseCase(
	repo IRepository.OrderRepository,
	productClient *grpcclient.ProductClient,
	pricingClient *grpcclient.DynamicPricingClient,
	outbox OutboxRepo,
) IUseCase.OrderUseCase {
	return &orderUseCase{
		repo:          repo,
		productClient: productClient,
		pricingClient: pricingClient,
		outbox:        outbox,
	}
}

func (uc *orderUseCase) Create(ctx context.Context, args dto.CreateOrderDto) (*domain.Order, error) {
	for _, item := range args.Items {
		if item.Quantity <= 0 || item.Quantity > math.MaxInt32 {
			return nil, errors.New(constants.ErrInvalidQuantity)
		}
	}

	orderId := uuid.New().String()
	orderRef := utils.GenerateOrderRef()

	var productIds []string
	for _, item := range args.Items {
		productIds = append(productIds, item.ProductId)
	}

	products, err := uc.productClient.GetProducts(ctx, productIds)
	if err != nil {
		return nil, err
	}

	productMap := make(map[string]*productpb.Product)
	for _, p := range products {
		productMap[p.Id] = p
	}

	for _, item := range args.Items {
		if _, ok := productMap[item.ProductId]; !ok {
			return nil, errors.New(constants.ErrProductNotFound)
		}
	}

	var totalAmount int64
	var subtotal int64
	var orderItems []domain.OrderItem
	var stockItems []*productpb.StockItem

	for _, item := range args.Items {
		product := productMap[item.ProductId]

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

		var unitPrice int64
		var availableStock int64
		if variant != nil {
			unitPrice = variant.Price
			availableStock = variant.Stock - variant.StockReserved
		} else {
			unitPrice = product.Price
			availableStock = product.Stock - product.StockReserved
		}
		if unitPrice < 0 || unitPrice > math.MaxInt64/int64(item.Quantity) {
			return nil, errors.New(constants.ErrOrderItemTotalOverflow)
		}
		itemSubtotal := unitPrice * int64(item.Quantity)
		if subtotal > math.MaxInt64-itemSubtotal {
			return nil, errors.New(constants.ErrOrderSubtotalOverflow)
		}

		if availableStock < int64(item.Quantity) {
			return nil, errors.New(constants.ErrInsufficientStock)
		}

		stockItem := &productpb.StockItem{
			ProductId: item.ProductId,
			VariantId: item.VariantId,
			Quantity:  int32(item.Quantity),
		}
		stockItems = append(stockItems, stockItem)

		var flashSaleId, flashSaleName *string
		var flashSaleDiscountPrice, flashSaleOriginalPrice *int64
		var flashSaleDiscountPercent *int
		var flashSaleQuantity *int

		var flashSale *dynamicpricingpb.FlashSale
		if variant != nil {
			flashSale, err = uc.pricingClient.GetFlashSaleByVariantId(ctx, item.ProductId, item.VariantId)
		} else {
			flashSale, err = uc.pricingClient.GetFlashSaleByProductId(ctx, item.ProductId)
		}

		if err != nil {
			return nil, err
		}
		if flashSale != nil && flashSale.IsActive && flashSale.DiscountPercent > 0 {
			fsId := flashSale.Id
			fsName := flashSale.Name
			fsDiscountPercent := int(flashSale.DiscountPercent)
			fsDiscountPrice := utils.DiscountedPrice(unitPrice, flashSale.DiscountPercent)
			fsOriginalPrice := unitPrice

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

		var finalPrice int64
		if flashSaleDiscountPrice != nil && flashSaleQuantity != nil {
			flashSaleTotal := *flashSaleDiscountPrice * int64(*flashSaleQuantity)
			normalQuantity := item.Quantity - *flashSaleQuantity
			normalTotal := unitPrice * int64(normalQuantity)
			finalPrice = flashSaleTotal + normalTotal
		} else {
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
		subtotal += itemSubtotal
	}

	allocationRequest := &dynamicpricingpb.AllocatePricingRequest{
		OrderId: orderId, UserId: args.UserId, PromoCode: args.PromoCode, Subtotal: subtotal,
	}
	for _, item := range orderItems {
		requestItem := &dynamicpricingpb.PricingAllocationItemRequest{
			ItemId: item.Id, Quantity: int32(item.Quantity), UnitPrice: item.ProductPrice,
		}
		if item.FlashSaleId != nil {
			requestItem.FlashSaleId = *item.FlashSaleId
		}
		allocationRequest.Items = append(allocationRequest.Items, requestItem)
	}

	allocation, err := uc.pricingClient.AllocatePricing(ctx, allocationRequest)
	if err != nil {
		return nil, errors.Join(err, uc.compensate(ctx, orderId, args.UserId, stockItems))
	}

	allocatedItems := make(map[string]*dynamicpricingpb.PricingAllocationItem, len(allocation.Items))
	for _, item := range allocation.Items {
		allocatedItems[item.ItemId] = item
	}
	totalAmount = 0
	for i := range orderItems {
		item := &orderItems[i]
		item.FlashSaleId = nil
		item.FlashSaleName = nil
		item.FlashSaleDiscountPercent = nil
		item.FlashSaleDiscountPrice = nil
		item.FlashSaleOriginalPrice = nil
		item.FlashSaleQuantity = nil
		item.FinalPrice = item.ProductPrice * int64(item.Quantity)
		if allocated := allocatedItems[item.Id]; allocated != nil {
			flashSaleID, name := allocated.FlashSaleId, allocated.Name
			discountPercent, quantity := int(allocated.DiscountPercent), int(allocated.Quantity)
			discountPrice, originalPrice := utils.DiscountedPrice(item.ProductPrice, allocated.DiscountPercent), item.ProductPrice
			item.FlashSaleId, item.FlashSaleName = &flashSaleID, &name
			item.FlashSaleDiscountPercent, item.FlashSaleQuantity = &discountPercent, &quantity
			item.FlashSaleDiscountPrice, item.FlashSaleOriginalPrice = &discountPrice, &originalPrice
			item.FinalPrice = discountPrice*int64(quantity) + item.ProductPrice*int64(item.Quantity-quantity)
		}
		if allocation.PromoId != "" {
			promoID, code, name := allocation.PromoId, allocation.PromoCode, allocation.PromoName
			discountType, discount := allocation.PromoDiscountType, allocation.PromoDiscountAmount
			item.PromoId, item.PromoCode, item.PromoName = &promoID, &code, &name
			item.PromoDiscountType, item.PromoDiscountAmount = &discountType, &discount
		}
		totalAmount += item.FinalPrice
	}
	totalAmount -= allocation.PromoDiscountAmount

	order := domain.Order{
		Id:          orderId,
		OrderRef:    orderRef,
		UserId:      args.UserId,
		AddressId:   args.AddressId,
		TotalAmount: totalAmount,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	_, err = uc.productClient.ReserveStock(ctx, orderId, stockItems)
	if err != nil {
		return nil, errors.Join(err, uc.compensate(ctx, orderId, args.UserId, stockItems))
	}

	if err := uc.repo.CreateWithItemsWithTx(ctx, order, orderItems); err != nil {
		return nil, errors.Join(err, uc.compensate(ctx, orderId, args.UserId, stockItems))
	}

	return &order, nil
}

func (uc *orderUseCase) releaseStock(orderID string, items []*productpb.StockItem) error {
	rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(context.Background()), 5*time.Second)
	defer cancel()
	_, err := uc.productClient.ReleaseStock(rollbackCtx, orderID, items)
	return err
}

func (uc *orderUseCase) compensate(ctx context.Context, orderID, userID string, stockItems []*productpb.StockItem) error {
	return errors.Join(uc.releaseStock(orderID, stockItems), uc.releasePricing(orderID, userID))
}

func (uc *orderUseCase) releasePricing(orderID, userID string) error {
	rollbackCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return uc.pricingClient.ReleasePricing(rollbackCtx, orderID, userID)
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
	if err := uc.repo.UpdateStatus(ctx, id, fromStatuses, status); err != nil {
		return err
	}

	if status == "cancelled" || status == "refunded" {
		items, err := uc.repo.ReadItemsByOrderId(ctx, id)
		if err != nil {
			return err
		}
		order, err := uc.repo.ReadById(ctx, id)
		if err != nil {
			return err
		}
		userID := ""
		if order != nil {
			userID = order.UserId
		}
		eventType := "order.cancelled"
		if status == "refunded" {
			eventType = "order.refunded"
		}
		event := kafka.OrderEvent{Type: eventType, OrderID: id, UserID: userID}
		for _, item := range items {
			event.Items = append(event.Items, kafka.OrderEventItem{
				ProductID: item.ProductId,
				VariantID: item.VariantId,
				Quantity:  item.Quantity,
			})
		}
		return uc.outbox.Insert(ctx, "order-events", id, event)
	}
	return nil
}
