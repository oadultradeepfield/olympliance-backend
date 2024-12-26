package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"gorm.io/gorm"
)

type Claims struct {
	UserID uint
	jwt.RegisteredClaims
}

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			if !handleRefreshFlow(c, db) {
				c.Abort()
			}
			return
		}

		claims := &Claims{}
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT_SECRET not set"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid || claims.ExpiresAt.Before(time.Now()) {
			if !handleRefreshFlow(c, db) {
				c.Abort()
			}
			return
		}

		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}

func handleRefreshFlow(c *gin.Context, db *gorm.DB) bool {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token found"})
		return false
	}

	secretKey := os.Getenv("JWT_SECRET")
	refreshClaims := &Claims{}
	refreshTokenParsed, err := jwt.ParseWithClaims(refreshToken, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !refreshTokenParsed.Valid || refreshClaims.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return false
	}

	var user models.User
	if err := db.First(&user, refreshClaims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return false
	}

	accessTokenClaims := Claims{
		UserID: user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	newAccessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).SignedString([]byte(secretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return false
	}

	backendDomain := os.Getenv("BACKEND_DOMAIN")
	if backendDomain == "" {
		backendDomain = "localhost"
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Path:     "/",
		Domain:   backendDomain,
		Expires:  time.Now().Add(15 * time.Minute),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	c.Set("user", &user)
	return true
}
