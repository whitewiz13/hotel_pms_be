package repository

import (
	"fmt"
	"time"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

type AnalyticsRepository struct {
	db *gorm.DB
}

func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// --- Summary queries ---

func (r *AnalyticsRepository) GetRevenueSummary(hotelID string, from, to time.Time) (room, service, activity float64, err error) {
	type result struct{ Total float64 }

	var res result
	err = r.db.Model(&models.Bill{}).
		Select("COALESCE(SUM(room_charges), 0) as total").
		Where("hotel_id = ? AND status = ? AND created_at BETWEEN ? AND ?", hotelID, models.BillStatusPaid, from, to).
		Scan(&res).Error
	if err != nil {
		return
	}
	room = res.Total

	err = r.db.Model(&models.Bill{}).
		Select("COALESCE(SUM(room_service_total), 0) as total").
		Where("hotel_id = ? AND status = ? AND created_at BETWEEN ? AND ?", hotelID, models.BillStatusPaid, from, to).
		Scan(&res).Error
	if err != nil {
		return
	}
	service = res.Total

	err = r.db.Model(&models.Bill{}).
		Select("COALESCE(SUM(activity_total), 0) as total").
		Where("hotel_id = ? AND status = ? AND created_at BETWEEN ? AND ?", hotelID, models.BillStatusPaid, from, to).
		Scan(&res).Error
	if err != nil {
		return
	}
	activity = res.Total

	return
}

func (r *AnalyticsRepository) CountReservationsByStatus(hotelID string, from, to time.Time) (total, checkIns, checkOuts, cancelled int64, err error) {
	base := r.db.Model(&models.Reservation{}).Where("hotel_id = ? AND created_at BETWEEN ? AND ?", hotelID, from, to)

	if err = base.Count(&total).Error; err != nil {
		return
	}
	if err = r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND status = ? AND check_in_date BETWEEN ? AND ?", hotelID, models.ReservationStatusCheckedIn, from, to).
		Count(&checkIns).Error; err != nil {
		return
	}
	if err = r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND status = ? AND check_out_date BETWEEN ? AND ?", hotelID, models.ReservationStatusCheckedOut, from, to).
		Count(&checkOuts).Error; err != nil {
		return
	}
	if err = r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND status = ? AND created_at BETWEEN ? AND ?", hotelID, models.ReservationStatusCancelled, from, to).
		Count(&cancelled).Error; err != nil {
		return
	}
	return
}

func (r *AnalyticsRepository) CountUniqueGuests(hotelID string, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND created_at BETWEEN ? AND ?", hotelID, from, to).
		Distinct("guest_id").
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) GetAvgStayLength(hotelID string, from, to time.Time) (float64, error) {
	var avg *float64
	err := r.db.Model(&models.Reservation{}).
		Select("AVG(check_out_date - check_in_date)").
		Where("hotel_id = ? AND status = ? AND check_out_date BETWEEN ? AND ?", hotelID, models.ReservationStatusCheckedOut, from, to).
		Scan(&avg).Error
	if err != nil || avg == nil {
		return 0, err
	}
	return *avg, nil
}

func (r *AnalyticsRepository) CountTotalActiveRooms(hotelID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Room{}).
		Where("hotel_id = ? AND is_active = true", hotelID).
		Count(&count).Error
	return count, err
}

func (r *AnalyticsRepository) CountRoomsSold(hotelID string, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.Reservation{}).
		Where("hotel_id = ? AND status IN ? AND check_in_date BETWEEN ? AND ?", hotelID,
			[]string{string(models.ReservationStatusCheckedIn), string(models.ReservationStatusCheckedOut)}, from, to).
		Count(&count).Error
	return count, err
}

// --- Trend queries ---

func (r *AnalyticsRepository) GetOccupancyTrend(hotelID string, from, to time.Time, trunc string) ([]dto.OccupancyPoint, error) {
	dateTrunc := pgDateTrunc(trunc)

	type row struct {
		Date          time.Time
		OccupiedRooms int64
	}

	var rows []row
	err := r.db.Model(&models.Reservation{}).
		Select(fmt.Sprintf("DATE_TRUNC('%s', check_in_date) as date, COUNT(*) as occupied_rooms", dateTrunc)).
		Where("hotel_id = ? AND status IN ? AND check_in_date BETWEEN ? AND ?", hotelID,
			[]string{string(models.ReservationStatusCheckedIn), string(models.ReservationStatusCheckedOut)}, from, to).
		Group("date").
		Order("date").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	totalRooms, err := r.CountTotalActiveRooms(hotelID)
	if err != nil {
		return nil, err
	}

	points := make([]dto.OccupancyPoint, len(rows))
	for i, row := range rows {
		var rate float64
		if totalRooms > 0 {
			rate = float64(row.OccupiedRooms) / float64(totalRooms) * 100
		}
		points[i] = dto.OccupancyPoint{
			Date:          row.Date.Format(dateFormat(trunc)),
			OccupiedRooms: row.OccupiedRooms,
			TotalRooms:    totalRooms,
			Rate:          rate,
		}
	}
	return points, nil
}

