package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/handler"
	"github.com/hotelpms/backend/middleware"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/repository"
	"github.com/hotelpms/backend/service"
)

type Handlers struct {
	Auth          *handler.AuthHandler
	Hotel         *handler.HotelHandler
	Room          *handler.RoomHandler
	RoomType      *handler.RoomTypeHandler
	Amenity       *handler.AmenityHandler
	User          *handler.UserHandler
	Reservation   *handler.ReservationHandler
	Housekeeping  *handler.HousekeepingHandler
	Dashboard     *handler.DashboardHandler
	Menu          *handler.MenuHandler
	Order         *handler.OrderHandler
	Activity      *handler.ActivityHandler
	Bill          *handler.BillHandler
	Guest         *handler.GuestHandler
	GuestSettings *handler.GuestSettingsHandler
	Role          *handler.RoleHandler
	Analytics     *handler.AnalyticsHandler
	Upload        *handler.UploadHandler
	FCMToken      *handler.FCMTokenHandler
	Plan          *handler.PlanHandler
}

func Setup(handlers *Handlers, authService *service.AuthService, roleRepo *repository.RoleRepository, planService *service.PlanService, uploadDir string) *gin.Engine {
	r := gin.Default()

	// Serve uploaded files
	r.Static("/uploads", uploadDir)

	// Shorthand for permission middleware
	perm := func(permissions ...string) gin.HandlerFunc {
		return middleware.RequirePermission(roleRepo, permissions...)
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173","http://192.168.29.194:5173", "https://wizhotelpms.netlify.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "hotel-pms"})
		})

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Auth.Login)
			auth.POST("/guest/login", handlers.Auth.GuestLogin)
		}

		// Public hotel lookup by slug (for guest portal URL)
		api.GET("/hotels/slug/:slug", handlers.Hotel.GetBySlug)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Current user info with live permissions
			protected.GET("/auth/me", handlers.Auth.Me)

			// FCM device tokens
			protected.POST("/fcm-tokens", handlers.FCMToken.SaveToken)
			protected.DELETE("/fcm-tokens", handlers.FCMToken.DeleteToken)

				// Plans (public list)
			protected.GET("/plans", handlers.Plan.GetAllPlans)

			// Super admin: hotel management
			hotels := protected.Group("/hotels")
			{
				hotels.POST("", middleware.RequireSuperAdmin(), handlers.Hotel.Create)
				hotels.GET("", middleware.RequireSuperAdmin(), handlers.Hotel.GetAll)
			}

			protected.GET("/users", middleware.RequireSuperAdmin(), handlers.User.GetAll)

			// Hotel-scoped routes
			hotel := protected.Group("/hotels/:hotel_id")
			hotel.Use(middleware.HotelAccessMiddleware())
			{
				hotel.GET("", handlers.Hotel.GetByID)
				hotel.PUT("", perm("hotels:update"), handlers.Hotel.Update)
				hotel.DELETE("", middleware.RequireSuperAdmin(), handlers.Hotel.Delete)

				// Subscription management
				hotel.GET("/subscription", handlers.Plan.GetHotelSubscription)
				hotel.PUT("/subscription", middleware.RequireSuperAdmin(), handlers.Plan.ChangeHotelPlan)
				hotel.GET("/subscription/usage", handlers.Plan.GetHotelPlanUsage)

				// Staff management
				hotel.POST("/staff", perm("staff:create"), handlers.Auth.CreateStaff)
				hotel.GET("/staff", perm("staff:read"), handlers.User.GetByHotelID)

				// Room Types
				hotel.POST("/room-types", perm("room_types:create"), handlers.RoomType.Create)
				hotel.GET("/room-types", handlers.RoomType.GetAll)
				hotel.GET("/room-types/:id", handlers.RoomType.GetByID)
				hotel.PUT("/room-types/:id", perm("room_types:update"), handlers.RoomType.Update)
				hotel.DELETE("/room-types/:id", perm("room_types:delete"), handlers.RoomType.Delete)

				// Rooms
				hotel.POST("/rooms", perm("rooms:create"), handlers.Room.Create)
				hotel.GET("/rooms", handlers.Room.GetByHotelID)
				hotel.GET("/rooms/:id", handlers.Room.GetByID)
				hotel.PUT("/rooms/:id", perm("rooms:update"), handlers.Room.Update)
				hotel.DELETE("/rooms/:id", perm("rooms:delete"), handlers.Room.Delete)
				hotel.POST("/rooms/:id/pin", perm("rooms:manage_pin"), handlers.Room.GeneratePin)
				hotel.DELETE("/rooms/:id/pin", perm("rooms:manage_pin"), handlers.Room.ClearPin)

				// Reservations
				hotel.GET("/availability", perm("reservations:read"), handlers.Reservation.GetAvailability)
				hotel.POST("/reservations", perm("reservations:create"), handlers.Reservation.Create)
				hotel.GET("/reservations", perm("reservations:read"), handlers.Reservation.List)
				hotel.GET("/reservations/:id", perm("reservations:read"), handlers.Reservation.GetByID)
				hotel.POST("/reservations/:id/check-in", perm("reservations:check_in"), handlers.Reservation.CheckIn)
				hotel.POST("/reservations/:id/check-out", perm("reservations:check_out"), handlers.Reservation.CheckOut)
				hotel.POST("/reservations/:id/cancel", perm("reservations:cancel"), handlers.Reservation.Cancel)

				// Amenities
				hotel.POST("/amenities", perm("amenities:create"), handlers.Amenity.Create)
				hotel.GET("/amenities", handlers.Amenity.GetAll)
				hotel.GET("/amenities/:id", handlers.Amenity.GetByID)
				hotel.PUT("/amenities/:id", perm("amenities:update"), handlers.Amenity.Update)
				hotel.DELETE("/amenities/:id", perm("amenities:delete"), handlers.Amenity.Delete)

				// Housekeeping
				hotel.POST("/housekeeping", perm("housekeeping:assign"), handlers.Housekeeping.Assign)
				hotel.GET("/housekeeping", perm("housekeeping:read"), handlers.Housekeeping.List)
				hotel.GET("/housekeeping/:id", perm("housekeeping:read"), handlers.Housekeeping.GetByID)
				hotel.POST("/housekeeping/:id/start", perm("housekeeping:update"), handlers.Housekeeping.Start)
				hotel.POST("/housekeeping/:id/complete", perm("housekeeping:update"), handlers.Housekeeping.Complete)

				// Menu (Room Service) — requires room_service feature
				menu := hotel.Group("", middleware.RequireFeature(planService, "room_service"))
				{
					menu.POST("/menu", perm("menu:create"), handlers.Menu.Create)
					menu.GET("/menu", perm("menu:read"), handlers.Menu.List)
					menu.GET("/menu/:id", perm("menu:read"), handlers.Menu.GetByID)
					menu.PUT("/menu/:id", perm("menu:update"), handlers.Menu.Update)
					menu.DELETE("/menu/:id", perm("menu:delete"), handlers.Menu.Delete)
				}

				// Orders (Room Service) — requires room_service feature
				orders := hotel.Group("", middleware.RequireFeature(planService, "room_service"))
				{
					orders.POST("/orders", perm("orders:create"), handlers.Order.Create)
					orders.GET("/orders", perm("orders:read"), handlers.Order.List)
					orders.GET("/orders/:id", perm("orders:read"), handlers.Order.GetByID)
					orders.POST("/orders/:id/status", perm("orders:update_status"), handlers.Order.UpdateStatus)
					orders.POST("/orders/:id/assign", perm("orders:assign"), handlers.Order.Assign)
				}

				// Activities — requires activities feature
				activities := hotel.Group("", middleware.RequireFeature(planService, "activities"))
				{
					activities.POST("/activities", perm("activities:create"), handlers.Activity.Create)
					activities.GET("/activities", perm("activities:read"), handlers.Activity.List)
					activities.GET("/activities/:id", perm("activities:read"), handlers.Activity.GetByID)
					activities.PUT("/activities/:id", perm("activities:update"), handlers.Activity.Update)
					activities.DELETE("/activities/:id", perm("activities:delete"), handlers.Activity.Delete)
				}

				// Activity Bookings — requires activities feature
				activityBookings := hotel.Group("", middleware.RequireFeature(planService, "activities"))
				{
					activityBookings.POST("/activity-bookings", perm("activity_bookings:create"), handlers.Activity.CreateBooking)
					activityBookings.GET("/activity-bookings", perm("activity_bookings:read"), handlers.Activity.ListBookings)
					activityBookings.GET("/activity-bookings/:id", perm("activity_bookings:read"), handlers.Activity.GetBookingByID)
					activityBookings.POST("/activity-bookings/:id/status", perm("activity_bookings:update_status"), handlers.Activity.UpdateBookingStatus)
				}

				// Billing
				hotel.POST("/reservations/:id/bill", perm("billing:generate"), handlers.Bill.Generate)
				hotel.GET("/reservations/:id/bill", perm("billing:read"), handlers.Bill.GetByReservation)
				hotel.GET("/bills", perm("billing:read"), handlers.Bill.List)
				hotel.GET("/bills/:id", perm("billing:read"), handlers.Bill.GetByID)
				hotel.POST("/bills/:id/pay", perm("billing:pay"), handlers.Bill.MarkPaid)

				// Dashboard
				hotel.GET("/dashboard/stats", perm("dashboard:view"), handlers.Dashboard.GetStats)
				hotel.GET("/activity", perm("dashboard:view"), handlers.Dashboard.GetActivity)

				// Analytics — requires analytics feature
				analytics := hotel.Group("/analytics", middleware.RequireFeature(planService, "analytics"))
				{
					analytics.GET("/summary", perm("analytics:view"), handlers.Analytics.GetSummary)
					analytics.GET("/occupancy", perm("analytics:view"), middleware.RequireFeature(planService, "adv_analytics"), handlers.Analytics.GetOccupancyTrend)
					analytics.GET("/revenue", perm("analytics:view"), middleware.RequireFeature(planService, "adv_analytics"), handlers.Analytics.GetRevenueTrend)
					analytics.GET("/reservations", perm("analytics:view"), middleware.RequireFeature(planService, "adv_analytics"), handlers.Analytics.GetReservationStats)
					analytics.GET("/room-types", perm("analytics:view"), middleware.RequireFeature(planService, "adv_analytics"), handlers.Analytics.GetRoomTypePerformance)
				}

				// Guest Settings
				hotel.POST("/guest-settings", perm("guest_settings:update"), handlers.GuestSettings.Save)
				hotel.GET("/guest-settings", perm("guest_settings:read"), handlers.GuestSettings.Get)

				// File Uploads — requires guest_uploads feature
				hotel.POST("/uploads", perm("reservations:check_in"), middleware.RequireFeature(planService, "guest_uploads"), handlers.Upload.Upload)

				// Roles & Permissions
				hotel.GET("/permissions", perm("roles:read"), handlers.Role.GetPermissions)
				hotel.GET("/roles", perm("roles:read"), handlers.Role.GetAll)
				hotel.GET("/roles/:id", perm("roles:read"), handlers.Role.GetByID)
				// Custom role management — requires custom_roles feature
				hotel.POST("/roles", perm("roles:create"), middleware.RequireFeature(planService, "custom_roles"), handlers.Role.Create)
				hotel.PUT("/roles/:id", perm("roles:update"), middleware.RequireFeature(planService, "custom_roles"), handlers.Role.Update)
				hotel.DELETE("/roles/:id", perm("roles:delete"), middleware.RequireFeature(planService, "custom_roles"), handlers.Role.Delete)
			}

				// Guest self-service (authenticated guests only) — requires guest_portal feature
			guest := protected.Group("/guest")
			guest.Use(middleware.RequireRole(models.RoleGuest))
			guest.Use(middleware.RequireFeature(planService, "guest_portal"))
			{
				guest.GET("/reservation", handlers.Guest.GetMyReservation)
				guest.GET("/menu", handlers.Guest.ListMenu)
				guest.GET("/activities", handlers.Guest.ListActivities)
				guest.POST("/orders", handlers.Guest.PlaceOrder)
				guest.GET("/orders", handlers.Guest.ListMyOrders)
				guest.POST("/activity-bookings", handlers.Guest.BookActivity)
				guest.GET("/activity-bookings", handlers.Guest.ListMyActivityBookings)
				guest.GET("/dashboard", handlers.Guest.GetDashboard)
				guest.GET("/settings", handlers.GuestSettings.GetForGuest)
			}
		}
	}

	return r
}
