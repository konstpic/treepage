package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/service"
)

func (h *Handler) requireSpaceAccess(c *gin.Context, space *models.Space) bool {
	userID := c.GetString("userID")
	roles := getRoles(c)
	ok, err := h.spaces.CanAccess(c.Request.Context(), space, userID, service.HasRole(roles, "super_admin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return false
	}
	return true
}

func (h *Handler) getEffectiveSpaceRole(c *gin.Context, space *models.Space) (string, error) {
	return h.spaces.EffectiveRole(
		c.Request.Context(),
		space.ID,
		c.GetString("userID"),
		getRoles(c),
	)
}

func (h *Handler) requireSpaceRole(c *gin.Context, space *models.Space, minRole string) bool {
	if !h.requireSpaceAccess(c, space) {
		return false
	}
	if service.HasRole(getRoles(c), "super_admin") {
		return true
	}
	effective, err := h.getEffectiveSpaceRole(c, space)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	if !service.HasSpaceRole(effective, minRole) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}
	return true
}

func (h *Handler) canEditInSpace(c *gin.Context, space *models.Space) (bool, error) {
	if service.HasRole(getRoles(c), "super_admin") {
		return true, nil
	}
	effective, err := h.getEffectiveSpaceRole(c, space)
	if err != nil {
		return false, err
	}
	return service.CanEditInSpace(effective, getRoles(c)), nil
}

func (h *Handler) canUseLLMInSpace(c *gin.Context, space *models.Space) bool {
	if service.HasRole(getRoles(c), "super_admin") {
		return true
	}
	effective, err := h.getEffectiveSpaceRole(c, space)
	if err != nil {
		return false
	}
	return service.CanUseLLMInSpace(effective, getRoles(c))
}

func getRoles(c *gin.Context) []string {
	v, _ := c.Get("roles")
	roles, _ := v.([]string)
	return roles
}
