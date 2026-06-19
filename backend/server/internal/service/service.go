package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

type SpaceService struct {
	db *gorm.DB
}

func NewSpaceService(db *gorm.DB) *SpaceService {
	return &SpaceService{db: db}
}

func (s *SpaceService) List(ctx context.Context, userID string, isSuperAdmin bool) ([]models.Space, error) {
	var spaces []models.Space
	q := s.db.WithContext(ctx).Model(&models.Space{})
	if userID == "" {
		q = q.Where("is_public = ?", true)
	} else if !isSuperAdmin {
		q = q.Where(`spaces.is_public = ? OR spaces.id IN (
			SELECT space_id FROM space_members WHERE user_id = ?
			UNION
			SELECT sg.space_id FROM space_groups sg
			JOIN group_members gm ON gm.group_id = sg.group_id
			WHERE gm.user_id = ?
		)`, true, userID, userID)
	}
	if err := q.Order("name ASC").Find(&spaces).Error; err != nil {
		return nil, err
	}
	return spaces, nil
}

func (s *SpaceService) CanAccess(ctx context.Context, space *models.Space, userID string, isSuperAdmin bool) (bool, error) {
	if space.IsPublic {
		return true, nil
	}
	if userID == "" {
		return false, nil
	}
	if isSuperAdmin {
		return true, nil
	}
	var directCount int64
	err := s.db.WithContext(ctx).Model(&models.SpaceMember{}).
		Where("space_id = ? AND user_id = ?", space.ID, userID).
		Count(&directCount).Error
	if err != nil {
		return false, err
	}
	if directCount > 0 {
		return true, nil
	}
	var groupCount int64
	err = s.db.WithContext(ctx).
		Table("space_groups sg").
		Joins("JOIN group_members gm ON gm.group_id = sg.group_id").
		Where("sg.space_id = ? AND gm.user_id = ?", space.ID, userID).
		Count(&groupCount).Error
	return groupCount > 0, err
}

// EffectiveRole returns the highest space-scoped role for a user:
// max(global role floor, direct space membership, group assignments).
func (s *SpaceService) EffectiveRole(ctx context.Context, spaceID, userID string, globalRoles []string) (string, error) {
	if userID == "" {
		return "", nil
	}
	if HasRole(globalRoles, "super_admin") {
		return SpaceRoleAdmin, nil
	}

	roles := []string{GlobalRolesAsSpaceRole(globalRoles)}

	var memberRole string
	err := s.db.WithContext(ctx).
		Table("space_members sm").
		Select("r.name").
		Joins("JOIN roles r ON r.id = sm.role_id").
		Where("sm.space_id = ? AND sm.user_id = ?", spaceID, userID).
		Limit(1).
		Scan(&memberRole).Error
	if err != nil {
		return "", err
	}
	if memberRole != "" {
		roles = append(roles, memberRole)
	}

	var groupRoles []string
	err = s.db.WithContext(ctx).
		Table("space_groups sg").
		Select("r.name").
		Joins("JOIN roles r ON r.id = sg.role_id").
		Joins("JOIN group_members gm ON gm.group_id = sg.group_id").
		Where("sg.space_id = ? AND gm.user_id = ?", spaceID, userID).
		Scan(&groupRoles).Error
	if err != nil {
		return "", err
	}
	roles = append(roles, groupRoles...)

	return MaxSpaceRole(roles...), nil
}

// AccessibleSpaceIDs returns space IDs visible to the caller.
// Super admins get nil (no filter). An empty slice means no accessible spaces.
func (s *SpaceService) AccessibleSpaceIDs(ctx context.Context, userID string, isSuperAdmin bool) ([]string, error) {
	if isSuperAdmin {
		return nil, nil
	}
	spaces, err := s.List(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(spaces))
	for i, sp := range spaces {
		ids[i] = sp.ID
	}
	return ids, nil
}

type UpdateSpaceInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
}

func (s *SpaceService) Update(ctx context.Context, id string, input UpdateSpaceInput) (*models.Space, error) {
	var space models.Space
	if err := s.db.WithContext(ctx).First(&space, "id = ?", id).Error; err != nil {
		return nil, err
	}
	if input.Name != nil {
		space.Name = *input.Name
	}
	if input.Description != nil {
		space.Description = *input.Description
	}
	if input.IsPublic != nil {
		space.IsPublic = *input.IsPublic
	}
	if err := s.db.WithContext(ctx).Save(&space).Error; err != nil {
		return nil, err
	}
	return &space, nil
}

