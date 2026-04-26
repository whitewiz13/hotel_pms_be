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

// RequirePermission checks if the authenticated user has at least one of the
// given permission codes. Super admins bypass all permission checks.
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			utils.RespondUnauthorized(c, "authentication required")
			c.Abort()
			return
		}

		// Super admin bypasses all permission checks
		if claims.Role == models.RoleSuperAdmin {
			c.Next()
			return
		}

		permSet := make(map[string]bool, len(claims.Permissions))
		for _, p := range claims.Permissions {
			permSet[p] = true
		}

		for _, required := range permissions {
			if permSet[required] {
				c.Next()
				return
			}
		}

		utils.RespondForbidden(c, "insufficient permissions")
		c.Abort()
	}
}
