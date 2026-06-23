package handler

import (
	"net/http"
	"strconv"
	"strings"

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
	if doc.RepositoryID != nil && h.sync != nil {
		if gitItems, gErr := h.sync.FileHistory(c.Request.Context(), *doc.RepositoryID, doc.Path, 40); gErr == nil {
			versions = service.MergeVersionHistory(versions, gitItems)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": versions})
}

func (h *Handler) DiffDocumentHistory(c *gin.Context) {
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

	fromSHA := strings.TrimSpace(c.Query("from_sha"))
	toSHA := strings.TrimSpace(c.Query("to_sha"))
	fromStr := strings.TrimSpace(c.Query("from"))
	toStr := strings.TrimSpace(c.Query("to"))

	var diff *service.VersionDiff
	switch {
	case fromSHA != "" && toSHA != "":
		diff, err = h.docs.DiffGitVersions(c.Request.Context(), h.sync, doc, fromSHA, toSHA)
	case fromStr != "" && toStr != "":
		fromVer, pErr := strconv.Atoi(fromStr)
		toVer, tErr := strconv.Atoi(toStr)
		if pErr != nil || tErr != nil || fromVer < 1 || toVer < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}
		diff, err = h.docs.DiffVersions(c.Request.Context(), doc.ID, fromVer, toVer)
	case fromSHA != "" && toStr != "":
		toVer, tErr := strconv.Atoi(toStr)
		if tErr != nil || toVer < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}
		diff, err = h.docs.DiffMixedVersionsReverse(c.Request.Context(), h.sync, doc, fromSHA, toVer)
	case fromStr != "" && toSHA != "":
		fromVer, pErr := strconv.Atoi(fromStr)
		if pErr != nil || fromVer < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}
		diff, err = h.docs.DiffMixedVersions(c.Request.Context(), h.sync, doc, fromVer, toSHA)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "specify from/to or from_sha/to_sha"})
		return
	}
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

func (h *Handler) GetDocumentHistoryContent(c *gin.Context) {
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

	sha := strings.TrimSpace(c.Query("sha"))
	if sha != "" {
		if doc.RepositoryID == nil || h.sync == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "document is not linked to git"})
			return
		}
		version, gErr := h.docs.GetGitVersion(c.Request.Context(), h.sync, doc, sha)
		if gErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gErr.Error()})
			return
		}
		c.JSON(http.StatusOK, version)
		return
	}

	versionStr := strings.TrimSpace(c.Query("version"))
	versionNum, pErr := strconv.Atoi(versionStr)
	if pErr != nil || versionNum < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "specify sha or version"})
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
	h.logAudit(c, "repo.sync", "repository", repoID)
	c.Data(code, "application/json", body)
}
