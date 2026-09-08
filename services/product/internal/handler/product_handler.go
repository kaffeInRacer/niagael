package handler

import (
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

type productHandler struct {
	usecase       IUseCase.ProductUseCase
	pricingClient *grpcclient.DynamicPricingClient
	logger        zerolog.Logger
	v             *validator.Validator
}

func NewProductHandler(
	usecase IUseCase.ProductUseCase,
	pricingClient *grpcclient.DynamicPricingClient,
	logger zerolog.Logger,
	engine *gin.Engine,
	imageHandler *productImageHandler,
	authorization *auth.Service) *productHandler {
	h := &productHandler{
		usecase:       usecase,
		pricingClient: pricingClient,
		logger:        logger,
		v:             validator.Get(),
	}

	// Buyer routes (read-only)
	products := engine.Group("/products")
	products.GET("", h.ListBuyer)
	products.GET("/slug/:slug", h.ReadBySlug)
	products.GET("/:id", h.ReadById)
	products.GET("/:id/images", imageHandler.ListByProductId)

	// Admin routes (full CRUD)
	adminProducts := engine.Group("/admin/products")
	adminProducts.Use(authorization.Authenticate())
	adminProducts.GET("", authorization.Authorize("products", "read"), h.List)
	adminProducts.POST("", authorization.Authorize("products", "create"), h.Create)
	adminProducts.GET("/slug/:slug", authorization.Authorize("products", "read"), h.ReadBySlug)
	adminProducts.GET("/:id", authorization.Authorize("products", "read"), h.ReadById)
	adminProducts.PUT("/:id", authorization.Authorize("products", "update"), h.Update)
	adminProducts.DELETE("/:id", authorization.Authorize("products", "delete"), h.Delete)
	adminProducts.GET("/:id/images", authorization.Authorize("products", "read"), imageHandler.ListByProductId)
	adminProducts.POST("/:id/images", authorization.Authorize("products", "create"), imageHandler.Create)
	adminProducts.DELETE("/:id/images/:image_id", authorization.Authorize("products", "delete"), imageHandler.Delete)

	return h
}

// getFlashSalesForProducts fetches flash sales for all products in bulk
func (h *productHandler) getFlashSalesForProducts(ctx *gin.Context, products []domain.Product) (map[string]*dto.FlashSaleInfo, map[string][]dto.VariantFlashSale) {
	if h.pricingClient == nil {
		return nil, nil
	}

	// Collect all product IDs (for products without variants)
	var productIds []string
	// Collect all variant requests (for products with variants)
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

	// Bulk fetch flash sales for products without variants
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

	// Bulk fetch flash sales for variants
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
	params.PageOffset *= params.PageSize

	products, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Bulk fetch flash sales
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
	params.PageOffset *= params.PageSize

	products, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list products")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Bulk fetch flash sales
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
}

func (h *productHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	product, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrProductNotFound})
		return
	}

	// Fetch flash sales for this single product
	productFlashSales, variantFlashSalesMap := h.getFlashSalesForProducts(c, []domain.Product{*product})

	flashSale := productFlashSales[product.Id]
	variantFlashSales := variantFlashSalesMap[product.Id]

	response := dto.ToProductDetailResponse(product, flashSale, variantFlashSales)
	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *productHandler) ReadBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.usecase.ReadBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read product by slug")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if product == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrProductNotFound})
		return
	}

	// Fetch flash sales for this single product
	productFlashSales, variantFlashSalesMap := h.getFlashSalesForProducts(c, []domain.Product{*product})

	flashSale := productFlashSales[product.Id]
	variantFlashSales := variantFlashSalesMap[product.Id]

	response := dto.ToProductDetailResponse(product, flashSale, variantFlashSales)
	c.JSON(http.StatusOK, gin.H{"data": response})
}
