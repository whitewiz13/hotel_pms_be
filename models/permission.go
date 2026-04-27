package models

// Permission represents a single permission in the system.
type Permission struct {
	BaseModel
	Code        string `gorm:"uniqueIndex;not null;size:100" json:"code"`
	Module      string `gorm:"not null;size:50;index" json:"module"`
	Action      string `gorm:"not null;size:50" json:"action"`
	Description string `gorm:"size:255" json:"description"`
}

func (Permission) TableName() string {
	return "permissions"
}

// PermissionDef defines a permission to be seeded.
type PermissionDef struct {
	Code        string
	Module      string
	Action      string
	Description string
}

// AllPermissionDefs returns all available permission definitions for seeding.
func AllPermissionDefs() []PermissionDef {
	return []PermissionDef{
		// Dashboard
		{"dashboard:view", "dashboard", "view", "View dashboard and statistics"},

		// Rooms
		{"rooms:create", "rooms", "create", "Create new rooms"},
		{"rooms:read", "rooms", "read", "View room details"},
		{"rooms:update", "rooms", "update", "Update room information"},
		{"rooms:delete", "rooms", "delete", "Delete rooms"},
		{"rooms:manage_pin", "rooms", "manage_pin", "Generate and clear room access PINs"},

		// Room Types
		{"room_types:create", "room_types", "create", "Create room types"},
		{"room_types:read", "room_types", "read", "View room types"},
		{"room_types:update", "room_types", "update", "Update room types"},
		{"room_types:delete", "room_types", "delete", "Delete room types"},

		// Reservations
		{"reservations:create", "reservations", "create", "Create reservations"},
		{"reservations:read", "reservations", "read", "View reservations"},
		{"reservations:check_in", "reservations", "check_in", "Check in guests"},
		{"reservations:check_out", "reservations", "check_out", "Check out guests"},
		{"reservations:cancel", "reservations", "cancel", "Cancel reservations"},

		// Amenities
		{"amenities:create", "amenities", "create", "Create amenities"},
		{"amenities:read", "amenities", "read", "View amenities"},
		{"amenities:update", "amenities", "update", "Update amenities"},
		{"amenities:delete", "amenities", "delete", "Delete amenities"},

		// Housekeeping
		{"housekeeping:assign", "housekeeping", "assign", "Assign housekeeping tasks"},
		{"housekeeping:read", "housekeeping", "read", "View housekeeping tasks"},
		{"housekeeping:update", "housekeeping", "update", "Update housekeeping task status"},

		// Menu
		{"menu:create", "menu", "create", "Create menu items"},
		{"menu:read", "menu", "read", "View menu items"},
		{"menu:update", "menu", "update", "Update menu items"},
		{"menu:delete", "menu", "delete", "Delete menu items"},

		// Orders
		{"orders:create", "orders", "create", "Create room service orders"},
		{"orders:read", "orders", "read", "View orders"},
		{"orders:update_status", "orders", "update_status", "Update order status"},
		{"orders:assign", "orders", "assign", "Assign orders to staff"},

		// Activities
		{"activities:create", "activities", "create", "Create activities"},
		{"activities:read", "activities", "read", "View activities"},
		{"activities:update", "activities", "update", "Update activities"},
		{"activities:delete", "activities", "delete", "Delete activities"},

		// Activity Bookings
		{"activity_bookings:create", "activity_bookings", "create", "Create activity bookings"},
		{"activity_bookings:read", "activity_bookings", "read", "View activity bookings"},
		{"activity_bookings:update_status", "activity_bookings", "update_status", "Update activity booking status"},

		// Billing
		{"billing:generate", "billing", "generate", "Generate bills"},
		{"billing:read", "billing", "read", "View bills"},
		{"billing:pay", "billing", "pay", "Process bill payments"},

		// Staff Management
		{"staff:create", "staff", "create", "Create staff members"},
		{"staff:read", "staff", "read", "View staff list"},
		{"staff:update", "staff", "update", "Update staff members"},

		// Guest Settings
		{"guest_settings:read", "guest_settings", "read", "View guest portal settings"},
		{"guest_settings:update", "guest_settings", "update", "Update guest portal settings"},

		// Role Management
		{"roles:create", "roles", "create", "Create custom roles"},
		{"roles:read", "roles", "read", "View roles and permissions"},
		{"roles:update", "roles", "update", "Update roles and permissions"},
		{"roles:delete", "roles", "delete", "Delete custom roles"},

		// Analytics
		{"analytics:view", "analytics", "view", "View analytics and reports"},

		// Hotel Settings
		{"hotels:read", "hotels", "read", "View hotel details"},
		{"hotels:update", "hotels", "update", "Update hotel settings"},
	}
}
