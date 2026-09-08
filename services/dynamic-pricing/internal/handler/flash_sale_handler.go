package handler

import (
	"kaffein/dynamic-pricing-service/internal/auth"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IUseCase"
	"kaffein/dynamic-pricing-service/utils/constants"
	"kaffein/dynamic-pricing-service/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type flashSaleHandler struct {
	usecase IUseCase.FlashSaleUseCase
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewFlashSaleHandler(usecase IUseCase.FlashSaleUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *flashSaleHandler {
	h := &flashSaleHandler{
		usecase: usecase,
		logger:  logger,
		v:       validator.Get(),
	}

	// Buyer: read-only flash sales
	flashSales := engine.Group("/flash-sales")
	{
		flashSales.GET("", h.List)
		flashSales.GET("/:id", h.ReadById)
	}

	// Admin: full CRUD flash sales
	adminFlashSales := engine.Group("/admin/flash-sales")
	{
		adminFlashSales.Use(authorization.Authenticate())
		adminFlashSales.GET("", authorization.Authorize("flash-sales", "read"), h.List)
		adminFlashSales.POST("", authorization.Authorize("flash-sales", "create"), h.Create)
		adminFlashSales.POST("/bulk", authorization.Authorize("flash-sales", "create"), h.CreateBulk)
		adminFlashSales.GET("/:id", authorization.Authorize("flash-sales", "read"), h.ReadById)
		adminFlashSales.PUT("/:id", authorization.Authorize("flash-sales", "update"), h.Update)
		adminFlashSales.DELETE("/:id", authorization.Authorize("flash-sales", "delete"), h.Delete)
	}

	return h
}

func (h *flashSaleHandler) List(c *gin.Context) {
	var params dto.ListFlashSaleParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	params.PageOffset *= params.PageSize

	flashSales, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list flash sales")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  flashSales,
		"count": count,
	})
}

func (h *flashSaleHandler) Create(c *gin.Context) {
	var args dto.CreateFlashSaleDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), args); err != nil {
		h.logger.Error().Err(err).Msg("failed to create flash sale")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "flash sale created"})
}

func (h *flashSaleHandler) CreateBulk(c *gin.Context) {
	var args dto.CreateFlashSaleBulkDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.CreateBulk(c.Request.Context(), args); err != nil {
		h.logger.Error().Err(err).Msg("failed to create flash sales bulk")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "flash sales created"})
}

func (h *flashSaleHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var args dto.UpdateFlashSaleDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, args); err != nil {
		if err.Error() == constants.ErrFlashSaleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update flash sale")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "flash sale updated"})
}

func (h *flashSaleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == constants.ErrFlashSaleNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete flash sale")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "flash sale deleted"})
}

func (h *flashSaleHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	flashSale, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read flash sale by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if flashSale == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrFlashSaleNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": flashSale})
}
