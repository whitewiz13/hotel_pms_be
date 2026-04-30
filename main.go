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
	roomTypeRepo := repository.NewRoomTypeRepository(db)
	amenityRepo := repository.NewAmenityRepository(db)
	reservationRepo := repository.NewReservationRepository(db)
	housekeepingRepo := repository.NewHousekeepingRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	billRepo := repository.NewBillRepository(db)
	guestRepo := repository.NewGuestRepository(db)
	guestSettingsRepo := repository.NewGuestSettingsRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	fcmTokenRepo := repository.NewFCMTokenRepository(db)

	// Initialize notification service
	notificationService := service.NewNotificationService(cfg.Firebase.CredentialsFile, cfg.Firebase.CredentialsJSON, fcmTokenRepo)

	// Initialize services
	roleService := service.NewRoleService(roleRepo, permissionRepo)
	authService := service.NewAuthService(userRepo, roomRepo, hotelRepo, roleRepo, cfg.JWT)
	hotelService := service.NewHotelService(db, hotelRepo, userRepo, roleRepo, permissionRepo)
	roomService := service.NewRoomService(roomRepo, amenityRepo)
	roomTypeService := service.NewRoomTypeService(roomTypeRepo)
	amenityService := service.NewAmenityService(amenityRepo)
	userService := service.NewUserService(userRepo)
	reservationService := service.NewReservationService(db, reservationRepo, roomRepo, billRepo, guestRepo, notificationService)
	housekeepingService := service.NewHousekeepingService(db, housekeepingRepo, roomRepo, userRepo, notificationService)
	dashboardService := service.NewDashboardService(dashboardRepo)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	menuService := service.NewMenuService(menuRepo)
	orderService := service.NewOrderService(db, orderRepo, menuRepo, roomRepo, reservationRepo, userRepo, notificationService)
	activityService := service.NewActivityService(activityRepo, roomRepo, reservationRepo, notificationService)
	billService := service.NewBillService(db, billRepo, reservationRepo, orderRepo, activityRepo, roomRepo, notificationService)
	guestService := service.NewGuestService(reservationRepo, orderService, activityService, menuService, orderRepo, activityRepo, guestSettingsRepo)
	guestSettingsService := service.NewGuestSettingsService(guestSettingsRepo)

	// Initialize upload service
	uploadService := service.NewUploadService(cfg.Upload.Dir, cfg.Upload.BaseURL)

	// Initialize handlers
	handlers := &router.Handlers{
		Auth:          handler.NewAuthHandler(authService),
		Hotel:         handler.NewHotelHandler(hotelService),
		Room:          handler.NewRoomHandler(roomService),
		RoomType:      handler.NewRoomTypeHandler(roomTypeService),
		Amenity:       handler.NewAmenityHandler(amenityService),
		User:         handler.NewUserHandler(userService),
		Reservation:  handler.NewReservationHandler(reservationService),
		Housekeeping: handler.NewHousekeepingHandler(housekeepingService, roleRepo),
		Dashboard:    handler.NewDashboardHandler(dashboardService),
		Menu:         handler.NewMenuHandler(menuService),
		Order:        handler.NewOrderHandler(orderService, roleRepo),
		Activity:     handler.NewActivityHandler(activityService),
		Bill:         handler.NewBillHandler(billService),
		Guest:        handler.NewGuestHandler(guestService),
		GuestSettings: handler.NewGuestSettingsHandler(guestSettingsService),
		Role:          handler.NewRoleHandler(roleService),
		Analytics:     handler.NewAnalyticsHandler(analyticsService),
		FCMToken:      handler.NewFCMTokenHandler(fcmTokenRepo),
	}

	handlers.Upload = handler.NewUploadHandler(uploadService)

	// Setup router
	r := router.Setup(handlers, authService, roleRepo, cfg.Upload.Dir)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Hotel PMS backend starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
