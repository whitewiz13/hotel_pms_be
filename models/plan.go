package models

// PlanID constants
const (
	PlanFree  = "free"
	PlanBasic = "basic"
	PlanPro   = "pro"
)

// Plan defines the subscription tier with resource limits and feature flags.
type Plan struct {
	ID   string `gorm:"primaryKey;size:20" json:"id"`
	Name string `gorm:"not null;size:50" json:"name"`

	// Resource limits (-1 = unlimited)
	MaxRooms             int `gorm:"not null" json:"max_rooms"`
	MaxStaff             int `gorm:"not null" json:"max_staff"`
	MaxReservationsMonth int `gorm:"not null" json:"max_reservations_month"`
	MaxStorageMB         int `gorm:"not null" json:"max_storage_mb"`

	// Feature flags
	FeatureRoomService   bool `gorm:"not null;default:false" json:"feature_room_service"`
	FeatureActivities    bool `gorm:"not null;default:false" json:"feature_activities"`
	FeatureGuestPortal   bool `gorm:"not null;default:false" json:"feature_guest_portal"`
	FeatureNotifications bool `gorm:"not null;default:false" json:"feature_notifications"`
	FeatureAnalytics     bool `gorm:"not null;default:false" json:"feature_analytics"`
	FeatureAdvAnalytics  bool `gorm:"not null;default:false" json:"feature_adv_analytics"`
	FeatureCustomRoles   bool `gorm:"not null;default:false" json:"feature_custom_roles"`
	FeatureGuestUploads  bool `gorm:"not null;default:false" json:"feature_guest_uploads"`
}

func (Plan) TableName() string {
	return "plans"
}
