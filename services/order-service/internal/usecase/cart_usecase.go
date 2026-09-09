package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IRepository"
	"kaffein/order-service/internal/interfaces/IUseCase"
	dynamicpricingpb "kaffein/order-service/proto/dynamic-pricing"
	productpb "kaffein/order-service/proto/product"
	"kaffein/order-service/utils"
	"kaffein/order-service/utils/constants"
)

type cartProductClient interface {
	GetProducts(ctx context.Context, ids []string) ([]*productpb.Product, error)
}

type cartPricingClient interface {
	GetFlashSaleByProductId(ctx context.Context, productId string) (*dynamicpricingpb.FlashSale, error)
	GetFlashSaleByVariantId(ctx context.Context, productId string, variantId string) (*dynamicpricingpb.FlashSale, error)
}

type cartUseCase struct {
	repo          IRepository.CartRepository
	productClient cartProductClient
	pricingClient cartPricingClient
}

func NewCartUseCase(repo IRepository.CartRepository, productClient cartProductClient, pricingClient cartPricingClient) IUseCase.CartUseCase {
	return &cartUseCase{
		repo:          repo,
		productClient: productClient,
		pricingClient: pricingClient,
	}
}

func (uc *cartUseCase) ReadByUserId(ctx context.Context, userId string) (*dto.CartResponse, error) {
	cart, err := uc.repo.ReadByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return &dto.CartResponse{UserId: userId, Items: []dto.CartItemResponse{}}, nil
	}
	return uc.quote(ctx, cart)
}

func (uc *cartUseCase) AddItem(ctx context.Context, args dto.AddCartItemDto) (*dto.CartResponse, error) {
	availableStock, err := uc.availableStock(ctx, args.ProductId, args.VariantId)
	if err != nil {
		return nil, err
	}
	if availableStock < int64(args.Quantity) {
		return nil, errors.New(constants.ErrCartQuantityLimit)
	}

	var variantId *string
	if args.VariantId != "" {
		variantId = &args.VariantId
	}
	item := domain.CartItem{
		Id:        uuid.New().String(),
		ProductId: args.ProductId,
		VariantId: variantId,
		Quantity:  args.Quantity,
	}
	if err := uc.repo.AddItemWithTx(ctx, args.UserId, uuid.New().String(), item, int(availableStock)); err != nil {
		return nil, err
	}
	return uc.ReadByUserId(ctx, args.UserId)
}

func (uc *cartUseCase) UpdateItem(ctx context.Context, itemId string, args dto.UpdateCartItemDto) (*dto.CartResponse, error) {
	cart, err := uc.repo.ReadByUserId(ctx, args.UserId)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, errors.New(constants.ErrCartItemNotFound)
	}

	var selected *domain.CartItem
	for i := range cart.Items {
		if cart.Items[i].Id == itemId {
			selected = &cart.Items[i]
			break
		}
	}
	if selected == nil {
		return nil, errors.New(constants.ErrCartItemNotFound)
	}

	variantId := ""
	if selected.VariantId != nil {
		variantId = *selected.VariantId
	}
	availableStock, err := uc.availableStock(ctx, selected.ProductId, variantId)
	if err != nil {
		return nil, err
	}
	if availableStock < int64(args.Quantity) {
		return nil, errors.New(constants.ErrCartQuantityLimit)
	}
	if err := uc.repo.UpdateItem(ctx, itemId, args.UserId, args.Quantity, int(availableStock)); err != nil {
		return nil, err
	}
	return uc.ReadByUserId(ctx, args.UserId)
}

func (uc *cartUseCase) DeleteItem(ctx context.Context, itemId string, userId string) error {
	return uc.repo.DeleteItem(ctx, itemId, userId)
}

func (uc *cartUseCase) Clear(ctx context.Context, userId string) error {
	return uc.repo.Clear(ctx, userId)
}

func (uc *cartUseCase) availableStock(ctx context.Context, productId string, variantId string) (int64, error) {
	products, err := uc.productClient.GetProducts(ctx, []string{productId})
	if err != nil {
		return 0, err
	}
	if len(products) == 0 {
		return 0, errors.New(constants.ErrProductNotFound)
	}
	product := products[0]
	if variantId == "" {
		return product.Stock - product.StockReserved, nil
	}
	for _, variant := range product.Variants {
		if variant.Id == variantId {
			return variant.Stock - variant.StockReserved, nil
		}
	}
	return 0, errors.New(constants.ErrVariantNotFound)
}

