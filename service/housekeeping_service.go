package service

import (
	"errors"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"gorm.io/gorm"
)

type HousekeepingService struct {
	db                *gorm.DB
	housekeepingRepo  *repository.HousekeepingRepository
	roomRepo          *repository.RoomRepository
	userRepo          *repository.UserRepository
}

func NewHousekeepingService(db *gorm.DB, housekeepingRepo *repository.HousekeepingRepository, roomRepo *repository.RoomRepository, userRepo *repository.UserRepository) *HousekeepingService {
	return &HousekeepingService{
		db:               db,
		housekeepingRepo: housekeepingRepo,
		roomRepo:         roomRepo,
		userRepo:         userRepo,
	}
}

// Assign creates a housekeeping task for a room and sets the room status to "cleaning".
func (s *HousekeepingService) Assign(hotelID, assignedByID string, req dto.AssignHousekeepingRequest) (*models.HousekeepingTask, error) {
	// Verify room belongs to hotel
	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil {
		return nil, errors.New("room not found")
	}
	if room.HotelID != hotelID {
		return nil, errors.New("room not found")
	}

	// Check no active task already exists for this room
	existing, _ := s.housekeepingRepo.FindActiveByRoomID(req.RoomID, hotelID)
	if existing != nil {
		return nil, errors.New("room already has an active housekeeping task")
	}

	// If assigning to a specific user, verify they belong to this hotel
	if req.AssignedToID != nil {
		user, err := s.userRepo.FindByID(*req.AssignedToID)
		if err != nil {
			return nil, errors.New("assigned user not found")
		}
		if user.HotelID == nil || *user.HotelID != hotelID {
			return nil, errors.New("assigned user does not belong to this hotel")
		}
	}

	task := &models.HousekeepingTask{
		HotelID:      hotelID,
		RoomID:       req.RoomID,
		AssignedToID: req.AssignedToID,
		AssignedByID: assignedByID,
		Status:       models.HousekeepingStatusPending,
		Priority:     req.Priority,
		Notes:        req.Notes,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(task).Error; err != nil {
			return errors.New("failed to create housekeeping task")
		}

		// Set room status to cleaning
		if err := tx.Model(&models.Room{}).Where("id = ?", req.RoomID).
			Update("status", models.RoomStatusCleaning).Error; err != nil {
			return errors.New("failed to update room status")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.housekeepingRepo.FindByID(task.ID.String())
}

// Complete marks a housekeeping task as completed and sets the room to "available".
func (s *HousekeepingService) Complete(id, hotelID, userID string, req dto.UpdateHousekeepingTaskRequest) (*models.HousekeepingTask, error) {
	task, err := s.housekeepingRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("housekeeping task not found")
	}

	if task.Status == models.HousekeepingStatusCompleted {
		return nil, errors.New("task is already completed")
	}

	// If task is assigned to someone, only that person can complete it
	if task.AssignedToID != nil && *task.AssignedToID != userID {
		return nil, errors.New("task is assigned to another housekeeper")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		task.Status = models.HousekeepingStatusCompleted
		if req.Notes != "" {
			task.Notes = req.Notes
		}
		// Track who actually completed it if it was unassigned
		if task.AssignedToID == nil {
			task.AssignedToID = &userID
		}

		if err := tx.Save(task).Error; err != nil {
			return errors.New("failed to update housekeeping task")
		}

		// Set room status to available
		if err := tx.Model(&models.Room{}).Where("id = ?", task.RoomID).
			Update("status", models.RoomStatusAvailable).Error; err != nil {
			return errors.New("failed to update room status")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.housekeepingRepo.FindByID(id)
}

// GetByID returns a single housekeeping task.
func (s *HousekeepingService) GetByID(id, hotelID string) (*models.HousekeepingTask, error) {
	task, err := s.housekeepingRepo.FindByIDAndHotel(id, hotelID)
	if err != nil {
		return nil, errors.New("housekeeping task not found")
	}
	return task, nil
}

// List returns filtered housekeeping tasks for a hotel, ordered by priority.
func (s *HousekeepingService) List(hotelID string, query dto.ListHousekeepingQuery) ([]models.HousekeepingTask, int64, error) {
	page := query.Page
	perPage := query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	return s.housekeepingRepo.FindByHotelID(hotelID, query.Status, query.AssignedToID, query.RoomID, query.Priority, page, perPage)
}
