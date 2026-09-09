package usecase

import (
	"golang.org/x/sync/errgroup"
	"context"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IRepository"
	"kaffein/product-service/utils"

	"github.com/google/uuid"
)

type productUseCase struct {
	repo        IRepository.ProductRepository
	imageRepo   IRepository.ProductImageRepository
	variantRepo IRepository.VariantRepository
	stockRepo   IRepository.StockRepository
}

func NewProductUseCase(repo IRepository.ProductRepository, imageRepo IRepository.ProductImageRepository, variantRepo IRepository.VariantRepository, stockRepo IRepository.StockRepository) *productUseCase {
	return &productUseCase{repo: repo, imageRepo: imageRepo, variantRepo: variantRepo, stockRepo: stockRepo}
}

func (uc *productUseCase) List(ctx context.Context, params dto.ListProductParams) ([]domain.Product, int64, error) {
	products, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	ids := make([]string, 0, len(products))
	for _, product := range products {
		ids = append(ids, product.Id)
	}

	var (
		count    int64
		images   []domain.ProductImage
		variants []domain.ProductVariant
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		total, err := uc.repo.ListCount(gctx, params)
		count = total
		return err
	})
	g.Go(func() error {
		data, err := uc.imageRepo.ListByProductIds(gctx, ids)
		images = data
		return err
	})
	g.Go(func() error {
		data, err := uc.variantRepo.ListByProductIds(gctx, ids)
		variants = data
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	byProduct := make(map[string][]domain.ProductImage)
	for _, image := range images {
		byProduct[image.ProductId] = append(byProduct[image.ProductId], image)
	}
	for i := range products {
		products[i].Images = byProduct[products[i].Id]
	}

	byProductVariants := make(map[string][]domain.ProductVariant)
	for _, v := range variants {
		byProductVariants[v.ProductId] = append(byProductVariants[v.ProductId], v)
	}
	for i := range products {
		products[i].Variants = byProductVariants[products[i].Id]
	}

	return products, count, nil
}

func (uc *productUseCase) Create(ctx context.Context, args dto.CreateProductDto) error {
	arg := dto.CreateProductDto{
		Id:          uuid.New().String(),
		CategoryId:  args.CategoryId,
		Name:        args.Name,
		Description: args.Description,
		Price:       args.Price,
		Stock:       args.Stock,
		IsActive:    args.IsActive,
		Slug:        utils.GenerateSlug(args.Name),
	}

	return uc.repo.Create(ctx, arg)
}

func (uc *productUseCase) Update(ctx context.Context, id string, args dto.UpdateProductDto) error {
	arg := dto.UpdateProductDto{
		Id:          id,
		CategoryId:  args.CategoryId,
		Name:        args.Name,
		Description: args.Description,
		Price:       args.Price,
		Stock:       args.Stock,
		IsActive:    args.IsActive,
		Slug:        utils.GenerateSlug(args.Name),
	}

	return uc.repo.Update(ctx, arg)
}

func (uc *productUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *productUseCase) ReadById(ctx context.Context, id string) (*domain.Product, error) {
	product, err := uc.repo.ReadById(ctx, id)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, nil
	}

	images, err := uc.imageRepo.ListByProductId(ctx, id)
	if err != nil {
		return nil, err
	}

	product.Images = images
	arr := make([]string, 0)
	arr = append(arr, id)
	variants, err := uc.variantRepo.ListByProductIds(ctx, arr)

	if err != nil {
		return nil, err
	}
	product.Variants = variants

	return product, nil
}

func (uc *productUseCase) ReadBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	product, err := uc.repo.ReadBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, nil
	}

	images, err := uc.imageRepo.ListByProductId(ctx, product.Id)
	if err != nil {
		return nil, err
	}

	product.Images = images
	variants, err := uc.variantRepo.ListByProductIds(ctx, []string{product.Id})
	if err != nil {
		return nil, err
	}

	product.Variants = variants

	return product, nil
}

func (uc *productUseCase) BatchById(ctx context.Context, ids []string) ([]domain.Product, error) {
	products, err := uc.repo.BatchById(ctx, ids)
	if err != nil {
		return nil, err
	}

	variants, err := uc.variantRepo.ListByProductIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	byProduct := make(map[string][]domain.ProductVariant)
	for _, variant := range variants {
		byProduct[variant.ProductId] = append(byProduct[variant.ProductId], variant)
	}

	for i := range products {
		products[i].Variants = byProduct[products[i].Id]
	}

	return products, nil
}

func (uc *productUseCase) ReserveStock(ctx context.Context, orderID string, items []domain.StockItem) error {
	return uc.stockRepo.ReserveWithTx(ctx, orderID, items)
}

func (uc *productUseCase) ReleaseStock(ctx context.Context, orderID string, items []domain.StockItem) error {
	return uc.stockRepo.ReleaseWithTx(ctx, orderID, items)
}

func (uc *productUseCase) ConfirmStock(ctx context.Context, orderID string, items []domain.StockItem) error {
	return uc.stockRepo.ConfirmWithTx(ctx, orderID, items)
}
