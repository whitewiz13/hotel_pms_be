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
	Auth    *handler.AuthHandler
	Hotel   *handler.HotelHandler
	Room    *handler.RoomHandler
	Amenity *handler.AmenityHandler
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
			// Staff management (super_admin + manager)
			protected.POST("/staff", middleware.RequireManagement(), handlers.Auth.CreateStaff)

			// Hotel routes
			hotels := protected.Group("/hotels")
			{
				hotels.POST("", middleware.RequireSuperAdmin(), handlers.Hotel.Create)
				hotels.GET("", middleware.RequireAnyStaff(), handlers.Hotel.GetAll)
				hotels.GET("/:id", middleware.RequireAnyStaff(), handlers.Hotel.GetByID)
				hotels.PUT("/:id", middleware.RequireManagement(), handlers.Hotel.Update)
				hotels.DELETE("/:id", middleware.RequireSuperAdmin(), handlers.Hotel.Delete)

				// Rooms under a hotel
				hotels.GET("/:hotelId/rooms", middleware.RequireAnyStaff(), handlers.Room.GetByHotelID)
			}

			// Room routes
			rooms := protected.Group("/rooms")
			{
				rooms.POST("", middleware.RequireFrontDeskOrAbove(), handlers.Room.Create)
				rooms.GET("/:id", middleware.RequireAnyAuthenticated(), handlers.Room.GetByID)
				rooms.PUT("/:id", middleware.RequireFrontDeskOrAbove(), handlers.Room.Update)
				rooms.DELETE("/:id", middleware.RequireManagement(), handlers.Room.Delete)
				rooms.POST("/:id/pin", middleware.RequireFrontDeskOrAbove(), handlers.Room.GeneratePin)
				rooms.DELETE("/:id/pin", middleware.RequireFrontDeskOrAbove(), handlers.Room.ClearPin)
			}

			// Amenity routes
			amenities := protected.Group("/amenities")
			{
				amenities.POST("", middleware.RequireManagement(), handlers.Amenity.Create)
				amenities.GET("", middleware.RequireAnyAuthenticated(), handlers.Amenity.GetAll)
				amenities.GET("/:id", middleware.RequireAnyAuthenticated(), handlers.Amenity.GetByID)
				amenities.PUT("/:id", middleware.RequireManagement(), handlers.Amenity.Update)
				amenities.DELETE("/:id", middleware.RequireSuperAdmin(), handlers.Amenity.Delete)
			}
		}
	}

	return r
}
