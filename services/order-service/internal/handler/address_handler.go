package handler

import (
	"kaffein/order-service/internal/auth"
	"kaffein/order-service/internal/dto"
	"kaffein/order-service/internal/interfaces/IUseCase"
	"kaffein/order-service/utils/constants"
	"kaffein/order-service/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type addressHandler struct {
	usecase IUseCase.AddressUseCase
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewAddressHandler(usecase IUseCase.AddressUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *auth.Service) *addressHandler {
	h := &addressHandler{
		usecase: usecase,
		logger:  logger,
		v:       validator.Get(),
	}

	addresses := engine.Group("/addresses")
	{
		addresses.Use(authorization.Authenticate())
		addresses.GET("", authorization.Authorize("addresses", "read"), h.List)
		addresses.POST("", authorization.Authorize("addresses", "create"), h.Create)
		addresses.GET("/:id", authorization.Authorize("addresses", "read"), h.ReadById)
		addresses.PUT("/:id", authorization.Authorize("addresses", "update"), h.Update)
		addresses.DELETE("/:id", authorization.Authorize("addresses", "delete"), h.Delete)
	}

	return h
}

func (h *addressHandler) List(c *gin.Context) {
	var params dto.ListAddressParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !auth.RequireOwner(c, params.UserId) {
		return
	}

	if errors := h.v.ValidateStruct(&params); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	addresses, err := h.usecase.List(c.Request.Context(), params)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list addresses")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": addresses})
}

func (h *addressHandler) Create(c *gin.Context) {
	var args dto.CreateAddressDto
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

	address, err := h.usecase.Create(c.Request.Context(), args)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create address")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": address, "message": "address created"})
}

func (h *addressHandler) Update(c *gin.Context) {
	id := c.Param("id")
	address, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read address by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrAddressNotFound})
		return
	}
	if !auth.RequireOwner(c, address.UserId) {
		return
	}

	var args dto.UpdateAddressDto
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	args.UserId = address.UserId

	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	if err := h.usecase.Update(c.Request.Context(), id, args); err != nil {
		if err.Error() == constants.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to update address")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "address updated"})
}

func (h *addressHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	address, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read address by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}
	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrAddressNotFound})
		return
	}
	if !auth.RequireOwner(c, address.UserId) {
		return
	}

	if err := h.usecase.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == constants.ErrAddressNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete address")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "address deleted"})
}

func (h *addressHandler) ReadById(c *gin.Context) {
	id := c.Param("id")

	address, err := h.usecase.ReadById(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to read address by id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	if address == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrAddressNotFound})
		return
	}
	if !auth.RequireOwner(c, address.UserId) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": address})
}
