package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type MenuService struct {
	menuRepo *repository.MenuRepository
}

func NewMenuService(menuRepo *repository.MenuRepository) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

func (s *MenuService) Create(hotelID string, req dto.CreateMenuItemRequest) (*models.MenuItem, error) {
	item := &models.MenuItem{
		HotelID:     hotelID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		IsAvailable: true,
	}

	if err := s.menuRepo.Create(item); err != nil {
		return nil, errors.New("failed to create menu item")
	}

	return item, nil
}

func (s *MenuService) GetByID(id, hotelID string) (*models.MenuItem, error) {
	item, err := s.menuRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("menu item not found")
	}
	return item, nil
}

func (s *MenuService) List(hotelID string, query dto.ListMenuQuery) ([]models.MenuItem, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.menuRepo.FindByHotelID(hotelID, query.Category, page, perPage)
}

func (s *MenuService) Update(id, hotelID string, req dto.UpdateMenuItemRequest) (*models.MenuItem, error) {
	item, err := s.menuRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("menu item not found")
	}

	if req.Name != nil {
		item.Name = *req.Name
	}
	if req.Description != nil {
		item.Description = *req.Description
	}
	if req.Price != nil {
		item.Price = *req.Price
	}
	if req.Category != nil {
		item.Category = *req.Category
	}
	if req.IsAvailable != nil {
		item.IsAvailable = *req.IsAvailable
	}

	if err := s.menuRepo.Update(item); err != nil {
		return nil, errors.New("failed to update menu item")
	}

	return item, nil
}

func (s *MenuService) Delete(id, hotelID string) error {
	_, err := s.menuRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return errors.New("menu item not found")
	}

	return s.menuRepo.Delete(id)
}
