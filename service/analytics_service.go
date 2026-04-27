package service

import (
	"fmt"
	"time"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/repository"
)

type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: analyticsRepo}
}

func (s *AnalyticsService) GetSummary(hotelID string, from, to time.Time) (*dto.AnalyticsSummaryResponse, error) {
	roomRev, serviceRev, activityRev, err := s.analyticsRepo.GetRevenueSummary(hotelID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue summary: %w", err)
	}

	total, checkIns, checkOuts, cancelled, err := s.analyticsRepo.CountReservationsByStatus(hotelID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to count reservations: %w", err)
	}

	guests, err := s.analyticsRepo.CountUniqueGuests(hotelID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to count guests: %w", err)
	}

	avgStay, err := s.analyticsRepo.GetAvgStayLength(hotelID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get avg stay: %w", err)
	}

	totalRooms, err := s.analyticsRepo.CountTotalActiveRooms(hotelID)
	if err != nil {
		return nil, fmt.Errorf("failed to count rooms: %w", err)
	}

	roomsSold, err := s.analyticsRepo.CountRoomsSold(hotelID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to count rooms sold: %w", err)
	}

	// ADR = room revenue / rooms sold
	var adr float64
	if roomsSold > 0 {
		adr = roomRev / float64(roomsSold)
	}

	// RevPAR = room revenue / total available room-nights
	days := to.Sub(from).Hours() / 24
	if days < 1 {
		days = 1
	}
	var revpar float64
	totalRoomNights := float64(totalRooms) * days
	if totalRoomNights > 0 {
		revpar = roomRev / totalRoomNights
	}

	// Occupancy rate for the period
	var occupancyRate float64
	if totalRoomNights > 0 {
		occupancyRate = float64(roomsSold) / totalRoomNights * 100
		if occupancyRate > 100 {
			occupancyRate = 100
		}
	}

	return &dto.AnalyticsSummaryResponse{
		TotalRevenue:      roomRev + serviceRev + activityRev,
		RoomRevenue:       roomRev,
		ServiceRevenue:    serviceRev,
		ActivityRevenue:   activityRev,
		TotalReservations: total,
		TotalCheckIns:     checkIns,
		TotalCheckOuts:    checkOuts,
		TotalCancelled:    cancelled,
		TotalGuests:       guests,
		AvgDailyRate:      adr,
		RevPAR:            revpar,
		AvgStayLength:     avgStay,
		OccupancyRate:     occupancyRate,
	}, nil
}

func (s *AnalyticsService) GetOccupancyTrend(hotelID string, from, to time.Time, period string) ([]dto.OccupancyPoint, error) {
	return s.analyticsRepo.GetOccupancyTrend(hotelID, from, to, period)
}

func (s *AnalyticsService) GetRevenueTrend(hotelID string, from, to time.Time, period string) ([]dto.RevenuePoint, error) {
	return s.analyticsRepo.GetRevenueTrend(hotelID, from, to, period)
}

func (s *AnalyticsService) GetReservationStats(hotelID string, from, to time.Time, period string) ([]dto.ReservationStatsPoint, error) {
	return s.analyticsRepo.GetReservationStatsTrend(hotelID, from, to, period)
}

func (s *AnalyticsService) GetRoomTypePerformance(hotelID string, from, to time.Time) ([]dto.RoomTypePerformance, error) {
	return s.analyticsRepo.GetRoomTypePerformance(hotelID, from, to)
}
