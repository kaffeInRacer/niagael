package handler

import (
	"kaffein/dynamic-pricing-service/internal/auth"
	"kaffein/dynamic-pricing-service/internal/domain"
	"kaffein/dynamic-pricing-service/internal/dto"
	"kaffein/dynamic-pricing-service/internal/interfaces/IUseCase"
	"kaffein/dynamic-pricing-service/utils/constants"
	"kaffein/dynamic-pricing-service/utils/validator"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type promoHandler struct {
	usecase IUseCase.PromoUseCase
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewPromoHandler(usecase IUseCase.PromoUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *promoHandler {
	h := &promoHandler{
		usecase: usecase,
		logger:  logger,
		v:       validator.Get(),
	}

	promos := engine.Group("/promos")
	{
		promos.GET("", h.ListCurrent)
		promos.GET("/:id", h.ReadCurrentById)
	}

	promoApply := engine.Group("/promos")
	{
		promoApply.Use(authorization.Authenticate())
		promoApply.POST("/apply/:code/:userId", authorization.Authorize("promos", "apply"), h.ApplyPromo)
	}

	adminPromos := engine.Group("/admin/promos")
	{
		adminPromos.Use(authorization.Authenticate())
		adminPromos.GET("", authorization.Authorize("promos", "read"), h.List)
		adminPromos.POST("", authorization.Authorize("promos", "create"), h.Create)
		adminPromos.GET("/:id", authorization.Authorize("promos", "read"), h.ReadById)
		adminPromos.PUT("/:id", authorization.Authorize("promos", "update"), h.Update)
		adminPromos.DELETE("/:id", authorization.Authorize("promos", "delete"), h.Delete)
	}

	return h
}

func (h *promoHandler) List(c *gin.Context) {
	h.list(c, false)
}

func (h *promoHandler) ListCurrent(c *gin.Context) {
	h.list(c, true)
}

func (h *promoHandler) list(c *gin.Context, currentOnly bool) {
	var params dto.ListPromoParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	if currentOnly {
		active := true
		params.IsActive = &active
		params.CurrentOnly = true
	}
	params.PageOffset = (params.Page - 1) * params.PageSize

	promos, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list promos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	data := any(promos)
	if currentOnly {
		publicPromos := make([]dto.PublicPromoResponse, 0, len(promos))
		for _, promo := range promos {
			publicPromos = append(publicPromos, dto.ToPublicPromoResponse(promo))
		}
		data = publicPromos
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"count": count,
	})
}

func (h *promoHandler) Create(c *gin.Context) {
	var args dto.CreatePromoDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), args); err != nil {
		h.logger.Error().Err(err).Msg("failed to create promo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "promo created"})
}

func (h *promoHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var args dto.UpdatePromoDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, args); err != nil {
		if err.Error() == constants.ErrPromoNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update promo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "promo updated"})
}

func (h *promoHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == constants.ErrPromoNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete promo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "promo deleted"})
}

func (h *promoHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	promo, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read promo by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if promo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrPromoNotFound})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": promo})
}

func (h *promoHandler) ReadCurrentById(c *gin.Context) {
	id := c.Param("id")
	promo, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read promo by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}
	if !currentPromo(promo, time.Now()) {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrPromoNotFound})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": dto.ToPublicPromoResponse(*promo)})
}

func currentPromo(promo *domain.Promo, now time.Time) bool {
	return promo != nil && promo.IsActive && !now.Before(promo.StartDate) && !now.After(promo.EndDate) &&
		(promo.Quantity == 0 || promo.UsedCount < promo.Quantity)
}

func (h *promoHandler) ApplyPromo(c *gin.Context) {
	claims, ok := auth.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": constants.ErrUnauthorized})
		return
	}
	if claims.Subject != c.Param("userId") {
		isPrivileged := false
		for _, role := range claims.Roles {
			isPrivileged = isPrivileged || role == "staff" || role == "admin"
		}
		if !isPrivileged {
			c.JSON(http.StatusForbidden, gin.H{"error": constants.ErrForbidden})
			return
		}
	}
	code := c.Param("code")
	userId := c.Param("userId")

	var request struct {
		TotalAmount int64 `json:"total_amount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	discount, err := h.usecase.ApplyPromo(c.Request.Context(), code, userId, request.TotalAmount)
	if err != nil {
		if err.Error() == constants.ErrPromoNotFound ||
			err.Error() == constants.ErrPromoNotActive ||
			err.Error() == constants.ErrPromoExpired ||
			err.Error() == constants.ErrPromoLimitReached ||
			err.Error() == constants.ErrPromoMinPurchase ||
			err.Error() == constants.ErrPromoMaxUsagePerUser {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to apply promo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"discount": discount,
		"message":  "promo applied successfully",
	})
}