func (s *SpaceService) GetBySlug(ctx context.Context, slug string) (*models.Space, error) {
	var space models.Space
	if err := s.db.WithContext(ctx).Where("slug = ?", slug).First(&space).Error; err != nil {
		return nil, err
	}
	return &space, nil
}

type CreateSpaceInput struct {
	Slug        string `json:"slug" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

func (s *SpaceService) Create(ctx context.Context, input CreateSpaceInput, userID string) (*models.Space, error) {
	slug := slugify(input.Slug)
	space := models.Space{
		Slug: slug, Name: input.Name, Description: input.Description,
		IsPublic: input.IsPublic, CreatedBy: &userID,
	}
	if err := s.db.WithContext(ctx).Create(&space).Error; err != nil {
		return nil, err
	}
	var adminRole models.Role
	if err := s.db.WithContext(ctx).Where("name = ?", "admin").First(&adminRole).Error; err == nil {
		s.db.WithContext(ctx).Create(&models.SpaceMember{
			SpaceID: space.ID, UserID: userID, RoleID: adminRole.ID,
		})
	}
	return &space, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.Trim(slugRe.ReplaceAllString(s, "-"), "-")
}

type DocumentService struct {
	db *gorm.DB
}

func NewDocumentService(db *gorm.DB) *DocumentService {
	return &DocumentService{db: db}
}

func (d *DocumentService) ListFullBySpace(ctx context.Context, spaceID string) ([]models.Document, error) {
	var docs []models.Document
	err := d.db.WithContext(ctx).
		Where("space_id = ? AND is_published = ?", spaceID, true).
		Order("path ASC").
		Find(&docs).Error
	return docs, err
}

func (d *DocumentService) ListBySpace(ctx context.Context, spaceID string) ([]models.Document, error) {
	var docs []models.Document
	err := d.db.WithContext(ctx).
		Select("id", "space_id", "repository_id", "slug", "title", "path", "tags", "updated_at", "is_published").
		Where("space_id = ? AND is_published = ?", spaceID, true).
		Order("path ASC").
		Find(&docs).Error
	return docs, err
}

func (d *DocumentService) ListBySpaceIncludingDrafts(ctx context.Context, spaceID string) ([]models.Document, error) {
	var docs []models.Document
	err := d.db.WithContext(ctx).
		Select("id", "space_id", "repository_id", "slug", "title", "path", "tags", "updated_at", "is_published").
		Where("space_id = ?", spaceID).
		Order("path ASC").
		Find(&docs).Error
	return docs, err
}

func (d *DocumentService) GetBySlug(ctx context.Context, spaceID, slug string) (*models.Document, error) {
	return d.GetBySlugVisible(ctx, spaceID, slug, false)
}

func (d *DocumentService) GetBySlugVisible(ctx context.Context, spaceID, slug string, includeDrafts bool) (*models.Document, error) {
	var doc models.Document
	if includeDrafts {
		err := d.db.WithContext(ctx).
			Where("space_id = ? AND slug = ? AND is_published = ?", spaceID, slug, true).
			First(&doc).Error
		if err == nil {
			return &doc, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		err = d.db.WithContext(ctx).
			Where("space_id = ? AND slug = ?", spaceID, slug).
			First(&doc).Error
		if err != nil {
			return nil, err
		}
		return &doc, nil
	}
	err := d.db.WithContext(ctx).
		Where("space_id = ? AND slug = ? AND is_published = ?", spaceID, slug, true).
		First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

type CreateDocumentInput struct {
	Title       string   `json:"title" binding:"required"`
	Path        string   `json:"path" binding:"required"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags"`
	IsPublished *bool    `json:"is_published"`
}

func (d *DocumentService) Create(ctx context.Context, spaceID, userID string, input CreateDocumentInput) (*models.Document, error) {
	slug := slugify(strings.TrimSuffix(input.Path, ".md"))
	if slug == "" {
		slug = slugify(input.Title)
	}
	isPublished := true
	if input.IsPublished != nil {
		isPublished = *input.IsPublished
	}
	doc := models.Document{
		SpaceID: spaceID, Slug: slug, Title: input.Title,
		Path: input.Path, Content: input.Content, Tags: input.Tags,
		AuthorID: &userID, IsPublished: isPublished,
	}
	if err := d.db.WithContext(ctx).Create(&doc).Error; err != nil {
		return nil, err
	}
	d.db.WithContext(ctx).Create(&models.DocumentVersion{
		DocumentID: doc.ID, VersionNumber: 1,
		Title: doc.Title, Content: doc.Content, AuthorID: &userID,
	})
	return &doc, nil
}

