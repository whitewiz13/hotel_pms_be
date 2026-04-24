package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type AmenityService struct {
	amenityRepo *repository.AmenityRepository
}

func NewAmenityService(amenityRepo *repository.AmenityRepository) *AmenityService {
	return &AmenityService{amenityRepo: amenityRepo}
}

func (s *AmenityService) Create(hotelID string, req dto.CreateAmenityRequest) (*models.Amenity, error) {
	amenity := &models.Amenity{
		HotelID:     hotelID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Category:    req.Category,
		IsActive:    true,
	}

	if err := s.amenityRepo.Create(amenity); err != nil {
		return nil, errors.New("failed to create amenity")
	}

	return amenity, nil
}

func (s *AmenityService) GetByID(id, hotelID string) (*models.Amenity, error) {
	amenity, err := s.amenityRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("amenity not found")
	}
	return amenity, nil
}

func (s *AmenityService) GetByHotelID(hotelID string, page, perPage int) ([]models.Amenity, int64, error) {
	return s.amenityRepo.FindByHotelID(hotelID, page, perPage)
}

func (s *AmenityService) GetByCategoryAndHotel(category, hotelID string, page, perPage int) ([]models.Amenity, int64, error) {
	return s.amenityRepo.FindByCategoryAndHotel(category, hotelID, page, perPage)
}

func (s *AmenityService) Update(id, hotelID string, req dto.UpdateAmenityRequest) (*models.Amenity, error) {
	amenity, err := s.amenityRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("amenity not found")
	}

	if req.Name != nil {
		amenity.Name = *req.Name
	}
	if req.Description != nil {
		amenity.Description = *req.Description
	}
	if req.Icon != nil {
		amenity.Icon = *req.Icon
	}
	if req.Category != nil {
		amenity.Category = *req.Category
	}
	if req.IsActive != nil {
		amenity.IsActive = *req.IsActive
	}

	if err := s.amenityRepo.Update(amenity); err != nil {
		return nil, errors.New("failed to update amenity")
	}

	return amenity, nil
}

func (s *AmenityService) Delete(id, hotelID string) error {
	_, err := s.amenityRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return errors.New("amenity not found")
	}
	return s.amenityRepo.Delete(id)
}
