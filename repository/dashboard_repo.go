package repository

import (
	"time"

	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) CountRoomsByStatus(hotelID string) (map[string]int64, int64, error) {
	type statusCount struct {
		Status string
		Count  int64
	}
	var rows []statusCount

	err := r.db.Model(&models.Room{}).
		Select("status, count(*) as count").
		Where("hotel_id = ? AND is_active = true", hotelID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make(map[string]int64)
	var total int64
	for _, row := range rows {
		result[row.Status] = row.Count
		total += row.Count
	}

	return result, total, nil
}

func (r *DashboardRepository) CountTodayCheckIns(hotelID string) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND check_in_date = ? AND status IN ?", hotelID, today,
			[]string{string(models.ReservationStatusReserved), string(models.ReservationStatusCheckedIn)}).
		Count(&count).Error
	return count, err
}

func (r *DashboardRepository) CountTodayCheckOuts(hotelID string) (int64, error) {
	var count int64
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND check_out_date = ? AND status IN ?", hotelID, today,
			[]string{string(models.ReservationStatusCheckedIn), string(models.ReservationStatusCheckedOut)}).
		Count(&count).Error
	return count, err
}

func (r *DashboardRepository) CountActiveReservations(hotelID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND status IN ?", hotelID,
			[]string{string(models.ReservationStatusReserved), string(models.ReservationStatusCheckedIn)}).
		Count(&count).Error
	return count, err
}

func (r *DashboardRepository) CountPendingHousekeeping(hotelID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.HousekeepingTask{}).
		Where("hotel_id = ? AND status IN ?", hotelID,
			[]string{string(models.HousekeepingStatusPending), string(models.HousekeepingStatusInProgress)}).
		Count(&count).Error
	return count, err
}

func (r *DashboardRepository) GetRecentReservations(hotelID string, limit int) ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.Preload("Room").
		Where("hotel_id = ?", hotelID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&reservations).Error
	return reservations, err
}

func (r *DashboardRepository) GetRecentHousekeepingTasks(hotelID string, limit int) ([]models.HousekeepingTask, error) {
	var tasks []models.HousekeepingTask
	err := r.db.Preload("Room").Preload("AssignedTo").
		Where("hotel_id = ?", hotelID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}
