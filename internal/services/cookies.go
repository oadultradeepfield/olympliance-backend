package services

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func SetCookie(c *gin.Context, name, value string, maxAge int) {
	backendDomain := os.Getenv("BACKEND_DOMAIN")
	if backendDomain == "" {
		log.Println("Warning: BACKEND_DOMAIN not set, using default localhost")
		backendDomain = "localhost"
	}

	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(
		name,
		value,
		maxAge,
		"/",
		backendDomain,
		os.Getenv("GO_ENVIRONMENT") == "production",
		true, // HTTPOnly flag
	)
}
