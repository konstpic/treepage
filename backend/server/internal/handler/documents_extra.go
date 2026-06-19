package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"gorm.io/gorm"
)

func (h *Handler) ListDocumentVersions(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	versions, err := h.docs.ListVersions(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": versions})
}

func (h *Handler) GetDocumentVersion(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	versionNum, err := strconv.Atoi(c.Param("version"))
	if err != nil || versionNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
		return
	}
	version, err := h.docs.GetVersion(c.Request.Context(), doc.ID, versionNum)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, version)
}

func (h *Handler) DiffDocumentVersions(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceAccess(c, space) {
		return
	}
	toVersion, err := strconv.Atoi(c.Param("version"))
	if err != nil || toVersion < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
		return
	}
	fromVersion := toVersion - 1
	if fromParam := c.Query("from"); fromParam != "" {
		fromVersion, err = strconv.Atoi(fromParam)
		if err != nil || fromVersion < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from version"})
			return
		}
	}
	diff, err := h.docs.DiffVersions(c.Request.Context(), doc.ID, fromVersion, toVersion)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, diff)
}

func (h *Handler) TriggerSpaceSync(c *gin.Context) {
	space, err := h.spaces.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	repoID := c.Param("repoId")
	repos, err := h.repos.ListBySpace(c.Request.Context(), space.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	found := false
	for _, r := range repos {
		if r.ID == repoID {
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "repository not found in this space"})
		return
	}
	if h.sync == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sync service not configured"})
		return
	}
	code, body, err := h.sync.TriggerSync(c.Request.Context(), repoID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Data(code, "application/json", body)
}
