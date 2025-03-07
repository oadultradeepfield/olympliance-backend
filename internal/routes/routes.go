package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers/auth"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers/comment"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers/interaction"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers/thread"
	"github.com/oadultradeepfield/olympliance-server/internal/handlers/user"
	"github.com/oadultradeepfield/olympliance-server/internal/middleware"
	"gorm.io/gorm"
)

func InitRoutes(r *gin.Engine, db *gorm.DB) {
	authHandler := auth.NewAuthHandler(db)
	userHandler := user.NewUserHandler(db)
	threadHandler := thread.NewThreadHandler(db)
	commentHandler := comment.NewCommentHandler(db)
	interactionHandler := interaction.NewInteractionHandler(db)

	r.Use(middleware.CorsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Unprotected Routes
	r.GET("/api/userinfo", userHandler.GetUserInformation)
	r.GET("/api/leaderboard", userHandler.GetLeaderboard)
	r.GET("/api/threads/:id", threadHandler.GetThread)
	r.GET("/api/threads/category/:category_id", threadHandler.GetAllThreadsByCategory)
	r.GET("/api/comments", commentHandler.GetAllComments)
	r.GET("/api/interactions", interactionHandler.GetInteraction)

	// Authentication Routes
	r.POST("/api/register", authHandler.Register)
	r.POST("/api/login", authHandler.Login)
	r.GET("/api/auth/google/", authHandler.GoogleLogin)
	r.GET("/api/auth/google/callback", authHandler.GoogleCallback)

	// Protected Routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(db))

	// Users
	api.GET("/users", userHandler.GetCurrentUserInformation)
	api.DELETE("/users/delete", userHandler.DeleteUser)
	api.POST("/logout", authHandler.Logout)

	// Below are routes protected from banned users
	api.Use(middleware.BanCheckMiddleware(db))

	// Users
	api.GET("/users/get-id/:username", userHandler.GetUserIDbyUsername)
	api.PUT("/users/change-username", userHandler.ChangeUsername)
	api.PUT("/users/change-password", userHandler.ChangePassword)
	api.PUT("/users/:id/toggle-ban", userHandler.ToggleBanUser)
	api.PUT("/users/:id/toggle-moderator", userHandler.ToggleAssignModerator)

	// Threads
	api.POST("/threads", threadHandler.CreateThread)
	api.PUT("/threads/:id", threadHandler.UpdateThread)
	api.DELETE("/threads/:id", threadHandler.DeleteThread)
	api.GET("/followed-threads/:id", threadHandler.GetFollowedThreads)

	// Comments
	api.POST("/comments", commentHandler.CreateComment)
	api.PUT("/comments/:id", commentHandler.UpdateComment)
	api.DELETE("/comments/:id", commentHandler.DeleteComment)

	// Interaction
	api.POST("/interactions", interactionHandler.CreateInteraction)
	api.PUT("/interactions/:id", interactionHandler.UpdateInteraction)
}
