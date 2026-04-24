package seeds

import (
	"log"

	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/utils"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	log.Println("Running database seeders...")

	if err := seedSuperAdmin(db); err != nil {
		return err
	}

	if err := seedAmenities(db); err != nil {
		return err
	}

	log.Println("Seeding completed successfully")
	return nil
}

func seedSuperAdmin(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count)
	if count > 0 {
		log.Println("Super admin already exists, skipping...")
		return nil
	}

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
}

func seedAmenities(db *gorm.DB) error {
	var count int64
	db.Model(&models.Amenity{}).Count(&count)
	if count > 0 {
		log.Println("Amenities already exist, skipping...")
		return nil
	}

	amenities := []models.Amenity{
		{Name: "WiFi", Description: "High-speed wireless internet", Icon: "wifi", Category: models.AmenityCategoryRoom, IsActive: true},
		{Name: "Air Conditioning", Description: "Climate control system", Icon: "ac", Category: models.AmenityCategoryRoom, IsActive: true},
		{Name: "TV", Description: "Flat-screen television", Icon: "tv", Category: models.AmenityCategoryRoom, IsActive: true},
		{Name: "Mini Bar", Description: "In-room mini bar with beverages", Icon: "minibar", Category: models.AmenityCategoryRoom, IsActive: true},
		{Name: "Safe", Description: "In-room safe for valuables", Icon: "safe", Category: models.AmenityCategoryRoom, IsActive: true},
		{Name: "Bathtub", Description: "Full-size bathtub", Icon: "bathtub", Category: models.AmenityCategoryBathroom, IsActive: true},
		{Name: "Rain Shower", Description: "Rain shower head", Icon: "shower", Category: models.AmenityCategoryBathroom, IsActive: true},
		{Name: "Hair Dryer", Description: "Professional hair dryer", Icon: "hairdryer", Category: models.AmenityCategoryBathroom, IsActive: true},
		{Name: "Swimming Pool", Description: "Outdoor swimming pool", Icon: "pool", Category: models.AmenityCategoryRecreation, IsActive: true},
		{Name: "Gym", Description: "Fully equipped fitness center", Icon: "gym", Category: models.AmenityCategoryRecreation, IsActive: true},
		{Name: "Spa", Description: "Full-service spa", Icon: "spa", Category: models.AmenityCategoryRecreation, IsActive: true},
		{Name: "Restaurant", Description: "On-site dining", Icon: "restaurant", Category: models.AmenityCategoryDining, IsActive: true},
		{Name: "Room Service", Description: "24-hour room service", Icon: "roomservice", Category: models.AmenityCategoryDining, IsActive: true},
		{Name: "Parking", Description: "On-site parking", Icon: "parking", Category: models.AmenityCategoryGeneral, IsActive: true},
		{Name: "Laundry", Description: "Laundry and dry cleaning service", Icon: "laundry", Category: models.AmenityCategoryGeneral, IsActive: true},
	}

	for i := range amenities {
		if err := db.Create(&amenities[i]).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded %d amenities", len(amenities))
	return nil
}
