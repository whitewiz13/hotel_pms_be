package dto

// AnalyticsSummaryResponse provides KPI card data for a given date range.
type AnalyticsSummaryResponse struct {
	TotalRevenue     float64 `json:"total_revenue"`
	RoomRevenue      float64 `json:"room_revenue"`
	ServiceRevenue   float64 `json:"service_revenue"`
	ActivityRevenue  float64 `json:"activity_revenue"`
	TotalReservations int64  `json:"total_reservations"`
	TotalCheckIns    int64   `json:"total_check_ins"`
	TotalCheckOuts   int64   `json:"total_check_outs"`
	TotalCancelled   int64   `json:"total_cancelled"`
	TotalGuests      int64   `json:"total_guests"`
	AvgDailyRate     float64 `json:"avg_daily_rate"`
	RevPAR           float64 `json:"rev_par"`
	AvgStayLength    float64 `json:"avg_stay_length"`
	OccupancyRate    float64 `json:"occupancy_rate"`
}

// OccupancyPoint represents occupancy data for a single time bucket.
type OccupancyPoint struct {
	Date          string  `json:"date"`
	OccupiedRooms int64   `json:"occupied_rooms"`
	TotalRooms    int64   `json:"total_rooms"`
	Rate          float64 `json:"rate"`
}

// RevenuePoint represents revenue data for a single time bucket.
type RevenuePoint struct {
	Date            string  `json:"date"`
	RoomRevenue     float64 `json:"room_revenue"`
	ServiceRevenue  float64 `json:"service_revenue"`
	ActivityRevenue float64 `json:"activity_revenue"`
	Total           float64 `json:"total"`
}

// ReservationStatsPoint represents reservation activity for a single time bucket.
type ReservationStatsPoint struct {
	Date        string `json:"date"`
	Reservations int64 `json:"reservations"`
	CheckIns    int64  `json:"check_ins"`
	CheckOuts   int64  `json:"check_outs"`
	Cancellations int64 `json:"cancellations"`
}

// RoomTypePerformance shows analytics per room type.
type RoomTypePerformance struct {
	RoomType      string  `json:"room_type"`
	TotalRooms    int64   `json:"total_rooms"`
	Reservations  int64   `json:"reservations"`
	Revenue       float64 `json:"revenue"`
	OccupancyRate float64 `json:"occupancy_rate"`
	AvgRate       float64 `json:"avg_rate"`
}
