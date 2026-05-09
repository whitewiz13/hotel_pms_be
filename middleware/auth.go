package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/models"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

const (
	ContextKeyClaims = "claims"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondUnauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			utils.RespondUnauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			utils.RespondUnauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		if claims.Role != models.RoleSuperAdmin && claims.HotelID != "" {
			if err := authService.CheckHotelAccess(claims.HotelID); err != nil {
				utils.RespondForbidden(c, err.Error())
				c.Abort()
				return
			}
		}

		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}
