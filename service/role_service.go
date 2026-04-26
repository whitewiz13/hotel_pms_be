package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type RoleService struct {
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
}

func NewRoleService(roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

func (s *RoleService) CreateRole(hotelID string, req dto.CreateRoleRequest) (*models.Role, error) {
	permissions, err := s.permissionRepo.FindByCodes(req.Permissions)
	if err != nil {
		return nil, errors.New("failed to validate permissions")
	}
	if len(permissions) != len(req.Permissions) {
		return nil, errors.New("one or more invalid permission codes")
	}

	role := &models.Role{
		HotelID:     &hotelID,
		Name:        req.Name,
		Slug:        slugify(req.Name),
		Description: req.Description,
		IsSystem:    false,
		Permissions: permissions,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, errors.New("failed to create role")
	}

	return role, nil
}

func (s *RoleService) GetRolesByHotel(hotelID string) ([]models.Role, error) {
	return s.roleRepo.FindByHotelID(hotelID)
}

func (s *RoleService) GetRoleByID(id string) (*models.Role, error) {
	return s.roleRepo.FindByID(id)
}

func (s *RoleService) UpdateRole(id string, req dto.UpdateRoleRequest) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	if role.IsSystem {
		return nil, errors.New("cannot modify system role")
	}

	if req.Name != nil {
		role.Name = *req.Name
		role.Slug = slugify(*req.Name)
	}
	if req.Description != nil {
		role.Description = *req.Description
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, errors.New("failed to update role")
	}

	if req.Permissions != nil {
		permissions, err := s.permissionRepo.FindByCodes(req.Permissions)
		if err != nil {
			return nil, errors.New("failed to validate permissions")
		}
		if len(permissions) != len(req.Permissions) {
			return nil, errors.New("one or more invalid permission codes")
		}
		if err := s.roleRepo.ReplacePermissions(role, permissions); err != nil {
			return nil, errors.New("failed to update permissions")
		}
		role.Permissions = permissions
	}

	return role, nil
}

func (s *RoleService) DeleteRole(id string) error {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return errors.New("role not found")
	}

	if role.IsSystem {
		return errors.New("cannot delete system role")
	}

	count, err := s.roleRepo.CountUsersWithRole(id)
	if err != nil {
		return errors.New("failed to check role usage")
	}
	if count > 0 {
		return errors.New("cannot delete role that is assigned to users")
	}

	return s.roleRepo.Delete(id)
}

func (s *RoleService) GetAllPermissions() ([]models.Permission, error) {
	return s.permissionRepo.FindAll()
}


