package user

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (h *UserHandler) ChangeUsername(c *gin.Context) {
	var input struct {
		NewUsername     string `json:"new_username" binding:"required"`
		ConfirmUsername string `json:"confirm_username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isValidUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString
	if !isValidUsername(input.NewUsername) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username can only contain letters, numbers, underscores, and dashes"})
		return
	}

	if input.NewUsername != input.ConfirmUsername {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New username and confirmation do not match"})
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

	var existingUser models.User
	if err := h.db.Where("username = ?", input.NewUsername).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already taken"})
		return
	}

	currentUser.Username = input.NewUsername
	if err := h.db.Save(currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Username updated successfully"})
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

func (h *UserHandler) DeleteUser(c *gin.Context) {
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

	currentUser.IsDeleted = true
	if err := h.db.Save(currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
