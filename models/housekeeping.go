package models

type HousekeepingStatus string

const (
	HousekeepingStatusPending    HousekeepingStatus = "pending"
	HousekeepingStatusInProgress HousekeepingStatus = "in_progress"
	HousekeepingStatusCompleted  HousekeepingStatus = "completed"
)

type HousekeepingPriority string

const (
	HousekeepingPriorityLow    HousekeepingPriority = "low"
	HousekeepingPriorityNormal HousekeepingPriority = "normal"
	HousekeepingPriorityHigh   HousekeepingPriority = "high"
	HousekeepingPriorityUrgent HousekeepingPriority = "urgent"
)

type HousekeepingTask struct {
	BaseModel
	HotelID      string               `gorm:"type:uuid;not null;index" json:"hotel_id"`
	RoomID       string               `gorm:"type:uuid;not null;index" json:"room_id"`
	AssignedToID *string              `gorm:"type:uuid;index" json:"assigned_to_id,omitempty"`
	AssignedByID string               `gorm:"type:uuid;not null" json:"assigned_by_id"`
	Status       HousekeepingStatus   `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	Priority     HousekeepingPriority `gorm:"type:varchar(20);not null;default:'normal'" json:"priority"`
	Notes        string               `gorm:"type:text" json:"notes"`

	Hotel      Hotel `gorm:"foreignKey:HotelID" json:"-"`
	Room       Room  `gorm:"foreignKey:RoomID" json:"room,omitempty"`
	AssignedTo *User `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	AssignedBy User  `gorm:"foreignKey:AssignedByID" json:"assigned_by,omitempty"`
}

func (HousekeepingTask) TableName() string {
	return "housekeeping_tasks"
}
