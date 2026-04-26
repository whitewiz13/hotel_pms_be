package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type RoomTypeService struct {
	roomTypeRepo *repository.RoomTypeRepository
}

func NewRoomTypeService(roomTypeRepo *repository.RoomTypeRepository) *RoomTypeService {
	return &RoomTypeService{roomTypeRepo: roomTypeRepo}
}

func (s *RoomTypeService) Create(hotelID string, req dto.CreateRoomTypeRequest) (*models.RoomType, error) {
	// Check for duplicate name within the same hotel
	existing, _ := s.roomTypeRepo.FindByNameAndHotel(req.Name, hotelID)
	if existing != nil {
		return nil, errors.New("room type with this name already exists")
	}

	roomType := &models.RoomType{
		HotelID:     hotelID,
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
	}

	if err := s.roomTypeRepo.Create(roomType); err != nil {
		return nil, errors.New("failed to create room type")
	}

	return roomType, nil
}

func (s *RoomTypeService) GetByID(id, hotelID string) (*models.RoomType, error) {
	roomType, err := s.roomTypeRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("room type not found")
	}
	return roomType, nil
}

func (s *RoomTypeService) GetByHotelID(hotelID string) ([]models.RoomType, error) {
	return s.roomTypeRepo.FindByHotelID(hotelID)
}

func (s *RoomTypeService) Update(id, hotelID string, req dto.UpdateRoomTypeRequest) (*models.RoomType, error) {
	roomType, err := s.roomTypeRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("room type not found")
	}

	if req.Name != nil {
		// Check for duplicate name (excluding current)
		existing, _ := s.roomTypeRepo.FindByNameAndHotel(*req.Name, hotelID)
		if existing != nil && existing.ID != roomType.ID {
			return nil, errors.New("room type with this name already exists")
		}
		roomType.Name = *req.Name
	}
	if req.Description != nil {
		roomType.Description = *req.Description
	}
	if req.IsActive != nil {
		roomType.IsActive = *req.IsActive
	}

	if err := s.roomTypeRepo.Update(roomType); err != nil {
		return nil, errors.New("failed to update room type")
	}

	return roomType, nil
}

func (s *RoomTypeService) Delete(id, hotelID string) error {
	_, err := s.roomTypeRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return errors.New("room type not found")
	}
	return s.roomTypeRepo.Delete(id)
}
