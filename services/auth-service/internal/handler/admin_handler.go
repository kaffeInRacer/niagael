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
	routes.GET("", authorization.Authorize("users", "read"), h.listUsers)
	routes.POST("", authorization.Authorize("users", "create"), h.createUser)
	routes.PATCH("/:id/role", authorization.Authorize("users", "update"), h.changeUserRole)
	routes.PATCH("/:id/status", authorization.Authorize("users", "update"), h.changeUserStatus)
	routes.PUT("/:id/email", authorization.Authorize("users", "update"), h.changeUserEmail)
	routes.DELETE("/:id", authorization.Authorize("users", "delete"), h.deleteUser)
	return h
}

func (h *adminHandler) listUsers(c *gin.Context) {
	page, size := parsePositiveInt(c.Query("page"), 1), parsePositiveInt(c.Query("page_size"), 20)
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

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidUserID})
		return uuid.Nil, false
	}
	return id, true
}

func (h *adminHandler) changeUserRole(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	var req dto.RoleRequest
	if !bindJSON(c, &req) {
		return
	}
	user, err := h.usecase.ChangeRole(c.Request.Context(), id, req.Role)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *adminHandler) createUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	user, err := h.usecase.CreateUser(c.Request.Context(), req)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *adminHandler) changeUserEmail(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	user, err := h.usecase.UpdateUserEmail(c.Request.Context(), id, req.Email)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *adminHandler) deleteUser(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	if err := h.usecase.DeleteUser(c.Request.Context(), id); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *adminHandler) changeUserStatus(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	var req dto.StatusRequest
	if !bindJSON(c, &req) {
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
