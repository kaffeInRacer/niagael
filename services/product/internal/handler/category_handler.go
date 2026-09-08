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

type categoryHandler struct {
	usecase IUseCase.CategoryUseCase
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewCategoryHandler(usecase IUseCase.CategoryUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *categoryHandler {
	h := &categoryHandler{
		usecase: usecase,
		logger:  logger,
		v:       validator.Get(),
	}

	// Buyer routes (read-only)
	categories := engine.Group("/categories")
	categories.GET("", h.List)
	categories.GET("/:id", h.ReadById)
	categories.GET("/slug/:slug", h.ReadBySlug)

	// Admin routes (full CRUD)
	adminCategories := engine.Group("/admin/categories")
	adminCategories.Use(authorization.Authenticate())
	adminCategories.GET("", authorization.Authorize("category", "read"), h.List)
	adminCategories.POST("", authorization.Authorize("category", "create"), h.Create)
	adminCategories.GET("/slug/:slug", authorization.Authorize("category", "read"), h.ReadBySlug)
	adminCategories.GET("/:id", authorization.Authorize("category", "read"), h.ReadById)
	adminCategories.PUT("/:id", authorization.Authorize("category", "update"), h.Update)
	adminCategories.DELETE("/:id", authorization.Authorize("category", "delete"), h.Delete)

	return h
}

func (h *categoryHandler) List(c *gin.Context) {
	var params dto.ListCategoryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	params.PageOffset *= params.PageSize

	categories, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list categories")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  categories,
		"count": count,
	})
}

func (h *categoryHandler) Create(c *gin.Context) {
	var args dto.CreateCategoryDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), args); err != nil {
		h.logger.Error().Err(err).Msg("failed to create category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "category created"})
}

func (h *categoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var args dto.UpdateCategoryDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, args); err != nil {
		if err.Error() == constants.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category updated"})
}

func (h *categoryHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == constants.ErrCategoryNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete category")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

func (h *categoryHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	category, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read category by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrCategoryNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": category})
}

func (h *categoryHandler) ReadBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.usecase.ReadBySlug(c.Request.Context(), slug)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read category by slug")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrCategoryNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": category})
}
