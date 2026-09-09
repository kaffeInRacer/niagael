package handler

import (
	"context"
	"kaffein/product-service/internal/auth"
	"kaffein/product-service/internal/domain"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IUseCase"
	grpcclient "kaffein/product-service/pkg/grpc/client"
	dynamicpricingpb "kaffein/product-service/proto/dynamic-pricing"
	"kaffein/product-service/utils/constants"
	"kaffein/product-service/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type FlashSaleProjectionReader interface {
	ListActive(ctx context.Context) ([]domain.FlashSaleProjection, error)
}

type productHandler struct {
	usecase       IUseCase.ProductUseCase
	imageHandler  *productImageHandler
	pricingClient *grpcclient.DynamicPricingClient
	logger        zerolog.Logger
	v             *validator.Validator
	projection    FlashSaleProjectionReader
}

func NewProductHandler(
	usecase IUseCase.ProductUseCase,
	pricingClient *grpcclient.DynamicPricingClient,
	logger zerolog.Logger,
	engine *gin.Engine,
	imageHandler *productImageHandler,
	authorization *auth.Service,
	projection FlashSaleProjectionReader) *productHandler {
	h := &productHandler{
		usecase:       usecase,
		imageHandler:  imageHandler,
		pricingClient: pricingClient,
		logger:        logger,
		v:             validator.Get(),
		projection:    projection,
	}

	products := engine.Group("/products")
	products.GET("", h.ListBuyer)
	products.GET("/slug/:slug", h.ReadActiveBySlug)
	products.GET("/:id", h.ReadActiveById)
	products.GET("/:id/images", h.ListActiveImages)

	adminProducts := engine.Group("/admin/products")
	adminProducts.Use(authorization.Authenticate())
	adminProducts.GET("", authorization.Authorize("products", "read"), h.List)
	adminProducts.POST("", authorization.Authorize("products", "create"), h.Create)
	adminProducts.GET("/slug/:slug", authorization.Authorize("products", "read"), h.ReadBySlug)
	adminProducts.GET("/:id", authorization.Authorize("products", "read"), h.ReadById)
	adminProducts.PUT("/:id", authorization.Authorize("products", "update"), h.Update)
	adminProducts.DELETE("/:id", authorization.Authorize("products", "delete"), h.Delete)
	adminProducts.GET("/:id/images", authorization.Authorize("product-images", "read"), imageHandler.ListByProductId)
	adminProducts.POST("/:id/images", authorization.Authorize("product-images", "create"), imageHandler.Create)
	adminProducts.DELETE("/:id/images/:image_id", authorization.Authorize("product-images", "delete"), imageHandler.Delete)

	return h
}

func (h *productHandler) getFlashSalesForProducts(ctx *gin.Context, products []domain.Product) (map[string]*dto.FlashSaleInfo, map[string][]dto.VariantFlashSale) {
	if h.projection != nil {
		info, variants, found := h.getFlashSalesFromProjection(ctx, products)
		if found {
			return info, variants
		}
	}
	if h.pricingClient == nil {
		return nil, nil
	}

	var productIds []string
	var variantRequests []*dynamicpricingpb.FlashSaleVariantRequest

	for _, p := range products {
		if len(p.Variants) == 0 {
			productIds = append(productIds, p.Id)
		} else {
			for _, v := range p.Variants {
				if v.IsActive {
					variantRequests = append(variantRequests, &dynamicpricingpb.FlashSaleVariantRequest{
						ProductId: p.Id,
						VariantId: v.Id,
					})
				}
			}
		}
	}

	productFlashSales := make(map[string]*dto.FlashSaleInfo)
	if len(productIds) > 0 {
		resp, err := h.pricingClient.GetFlashSalesByProductIds(ctx.Request.Context(), productIds)
		if err == nil {
			for _, fs := range resp.FlashSales {
				if fs.IsActive && fs.Id != "" {
					productFlashSales[fs.ProductId] = &dto.FlashSaleInfo{
						ProductId:       fs.ProductId,
						DiscountPercent: int32(fs.DiscountPercent),
					}
				}
			}
		}
	}

	variantFlashSalesMap := make(map[string][]dto.VariantFlashSale)
	if len(variantRequests) > 0 {
		resp, err := h.pricingClient.GetFlashSalesByVariantIds(ctx.Request.Context(), variantRequests)
		if err == nil {
			for _, fs := range resp.FlashSales {
				if fs.IsActive && fs.Id != "" && fs.DiscountPercent > 0 {
					variantFlashSalesMap[fs.ProductId] = append(variantFlashSalesMap[fs.ProductId], dto.VariantFlashSale{
						VariantId:       fs.VariantId,
						DiscountPercent: int32(fs.DiscountPercent),
					})
				}
			}
		}
	}

	return productFlashSales, variantFlashSalesMap
}

