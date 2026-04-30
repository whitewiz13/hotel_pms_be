package repository

import (
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) FindByID(id string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Items.MenuItem").Preload("Room").Preload("Guest").Preload("AssignedTo").
		Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByIDAndHotel(id, hotelID string) (*models.Order, error) {
	var order models.Order
	if err := r.db.Preload("Items.MenuItem").Preload("Room").Preload("Guest").Preload("AssignedTo").
		Where("id = ? AND hotel_id = ?", id, hotelID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) FindByHotelID(hotelID, status, roomID, reservationID, assignedToID string, page, perPage int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Where("hotel_id = ?", hotelID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if roomID != "" {
		query = query.Where("room_id = ?", roomID)
	}
	if reservationID != "" {
		query = query.Where("reservation_id = ?", reservationID)
	}
	if assignedToID != "" {
		query = query.Where("assigned_to_id = ?", assignedToID)
	}

	if err := query.Model(&models.Order{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	if err := query.Preload("Items.MenuItem").Preload("Room").Preload("Guest").Preload("AssignedTo").
		Order("created_at DESC").Offset(offset).Limit(perPage).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// FindByReservationID returns all non-cancelled orders for a reservation (used for billing).
func (r *OrderRepository) FindByReservationID(reservationID string) ([]models.Order, error) {
	var orders []models.Order
	if err := r.db.Preload("Items.MenuItem").
		Where("reservation_id = ? AND status != ?", reservationID, models.OrderStatusCancelled).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *OrderRepository) UpdateAssignedTo(orderID, assignedToID string) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).Update("assigned_to_id", assignedToID).Error
}

// CountByReservationGroupedByStatus returns order counts per status for a reservation.
func (r *OrderRepository) CountByReservationGroupedByStatus(reservationID string) (map[string]int64, error) {
	type result struct {
		Status string
		Count  int64
	}
	var results []result
	if err := r.db.Model(&models.Order{}).
		Select("status, count(*) as count").
		Where("reservation_id = ?", reservationID).
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}
