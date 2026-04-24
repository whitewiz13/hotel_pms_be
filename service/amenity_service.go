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

func (s *AmenityService) Create(req dto.CreateAmenityRequest) (*models.Amenity, error) {
	amenity := &models.Amenity{
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

func (s *AmenityService) GetByID(id string) (*models.Amenity, error) {
	amenity, err := s.amenityRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("amenity not found")
	}
	return amenity, nil
}

func (s *AmenityService) GetAll(page, perPage int) ([]models.Amenity, int64, error) {
	return s.amenityRepo.FindAll(page, perPage)
}

func (s *AmenityService) GetByCategory(category string, page, perPage int) ([]models.Amenity, int64, error) {
	return s.amenityRepo.FindByCategory(category, page, perPage)
}

func (s *AmenityService) Update(id string, req dto.UpdateAmenityRequest) (*models.Amenity, error) {
	amenity, err := s.amenityRepo.FindByID(id)
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

func (s *AmenityService) Delete(id string) error {
	_, err := s.amenityRepo.FindByID(id)
	if err != nil {
		return errors.New("amenity not found")
	}
	return s.amenityRepo.Delete(id)
}
