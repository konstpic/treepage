package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/server/internal/service"
	"github.com/konstpic/treepage/backend/server/internal/syncclient"
)

func (h *Handler) PublishDocument(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	if doc.RepositoryID == nil || *doc.RepositoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "document is not linked to a git repository"})
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

	var body struct {
		Title         string `json:"title"`
		Content       string `json:"content"`
		Branch        string `json:"branch" binding:"required"`
		CommitMessage string `json:"commit_message" binding:"required"`
		PRTitle       string `json:"pr_title"`
		PRBody        string `json:"pr_body"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content := body.Content
	title := body.Title
	if content != "" || title != "" {
		if content == "" {
			content = doc.Content
		}
		if title == "" {
			title = doc.Title
		}
		doc, err = h.docs.Update(c.Request.Context(), doc.ID, c.GetString("userID"), content, title)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		content = doc.Content
	}

	status, result, respBody, err := h.sync.PublishDocument(c.Request.Context(), *doc.RepositoryID, syncclient.PublishInput{
		DocumentID:    doc.ID,
		Path:          doc.Path,
		Content:       content,
		Branch:        body.Branch,
		CommitMessage: body.CommitMessage,
		PRTitle:       body.PRTitle,
		PRBody:        body.PRBody,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	if status >= 300 {
		c.JSON(status, gin.H{"error": string(respBody)})
		return
	}
	h.logAudit(c, "document.publish", "document", doc.ID)
	h.notifications.NotifySpaceEditors(
		c.Request.Context(), space.ID, c.GetString("userID"),
		"document.git_publish", "Changes pushed to Git", doc.Title,
		"document", doc.ID,
	)
	c.JSON(http.StatusOK, result)
}
