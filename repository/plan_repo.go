package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type PlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) *PlanRepository {
	return &PlanRepository{db: db}
}

func (r *PlanRepository) FindAll() ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.Order("CASE id WHEN 'free' THEN 1 WHEN 'basic' THEN 2 WHEN 'pro' THEN 3 END").Find(&plans).Error
	return plans, err
}

func (r *PlanRepository) FindByID(id string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.Where("id = ?", id).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *PlanRepository) FindSubscriptionByHotelID(hotelID string) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.Preload("Plan").Where("hotel_id = ?", hotelID).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *PlanRepository) CreateSubscription(sub *models.Subscription) error {
	return r.db.Create(sub).Error
}

func (r *PlanRepository) UpdateSubscription(sub *models.Subscription) error {
	return r.db.Model(&models.Subscription{}).
		Where("id = ?", sub.ID).
		Updates(map[string]interface{}{
			"plan_id":           sub.PlanID,
			"status":            sub.Status,
			"assigned_at":       sub.AssignedAt,
			"renewed_at":        sub.RenewedAt,
			"expires_at":        sub.ExpiresAt,
			"suspended_at":      sub.SuspendedAt,
			"suspension_reason": sub.SuspensionReason,
		}).Error
}

func (r *PlanRepository) CountRoomsByHotel(hotelID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Room{}).Where("hotel_id = ? AND deleted_at IS NULL", hotelID).Count(&count).Error
	return count, err
}

func (r *PlanRepository) CountStaffByHotel(hotelID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("hotel_id = ? AND deleted_at IS NULL AND role != ?", hotelID, models.RoleSuperAdmin).Count(&count).Error
	return count, err
}

func (r *PlanRepository) CountReservationsThisMonth(hotelID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND deleted_at IS NULL AND created_at >= date_trunc('month', CURRENT_DATE)", hotelID).
		Count(&count).Error
	return count, err
}
