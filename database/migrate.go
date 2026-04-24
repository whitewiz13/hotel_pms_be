package database

import (
	"log"

	"github.com/hotelpms/backend/models"
	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Enable uuid-ossp extension for UUID generation
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	err := db.AutoMigrate(
		&models.User{},
		&models.Hotel{},
		&models.Room{},
		&models.Amenity{},
	)
	if err != nil {
		return err
	}

	// Add unique composite index for room_number + hotel_id
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_hotel_room_number ON rooms(hotel_id, room_number) WHERE deleted_at IS NULL`)

	log.Println("Migrations completed successfully")
	return nil
}
