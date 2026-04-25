package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type HousekeepingRepository struct {
	db *gorm.DB
}

func NewHousekeepingRepository(db *gorm.DB) *HousekeepingRepository {
	return &HousekeepingRepository{db: db}
}

func (r *HousekeepingRepository) Create(task *models.HousekeepingTask) error {
	return r.db.Create(task).Error
}

func (r *HousekeepingRepository) FindByID(id string) (*models.HousekeepingTask, error) {
	var task models.HousekeepingTask
	err := r.db.Preload("Room").Preload("AssignedTo").Preload("AssignedBy").
		Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *HousekeepingRepository) FindByIDAndHotel(id, hotelID string) (*models.HousekeepingTask, error) {
	var task models.HousekeepingTask
	err := r.db.Preload("Room").Preload("AssignedTo").Preload("AssignedBy").
		Where("id = ? AND hotel_id = ?", id, hotelID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *HousekeepingRepository) FindByHotelID(hotelID, status, assignedToID, roomID, priority string, page, perPage int) ([]models.HousekeepingTask, int64, error) {
	var tasks []models.HousekeepingTask
	var total int64

	query := r.db.Where("housekeeping_tasks.hotel_id = ?", hotelID)

	if status != "" {
		query = query.Where("housekeeping_tasks.status = ?", status)
	}
	if assignedToID != "" {
		query = query.Where("housekeeping_tasks.assigned_to_id = ?", assignedToID)
	}
	if roomID != "" {
		query = query.Where("housekeeping_tasks.room_id = ?", roomID)
	}
	if priority != "" {
		query = query.Where("housekeeping_tasks.priority = ?", priority)
	}

	query.Model(&models.HousekeepingTask{}).Count(&total)

	err := query.Preload("Room").Preload("AssignedTo").Preload("AssignedBy").
		Order("CASE housekeeping_tasks.priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'normal' THEN 3 WHEN 'low' THEN 4 END, housekeeping_tasks.created_at ASC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&tasks).Error

	return tasks, total, err
}

func (r *HousekeepingRepository) FindActiveByRoomID(roomID, hotelID string) (*models.HousekeepingTask, error) {
	var task models.HousekeepingTask
	err := r.db.Where("room_id = ? AND hotel_id = ? AND status IN ?", roomID, hotelID,
		[]string{string(models.HousekeepingStatusPending), string(models.HousekeepingStatusInProgress)}).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *HousekeepingRepository) Update(task *models.HousekeepingTask) error {
	return r.db.Save(task).Error
}
