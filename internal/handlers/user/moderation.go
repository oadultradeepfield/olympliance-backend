package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oadultradeepfield/olympliance-server/internal/models"
)

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
