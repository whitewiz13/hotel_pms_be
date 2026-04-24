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
// Convenience helpers – update these when you add/remove roles in models.ValidRoles
// ---------------------------------------------------------------------------

func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin)
}

// Management = super_admin + manager
func RequireManagement() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin, models.RoleManager)
}

// FrontDeskOrAbove = super_admin + manager + front_desk
func RequireFrontDeskOrAbove() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin, models.RoleManager, models.RoleFrontDesk)
}

// AnyStaff = every role in ValidRoles (all non-guest authenticated users)
func RequireAnyStaff() gin.HandlerFunc {
	return RequireRole(models.ValidRoles...)
}

// AnyAuthenticated = any staff role + guest
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
