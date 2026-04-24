package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get(ContextKeyClaims)
		if !exists {
			utils.RespondUnauthorized(c, "authentication required")
			c.Abort()
			return
		}

		claims, ok := claimsVal.(*service.JWTClaims)
		if !ok {
			utils.RespondUnauthorized(c, "invalid token claims")
			c.Abort()
			return
		}

		for _, role := range roles {
			if claims.Role == role {
				c.Next()
				return
			}
		}

		// Also check for guest role
		if claims.IsGuest {
			for _, role := range roles {
				if role == "guest" {
					c.Next()
					return
				}
			}
		}

		utils.RespondForbidden(c, "insufficient permissions")
		c.Abort()
	}
}

// ---------------------------------------------------------------------------
// Hotel access middleware – verifies the user belongs to the hotel in the URL
// ---------------------------------------------------------------------------

func HotelAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			utils.RespondUnauthorized(c, "authentication required")
			c.Abort()
			return
		}

		// Super admin has access to all hotels
		if claims.Role == models.RoleSuperAdmin {
			c.Next()
			return
		}

		// For other roles, hotel_id in URL must match their hotel_id
		hotelID := c.Param("hotel_id")
		if claims.HotelID != hotelID {
			utils.RespondForbidden(c, "access denied to this hotel")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ---------------------------------------------------------------------------
// Convenience helpers
// ---------------------------------------------------------------------------

func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin)
}

// RequireHotelAdminOrAbove = super_admin + hotel_admin
func RequireHotelAdminOrAbove() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin, models.RoleHotelAdmin)
}

// RequireHotelManagement = super_admin + hotel_admin + manager
func RequireHotelManagement() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin, models.RoleHotelAdmin, models.RoleManager)
}

// RequireHotelFrontDeskOrAbove = super_admin + hotel_admin + manager + front_desk
func RequireHotelFrontDeskOrAbove() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin, models.RoleHotelAdmin, models.RoleManager, models.RoleFrontDesk)
}

// RequireAnyStaff = every role in ValidRoles (all non-guest authenticated users)
func RequireAnyStaff() gin.HandlerFunc {
	return RequireRole(models.ValidRoles...)
}

// RequireAnyAuthenticated = any staff role + guest
func RequireAnyAuthenticated() gin.HandlerFunc {
	roles := append([]models.UserRole{}, models.ValidRoles...)
	roles = append(roles, models.RoleGuest)
	return RequireRole(roles...)
}

func GetClaims(c *gin.Context) *service.JWTClaims {
	claimsVal, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil
	}
	claims, ok := claimsVal.(*service.JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
