package dto

type SaveGuestSettingsRequest struct {
	WifiPassword    string `json:"wifi_password"`
	AllowOrders     *bool  `json:"allow_orders" binding:"required"`
	AllowActivities *bool  `json:"allow_activities" binding:"required"`
}
