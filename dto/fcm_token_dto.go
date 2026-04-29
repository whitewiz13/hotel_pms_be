package dto

type SaveFCMTokenRequest struct {
	DeviceToken string `json:"device_token" binding:"required,min=10"`
}
