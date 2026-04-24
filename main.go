package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/config"
	"github.com/hotelpms/backend/database"
	"github.com/hotelpms/backend/handler"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/router"
	"github.com/hotelpms/backend/seeds"
	"github.com/hotelpms/backend/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	gin.SetMode(cfg.Server.GinMode)

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if config.ShouldRunMigrations() {
		if err := database.RunMigrations(db); err != nil {
			log.Fatal("Failed to run migrations:", err)
		}
	}

	// Run seeders
	if config.ShouldSeedDB() {
		if err := seeds.Run(db); err != nil {
			log.Fatal("Failed to seed database:", err)
		}
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	hotelRepo := repository.NewHotelRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	amenityRepo := repository.NewAmenityRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, roomRepo, cfg.JWT)
	hotelService := service.NewHotelService(db, hotelRepo, userRepo)
	roomService := service.NewRoomService(roomRepo, amenityRepo)
	amenityService := service.NewAmenityService(amenityRepo)
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	handlers := &router.Handlers{
		Auth:    handler.NewAuthHandler(authService),
		Hotel:   handler.NewHotelHandler(hotelService),
		Room:    handler.NewRoomHandler(roomService),
		Amenity: handler.NewAmenityHandler(amenityService),
		User:    handler.NewUserHandler(userService),
	}

	// Setup router
	r := router.Setup(handlers, authService)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Hotel PMS backend starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
