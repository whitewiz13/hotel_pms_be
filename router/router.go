package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/handler"
	"github.com/hotelpms/backend/middleware"
	"github.com/hotelpms/backend/models"
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
}

func Setup(handlers *Handlers, authService *service.AuthService) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
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
				hotel.PUT("", middleware.RequirePermission("hotels:update"), handlers.Hotel.Update)
				hotel.DELETE("", middleware.RequireSuperAdmin(), handlers.Hotel.Delete)

				// Staff management
				hotel.POST("/staff", middleware.RequirePermission("staff:create"), handlers.Auth.CreateStaff)
				hotel.GET("/staff", middleware.RequirePermission("staff:read"), handlers.User.GetByHotelID)

				// Room Types
				hotel.POST("/room-types", middleware.RequirePermission("room_types:create"), handlers.RoomType.Create)
				hotel.GET("/room-types", handlers.RoomType.GetAll)
				hotel.GET("/room-types/:id", handlers.RoomType.GetByID)
				hotel.PUT("/room-types/:id", middleware.RequirePermission("room_types:update"), handlers.RoomType.Update)
				hotel.DELETE("/room-types/:id", middleware.RequirePermission("room_types:delete"), handlers.RoomType.Delete)

				// Rooms
				hotel.POST("/rooms", middleware.RequirePermission("rooms:create"), handlers.Room.Create)
				hotel.GET("/rooms", handlers.Room.GetByHotelID)
				hotel.GET("/rooms/:id", handlers.Room.GetByID)
				hotel.PUT("/rooms/:id", middleware.RequirePermission("rooms:update"), handlers.Room.Update)
				hotel.DELETE("/rooms/:id", middleware.RequirePermission("rooms:delete"), handlers.Room.Delete)
				hotel.POST("/rooms/:id/pin", middleware.RequirePermission("rooms:manage_pin"), handlers.Room.GeneratePin)
				hotel.DELETE("/rooms/:id/pin", middleware.RequirePermission("rooms:manage_pin"), handlers.Room.ClearPin)

				// Reservations
				hotel.GET("/availability", middleware.RequirePermission("reservations:read"), handlers.Reservation.GetAvailability)
				hotel.POST("/reservations", middleware.RequirePermission("reservations:create"), handlers.Reservation.Create)
				hotel.GET("/reservations", middleware.RequirePermission("reservations:read"), handlers.Reservation.List)
				hotel.GET("/reservations/:id", middleware.RequirePermission("reservations:read"), handlers.Reservation.GetByID)
				hotel.POST("/reservations/:id/check-in", middleware.RequirePermission("reservations:check_in"), handlers.Reservation.CheckIn)
				hotel.POST("/reservations/:id/check-out", middleware.RequirePermission("reservations:check_out"), handlers.Reservation.CheckOut)
				hotel.POST("/reservations/:id/cancel", middleware.RequirePermission("reservations:cancel"), handlers.Reservation.Cancel)

				// Amenities
				hotel.POST("/amenities", middleware.RequirePermission("amenities:create"), handlers.Amenity.Create)
				hotel.GET("/amenities", handlers.Amenity.GetAll)
				hotel.GET("/amenities/:id", handlers.Amenity.GetByID)
				hotel.PUT("/amenities/:id", middleware.RequirePermission("amenities:update"), handlers.Amenity.Update)
				hotel.DELETE("/amenities/:id", middleware.RequirePermission("amenities:delete"), handlers.Amenity.Delete)

				// Housekeeping
				hotel.POST("/housekeeping", middleware.RequirePermission("housekeeping:assign"), handlers.Housekeeping.Assign)
				hotel.GET("/housekeeping", middleware.RequirePermission("housekeeping:read"), handlers.Housekeeping.List)
				hotel.GET("/housekeeping/:id", middleware.RequirePermission("housekeeping:read"), handlers.Housekeeping.GetByID)
				hotel.POST("/housekeeping/:id/start", middleware.RequirePermission("housekeeping:update"), handlers.Housekeeping.Start)
				hotel.POST("/housekeeping/:id/complete", middleware.RequirePermission("housekeeping:update"), handlers.Housekeeping.Complete)

				// Menu (Room Service)
				hotel.POST("/menu", middleware.RequirePermission("menu:create"), handlers.Menu.Create)
				hotel.GET("/menu", middleware.RequirePermission("menu:read"), handlers.Menu.List)
				hotel.GET("/menu/:id", middleware.RequirePermission("menu:read"), handlers.Menu.GetByID)
				hotel.PUT("/menu/:id", middleware.RequirePermission("menu:update"), handlers.Menu.Update)
				hotel.DELETE("/menu/:id", middleware.RequirePermission("menu:delete"), handlers.Menu.Delete)

				// Orders (Room Service)
				hotel.POST("/orders", middleware.RequirePermission("orders:create"), handlers.Order.Create)
				hotel.GET("/orders", middleware.RequirePermission("orders:read"), handlers.Order.List)
				hotel.GET("/orders/:id", middleware.RequirePermission("orders:read"), handlers.Order.GetByID)
				hotel.POST("/orders/:id/status", middleware.RequirePermission("orders:update_status"), handlers.Order.UpdateStatus)
				hotel.POST("/orders/:id/assign", middleware.RequirePermission("orders:assign"), handlers.Order.Assign)

				// Activities
				hotel.POST("/activities", middleware.RequirePermission("activities:create"), handlers.Activity.Create)
				hotel.GET("/activities", middleware.RequirePermission("activities:read"), handlers.Activity.List)
				hotel.GET("/activities/:id", middleware.RequirePermission("activities:read"), handlers.Activity.GetByID)
				hotel.PUT("/activities/:id", middleware.RequirePermission("activities:update"), handlers.Activity.Update)
				hotel.DELETE("/activities/:id", middleware.RequirePermission("activities:delete"), handlers.Activity.Delete)

				// Activity Bookings
				hotel.POST("/activity-bookings", middleware.RequirePermission("activity_bookings:create"), handlers.Activity.CreateBooking)
				hotel.GET("/activity-bookings", middleware.RequirePermission("activity_bookings:read"), handlers.Activity.ListBookings)
				hotel.GET("/activity-bookings/:id", middleware.RequirePermission("activity_bookings:read"), handlers.Activity.GetBookingByID)
				hotel.POST("/activity-bookings/:id/status", middleware.RequirePermission("activity_bookings:update_status"), handlers.Activity.UpdateBookingStatus)

				// Billing
				hotel.POST("/reservations/:id/bill", middleware.RequirePermission("billing:generate"), handlers.Bill.Generate)
				hotel.GET("/reservations/:id/bill", middleware.RequirePermission("billing:read"), handlers.Bill.GetByReservation)
				hotel.GET("/bills", middleware.RequirePermission("billing:read"), handlers.Bill.List)
				hotel.GET("/bills/:id", middleware.RequirePermission("billing:read"), handlers.Bill.GetByID)
				hotel.POST("/bills/:id/pay", middleware.RequirePermission("billing:pay"), handlers.Bill.MarkPaid)

				// Dashboard
				hotel.GET("/dashboard/stats", middleware.RequirePermission("dashboard:view"), handlers.Dashboard.GetStats)
				hotel.GET("/activity", middleware.RequirePermission("dashboard:view"), handlers.Dashboard.GetActivity)

				// Guest Settings
				hotel.POST("/guest-settings", middleware.RequirePermission("guest_settings:update"), handlers.GuestSettings.Save)
				hotel.GET("/guest-settings", middleware.RequirePermission("guest_settings:read"), handlers.GuestSettings.Get)

				// Roles & Permissions
				hotel.GET("/permissions", middleware.RequirePermission("roles:read"), handlers.Role.GetPermissions)
				hotel.POST("/roles", middleware.RequirePermission("roles:create"), handlers.Role.Create)
				hotel.GET("/roles", middleware.RequirePermission("roles:read"), handlers.Role.GetAll)
				hotel.GET("/roles/:id", middleware.RequirePermission("roles:read"), handlers.Role.GetByID)
				hotel.PUT("/roles/:id", middleware.RequirePermission("roles:update"), handlers.Role.Update)
				hotel.DELETE("/roles/:id", middleware.RequirePermission("roles:delete"), handlers.Role.Delete)
			}

			// Guest self-service (authenticated guests only)
			guest := protected.Group("/guest")
			guest.Use(middleware.RequireRole(models.RoleGuest))
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
