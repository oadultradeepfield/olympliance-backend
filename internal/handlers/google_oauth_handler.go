package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oadultradeepfield/olympliance-server/internal/middleware"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOAuth2Config = oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	url := googleOAuth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.DefaultQuery("code", "")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Code is missing"})
		return
	}

	token, err := googleOAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := googleOAuth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		Name string `json:"name"`
		Sub  string `json:"sub"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	var user models.User
	if err := h.db.Where("google_id = ?", userInfo.Sub).First(&user).Error; err != nil {
		if err := h.db.Create(&models.User{
			GoogleID:  userInfo.Sub,
			Username:  userInfo.Name,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT_SECRET not set"})
		return
	}

	accessTokenClaims := middleware.Claims{
		UserID: user.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshTokenClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	backendDomain := os.Getenv("BACKEND_DOMAIN")
	if backendDomain == "" {
		backendDomain = "localhost"
	}

	c.SetCookie("refresh_token", refreshToken, 7*24*60*60, "/", backendDomain, true, true)
	c.Writer.Header().Set("Set-Cookie", fmt.Sprintf("refresh_token=%s; Path=/; HttpOnly; Secure; SameSite=None", refreshToken))

	redirectURL := os.Getenv("FRONTEND_REDIRECT_URL") + "?access_token=" + accessToken
	c.Redirect(http.StatusFound, redirectURL)
}
