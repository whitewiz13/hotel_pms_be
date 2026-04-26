package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

type HotelService struct {
	db             *gorm.DB
	hotelRepo      *repository.HotelRepository
	userRepo       *repository.UserRepository
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
}

func NewHotelService(db *gorm.DB, hotelRepo *repository.HotelRepository, userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, permissionRepo *repository.PermissionRepository) *HotelService {
	return &HotelService{db: db, hotelRepo: hotelRepo, userRepo: userRepo, roleRepo: roleRepo, permissionRepo: permissionRepo}
}

func (s *HotelService) Create(req dto.CreateHotelRequest) (*models.Hotel, *models.User, error) {
	exists, err := s.userRepo.ExistsByEmail(req.AdminEmail)
	if err != nil {
		return nil, nil, errors.New("failed to check existing user")
	}
	if exists {
		return nil, nil, errors.New("admin email already registered")
	}

	hash, err := utils.HashPassword(req.AdminPassword)
	if err != nil {
		return nil, nil, errors.New("failed to hash password")
	}

	slug, err := s.generateUniqueSlug(req.Name)
	if err != nil {
		return nil, nil, errors.New("failed to generate hotel slug")
	}

	hotel := &models.Hotel{
		Name:        req.Name,
		Slug:        slug,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		ZipCode:     req.ZipCode,
		Phone:       req.Phone,
		Email:       req.Email,
		Description: req.Description,
		IsActive:    true,
	}

	var admin *models.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(hotel).Error; err != nil {
			return err
		}

		hotelID := hotel.ID.String()

		// Create Hotel Admin role with all permissions except super-admin-only ones
		allPerms, err := s.permissionRepo.FindAll()
		if err != nil {
			return fmt.Errorf("failed to load permissions: %w", err)
		}

		// Exclude super-admin-only permissions
		excluded := map[string]bool{
			"hotels:read":   true,
			"hotels:update": true,
		}
		var rolePerms []models.Permission
		for _, p := range allPerms {
			if !excluded[p.Code] {
				rolePerms = append(rolePerms, p)
			}
		}

		adminRole := &models.Role{
			HotelID:     &hotelID,
			Name:        "Hotel Admin",
			Slug:        "hotel-admin",
			Description: "Full access to all hotel features",
			IsSystem:    false,
			Permissions: rolePerms,
		}
		if err := tx.Create(adminRole).Error; err != nil {
			return fmt.Errorf("failed to create admin role: %w", err)
		}

		roleID := adminRole.ID.String()
		admin = &models.User{
			Email:        req.AdminEmail,
			PasswordHash: hash,
			Name:         req.AdminName,
			Role:         models.RoleHotelAdmin,
			RoleID:       &roleID,
			HotelID:      &hotelID,
			IsActive:     true,
		}
		return tx.Create(admin).Error
	})

	if err != nil {
		return nil, nil, errors.New("failed to create hotel")
	}

	return hotel, admin, nil
}

func (s *HotelService) GetByID(id string) (*models.Hotel, error) {
	hotel, err := s.hotelRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("hotel not found")
	}
	return hotel, nil
}

func (s *HotelService) GetAll(page, perPage int) ([]models.Hotel, int64, error) {
	return s.hotelRepo.FindAll(page, perPage)
}

func (s *HotelService) Update(id string, req dto.UpdateHotelRequest) (*models.Hotel, error) {
	hotel, err := s.hotelRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("hotel not found")
	}

	if req.Name != nil {
		hotel.Name = *req.Name
	}
	if req.Address != nil {
		hotel.Address = *req.Address
	}
	if req.City != nil {
		hotel.City = *req.City
	}
	if req.State != nil {
		hotel.State = *req.State
	}
	if req.Country != nil {
		hotel.Country = *req.Country
	}
	if req.ZipCode != nil {
		hotel.ZipCode = *req.ZipCode
	}
	if req.Phone != nil {
		hotel.Phone = *req.Phone
	}
	if req.Email != nil {
		hotel.Email = *req.Email
	}
	if req.Description != nil {
		hotel.Description = *req.Description
	}
	if req.IsActive != nil {
		hotel.IsActive = *req.IsActive
	}

	if err := s.hotelRepo.Update(hotel); err != nil {
		return nil, errors.New("failed to update hotel")
	}

	return hotel, nil
}

func (s *HotelService) Delete(id string) error {
	_, err := s.hotelRepo.FindByID(id)
	if err != nil {
		return errors.New("hotel not found")
	}
	return s.hotelRepo.Delete(id)
}

func (s *HotelService) GetBySlug(slug string) (*models.Hotel, error) {
	hotel, err := s.hotelRepo.FindBySlug(slug)
	if err != nil {
		return nil, errors.New("hotel not found")
	}
	return hotel, nil
}

// generateUniqueSlug creates a URL-friendly slug from the hotel name
// with a 4-character random suffix, e.g. "the-london-tipton-a3f9".
func (s *HotelService) generateUniqueSlug(name string) (string, error) {
	base := slugify(name)

	const maxAttempts = 5
	for range maxAttempts {
		slug := fmt.Sprintf("%s-%s", base, randomSuffix(4))
		exists, err := s.hotelRepo.ExistsBySlug(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
	return "", errors.New("could not generate unique slug")
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9-]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = multiDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func randomSuffix(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	rand.Read(b)
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}
