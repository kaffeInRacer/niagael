package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"kaffein/order-service/internal/auth"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IUseCase"
	"kaffein/order-service/utils/constants"
	"kaffein/order-service/utils/validator"
)

type cartHandler struct {
	usecase IUseCase.CartUseCase
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewCartHandler(usecase IUseCase.CartUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *cartHandler {
	h := &cartHandler{usecase: usecase, logger: logger, v: validator.Get()}

	carts := engine.Group("/carts")
	{
		carts.Use(authorization.Authenticate())
		carts.GET("/user/:userId", authorization.Authorize("carts", "read"), h.ReadByUserId)
		carts.POST("/items", authorization.Authorize("carts", "create"), h.AddItem)
		carts.PUT("/items/:id", authorization.Authorize("carts", "update"), h.UpdateItem)
		carts.DELETE("/items/:id", authorization.Authorize("carts", "delete"), h.DeleteItem)
		carts.DELETE("/user/:userId", authorization.Authorize("carts", "delete"), h.Clear)
	}

	return h
}

func (h *cartHandler) ReadByUserId(c *gin.Context) {
	userId := c.Param("userId")
	if !auth.RequireOwner(c, userId) {
		return
	}
	if errors := h.v.ValidateStruct(dto.CartUserParams{UserId: userId}); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	cart, err := h.usecase.ReadByUserId(c.Request.Context(), userId)
	if err != nil {
		h.respondError(c, err, "failed to read cart")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cart})
}

func (h *cartHandler) AddItem(c *gin.Context) {
	var args dto.AddCartItemDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !auth.RequireOwner(c, args.UserId) {
		return
	}
	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	cart, err := h.usecase.AddItem(c.Request.Context(), args)
	if err != nil {
		h.respondError(c, err, "failed to add cart item")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": cart, "message": "cart item added"})
}

func (h *cartHandler) UpdateItem(c *gin.Context) {
	var args dto.UpdateCartItemDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !auth.RequireOwner(c, args.UserId) {
		return
	}
	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	cart, err := h.usecase.UpdateItem(c.Request.Context(), c.Param("id"), args)
	if err != nil {
		h.respondError(c, err, "failed to update cart item")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": cart, "message": "cart item updated"})
}

func (h *cartHandler) DeleteItem(c *gin.Context) {
	var params dto.CartUserParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !auth.RequireOwner(c, params.UserId) {
		return
	}
	if errors := h.v.ValidateStruct(params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.DeleteItem(c.Request.Context(), c.Param("id"), params.UserId); err != nil {
		h.respondError(c, err, "failed to delete cart item")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cart item deleted"})
}

func (h *cartHandler) Clear(c *gin.Context) {
	userId := c.Param("userId")
	if !auth.RequireOwner(c, userId) {
		return
	}
	if errors := h.v.ValidateStruct(dto.CartUserParams{UserId: userId}); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	if err := h.usecase.Clear(c.Request.Context(), userId); err != nil {
		h.respondError(c, err, "failed to clear cart")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cart cleared"})
}

func (h *cartHandler) respondError(c *gin.Context, err error, logMessage string) {
	switch err.Error() {
	case constants.ErrProductNotFound, constants.ErrVariantNotFound, constants.ErrCartItemNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case constants.ErrCartQuantityLimit, constants.ErrInsufficientStock:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		h.logger.Error().Err(err).Msg(logMessage)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
