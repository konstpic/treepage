package service

import (
	"context"
	"errors"

	"github.com/konstpic/treepage/backend/pkg/models"
	"gorm.io/gorm"
)

type SpaceMemberRow struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type AssignSpaceMemberInput struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role"`
}

func (s *SpaceService) ListMembers(ctx context.Context, spaceID string) ([]SpaceMemberRow, error) {
	type row struct {
		UserID      string
		Email       string
		DisplayName string
		RoleName    string
	}
	var rows []row
	err := s.db.WithContext(ctx).
		Table("space_members sm").
		Select("u.id AS user_id, u.email, u.display_name, r.name AS role_name").
		Joins("JOIN users u ON u.id = sm.user_id").
		Joins("JOIN roles r ON r.id = sm.role_id").
		Where("sm.space_id = ?", spaceID).
		Order("u.email ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]SpaceMemberRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, SpaceMemberRow{
			UserID:      r.UserID,
			Email:       r.Email,
			DisplayName: r.DisplayName,
			Role:        r.RoleName,
		})
	}
	return out, nil
}

func (s *SpaceService) AssignMember(ctx context.Context, spaceID string, input AssignSpaceMemberInput) error {
	if _, err := s.GetByID(ctx, spaceID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("space not found")
		}
		return err
	}
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", input.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
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
	if err := s.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		return errors.New("invalid role: " + roleName)
	}
	sm := models.SpaceMember{SpaceID: spaceID, UserID: input.UserID, RoleID: role.ID}
	return s.db.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, input.UserID).
		Assign(models.SpaceMember{RoleID: role.ID}).
		FirstOrCreate(&sm).Error
}

func (s *SpaceService) RemoveMember(ctx context.Context, spaceID, userID string) error {
	res := s.db.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, userID).
		Delete(&models.SpaceMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("member not found")
	}
	return nil
}
