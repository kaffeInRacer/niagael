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

type categoryUseCase struct {
	repo IRepository.CategoryRepository
}

func NewCategoryUseCase(repo IRepository.CategoryRepository) *categoryUseCase {
	return &categoryUseCase{repo: repo}
}

func (uc *categoryUseCase) List(ctx context.Context, params dto.ListCategoryParams) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var err error

	if params.IncludeProductCount {
		categories, err = uc.repo.ListWithProductCount(ctx, params)
	} else {
		categories, err = uc.repo.List(ctx, params)
	}

	if err != nil {
		return nil, 0, err
	}

	var count int64
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		total, err := uc.repo.ListCount(gctx, params)
		count = total
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}

	return categories, count, nil
}

func (uc *categoryUseCase) Create(ctx context.Context, args dto.CreateCategoryDto) error {
	arg := dto.CreateCategoryDto{
		Id:          uuid.New().String(),
		Name:        args.Name,
		Description: args.Description,
		Slug:        utils.GenerateSlug(args.Name),
	}

	return uc.repo.Create(ctx, arg)
}

func (uc *categoryUseCase) Update(ctx context.Context, id string, args dto.UpdateCategoryDto) error {
	arg := dto.UpdateCategoryDto{
		Id:          id,
		Name:        args.Name,
		Description: args.Description,
		Slug:        utils.GenerateSlug(args.Name),
	}

	return uc.repo.Update(ctx, arg)
}

func (uc *categoryUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *categoryUseCase) ReadById(ctx context.Context, id string) (*domain.Category, error) {
	return uc.repo.ReadById(ctx, id)
}

func (uc *categoryUseCase) ReadBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return uc.repo.ReadBySlug(ctx, slug)
}
