package service

import (
	"context"
	"errors"
	"strings"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

var validSpaceRoles = map[string]bool{
	"viewer": true,
	"editor": true,
	"admin":  true,
}

type GroupService struct {
	db *gorm.DB
}

func NewGroupService(db *gorm.DB) *GroupService {
	return &GroupService{db: db}
}

type GroupRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ExternalID  string `json:"external_id,omitempty"`
	MemberCount int64  `json:"member_count"`
	SpaceCount  int64  `json:"space_count"`
}

type GroupMemberRow struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type SpaceGroupRow struct {
	GroupID     string `json:"group_id"`
	GroupName   string `json:"group_name"`
	Role        string `json:"role"`
	Description string `json:"description,omitempty"`
}

type CreateGroupInput struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ExternalID  string `json:"external_id"`
}

type UpdateGroupInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ExternalID  *string `json:"external_id"`
}

type AssignSpaceGroupInput struct {
	GroupID string `json:"group_id" binding:"required"`
	Role    string `json:"role"`
}

type AddGroupMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
}

func (g *GroupService) List(ctx context.Context) ([]GroupRow, error) {
	var groups []models.Group
	if err := g.db.WithContext(ctx).Order("name ASC").Find(&groups).Error; err != nil {
		return nil, err
	}
	out := make([]GroupRow, 0, len(groups))
	for _, grp := range groups {
		memberCount, _ := g.countMembers(ctx, grp.ID)
		spaceCount, _ := g.countSpaces(ctx, grp.ID)
		out = append(out, GroupRow{
			ID:          grp.ID,
			Name:        grp.Name,
			Description: grp.Description,
			ExternalID:  grp.ExternalID,
			MemberCount: memberCount,
			SpaceCount:  spaceCount,
		})
	}
	return out, nil
}

func (g *GroupService) Get(ctx context.Context, id string) (*models.Group, error) {
	var grp models.Group
	if err := g.db.WithContext(ctx).First(&grp, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("group not found")
		}
		return nil, err
	}
	return &grp, nil
}

func (g *GroupService) Create(ctx context.Context, input CreateGroupInput) (*models.Group, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	grp := models.Group{
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		ExternalID:  strings.TrimSpace(input.ExternalID),
	}
	if err := g.db.WithContext(ctx).Create(&grp).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("group name already exists")
		}
		return nil, err
	}
	return &grp, nil
}

func (g *GroupService) Update(ctx context.Context, id string, input UpdateGroupInput) (*models.Group, error) {
	grp, err := g.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, errors.New("name is required")
		}
		grp.Name = name
	}
	if input.Description != nil {
		grp.Description = strings.TrimSpace(*input.Description)
	}
	if input.ExternalID != nil {
		grp.ExternalID = strings.TrimSpace(*input.ExternalID)
	}
	if err := g.db.WithContext(ctx).Save(grp).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("group name already exists")
		}
		return nil, err
	}
	return grp, nil
}

func (g *GroupService) Delete(ctx context.Context, id string) error {
	res := g.db.WithContext(ctx).Delete(&models.Group{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("group not found")
	}
	return nil
}

func (g *GroupService) ListMembers(ctx context.Context, groupID string) ([]GroupMemberRow, error) {
	if _, err := g.Get(ctx, groupID); err != nil {
		return nil, err
	}
	var users []models.User
	err := g.db.WithContext(ctx).
		Joins("JOIN group_members ON group_members.user_id = users.id").
		Where("group_members.group_id = ?", groupID).
		Order("users.email ASC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	out := make([]GroupMemberRow, 0, len(users))
	for _, u := range users {
		out = append(out, GroupMemberRow{
			UserID:      u.ID,
			Email:       u.Email,
			DisplayName: u.DisplayName,
		})
	}
	return out, nil
}

func (g *GroupService) AddMember(ctx context.Context, groupID string, userID string) error {
	if _, err := g.Get(ctx, groupID); err != nil {
		return err
	}
	var user models.User
	if err := g.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}
	var existing int64
	if err := g.db.WithContext(ctx).Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&existing).Error; err != nil {
		return err
	}
	if existing > 0 {
		return errors.New("user already in group")
	}
	return g.db.WithContext(ctx).Create(&models.GroupMember{GroupID: groupID, UserID: userID}).Error
}

func (g *GroupService) RemoveMember(ctx context.Context, groupID, userID string) error {
	res := g.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("member not found")
	}
	return nil
}

func (g *GroupService) ListSpaceGroups(ctx context.Context, spaceID string) ([]SpaceGroupRow, error) {
	type row struct {
		GroupID     string
		GroupName   string
		RoleName    string
		Description string
	}
	var rows []row
	err := g.db.WithContext(ctx).
		Table("space_groups sg").
		Select("g.id AS group_id, g.name AS group_name, r.name AS role_name, g.description").
		Joins("JOIN groups g ON g.id = sg.group_id").
		Joins("JOIN roles r ON r.id = sg.role_id").
		Where("sg.space_id = ?", spaceID).
		Order("g.name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]SpaceGroupRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, SpaceGroupRow{
			GroupID:     r.GroupID,
			GroupName:   r.GroupName,
			Role:        r.RoleName,
			Description: r.Description,
		})
	}
	return out, nil
}

func (g *GroupService) AssignToSpace(ctx context.Context, spaceID string, input AssignSpaceGroupInput) error {
	if _, err := g.Get(ctx, input.GroupID); err != nil {
		return err
	}
	var space models.Space
	if err := g.db.WithContext(ctx).First(&space, "id = ?", spaceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("space not found")
		}
		return err
	}
	roleName := input.Role
	if roleName == "" {
		roleName = "viewer"
	}
	if !validSpaceRoles[roleName] {
		return errors.New("invalid role: " + roleName)
	}
	var role models.Role
	if err := g.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return errors.New("invalid role: " + roleName)
	}
	sg := models.SpaceGroup{SpaceID: spaceID, GroupID: input.GroupID, RoleID: role.ID}
	return g.db.WithContext(ctx).
		Where("space_id = ? AND group_id = ?", spaceID, input.GroupID).
		Assign(models.SpaceGroup{RoleID: role.ID}).
		FirstOrCreate(&sg).Error
}

func (g *GroupService) RemoveFromSpace(ctx context.Context, spaceID, groupID string) error {
	res := g.db.WithContext(ctx).
		Where("space_id = ? AND group_id = ?", spaceID, groupID).
		Delete(&models.SpaceGroup{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("group not assigned to space")
	}
	return nil
}

func (g *GroupService) countMembers(ctx context.Context, groupID string) (int64, error) {
	var count int64
	err := g.db.WithContext(ctx).Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

func (g *GroupService) countSpaces(ctx context.Context, groupID string) (int64, error) {
	var count int64
	err := g.db.WithContext(ctx).Model(&models.SpaceGroup{}).Where("group_id = ?", groupID).Count(&count).Error
	return count, err
}

func (g *GroupService) UserHasSpaceAccessViaGroup(ctx context.Context, spaceID, userID string) (bool, error) {
	var count int64
	err := g.db.WithContext(ctx).
		Table("space_groups sg").
		Joins("JOIN group_members gm ON gm.group_id = sg.group_id").
		Where("sg.space_id = ? AND gm.user_id = ?", spaceID, userID).
		Count(&count).Error
	return count > 0, err
}
