package service

import (
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetAll(page, perPage int, role string) ([]models.User, int64, error) {
	return s.userRepo.FindAll(page, perPage, role)
}

func (s *UserService) GetByHotelID(hotelID string, page, perPage int) ([]models.User, int64, error) {
	return s.userRepo.FindByHotelID(hotelID, page, perPage)
}