func (d *DocumentService) GetByID(ctx context.Context, id string) (*models.Document, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (d *DocumentService) Update(ctx context.Context, docID, userID string, content, title string) (*models.Document, error) {
	return d.UpdateFull(ctx, docID, userID, content, title, nil)
}

func (d *DocumentService) UpdateFull(ctx context.Context, docID, userID string, content, title string, isPublished *bool) (*models.Document, error) {
	var doc models.Document
	if err := d.db.WithContext(ctx).First(&doc, "id = ?", docID).Error; err != nil {
		return nil, err
	}
	if title != "" {
		doc.Title = title
	}
	doc.Content = content
	if isPublished != nil {
		doc.IsPublished = *isPublished
	}
	if doc.RepositoryID != nil {
		doc.HasPendingChanges = true
	}
	if err := d.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return nil, err
	}
	var maxVersion int
	d.db.WithContext(ctx).Model(&models.DocumentVersion{}).
		Where("document_id = ?", docID).
		Select("COALESCE(MAX(version_number), 0)").Scan(&maxVersion)
	d.db.WithContext(ctx).Create(&models.DocumentVersion{
		DocumentID: doc.ID, VersionNumber: maxVersion + 1,
		Title: doc.Title, Content: doc.Content, AuthorID: &userID,
	})
	return &doc, nil
}

type RepositoryService struct {
	db *gorm.DB
}

func NewRepositoryService(db *gorm.DB) *RepositoryService {
	return &RepositoryService{db: db}
}

func (r *RepositoryService) ListBySpace(ctx context.Context, spaceID string) ([]models.Repository, error) {
	var repos []models.Repository
	err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Find(&repos).Error
	return repos, err
}

type CreateRepositoryInput struct {
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required"`
	Branch   string `json:"branch"`
	Provider string `json:"provider"`
	DocsPath string `json:"docs_path"`
}

func (r *RepositoryService) Create(ctx context.Context, spaceID string, input CreateRepositoryInput) (*models.Repository, error) {
	branch := input.Branch
	if branch == "" {
		branch = "main"
	}
	provider := input.Provider
	if provider == "" {
		provider = "generic"
	}
	docsPath := input.DocsPath
	if docsPath == "" {
		docsPath = "docs"
	}
	repo := models.Repository{
		SpaceID: spaceID, Name: input.Name, URL: input.URL,
		Branch: branch, Provider: provider, DocsPath: docsPath,
		Enabled: true,
	}
	if err := r.db.WithContext(ctx).Create(&repo).Error; err != nil {
		return nil, err
	}
	return &repo, nil
}

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (a *AuditService) Log(ctx context.Context, userID, action, resourceType, resourceID, ip, ua string) {
	entry := models.AuditLog{
		Action: action, ResourceType: resourceType,
		IPAddress: ip, UserAgent: ua,
	}
	if userID != "" {
		entry.UserID = &userID
	}
	if resourceID != "" {
		entry.ResourceID = &resourceID
	}
	_ = a.db.WithContext(ctx).Create(&entry).Error
}

type AuditLogView struct {
	models.AuditLog
	UserEmail string `json:"user_email,omitempty"`
}

func (a *AuditService) List(ctx context.Context, limit, offset int, action string) ([]AuditLogView, int64, error) {
	q := a.db.WithContext(ctx).Model(&models.AuditLog{})
	if action != "" {
		q = q.Where("action = ?", action)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.AuditLog
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := make([]AuditLogView, 0, len(rows))
	for _, row := range rows {
		view := AuditLogView{AuditLog: row}
		if row.UserID != nil {
			var email string
			a.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", *row.UserID).Pluck("email", &email)
			view.UserEmail = email
		}
		out = append(out, view)
	}
	return out, total, nil
}

func HasRole(roles []string, allowed ...string) bool {
	for _, r := range roles {
		if r == "super_admin" {
			return true
		}
		for _, a := range allowed {
			if r == a {
				return true
			}
		}
	}
	return false
}

// CanUseLLM is true for roles that may trigger LLM features (auto-translate, book generation).
func CanUseLLM(roles []string) bool {
	return HasRole(roles, "super_admin", "admin", "editor")
}

// CanManageBooks is true for roles that may create, generate, or delete books (global fallback).
func CanManageBooks(roles []string) bool {
	return CanUseLLM(roles)
}

var ErrForbidden = errors.New("forbidden")
