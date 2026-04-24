package seeds

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

// SeederRecord tracks which seeders have been applied.
type SeederRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"uniqueIndex;not null;size:100"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

func (SeederRecord) TableName() string {
	return "schema_seeders"
}

type SeederFunc func(db *gorm.DB) error

type seeder struct {
	Name  string
	Apply SeederFunc
}

// seeders returns the ordered list of all seeders.
// Add new seeders to the end with a new name.
func seeders() []seeder {
	return []seeder{
		{
			Name: "20260424_001_super_admin",
			Apply: func(db *gorm.DB) error {
				hash, err := utils.HashPassword("admin@123")
				if err != nil {
					return err
				}

				admin := &models.User{
					Email:        "admin@hotelpms.com",
					PasswordHash: hash,
					Name:         "Super Admin",
					Role:         models.RoleSuperAdmin,
					IsActive:     true,
				}

				if err := db.Create(admin).Error; err != nil {
					return err
				}

				log.Println("Super admin created: admin@hotelpms.com / admin@123")
				return nil
			},
		},
	}
}

func Run(db *gorm.DB) error {
	log.Println("Running database seeders...")

	// Ensure the tracking table exists
	if err := db.AutoMigrate(&SeederRecord{}); err != nil {
		return fmt.Errorf("failed to create seeder tracking table: %w", err)
	}

	// Load already-applied seeders
	var applied []SeederRecord
	if err := db.Find(&applied).Error; err != nil {
		return fmt.Errorf("failed to read seeder records: %w", err)
	}
	appliedSet := make(map[string]bool, len(applied))
	for _, s := range applied {
		appliedSet[s.Name] = true
	}

	allSeeders := seeders()
	sort.Slice(allSeeders, func(i, j int) bool {
		return allSeeders[i].Name < allSeeders[j].Name
	})

	ran := 0
	for _, s := range allSeeders {
		if appliedSet[s.Name] {
			continue
		}

		log.Printf("Running seeder: %s", s.Name)
		if err := s.Apply(db); err != nil {
			return fmt.Errorf("seeder %s failed: %w", s.Name, err)
		}

		if err := db.Create(&SeederRecord{Name: s.Name}).Error; err != nil {
			return fmt.Errorf("failed to record seeder %s: %w", s.Name, err)
		}
		ran++
	}

	if ran == 0 {
		log.Println("No new seeders to apply")
	} else {
		log.Printf("Applied %d seeder(s) successfully", ran)
	}
	return nil
}
