package models

// FCMToken stores a Firebase Cloud Messaging device token for a user.
// Multiple tokens per user are allowed (multi-device support).
type FCMToken struct {
	BaseModel
	UserID      string `gorm:"type:uuid;not null;index" json:"user_id"`
	HotelID     string `gorm:"type:uuid;not null;index" json:"hotel_id"`
	DeviceToken string `gorm:"not null;size:500" json:"device_token"`

	User  User  `gorm:"foreignKey:UserID" json:"-"`
	Hotel Hotel `gorm:"foreignKey:HotelID" json:"-"`
}

func (FCMToken) TableName() string {
	return "fcm_tokens"
}
