package service

import (
	"fmt"
	"sort"
	"sync"

	"github.com/hotelpms/backend/dto"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
)

type DashboardService struct {
	dashboardRepo *repository.DashboardRepository
}

func NewDashboardService(dashboardRepo *repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

func (s *DashboardService) GetStats(hotelID string) (*dto.DashboardStatsResponse, error) {
	var (
		roomsByStatus       map[string]int64
		totalRooms          int64
		todayCheckIns       int64
		todayCheckOuts      int64
		activeReservations  int64
		pendingHousekeeping int64
		mu                  sync.Mutex
		firstErr            error
		wg                  sync.WaitGroup
	)

	setErr := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}

	wg.Add(5)

	go func() {
		defer wg.Done()
		r, t, err := s.dashboardRepo.CountRoomsByStatus(hotelID)
		if err != nil {
			setErr(fmt.Errorf("failed to count rooms: %w", err))
			return
		}
		mu.Lock()
		roomsByStatus = r
		totalRooms = t
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		c, err := s.dashboardRepo.CountTodayCheckIns(hotelID)
		if err != nil {
			setErr(fmt.Errorf("failed to count today's check-ins: %w", err))
			return
		}
		mu.Lock()
		todayCheckIns = c
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		c, err := s.dashboardRepo.CountTodayCheckOuts(hotelID)
		if err != nil {
			setErr(fmt.Errorf("failed to count today's check-outs: %w", err))
			return
		}
		mu.Lock()
		todayCheckOuts = c
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		c, err := s.dashboardRepo.CountActiveReservations(hotelID)
		if err != nil {
			setErr(fmt.Errorf("failed to count active reservations: %w", err))
			return
		}
		mu.Lock()
		activeReservations = c
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		c, err := s.dashboardRepo.CountPendingHousekeeping(hotelID)
		if err != nil {
			setErr(fmt.Errorf("failed to count pending housekeeping: %w", err))
			return
		}
		mu.Lock()
		pendingHousekeeping = c
		mu.Unlock()
	}()

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	var occupancyRate float64
	if totalRooms > 0 {
		occupied := roomsByStatus[string(models.RoomStatusOccupied)]
		occupancyRate = float64(occupied) / float64(totalRooms) * 100
	}

	return &dto.DashboardStatsResponse{
		TotalRooms:          totalRooms,
		RoomsByStatus:       roomsByStatus,
		OccupancyRate:       occupancyRate,
		TodayCheckIns:       todayCheckIns,
		TodayCheckOuts:      todayCheckOuts,
		ActiveReservations:  activeReservations,
		PendingHousekeeping: pendingHousekeeping,
	}, nil
}

func (s *DashboardService) GetActivity(hotelID string, limit int) ([]dto.ActivityItem, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	reservations, err := s.dashboardRepo.GetRecentReservations(hotelID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent reservations: %w", err)
	}

	tasks, err := s.dashboardRepo.GetRecentHousekeepingTasks(hotelID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent housekeeping: %w", err)
	}

	var items []dto.ActivityItem

	for _, r := range reservations {
		item := dto.ActivityItem{
			ID:        r.ID.String(),
			Timestamp: r.UpdatedAt,
		}

		switch r.Status {
		case models.ReservationStatusReserved:
			item.Type = "reservation_created"
			item.Message = fmt.Sprintf("New reservation for %s in room %s", r.GuestName, r.Room.RoomNumber)
		case models.ReservationStatusCheckedIn:
			item.Type = "check_in"
			item.Message = fmt.Sprintf("%s checked in to room %s", r.GuestName, r.Room.RoomNumber)
		case models.ReservationStatusCheckedOut:
			item.Type = "check_out"
			item.Message = fmt.Sprintf("%s checked out from room %s", r.GuestName, r.Room.RoomNumber)
		case models.ReservationStatusCancelled:
			item.Type = "reservation_cancelled"
			item.Message = fmt.Sprintf("Reservation for %s in room %s was cancelled", r.GuestName, r.Room.RoomNumber)
		}

		items = append(items, item)
	}

	for _, t := range tasks {
		item := dto.ActivityItem{
			ID:        t.ID.String(),
			Timestamp: t.UpdatedAt,
		}

		switch t.Status {
		case models.HousekeepingStatusPending:
			item.Type = "housekeeping_assigned"
			item.Message = fmt.Sprintf("Housekeeping task assigned for room %s", t.Room.RoomNumber)
		case models.HousekeepingStatusInProgress:
			item.Type = "housekeeping_started"
			item.Message = fmt.Sprintf("Housekeeping started for room %s", t.Room.RoomNumber)
		case models.HousekeepingStatusCompleted:
			item.Type = "housekeeping_completed"
			item.Message = fmt.Sprintf("Housekeeping completed for room %s", t.Room.RoomNumber)
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp.After(items[j].Timestamp)
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}
