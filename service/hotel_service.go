package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

type HotelService struct {
	db        *gorm.DB
	hotelRepo *repository.HotelRepository
	userRepo  *repository.UserRepository
}

func NewHotelService(db *gorm.DB, hotelRepo *repository.HotelRepository, userRepo *repository.UserRepository) *HotelService {
	return &HotelService{db: db, hotelRepo: hotelRepo, userRepo: userRepo}
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

	hotel := &models.Hotel{
		Name:        req.Name,
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
		admin = &models.User{
			Email:        req.AdminEmail,
			PasswordHash: hash,
			Name:         req.AdminName,
			Role:         models.RoleHotelAdmin,
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