func (uc *cartUseCase) quote(ctx context.Context, cart *domain.Cart) (*dto.CartResponse, error) {
	response := &dto.CartResponse{
		Id:        cart.Id,
		UserId:    cart.UserId,
		Items:     make([]dto.CartItemResponse, 0, len(cart.Items)),
		CreatedAt: &cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
	if len(cart.Items) == 0 {
		return response, nil
	}

	productIds := make([]string, 0, len(cart.Items))
	seen := make(map[string]struct{}, len(cart.Items))
	for _, item := range cart.Items {
		if _, ok := seen[item.ProductId]; !ok {
			seen[item.ProductId] = struct{}{}
			productIds = append(productIds, item.ProductId)
		}
	}
	products, err := uc.productClient.GetProducts(ctx, productIds)
	if err != nil {
		return nil, err
	}
	productMap := make(map[string]*productpb.Product, len(products))
	for _, product := range products {
		productMap[product.Id] = product
	}

	for _, item := range cart.Items {
		product := productMap[item.ProductId]
		if product == nil {
			return nil, errors.New(constants.ErrProductNotFound)
		}

		unitPrice := product.Price
		availableStock := product.Stock - product.StockReserved
		variantName := ""
		var variantAttributes map[string]string
		variantId := ""
		if item.VariantId != nil {
			variantId = *item.VariantId
			var selected *productpb.ProductVariant
			for _, variant := range product.Variants {
				if variant.Id == variantId {
					selected = variant
					break
				}
			}
			if selected == nil {
				return nil, errors.New(constants.ErrVariantNotFound)
			}
			unitPrice = selected.Price
			availableStock = selected.Stock - selected.StockReserved
			variantName = selected.Name
			variantAttributes = selected.Attributes
		}

		var flashSale *dynamicpricingpb.FlashSale
		if variantId == "" {
			flashSale, err = uc.pricingClient.GetFlashSaleByProductId(ctx, item.ProductId)
		} else {
			flashSale, err = uc.pricingClient.GetFlashSaleByVariantId(ctx, item.ProductId, variantId)
		}
		if err != nil {
			return nil, err
		}

		flashSaleUnitPrice := unitPrice
		flashSaleQuantity := 0
		var flashSaleId, flashSaleName *string
		var discountPercent int32
		if flashSale != nil && flashSale.Id != "" && flashSale.IsActive && flashSale.DiscountPercent > 0 {
			discountPercent = flashSale.DiscountPercent
			flashSaleUnitPrice = utils.DiscountedPrice(unitPrice, discountPercent)
			flashSaleQuantity = item.Quantity
			if flashSale.MaxPerUser > 0 && flashSaleQuantity > int(flashSale.MaxPerUser) {
				flashSaleQuantity = int(flashSale.MaxPerUser)
			}
			if int64(flashSaleQuantity) > flashSale.Stock {
				flashSaleQuantity = int(flashSale.Stock)
			}
			id, name := flashSale.Id, flashSale.Name
			flashSaleId, flashSaleName = &id, &name
		}

		normalQuantity := item.Quantity - flashSaleQuantity
		lineTotal := flashSaleUnitPrice*int64(flashSaleQuantity) + unitPrice*int64(normalQuantity)
		response.Items = append(response.Items, dto.CartItemResponse{
			Id:                       item.Id,
			ProductId:                item.ProductId,
			VariantId:                item.VariantId,
			ProductName:              product.Name,
			VariantName:              variantName,
			VariantAttributes:        variantAttributes,
			Quantity:                 item.Quantity,
			AvailableStock:           availableStock,
			OriginalUnitPrice:        unitPrice,
			FlashSaleUnitPrice:       flashSaleUnitPrice,
			FlashSaleQuantity:        flashSaleQuantity,
			NormalQuantity:           normalQuantity,
			LineTotal:                lineTotal,
			FlashSaleId:              flashSaleId,
			FlashSaleName:            flashSaleName,
			FlashSaleDiscountPercent: discountPercent,
		})
		response.Subtotal += lineTotal
		response.TotalItems += item.Quantity
	}
	return response, nil
}
