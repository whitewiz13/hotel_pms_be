package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/utils"
)

type RoomService struct {
	roomRepo    *repository.RoomRepository
	amenityRepo *repository.AmenityRepository
}

func NewRoomService(roomRepo *repository.RoomRepository, amenityRepo *repository.AmenityRepository) *RoomService {
	return &RoomService{
		roomRepo:    roomRepo,
		amenityRepo: amenityRepo,
	}
}

func (s *RoomService) Create(req dto.CreateRoomRequest) (*models.Room, error) {
	room := &models.Room{
		HotelID:       req.HotelID,
		RoomNumber:    req.RoomNumber,
		RoomType:      req.RoomType,
		Floor:         req.Floor,
		Status:        models.RoomStatusAvailable,
		PricePerNight: req.PricePerNight,
		Description:   req.Description,
		MaxOccupancy:  req.MaxOccupancy,
		IsActive:      true,
	}

	if err := s.roomRepo.Create(room); err != nil {
		return nil, errors.New("failed to create room (room number may already exist for this hotel)")
	}

	if len(req.AmenityIDs) > 0 {
		amenities, err := s.amenityRepo.FindByIDs(req.AmenityIDs)
		if err == nil {
			_ = s.roomRepo.UpdateAmenities(room, amenities)
		}
	}

	return s.roomRepo.FindByID(room.ID.String())
}

func (s *RoomService) GetByID(id string) (*models.Room, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("room not found")
	}
	return room, nil
}

func (s *RoomService) GetByHotelID(hotelID string, page, perPage int) ([]models.Room, int64, error) {
	return s.roomRepo.FindByHotelID(hotelID, page, perPage)
}

func (s *RoomService) Update(id string, req dto.UpdateRoomRequest) (*models.Room, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("room not found")
	}

	if req.RoomNumber != nil {
		room.RoomNumber = *req.RoomNumber
	}
	if req.RoomType != nil {
		room.RoomType = *req.RoomType
	}
	if req.Floor != nil {
		room.Floor = *req.Floor
	}
	if req.Status != nil {
		room.Status = *req.Status
	}
	if req.PricePerNight != nil {
		room.PricePerNight = *req.PricePerNight
	}
	if req.Description != nil {
		room.Description = *req.Description
	}
	if req.MaxOccupancy != nil {
		room.MaxOccupancy = *req.MaxOccupancy
	}
	if req.IsActive != nil {
		room.IsActive = *req.IsActive
	}

	if err := s.roomRepo.Update(room); err != nil {
		return nil, errors.New("failed to update room")
	}

	if req.AmenityIDs != nil {
		amenities, err := s.amenityRepo.FindByIDs(req.AmenityIDs)
		if err == nil {
			_ = s.roomRepo.UpdateAmenities(room, amenities)
		}
	}

	return s.roomRepo.FindByID(id)
}

func (s *RoomService) Delete(id string) error {
	_, err := s.roomRepo.FindByID(id)
	if err != nil {
		return errors.New("room not found")
	}
	return s.roomRepo.Delete(id)
}

func (s *RoomService) SetAccessPin(id string) (string, error) {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		return "", errors.New("room not found")
	}

	pin, err := utils.GeneratePin(6)
	if err != nil {
		return "", errors.New("failed to generate pin")
	}

	room.AccessPin = pin
	if err := s.roomRepo.Update(room); err != nil {
		return "", errors.New("failed to set pin")
	}

	return pin, nil
}

func (s *RoomService) ClearAccessPin(id string) error {
	room, err := s.roomRepo.FindByID(id)
	if err != nil {
		return errors.New("room not found")
	}

	room.AccessPin = ""
	return s.roomRepo.Update(room)
}
