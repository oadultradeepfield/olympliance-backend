package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
)

func (h *UserHandler) GetUserInformation(c *gin.Context) {
	userID := c.DefaultQuery("id", "")
	username := c.DefaultQuery("username", "")

	if userID == "" && username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either User ID or Username is required"})
		return
	}

	var user models.User
	var err error
	if userID != "" {
		err = h.db.First(&user, "user_id = ?", userID).Error
	} else if username != "" {
		err = h.db.First(&user, "username = ?", username).Error
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    user.UserID,
		"username":   user.Username,
		"role_id":    user.RoleID,
		"reputation": user.Reputation,
		"is_banned":  user.IsBanned,
		"is_deleted": user.IsDeleted,
		"created_at": user.CreatedAt,
	})
}

func (h *UserHandler) GetUserIDbyUsername(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	currentUserData, ok := currentUser.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	if currentUserData.RoleID <= 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to view user IDs"})
		return
	}

	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var targetUser models.User
	if err := h.db.Where("username = ?", username).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if targetUser.RoleID >= currentUserData.RoleID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot view this user's ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": targetUser.UserID})
}

func (h *UserHandler) GetCurrentUserInformation(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    currentUser.UserID,
		"username":   currentUser.Username,
		"role_id":    currentUser.RoleID,
		"reputation": currentUser.Reputation,
		"is_banned":  currentUser.IsBanned,
		"is_deleted": currentUser.IsDeleted,
	})
}

func (h *UserHandler) GetLeaderboard(c *gin.Context) {
	var users []models.User

	if err := h.db.Order("reputation desc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	var leaderboard []gin.H
	for _, user := range users {
		if !user.IsBanned && !user.IsDeleted {
			leaderboard = append(leaderboard, gin.H{
				"user_id":    user.UserID,
				"username":   user.Username,
				"reputation": user.Reputation,
			})
			if len(leaderboard) == 10 {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": leaderboard,
	})
}
