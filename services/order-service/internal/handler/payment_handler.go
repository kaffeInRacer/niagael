package handler

import (
	"errors"
	"kaffein/order-service/internal/auth"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IUseCase"
	"kaffein/order-service/utils/constants"
	"kaffein/order-service/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type paymentHandler struct {
	usecase      IUseCase.PaymentUseCase
	orderUseCase IUseCase.OrderUseCase
	logger       zerolog.Logger
	v            *validator.Validator
}

func NewPaymentHandler(usecase IUseCase.PaymentUseCase, orderUseCase IUseCase.OrderUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *paymentHandler {
	h := &paymentHandler{
		usecase:      usecase,
		orderUseCase: orderUseCase,
		logger:       logger,
		v:            validator.Get(),
	}

	payments := engine.Group("/payments")
	{
		payments.Use(authorization.Authenticate())
		payments.POST("/:orderId", authorization.Authorize("payments", "create"), h.Create)
	}

	webhooks := engine.Group("/webhooks")
	{
		webhooks.POST("/midtrans", h.Callback)
	}

	return h
}

func (h *paymentHandler) Create(c *gin.Context) {
	orderId := c.Param("orderId")
	order, err := h.orderUseCase.ReadById(c.Request.Context(), orderId)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read order by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrOrderNotFound})
		return
	}
	if !auth.RequireOwner(c, order.UserId) {
		return
	}

	var args dto.CreatePaymentDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	args.OrderId = orderId

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	redirectUrl, err := h.usecase.Create(c.Request.Context(), args)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create payment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"redirect_url": redirectUrl,
		"message":      "payment created successfully",
	})
}

func (h *paymentHandler) Callback(c *gin.Context) {
	var args dto.MidtransCallbackDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.usecase.Callback(c.Request.Context(), args); err != nil {
		if err.Error() == constants.ErrInvalidPaymentSignature {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrInvalidPaymentAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to process payment callback")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "callback processed successfully"})
}