func (h *productHandler) getFlashSalesFromProjection(ctx *gin.Context, products []domain.Product) (map[string]*dto.FlashSaleInfo, map[string][]dto.VariantFlashSale, bool) {
	items, err := h.projection.ListActive(ctx.Request.Context())
	if err != nil || len(items) == 0 {
		return nil, nil, false
	}

	productFlashSales := make(map[string]*dto.FlashSaleInfo)
	variantFlashSalesMap := make(map[string][]dto.VariantFlashSale)
	for _, p := range items {
		if p.VariantID == "" {
			productFlashSales[p.ProductID] = &dto.FlashSaleInfo{
				ProductId:       p.ProductID,
				DiscountPercent: int32(p.DiscountPercent),
			}
		} else if p.DiscountPercent > 0 {
			variantFlashSalesMap[p.ProductID] = append(variantFlashSalesMap[p.ProductID], dto.VariantFlashSale{
				VariantId:       p.VariantID,
				DiscountPercent: int32(p.DiscountPercent),
			})
		}
	}
	return productFlashSales, variantFlashSalesMap, true
}

func (h *productHandler) List(c *gin.Context) {
	var params dto.ListProductParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	params.PageOffset = (params.Page - 1) * params.PageSize

	products, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}
	for i := range products {
		activeProduct(&products[i])
	}

	productFlashSales, _ := h.getFlashSalesForProducts(c, products)

	var responses []dto.AdminProductListResponse
	for _, p := range products {
		flashSale := productFlashSales[p.Id]
		responses = append(responses, dto.ToAdminProductListResponse(&p, flashSale))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"count": count,
	})
}

func (h *productHandler) ListActiveImages(c *gin.Context) {
	product, err := h.usecase.ReadById(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if activeProduct(product) == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrProductNotFound})
		return
	}
	h.imageHandler.ListByProductId(c)
}

func (h *productHandler) ListBuyer(c *gin.Context) {
	var params dto.ListProductParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	active := true
	params.IsActive = &active
	params.PublicOnly = true
	params.PageOffset = (params.Page - 1) * params.PageSize

	products, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	productFlashSales, variantFlashSalesMap := h.getFlashSalesForProducts(c, products)

	var responses []dto.ProductListResponse
	for _, p := range products {
		flashSale := productFlashSales[p.Id]
		variantFlashSales := variantFlashSalesMap[p.Id]
		responses = append(responses, dto.ToProductListResponse(&p, flashSale, variantFlashSales))
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  responses,
		"count": count,
	})
}

func activeProduct(product *domain.Product) *domain.Product {
	if product == nil || !product.IsActive || !product.CategoryActive {
		return nil
	}

	variants := product.Variants[:0]
	for _, variant := range product.Variants {
		if variant.IsActive {
			variants = append(variants, variant)
		}
	}
	product.Variants = variants
	return product
}

func (h *productHandler) Create(c *gin.Context) {
	var args dto.CreateProductDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), args); err != nil {
		h.logger.Error().Err(err).Msg("failed to create product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "product created"})
}

func (h *productHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var args dto.UpdateProductDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, args); err != nil {
		if err.Error() == constants.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product updated"})
}

func (h *productHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == constants.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete product")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *productHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	product, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrProductNotFound})
		return
	}

	productFlashSales, variantFlashSalesMap := h.getFlashSalesForProducts(c, []domain.Product{*product})

	flashSale := productFlashSales[product.Id]
	variantFlashSales := variantFlashSalesMap[product.Id]

	response := dto.ToProductDetailResponse(product, flashSale, variantFlashSales)
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *productHandler) ReadActiveById(c *gin.Context) {
	id := c.Param("id")
	product, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	h.writeActiveProduct(c, product)
}

func (h *productHandler) ReadBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.usecase.ReadBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by slug")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrProductNotFound})
		return
	}

	productFlashSales, variantFlashSalesMap := h.getFlashSalesForProducts(c, []domain.Product{*product})

	flashSale := productFlashSales[product.Id]
	variantFlashSales := variantFlashSalesMap[product.Id]

	response := dto.ToProductDetailResponse(product, flashSale, variantFlashSales)
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *productHandler) ReadActiveBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.usecase.ReadBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by slug")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	h.writeActiveProduct(c, product)
}

func (h *productHandler) writeActiveProduct(c *gin.Context, product *domain.Product) {
	product = activeProduct(product)
	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrProductNotFound})
		return
	}

	productFlashSales, variantFlashSalesMap := h.getFlashSalesForProducts(c, []domain.Product{*product})
	response := dto.ToProductDetailResponse(product, productFlashSales[product.Id], variantFlashSalesMap[product.Id])
	c.JSON(http.StatusOK, gin.H{"data": response})
}
