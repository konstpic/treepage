package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/service"
)

func (h *Handler) ListFavorites(c *gin.Context) {
	items, err := h.prefs.ListFavorites(c.Request.Context(), c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) AddFavorite(c *gin.Context) {
	docID := c.Param("documentId")
	doc, err := h.docs.GetByID(c.Request.Context(), docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireSpaceAccess(c, space) {
		return
	}
	if err := h.prefs.AddFavorite(c.Request.Context(), c.GetString("userID"), docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

func (h *Handler) RemoveFavorite(c *gin.Context) {
	if err := h.prefs.RemoveFavorite(c.Request.Context(), c.GetString("userID"), c.Param("documentId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ListRecent(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.prefs.ListRecent(c.Request.Context(), c.GetString("userID"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) ListNotifications(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := h.notifications.ListViews(c.Request.Context(), c.GetString("userID"), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) NotificationUnreadCount(c *gin.Context) {
	count, err := h.notifications.UnreadCount(c.Request.Context(), c.GetString("userID"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *Handler) MarkNotificationRead(c *gin.Context) {
	if err := h.notifications.MarkRead(c.Request.Context(), c.GetString("userID"), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) MarkAllNotificationsRead(c *gin.Context) {
	if err := h.notifications.MarkAllRead(c.Request.Context(), c.GetString("userID")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ListDocumentAttachments(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireSpaceAccess(c, space) {
		return
	}
	items, err := h.attachments.ListByDocument(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) UploadDocumentAttachment(c *gin.Context) {
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
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer src.Close()
	mime := file.Header.Get("Content-Type")
	att, err := h.attachments.Upload(c.Request.Context(), doc.ID, c.GetString("userID"), file.Filename, mime, src, file.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "attachment.upload", "document", doc.ID)
	c.JSON(http.StatusCreated, att)
}

func (h *Handler) DownloadAttachment(c *gin.Context) {
	att, f, err := h.attachments.Open(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}
	defer f.Close()
	doc, err := h.docs.GetByID(c.Request.Context(), att.DocumentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireSpaceAccess(c, space) {
		return
	}
	c.Header("Content-Disposition", "inline; filename=\""+att.Filename+"\"")
	c.DataFromReader(http.StatusOK, att.SizeBytes, att.MimeType, f, nil)
}

func (h *Handler) DeleteAttachment(c *gin.Context) {
	att, err := h.attachments.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attachment not found"})
		return
	}
	doc, err := h.docs.GetByID(c.Request.Context(), att.DocumentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	if err := h.attachments.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "attachment.delete", "document", doc.ID)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) PublishDocumentLocal(c *gin.Context) {
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
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	published := true
	doc, err = h.docs.UpdateFull(c.Request.Context(), doc.ID, c.GetString("userID"), doc.Content, doc.Title, &published)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.notifications.NotifySpaceEditors(
		c.Request.Context(), space.ID, c.GetString("userID"),
		"document.published", "Document published", doc.Title,
		"document", doc.ID,
	)
	h.logAudit(c, "document.publish_local", "document", doc.ID)
	c.JSON(http.StatusOK, doc)
}
