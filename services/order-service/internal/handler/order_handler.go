package handler

import (
	"kaffein/order-service/internal/auth"
	"kaffein/order-service/internal/domain"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IUseCase"
	"kaffein/order-service/utils/constants"
	"kaffein/order-service/utils/validator"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type orderHandler struct {
	usecase        IUseCase.OrderUseCase
	addressUseCase IUseCase.AddressUseCase
	logger         zerolog.Logger
	v              *validator.Validator
}

func NewOrderHandler(usecase IUseCase.OrderUseCase, addressUseCase IUseCase.AddressUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *orderHandler {
	h := &orderHandler{
		usecase:        usecase,
		addressUseCase: addressUseCase,
		logger:         logger,
		v:              validator.Get(),
	}

	orders := engine.Group("/orders")
	{
		orders.Use(authorization.Authenticate())
		orders.POST("", authorization.Authorize("orders", "create"), h.Create)
		orders.GET("", authorization.Authorize("orders", "read"), authorization.RequireRoles("staff", "admin"), h.List)
		orders.GET("/user/:userId", authorization.Authorize("orders", "read"), h.ReadByUserId)
		orders.PATCH("/:id/status", authorization.Authorize("orders", "update"), authorization.RequireRoles("staff", "admin"), h.UpdateStatus)
		orders.GET("/:id", authorization.Authorize("orders", "read"), h.ReadById)
	}

	return h
}

func (h *orderHandler) UpdateStatus(c *gin.Context) {
	var args dto.UpdateOrderStatusDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	if err := h.usecase.UpdateStatus(c.Request.Context(), c.Param("id"), args.Status); err != nil {
		if err.Error() == constants.ErrInvalidOrderStatusTransition {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update order status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order status updated"})
}

func (h *orderHandler) Create(c *gin.Context) {
	var args dto.CreateOrderDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	if !auth.RequireOwner(c, args.UserId) {
		return
	}
	address, err := h.addressUseCase.ReadById(c.Request.Context(), args.AddressId)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read order address")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrAddressNotFound})
		return
	}
	if address.UserId != args.UserId {
		c.JSON(http.StatusForbidden, gin.H{"error": constants.ErrForbidden})
		return
	}

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	order, err := h.usecase.Create(c.Request.Context(), args)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create order")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func (h *orderHandler) List(c *gin.Context) {
	var params dto.ListOrderParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	params.PageOffset = (params.Page - 1) * params.PageSize

	orders, count, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list orders")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  orders,
		"count": count,
	})
}

func (h *orderHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	order, items, err := h.usecase.ReadByIdWithItems(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read order by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrOrderNotFound})
		return
	}
	if !auth.RequireOwner(c, order.UserId) {
		return
	}

	if items == nil {
		items = []domain.OrderItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  order,
		"items": items,
	})
}

func (h *orderHandler) ReadByUserId(c *gin.Context) {
	userId := c.Param("userId")
	if !auth.RequireOwner(c, userId) {
		return
	}

	var params struct {
		Before   string `form:"before" validate:"omitempty"`
		PageSize int32  `form:"page_size" validate:"clamp=5 50"`
	}
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	orders, err := h.usecase.ReadByUserId(c.Request.Context(), userId, params.Before, params.PageSize)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read orders by user id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if orders == nil {
		orders = []domain.Order{}
	}

	var nextCursor string
	if len(orders) == int(params.PageSize) {
		nextCursor = orders[len(orders)-1].CreatedAt.Format(time.RFC3339Nano)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        orders,
		"next_cursor": nextCursor,
	})
}
