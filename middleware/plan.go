package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hotelpms/backend/service"
	"github.com/hotelpms/backend/utils"
)

// RequireFeature returns a middleware that checks if the hotel's plan has the given feature enabled.
// It uses the hotel_id from the URL parameter.
func RequireFeature(planService *service.PlanService, feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Super admins bypass feature checks
		claims := GetClaims(c)
		if claims != nil && claims.Role == "super_admin" {
			c.Next()
			return
		}

		hotelID := c.Param("hotel_id")
		if hotelID == "" {
			// For guest routes, get hotel_id from claims
			if claims != nil {
				hotelID = claims.HotelID
			}
		}

		if hotelID == "" {
			utils.RespondForbidden(c, "unable to determine hotel")
			c.Abort()
			return
		}

		has, err := planService.HasFeature(hotelID, feature)
		if err != nil {
			utils.RespondInternalError(c, "failed to check plan feature")
			c.Abort()
			return
		}

		if !has {
			utils.RespondForbidden(c, "this feature is not available on your current plan. Please upgrade to access it")
			c.Abort()
			return
		}

		c.Next()
	}
}
