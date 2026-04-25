package router

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/handler"
	"github.com/hotelpms/backend/middleware"
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

				// Dashboard
				hotel.GET("/dashboard/stats", middleware.RequireHotelFrontDeskOrAbove(), handlers.Dashboard.GetStats)
				hotel.GET("/activity", middleware.RequireHotelFrontDeskOrAbove(), handlers.Dashboard.GetActivity)
			}
		}
	}

	return r
}
