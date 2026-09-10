package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"kaffein/auth-service/internal/interfaces/IUseCase"
	"kaffein/auth-service/internal/middleware"
	"kaffein/auth-service/utils/constants"
)

type rbacHandler struct {
	usecase       IUseCase.RBACUseCase
	logger        zerolog.Logger
	authorization *middleware.Service
}

type AddPolicyRequest struct {
	Service  string `json:"service" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

type DeletePolicyRequest struct {
	Service  string `json:"service" binding:"required"`
	Role     string `json:"role" binding:"required"`
	Resource string `json:"resource" binding:"required"`
	Action   string `json:"action" binding:"required"`
}

func NewRBACHandler(usecase IUseCase.RBACUseCase, logger zerolog.Logger, engine *gin.Engine, authorization *middleware.Service) *rbacHandler {
	h := &rbacHandler{usecase: usecase, logger: logger, authorization: authorization}
	routes := engine.Group("/admin/rbac")
	routes.Use(authorization.Authenticate(), authorization.Admin())
	routes.GET("/policies", authorization.Authorize("policies", "read"), h.listPolicies)
	routes.GET("/resources", authorization.Authorize("policies", "read"), h.getResources)
	routes.POST("/policies", authorization.Authorize("policies", "create"), h.addPolicy)
	routes.DELETE("/policies", authorization.Authorize("policies", "delete"), h.deletePolicy)
	return h
}

func (h *rbacHandler) listPolicies(c *gin.Context) {
	policies, err := h.usecase.GetAllPolicies(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list policies")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrFailedListPolicies})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policies, "count": len(policies)})
}

func (h *rbacHandler) getResources(c *gin.Context) {
	resources, err := h.usecase.GetResources(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get resources")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrFailedGetResources})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resources})
}

func (h *rbacHandler) addPolicy(c *gin.Context) {
	var req AddPolicyRequest
	if !bindJSON(c, &req) {
		return
	}

	err := h.usecase.AddPolicy(c.Request.Context(), req.Service, req.Role, req.Resource, req.Action)
	if err != nil {
		if err.Error() == "policy already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": constants.ErrPolicyExists})
			return
		}
		h.logger.Error().Err(err).Msg("failed to add policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrFailedAddPolicy})
		return
	}

	if err := h.authorization.ReloadPolicies(); err != nil {
		h.logger.Error().Err(err).Msg("failed to reload auth policies")
	}
	c.JSON(http.StatusCreated, gin.H{"message": "policy added successfully"})
}

func (h *rbacHandler) deletePolicy(c *gin.Context) {
	var req DeletePolicyRequest
	if !bindJSON(c, &req) {
		return
	}

	err := h.usecase.DeletePolicy(c.Request.Context(), req.Service, req.Role, req.Resource, req.Action)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrPolicyNotFound})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": constants.ErrFailedDeletePolicy})
		return
	}

	if err := h.authorization.ReloadPolicies(); err != nil {
		h.logger.Error().Err(err).Msg("failed to reload auth policies")
	}
	c.JSON(http.StatusOK, gin.H{"message": "policy deleted successfully"})
}
