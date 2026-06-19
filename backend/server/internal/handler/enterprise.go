package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/server/internal/rag"
	"github.com/konstpic/treepage/backend/server/internal/service"
)

func (h *Handler) ListPageACLRules(c *gin.Context) {
	space, err := h.spaces.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleAdmin) {
		return
	}
	rules, err := h.pageACL.ListRules(c.Request.Context(), space.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules})
}

func (h *Handler) CreatePageACLRule(c *gin.Context) {
	space, err := h.spaces.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleAdmin) {
		return
	}
	var input service.PageACLRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.pageACL.CreateRule(c.Request.Context(), space.ID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *Handler) DeletePageACLRule(c *gin.Context) {
	rule, err := h.pageACL.GetRule(c.Request.Context(), c.Param("ruleId"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), rule.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleAdmin) {
		return
	}
	if err := h.pageACL.DeleteRule(c.Request.Context(), c.Param("ruleId")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ListDocumentComments(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireDocumentAccess(c, space, doc) {
		return
	}
	items, err := h.comments.ListThread(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) CreateDocumentComment(c *gin.Context) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil || !h.requireDocumentAccess(c, space, doc) {
		return
	}
	var body struct {
		Body     string  `json:"body" binding:"required"`
		ParentID *string `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	comment, mentions, err := h.comments.Create(c.Request.Context(), doc.ID, c.GetString("userID"), body.Body, body.ParentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rt, rid := "document", doc.ID
	for _, uid := range mentions {
		_, _ = h.notifications.Create(c.Request.Context(), uid, "comment.mention", "You were mentioned", doc.Title, &rt, &rid)
	}
	h.logAudit(c, "comment.create", "document", doc.ID)
	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) DeleteComment(c *gin.Context) {
	comment, err := h.comments.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}
	doc, err := h.docs.GetByID(c.Request.Context(), comment.DocumentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return
	}
	isAdmin := service.HasRole(getRoles(c), "super_admin", "admin") || h.canEditSpaceAdmin(c, space)
	if err := h.comments.Delete(c.Request.Context(), c.Param("id"), c.GetString("userID"), isAdmin); err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) SubmitDocumentReview(c *gin.Context) {
	doc, space, ok := h.loadDocForWorkflow(c)
	if !ok {
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	doc, err := h.docs.SubmitReview(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.notifications.NotifySpaceEditors(c.Request.Context(), space.ID, c.GetString("userID"),
		"workflow.review", "Review requested", doc.Title, "document", doc.ID)
	h.logAudit(c, "document.submit_review", "document", doc.ID)
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) ApproveDocumentReview(c *gin.Context) {
	doc, space, ok := h.loadDocForWorkflow(c)
	if !ok {
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleAdmin) {
		return
	}
	doc, err := h.docs.ApproveReview(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "document.approve", "document", doc.ID)
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) RejectDocumentReview(c *gin.Context) {
	doc, space, ok := h.loadDocForWorkflow(c)
	if !ok {
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleAdmin) {
		return
	}
	doc, err := h.docs.RejectReview(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logAudit(c, "document.reject_review", "document", doc.ID)
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) PublishDocumentWorkflow(c *gin.Context) {
	doc, space, ok := h.loadDocForWorkflow(c)
	if !ok {
		return
	}
	if !h.requireSpaceRole(c, space, service.SpaceRoleEditor) {
		return
	}
	doc, err := h.docs.PublishWorkflow(c.Request.Context(), doc.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.notifications.NotifySpaceEditors(c.Request.Context(), space.ID, c.GetString("userID"),
		"document.published", "Document published", doc.Title, "document", doc.ID)
	h.logAudit(c, "document.publish_workflow", "document", doc.ID)
	c.JSON(http.StatusOK, doc)
}

func (h *Handler) RAGAsk(c *gin.Context) {
	var body struct {
		Question string `json:"question" binding:"required"`
		Limit    int    `json:"limit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if c.GetString("userID") == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	answer, err := h.rag.Ask(c.Request.Context(), body.Question, c.GetString("userID"), getRoles(c), body.Limit)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, answer)
}

func (h *Handler) RAGFeedback(c *gin.Context) {
	var body rag.FeedbackInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if c.GetString("userID") == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	if err := h.rag.SubmitFeedback(c.Request.Context(), c.GetString("userID"), body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) AnalyticsOverview(c *gin.Context) {
	overview, err := h.analytics.Overview(c.Request.Context(), 90)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *Handler) loadDocForWorkflow(c *gin.Context) (*models.Document, *models.Space, bool) {
	doc, err := h.docs.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
		return nil, nil, false
	}
	space, err := h.spaces.GetByID(c.Request.Context(), doc.SpaceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "space not found"})
		return nil, nil, false
	}
	if !h.requireDocumentAccess(c, space, doc) {
		return nil, nil, false
	}
	return doc, space, true
}

func (h *Handler) canEditSpaceAdmin(c *gin.Context, space *models.Space) bool {
	if service.HasRole(getRoles(c), "super_admin") {
		return true
	}
	effective, err := h.getEffectiveSpaceRole(c, space)
	return err == nil && service.HasSpaceRole(effective, service.SpaceRoleAdmin)
}
