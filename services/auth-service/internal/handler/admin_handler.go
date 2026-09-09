package handler

import (
	"net/http"
	"strconv"

	"kaffein/auth-service/internal/dto"
	"kaffein/auth-service/internal/interfaces/IUseCase"
	"kaffein/auth-service/internal/middleware"
	"kaffein/auth-service/utils/constants"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type adminHandler struct {
	usecase IUseCase.AdminUseCase
	logger  zerolog.Logger
}

func NewAdminHandler(usecase IUseCase.AdminUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *middleware.Service) *adminHandler {
	h := &adminHandler{usecase: usecase, logger: logger}

	routes := engine.Group("/admin/users")
	routes.Use(authorization.Authenticate(), authorization.Admin())
	routes.GET("", authorization.Authorize("users", "read"), h.list)
	routes.POST("", authorization.Authorize("users", "create"), h.create)
	routes.PATCH("/:id/role", authorization.Authorize("users", "update"), h.role)
	routes.PATCH("/:id/status", authorization.Authorize("users", "update"), h.status)
	routes.PUT("/:id/email", authorization.Authorize("users", "update"), h.email)
	routes.DELETE("/:id", authorization.Authorize("users", "delete"), h.delete)
	return h
}

func (h *adminHandler) list(c *gin.Context) {
	page, size := positiveInt(c.Query("page"), 1), positiveInt(c.Query("page_size"), 20)
	if size > 100 {
		size = 100
	}
	result, err := h.usecase.ListUsers(c.Request.Context(), page, size)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func positiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidUserID})
		return uuid.Nil, false
	}
	return id, true
}

func (h *adminHandler) role(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.RoleRequest
	if !bind(c, &req) {
		return
	}
	user, err := h.usecase.ChangeRole(c.Request.Context(), id, req.Role)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *adminHandler) create(c *gin.Context) {
	var req dto.CreateUserRequest
	if !bind(c, &req) {
		return
	}
	user, err := h.usecase.CreateUser(c.Request.Context(), req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *adminHandler) email(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if !bind(c, &req) {
		return
	}
	user, err := h.usecase.UpdateUserEmail(c.Request.Context(), id, req.Email)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *adminHandler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.usecase.DeleteUser(c.Request.Context(), id); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *adminHandler) status(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.StatusRequest
	if !bind(c, &req) {
		return
	}
	user, err := h.usecase.ChangeStatus(c.Request.Context(), id, *req.IsActive)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *adminHandler) fail(c *gin.Context, err error) {
	switch err.Error() {
	case constants.ErrUserNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	case constants.ErrInvalidRole:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	case constants.ErrEmailExists:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	h.logger.Error().Err(err).Msg("admin request failed")
	c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrInternalServer})
}
