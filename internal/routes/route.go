package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers"
	"github.com/oadultradeepfield/olympliance-server/internal/middleware"
	"gorm.io/gorm"
)

func InitRoutes(r *gin.Engine, db *gorm.DB) {
	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	threadHandler := handlers.NewThreadHandler(db)
	commentHandler := handlers.NewCommentHandler(db)
	interactionHandler := handlers.NewInteractionHandler(db)

	r.Use(middleware.CorsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Unprotected Routes
	r.GET("/api/threads/:id", threadHandler.GetThread)
	r.GET("/api/threads/category/:category_id", threadHandler.GetAllThreadsByCategory)
	r.GET("/api/comments", commentHandler.GetAllComments)

	// Authentication Routes
	r.POST("/api/register", authHandler.Register)
	r.POST("/api/login", authHandler.Login)

	// Protected Routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(db))
	api.Use(middleware.BanCheckMiddleware(db))

	// Users
	api.POST("/users/change-password", userHandler.ChangePassword)
	api.POST("/users/:id/toggle-ban", userHandler.ToggleBanUser)
	api.POST("/users/:id/toggle-moderator", userHandler.ToggleAssignModerator)

	// Threads
	api.POST("/threads", threadHandler.CreateThread)
	api.PUT("/threads/:id", threadHandler.UpdateThread)
	api.DELETE("/threads/:id", threadHandler.DeleteThread)

	// Comments
	api.POST("/comments", commentHandler.CreateComment)
	api.PUT("/comments/:id", commentHandler.UpdateComment)
	api.DELETE("/comments/:id", commentHandler.DeleteComment)

	// Interaction
	api.POST("/interactions", interactionHandler.CreateInteraction)
	api.PUT("/interactions/:id", interactionHandler.UpdateInteraction)
}
