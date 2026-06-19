package service

import (
	"context"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type PageACLService struct {
	db *gorm.DB
}

func NewPageACLService(db *gorm.DB) *PageACLService {
	return &PageACLService{db: db}
}

type PageACLRuleInput struct {
	PathPrefix  string `json:"path_prefix"`
	SubjectType string `json:"subject_type" binding:"required"`
	SubjectID   string `json:"subject_id" binding:"required"`
	Role        string `json:"role" binding:"required"`
}

func normalizePathPrefix(p string) string {
	p = strings.TrimSpace(strings.ReplaceAll(p, "\\", "/"))
	p = strings.Trim(p, "/")
	return p
}

func pathUnderPrefix(prefix, docPath string) bool {
	prefix = normalizePathPrefix(prefix)
	docPath = normalizePathPrefix(docPath)
	if prefix == "" {
		return true
	}
	return docPath == prefix || strings.HasPrefix(docPath, prefix+"/")
}

func prefixLen(prefix string) int {
	return len(normalizePathPrefix(prefix))
}

func (s *PageACLService) ListRules(ctx context.Context, spaceID string) ([]models.PageACLRule, error) {
	var rules []models.PageACLRule
	err := s.db.WithContext(ctx).Where("space_id = ?", spaceID).Order("path_prefix ASC").Find(&rules).Error
	return rules, err
}

func (s *PageACLService) CreateRule(ctx context.Context, spaceID string, input PageACLRuleInput) (*models.PageACLRule, error) {
	role := strings.ToLower(strings.TrimSpace(input.Role))
	if role == "" {
		role = "viewer"
	}
	rule := models.PageACLRule{
		SpaceID: spaceID, PathPrefix: normalizePathPrefix(input.PathPrefix),
		SubjectType: strings.ToLower(input.SubjectType),
		SubjectID: input.SubjectID, Role: role,
	}
	if err := s.db.WithContext(ctx).Create(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *PageACLService) DeleteRule(ctx context.Context, ruleID string) error {
	return s.db.WithContext(ctx).Delete(&models.PageACLRule{}, "id = ?", ruleID).Error
}

func (s *PageACLService) GetRule(ctx context.Context, ruleID string) (*models.PageACLRule, error) {
	var rule models.PageACLRule
	if err := s.db.WithContext(ctx).First(&rule, "id = ?", ruleID).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *PageACLService) userGroupIDs(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, nil
	}
	var ids []string
	err := s.db.WithContext(ctx).Table("group_members").Where("user_id = ?", userID).Pluck("group_id", &ids).Error
	return ids, err
}

// EffectiveDocumentRole returns the effective role for a document path.
// Page ACL rules on the longest matching prefix override space role when the user/group matches.
func (s *PageACLService) EffectiveDocumentRole(ctx context.Context, spaceID, docPath, userID, spaceRole string, isSuperAdmin bool) (string, bool) {
	if isSuperAdmin {
		return SpaceRoleAdmin, true
	}
	var rules []models.PageACLRule
	if err := s.db.WithContext(ctx).Where("space_id = ?", spaceID).Find(&rules).Error; err != nil || len(rules) == 0 {
		return spaceRole, spaceRole != "none" && HasSpaceRole(spaceRole, SpaceRoleViewer)
	}
	groupIDs, _ := s.userGroupIDs(ctx, userID)
	bestPrefix := -1
	var matchedRoles []string
	for _, r := range rules {
		if !pathUnderPrefix(r.PathPrefix, docPath) {
			continue
		}
		pl := prefixLen(r.PathPrefix)
		if pl < bestPrefix {
			continue
		}
		subjectMatch := (r.SubjectType == "user" && r.SubjectID == userID) ||
			(r.SubjectType == "group" && contains(groupIDs, r.SubjectID))
		if !subjectMatch {
			continue
		}
		if pl > bestPrefix {
			bestPrefix = pl
			matchedRoles = []string{r.Role}
		} else if pl == bestPrefix {
			matchedRoles = append(matchedRoles, r.Role)
		}
	}
	if len(matchedRoles) == 0 {
		return spaceRole, HasSpaceRole(spaceRole, SpaceRoleViewer)
	}
	for _, r := range matchedRoles {
		if r == "none" {
			return "none", false
		}
	}
	pageRole := MaxSpaceRole(matchedRoles...)
	effective := MaxSpaceRole(spaceRole, pageRole)
	return effective, HasSpaceRole(effective, SpaceRoleViewer)
}

func (s *PageACLService) FilterDocuments(ctx context.Context, userID, spaceRole string, isSuperAdmin bool, docs []models.Document) []models.Document {
	if isSuperAdmin {
		return docs
	}
	out := make([]models.Document, 0, len(docs))
	for _, d := range docs {
		if _, ok := s.EffectiveDocumentRole(ctx, d.SpaceID, d.Path, userID, spaceRole, isSuperAdmin); ok {
			out = append(out, d)
		}
	}
	return out
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
