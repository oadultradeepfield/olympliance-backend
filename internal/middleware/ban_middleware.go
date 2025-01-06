package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

func BanCheckMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user data"})
			c.Abort()
			return
		}

		if user.IsBanned {
			c.JSON(http.StatusForbidden, gin.H{"error": "Your account has been banned"})
			c.Abort()
			return
		}
	}
}
