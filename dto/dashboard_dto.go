package dto

import "time"

type DashboardStatsResponse struct {
	TotalRooms          int64            `json:"total_rooms"`
	RoomsByStatus       map[string]int64 `json:"rooms_by_status"`
	OccupancyRate       float64          `json:"occupancy_rate"`
	TodayCheckIns       int64            `json:"today_check_ins"`
	TodayCheckOuts      int64            `json:"today_check_outs"`
	ActiveReservations  int64            `json:"active_reservations"`
	PendingHousekeeping int64            `json:"pending_housekeeping"`
}

type ActivityItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}
