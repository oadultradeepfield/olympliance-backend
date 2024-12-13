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

	c.JSON(http.StatusOK, gin.H{"message": "User ban status toggled successfully"})
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

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}
