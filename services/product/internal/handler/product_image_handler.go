package handler

import (
	"fmt"
	"kaffein/product-service/internal/interfaces/IUseCase"
	"kaffein/product-service/pkg/minio/storage"
	"kaffein/product-service/utils/constants"
	"kaffein/product-service/internal/dto"
	"kaffein/product-service/utils/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type productImageHandler struct {
	usecase IUseCase.ProductImageUseCase
	storage *storage.Storage
	logger  zerolog.Logger
	v       *validator.Validator
}

func NewProductImageHandler(usecase IUseCase.ProductImageUseCase, storage *storage.Storage, logger zerolog.Logger) *productImageHandler {
	return &productImageHandler{
		usecase: usecase,
		storage: storage,
		logger:  logger,
		v:       validator.Get(),
	}
}

func (h *productImageHandler) Create(c *gin.Context) {
	productId := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrFileRequired})
		return
	}

	var args dto.UploadProductImageDto
	if err := c.ShouldBind(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors := h.v.ValidateStruct(args); errors != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}
	sortOrder := args.SortOrder

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrFailedOpenFile})
		return
	}
	defer src.Close()

	fileName := fmt.Sprintf("%s/%s", productId, uuid.New().String())
	_, err = h.storage.Upload(c.Request.Context(), fileName, src, file.Size, file.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to upload file to minio")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrFailedUploadFile})
		return
	}

	if err := h.usecase.Create(c.Request.Context(), productId, fileName, int(sortOrder)); err != nil {
		h.logger.Error().Err(err).Msg("failed to create product image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "image uploaded"})
}

func (h *productImageHandler) Delete(c *gin.Context) {
	productId := c.Param("id")
	id := c.Param("image_id")

	if err := h.usecase.Delete(c.Request.Context(), productId, id); err != nil {
		if err.Error() == constants.ErrImageNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		h.logger.Error().Err(err).Msg("failed to delete product image")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "image deleted"})
}

func (h *productImageHandler) ListByProductId(c *gin.Context) {
	productId := c.Param("id")

	images, err := h.usecase.ListByProductId(c.Request.Context(), productId)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list product images")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": images})
}
