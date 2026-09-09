package handler

import (
	"kaffein/product-service/internal/auth"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/internal/interfaces/IUseCase"
	"kaffein/product-service/utils/constants"
	"kaffein/product-service/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type variantHandler struct {
	usecase IUseCase.VariantUseCase
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewVariantHandler(usecase IUseCase.VariantUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *variantHandler {
	h := &variantHandler{
		usecase: usecase,
		logger:  logger,
		v:       validator.Get(),
	}

	variants := engine.Group("/variants")
	variants.GET("", h.ListActive)

	adminVariants := engine.Group("/admin/variants")
	adminVariants.Use(authorization.Authenticate())
	adminVariants.GET("", authorization.Authorize("variants", "read"), h.List)
	adminVariants.POST("", authorization.Authorize("variants", "create"), h.Create)
	adminVariants.PUT("/:id", authorization.Authorize("variants", "update"), h.Update)
	adminVariants.DELETE("/:id", authorization.Authorize("variants", "delete"), h.Delete)

	return h
}

func (h *variantHandler) List(c *gin.Context) {
	h.list(c)
}

func (h *variantHandler) ListActive(c *gin.Context) {
	productId := c.Query("product_id")

	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	variants, err := h.usecase.ListActiveByProductId(c.Request.Context(), productId)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list variants")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variants})
}

func (h *variantHandler) list(c *gin.Context) {
	productId := c.Query("product_id")
	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	variants, err := h.usecase.ListByProductId(c.Request.Context(), productId)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list variants")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variants})
}

func (h *variantHandler) Create(c *gin.Context) {
	productId := c.Query("product_id")
	if productId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product_id is required"})
		return
	}

	var args dto.CreateVariantDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), productId, args); err != nil {
		if err.Error() == constants.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to create variant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "variant created"})
}

func (h *variantHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var args dto.UpdateVariantDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, args); err != nil {
		if err.Error() == constants.ErrVariantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update variant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "variant updated"})
}

func (h *variantHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == constants.ErrVariantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete variant")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "variant deleted"})
}
