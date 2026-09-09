package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"kaffein/auth-service/internal/interfaces/IUseCase"
)

type rbacHandler struct {
	usecase IUseCase.RBACUseCase
	logger  zerolog.Logger
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

func NewRBACHandler(usecase IUseCase.RBACUseCase, logger zerolog.Logger, engine *gin.Engine, auth, admin, read, create, remove gin.HandlerFunc) *rbacHandler {
	h := &rbacHandler{usecase: usecase, logger: logger}
	routes := engine.Group("/admin/rbac", auth, admin)
	routes.GET("/policies", read, h.listPolicies)
	routes.GET("/resources", read, h.getResources)
	routes.POST("/policies", create, h.addPolicy)
	routes.DELETE("/policies", remove, h.deletePolicy)
	return h
}

func (h *rbacHandler) listPolicies(c *gin.Context) {
	policies, err := h.usecase.GetAllPolicies(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list policies")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list policies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": policies, "count": len(policies)})
}

func (h *rbacHandler) getResources(c *gin.Context) {
	resources, err := h.usecase.GetResources(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get resources")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get resources"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resources})
}

func (h *rbacHandler) addPolicy(c *gin.Context) {
	var req AddPolicyRequest
	if !bind(c, &req) {
		return
	}

	err := h.usecase.AddPolicy(c.Request.Context(), req.Service, req.Role, req.Resource, req.Action)
	if err != nil {
		if err.Error() == "policy already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "policy already exists"})
			return
		}
		h.logger.Error().Err(err).Msg("failed to add policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add policy"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "policy added successfully"})
}

func (h *rbacHandler) deletePolicy(c *gin.Context) {
	var req DeletePolicyRequest
	if !bind(c, &req) {
		return
	}

	err := h.usecase.DeletePolicy(c.Request.Context(), req.Service, req.Role, req.Resource, req.Action)
	if err != nil {
		if err.Error() == "policy not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "policy not found"})
			return
		}
		h.logger.Error().Err(err).Msg("failed to delete policy")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete policy"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "policy deleted successfully"})
}
