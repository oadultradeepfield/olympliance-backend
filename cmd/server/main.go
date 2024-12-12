package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/oadultradeepfield/olympliance-server/internal/databases"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers"
	"github.com/oadultradeepfield/olympliance-server/internal/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	if os.Getenv("GO_ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db := databases.InitDB()
	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	threadHandler := handlers.NewThreadHandler(db)

	r := gin.New()
	r.Use(middleware.CorsMiddleware())

	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Authentication routes
	r.POST("/api/register", authHandler.Register)
	r.POST("/api/login", authHandler.Login)
	r.GET("/api/threads/:id", threadHandler.GetThread)
	r.GET("/api/threads/category/:category_id", threadHandler.GetAllThreadsByCategory)

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(db))

	api.POST("/change-password", userHandler.ChangePassword)

	api.POST("/threads", threadHandler.CreateThread)
	api.PUT("/threads/:id", threadHandler.UpdateThread)
	api.DELETE("/threads/:id", threadHandler.DeleteThread)

	r.Run(":" + os.Getenv("PORT"))
}
