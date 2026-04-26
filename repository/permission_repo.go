package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (r *PermissionRepository) FindAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Order("module, action").Find(&permissions).Error
	return permissions, err
}

func (r *PermissionRepository) FindByCodes(codes []string) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}
