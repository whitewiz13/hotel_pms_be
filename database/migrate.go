package database

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

// MigrationRecord tracks which migrations have been applied.
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Version   string    `gorm:"uniqueIndex;not null;size:100"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

func (MigrationRecord) TableName() string {
	return "schema_migrations"
}

// MigrationFunc is a function that applies a single migration step.
type MigrationFunc func(db *gorm.DB) error

type migration struct {
	Version string
	Apply   MigrationFunc
}

// migrations returns the ordered list of all migrations.
// Add new migrations to the end with a new version string.
func migrations() []migration {
	return []migration{
		{
			Version: "20260424_001_initial_schema",
			Apply: func(db *gorm.DB) error {
				db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

				if err := db.AutoMigrate(
					&models.User{},
					&models.Hotel{},
					&models.Room{},
					&models.Amenity{},
				); err != nil {
					return err
				}

				db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_hotel_room_number ON rooms(hotel_id, room_number) WHERE deleted_at IS NULL`)
				return nil
			},
		},
		{
			Version: "20260425_002_reservations",
			Apply: func(db *gorm.DB) error {
				if err := db.AutoMigrate(
					&models.Reservation{},
					&models.RoomInventory{},
				); err != nil {
					return err
				}

				// Unique constraint: one inventory record per room per date
				db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_room_inventories_room_date ON room_inventories(room_id, date)`)

				// Index for fast availability lookups
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_room_inventories_hotel_date ON room_inventories(hotel_id, date, is_available)`)

				return nil
			},
		},
		{
			Version: "20260425_003_housekeeping",
			Apply: func(db *gorm.DB) error {
				if err := db.AutoMigrate(&models.HousekeepingTask{}); err != nil {
					return err
				}

				// Index for listing active tasks per hotel
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_housekeeping_hotel_status ON housekeeping_tasks(hotel_id, status)`)

				return nil
			},
		},
		{
			Version: "20260425_004_room_service",
			Apply: func(db *gorm.DB) error {
				if err := db.AutoMigrate(
					&models.MenuItem{},
					&models.Order{},
					&models.OrderItem{},
				); err != nil {
					return err
				}

				db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_hotel_status ON orders(hotel_id, status)`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_reservation ON orders(reservation_id)`)

				return nil
			},
		},
		{
			Version: "20260425_005_activities",
			Apply: func(db *gorm.DB) error {
				if err := db.AutoMigrate(
					&models.Activity{},
					&models.ActivityBooking{},
				); err != nil {
					return err
				}

				db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_bookings_hotel_status ON activity_bookings(hotel_id, status)`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_bookings_reservation ON activity_bookings(reservation_id)`)

				return nil
			},
		},
		{
			Version: "20260425_006_billing",
			Apply: func(db *gorm.DB) error {
				if err := db.AutoMigrate(
					&models.Bill{},
					&models.BillLineItem{},
				); err != nil {
					return err
				}

				db.Exec(`CREATE INDEX IF NOT EXISTS idx_bills_hotel_status ON bills(hotel_id, status)`)

				return nil
			},
		},
		{
			Version: "20260426_007_guests",
			Apply: func(db *gorm.DB) error {
				// Create guests table
				if err := db.AutoMigrate(&models.Guest{}); err != nil {
					return err
				}

				// Add guest_id columns as nullable first
				for _, table := range []string{"reservations", "orders", "activity_bookings", "bills"} {
					db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS guest_id UUID`, table))
				}

				// Backfill: create guest records from existing reservations
				db.Exec(`
					INSERT INTO guests (id, hotel_id, name, phone, created_at, updated_at)
					SELECT gen_random_uuid(), hotel_id, guest_name, guest_phone, NOW(), NOW()
					FROM reservations
					WHERE guest_name IS NOT NULL AND guest_name != ''
					ON CONFLICT DO NOTHING
				`)

				// Link reservations to guests by matching hotel_id + phone
				db.Exec(`
					UPDATE reservations r
					SET guest_id = g.id
					FROM guests g
					WHERE r.hotel_id = g.hotel_id AND r.guest_phone = g.phone AND r.guest_id IS NULL
				`)

				// Link orders/activity_bookings/bills via reservation
				db.Exec(`
					UPDATE orders o SET guest_id = r.guest_id
					FROM reservations r WHERE o.reservation_id = r.id AND o.guest_id IS NULL
				`)
				db.Exec(`
					UPDATE activity_bookings ab SET guest_id = r.guest_id
					FROM reservations r WHERE ab.reservation_id = r.id AND ab.guest_id IS NULL
				`)
				db.Exec(`
					UPDATE bills b SET guest_id = r.guest_id
					FROM reservations r WHERE b.reservation_id = r.id AND b.guest_id IS NULL
				`)

				// Now make guest_id NOT NULL (safe after backfill)
				for _, table := range []string{"reservations", "orders", "activity_bookings", "bills"} {
					db.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN guest_id SET NOT NULL`, table))
				}

				db.Exec(`CREATE INDEX IF NOT EXISTS idx_guests_hotel ON guests(hotel_id)`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_guests_hotel_phone ON guests(hotel_id, phone) WHERE deleted_at IS NULL`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_reservations_guest ON reservations(guest_id)`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_orders_guest ON orders(guest_id)`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_activity_bookings_guest ON activity_bookings(guest_id)`)
				db.Exec(`CREATE INDEX IF NOT EXISTS idx_bills_guest ON bills(guest_id)`)

				return nil
			},
		},
	}
}

func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Ensure the tracking table exists
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration tracking table: %w", err)
	}

	// Load already-applied versions
	var applied []MigrationRecord
	if err := db.Find(&applied).Error; err != nil {
		return fmt.Errorf("failed to read migration records: %w", err)
	}
	appliedSet := make(map[string]bool, len(applied))
	for _, m := range applied {
		appliedSet[m.Version] = true
	}

	// Sort migrations by version (they're already ordered, but be safe)
	allMigrations := migrations()
	sort.Slice(allMigrations, func(i, j int) bool {
		return allMigrations[i].Version < allMigrations[j].Version
	})

	ran := 0
	for _, m := range allMigrations {
		if appliedSet[m.Version] {
			continue
		}

		log.Printf("Applying migration: %s", m.Version)
		if err := m.Apply(db); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.Version, err)
		}

		// Record it
		if err := db.Create(&MigrationRecord{Version: m.Version}).Error; err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.Version, err)
		}
		ran++
	}

	if ran == 0 {
		log.Println("No new migrations to apply")
	} else {
		log.Printf("Applied %d migration(s) successfully", ran)
	}
	return nil
}
