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
	Auth         *handler.AuthHandler
	Hotel        *handler.HotelHandler
	Room         *handler.RoomHandler
	Amenity      *handler.AmenityHandler
	User         *handler.UserHandler
	Reservation  *handler.ReservationHandler
	Housekeeping *handler.HousekeepingHandler
	Dashboard    *handler.DashboardHandler
	Menu         *handler.MenuHandler
	Order        *handler.OrderHandler
	Activity     *handler.ActivityHandler
	Bill         *handler.BillHandler
	Guest        *handler.GuestHandler
	GuestSettings *handler.GuestSettingsHandler
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
				hotel.PUT("", middleware.RequireHotelAdminOrAbove(), handlers.Hotel.Update)
				hotel.DELETE("", middleware.RequireSuperAdmin(), handlers.Hotel.Delete)

				// Staff management
				hotel.POST("/staff", middleware.RequireHotelAdminOrAbove(), handlers.Auth.CreateStaff)
				hotel.GET("/staff", middleware.RequireHotelAdminOrAbove(), handlers.User.GetByHotelID)

				// Rooms
				hotel.POST("/rooms", middleware.RequireHotelFrontDeskOrAbove(), handlers.Room.Create)
				hotel.GET("/rooms", handlers.Room.GetByHotelID)
				hotel.GET("/rooms/:id", handlers.Room.GetByID)
				hotel.PUT("/rooms/:id", middleware.RequireHotelFrontDeskOrAbove(), handlers.Room.Update)
				hotel.DELETE("/rooms/:id", middleware.RequireHotelAdminOrAbove(), handlers.Room.Delete)
				hotel.POST("/rooms/:id/pin", middleware.RequireHotelFrontDeskOrAbove(), handlers.Room.GeneratePin)
				hotel.DELETE("/rooms/:id/pin", middleware.RequireHotelFrontDeskOrAbove(), handlers.Room.ClearPin)

				// Reservations
				hotel.GET("/availability", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.GetAvailability)
				hotel.POST("/reservations", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.Create)
				hotel.GET("/reservations", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.List)
				hotel.GET("/reservations/:id", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.GetByID)
				hotel.POST("/reservations/:id/check-in", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.CheckIn)
				hotel.POST("/reservations/:id/check-out", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.CheckOut)
				hotel.POST("/reservations/:id/cancel", middleware.RequireHotelFrontDeskOrAbove(), handlers.Reservation.Cancel)

				// Amenities
				hotel.POST("/amenities", middleware.RequireHotelAdminOrAbove(), handlers.Amenity.Create)
				hotel.GET("/amenities", handlers.Amenity.GetAll)
				hotel.GET("/amenities/:id", handlers.Amenity.GetByID)
				hotel.PUT("/amenities/:id", middleware.RequireHotelAdminOrAbove(), handlers.Amenity.Update)
				hotel.DELETE("/amenities/:id", middleware.RequireHotelAdminOrAbove(), handlers.Amenity.Delete)

				// Housekeeping
				hotel.POST("/housekeeping", middleware.RequireHotelFrontDeskOrAbove(), handlers.Housekeeping.Assign)
				hotel.GET("/housekeeping", middleware.RequireHousekeepingOrAbove(), handlers.Housekeeping.List)
				hotel.GET("/housekeeping/:id", middleware.RequireHousekeepingOrAbove(), handlers.Housekeeping.GetByID)
				hotel.POST("/housekeeping/:id/start", middleware.RequireHousekeepingOrAbove(), handlers.Housekeeping.Start)
				hotel.POST("/housekeeping/:id/complete", middleware.RequireHousekeepingOrAbove(), handlers.Housekeeping.Complete)

				// Menu (Room Service)
				hotel.POST("/menu", middleware.RequireHotelManagement(), handlers.Menu.Create)
				hotel.GET("/menu", middleware.RequireAnyAuthenticated(), handlers.Menu.List)
				hotel.GET("/menu/:id", middleware.RequireAnyAuthenticated(), handlers.Menu.GetByID)
				hotel.PUT("/menu/:id", middleware.RequireHotelManagement(), handlers.Menu.Update)
				hotel.DELETE("/menu/:id", middleware.RequireHotelManagement(), handlers.Menu.Delete)

				// Orders (Room Service)
				hotel.POST("/orders", middleware.RequireHotelFrontDeskOrAbove(), handlers.Order.Create)
				hotel.GET("/orders", middleware.RequireRoomServiceOrAbove(), handlers.Order.List)
				hotel.GET("/orders/:id", middleware.RequireRoomServiceOrAbove(), handlers.Order.GetByID)
				hotel.POST("/orders/:id/status", middleware.RequireRoomServiceOrAbove(), handlers.Order.UpdateStatus)
				hotel.POST("/orders/:id/assign", middleware.RequireHotelFrontDeskOrAbove(), handlers.Order.Assign)

				// Activities
				hotel.POST("/activities", middleware.RequireHotelManagement(), handlers.Activity.Create)
				hotel.GET("/activities", middleware.RequireAnyAuthenticated(), handlers.Activity.List)
				hotel.GET("/activities/:id", middleware.RequireAnyAuthenticated(), handlers.Activity.GetByID)
				hotel.PUT("/activities/:id", middleware.RequireHotelManagement(), handlers.Activity.Update)
				hotel.DELETE("/activities/:id", middleware.RequireHotelManagement(), handlers.Activity.Delete)

				// Activity Bookings
				hotel.POST("/activity-bookings", middleware.RequireHotelFrontDeskOrAbove(), handlers.Activity.CreateBooking)
				hotel.GET("/activity-bookings", middleware.RequireHotelFrontDeskOrAbove(), handlers.Activity.ListBookings)
				hotel.GET("/activity-bookings/:id", middleware.RequireHotelFrontDeskOrAbove(), handlers.Activity.GetBookingByID)
				hotel.POST("/activity-bookings/:id/status", middleware.RequireHotelFrontDeskOrAbove(), handlers.Activity.UpdateBookingStatus)

				// Billing
				hotel.POST("/reservations/:id/bill", middleware.RequireHotelFrontDeskOrAbove(), handlers.Bill.Generate)
				hotel.GET("/reservations/:id/bill", middleware.RequireHotelFrontDeskOrAbove(), handlers.Bill.GetByReservation)
				hotel.GET("/bills", middleware.RequireHotelFrontDeskOrAbove(), handlers.Bill.List)
				hotel.GET("/bills/:id", middleware.RequireHotelFrontDeskOrAbove(), handlers.Bill.GetByID)
				hotel.POST("/bills/:id/pay", middleware.RequireHotelFrontDeskOrAbove(), handlers.Bill.MarkPaid)

				// Dashboard
				hotel.GET("/dashboard/stats", middleware.RequireHotelFrontDeskOrAbove(), handlers.Dashboard.GetStats)
				hotel.GET("/activity", middleware.RequireHotelFrontDeskOrAbove(), handlers.Dashboard.GetActivity)

				// Guest Settings (admin)
				hotel.POST("/guest-settings", middleware.RequireHotelAdminOrAbove(), handlers.GuestSettings.Save)
				hotel.GET("/guest-settings", middleware.RequireHotelAdminOrAbove(), handlers.GuestSettings.Get)
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
