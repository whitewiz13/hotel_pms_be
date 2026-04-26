package service

import (
	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type GuestSettingsService struct {
	settingsRepo *repository.GuestSettingsRepository
}

func NewGuestSettingsService(settingsRepo *repository.GuestSettingsRepository) *GuestSettingsService {
	return &GuestSettingsService{settingsRepo: settingsRepo}
}

func (s *GuestSettingsService) Get(hotelID string) (*models.GuestSettings, error) {
	settings, err := s.settingsRepo.FindByHotelID(hotelID)
	if err != nil {
		// Return defaults if no settings saved yet
		return &models.GuestSettings{
			HotelID:         hotelID,
			WifiPassword:    "",
			AllowOrders:     true,
			AllowActivities: true,
		}, nil
	}
	return settings, nil
}

func (s *GuestSettingsService) Save(hotelID string, req dto.SaveGuestSettingsRequest) (*models.GuestSettings, error) {
	settings := &models.GuestSettings{
		HotelID:         hotelID,
		WifiPassword:    req.WifiPassword,
		AllowOrders:     *req.AllowOrders,
		AllowActivities: *req.AllowActivities,
	}

	if err := s.settingsRepo.Upsert(settings); err != nil {
		return nil, err
	}

	return s.settingsRepo.FindByHotelID(hotelID)
}
