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
}
