package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *models.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) FindByID(id string) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) FindByHotelID(hotelID string) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Preload("Permissions").
		Where("hotel_id = ?", hotelID).
		Order("is_system DESC, name ASC").
		Find(&roles).Error
	return roles, err
}

func (r *RoleRepository) FindBySlugAndHotel(slug, hotelID string) (*models.Role, error) {
	var role models.Role
	err := r.db.Preload("Permissions").
		Where("slug = ? AND hotel_id = ?", slug, hotelID).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepository) Update(role *models.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Role{}).Error
}

func (r *RoleRepository) ReplacePermissions(role *models.Role, permissions []models.Permission) error {
	return r.db.Model(role).Association("Permissions").Replace(permissions)
}

func (r *RoleRepository) GetPermissionCodes(roleID string) ([]string, error) {
	var role models.Role
	err := r.db.Preload("Permissions").Where("id = ?", roleID).First(&role).Error
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(role.Permissions))
	for i, p := range role.Permissions {
		codes[i] = p.Code
	}
	return codes, nil
}

func (r *RoleRepository) CountUsersWithRole(roleID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}