func (r *AnalyticsRepository) GetRevenueTrend(hotelID string, from, to time.Time, trunc string) ([]dto.RevenuePoint, error) {
	dateTrunc := pgDateTrunc(trunc)

	type row struct {
		Date            time.Time
		RoomRevenue     float64
		ServiceRevenue  float64
		ActivityRevenue float64
	}

	var rows []row
	err := r.db.Model(&models.Bill{}).
		Select(fmt.Sprintf(`DATE_TRUNC('%s', created_at) as date,
			COALESCE(SUM(room_charges), 0) as room_revenue,
			COALESCE(SUM(room_service_total), 0) as service_revenue,
			COALESCE(SUM(activity_total), 0) as activity_revenue`, dateTrunc)).
		Where("hotel_id = ? AND status = ? AND created_at BETWEEN ? AND ?", hotelID, models.BillStatusPaid, from, to).
		Group("date").
		Order("date").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	points := make([]dto.RevenuePoint, len(rows))
	for i, row := range rows {
		points[i] = dto.RevenuePoint{
			Date:            row.Date.Format(dateFormat(trunc)),
			RoomRevenue:     row.RoomRevenue,
			ServiceRevenue:  row.ServiceRevenue,
			ActivityRevenue: row.ActivityRevenue,
			Total:           row.RoomRevenue + row.ServiceRevenue + row.ActivityRevenue,
		}
	}
	return points, nil
}

func (r *AnalyticsRepository) GetReservationStatsTrend(hotelID string, from, to time.Time, trunc string) ([]dto.ReservationStatsPoint, error) {
	dateTrunc := pgDateTrunc(trunc)

	type row struct {
		Date          time.Time
		Reservations  int64
		CheckIns      int64
		CheckOuts     int64
		Cancellations int64
	}

	var rows []row
	err := r.db.Model(&models.Reservation{}).
		Select(fmt.Sprintf(`DATE_TRUNC('%s', created_at) as date,
			COUNT(*) as reservations,
			SUM(CASE WHEN status = 'checked_in' THEN 1 ELSE 0 END) as check_ins,
			SUM(CASE WHEN status = 'checked_out' THEN 1 ELSE 0 END) as check_outs,
			SUM(CASE WHEN status = 'cancelled' THEN 1 ELSE 0 END) as cancellations`, dateTrunc)).
		Where("hotel_id = ? AND created_at BETWEEN ? AND ?", hotelID, from, to).
		Group("date").
		Order("date").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	points := make([]dto.ReservationStatsPoint, len(rows))
	for i, row := range rows {
		points[i] = dto.ReservationStatsPoint{
			Date:          row.Date.Format(dateFormat(trunc)),
			Reservations:  row.Reservations,
			CheckIns:      row.CheckIns,
			CheckOuts:     row.CheckOuts,
			Cancellations: row.Cancellations,
		}
	}
	return points, nil
}

func (r *AnalyticsRepository) GetRoomTypePerformance(hotelID string, from, to time.Time) ([]dto.RoomTypePerformance, error) {
	type row struct {
		RoomType     string
		TotalRooms   int64
		Reservations int64
		Revenue      float64
	}

	// Room counts per type
	type roomCount struct {
		RoomType   string
		TotalRooms int64
	}
	var roomCounts []roomCount
	if err := r.db.Model(&models.Room{}).
		Select("room_type, COUNT(*) as total_rooms").
		Where("hotel_id = ? AND is_active = true", hotelID).
		Group("room_type").
		Scan(&roomCounts).Error; err != nil {
		return nil, err
	}

	roomCountMap := make(map[string]int64)
	for _, rc := range roomCounts {
		roomCountMap[rc.RoomType] = rc.TotalRooms
	}

	// Reservation + revenue per room type
	type perfRow struct {
		RoomType     string
		Reservations int64
		Revenue      float64
	}
	var perfRows []perfRow
	if err := r.db.Model(&models.Bill{}).
		Joins("JOIN rooms ON rooms.id = bills.room_id").
		Select("rooms.room_type, COUNT(bills.id) as reservations, COALESCE(SUM(bills.total_amount), 0) as revenue").
		Where("bills.hotel_id = ? AND bills.status = ? AND bills.created_at BETWEEN ? AND ?", hotelID, models.BillStatusPaid, from, to).
		Group("rooms.room_type").
		Scan(&perfRows).Error; err != nil {
		return nil, err
	}

	// Calculate the total number of days in the range
	days := to.Sub(from).Hours() / 24
	if days < 1 {
		days = 1
	}

	results := make([]dto.RoomTypePerformance, len(perfRows))
	for i, pr := range perfRows {
		totalRooms := roomCountMap[pr.RoomType]
		var occRate, avgRate float64
		if totalRooms > 0 && days > 0 {
			occRate = float64(pr.Reservations) / (float64(totalRooms) * days) * 100
			if occRate > 100 {
				occRate = 100
			}
		}
		if pr.Reservations > 0 {
			avgRate = pr.Revenue / float64(pr.Reservations)
		}

		results[i] = dto.RoomTypePerformance{
			RoomType:      pr.RoomType,
			TotalRooms:    totalRooms,
			Reservations:  pr.Reservations,
			Revenue:       pr.Revenue,
			OccupancyRate: occRate,
			AvgRate:       avgRate,
		}
	}
	return results, nil
}

// --- helpers ---

func pgDateTrunc(period string) string {
	switch period {
	case "monthly":
		return "month"
	case "yearly":
		return "year"
	default:
		return "day"
	}
}

func dateFormat(period string) string {
	switch period {
	case "monthly":
		return "2006-01"
	case "yearly":
		return "2006"
	default:
		return "2006-01-02"
	}
}
