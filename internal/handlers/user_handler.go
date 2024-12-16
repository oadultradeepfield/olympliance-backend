package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetUserInformation(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":    user.UserID,
		"username":   user.Username,
		"role_id":    user.RoleID,
		"reputation": user.Reputation,
		"is_banned":  user.IsBanned,
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
	})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var input struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.NewPassword) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}

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

	if err := bcrypt.CompareHashAndPassword([]byte(currentUser.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	if input.NewPassword != input.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password and confirmation do not match"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	currentUser.PasswordHash = string(hashedPassword)
	if err := h.db.Save(currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (h *UserHandler) ToggleBanUser(c *gin.Context) {
	userID := c.Param("id")

	var userToBan models.User
	if err := h.db.First(&userToBan, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

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
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to ban users"})
		return
	}

	if currentUserData.RoleID > 1 && userToBan.RoleID > 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot ban another admin"})
		return
	}

	if currentUserData.RoleID == 1 && userToBan.RoleID == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Moderators cannot ban other moderators"})
		return
	}

	userToBan.IsBanned = !userToBan.IsBanned
	if err := h.db.Save(&userToBan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ban status"})
		return
	}

	if userToBan.IsBanned {
		if err := h.db.Model(&models.Comment{}).Where("user_id = ?", userID).Update("is_deleted", true).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user comments"})
			return
		}

		if err := h.db.Model(&models.Thread{}).Where("user_id = ?", userID).Update("is_deleted", true).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user threads"})
			return
		}
	}

	if userToBan.IsBanned {
		c.JSON(http.StatusOK, gin.H{"message": "Successfully banned the user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully unbanned the user"})
}

func (h *UserHandler) ToggleAssignModerator(c *gin.Context) {
	userID := c.Param("id")

	var userToAssign models.User
	if err := h.db.First(&userToAssign, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

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

	if currentUserData.RoleID <= 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can assign moderators"})
		return
	}

	if userToAssign.RoleID > 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot assign moderator to another admin"})
		return
	}

	if userToAssign.RoleID == 1 {
		userToAssign.RoleID = 0
	} else {
		userToAssign.RoleID = 1
	}

	if err := h.db.Save(&userToAssign).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	if userToAssign.RoleID == 1 {
		c.JSON(http.StatusOK, gin.H{"message": "Successfully assigned user as moderator"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully remove user from moderators"})
}
